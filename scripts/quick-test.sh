#!/bin/bash
# One-liner test para Smart Bin en AWS

echo "🚀 Iniciando prueba rápida..."

# Obtener ALB DNS
ALB=$(aws elbv2 describe-load-balancers --region us-east-1 --query 'LoadBalancers[?contains(LoadBalancerName, `smartbin-dev`)].DNSName' --output text)
echo "📡 ALB: http://$ALB"

# 1. Health Check
echo ""
echo "1️⃣  Health Check..."
curl -s http://$ALB/health | jq '.'

# 2. Crear Job
echo ""
echo "2️⃣  Creando Job..."
RESPONSE=$(curl -s -X POST http://$ALB/api/v1/jobs -H "Content-Type: application/json" -d "{\"device_id\": \"test-$(date +%s)\", \"capture_type\": \"manual\"}")
echo "$RESPONSE" | jq '.'

JOB_ID=$(echo "$RESPONSE" | jq -r '.job_id')
UPLOAD_URL=$(echo "$RESPONSE" | jq -r '.upload_url')

# 3. Subir Imagen
echo ""
echo "3️⃣  Subiendo imagen al job: $JOB_ID"
echo 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==' | base64 -d > /tmp/test.png
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "$UPLOAD_URL" -H "Content-Type: image/png" --data-binary @/tmp/test.png)
echo "   Status: $STATUS"

# 4. Verificar en DynamoDB
echo ""
echo "4️⃣  Verificando en DynamoDB..."
aws dynamodb get-item --table-name smartbin-dev-jobs --key "{\"job_id\": {\"S\": \"$JOB_ID\"}}" --region us-east-1 | jq '.Item.status'

# 5. Verificar en S3
echo ""
echo "5️⃣  Verificando en S3..."
aws s3 ls s3://smartbin-dev-images-940482438767/ --region us-east-1 | grep -i jpg | tail -3

# 6. Esperar y verificar cambio de estado
echo ""
echo "6️⃣  Esperando procesamiento (3s)..."
sleep 3
echo "   Estado del job:"
curl -s http://$ALB/api/v1/jobs/$JOB_ID | jq '.status'

# Resumen
echo ""
echo "✅ Prueba completada!"
echo "📊 Stats:"
aws dynamodb scan --table-name smartbin-dev-jobs --select COUNT --region us-east-1 --query 'Count' | xargs echo "   Jobs totales:"
echo "   Último job: $JOB_ID"

rm /tmp/test.png 2>/dev/null