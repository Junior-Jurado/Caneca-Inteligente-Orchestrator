#!/bin/bash

# =============================================================================
# SCRIPT DE BUILD Y DEPLOY - SMART BIN ORCHESTRATOR
# =============================================================================
# Este script:
# 1. Construye la imagen Docker
# 2. La sube a AWS ECR
# 3. Actualiza el servicio ECS para que use la nueva imagen
# =============================================================================

set -e  # Exit on error

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# =============================================================================
# CONFIGURACIÓN
# =============================================================================

# Cambiar estos valores según tu ambiente
ENVIRONMENT="${ENVIRONMENT:-dev}"  # dev o prod
AWS_REGION="${AWS_REGION:-us-east-1}"
AWS_ACCOUNT_ID="${AWS_ACCOUNT_ID:-940482438767}"

# Valores derivados
ECR_REPOSITORY="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/smartbin-${ENVIRONMENT}-service"
IMAGE_TAG="${IMAGE_TAG:-orchestrator-$(date +%Y%m%d-%H%M%S)}"
ECS_CLUSTER="smartbin-${ENVIRONMENT}-cluster"
ECS_SERVICE="smartbin-${ENVIRONMENT}-orchestrator"
SERVICE_NAME="orchestrator"

# Directorio del proyecto (asume que el script está en la raíz del proyecto)
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# =============================================================================
# FUNCIONES
# =============================================================================

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

header() {
    echo ""
    echo -e "${CYAN}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}${BOLD}  $1${NC}"
    echo -e "${CYAN}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

check_prerequisites() {
    header "VERIFICANDO PRE-REQUISITOS"
    
    # Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker no está instalado"
        exit 1
    fi
    log_success "Docker instalado: $(docker --version)"
    
    # AWS CLI
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI no está instalado"
        exit 1
    fi
    log_success "AWS CLI instalado: $(aws --version)"
    
    # jq (opcional)
    if ! command -v jq &> /dev/null; then
        log_warning "jq no instalado (opcional, pero recomendado)"
    else
        log_success "jq instalado"
    fi
    
    # Verificar credenciales AWS
    if ! aws sts get-caller-identity &> /dev/null; then
        log_error "Credenciales AWS no configuradas"
        exit 1
    fi
    
    local account=$(aws sts get-caller-identity --query 'Account' --output text)
    local user=$(aws sts get-caller-identity --query 'Arn' --output text | cut -d'/' -f2)
    log_success "Autenticado como: $user (Account: $account)"
    
    # Verificar que estamos en el directorio correcto
    if [ ! -f "$PROJECT_DIR/go.mod" ]; then
        log_error "No se encontró go.mod. Asegúrate de ejecutar este script desde la raíz del proyecto"
        exit 1
    fi
    log_success "Directorio del proyecto válido: $PROJECT_DIR"
}

show_config() {
    header "CONFIGURACIÓN"
    
    echo "  Environment:     ${BOLD}$ENVIRONMENT${NC}"
    echo "  AWS Region:      $AWS_REGION"
    echo "  AWS Account:     $AWS_ACCOUNT_ID"
    echo "  ECR Repository:  $ECR_REPOSITORY"
    echo "  Image Tag:       $IMAGE_TAG"
    echo "  ECS Cluster:     $ECS_CLUSTER"
    echo "  ECS Service:     $ECS_SERVICE"
    echo "  Project Dir:     $PROJECT_DIR"
    echo ""
}

build_docker_image() {
    header "PASO 1: CONSTRUYENDO IMAGEN DOCKER"
    
    log_info "Construyendo imagen: $ECR_REPOSITORY:$IMAGE_TAG"
    
    # Build con multi-stage
    if docker build \
        -t "${ECR_REPOSITORY}:${IMAGE_TAG}" \
        -t "${ECR_REPOSITORY}:latest" \
        -f "$PROJECT_DIR/Dockerfile" \
        "$PROJECT_DIR"; then
        log_success "Imagen construida exitosamente"
    else
        log_error "Falló la construcción de la imagen"
        exit 1
    fi
    
    # Mostrar tamaño de la imagen
    local size=$(docker images "${ECR_REPOSITORY}:${IMAGE_TAG}" --format "{{.Size}}")
    log_info "Tamaño de la imagen: $size"
}

test_docker_image() {
    header "PASO 2: PROBANDO IMAGEN LOCALMENTE"
    
    log_info "Iniciando contenedor de prueba..."
    
    # Ejecutar contenedor en background
    local container_id=$(docker run -d \
        -p 18080:8080 \
        -e APP_ENV=test \
        -e AWS_REGION=$AWS_REGION \
        -e PORT=8080 \
        "${ECR_REPOSITORY}:${IMAGE_TAG}")
    
    log_info "Contenedor iniciado: ${container_id:0:12}"
    
    # Esperar a que el servicio inicie
    log_info "Esperando 5 segundos para que el servicio inicie..."
    sleep 5
    
    # Probar health check
    log_info "Probando health check..."
    if curl -s -f http://localhost:18080/health > /dev/null; then
        log_success "Health check exitoso"
    else
        log_warning "Health check falló (puede ser normal si las dependencias AWS no están configuradas)"
    fi
    
    # Ver logs
    log_info "Últimos logs del contenedor:"
    docker logs "$container_id" | tail -10
    
    # Detener y eliminar contenedor
    log_info "Deteniendo contenedor de prueba..."
    docker stop "$container_id" > /dev/null
    docker rm "$container_id" > /dev/null
    
    log_success "Prueba local completada"
}

login_ecr() {
    header "PASO 3: LOGIN A AWS ECR"
    
    log_info "Obteniendo credenciales de ECR..."
    
    if aws ecr get-login-password --region "$AWS_REGION" | \
       docker login --username AWS --password-stdin "$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com"; then
        log_success "Login exitoso a ECR"
    else
        log_error "Falló el login a ECR"
        exit 1
    fi
}

push_to_ecr() {
    header "PASO 4: SUBIENDO IMAGEN A ECR"
    
    log_info "Subiendo imagen con tag: $IMAGE_TAG"
    
    if docker push "${ECR_REPOSITORY}:${IMAGE_TAG}"; then
        log_success "Imagen con tag $IMAGE_TAG subida"
    else
        log_error "Falló el push de la imagen"
        exit 1
    fi
    
    log_info "Subiendo imagen con tag: latest"
    
    if docker push "${ECR_REPOSITORY}:latest"; then
        log_success "Imagen latest subida"
    else
        log_error "Falló el push de latest"
        exit 1
    fi
}

update_ecs_service() {
    header "PASO 5: ACTUALIZANDO SERVICIO ECS"
    
    log_info "Forzando nuevo deployment en ECS..."
    
    # Forzar nuevo deployment
    if aws ecs update-service \
        --cluster "$ECS_CLUSTER" \
        --service "$ECS_SERVICE" \
        --force-new-deployment \
        --region "$AWS_REGION" > /dev/null; then
        log_success "Deployment iniciado"
    else
        log_error "Falló la actualización del servicio"
        exit 1
    fi
    
    log_info "Esperando a que el servicio se estabilice..."
    log_warning "Esto puede tomar 2-5 minutos..."
    
    # Esperar a que el servicio se estabilice
    if aws ecs wait services-stable \
        --cluster "$ECS_CLUSTER" \
        --services "$ECS_SERVICE" \
        --region "$AWS_REGION"; then
        log_success "Servicio actualizado y estable"
    else
        log_error "El servicio no se estabilizó en el tiempo esperado"
        log_warning "Verifica el estado del servicio en la consola de AWS"
        exit 1
    fi
}

verify_deployment() {
    header "PASO 6: VERIFICANDO DEPLOYMENT"
    
    # Obtener información del servicio
    local service_info=$(aws ecs describe-services \
        --cluster "$ECS_CLUSTER" \
        --services "$ECS_SERVICE" \
        --region "$AWS_REGION" \
        --query 'services[0]')
    
    local running=$(echo "$service_info" | jq -r '.runningCount')
    local desired=$(echo "$service_info" | jq -r '.desiredCount')
    
    echo "  Running Tasks: $running/$desired"
    
    if [ "$running" == "$desired" ]; then
        log_success "Todas las tasks están corriendo"
    else
        log_warning "Algunas tasks no están corriendo ($running/$desired)"
    fi
    
    # Obtener ID de la task más reciente
    log_info "Obteniendo logs de la task más reciente..."
    
    local task_arn=$(aws ecs list-tasks \
        --cluster "$ECS_CLUSTER" \
        --service-name "$ECS_SERVICE" \
        --region "$AWS_REGION" \
        --query 'taskArns[0]' \
        --output text)
    
    if [ ! -z "$task_arn" ] && [ "$task_arn" != "None" ]; then
        log_info "Task ARN: ${task_arn##*/}"
        
        # Mostrar últimos logs
        log_info "Últimos logs:"
        aws logs tail "/ecs/$ECS_SERVICE" \
            --since 2m \
            --region "$AWS_REGION" \
            --format short 2>/dev/null | tail -10 || log_warning "No se pudieron obtener logs"
    fi
}

show_summary() {
    header "RESUMEN DEL DEPLOYMENT"
    
    echo -e "  ${GREEN}✓${NC} Imagen construida: ${BOLD}$IMAGE_TAG${NC}"
    echo -e "  ${GREEN}✓${NC} Subida a ECR: ${BOLD}$ECR_REPOSITORY${NC}"
    echo -e "  ${GREEN}✓${NC} Servicio actualizado: ${BOLD}$ECS_SERVICE${NC}"
    echo ""
    echo -e "${CYAN}Para ver los logs en tiempo real:${NC}"
    echo -e "  ${BOLD}aws logs tail /ecs/$ECS_SERVICE --follow --region $AWS_REGION${NC}"
    echo ""
    echo -e "${CYAN}Para ver el estado del servicio:${NC}"
    echo -e "  ${BOLD}aws ecs describe-services --cluster $ECS_CLUSTER --services $ECS_SERVICE --region $AWS_REGION${NC}"
    echo ""
}

cleanup() {
    log_info "Limpiando imágenes Docker locales antiguas..."
    docker image prune -f > /dev/null 2>&1 || true
}

# =============================================================================
# MAIN
# =============================================================================

main() {
    clear
    
    echo -e "${CYAN}${BOLD}"
    cat << "EOF"
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║        SMART BIN - BUILD & DEPLOY TO AWS                     ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
    
    # Verificar argumentos
    if [ "$1" == "--skip-test" ]; then
        SKIP_TEST=true
        log_warning "Saltando pruebas locales"
    else
        SKIP_TEST=false
    fi
    
    # Ejecutar pasos
    check_prerequisites
    show_config
    
    # Confirmar
    echo -n -e "${YELLOW}¿Continuar con el deployment? (y/n): ${NC}"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        log_warning "Deployment cancelado"
        exit 0
    fi
    
    build_docker_image
    
    if [ "$SKIP_TEST" = false ]; then
        test_docker_image
    fi
    
    login_ecr
    push_to_ecr
    update_ecs_service
    verify_deployment
    cleanup
    show_summary
    
    log_success "DEPLOYMENT COMPLETADO EXITOSAMENTE! 🎉"
}

# Ejecutar
main "$@"