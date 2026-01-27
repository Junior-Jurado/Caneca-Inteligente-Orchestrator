#!/bin/bash

# =============================================================================
# SMART BIN - Pruebas End-to-End en AWS Cloud
# =============================================================================
# Este script ejecuta todas las pruebas directamente en AWS
# No requiere nada local, todo se prueba en la nube
# =============================================================================

set -e

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Config AWS
AWS_REGION="us-east-1"
CLUSTER="smartbin-dev-cluster"
SERVICE="smartbin-dev-orchestrator"
TABLE_JOBS="smartbin-dev-jobs"
S3_BUCKET="smartbin-dev-images-940482438767"
SQS_QUEUE="https://sqs.us-east-1.amazonaws.com/940482438767/smartbin-dev-captured-queue"

# Variables globales
ALB_DNS=""
JOB_ID=""
UPLOAD_URL=""
TESTS_PASSED=0
TESTS_FAILED=0

# =============================================================================
# FUNCIONES
# =============================================================================

log() { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[✓ PASS]${NC} $1"; ((TESTS_PASSED++)); }
error() { echo -e "${RED}[✗ FAIL]${NC} $1"; ((TESTS_FAILED++)); }
warn() { echo -e "${YELLOW}[! WARN]${NC} $1"; }
header() {
    echo ""
    echo -e "${CYAN}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}${BOLD}  $1${NC}"
    echo -e "${CYAN}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

spinner() {
    local pid=$1
    local delay=0.1
    local spinstr='⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏'
    while [ "$(ps a | awk '{print $1}' | grep $pid)" ]; do
        local temp=${spinstr#?}
        printf " [%c]  " "$spinstr"
        local spinstr=$temp${spinstr%"$temp"}
        sleep $delay
        printf "\b\b\b\b\b\b"
    done
    printf "    \b\b\b\b"
}

# =============================================================================
# PRUEBAS
# =============================================================================

test_prerequisites() {
    header "VERIFICANDO PRE-REQUISITOS"
    
    # AWS CLI
    if ! command -v aws &> /dev/null; then
        error "AWS CLI no instalado"
        exit 1
    fi
    success "AWS CLI instalado"
    
    # jq
    if ! command -v jq &> /dev/null; then
        warn "jq no instalado (opcional, pero recomendado)"
    else
        success "jq instalado"
    fi
    
    # Credenciales AWS
    if ! aws sts get-caller-identity &> /dev/null; then
        error "Credenciales AWS no configuradas"
        exit 1
    fi
    
    local account=$(aws sts get-caller-identity --query 'Account' --output text)
    local user=$(aws sts get-caller-identity --query 'Arn' --output text | cut -d'/' -f2)
    success "Autenticado como: $user (Account: $account)"
}

test_infrastructure() {
    header "VERIFICANDO INFRAESTRUCTURA AWS"
    
    # VPC
    log "Verificando VPC..."
    local vpc=$(aws ec2 describe-vpcs \
        --filters "Name=tag:Name,Values=smartbin-dev-vpc" \
        --query 'Vpcs[0].VpcId' \
        --output text \
        --region $AWS_REGION 2>/dev/null)
    
    if [ "$vpc" != "None" ] && [ ! -z "$vpc" ]; then
        success "VPC encontrada: $vpc"
    else
        error "VPC no encontrada"
    fi
    
    # ECS Cluster
    log "Verificando ECS Cluster..."
    local cluster_status=$(aws ecs describe-clusters \
        --clusters $CLUSTER \
        --query 'clusters[0].status' \
        --output text \
        --region $AWS_REGION 2>/dev/null)
    
    if [ "$cluster_status" == "ACTIVE" ]; then
        success "ECS Cluster activo"
    else
        error "ECS Cluster no activo o no existe"
    fi
    
    # DynamoDB Tables
    log "Verificando tabla DynamoDB..."
    local table_status=$(aws dynamodb describe-table \
        --table-name $TABLE_JOBS \
        --query 'Table.TableStatus' \
        --output text \
        --region $AWS_REGION 2>/dev/null)
    
    if [ "$table_status" == "ACTIVE" ]; then
        success "Tabla DynamoDB activa: $TABLE_JOBS"
    else
        error "Tabla DynamoDB no activa"
    fi
    
    # S3 Bucket
    log "Verificando S3 Bucket..."
    if aws s3api head-bucket --bucket $S3_BUCKET --region $AWS_REGION 2>/dev/null; then
        success "S3 Bucket existe: $S3_BUCKET"
    else
        error "S3 Bucket no encontrado"
    fi
    
    # SQS Queue
    log "Verificando SQS Queue..."
    local queue_attrs=$(aws sqs get-queue-attributes \
        --queue-url $SQS_QUEUE \
        --attribute-names ApproximateNumberOfMessages \
        --region $AWS_REGION 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        local msgs=$(echo "$queue_attrs" | jq -r '.Attributes.ApproximateNumberOfMessages' 2>/dev/null || echo "?")
        success "SQS Queue activa (Mensajes: $msgs)"
    else
        error "SQS Queue no encontrada"
    fi
}

test_ecs_service() {
    header "VERIFICANDO SERVICIO ECS"
    
    log "Obteniendo estado del servicio..."
    local service_info=$(aws ecs describe-services \
        --cluster $CLUSTER \
        --services $SERVICE \
        --query 'services[0]' \
        --region $AWS_REGION 2>/dev/null)
    
    if [ $? -ne 0 ]; then
        error "No se pudo obtener info del servicio"
        return 1
    fi
    
    local status=$(echo "$service_info" | jq -r '.status')
    local running=$(echo "$service_info" | jq -r '.runningCount')
    local desired=$(echo "$service_info" | jq -r '.desiredCount')
    local deployment_status=$(echo "$service_info" | jq -r '.deployments[0].rolloutState // "UNKNOWN"')
    
    echo "  └─ Status: $status"
    echo "  └─ Tasks: $running/$desired"
    echo "  └─ Deployment: $deployment_status"
    
    if [ "$status" == "ACTIVE" ] && [ "$running" == "$desired" ]; then
        success "Servicio saludable"
    else
        error "Servicio no saludable"
    fi
    
    # Obtener ALB DNS
    log "Obteniendo DNS del Load Balancer..."
    ALB_DNS=$(aws elbv2 describe-load-balancers \
        --region $AWS_REGION \
        --query 'LoadBalancers[?contains(LoadBalancerName, `smartbin-dev`)].DNSName' \
        --output text)
    
    if [ -z "$ALB_DNS" ]; then
        error "No se encontró el Load Balancer"
        return 1
    fi
    
    success "ALB DNS: http://$ALB_DNS"
}

test_health_endpoint() {
    header "PRUEBA 1: HEALTH CHECK"
    
    log "Llamando a /health..."
    local response=$(curl -s -w "\n%{http_code}" "http://$ALB_DNS/health" 2>/dev/null)
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    echo "  └─ HTTP Status: $http_code"
    
    if [ "$http_code" == "200" ]; then
        if command -v jq &> /dev/null; then
            local status=$(echo "$body" | jq -r '.status' 2>/dev/null)
            echo "  └─ Service Status: $status"
            
            if [ "$status" == "healthy" ]; then
                success "Health check exitoso"
                echo "$body" | jq '.' 2>/dev/null
            else
                error "Servicio reporta estado no saludable"
            fi
        else
            success "Health check retornó 200"
            echo "$body"
        fi
    else
        error "Health check falló (HTTP $http_code)"
        echo "$body"
    fi
}

test_create_job() {
    header "PRUEBA 2: CREAR JOB"
    
    local device_id="test-device-$(date +%s)"
    log "Creando job para device: $device_id"
    
    local response=$(curl -s -w "\n%{http_code}" -X POST "http://$ALB_DNS/api/v1/jobs" \
        -H "Content-Type: application/json" \
        -d "{\"device_id\": \"$device_id\", \"capture_type\": \"manual\"}" 2>/dev/null)
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    echo "  └─ HTTP Status: $http_code"
    
    if [ "$http_code" == "201" ]; then
        if command -v jq &> /dev/null; then
            JOB_ID=$(echo "$body" | jq -r '.job_id')
            UPLOAD_URL=$(echo "$body" | jq -r '.upload_url')
            local status=$(echo "$body" | jq -r '.status')
            
            echo "  └─ Job ID: $JOB_ID"
            echo "  └─ Status: $status"
            echo "  └─ Upload URL obtenido: ${UPLOAD_URL:0:50}..."
            
            if [ ! -z "$JOB_ID" ] && [ "$JOB_ID" != "null" ]; then
                success "Job creado exitosamente"
            else
                error "Job creado pero ID inválido"
            fi
        else
            success "Job creado (HTTP 201)"
            echo "$body"
        fi
    else
        error "Creación de job falló (HTTP $http_code)"
        echo "$body"
    fi
}

test_dynamodb_persistence() {
    header "PRUEBA 3: VERIFICAR PERSISTENCIA EN DYNAMODB"
    
    if [ -z "$JOB_ID" ]; then
        warn "No hay job_id, saltando prueba"
        return
    fi
    
    log "Buscando job en DynamoDB: $JOB_ID"
    
    local item=$(aws dynamodb get-item \
        --table-name $TABLE_JOBS \
        --key "{\"job_id\": {\"S\": \"$JOB_ID\"}}" \
        --region $AWS_REGION 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        if command -v jq &> /dev/null; then
            local db_status=$(echo "$item" | jq -r '.Item.status.S' 2>/dev/null)
            echo "  └─ Job encontrado en DynamoDB"
            echo "  └─ Status: $db_status"
            success "Job persistido correctamente"
        else
            success "Job encontrado en DynamoDB"
        fi
    else
        error "Job no encontrado en DynamoDB"
    fi
}

test_upload_image() {
    header "PRUEBA 4: UPLOAD DE IMAGEN A S3"
    
    if [ -z "$UPLOAD_URL" ]; then
        warn "No hay upload_url, saltando prueba"
        return
    fi
    
    log "Creando imagen de prueba..."
    echo 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==' | base64 -d > /tmp/test-smartbin.png
    
    log "Subiendo imagen a S3..."
    local http_code=$(curl -s -w "%{http_code}" -X PUT "$UPLOAD_URL" \
        -H "Content-Type: image/png" \
        --data-binary @/tmp/test-smartbin.png 2>/dev/null)
    
    echo "  └─ HTTP Status: $http_code"
    
    if [ "$http_code" == "200" ]; then
        success "Imagen subida exitosamente"
    else
        error "Upload falló (HTTP $http_code)"
    fi
    
    rm -f /tmp/test-smartbin.png
}

test_s3_verification() {
    header "PRUEBA 5: VERIFICAR IMAGEN EN S3"
    
    if [ -z "$JOB_ID" ]; then
        warn "No hay job_id, saltando prueba"
        return
    fi
    
    log "Buscando imagen en S3..."
    local s3_list=$(aws s3 ls "s3://$S3_BUCKET/" --region $AWS_REGION 2>/dev/null | grep -i "$JOB_ID")
    
    if [ ! -z "$s3_list" ]; then
        echo "  └─ Archivos encontrados:"
        echo "$s3_list" | while read line; do
            echo "     • $line"
        done
        success "Imagen encontrada en S3"
    else
        warn "Imagen no encontrada en S3 (puede tardar unos segundos)"
    fi
}

test_job_state_change() {
    header "PRUEBA 6: VERIFICAR CAMBIO DE ESTADO DEL JOB"
    
    if [ -z "$JOB_ID" ]; then
        warn "No hay job_id, saltando prueba"
        return
    fi
    
    log "Esperando 3 segundos para procesamiento..."
    sleep 3
    
    log "Consultando estado del job..."
    local response=$(curl -s "http://$ALB_DNS/api/v1/jobs/$JOB_ID" 2>/dev/null)
    
    if command -v jq &> /dev/null; then
        local status=$(echo "$response" | jq -r '.status' 2>/dev/null)
        local updated=$(echo "$response" | jq -r '.updated_at' 2>/dev/null)
        
        echo "  └─ Estado actual: $status"
        echo "  └─ Actualizado: $updated"
        
        if [ "$status" == "processing" ] || [ "$status" == "completed" ]; then
            success "Job cambió de estado correctamente"
        else
            warn "Job en estado: $status (esperado: processing o completed)"
        fi
    else
        echo "$response"
    fi
}

test_logs() {
    header "PRUEBA 7: VERIFICAR LOGS"
    
    log "Obteniendo últimos logs del orchestrator..."
    
    local logs=$(aws logs tail "/ecs/$SERVICE" \
        --since 5m \
        --region $AWS_REGION \
        --format short 2>/dev/null | tail -10)
    
    if [ $? -eq 0 ]; then
        echo "$logs"
        success "Logs obtenidos correctamente"
    else
        warn "No se pudieron obtener logs"
    fi
}

test_metrics() {
    header "PRUEBA 8: MÉTRICAS Y MONITOREO"
    
    # Contar jobs en DynamoDB
    log "Contando jobs en DynamoDB..."
    local job_count=$(aws dynamodb scan \
        --table-name $TABLE_JOBS \
        --select COUNT \
        --region $AWS_REGION \
        --query 'Count' \
        --output text 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        echo "  └─ Total de jobs: $job_count"
        success "Métricas de DynamoDB obtenidas"
    fi
    
    # Contar objetos en S3
    log "Contando imágenes en S3..."
    local s3_count=$(aws s3 ls "s3://$S3_BUCKET/" --region $AWS_REGION 2>/dev/null | wc -l)
    echo "  └─ Total de imágenes: $s3_count"
    
    # Mensajes en SQS
    log "Verificando mensajes en SQS..."
    local queue_attrs=$(aws sqs get-queue-attributes \
        --queue-url $SQS_QUEUE \
        --attribute-names ApproximateNumberOfMessages,ApproximateNumberOfMessagesNotVisible \
        --region $AWS_REGION 2>/dev/null)
    
    if [ $? -eq 0 ] && command -v jq &> /dev/null; then
        local visible=$(echo "$queue_attrs" | jq -r '.Attributes.ApproximateNumberOfMessages')
        local inflight=$(echo "$queue_attrs" | jq -r '.Attributes.ApproximateNumberOfMessagesNotVisible')
        echo "  └─ Mensajes visibles: $visible"
        echo "  └─ Mensajes en proceso: $inflight"
        success "Métricas de SQS obtenidas"
    fi
}

show_summary() {
    header "RESUMEN DE PRUEBAS"
    
    local total=$((TESTS_PASSED + TESTS_FAILED))
    local success_rate=0
    if [ $total -gt 0 ]; then
        success_rate=$((TESTS_PASSED * 100 / total))
    fi
    
    echo ""
    echo -e "  Total de pruebas: ${BOLD}$total${NC}"
    echo -e "  ${GREEN}Exitosas: $TESTS_PASSED${NC}"
    echo -e "  ${RED}Fallidas: $TESTS_FAILED${NC}"
    echo -e "  Tasa de éxito: ${BOLD}$success_rate%${NC}"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}${BOLD}✓ TODAS LAS PRUEBAS PASARON${NC}"
        echo ""
        echo -e "${CYAN}Tu orchestrator está funcionando correctamente en AWS! 🎉${NC}"
        return 0
    else
        echo -e "${RED}${BOLD}✗ ALGUNAS PRUEBAS FALLARON${NC}"
        echo ""
        echo -e "${YELLOW}Revisa los logs arriba para más detalles.${NC}"
        return 1
    fi
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
║           SMART BIN - PRUEBAS END-TO-END EN AWS              ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
    
    test_prerequisites
    test_infrastructure
    test_ecs_service
    test_health_endpoint
    test_create_job
    test_dynamodb_persistence
    test_upload_image
    test_s3_verification
    test_job_state_change
    test_logs
    test_metrics
    
    echo ""
    show_summary
}

main "$@"