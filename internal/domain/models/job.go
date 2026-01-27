// Package models defines the domain models for the Smart Bin Orchestrator.
// It includes Job, Device, Classification, and Decision types with their validation logic.
package models

import (
	"time"
)

// JobStatus represents the state of a classification job.
type JobStatus string

const (
	// JobStatusPending indicates the job is created and waiting for image upload.
	JobStatusPending JobStatus = "pending"

	// JobStatusUploading indicates the image is being uploaded to S3.
	JobStatusUploading JobStatus = "uploading"

	// JobStatusProcessing indicates the image is being processed by the Classifier.
	JobStatusProcessing JobStatus = "processing"

	// JobStatusCompleted indicates the job completed successfully.
	JobStatusCompleted JobStatus = "completed"

	// JobStatusFailed indicates the job failed at some step.
	JobStatusFailed JobStatus = "failed"
)

// Job represents a waste classification task that progresses through various states.
type Job struct {
	JobID               string                 `json:"job_id" dynamodbav:"job_id"`
	DeviceID            string                 `json:"device_id" dynamodbav:"device_id"`
	Status              JobStatus              `json:"status" dynamodbav:"status"`
	ErrorMessage        string                 `json:"error_message,omitempty" dynamodbav:"error_message,omitempty"`
	ImageKey            string                 `json:"image_key,omitempty" dynamodbav:"image_key,omitempty"`
	UploadURL           string                 `json:"upload_url,omitempty" dynamodbav:"-"`
	Classification      *Classification        `json:"classification,omitempty" dynamodbav:"classification,omitempty"`
	Decision            *Decision              `json:"decision,omitempty" dynamodbav:"decision,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty" dynamodbav:"metadata,omitempty"`
	CreatedAt           time.Time              `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at" dynamodbav:"updated_at"`
	CompletedAt         *time.Time             `json:"completed_at,omitempty" dynamodbav:"completed_at,omitempty"`
	ProcessingStartedAt *time.Time             `json:"processing_started_at,omitempty" dynamodbav:"processing_started_at,omitempty"`
	ClassificationTime  *int64                 `json:"classification_time_ms,omitempty" dynamodbav:"classification_time_ms,omitempty"`
}

// CreateJobRequest contains the parameters for creating a new classification job.
type CreateJobRequest struct {
	DeviceID  string                 `json:"device_id" binding:"required"`
	Timestamp string                 `json:"timestamp" binding:"required"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// CreateJobResponse contains the response data when creating a new job.
type CreateJobResponse struct {
	JobID           string    `json:"job_id"`
	DeviceID        string    `json:"device_id"`
	Status          JobStatus `json:"status"`
	UploadURL       string    `json:"upload_url"`
	UploadExpiresAt time.Time `json:"upload_expires_at"`
	CreatedAt       time.Time `json:"created_at"`
}

// GetJobResponse wraps the complete Job data for retrieval responses.
type GetJobResponse struct {
	*Job
}

// ListJobsRequest contains filter parameters for listing jobs.
type ListJobsRequest struct {
	DeviceID string    `form:"device_id"`
	Status   JobStatus `form:"status"`
	Limit    int       `form:"limit,default=10"`
	Offset   int       `form:"offset,default=0"`
}

// ListJobsResponse contains the paginated list of jobs.
type ListJobsResponse struct {
	Jobs   []*Job `json:"jobs"`
	Total  int    `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// UpdateJobRequest contains fields that can be updated on a job.
type UpdateJobRequest struct {
	Status         *JobStatus      `json:"status,omitempty"`
	Classification *Classification `json:"classification,omitempty"`
	Decision       *Decision       `json:"decision,omitempty"`
	ErrorMessage   *string         `json:"error_message,omitempty"`
}

// IsCompleted returns true if the job is in a terminal state (completed or failed).
func (j *Job) IsCompleted() bool {
	return j.Status == JobStatusCompleted || j.Status == JobStatusFailed
}

// IsPending returns true if the job is pending image upload.
func (j *Job) IsPending() bool {
	return j.Status == JobStatusPending
}

// IsProcessing returns true if the job is currently being processed.
func (j *Job) IsProcessing() bool {
	return j.Status == JobStatusProcessing
}

// CanTransitionTo checks if the job can transition from its current state to the new state.
func (j *Job) CanTransitionTo(newStatus JobStatus) bool {
	validTransitions := map[JobStatus][]JobStatus{
		JobStatusPending:    {JobStatusUploading, JobStatusProcessing, JobStatusFailed},
		JobStatusUploading:  {JobStatusProcessing, JobStatusFailed},
		JobStatusProcessing: {JobStatusCompleted, JobStatusFailed},
		JobStatusCompleted:  {},
		JobStatusFailed:     {},
	}

	allowedStatuses, exists := validTransitions[j.Status]
	if !exists {
		return false
	}

	for _, status := range allowedStatuses {
		if status == newStatus {
			return true
		}
	}

	return false
}

// MarkAsProcessing updates the job status to processing and records the start time.
func (j *Job) MarkAsProcessing() {
	now := time.Now()
	j.Status = JobStatusProcessing
	j.ProcessingStartedAt = &now
	j.UpdatedAt = now
}

// MarkAsCompleted updates the job status to completed and records the completion time.
func (j *Job) MarkAsCompleted() {
	now := time.Now()
	j.Status = JobStatusCompleted
	j.CompletedAt = &now
	j.UpdatedAt = now
}

// MarkAsFailed updates the job status to failed with an error message.
func (j *Job) MarkAsFailed(errorMsg string) {
	now := time.Now()
	j.Status = JobStatusFailed
	j.ErrorMessage = errorMsg
	j.CompletedAt = &now
	j.UpdatedAt = now
}

// GetProcessingDuration returns the time taken to process the job.
// If the job is still processing, it returns the elapsed time since processing started.
func (j *Job) GetProcessingDuration() *time.Duration {
	if j.ProcessingStartedAt == nil {
		return nil
	}

	var endTime time.Time
	if j.CompletedAt != nil {
		endTime = *j.CompletedAt
	} else {
		endTime = time.Now()
	}

	duration := endTime.Sub(*j.ProcessingStartedAt)
	return &duration
}

// HasClassification returns true if the job has classification results.
func (j *Job) HasClassification() bool {
	return j.Classification != nil && j.Classification.Label != ""
}

// HasDecision returns true if the job has a decision.
func (j *Job) HasDecision() bool {
	return j.Decision != nil && j.Decision.Action != ""
}

// Validate checks that the job has all required fields.
func (j *Job) Validate() error {
	if j.JobID == "" {
		return ErrInvalidJobID
	}
	if j.DeviceID == "" {
		return ErrInvalidDeviceID
	}
	if j.Status == "" {
		return ErrInvalidStatus
	}
	return nil
}
