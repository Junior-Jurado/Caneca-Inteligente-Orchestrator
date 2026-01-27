// Package router configures the HTTP routes and middleware for the Smart Bin Orchestrator API.
// It sets up health checks, API v1 endpoints, and global middleware including CORS, logging, and security.
package router

import (
	"time"

	"github.com/Junior_Jurado/Caneca-Inteligente-Orchestrator/internal/api/handlers"
	"github.com/Junior_Jurado/Caneca-Inteligente-Orchestrator/internal/api/middleware"
	"github.com/Junior_Jurado/Caneca-Inteligente-Orchestrator/internal/config"
	"github.com/gin-gonic/gin"
)

// NewRouter crea y configura el router HTTP principal.
func NewRouter(cfg *config.Config) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(middleware.SecurityHeaders())

	healthHandler := handlers.NewHealthHandler(cfg)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)
	router.GET("/metrics", healthHandler.Metrics)

	v1 := router.Group("/api/v1")

	jobsHandler := handlers.NewJobsHandler(cfg)
	jobs := v1.Group("/jobs")
	jobs.POST("", jobsHandler.CreateJob)
	jobs.GET("/:job_id", jobsHandler.GetJob)
	jobs.GET("", jobsHandler.ListJobs)
	jobs.PATCH("/:job_id", jobsHandler.UpdateJob)
	jobs.DELETE("/:job_id", jobsHandler.DeleteJob)

	devicesHandler := handlers.NewDevicesHandler(cfg)
	devices := v1.Group("/devices")
	devices.POST("/register", devicesHandler.RegisterDevice)
	devices.GET("/:device_id", devicesHandler.GetDevice)
	devices.GET("", devicesHandler.ListDevices)
	devices.PATCH("/:device_id", devicesHandler.UpdateDevice)
	devices.DELETE("/:device_id", devicesHandler.DeleteDevice)

	webhooksHandler := handlers.NewWebhooksHandler(cfg)
	webhooks := v1.Group("/webhooks")
	webhooks.POST("/classification", webhooksHandler.ClassificationCallback)
	webhooks.POST("/device-event", webhooksHandler.DeviceEventCallback)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Endpoint not found",
			},
			"metadata": gin.H{
				"timestamp":  time.Now(),
				"request_id": c.GetString("request_id"),
				"service":    cfg.Server.ServiceName,
			},
		})
	})

	return router
}
