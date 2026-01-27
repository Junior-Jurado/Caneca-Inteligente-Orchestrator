// Package handlers provides HTTP request handlers for the Smart Bin Orchestrator API.
package handlers

import (
	"net/http"
	"time"

	"github.com/Junior_Jurado/Caneca-Inteligente-Orchestrator/internal/config"
	"github.com/Junior_Jurado/Caneca-Inteligente-Orchestrator/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// WebhooksHandler maneja los endpoints de webhooks y callbacks.
type WebhooksHandler struct {
	config *config.Config
}

// NewWebhooksHandler crea una nueva instancia de WebhooksHandler.
func NewWebhooksHandler(cfg *config.Config) *WebhooksHandler {
	return &WebhooksHandler{
		config: cfg,
	}
}

// ClassificationCallback receives callbacks from the Classifier Service when classification completes.
func (h *WebhooksHandler) ClassificationCallback(c *gin.Context) {
	var payload struct {
		JobID          string                 `json:"job_id" binding:"required"`
		Status         string                 `json:"status" binding:"required"`
		Classification *models.Classification `json:"classification"`
		Error          string                 `json:"error,omitempty"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Error().
			Err(err).
			Str("request_id", c.GetString("request_id")).
			Msg("Failed to parse classification callback")

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "Invalid callback payload",
			},
			"metadata": h.buildMetadata(c),
		})
		return
	}

	log.Info().
		Str("job_id", payload.JobID).
		Str("status", payload.Status).
		Str("request_id", c.GetString("request_id")).
		Msg("Received classification callback")

	if payload.Status == "completed" && payload.Classification != nil {
		log.Info().
			Str("job_id", payload.JobID).
			Str("label", payload.Classification.Label).
			Float64("confidence", payload.Classification.Confidence).
			Msg("Classification successful, would call Decision Service")
	}

	if payload.Status == "failed" {
		log.Error().
			Str("job_id", payload.JobID).
			Str("error", payload.Error).
			Msg("Classification failed")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"job_id":       payload.JobID,
			"received":     true,
			"processed_at": time.Now(),
		},
		"metadata": h.buildMetadata(c),
	})
}

// DeviceEventCallback receives events from IoT devices.
func (h *WebhooksHandler) DeviceEventCallback(c *gin.Context) {
	var event struct {
		EventType string                 `json:"event_type" binding:"required"`
		DeviceID  string                 `json:"device_id" binding:"required"`
		Timestamp string                 `json:"timestamp" binding:"required"`
		Data      map[string]interface{} `json:"data,omitempty"`
	}

	if err := c.ShouldBindJSON(&event); err != nil {
		log.Error().
			Err(err).
			Str("request_id", c.GetString("request_id")).
			Msg("Failed to parse device event")

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "Invalid event payload",
			},
			"metadata": h.buildMetadata(c),
		})
		return
	}

	log.Info().
		Str("device_id", event.DeviceID).
		Str("event_type", event.EventType).
		Str("request_id", c.GetString("request_id")).
		Msg("Received device event")

	switch event.EventType {
	case "image_captured":
		log.Info().
			Str("device_id", event.DeviceID).
			Msg("Image captured event - would create new job")

	case "device_status":
		log.Info().
			Str("device_id", event.DeviceID).
			Interface("data", event.Data).
			Msg("Device status update")

	case "error":
		log.Error().
			Str("device_id", event.DeviceID).
			Interface("data", event.Data).
			Msg("Device error reported")

	default:
		log.Warn().
			Str("device_id", event.DeviceID).
			Str("event_type", event.EventType).
			Msg("Unknown event type")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"event_id":     "evt_" + c.GetString("request_id"),
			"received":     true,
			"processed_at": time.Now(),
		},
		"metadata": h.buildMetadata(c),
	})
}

func (h *WebhooksHandler) buildMetadata(c *gin.Context) gin.H {
	return gin.H{
		"timestamp":  time.Now(),
		"request_id": c.GetString("request_id"),
		"service":    h.config.Server.ServiceName,
		"version":    h.config.Server.Version,
	}
}
