package models

import (
	"testing"
	"time"
)

// TestJobStatusConstants verifica que las constantes de estado están definidas correctamente.
func TestJobStatusConstants(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		want   string
	}{
		{"Pending status", JobStatusPending, "pending"},
		{"Uploading status", JobStatusUploading, "uploading"},
		{"Processing status", JobStatusProcessing, "processing"},
		{"Completed status", JobStatusCompleted, "completed"},
		{"Failed status", JobStatusFailed, "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("JobStatus = %v, want %v", tt.status, tt.want)
			}
		})
	}
}

// TestJobIsCompleted verifica el método IsCompleted.
func TestJobIsCompleted(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		want   bool
	}{
		{"Completed job", JobStatusCompleted, true},
		{"Failed job", JobStatusFailed, true},
		{"Pending job", JobStatusPending, false},
		{"Processing job", JobStatusProcessing, false},
		{"Uploading job", JobStatusUploading, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{Status: tt.status}
			if got := job.IsCompleted(); got != tt.want {
				t.Errorf("Job.IsCompleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestJobIsPending verifica el método IsPending.
func TestJobIsPending(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		want   bool
	}{
		{"Pending job", JobStatusPending, true},
		{"Completed job", JobStatusCompleted, false},
		{"Processing job", JobStatusProcessing, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{Status: tt.status}
			if got := job.IsPending(); got != tt.want {
				t.Errorf("Job.IsPending() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestJobIsProcessing verifica el método IsProcessing.
func TestJobIsProcessing(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		want   bool
	}{
		{"Processing job", JobStatusProcessing, true},
		{"Pending job", JobStatusPending, false},
		{"Completed job", JobStatusCompleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{Status: tt.status}
			if got := job.IsProcessing(); got != tt.want {
				t.Errorf("Job.IsProcessing() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestJobCanTransitionTo verifica las transiciones de estado válidas.
func TestJobCanTransitionTo(t *testing.T) {
	tests := []struct {
		name       string
		fromStatus JobStatus
		toStatus   JobStatus
		canTransit bool
	}{
		// Desde Pending
		{"Pending to Uploading", JobStatusPending, JobStatusUploading, true},
		{"Pending to Processing", JobStatusPending, JobStatusProcessing, true},
		{"Pending to Failed", JobStatusPending, JobStatusFailed, true},
		{"Pending to Completed", JobStatusPending, JobStatusCompleted, false},

		// Desde Uploading
		{"Uploading to Processing", JobStatusUploading, JobStatusProcessing, true},
		{"Uploading to Failed", JobStatusUploading, JobStatusFailed, true},
		{"Uploading to Completed", JobStatusUploading, JobStatusCompleted, false},

		// Desde Processing
		{"Processing to Completed", JobStatusProcessing, JobStatusCompleted, true},
		{"Processing to Failed", JobStatusProcessing, JobStatusFailed, true},
		{"Processing to Pending", JobStatusProcessing, JobStatusPending, false},

		// Desde estados finales
		{"Completed to any", JobStatusCompleted, JobStatusPending, false},
		{"Failed to any", JobStatusFailed, JobStatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{Status: tt.fromStatus}
			if got := job.CanTransitionTo(tt.toStatus); got != tt.canTransit {
				t.Errorf("Job.CanTransitionTo(%v) = %v, want %v", tt.toStatus, got, tt.canTransit)
			}
		})
	}
}

// TestJobMarkAsProcessing verifica que el método actualiza correctamente el estado.
func TestJobMarkAsProcessing(t *testing.T) {
	job := &Job{
		JobID:    "test-job-1",
		DeviceID: "device-1",
		Status:   JobStatusPending,
	}

	beforeTime := time.Now()
	job.MarkAsProcessing()
	afterTime := time.Now()

	if job.Status != JobStatusProcessing {
		t.Errorf("Expected status to be Processing, got %v", job.Status)
	}

	if job.ProcessingStartedAt == nil {
		t.Error("Expected ProcessingStartedAt to be set")
	}

	if job.ProcessingStartedAt.Before(beforeTime) || job.ProcessingStartedAt.After(afterTime) {
		t.Error("ProcessingStartedAt timestamp is out of expected range")
	}

	if job.UpdatedAt.Before(beforeTime) || job.UpdatedAt.After(afterTime) {
		t.Error("UpdatedAt timestamp is out of expected range")
	}
}

// TestJobMarkAsCompleted verifica que el método actualiza correctamente el estado.
func TestJobMarkAsCompleted(t *testing.T) {
	job := &Job{
		JobID:    "test-job-1",
		DeviceID: "device-1",
		Status:   JobStatusProcessing,
	}

	beforeTime := time.Now()
	job.MarkAsCompleted()
	afterTime := time.Now()

	if job.Status != JobStatusCompleted {
		t.Errorf("Expected status to be Completed, got %v", job.Status)
	}

	if job.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}

	if job.CompletedAt.Before(beforeTime) || job.CompletedAt.After(afterTime) {
		t.Error("CompletedAt timestamp is out of expected range")
	}
}

// TestJobMarkAsFailed verifica que el método actualiza correctamente el estado.
func TestJobMarkAsFailed(t *testing.T) {
	job := &Job{
		JobID:    "test-job-1",
		DeviceID: "device-1",
		Status:   JobStatusProcessing,
	}

	errorMsg := "Test error message"
	beforeTime := time.Now()
	job.MarkAsFailed(errorMsg)
	afterTime := time.Now()

	if job.Status != JobStatusFailed {
		t.Errorf("Expected status to be Failed, got %v", job.Status)
	}

	if job.ErrorMessage != errorMsg {
		t.Errorf("Expected error message to be %v, got %v", errorMsg, job.ErrorMessage)
	}

	if job.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}

	if job.CompletedAt.Before(beforeTime) || job.CompletedAt.After(afterTime) {
		t.Error("CompletedAt timestamp is out of expected range")
	}
}

// TestJobHasClassification verifica si el job tiene clasificación.
func TestJobHasClassification(t *testing.T) {
	tests := []struct {
		name           string
		classification *Classification
		want           bool
	}{
		{
			name: "With classification",
			classification: &Classification{
				Label:      "plastic_bottle",
				Confidence: 0.95,
			},
			want: true,
		},
		{
			name:           "Without classification",
			classification: nil,
			want:           false,
		},
		{
			name: "Classification with empty label",
			classification: &Classification{
				Label:      "",
				Confidence: 0.95,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{Classification: tt.classification}
			if got := job.HasClassification(); got != tt.want {
				t.Errorf("Job.HasClassification() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestJobHasDecision verifica si el job tiene decisión.
func TestJobHasDecision(t *testing.T) {
	tests := []struct {
		name     string
		decision *Decision
		want     bool
	}{
		{
			name: "With decision",
			decision: &Decision{
				Action:         "accept",
				BinCompartment: "recyclable",
			},
			want: true,
		},
		{
			name:     "Without decision",
			decision: nil,
			want:     false,
		},
		{
			name: "Decision with empty action",
			decision: &Decision{
				Action:         "",
				BinCompartment: "recyclable",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{Decision: tt.decision}
			if got := job.HasDecision(); got != tt.want {
				t.Errorf("Job.HasDecision() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestJobValidate verifica la validación del job.
func TestJobValidate(t *testing.T) {
	tests := []struct {
		name    string
		job     *Job
		wantErr bool
	}{
		{
			name: "Valid job",
			job: &Job{
				JobID:    "job-123",
				DeviceID: "device-001",
				Status:   JobStatusPending,
			},
			wantErr: false,
		},
		{
			name: "Missing JobID",
			job: &Job{
				JobID:    "",
				DeviceID: "device-001",
				Status:   JobStatusPending,
			},
			wantErr: true,
		},
		{
			name: "Missing DeviceID",
			job: &Job{
				JobID:    "job-123",
				DeviceID: "",
				Status:   JobStatusPending,
			},
			wantErr: true,
		},
		{
			name: "Missing Status",
			job: &Job{
				JobID:    "job-123",
				DeviceID: "device-001",
				Status:   "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.job.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Job.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestJobGetProcessingDuration verifica el cálculo de duración.
func TestJobGetProcessingDuration(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name                string
		processingStartedAt *time.Time
		completedAt         *time.Time
		wantNil             bool
	}{
		{
			name:                "No processing started",
			processingStartedAt: nil,
			completedAt:         nil,
			wantNil:             true,
		},
		{
			name: "Completed job",
			processingStartedAt: func() *time.Time {
				t := now.Add(-5 * time.Minute)
				return &t
			}(),
			completedAt: &now,
			wantNil:     false,
		},
		{
			name: "Still processing",
			processingStartedAt: func() *time.Time {
				t := now.Add(-2 * time.Minute)
				return &t
			}(),
			completedAt: nil,
			wantNil:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{
				ProcessingStartedAt: tt.processingStartedAt,
				CompletedAt:         tt.completedAt,
			}

			duration := job.GetProcessingDuration()

			if tt.wantNil {
				if duration != nil {
					t.Error("Expected nil duration")
				}
			} else {
				if duration == nil {
					t.Error("Expected non-nil duration")
				}
			}
		})
	}
}
