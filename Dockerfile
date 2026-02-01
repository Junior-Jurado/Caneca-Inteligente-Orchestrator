# =============================================================================
# MULTI-STAGE DOCKERFILE - SMART BIN ORCHESTRATOR
# =============================================================================
# Optimizado para el proyecto actual con estructura:
# - cmd/server/main.go (entry point)
# - internal/ (código privado)
# - go.mod/go.sum (dependencias)
# =============================================================================

# -----------------------------------------------------------------------------
# STAGE 1: BUILD
# -----------------------------------------------------------------------------
FROM golang:1.24.3-alpine AS builder

LABEL maintainer="Smart Bin Team"
LABEL description="Smart Bin Orchestrator Service - Build Stage"

# Instalar dependencias de compilación
RUN apk add --no-cache git ca-certificates tzdata

# Configurar directorio de trabajo
WORKDIR /build

# Copiar archivos de dependencias primero (para cache de Docker)
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download
RUN go mod verify

# Copiar TODO el código fuente
# Necesitamos copiar internal/, cmd/, y cualquier otro directorio
COPY . .

# Compilar el binario
# IMPORTANTE: El path del main es cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s -X main.Version=1.0.0 -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o orchestrator \
    ./cmd/server/main.go

# Verificar que el binario existe
RUN ls -lh /build/orchestrator && echo "Binary compiled successfully"

# -----------------------------------------------------------------------------
# STAGE 2: RUNTIME
# -----------------------------------------------------------------------------
FROM alpine:3.19

LABEL maintainer="Smart Bin Team"
LABEL version="1.0.0"
LABEL description="Smart Bin Orchestrator Service - Production Runtime"

# Instalar certificados SSL y timezone data
RUN apk --no-cache add ca-certificates tzdata wget

# Crear usuario no-root para ejecutar la aplicación
RUN addgroup -g 1000 orchestrator && \
    adduser -D -u 1000 -G orchestrator orchestrator

# Configurar directorio de trabajo
WORKDIR /app

# Copiar el binario compilado desde el stage anterior
COPY --from=builder /build/orchestrator .

# Cambiar ownership
RUN chown -R orchestrator:orchestrator /app

# Cambiar a usuario no-root
USER orchestrator

# Exponer el puerto
EXPOSE 8080

# Health check (usa wget porque alpine no tiene curl por defecto)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Comando para ejecutar la aplicación
ENTRYPOINT ["/app/orchestrator"]

# =============================================================================
# NOTAS DE USO:
# =============================================================================
# 
# Build local:
#   docker build -t smartbin-orchestrator:latest .
#
# Run local (básico):
#   docker run -p 8080:8080 smartbin-orchestrator:latest
#
# Run local (con variables de entorno):
#   docker run -p 8080:8080 \
#     -e APP_ENV=dev \
#     -e AWS_REGION=us-east-1 \
#     -e PORT=8080 \
#     -e LOG_LEVEL=debug \
#     smartbin-orchestrator:latest
#
# Build para AWS ECR:
#   docker build -t 940482438767.dkr.ecr.us-east-1.amazonaws.com/smartbin-dev-service:orchestrator .
#
# =============================================================================