package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Junior_Jurado/Caneca-Inteligente-Orchestrator/internal/config"
	"github.com/Junior_Jurado/Caneca-Inteligente-Orchestrator/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

/*
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║  JOBS.GO - HANDLER DE JOBS DE CLASIFICACIÓN                   ║
║                                                               ║
║  Endpoints para gestionar jobs de clasificación de residuos:  ║
║  - POST   /api/v1/jobs           - Crear job                  ║
║  - GET    /api/v1/jobs/:job_id   - Obtener job                ║
║  - GET    /api/v1/jobs           - Listar jobs                ║
║  - PATCH  /api/v1/jobs/:job_id   - Actualizar job             ║
║  - DELETE /api/v1/jobs/:job_id   - Eliminar job               ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
*/

// JobsHandler maneja los endpoints relacionados con jobs.
type JobsHandler struct {
	config *config.Config
	// TODO: Agregar dependencies:
	// jobRepository    ports.JobRepository
	// s3Presigner      ports.S3Presigner
	// classifierClient ports.ClassifierClient
}

// NewJobsHandler crea una nueva instancia de JobsHandler.
func NewJobsHandler(cfg *config.Config) *JobsHandler {
	return &JobsHandler{
		config: cfg,
		// TODO: Inyectar dependencies cuando estén listas
	}
}

// ENDPOINT: POST /api/v1/jobs.
func (h *JobsHandler) CreateJob(c *gin.Context) {
	// ───────────────────────────────────────────────────────────────
	// 1. PARSEAR Y VALIDAR REQUEST
	// ───────────────────────────────────────────────────────────────
	var req models.CreateJobRequest

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
	// 2. GENERAR JOB ID ÚNICO
	// ───────────────────────────────────────────────────────────────
	jobID := h.generateJobID()

	// ───────────────────────────────────────────────────────────────
	// 3. CREAR OBJETO JOB
	// ───────────────────────────────────────────────────────────────
	job := &models.Job{
		JobID:     jobID,
		DeviceID:  req.DeviceID,
		Status:    models.JobStatusPending,
		ImageKey:  h.generateImageKey(req.DeviceID, jobID),
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// ───────────────────────────────────────────────────────────────
	// 4. GENERAR URL PREFIRMADA DE S3
	// ───────────────────────────────────────────────────────────────
	// TODO: Usar S3Presigner real
	// uploadURL, err := h.s3Presigner.GeneratePresignedPutURL(
	//     c.Request.Context(),
	//     h.config.AWS.S3.BucketImages,
	//     job.ImageKey,
	//     h.config.AWS.S3.PresignedURLExpiry,
	// )

	// Por ahora, generar URL mock
	uploadURL := h.generateMockS3URL(job.ImageKey)
	uploadExpiresAt := time.Now().Add(h.config.AWS.S3.PresignedURLExpiry)

	response := models.CreateJobResponse{
		JobID:           job.JobID,
		DeviceID:        job.DeviceID,
		Status:          job.Status,
		UploadURL:       uploadURL,
		UploadExpiresAt: uploadExpiresAt,
		CreatedAt:       job.CreatedAt,
	}

	// ───────────────────────────────────────────────────────────────
	// 7. RESPONDER
	// ───────────────────────────────────────────────────────────────
	c.JSON(http.StatusCreated, gin.H{
		"success":  true,
		"data":     response,
		"metadata": h.buildMetadata(c),
	})
}

// ═══════════════════════════════════════════════════════════════════
// ENDPOINT: GET /api/v1/jobs/:job_id
// ═══════════════════════════════════════════════════════════════════
// Obtiene un job por su ID
//
// Response (200 OK):
//
//	{
//	  "success": true,
//	  "data": {
//	    "job_id": "job_abc123",
//	    "device_id": "smart-bin-001",
//	    "status": "completed",
//	    "classification": {
//	      "label": "plastic_bottle",
//	      "confidence": 0.94
//	    },
//	    "decision": {
//	      "action": "accept",
//	      "bin_compartment": "recyclable"
//	    }
//	  }
//	}
func (h *JobsHandler) GetJob(c *gin.Context) {
	// ───────────────────────────────────────────────────────────────
	// 1. OBTENER JOB ID DEL PATH
	// ───────────────────────────────────────────────────────────────
	jobID := c.Param("job_id")

	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "Job ID is required",
			},
			"metadata": h.buildMetadata(c),
		})
		return
	}

	job := h.getMockJob(jobID)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"data":     job,
		"metadata": h.buildMetadata(c),
	})
}

/* ENDPOINT: GET /api/v1/jobs. */
func (h *JobsHandler) ListJobs(c *gin.Context) {
	// 1. PARSEAR QUERY PARAMETERS
	var req models.ListJobsRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "Invalid query parameters",
				"details": gin.H{
					"error": err.Error(),
				},
			},
			"metadata": h.buildMetadata(c),
		})
		return
	}

	// Valores por defecto
	if req.Limit == 0 {
		req.Limit = 10
	}

	// ───────────────────────────────────────────────────────────────
	// 2. OBTENER JOBS DE DYNAMODB
	// ───────────────────────────────────────────────────────────────
	// TODO: Consultar DynamoDB con filtros
	// jobs, total, err := h.jobRepository.List(c.Request.Context(), &req)

	// Por ahora, retornar jobs mock
	jobs := h.getMockJobs()
	total := len(jobs)

	// ───────────────────────────────────────────────────────────────
	// 3. CONSTRUIR RESPONSE
	// ───────────────────────────────────────────────────────────────
	response := models.ListJobsResponse{
		Jobs:   jobs,
		Total:  total,
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"data":     response,
		"metadata": h.buildMetadata(c),
	})
}

// ═══════════════════════════════════════════════════════════════════
// ENDPOINT: PATCH /api/v1/jobs/:job_id
// ═══════════════════════════════════════════════════════════════════
// Actualiza un job (usado internamente por callbacks)

func (h *JobsHandler) UpdateJob(c *gin.Context) {
	// jobID := c.Param("job_id")

	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Job update not yet implemented",
		},
		"metadata": h.buildMetadata(c),
	})
}

// ═══════════════════════════════════════════════════════════════════
// ENDPOINT: DELETE /api/v1/jobs/:job_id
// ═══════════════════════════════════════════════════════════════════
// Elimina un job

func (h *JobsHandler) DeleteJob(c *gin.Context) {
	// jobID := c.Param("job_id")

	// TODO: Implementar eliminación de job
	// Por ahora, retornar 501 Not Implemented

	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Job deletion not yet implemented",
		},
		"metadata": h.buildMetadata(c),
	})
}

// ═══════════════════════════════════════════════════════════════════
//                     HELPER FUNCTIONS
// ═══════════════════════════════════════════════════════════════════

// generateJobID genera un ID único para un job.
func (h *JobsHandler) generateJobID() string {
	return "job_" + uuid.New().String()[:8]
}

// generateImageKey genera la key S3 para la imagen del job.
func (h *JobsHandler) generateImageKey(deviceID, jobID string) string {
	return fmt.Sprintf("uploads/%s/%s.jpg", deviceID, jobID)
}

// generateMockS3URL genera una URL S3 mock (temporal).
func (h *JobsHandler) generateMockS3URL(imageKey string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s?X-Amz-Algorithm=...",
		h.config.AWS.S3.BucketImages,
		h.config.AWS.Region,
		imageKey,
	)
}

// buildMetadata construye el objeto metadata estándar.
func (h *JobsHandler) buildMetadata(c *gin.Context) gin.H {
	return gin.H{
		"timestamp":  time.Now(),
		"request_id": c.GetString("request_id"),
		"service":    h.config.Server.ServiceName,
		"version":    h.config.Server.Version,
	}
}

// getMockJob retorna un job mock para testing.
func (h *JobsHandler) getMockJob(jobID string) *models.Job {
	completedAt := time.Now().Add(-1 * time.Minute)

	return &models.Job{
		JobID:       jobID,
		DeviceID:    "smart-bin-001",
		Status:      models.JobStatusCompleted,
		ImageKey:    fmt.Sprintf("uploads/smart-bin-001/%s.jpg", jobID),
		CreatedAt:   time.Now().Add(-5 * time.Minute),
		UpdatedAt:   time.Now().Add(-1 * time.Minute),
		CompletedAt: &completedAt,
		Classification: &models.Classification{
			Label:        "plastic_bottle",
			Confidence:   0.94,
			ModelVersion: "v1.2.3",
			Alternatives: []models.Alternative{
				{Label: "glass_bottle", Confidence: 0.03},
				{Label: "aluminum_can", Confidence: 0.02},
			},
			ProcessingTime: 342,
		},
		Decision: &models.Decision{
			Action:                 "accept",
			BinCompartment:         "recyclable",
			Message:                "Item classified correctly as recyclable plastic",
			ConfidenceThresholdMet: true,
			RuleApplied:            "recyclable_plastics_high_confidence",
		},
	}
}

// getMockJobs retorna lista de jobs mock para testing.
func (h *JobsHandler) getMockJobs() []*models.Job {
	return []*models.Job{
		{
			JobID:     "job_abc123",
			DeviceID:  "smart-bin-001",
			Status:    models.JobStatusCompleted,
			CreatedAt: time.Now().Add(-10 * time.Minute),
		},
		{
			JobID:     "job_def456",
			DeviceID:  "smart-bin-002",
			Status:    models.JobStatusPending,
			CreatedAt: time.Now().Add(-5 * time.Minute),
		},
		{
			JobID:     "job_ghi789",
			DeviceID:  "smart-bin-001",
			Status:    models.JobStatusProcessing,
			CreatedAt: time.Now().Add(-2 * time.Minute),
		},
	}
}
