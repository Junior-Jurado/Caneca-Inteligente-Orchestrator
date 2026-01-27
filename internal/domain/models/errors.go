package models

import (
	"errors"
)

/*
╔════════════════════════════════════════════════════════════════╗
║                                                                ║
║  ERRORS.GO - ERRORES DEL DOMINIO                               ║
║                                                                ║
║  Define todos los errores específicos del dominio para         ║
║  tener mejor control y mensajes consistentes.                  ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝
*/

// ═══════════════════════════════════════════════════════════════════
//                     ERRORES DE JOB
// ═══════════════════════════════════════════════════════════════════

var (
	// ErrJobNotFound - Job no encontrado.
	ErrJobNotFound = errors.New("job not found")

	// ErrInvalidJobID - Job ID inválido.
	ErrInvalidJobID = errors.New("invalid job ID")

	// ErrInvalidStatus - Status inválido.
	ErrInvalidStatus = errors.New("invalid status")

	// ErrInvalidTransition - Transición de estado inválida.
	ErrInvalidTransition = errors.New("invalid status transition")

	// ErrJobAlreadyCompleted - Job ya fue completado.
	ErrJobAlreadyCompleted = errors.New("job already completed")

	// ErrJobAlreadyFailed - Job ya falló.
	ErrJobAlreadyFailed = errors.New("job already failed")
)

// ═══════════════════════════════════════════════════════════════════
//                     ERRORES DE DEVICE
// ═══════════════════════════════════════════════════════════════════

var (
	// ErrDeviceNotFound - Device no encontrado.
	ErrDeviceNotFound = errors.New("device not found")

	// ErrInvalidDeviceID - Device ID inválido.
	ErrInvalidDeviceID = errors.New("invalid device ID")

	// ErrInvalidDeviceType - Device type inválido.
	ErrInvalidDeviceType = errors.New("invalid device type")

	// ErrDeviceAlreadyExists - Device ya existe.
	ErrDeviceAlreadyExists = errors.New("device already exists")

	// ErrDeviceInactive - Device está inactivo.
	ErrDeviceInactive = errors.New("device is inactive")

	// ErrDeviceNotOnline - Device no está conectado.
	ErrDeviceNotOnline = errors.New("device not online")
)

// ═══════════════════════════════════════════════════════════════════
//                     ERRORES DE CLASSIFICATION
// ═══════════════════════════════════════════════════════════════════

var (
	// ErrInvalidLabel - Label de clasificación inválido.
	ErrInvalidLabel = errors.New("invalid classification label")

	// ErrInvalidConfidence - Confidence value inválido (debe estar entre 0 y 1).
	ErrInvalidConfidence = errors.New("invalid confidence value")

	// ErrInvalidModelVersion - Model version inválido.
	ErrInvalidModelVersion = errors.New("invalid model version")

	// ErrLowConfidence - Confianza muy baja para tomar decisión.
	ErrLowConfidence = errors.New("classification confidence too low")

	// ErrClassificationFailed - Clasificación falló.
	ErrClassificationFailed = errors.New("classification failed")
)

// ═══════════════════════════════════════════════════════════════════
//                     ERRORES DE DECISION
// ═══════════════════════════════════════════════════════════════════

var (
	// ErrInvalidAction - Action de decisión inválida.
	ErrInvalidAction = errors.New("invalid decision action")

	// ErrInvalidBinCompartment - Bin compartment inválido.
	ErrInvalidBinCompartment = errors.New("invalid bin compartment")

	// ErrInvalidMessage - Message vacío.
	ErrInvalidMessage = errors.New("decision message cannot be empty")

	// ErrDecisionFailed - Decisión falló.
	ErrDecisionFailed = errors.New("decision failed")
)

// ═══════════════════════════════════════════════════════════════════
//                     ERRORES DE AWS
// ═══════════════════════════════════════════════════════════════════

var (
	// ErrDynamoDBOperation - Error en operación de DynamoDB.
	ErrDynamoDBOperation = errors.New("dynamodb operation failed")

	// ErrS3Operation - Error en operación de S3.
	ErrS3Operation = errors.New("s3 operation failed")

	// ErrSQSOperation - Error en operación de SQS.
	ErrSQSOperation = errors.New("sqs operation failed")

	// ErrIoTOperation - Error en operación de IoT Core.
	ErrIoTOperation = errors.New("iot operation failed")

	// ErrPresignedURLExpired - URL prefirmada expiró.
	ErrPresignedURLExpired = errors.New("presigned URL expired")
)

// ═══════════════════════════════════════════════════════════════════
//                     ERRORES DE SERVICIOS EXTERNOS
// ═══════════════════════════════════════════════════════════════════

var (
	// ErrClassifierServiceUnavailable - Classifier service no disponible.
	ErrClassifierServiceUnavailable = errors.New("classifier service unavailable")

	// ErrDecisionServiceUnavailable - Decision service no disponible.
	ErrDecisionServiceUnavailable = errors.New("decision service unavailable")

	// ErrServiceTimeout - Timeout llamando a servicio externo.
	ErrServiceTimeout = errors.New("service timeout")

	// ErrServiceError - Error en servicio externo.
	ErrServiceError = errors.New("external service error")
)

// ═══════════════════════════════════════════════════════════════════
//                     ERRORES DE VALIDACIÓN
// ═══════════════════════════════════════════════════════════════════

var (
	// ErrInvalidInput - Input inválido.
	ErrInvalidInput = errors.New("invalid input")

	// ErrMissingRequiredField - Campo obligatorio faltante.
	ErrMissingRequiredField = errors.New("missing required field")

	// ErrInvalidTimestamp - Timestamp inválido.
	ErrInvalidTimestamp = errors.New("invalid timestamp")

	// ErrInvalidURL - URL inválida.
	ErrInvalidURL = errors.New("invalid URL")
)

// ═══════════════════════════════════════════════════════════════════
//                     ERRORES GENERALES
// ═══════════════════════════════════════════════════════════════════

var (
	// ErrInternalServer - Error interno del servidor.
	ErrInternalServer = errors.New("internal server error")

	// ErrNotImplemented - Funcionalidad no implementada.
	ErrNotImplemented = errors.New("not implemented")

	// ErrUnauthorized - No autorizado.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden - Prohibido.
	ErrForbidden = errors.New("forbidden")

	// ErrRateLimitExceeded - Rate limit excedido.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)
