package models

import (
	"time"
)

/*
╔════════════════════════════════════════════════════════════════╗
║                                                                ║
║  DEVICE.GO - MODELO DE DISPOSITIVO IOT                         ║
║                                                                ║
║  Representa un dispositivo Smart Bin que captura imágenes      ║
║  y envía residuos para clasificación.                          ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝
*/

// DeviceStatus representa el estado de un dispositivo.
type DeviceStatus string

const (
	// DeviceStatusActive - Dispositivo activo y funcionando.
	DeviceStatusActive DeviceStatus = "active"

	// DeviceStatusInactive - Dispositivo inactivo o apagado.
	DeviceStatusInactive DeviceStatus = "inactive"

	// DeviceStatusMaintenance - Dispositivo en mantenimiento.
	DeviceStatusMaintenance DeviceStatus = "maintenance"

	// DeviceStatusError - Dispositivo con error.
	DeviceStatusError DeviceStatus = "error"

	// DeviceStatusDecommissioned - Dispositivo dado de baja.
	DeviceStatusDecommissioned DeviceStatus = "decommissioned"
)

// DeviceType representa el tipo/modelo de dispositivo.
type DeviceType string

const (
	// Versión 1 del Smart Bin.
	DeviceTypeSmartBinV1 DeviceType = "smart_bin_v1"

	// Versión 2 del Smart Bin (con más sensors).
	DeviceTypeSmartBinV2 DeviceType = "smart_bin_v2"

	// Smart Bin industrial (mayor capacidad).
	DeviceTypeSmartBinIndustrial DeviceType = "smart_bin_industrial"
)

// BinType representa el tipo de contenedor.
type BinType string

const (
	// BinTypeRecyclable - Contenedor para reciclables.
	BinTypeRecyclable BinType = "recyclable"

	// BinTypeOrganic - Contenedor para residuos orgánicos.
	BinTypeOrganic BinType = "organic"

	// BinTypeGeneral - Contenedor para residuos generales.
	BinTypeGeneral BinType = "general"

	// BinTypeMixed - Contenedor mixto (múltiples compartimentos).
	BinTypeMixed BinType = "mixed"
)

// Device representa un dispositivo IoT Smart Bin.
type Device struct {
	// ═══════════════════════════════════════════════════════════════
	// IDENTIFICACIÓN
	// ═══════════════════════════════════════════════════════════════

	// DeviceID - ID único del dispositivo (ej: "smart-bin-001")
	DeviceID string `json:"device_id" dynamodbav:"device_id"`

	// DeviceType - Tipo/modelo del dispositivo
	DeviceType string `json:"device_type" dynamodbav:"device_type"`

	// SerialNumber - Número de serie del hardware (opcional)
	SerialNumber string `json:"serial_number,omitempty" dynamodbav:"serial_number,omitempty"`

	// ═══════════════════════════════════════════════════════════════
	// ESTADO
	// ═══════════════════════════════════════════════════════════════

	// Status - Estado actual del dispositivo
	Status DeviceStatus `json:"status" dynamodbav:"status"`

	// StatusReason - Razón del estado actual (si aplica)
	StatusReason string `json:"status_reason,omitempty" dynamodbav:"status_reason,omitempty"`

	// ═══════════════════════════════════════════════════════════════
	// UBICACIÓN
	// ═══════════════════════════════════════════════════════════════

	// Location - Ubicación física del dispositivo
	Location *Location `json:"location,omitempty" dynamodbav:"location,omitempty"`

	// ═══════════════════════════════════════════════════════════════
	// CONFIGURACIÓN
	// ═══════════════════════════════════════════════════════════════

	// BinType - Tipo de contenedor
	BinType string `json:"bin_type,omitempty" dynamodbav:"bin_type,omitempty"`

	// Capacity - Capacidad del contenedor en litros (opcional)
	Capacity int `json:"capacity,omitempty" dynamodbav:"capacity,omitempty"`

	// ═══════════════════════════════════════════════════════════════
	// CONECTIVIDAD IOT
	// ═══════════════════════════════════════════════════════════════

	// Certificate - Certificado de AWS IoT Core
	Certificate string `json:"certificate,omitempty" dynamodbav:"certificate,omitempty"`

	// ThingName - Nombre del Thing en AWS IoT Core
	ThingName string `json:"thing_name,omitempty" dynamodbav:"thing_name,omitempty"`

	// ═══════════════════════════════════════════════════════════════
	// ESTADO DEL HARDWARE
	// ═══════════════════════════════════════════════════════════════

	// BatteryLevel - Nivel de batería (0-100)
	BatteryLevel *int `json:"battery_level,omitempty" dynamodbav:"battery_level,omitempty"`

	// FillLevel - Nivel de llenado del contenedor (0-100)
	FillLevel *int `json:"fill_level,omitempty" dynamodbav:"fill_level,omitempty"`

	// SignalStrength - Fuerza de señal en dBm (ej: -45)
	SignalStrength *int `json:"signal_strength,omitempty" dynamodbav:"signal_strength,omitempty"`

	// ═══════════════════════════════════════════════════════════════
	// METADATA
	// ═══════════════════════════════════════════════════════════════

	// Metadata - Datos adicionales del dispositivo
	Metadata map[string]interface{} `json:"metadata,omitempty" dynamodbav:"metadata,omitempty"`

	// ═══════════════════════════════════════════════════════════════
	// TIMESTAMPS
	// ═══════════════════════════════════════════════════════════════

	// LastSeen - Última vez que el dispositivo se conectó
	LastSeen *time.Time `json:"last_seen,omitempty" dynamodbav:"last_seen,omitempty"`

	// CreatedAt - Moment en que se registró el dispositivo
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`

	// UpdatedAt - Última actualización del dispositivo
	UpdatedAt time.Time `json:"updated_at" dynamodbav:"updated_at"`

	// ═══════════════════════════════════════════════════════════════
	// ESTADÍSTICAS
	// ═══════════════════════════════════════════════════════════════

	// TotalJobs - Total de jobs procesados por este dispositivo
	TotalJobs int `json:"total_jobs,omitempty" dynamodbav:"total_jobs,omitempty"`

	// TotalErrors - Total de errores del dispositivo
	TotalErrors int `json:"total_errors,omitempty" dynamodbav:"total_errors,omitempty"`
}

// Location representa la ubicación física de un dispositivo.
type Location struct {
	// Building - Nombre del edificio (ej: "Edificio A")
	Building string `json:"building,omitempty" dynamodbav:"building,omitempty"`

	// Floor - Piso (ej: 2)
	Floor int `json:"floor,omitempty" dynamodbav:"floor,omitempty"`

	// Area - Área específica (ej: "Cafeteria", "Recepción")
	Area string `json:"area,omitempty" dynamodbav:"area,omitempty"`

	// Zone - Zona o sector (ej: "Norte", "Zona A")
	Zone string `json:"zone,omitempty" dynamodbav:"zone,omitempty"`

	// Latitude - Latitud GPS (ej: 4.6097)
	Latitude float64 `json:"latitude,omitempty" dynamodbav:"latitude,omitempty"`

	// Longitude - Longitud GPS (ej: -74.0817)
	Longitude float64 `json:"longitude,omitempty" dynamodbav:"longitude,omitempty"`
}

// ═══════════════════════════════════════════════════════════════════
//                     DTOs (Data Transfer Objects)
// ═══════════════════════════════════════════════════════════════════

// RegisterDeviceRequest - Request para registrar un nuevo dispositivo.
type RegisterDeviceRequest struct {
	// DeviceID - ID del dispositivo (obligatorio)
	DeviceID string `json:"device_id" binding:"required"`

	// DeviceType - Tipo de dispositivo (obligatorio)
	DeviceType string `json:"device_type" binding:"required"`

	// SerialNumber - Número de serie (opcional)
	SerialNumber string `json:"serial_number,omitempty"`

	// Location - Ubicación del dispositivo (opcional)
	Location *Location `json:"location,omitempty"`

	// BinType - Tipo de contenedor (opcional)
	BinType string `json:"bin_type,omitempty"`

	// Capacity - Capacidad en litros (opcional)
	Capacity int `json:"capacity,omitempty"`

	// Metadata - Datos adicionales (opcional)
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// RegisterDeviceResponse - Response al registrar un dispositivo.
type RegisterDeviceResponse struct {
	DeviceID    string       `json:"device_id"`
	DeviceType  string       `json:"device_type"`
	Status      DeviceStatus `json:"status"`
	Certificate string       `json:"certificate"`
	CreatedAt   time.Time    `json:"created_at"`
}

// UpdateDeviceRequest - Request para actualizar un dispositivo.
type UpdateDeviceRequest struct {
	Status         *DeviceStatus `json:"status,omitempty"`
	StatusReason   *string       `json:"status_reason,omitempty"`
	Location       *Location     `json:"location,omitempty"`
	BatteryLevel   *int          `json:"battery_level,omitempty"`
	FillLevel      *int          `json:"fill_level,omitempty"`
	SignalStrength *int          `json:"signal_strength,omitempty"`
}

// ═══════════════════════════════════════════════════════════════════
//                     MÉTODOS DEL MODELO
// ═══════════════════════════════════════════════════════════════════

// IsActive retorna true si el dispositivo está activo.
func (d *Device) IsActive() bool {
	return d.Status == DeviceStatusActive
}

// Se considera online si se conectó en los últimos 5 minutos.
func (d *Device) IsOnline() bool {
	if d.LastSeen == nil {
		return false
	}
	return time.Since(*d.LastSeen) < 5*time.Minute
}

// MarkAsSeen actualiza el timestamp de LastSeen.
func (d *Device) MarkAsSeen() {
	now := time.Now()
	d.LastSeen = &now
	d.UpdatedAt = now
}

// NeedsMaintenance retorna true si el dispositivo requiere mantenimiento.
func (d *Device) NeedsMaintenance() bool {
	// Si está en estado de mantenimiento
	if d.Status == DeviceStatusMaintenance {
		return true
	}

	// Si la batería está baja (< 20%)
	if d.BatteryLevel != nil && *d.BatteryLevel < 20 {
		return true
	}

	// Si el contenedor está casi lleno (> 90%)
	if d.FillLevel != nil && *d.FillLevel > 90 {
		return true
	}

	// Si tiene muchos errores
	if d.TotalErrors > 10 {
		return true
	}

	return false
}

// GetBatteryStatus retorna el estado de la batería como string.
func (d *Device) GetBatteryStatus() string {
	if d.BatteryLevel == nil {
		return "unknown"
	}

	level := *d.BatteryLevel
	switch {
	case level > 80:
		return "high"
	case level > 40:
		return "medium"
	case level > 20:
		return "low"
	default:
		return "critical"
	}
}

// GetFillStatus retorna el estado de llenado del contenedor.
func (d *Device) GetFillStatus() string {
	if d.FillLevel == nil {
		return "unknown"
	}

	level := *d.FillLevel
	switch {
	case level < 30:
		return "empty"
	case level < 70:
		return "half_full"
	case level < 90:
		return "almost_full"
	default:
		return "full"
	}
}

// HasGoodSignal retorna true si la señal es buena (> -70 dBm).
func (d *Device) HasGoodSignal() bool {
	if d.SignalStrength == nil {
		return false
	}
	return *d.SignalStrength > -70
}

// IncrementJobCount incrementa el contador de jobs.
func (d *Device) IncrementJobCount() {
	d.TotalJobs++
	d.UpdatedAt = time.Now()
}

// IncrementErrorCount incrementa el contador de errores.
func (d *Device) IncrementErrorCount() {
	d.TotalErrors++
	d.UpdatedAt = time.Now()
}

// Validate valida que el dispositivo tenga los campos obligatorios.
func (d *Device) Validate() error {
	if d.DeviceID == "" {
		return ErrInvalidDeviceID
	}
	if d.DeviceType == "" {
		return ErrInvalidDeviceType
	}
	if d.Status == "" {
		return ErrInvalidStatus
	}
	return nil
}

// GetDisplayName retorna un nombre legible del dispositivo.
func (d *Device) GetDisplayName() string {
	if d.Location != nil && d.Location.Area != "" {
		return d.Location.Area + " - " + d.DeviceID
	}
	return d.DeviceID
}

/*
═══════════════════════════════════════════════════════════════════
                    EJEMPLO DE USO
═══════════════════════════════════════════════════════════════════

// Crear un nuevo dispositivo
device := &Device{
    DeviceID:   "smart-bin-001",
    DeviceType: "smart_bin_v1",
    Status:     DeviceStatusActive,
    BinType:    "recyclable",
    Location: &Location{
        Building: "Edificio A",
        Floor:    2,
        Area:     "Cafeteria",
        Latitude:  4.6097,
        Longitude: -74.0817,
    },
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}

// Actualizar estado
device.MarkAsSeen()

// Verificar si está online
if device.IsOnline() {
    fmt.Println("Device is online")
}

// Verificar batería
batteryLevel := 75
device.BatteryLevel = &batteryLevel
fmt.Printf("Battery status: %s\n", device.GetBatteryStatus())

// Verificar si necesita mantenimiento
if device.NeedsMaintenance() {
    fmt.Println("Device needs maintenance")
}

// Incrementar contador de jobs
device.IncrementJobCount()

// Validar
if err := device.Validate(); err != nil {
    fmt.Printf("Error: %s\n", err)
}

═══════════════════════════════════════════════════════════════════
*/
