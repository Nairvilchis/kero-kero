#!/bin/bash

# Script de prueba para Kero-Kero WhatsApp API
# Este script demuestra el flujo completo de uso de la API

set -e

API_URL="http://localhost:8080"
INSTANCE_ID="test_$(date +%s)"

echo "🐸 Kero-Kero WhatsApp API - Script de Prueba"
echo "=============================================="
echo ""

# Colores para output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 1. Crear instancia
echo -e "${BLUE}[1/5]${NC} Creando instancia: ${INSTANCE_ID}"
curl -s -X POST -H "Content-Type: application/json" \
  -d "{\"instance_id\": \"${INSTANCE_ID}\"}" \
  ${API_URL}/instances | jq .
echo ""

# 2. Conectar instancia
echo -e "${BLUE}[2/5]${NC} Conectando instancia..."
curl -s -X POST ${API_URL}/instances/${INSTANCE_ID}/connect
echo ""
echo ""

# Esperar un momento para que se genere el QR
sleep 2

# 3. Obtener QR
echo -e "${BLUE}[3/5]${NC} Descargando código QR..."
curl -s ${API_URL}/instances/${INSTANCE_ID}/qr -o qr_${INSTANCE_ID}.png
echo -e "${GREEN}✓${NC} QR guardado en: qr_${INSTANCE_ID}.png"
echo ""

# Intentar abrir el QR automáticamente
if command -v xdg-open &> /dev/null; then
    xdg-open qr_${INSTANCE_ID}.png 2>/dev/null &
elif command -v open &> /dev/null; then
    open qr_${INSTANCE_ID}.png 2>/dev/null &
fi

# 4. Verificar estado
echo -e "${BLUE}[4/5]${NC} Verificando estado de la conexión..."
curl -s ${API_URL}/instances/${INSTANCE_ID}/status | jq .
echo ""

# 5. Esperar autenticación
echo -e "${YELLOW}⏳${NC} Esperando autenticación..."
echo "   Por favor, escanea el código QR con WhatsApp"
echo ""

# Polling del estado hasta que esté autenticado
MAX_ATTEMPTS=60
ATTEMPT=0
while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
    STATUS=$(curl -s ${API_URL}/instances/${INSTANCE_ID}/status | jq -r '.status')
    
    if [ "$STATUS" == "authenticated" ]; then
        echo -e "${GREEN}✓${NC} ¡Autenticación exitosa!"
        break
    fi
    
    echo -ne "   Intento $((ATTEMPT+1))/${MAX_ATTEMPTS} - Estado: ${STATUS}\r"
    sleep 2
    ATTEMPT=$((ATTEMPT+1))
done

echo ""

if [ "$STATUS" != "authenticated" ]; then
    echo -e "${YELLOW}⚠${NC}  No se pudo autenticar en el tiempo esperado"
    echo "   Puedes intentar enviar un mensaje manualmente más tarde"
    exit 0
fi

# 6. Enviar mensaje de prueba (opcional)
echo ""
read -p "¿Deseas enviar un mensaje de prueba? (s/N): " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Ss]$ ]]; then
    read -p "Ingresa el número de teléfono (con código de país, sin +): " PHONE
    
    if [ ! -z "$PHONE" ]; then
        echo -e "${BLUE}[5/5]${NC} Enviando mensaje de prueba..."
        
        curl -s -X POST -H "Content-Type: application/json" \
          -d "{
            \"phone\": \"${PHONE}\",
            \"message\": \"¡Hola! Este es un mensaje de prueba desde Kero-Kero API 🐸\"
          }" \
          ${API_URL}/instances/${INSTANCE_ID}/messages/text | jq .
        
        echo ""
        echo -e "${GREEN}✓${NC} Proceso completado!"
    fi
else
    echo "Proceso completado sin enviar mensaje"
fi

echo ""
echo "=============================================="
echo "Instance ID: ${INSTANCE_ID}"
echo "Para usar esta instancia, utiliza el ID en tus requests"
echo ""
