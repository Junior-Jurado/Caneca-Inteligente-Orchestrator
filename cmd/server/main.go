// Package main is the entry point for the Smart Bin Orchestrator Service.
// It initializes the HTTP server, loads configuration, and manages graceful shutdown.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Junior_Jurado/Caneca-Inteligente-Orchestrator/internal/api/router"
	"github.com/Junior_Jurado/Caneca-Inteligente-Orchestrator/internal/config"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

/*
╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║  MAIN.GO - PUNTO DE ENTRADA DEL SERVICIO ORCHESTRATOR            ║
║                                                                  ║
║  Este archivo es el corazón del servicio. Se encarga de:         ║
║  1. Cargar configuración                                         ║
║  2. Inicializar dependencies                                     ║
║  3. Iniciar servidor HTTP                                        ║
║  4. Manejar shutdown gracefully                                  ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
*/

func main() {
	// ═══════════════════════════════════════════════════════════════
	// PASO 1: CARGAR VARIABLES DE ENTORNO
	// ═══════════════════════════════════════════════════════════════
	// Intenta cargar el archivo .env si existe (para desarrollo local)
	// En producción (ECS, K8s) las variables vienen del sistema
	if err := godotenv.Load(); err != nil {
		// No es un error fatal si no existe .env en producción
		log.Debug().Msg(".env file not found, using environment variables")
	}

	// ═══════════════════════════════════════════════════════════════
	// PASO 2: CONFIGURAR LOGGER
	// ═══════════════════════════════════════════════════════════════
	setupLogger()

	log.Info().Msg("Starting Smart Bin Orchestrator Service")

	// ═══════════════════════════════════════════════════════════════
	// PASO 3: CARGAR CONFIGURACIÓN
	// ═══════════════════════════════════════════════════════════════
	// Carga y valida toda la configuración desde variables de entorno
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	log.Info().
		Str("service", cfg.Server.ServiceName).
		Str("version", cfg.Server.Version).
		Str("environment", cfg.Server.Environment).
		Str("port", cfg.Server.Port).
		Bool("use_localstack", cfg.UseLocalStack()).
		Msg("Configuration loaded successfully")

	// ═══════════════════════════════════════════════════════════════
	// PASO 4: INICIALIZAR dependencies
	// ═══════════════════════════════════════════════════════════════
	// TODO: En pasos futuros aquí inicializaremos:
	//
	// 1. AWS Clients:
	//    - DynamoDB (para guardar jobs y devices)
	//    - S3 (para generar URLs prefirmadas)
	//    - SQS (para publicar mensajes al Classifier)
	//    - IoT Core (para enviar resultados a dispositivos)
	//
	// 2. HTTP Clients:
	//    - Classifier Service client
	//    - Decision Service client
	//
	// 3. Repositories:
	//    - JobRepository (abstracción de DynamoDB)
	//    - DeviceRepository (abstracción de DynamoDB)
	//
	// 4. Domain Services:
	//    - OrchestrationService (lógica de orquestación)
	//    - JobManager
	//    - DeviceManager
	//
	// Por ahora lo dejamos comentado:
	// deps := initializeDependencies(cfg)

	// ═══════════════════════════════════════════════════════════════
	// PASO 5: CONFIGURAR ROUTER HTTP
	// ═══════════════════════════════════════════════════════════════
	// El router maneja todas las rutas HTTP:
	// - Health checks (/health, /ready, /metrics)
	// - Jobs API (/api/v1/jobs)
	// - Devices API (/api/v1/devices)
	// - Webhooks (/api/v1/webhooks)
	httpRouter := router.NewRouter(cfg)

	// ═══════════════════════════════════════════════════════════════
	// PASO 6: CREAR SERVIDOR HTTP
	// ═══════════════════════════════════════════════════════════════
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: httpRouter,

		// Timeouts importantes para producción
		ReadTimeout:  30 * time.Second,  // Tiempo máximo para leer request
		WriteTimeout: 30 * time.Second,  // Tiempo máximo para escribir response
		IdleTimeout:  120 * time.Second, // Tiempo máximo de conexión idle

		// Limite de tamaño de headers (1 MB)
		MaxHeaderBytes: 1 << 20,
	}

	// ═══════════════════════════════════════════════════════════════
	// PASO 7: INICIAR SERVIDOR HTTP (EN GOROUTINE)
	// ═══════════════════════════════════════════════════════════════
	// Lo iniciamos en una goroutine para no bloquear el main thread
	// Así podemos escuchar señales de shutdown en el thread principal
	go func() {
		log.Info().
			Str("port", cfg.Server.Port).
			Str("environment", cfg.Server.Environment).
			Msgf("HTTP server listening on http://localhost:%s", cfg.Server.Port)

		// ListenAndServe bloquea hasta que ocurre un error o shutdown
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
	}()

	// ═══════════════════════════════════════════════════════════════
	// PASO 8: ESPERAR SEÑAL DE TERMINACIÓN
	// ═══════════════════════════════════════════════════════════════
	// Crear canal para recibir señales del sistema operativo
	quit := make(chan os.Signal, 1)

	// Escuchar señales de terminación:
	// - SIGINT: (Ctrl+C) en terminal
	// - SIGTERM: Señal de ECS/K8s cuando hace stop del contenedor
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Bloquear aquí hasta recibir una señal
	sig := <-quit

	log.Info().
		Str("signal", sig.String()).
		Msg("Received shutdown signal")

	// ═══════════════════════════════════════════════════════════════
	// PASO 9: GRACEFUL SHUTDOWN
	// ═══════════════════════════════════════════════════════════════
	// Crear contexto con timeout para el shutdown
	ctx, cancel := context.WithTimeout(
		context.Background(),
		cfg.Server.GracefulShutdownTimeout,
	)
	defer cancel()

	log.Info().
		Dur("timeout", cfg.Server.GracefulShutdownTimeout).
		Msg("Shutting down HTTP server gracefully...")

	// Shutdown graceful:
	// 1. Deja de aceptar nuevas conexiones
	// 2. Espera a que las conexiones activas terminen
	// 3. Si no terminan antes del timeout, las fuerza a cerrar
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	// TODO: Cerrar otras conexiones
	// En pasos futuros aquí cerraremos:
	// - Conexiones de AWS clients
	// - Conexiones de HTTP clients
	// - Conexiones de base de datos (si hay)
	// deps.Close()

	log.Info().Msg("Server stopped gracefully")
}

// setupLogger configures the global logger based on environment.
// Development mode uses colorized console output, while production uses structured JSON.
func setupLogger() {
	// Formato de timestamp en los logs
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Obtener configuración del ambiente
	env := os.Getenv("APP_ENV")
	logLevel := os.Getenv("LOG_LEVEL")

	// ───────────────────────────────────────────────────────────────
	// Configurar nivel de log
	// ───────────────────────────────────────────────────────────────
	switch logLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		// Si no se especifica, usar DEBUG en dev e INFO en prod
		if env == "development" || env == "dev" {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}
	}

	// ───────────────────────────────────────────────────────────────
	// Configurar formato de salida
	// ───────────────────────────────────────────────────────────────
	if env == "development" || env == "dev" {
		// DEVELOPMENT: Logs coloridos en consola
		//
		// Ejemplo de salida:
		// 15:04:05 INF Starting server port=8080
		//
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
			NoColor:    false,
		})
	} else {
		// PRODUCTION: JSON estructurado
		//
		// Ejemplo de salida:
		// {"level":"info","time":"2026-01-20T22:45:00Z","caller":"main.go:50","message":"Starting server","port":"8080"}
		//
		log.Logger = zerolog.New(os.Stderr).
			With().
			Timestamp(). // Agrega timestamp automático
			Caller().    // Agrega archivo: línea del caller
			Logger()
	}

	log.Debug().Msg("Logger configured")
}

/*
═══════════════════════════════════════════════════════════════════
                    PRÓXIMOS PASOS
═══════════════════════════════════════════════════════════════════

Para completar este archivo, necesitaremos implementar:

1. initializeDependencies(cfg) - Función que crea todas las dependencies
   └─ Crea AWS clients
   └─ Crea HTTP clients
   └─ Crea repositories
   └─ Crea domain services
   └─ Los inyecta al router

2. Dependencies struct - Contiene todas las dependencies
   └─ JobRepository
   └─ DeviceRepository
   └─ ClassifierClient
   └─ DecisionClient
   └─ S3Presigner
   └─ SQSPublisher
   └─ IoTPublisher

3. Dependencies.Close() - Cierra todas las conexiones gracefully

Por ahora, main.go está funcional y puede arrancar el servidor,
pero falta implementar la lógica de negocio (siguiente paso).

═══════════════════════════════════════════════════════════════════
*/
