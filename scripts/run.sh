#!/bin/bash

# Script para iniciar el servidor Kero-Kero

set -e

echo "🐸 Iniciando Kero-Kero WhatsApp API..."
echo ""

# Verificar que existe .env.local
if [ ! -f .env.local ]; then
    echo "⚠️  No se encontró .env.local, copiando desde .env.example..."
    cp .env.example .env.local
    echo "✅ Archivo .env.local creado. Por favor, configúralo antes de continuar."
    exit 1
fi

# Crear directorio de datos si no existe
mkdir -p data

# Compilar
echo "📦 Compilando..."
go build -o server cmd/server/main.go

# Verificar servicios requeridos
echo ""
echo "🔍 Verificando servicios..."

# Verificar Redis
if ! redis-cli ping > /dev/null 2>&1; then
    echo "⚠️  Redis no está corriendo. Iniciando Redis..."
    echo "   Ejecuta: redis-server &"
    echo "   O usa Docker: docker run -d -p 6379:6379 redis:alpine"
fi

# Si usa PostgreSQL, verificar conexión
DB_DRIVER=$(grep "^DB_DRIVER=" .env.local | cut -d'=' -f2)
if [ "$DB_DRIVER" = "postgres" ]; then
    echo "   Base de datos: PostgreSQL"
    # Aquí podrías agregar verificación de PostgreSQL
else
    echo "   Base de datos: SQLite (./data/kerokero_dev.db)"
fi

echo ""
echo "🚀 Iniciando servidor..."
echo ""

# Ejecutar servidor
./server
