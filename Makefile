# =============================================================================
# Makefile - Smart Bin Orchestrator
# =============================================================================
# Comandos disponibles para desarrollo, testing, linting y deployment
# =============================================================================

.PHONY: help setup build run test lint security clean docker all

# Variables
APP_NAME=orchestrator
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"
GOLANGCI_LINT_VERSION=v1.55.2

# Colores para output
CYAN=\033[0;36m
GREEN=\033[0;32m
RED=\033[0;31m
YELLOW=\033[1;33m
NC=\033[0m # No Color

# =============================================================================
# HELP
# =============================================================================
help: ## Muestra esta ayuda
	@echo "$(CYAN)Smart Bin Orchestrator - Makefile$(NC)"
	@echo ""
	@echo "$(GREEN)Comandos disponibles:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-20s$(NC) %s\n", $$1, $$2}'

# =============================================================================
# SETUP Y DEPENDENCIES
# =============================================================================
setup: ## Instala todas las dependencias necesarias
	@echo "$(CYAN)📦 Instalando dependencias...$(NC)"
	go mod download
	go mod verify
	@echo "$(CYAN)🔧 Instalando herramientas...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/sonatype-nexus-community/nancy@latest
	@echo "$(GREEN)✓ Setup completado$(NC)"

deps-update: ## Actualiza todas las dependencias
	@echo "$(CYAN)⬆️  Actualizando dependencias...$(NC)"
	go get -u ./...
	go mod tidy
	@echo "$(GREEN)✓ Dependencias actualizadas$(NC)"

# =============================================================================
# BUILD
# =============================================================================
build: ## Compila el binario
	@echo "$(CYAN)🔨 Compilando...$(NC)"
	go build $(LDFLAGS) -o bin/$(APP_NAME) cmd/server/main.go
	@echo "$(GREEN)✓ Binario compilado: bin/$(APP_NAME)$(NC)"

build-all: ## Compila para todas las plataformas
	@echo "$(CYAN)🔨 Compilando para múltiples plataformas...$(NC)"
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux-amd64 cmd/server/main.go
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-amd64 cmd/server/main.go
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-windows-amd64.exe cmd/server/main.go
	@echo "$(GREEN)✓ Binarios compilados$(NC)"

# =============================================================================
# RUN
# =============================================================================
run: ## Ejecuta la aplicación
	@echo "$(CYAN)🚀 Iniciando aplicación...$(NC)"
	go run cmd/server/main.go

dev: ## Ejecuta con hot reload (requiere air)
	@echo "$(CYAN)🔥 Iniciando con hot reload...$(NC)"
	air

# =============================================================================
# TESTING
# =============================================================================
test: ## Ejecuta los tests
	@echo "$(CYAN)🧪 Ejecutando tests...$(NC)"
	go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✓ Tests completados$(NC)"

test-coverage: test ## Genera reporte de cobertura HTML
	@echo "$(CYAN)📊 Generando reporte de cobertura...$(NC)"
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Reporte generado: coverage.html$(NC)"

test-coverage-text: test ## Muestra cobertura en terminal
	@echo "$(CYAN)📊 Cobertura de tests:$(NC)"
	@go tool cover -func=coverage.out | tail -1

test-integration: ## Ejecuta tests de integración
	@echo "$(CYAN)🧪 Ejecutando tests de integración...$(NC)"
	go test -v -tags=integration ./test/integration/...

test-e2e: ## Ejecuta tests end-to-end
	@echo "$(CYAN)🧪 Ejecutando tests E2E...$(NC)"
	go test -v -tags=e2e ./test/e2e/...

# =============================================================================
# LINTING Y CALIDAD DE CÓDIGO
# =============================================================================
lint: ## Ejecuta golangci-lint
	@echo "$(CYAN)🔍 Ejecutando linter...$(NC)"
	golangci-lint run --config=.golangci.yml --timeout=5m
	@echo "$(GREEN)✓ Linting completado$(NC)"

lint-fix: ## Ejecuta linter y arregla problemas automáticamente
	@echo "$(CYAN)🔧 Ejecutando linter con auto-fix...$(NC)"
	golangci-lint run --config=.golangci.yml --fix --timeout=5m
	@echo "$(GREEN)✓ Problemas arreglados$(NC)"

fmt: ## Formatea el código
	@echo "$(CYAN)💅 Formateando código...$(NC)"
	go fmt ./...
	goimports -w .
	@echo "$(GREEN)✓ Código formateado$(NC)"

vet: ## Ejecuta go vet
	@echo "$(CYAN)🔍 Ejecutando go vet...$(NC)"
	go vet ./...
	@echo "$(GREEN)✓ Vet completado$(NC)"

# =============================================================================
# SEGURIDAD
# =============================================================================
security: ## Ejecuta escaneo de seguridad completo
	@echo "$(CYAN)🔒 Ejecutando escaneo de seguridad...$(NC)"
	@$(MAKE) security-gosec
	@$(MAKE) security-nancy
	@echo "$(GREEN)✓ Escaneo de seguridad completado$(NC)"

security-gosec: ## Ejecuta gosec (Go Security Scanner)
	@echo "$(CYAN)🔒 Ejecutando gosec...$(NC)"
	gosec -fmt=json -out=gosec-report.json -stdout -verbose=text ./...
	@echo "$(GREEN)✓ Gosec completado - Ver: gosec-report.json$(NC)"

security-nancy: ## Ejecuta nancy (dependency vulnerability scanner)
	@echo "$(CYAN)🔒 Ejecutando nancy...$(NC)"
	go list -json -deps ./... | nancy sleuth
	@echo "$(GREEN)✓ Nancy completado$(NC)"

security-trivy: ## Escanea la imagen Docker con Trivy
	@echo "$(CYAN)🔒 Ejecutando Trivy en imagen Docker...$(NC)"
	trivy image --severity HIGH,CRITICAL $(APP_NAME):latest
	@echo "$(GREEN)✓ Trivy completado$(NC)"

# =============================================================================
# DOCKER
# =============================================================================
docker-build: ## Construye imagen Docker
	@echo "$(CYAN)🐳 Construyendo imagen Docker...$(NC)"
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .
	@echo "$(GREEN)✓ Imagen construida: $(APP_NAME):$(VERSION)$(NC)"

docker-run: ## Ejecuta contenedor Docker
	@echo "$(CYAN)🐳 Ejecutando contenedor...$(NC)"
	docker run -p 8080:8080 --env-file .env $(APP_NAME):latest

docker-scan: docker-build ## Escanea la imagen Docker con Trivy
	@$(MAKE) security-trivy

# =============================================================================
# LIMPIEZA
# =============================================================================
clean: ## Limpia archivos generados
	@echo "$(CYAN)🧹 Limpiando...$(NC)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f gosec-report.json
	go clean -cache
	@echo "$(GREEN)✓ Limpieza completada$(NC)"

# =============================================================================
# CI/CD
# =============================================================================
ci: lint security test build ## Ejecuta CI pipeline localmente
	@echo "$(GREEN)✓ CI checks pasaron$(NC)"

pre-commit: fmt lint test ## Ejecuta checks antes de commit
	@echo "$(GREEN)✓ Pre-commit checks pasaron$(NC)"

# =============================================================================
# ANÁLISIS
# =============================================================================
complexity: ## Muestra complejidad del código
	@echo "$(CYAN)📊 Analizando complejidad...$(NC)"
	gocyclo -over 15 .

duplicates: ## Detecta código duplicado
	@echo "$(CYAN)📊 Detectando código duplicado...$(NC)"
	dupl -threshold 100 ./...

# =============================================================================
# UTILIDADES
# =============================================================================
mod-tidy: ## Limpia go.mod y go.sum
	@echo "$(CYAN)🔧 Limpiando módulos...$(NC)"
	go mod tidy
	@echo "$(GREEN)✓ Módulos limpios$(NC)"

mod-verify: ## Verifica integridad de dependencias
	@echo "$(CYAN)🔍 Verificando módulos...$(NC)"
	go mod verify
	@echo "$(GREEN)✓ Módulos verificados$(NC))"

generate: ## Genera código (mocks, etc)
	@echo "$(CYAN)🔧 Generando código...$(NC)"
	go generate ./...
	@echo "$(GREEN)✓ Código generado$(NC)"

# =============================================================================
# ALL
# =============================================================================
all: clean fmt lint security test build ## Ejecuta todo el pipeline
	@echo "$(GREEN)✓✓✓ Pipeline completo ejecutado$(NC)"

# Default target
.DEFAULT_GOAL := help