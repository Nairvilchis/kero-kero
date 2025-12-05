# 🐸 Kero-Kero WhatsApp API

API REST completa para gestionar múltiples instancias de WhatsApp usando `whatsmeow`.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## ✨ Características

- 📱 **Gestión de Instancias**: Crear, listar, conectar y desconectar múltiples cuentas de WhatsApp
- 💬 **Mensajería Completa**: Texto, imágenes, videos, audio, documentos y ubicaciones
- 👥 **Grupos**: Crear, gestionar participantes, obtener información
- 📇 **Contactos**: Verificar números, obtener fotos de perfil, sincronizar contactos
- 🔒 **Privacidad**: Configurar quién puede ver tu información
- 🔔 **Webhooks**: Notificaciones en tiempo real de mensajes y eventos
- 🚀 **Alto Rendimiento**: Arquitectura limpia con Go
- 🐳 **Docker Ready**: Despliegue fácil con Docker Compose
- 🔐 **Seguro**: Autenticación con API Key, rate limiting, CORS configurable

## 🏗️ Arquitectura

```
kero-kero/
├── cmd/
│   └── server/          # Punto de entrada de la aplicación
├── internal/
│   ├── config/          # Configuración
│   ├── handlers/        # Controladores HTTP
│   ├── models/          # Modelos de datos
│   ├── repository/      # Capa de datos (PostgreSQL, Redis)
│   ├── routes/          # Definición de rutas
│   ├── server/          # Middlewares
│   ├── services/        # Lógica de negocio
│   └── whatsapp/        # Cliente de WhatsApp
├── pkg/
│   ├── errors/          # Manejo de errores
│   └── logger/          # Logging estructurado
└── docs/                # Documentación
```

## 🚀 Inicio Rápido

### Opción 1: Docker (Recomendado)

```bash
# Clonar repositorio
git clone <repository-url>
cd kero-kero

# Configurar variables de entorno
cp .env.example .env
nano .env  # Editar valores

# Iniciar servicios
docker-compose up -d

# Verificar estado
curl http://localhost:8080/health
```

### Opción 2: Local

```bash
# Requisitos: Go 1.21+, PostgreSQL, Redis

# Instalar dependencias
go mod download

# Configurar .env
cp .env.example .env

# Ejecutar
go run cmd/server/main.go
```

## 📚 Documentación

- [📋 Endpoints de la API](API_ENDPOINTS.md)
- [🐳 Guía de Despliegue con Docker](DOCKER_DEPLOYMENT.md)
- [🔧 Configuración](docs/CONFIGURATION.md)

## 🎯 Uso Básico

### 1. Crear una instancia

```bash
curl -X POST http://localhost:8080/instances \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"instance_id": "mi-whatsapp"}'
```

### 2. Conectar y obtener QR

```bash
# Conectar
curl -X POST http://localhost:8080/instances/mi-whatsapp/connect \
  -H "X-API-Key: your-api-key"

# Obtener QR
curl http://localhost:8080/instances/mi-whatsapp/qr \
  -H "X-API-Key: your-api-key" \
  --output qr.png
```

### 3. Enviar mensaje

```bash
curl -X POST http://localhost:8080/instances/mi-whatsapp/messages/text \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5215512345678",
    "message": "¡Hola desde Kero-Kero!"
  }'
```

### 4. Configurar webhook

```bash
curl -X POST http://localhost:8080/instances/mi-whatsapp/webhook \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://tu-servidor.com/webhook",
    "events": ["message", "status"],
    "secret": "tu-secreto"
  }'
```

## 🛠️ Comandos Make

```bash
make help              # Ver todos los comandos disponibles
make build             # Compilar aplicación
make run               # Ejecutar localmente
make test              # Ejecutar tests
make docker-up         # Iniciar con Docker
make docker-logs       # Ver logs
make docker-down       # Detener servicios
```

## 📊 Endpoints Disponibles

### Instancias
- `POST /instances` - Crear instancia
- `GET /instances` - Listar instancias
- `GET /instances/{id}` - Obtener detalles
- `DELETE /instances/{id}` - Eliminar instancia
- `POST /instances/{id}/connect` - Conectar
- `GET /instances/{id}/qr` - Obtener QR
- `GET /instances/{id}/status` - Ver estado

### Mensajes
- `POST /instances/{id}/messages/text` - Enviar texto
- `POST /instances/{id}/messages/image` - Enviar imagen
- `POST /instances/{id}/messages/video` - Enviar video
- `POST /instances/{id}/messages/audio` - Enviar audio
- `POST /instances/{id}/messages/document` - Enviar documento
- `POST /instances/{id}/messages/location` - Enviar ubicación

### Grupos
- `POST /instances/{id}/groups` - Crear grupo
- `GET /instances/{id}/groups` - Listar grupos
- `GET /instances/{id}/groups/{groupID}` - Info del grupo
- `PATCH /instances/{id}/groups/{groupID}/participants` - Gestionar participantes

### Contactos
- `POST /instances/{id}/contacts/check` - Verificar números
- `GET /instances/{id}/contacts` - Listar contactos
- `GET /instances/{id}/contacts/profile-picture` - Foto de perfil

### Privacidad
- `GET /instances/{id}/privacy` - Obtener configuración
- `PATCH /instances/{id}/privacy` - Actualizar configuración

### Webhooks
- `POST /instances/{id}/webhook` - Configurar webhook
- `GET /instances/{id}/webhook` - Ver configuración
- `DELETE /instances/{id}/webhook` - Eliminar webhook

## 🔧 Configuración

### Variables de Entorno

```env
# Aplicación
APP_NAME=Kero-Kero
APP_ENV=production
APP_PORT=8080

# Base de Datos
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=kerokero
DB_USER=kerokero
DB_PASSWORD=secret

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Seguridad
API_KEY=your-secret-api-key

# CORS
CORS_ALLOWED_ORIGINS=*
```

## 🧪 Testing

```bash
# Ejecutar todos los tests
make test

# Tests con cobertura
make test-coverage

# Linter
make lint
```

## 📈 Monitoreo

### Health Check

```bash
curl http://localhost:8080/health
```

Respuesta:
```json
{
  "status": "healthy",
  "database": "ok",
  "redis": "ok"
}
```

### Logs

```bash
# Docker
docker-compose logs -f api

# Local
tail -f logs/app.log
```

## 🔒 Seguridad

- ✅ Autenticación con API Key
- ✅ Rate limiting configurable
- ✅ CORS configurable
- ✅ Validación de entrada
- ✅ Webhooks firmados con HMAC-SHA256
- ✅ Logs de auditoría

## 🤝 Contribuir

Las contribuciones son bienvenidas. Por favor:

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📝 Licencia

Este proyecto está bajo la Licencia MIT. Ver el archivo [LICENSE](LICENSE) para más detalles.

## 🆘 Soporte

- 📧 Email: support@example.com
- 💬 Discord: [Link al servidor]
- 📖 Documentación: [docs/](docs/)
- 🐛 Issues: [GitHub Issues](https://github.com/user/repo/issues)

## 🙏 Agradecimientos

- [whatsmeow](https://github.com/tulir/whatsmeow) - Cliente de WhatsApp
- [chi](https://github.com/go-chi/chi) - Router HTTP
- [zerolog](https://github.com/rs/zerolog) - Logger estructurado

---

Hecho con ❤️ y Go
