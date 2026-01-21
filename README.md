# 🏗️ Smart Bin Orchestrator - Repositorio Completo

## 📋 Índice
1. [Visión General](#visión-general)
2. [Estructura del Repositorio](#estructura-del-repositorio)
3. [Arquitectura del Servicio](#arquitectura-del-servicio)
4. [Integración con Otros Servicios](#integración-con-otros-servicios)
5. [Setup del Proyecto](#setup-del-proyecto)
6. [Desarrollo](#desarrollo)
7. [CI/CD](#cicd)
8. [Deployment](#deployment)

---

## 🎯 Visión General

### Responsabilidades del Orchestrator

```
┌─────────────────────────────────────────────────────────────────┐
│                     ORCHESTRATOR SERVICE                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. API Gateway / Entry Point                                  │
│     - Recibe peticiones de dispositivos IoT                    │
│     - Expone API REST para dashboard/mobile                    │
│     - Maneja autenticación y autorización                      │
│                                                                 │
│  2. Coordinación de Flujo                                      │
│     - Crea jobs de clasificación                               │
│     - Orquesta llamadas entre servicios                        │
│     - Mantiene estado en DynamoDB                              │
│                                                                 │
│  3. Gestión de Recursos AWS                                    │
│     - URLs prefirmadas S3 (upload de imágenes)                 │
│     - Publicación a SQS (jobs para Classifier)                 │
│     - Comunicación IoT Core (resultados a dispositivos)        │
│                                                                 │
│  4. Integración de Servicios                                   │
│     - Invoca Classifier Service (clasificación ML)             │
│     - Invoca Decision Service (reglas de negocio)              │
│     - Consolida respuestas                                     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Flujo Completo

```
IoT Device → Orchestrator → S3 (upload) → SQS → Classifier → Orchestrator
                ↓                                                    ↓
            DynamoDB ←──────────────────────────────────────────────┘
                ↓
            Decision Service ←─────────────── Orchestrator
                ↓
            IoT Core ←────────────────────────── Orchestrator
                ↓
            Device (resultado)
```

---

## 📁 Estructura del Repositorio

```
smart-bin-orchestrator/
│
├── .github/                          # GitHub Actions workflows
│   ├── workflows/
│   │   ├── ci.yml                   # Tests, lint, build
│   │   ├── cd-dev.yml               # Deploy to dev
│   │   ├── cd-staging.yml           # Deploy to staging
│   │   └── cd-prod.yml              # Deploy to production
│   └── dependabot.yml               # Dependency updates
│
├── cmd/
│   └── server/
│       └── main.go                   # Entry point
│
├── internal/                         # Private application code
│   ├── api/                         # API layer
│   │   ├── handlers/
│   │   │   ├── health.go
│   │   │   ├── jobs.go
│   │   │   ├── devices.go
│   │   │   └── webhooks.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── logger.go
│   │   │   ├── ratelimit.go
│   │   │   └── recovery.go
│   │   └── router/
│   │       └── router.go
│   │
│   ├── domain/                      # Business logic
│   │   ├── models/
│   │   │   ├── job.go
│   │   │   ├── device.go
│   │   │   └── classification.go
│   │   ├── services/
│   │   │   ├── orchestration.go    # Core orchestration logic
│   │   │   ├── job_manager.go
│   │   │   └── device_manager.go
│   │   └── ports/                   # Interfaces
│   │       ├── repositories.go
│   │       └── clients.go
│   │
│   ├── infrastructure/              # External integrations
│   │   ├── aws/
│   │   │   ├── dynamodb/
│   │   │   │   ├── client.go
│   │   │   │   ├── job_repository.go
│   │   │   │   └── device_repository.go
│   │   │   ├── s3/
│   │   │   │   ├── client.go
│   │   │   │   └── presigner.go
│   │   │   ├── sqs/
│   │   │   │   ├── client.go
│   │   │   │   ├── publisher.go
│   │   │   │   └── consumer.go
│   │   │   └── iot/
│   │   │       ├── client.go
│   │   │       └── publisher.go
│   │   │
│   │   ├── http/                    # HTTP clients for other services
│   │   │   ├── classifier_client.go
│   │   │   └── decision_client.go
│   │   │
│   │   └── cache/
│   │       └── redis.go             # Optional: Redis cache
│   │
│   ├── config/                      # Configuration
│   │   ├── config.go
│   │   └── validator.go
│   │
│   └── pkg/                         # Shared utilities
│       ├── errors/
│       │   └── errors.go
│       ├── logger/
│       │   └── logger.go
│       └── validator/
│           └── validator.go
│
├── pkg/                             # Public libraries (if any)
│
├── test/                            # Test files
│   ├── integration/
│   │   ├── jobs_test.go
│   │   └── devices_test.go
│   ├── e2e/
│   │   └── flow_test.go
│   └── fixtures/
│       ├── jobs.json
│       └── devices.json
│
├── scripts/                         # Utility scripts
│   ├── build.sh
│   ├── test.sh
│   ├── migrate.sh
│   └── local-dev.sh
│
├── deployments/                     # Deployment configurations
│   ├── docker/
│   │   ├── Dockerfile
│   │   ├── Dockerfile.dev
│   │   └── .dockerignore
│   ├── kubernetes/                  # K8s manifests (optional)
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── configmap.yaml
│   └── terraform/                   # Service-specific infrastructure
│       ├── main.tf
│       ├── variables.tf
│       └── outputs.tf
│
├── docs/                            # Documentation
│   ├── architecture.md
│   ├── api.md
│   ├── integration.md
│   └── deployment.md
│
├── .env.example                     # Environment variables template
├── .gitignore
├── .golangci.yml                    # Linter configuration
├── docker-compose.yml               # Local development
├── docker-compose.test.yml          # Testing environment
├── go.mod
├── go.sum
├── Makefile                         # Build commands
└── README.md

```

---

## 🏛️ Arquitectura del Servicio

### Arquitectura Hexagonal (Ports & Adapters)

```
┌─────────────────────────────────────────────────────────────────┐
│                        API LAYER (Handlers)                     │
│  HTTP Handlers │ Middleware │ Request/Response │ Validation     │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                      DOMAIN LAYER (Core)                        │
│  Business Logic │ Orchestration │ Models │ Interfaces (Ports)  │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                  INFRASTRUCTURE LAYER (Adapters)                │
│  DynamoDB │ S3 │ SQS │ IoT │ HTTP Clients │ Cache │ Metrics   │
└─────────────────────────────────────────────────────────────────┘
```

### Dependencias entre Capas

```go
// API Layer → Domain Layer
handlers → domain.services

// Domain Layer → Infrastructure (via interfaces)
domain.services → domain.ports (interfaces)
infrastructure  → domain.ports (implementations)
```

---

## 🔗 Integración con Otros Servicios

### 1. Classifier Service

**Comunicación:** HTTP REST + SQS (asíncrono)

```go
// Opción A: Síncrono (para MVP/testing)
POST https://classifier-service.com/api/v1/classify
{
  "image_url": "s3://bucket/path/to/image.jpg",
  "job_id": "job_123"
}

// Opción B: Asíncrono (producción)
// Orchestrator → SQS Queue → Classifier consume
SQS Message: {
  "job_id": "job_123",
  "image_key": "uploads/device-001/job_123.jpg",
  "device_id": "device-001",
  "timestamp": "2026-01-20T..."
}
```

**Interface en Orchestrator:**

```go
type ClassifierClient interface {
    ClassifySync(ctx context.Context, req *ClassifyRequest) (*ClassifyResponse, error)
    PublishJobToQueue(ctx context.Context, job *Job) error
}
```

### 2. Decision Service

**Comunicación:** HTTP REST (síncrono)

```go
POST https://decision-service.com/api/v1/decide
{
  "classification": {
    "label": "plastic_bottle",
    "confidence": 0.94
  },
  "device_info": {
    "device_id": "device-001",
    "bin_type": "recyclable"
  },
  "context": {
    "location": "cafeteria-piso-2",
    "timestamp": "2026-01-20T..."
  }
}

Response:
{
  "action": "accept",
  "bin_compartment": "recyclable",
  "message": "Item classified correctly",
  "confidence_threshold_met": true,
  "rule_applied": "recyclable_plastics"
}
```

**Interface en Orchestrator:**

```go
type DecisionClient interface {
    Decide(ctx context.Context, req *DecisionRequest) (*DecisionResponse, error)
}
```

### 3. Service Discovery

**Opción A: Environment Variables (MVP)**
```env
CLASSIFIER_SERVICE_URL=http://classifier-service:8081
DECISION_SERVICE_URL=http://decision-service:8082
```

**Opción B: AWS Service Discovery (Producción)**
```go
// Use AWS Cloud Map for service discovery
classifierURL, err := serviceDiscovery.GetServiceURL("classifier-service")
```

**Opción C: Load Balancer (Recomendado para AWS)**
```env
# Internal ALB endpoints
CLASSIFIER_SERVICE_URL=http://classifier-internal-alb.smart-bin.local
DECISION_SERVICE_URL=http://decision-internal-alb.smart-bin.local
```

### 4. Circuit Breaker & Retry

```go
// Usar go-resilience o similar
import "github.com/sony/gobreaker"

type ResilientHTTPClient struct {
    client  *http.Client
    breaker *gobreaker.CircuitBreaker
}

func (c *ResilientHTTPClient) Do(req *http.Request) (*http.Response, error) {
    resp, err := c.breaker.Execute(func() (interface{}, error) {
        return c.client.Do(req)
    })
    
    if err != nil {
        return nil, err
    }
    
    return resp.(*http.Response), nil
}
```

---

## ⚙️ Setup del Proyecto

### Paso 1: Crear el Repositorio

```bash
# Crear nuevo repo
mkdir smart-bin-orchestrator
cd smart-bin-orchestrator

# Inicializar Git
git init
git remote add origin https://github.com/tu-org/smart-bin-orchestrator.git

# Inicializar Go module
go mod init github.com/tu-org/smart-bin-orchestrator
```

### Paso 2: Estructura Base

```bash
# Crear estructura de directorios
mkdir -p cmd/server
mkdir -p internal/{api/{handlers,middleware,router},domain/{models,services,ports},infrastructure/{aws/{dynamodb,s3,sqs,iot},http,cache},config,pkg/{errors,logger,validator}}
mkdir -p test/{integration,e2e,fixtures}
mkdir -p scripts deployments/{docker,kubernetes,terraform} docs

# Crear archivos base
touch cmd/server/main.go
touch internal/config/config.go
touch .env.example
touch .gitignore
touch Makefile
touch README.md
```

### Paso 3: Dependencias Core

```bash
# Framework HTTP
go get github.com/gin-gonic/gin@latest

# AWS SDK v2
go get github.com/aws/aws-sdk-go-v2@latest
go get github.com/aws/aws-sdk-go-v2/config@latest
go get github.com/aws/aws-sdk-go-v2/service/dynamodb@latest
go get github.com/aws/aws-sdk-go-v2/service/s3@latest
go get github.com/aws/aws-sdk-go-v2/service/sqs@latest
go get github.com/aws/aws-sdk-go-v2/service/iotdataplane@latest

# Utilities
go get github.com/google/uuid@latest
go get github.com/rs/zerolog@latest
go get github.com/joho/godotenv@latest
go get github.com/go-playground/validator/v10@latest

# Circuit Breaker & Resilience
go get github.com/sony/gobreaker@latest
go get github.com/cenkalti/backoff/v4@latest

# Testing
go get github.com/stretchr/testify@latest
go get github.com/golang/mock/gomock@latest

# Metrics & Observability
go get github.com/prometheus/client_golang/prometheus@latest
go get go.opentelemetry.io/otel@latest
```

---

## 🔨 Desarrollo

### Variables de Entorno

```bash
# .env.example

# Server
PORT=8080
APP_ENV=development
SERVICE_NAME=orchestrator
VERSION=1.0.0

# AWS
AWS_REGION=us-east-1
AWS_ACCOUNT_ID=123456789012

# DynamoDB Tables
DYNAMODB_TABLE_JOBS=smart-bin-dev-jobs
DYNAMODB_TABLE_DEVICES=smart-bin-dev-devices

# S3
S3_BUCKET_IMAGES=smart-bin-dev-images
S3_PRESIGNED_URL_EXPIRY=15m

# SQS
SQS_QUEUE_URL_CLASSIFICATION=https://sqs.us-east-1.amazonaws.com/123456789012/classification-jobs
SQS_QUEUE_URL_DLQ=https://sqs.us-east-1.amazonaws.com/123456789012/classification-dlq

# IoT Core
IOT_ENDPOINT=xxxxx-ats.iot.us-east-1.amazonaws.com

# Service URLs (other microservices)
CLASSIFIER_SERVICE_URL=http://classifier-service:8081
DECISION_SERVICE_URL=http://decision-service:8082

# Service Discovery (optional)
USE_SERVICE_DISCOVERY=false
SERVICE_DISCOVERY_NAMESPACE=smart-bin.local

# Security
COGNITO_USER_POOL_ID=us-east-1_XXXXXXXXX
COGNITO_CLIENT_ID=xxxxxxxxxxxxxxxxxxxxx
JWT_SECRET=your-jwt-secret-here

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# Circuit Breaker
CIRCUIT_BREAKER_TIMEOUT=30s
CIRCUIT_BREAKER_MAX_REQUESTS=3
CIRCUIT_BREAKER_INTERVAL=60s

# Timeouts
HTTP_CLIENT_TIMEOUT=30s
CLASSIFIER_TIMEOUT=60s
DECISION_TIMEOUT=10s

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json

# Metrics
ENABLE_METRICS=true
METRICS_PORT=9090

# Feature Flags
ENABLE_ASYNC_CLASSIFICATION=true
ENABLE_CACHE=false
ENABLE_TRACING=false
```

### Makefile

```makefile
.PHONY: help setup build run test lint clean docker-build docker-run

# Variables
APP_NAME=orchestrator
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Install dependencies
	@echo "📦 Installing dependencies..."
	go mod download
	go mod verify
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

build: ## Build the application
	@echo "🔨 Building..."
	go build $(LDFLAGS) -o bin/$(APP_NAME) cmd/server/main.go

run: ## Run the application
	@echo "🚀 Running..."
	go run cmd/server/main.go

dev: ## Run with hot reload (requires air)
	@echo "🔥 Running with hot reload..."
	air

test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	@echo "📊 Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-integration: ## Run integration tests
	@echo "🧪 Running integration tests..."
	go test -v -tags=integration ./test/integration/...

lint: ## Run linter
	@echo "🔍 Running linter..."
	golangci-lint run --timeout=5m

fmt: ## Format code
	@echo "💅 Formatting code..."
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	go vet ./...

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache

# Docker targets
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) -f deployments/docker/Dockerfile .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

docker-run: ## Run Docker container
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(APP_NAME):latest

docker-compose-up: ## Start all services with docker-compose
	@echo "🐳 Starting services..."
	docker-compose up -d

docker-compose-down: ## Stop all services
	@echo "🐳 Stopping services..."
	docker-compose down

docker-compose-logs: ## View logs
	docker-compose logs -f orchestrator

# AWS targets
aws-login: ## Login to AWS ECR
	@echo "🔐 Logging in to AWS ECR..."
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com

docker-push: aws-login docker-build ## Push Docker image to ECR
	@echo "📤 Pushing to ECR..."
	docker tag $(APP_NAME):$(VERSION) $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(APP_NAME):$(VERSION)
	docker tag $(APP_NAME):$(VERSION) $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(APP_NAME):latest
	docker push $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(APP_NAME):$(VERSION)
	docker push $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(APP_NAME):latest

# Database migrations
migrate-up: ## Run database migrations up
	@echo "⬆️  Running migrations..."
	# Add migration tool command here

migrate-down: ## Run database migrations down
	@echo "⬇️  Rolling back migrations..."
	# Add migration tool command here

# Development
generate: ## Generate code (mocks, etc.)
	@echo "🔧 Generating code..."
	go generate ./...

watch: ## Watch for changes and rebuild
	@echo "👀 Watching for changes..."
	# Requires air: go install github.com/cosmtrek/air@latest
	air

# Security
security-scan: ## Run security scan
	@echo "🔒 Running security scan..."
	gosec ./...

deps-update: ## Update dependencies
	@echo "⬆️  Updating dependencies..."
	go get -u ./...
	go mod tidy

# CI/CD
ci: lint test ## Run CI pipeline locally
	@echo "✅ CI checks passed"

all: clean lint test build ## Run all checks and build
```

### docker-compose.yml (Desarrollo Local)

```yaml
version: '3.8'

services:
  orchestrator:
    build:
      context: .
      dockerfile: deployments/docker/Dockerfile.dev
    ports:
      - "8080:8080"
      - "9090:9090"  # Metrics
    environment:
      - APP_ENV=development
      - PORT=8080
      - AWS_REGION=us-east-1
      - AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
      - AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
      # Use LocalStack for local development
      - DYNAMODB_ENDPOINT=http://localstack:4566
      - S3_ENDPOINT=http://localstack:4566
      - SQS_ENDPOINT=http://localstack:4566
      - IOT_ENDPOINT=http://localstack:4566
      # Service URLs
      - CLASSIFIER_SERVICE_URL=http://classifier:8081
      - DECISION_SERVICE_URL=http://decision:8082
    volumes:
      - .:/app
      - go-modules:/go/pkg/mod
    depends_on:
      - localstack
      - classifier
      - decision
    networks:
      - smart-bin-network
    command: air  # Hot reload

  # Mock services for development
  classifier:
    image: mockserver/mockserver:latest
    ports:
      - "8081:1080"
    environment:
      MOCKSERVER_INITIALIZATION_JSON_PATH: /config/classifier-mock.json
    volumes:
      - ./test/mocks:/config
    networks:
      - smart-bin-network

  decision:
    image: mockserver/mockserver:latest
    ports:
      - "8082:1080"
    environment:
      MOCKSERVER_INITIALIZATION_JSON_PATH: /config/decision-mock.json
    volumes:
      - ./test/mocks:/config
    networks:
      - smart-bin-network

  localstack:
    image: localstack/localstack:latest
    ports:
      - "4566:4566"
    environment:
      - SERVICES=dynamodb,s3,sqs,iot
      - DEBUG=1
      - DATA_DIR=/tmp/localstack/data
      - DOCKER_HOST=unix:///var/run/docker.sock
    volumes:
      - "./scripts/localstack-init:/etc/localstack/init/ready.d"
      - "/var/run/docker.sock:/var/run/docker.sock"
      - "localstack-data:/tmp/localstack"
    networks:
      - smart-bin-network

  # Optional: Redis for caching
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    networks:
      - smart-bin-network

  # Optional: Prometheus for metrics
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    volumes:
      - ./deployments/prometheus.yml:/etc/prometheus/prometheus.yml
    networks:
      - smart-bin-network

networks:
  smart-bin-network:
    driver: bridge

volumes:
  localstack-data:
  go-modules:
```

---

## 🚀 CI/CD

### GitHub Actions - CI Pipeline

**`.github/workflows/ci.yml`**

```yaml
name: CI Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

env:
  GO_VERSION: '1.21'

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}
      
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m

  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}
      
      - name: Run tests
        run: make test
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

  security:
    name: Security Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Run Gosec Security Scanner
        uses: securego/gosec@master
        with:
          args: ./...

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}
      
      - name: Build
        run: make build
      
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: orchestrator-binary
          path: bin/orchestrator

  docker:
    name: Build Docker Image
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      
      - name: Build Docker image
        run: make docker-build
      
      - name: Save Docker image
        run: docker save orchestrator:latest | gzip > orchestrator-image.tar.gz
      
      - name: Upload Docker image
        uses: actions/upload-artifact@v3
        with:
          name: docker-image
          path: orchestrator-image.tar.gz
```

### CD Pipeline - Development

**`.github/workflows/cd-dev.yml`**

```yaml
name: Deploy to Development

on:
  push:
    branches: [ develop ]

env:
  AWS_REGION: us-east-1
  ECR_REPOSITORY: smart-bin/orchestrator
  ECS_CLUSTER: smart-bin-dev-cluster
  ECS_SERVICE: orchestrator-service
  ECS_TASK_DEFINITION: orchestrator-task-dev

jobs:
  deploy:
    name: Deploy to Dev
    runs-on: ubuntu-latest
    environment: development
    
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ env.AWS_REGION }}
      
      - name: Login to Amazon ECR
        id: login-ecr
        uses: aws-actions/amazon-ecr-login@v2
      
      - name: Build, tag, and push image to Amazon ECR
        id: build-image
        env:
          ECR_REGISTRY: ${{ steps.login-ecr.outputs.registry }}
          IMAGE_TAG: ${{ github.sha }}
        run: |
          docker build -t $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG -f deployments/docker/Dockerfile .
          docker push $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG
          echo "image=$ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG" >> $GITHUB_OUTPUT
      
      - name: Download task definition
        run: |
          aws ecs describe-task-definition \
            --task-definition ${{ env.ECS_TASK_DEFINITION }} \
            --query taskDefinition > task-definition.json
      
      - name: Fill in the new image ID in the Amazon ECS task definition
        id: task-def
        uses: aws-actions/amazon-ecs-render-task-definition@v1
        with:
          task-definition: task-definition.json
          container-name: orchestrator
          image: ${{ steps.build-image.outputs.image }}
      
      - name: Deploy Amazon ECS task definition
        uses: aws-actions/amazon-ecs-deploy-task-definition@v1
        with:
          task-definition: ${{ steps.task-def.outputs.task-definition }}
          service: ${{ env.ECS_SERVICE }}
          cluster: ${{ env.ECS_CLUSTER }}
          wait-for-service-stability: true
      
      - name: Notify deployment
        if: always()
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          text: 'Deployment to DEV: ${{ job.status }}'
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

---
