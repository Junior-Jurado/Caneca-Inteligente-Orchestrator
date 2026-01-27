// Package handlers provides HTTP request handlers for the Smart Bin Orchestrator API.
// It includes handlers for devices, jobs, webhooks, and health checks.
package handlers

import (
	"net/http"
	"time"

	"github.com/Junior_Jurado/Caneca-Inteligente-Orchestrator/internal/config"
	"github.com/Junior_Jurado/Caneca-Inteligente-Orchestrator/internal/domain/models"
	"github.com/gin-gonic/gin"
)

/*
╔════════════════════════════════════════════════════════════════╗
║                                                            ║
║  DEVICES.GO - HANDLER DE DISPOSITIVOS IOT                  ║
║                                                            ║
║  Endpoints para gestionar dispositivos Smart Bin:          ║
║  - POST   /api/v1/devices/register   - Registrar device    ║
║  - GET    /api/v1/devices/:device_id - Obtener device      ║
║  - GET    /api/v1/devices            - Listar devices      ║
║  - PATCH  /api/v1/devices/:device_id - Actualizar device   ║
║  - DELETE /api/v1/devices/:device_id - Eliminar device     ║
║                                                            ║
╚════════════════════════════════════════════════════════════════╝
*/

// DevicesHandler maneja los endpoints relacionados con dispositivos.
type DevicesHandler struct {
	config *config.Config
	// TODO: Agregar dependencies:
	// deviceRepository ports.DeviceRepository
	// iotClient        ports.IoTPublisher
}

// NewDevicesHandler crea una nueva instancia de DevicesHandler.
func NewDevicesHandler(cfg *config.Config) *DevicesHandler {
	return &DevicesHandler{
		config: cfg,
	}
}

// RegisterDevice handles device registration requests.
// It creates a new IoT device, generates certificates, and stores device info.
//
// Request Body:
//
//	{
//	  "device_id": "smart-bin-001",
//	  "device_type": "smart_bin_v1",
//	  "location": {
//	    "building": "Edificio A",
//	    "floor": 2,
//	    "area": "Cafeteria"
//	  },
//	  "bin_type": "recyclable"
//	}
//
// Response (201 Created):
//
//	{
//	  "success": true,
//	  "data": {
//	    "device_id": "smart-bin-001",
//	    "device_type": "smart_bin_v1",
//	    "status": "active",
//	    "certificate": "-----BEGIN CERTIFICATE-----...",
//	    "created_at": "2026-01-21T10:30:00Z"
//	  }
//	}
func (h *DevicesHandler) RegisterDevice(c *gin.Context) {
	// ───────────────────────────────────────────────────────────────
	// 1. PARSEAR Y VALIDAR REQUEST
	// ───────────────────────────────────────────────────────────────
	var req models.RegisterDeviceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "Invalid request body",
				"details": gin.H{
					"error": err.Error(),
				},
			},
			"metadata": h.buildMetadata(c),
		})
		return
	}

	// ───────────────────────────────────────────────────────────────
	// 2. VERIFICAR QUE EL DEVICE NO EXISTA
	// ───────────────────────────────────────────────────────────────
	// TODO: Verificar en DynamoDB
	// existingDevice, err := h.deviceRepository.GetByID(c.Request.Context(), req.DeviceID)
	// if existingDevice != nil {
	//     c.JSON(http.StatusConflict, gin.H{
	//         "success": false,
	//         "error": gin.H{
	//             "code":    "DEVICE_ALREADY_EXISTS",
	//             "message": "Device with this ID already exists",
	//         },
	//     })
	//     return
	// }

	// ───────────────────────────────────────────────────────────────
	// 3. CREAR DEVICE EN AWS IOT CORE
	// ───────────────────────────────────────────────────────────────
	// TODO: Crear Thing en IoT Core y generar certificados
	// certificate, err := h.iotClient.CreateThing(c.Request.Context(), req.DeviceID)

	// Por ahora, generar certificado mock
	certificate := h.generateMockCertificate(req.DeviceID)

	// ───────────────────────────────────────────────────────────────
	// 4. CREAR OBJETO DEVICE
	// ───────────────────────────────────────────────────────────────
	now := time.Now()
	device := &models.Device{
		DeviceID:    req.DeviceID,
		DeviceType:  req.DeviceType,
		Status:      models.DeviceStatusActive,
		Certificate: certificate,
		CreatedAt:   now,
	}

	// ───────────────────────────────────────────────────────────────
	// 5. GUARDAR DEVICE EN DYNAMODB
	// ───────────────────────────────────────────────────────────────
	// TODO: Guardar en DynamoDB
	// if err := h.deviceRepository.Create(c.Request.Context(), device); err != nil {
	//     c.JSON(http.StatusInternalServerError, gin.H{...})
	//     return
	// }

	// ───────────────────────────────────────────────────────────────
	// 6. CONSTRUIR RESPONSE
	// ───────────────────────────────────────────────────────────────
	response := models.RegisterDeviceResponse{
		DeviceID:    device.DeviceID,
		DeviceType:  device.DeviceType,
		Status:      device.Status,
		Certificate: device.Certificate,
		CreatedAt:   device.CreatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":  true,
		"data":     response,
		"metadata": h.buildMetadata(c),
	})
}

// GetDevice retrieves device information by ID.
//
// Response (200 OK):
//
//	{
//	  "success": true,
//	  "data": {
//	    "device_id": "smart-bin-001",
//	    "status": "active",
//	    "location": {...},
//	    "bin_type": "recyclable",
//	    "last_seen": "2026-01-21T10:28:00Z"
//	  }
//	}
func (h *DevicesHandler) GetDevice(c *gin.Context) {
	// ───────────────────────────────────────────────────────────────
	// 1. OBTENER DEVICE ID DEL PATH
	// ───────────────────────────────────────────────────────────────
	deviceID := c.Param("device_id")

	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "Device ID is required",
			},
			"metadata": h.buildMetadata(c),
		})
		return
	}

	// ───────────────────────────────────────────────────────────────
	// 2. OBTENER DEVICE DE DYNAMODB
	// ───────────────────────────────────────────────────────────────
	// TODO: Obtener de DynamoDB
	// device, err := h.deviceRepository.GetByID(c.Request.Context(), deviceID)
	// if err != nil {
	//     if err == ErrNotFound {
	//         c.JSON(http.StatusNotFound, gin.H{...})
	//         return
	//     }
	//     c.JSON(http.StatusInternalServerError, gin.H{...})
	//     return
	// }

	// Por ahora, retornar device mock
	device := h.getMockDevice(deviceID)

	// ───────────────────────────────────────────────────────────────
	// 3. RESPONDER
	// ───────────────────────────────────────────────────────────────
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"data":     device,
		"metadata": h.buildMetadata(c),
	})
}

// ListDevices returns a paginated list of all devices.
//
// Query Parameters:
//   - limit (default: 10)
//   - offset (default: 0)
//
// Response (200 OK):
//
//	{
//	  "success": true,
//	  "data": {
//	    "devices": [...],
//	    "total": 50,
//	    "limit": 10,
//	    "offset": 0
//	  }
//	}
func (h *DevicesHandler) ListDevices(c *gin.Context) {
	// TODO: Parsear query params y obtener de DynamoDB

	// Por ahora, retornar lista mock
	devices := h.getMockDevices()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"devices": devices,
			"total":   len(devices),
			"limit":   10,
			"offset":  0,
		},
		"metadata": h.buildMetadata(c),
	})
}

// UpdateDevice updates device information.
// This endpoint is not yet implemented.
func (h *DevicesHandler) UpdateDevice(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Device update not yet implemented",
		},
		"metadata": h.buildMetadata(c),
	})
}

// DeleteDevice removes a device from the system.
// This endpoint is not yet implemented.
func (h *DevicesHandler) DeleteDevice(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Device deletion not yet implemented",
		},
		"metadata": h.buildMetadata(c),
	})
}

// ═══════════════════════════════════════════════════════════════════
//                     HELPER FUNCTIONS
// ═══════════════════════════════════════════════════════════════════

// generateMockCertificate genera un certificado IoT mock.
func (h *DevicesHandler) generateMockCertificate(deviceID string) string {
	return "-----BEGIN CERTIFICATE-----\nMOCK_CERTIFICATE_FOR_" + deviceID + "\n-----END CERTIFICATE-----"
}

// buildMetadata construye el objeto metadata estándar.
func (h *DevicesHandler) buildMetadata(c *gin.Context) gin.H {
	return gin.H{
		"timestamp":  time.Now(),
		"request_id": c.GetString("request_id"),
		"service":    h.config.Server.ServiceName,
		"version":    h.config.Server.Version,
	}
}

// getMockDevice retorna un device mock para testing.
func (h *DevicesHandler) getMockDevice(deviceID string) *models.Device {
	lastSeen := time.Now().Add(-2 * time.Minute)

	return &models.Device{
		DeviceID:   deviceID,
		DeviceType: "smart_bin_v1",
		Status:     models.DeviceStatusActive,
		Location: &models.Location{
			Building:  "Edificio A",
			Floor:     2,
			Area:      "Cafeteria",
			Latitude:  4.6097,
			Longitude: -74.0817,
		},
		BinType:   "recyclable",
		LastSeen:  &lastSeen,
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-2 * time.Minute),
	}
}

// getMockDevices retorna lista de devices mock.
func (h *DevicesHandler) getMockDevices() []*models.Device {
	return []*models.Device{
		{
			DeviceID:   "smart-bin-001",
			DeviceType: "smart_bin_v1",
			Status:     models.DeviceStatusActive,
			BinType:    "recyclable",
			CreatedAt:  time.Now().Add(-48 * time.Hour),
		},
		{
			DeviceID:   "smart-bin-002",
			DeviceType: "smart_bin_v1",
			Status:     models.DeviceStatusActive,
			BinType:    "organic",
			CreatedAt:  time.Now().Add(-36 * time.Hour),
		},
		{
			DeviceID:   "smart-bin-003",
			DeviceType: "smart_bin_v2",
			Status:     models.DeviceStatusMaintenance,
			BinType:    "general",
			CreatedAt:  time.Now().Add(-12 * time.Hour),
		},
	}
}
