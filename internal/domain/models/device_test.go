package models

import (
	"testing"
	"time"
)

// TestDeviceStatusConstants verifica las constantes de estado.
func TestDeviceStatusConstants(t *testing.T) {
	tests := []struct {
		name   string
		status DeviceStatus
		want   string
	}{
		{"Active status", DeviceStatusActive, "active"},
		{"Inactive status", DeviceStatusInactive, "inactive"},
		{"Maintenance status", DeviceStatusMaintenance, "maintenance"},
		{"Error status", DeviceStatusError, "error"},
		{"Decommissioned status", DeviceStatusDecommissioned, "decommissioned"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("DeviceStatus = %v, want %v", tt.status, tt.want)
			}
		})
	}
}

// TestDeviceIsActive verifica el método IsActive.
func TestDeviceIsActive(t *testing.T) {
	tests := []struct {
		name   string
		status DeviceStatus
		want   bool
	}{
		{"Active device", DeviceStatusActive, true},
		{"Inactive device", DeviceStatusInactive, false},
		{"Maintenance device", DeviceStatusMaintenance, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &Device{Status: tt.status}
			if got := device.IsActive(); got != tt.want {
				t.Errorf("Device.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDeviceIsOnline verifica el método IsOnline.
func TestDeviceIsOnline(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		lastSeen *time.Time
		want     bool
	}{
		{
			name:     "Never seen",
			lastSeen: nil,
			want:     false,
		},
		{
			name: "Seen 2 minutes ago",
			lastSeen: func() *time.Time {
				t := now.Add(-2 * time.Minute)
				return &t
			}(),
			want: true,
		},
		{
			name: "Seen 6 minutes ago",
			lastSeen: func() *time.Time {
				t := now.Add(-6 * time.Minute)
				return &t
			}(),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &Device{LastSeen: tt.lastSeen}
			if got := device.IsOnline(); got != tt.want {
				t.Errorf("Device.IsOnline() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDeviceMarkAsSeen verifica que se actualiza el timestamp.
func TestDeviceMarkAsSeen(t *testing.T) {
	device := &Device{
		DeviceID: "device-001",
		Status:   DeviceStatusActive,
	}

	beforeTime := time.Now()
	device.MarkAsSeen()
	afterTime := time.Now()

	if device.LastSeen == nil {
		t.Error("Expected LastSeen to be set")
	}

	if device.LastSeen.Before(beforeTime) || device.LastSeen.After(afterTime) {
		t.Error("LastSeen timestamp is out of expected range")
	}

	if device.UpdatedAt.Before(beforeTime) || device.UpdatedAt.After(afterTime) {
		t.Error("UpdatedAt timestamp is out of expected range")
	}
}

// TestDeviceNeedsMaintenance verifica la lógica de mantenimiento.
func TestDeviceNeedsMaintenance(t *testing.T) {
	tests := []struct {
		name        string
		status      DeviceStatus
		battery     *int
		fillLevel   *int
		totalErrors int
		want        bool
	}{
		{
			name:   "Status is maintenance",
			status: DeviceStatusMaintenance,
			want:   true,
		},
		{
			name:    "Low battery",
			status:  DeviceStatusActive,
			battery: func() *int { i := 15; return &i }(),
			want:    true,
		},
		{
			name:      "Container almost full",
			status:    DeviceStatusActive,
			fillLevel: func() *int { i := 95; return &i }(),
			want:      true,
		},
		{
			name:        "Too many errors",
			status:      DeviceStatusActive,
			totalErrors: 15,
			want:        true,
		},
		{
			name:    "Good condition",
			status:  DeviceStatusActive,
			battery: func() *int { i := 80; return &i }(),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &Device{
				Status:       tt.status,
				BatteryLevel: tt.battery,
				FillLevel:    tt.fillLevel,
				TotalErrors:  tt.totalErrors,
			}
			if got := device.NeedsMaintenance(); got != tt.want {
				t.Errorf("Device.NeedsMaintenance() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDeviceGetBatteryStatus verifica el estado de batería.
func TestDeviceGetBatteryStatus(t *testing.T) {
	tests := []struct {
		name    string
		battery *int
		want    string
	}{
		{
			name:    "Unknown battery",
			battery: nil,
			want:    "unknown",
		},
		{
			name:    "High battery",
			battery: func() *int { i := 85; return &i }(),
			want:    "high",
		},
		{
			name:    "Medium battery",
			battery: func() *int { i := 60; return &i }(),
			want:    "medium",
		},
		{
			name:    "Low battery",
			battery: func() *int { i := 30; return &i }(),
			want:    "low",
		},
		{
			name:    "Critical battery",
			battery: func() *int { i := 10; return &i }(),
			want:    "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &Device{BatteryLevel: tt.battery}
			if got := device.GetBatteryStatus(); got != tt.want {
				t.Errorf("Device.GetBatteryStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDeviceGetFillStatus verifica el estado de llenado.
func TestDeviceGetFillStatus(t *testing.T) {
	tests := []struct {
		name      string
		fillLevel *int
		want      string
	}{
		{
			name:      "Unknown fill",
			fillLevel: nil,
			want:      "unknown",
		},
		{
			name:      "Empty",
			fillLevel: func() *int { i := 20; return &i }(),
			want:      "empty",
		},
		{
			name:      "Half full",
			fillLevel: func() *int { i := 50; return &i }(),
			want:      "half_full",
		},
		{
			name:      "Almost full",
			fillLevel: func() *int { i := 85; return &i }(),
			want:      "almost_full",
		},
		{
			name:      "Full",
			fillLevel: func() *int { i := 95; return &i }(),
			want:      "full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &Device{FillLevel: tt.fillLevel}
			if got := device.GetFillStatus(); got != tt.want {
				t.Errorf("Device.GetFillStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDeviceHasGoodSignal verifica la señal del dispositivo.
func TestDeviceHasGoodSignal(t *testing.T) {
	tests := []struct {
		name   string
		signal *int
		want   bool
	}{
		{
			name:   "No signal data",
			signal: nil,
			want:   false,
		},
		{
			name:   "Good signal",
			signal: func() *int { i := -50; return &i }(),
			want:   true,
		},
		{
			name:   "Poor signal",
			signal: func() *int { i := -80; return &i }(),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &Device{SignalStrength: tt.signal}
			if got := device.HasGoodSignal(); got != tt.want {
				t.Errorf("Device.HasGoodSignal() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDeviceIncrementJobCount verifica el incremento de jobs.
func TestDeviceIncrementJobCount(t *testing.T) {
	device := &Device{
		DeviceID:  "device-001",
		TotalJobs: 5,
	}

	beforeTime := time.Now()
	device.IncrementJobCount()
	afterTime := time.Now()

	if device.TotalJobs != 6 {
		t.Errorf("Expected TotalJobs to be 6, got %d", device.TotalJobs)
	}

	if device.UpdatedAt.Before(beforeTime) || device.UpdatedAt.After(afterTime) {
		t.Error("UpdatedAt timestamp is out of expected range")
	}
}

// TestDeviceIncrementErrorCount verifica el incremento de errores.
func TestDeviceIncrementErrorCount(t *testing.T) {
	device := &Device{
		DeviceID:    "device-001",
		TotalErrors: 3,
	}

	beforeTime := time.Now()
	device.IncrementErrorCount()
	afterTime := time.Now()

	if device.TotalErrors != 4 {
		t.Errorf("Expected TotalErrors to be 4, got %d", device.TotalErrors)
	}

	if device.UpdatedAt.Before(beforeTime) || device.UpdatedAt.After(afterTime) {
		t.Error("UpdatedAt timestamp is out of expected range")
	}
}

// TestDeviceValidate verifica la validación del dispositivo.
func TestDeviceValidate(t *testing.T) {
	tests := []struct {
		name    string
		device  *Device
		wantErr bool
	}{
		{
			name: "Valid device",
			device: &Device{
				DeviceID:   "device-001",
				DeviceType: "smart_bin_v1",
				Status:     DeviceStatusActive,
			},
			wantErr: false,
		},
		{
			name: "Missing DeviceID",
			device: &Device{
				DeviceID:   "",
				DeviceType: "smart_bin_v1",
				Status:     DeviceStatusActive,
			},
			wantErr: true,
		},
		{
			name: "Missing DeviceType",
			device: &Device{
				DeviceID:   "device-001",
				DeviceType: "",
				Status:     DeviceStatusActive,
			},
			wantErr: true,
		},
		{
			name: "Missing Status",
			device: &Device{
				DeviceID:   "device-001",
				DeviceType: "smart_bin_v1",
				Status:     "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.device.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Device.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDeviceGetDisplayName verifica el nombre de display.
func TestDeviceGetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		device   *Device
		expected string
	}{
		{
			name: "With location area",
			device: &Device{
				DeviceID: "device-001",
				Location: &Location{
					Area: "Cafeteria",
				},
			},
			expected: "Cafeteria - device-001",
		},
		{
			name: "Without location",
			device: &Device{
				DeviceID: "device-001",
				Location: nil,
			},
			expected: "device-001",
		},
		{
			name: "Location without area",
			device: &Device{
				DeviceID: "device-001",
				Location: &Location{
					Building: "Building A",
				},
			},
			expected: "device-001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.device.GetDisplayName()
			if got != tt.expected {
				t.Errorf("Device.GetDisplayName() = %v, want %v", got, tt.expected)
			}
		})
	}
}
