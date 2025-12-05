# Documentación Técnica del Sistema Kero-Kero

## 1. Visión General del Sistema

Kero-Kero es una API REST robusta y escalable diseñada para interactuar con la red de WhatsApp utilizando la librería `whatsmeow`. El sistema permite la gestión de múltiples instancias de WhatsApp, envío y recepción de mensajes, administración de grupos y contactos, y manejo de eventos en tiempo real a través de webhooks.

### 1.1. Tecnologías Principales

*   **Lenguaje:** Go (Golang) 1.21+
*   **Core WhatsApp:** `go.mau.fi/whatsmeow`
*   **Router HTTP:** `go-chi/chi/v5`
*   **Base de Datos:** PostgreSQL (driver `pgx/v5`)
*   **Caché y Colas:** Redis (`go-redis/v9`)
*   **Logging:** `zerolog`
*   **Testing:** `testify`, `miniredis`
*   **Contenedorización:** Docker

---

## 2. Arquitectura del Sistema

El proyecto sigue una arquitectura limpia (Clean Architecture) para asegurar la mantenibilidad y escalabilidad.

### 2.1. Estructura de Directorios

```
kero-kero/
├── cmd/
│   └── server/          # Punto de entrada de la aplicación (main.go)
├── internal/
│   ├── config/          # Carga y validación de configuración (.env)
│   ├── handlers/        # Controladores HTTP (capa de presentación)
│   ├── models/          # Definiciones de estructuras de datos (DTOs y Entidades)
│   ├── repository/      # Capa de acceso a datos (PostgreSQL y Redis)
│   ├── routes/          # Definición de rutas y middlewares
│   ├── server/          # Configuración del servidor HTTP y middlewares globales
│   ├── services/        # Lógica de negocio
│   └── whatsapp/        # Gestor de instancias y cliente whatsmeow
├── pkg/
│   ├── errors/          # Sistema de manejo de errores centralizado
│   └── logger/          # Configuración de logging estructurado
├── docs/                # Documentación del proyecto
├── Dockerfile           # Definición de imagen Docker
├── docker-compose.yml   # Orquestación de servicios
└── Makefile             # Comandos de utilidad
```

### 2.2. Componentes Principales

1.  **Manager (`internal/whatsapp/manager.go`):**
    *   Gestiona el ciclo de vida de múltiples instancias de WhatsApp.
    *   Mantiene un mapa concurrente de `instance_id` a clientes activos.
    *   Se encarga de la reconexión automática y persistencia de sesiones.

2.  **Client (`internal/whatsapp/client.go`):**
    *   Wrapper alrededor de `whatsmeow.Client`.
    *   Maneja eventos específicos de una conexión (QR, mensajes, estado).
    *   Procesa eventos entrantes y los distribuye a los webhooks configurados.

3.  **Services (`internal/services/`):**
    *   Contienen la lógica de negocio pura.
    *   Ejemplos: `AuthService`, `MessageService`, `GroupService`.
    *   Orquestan llamadas entre repositorios y el manager de WhatsApp.

4.  **Repositories (`internal/repository/`):**
    *   Abstraen el acceso a la base de datos y caché.
    *   `MessageRepository`: Persistencia de historial de chat.
    *   `InstanceRepository`: Gestión de metadatos de instancias.

### 2.3. Flujo de Datos

**Envío de Mensaje:**
1.  Cliente HTTP realiza POST a `/instances/{id}/messages/text`.
2.  `MessageHandler` valida la petición y llama a `MessageService`.
3.  `MessageService` verifica la instancia y usa `WhatsAppManager` para obtener el cliente.
4.  El cliente envía el mensaje a través de `whatsmeow`.
5.  `MessageRepository` guarda el mensaje enviado en PostgreSQL.
6.  Se retorna la respuesta al cliente HTTP.

**Recepción de Mensaje:**
1.  WhatsApp envía evento a `whatsmeow`.
2.  `Client` captura el evento en `EventHandler`.
3.  Se extraen los datos y se normalizan.
4.  `MessageRepository` guarda el mensaje recibido.
5.  Si hay webhooks configurados, se envía una petición POST al endpoint del usuario.

---

## 3. Base de Datos y Persistencia

### 3.1. Esquema de Base de Datos (PostgreSQL)

El sistema utiliza PostgreSQL para persistencia relacional.

**Tabla `messages`:**
Almacena el historial de todos los mensajes enviados y recibidos.

```sql
CREATE TABLE messages (
    id TEXT PRIMARY KEY,              -- ID único del mensaje (WhatsApp ID)
    instance_id TEXT NOT NULL,        -- ID de la instancia asociada
    jid TEXT NOT NULL,                -- JID del chat (remote JID)
    from_me BOOLEAN NOT NULL,         -- true si fue enviado por la instancia
    content TEXT,                     -- Contenido del mensaje (texto, caption, etc.)
    timestamp TIMESTAMP,              -- Hora del mensaje
    status TEXT,                      -- Estado: sent, delivered, read, received
    type TEXT,                        -- Tipo: text, image, video, audio, etc.
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Tabla `instances` (Conceptual):**
Gestionada principalmente por `whatsmeow` en sus tablas internas, pero el sistema puede mantener metadatos adicionales.

### 3.2. Redis

Redis se utiliza para:
*   **Caché de sesiones:** Almacenamiento temporal de estados.
*   **Colas de tareas:** (Si aplica) Procesamiento asíncrono.
*   **Rate Limiting:** Control de frecuencia de peticiones.

---

## 4. Referencia de la API

Todas las respuestas de la API siguen un formato estándar JSON:

**Éxito:**
```json
{
  "success": true,
  "data": { ... }
}
```

**Error:**
```json
{
  "success": false,
  "error": "Descripción del error",
  "code": "ERROR_CODE"
}
```

### 4.1. Autenticación

Kero-Kero soporta dos métodos de autenticación:

1. **API Key (Directo)**: Para acceso directo a la API
   - Header: `X-API-Key: tu-api-key`
   - O: `Authorization: Bearer tu-api-key`

2. **JWT (Dashboard)**: Para aplicaciones web que requieren sesiones
   - Primero autenticarse con API Key en `/auth/login`
   - Recibir un token JWT válido por 24 horas
   - Usar token en peticiones: `Authorization: Bearer <jwt-token>`

**Endpoints de Autenticación:**

| Método | Endpoint | Descripción | Requiere Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/auth/login` | Autenticar con API Key y obtener JWT | No |
| `GET` | `/auth/validate` | Validar token JWT | Sí (JWT) |

**Ejemplo de Login:**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"api_key":"dev-api-key-12345"}'
```

**Respuesta:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": 1733444123,
  "type": "Bearer"
}
```

📖 **Para más información**: Ver [Sistema de Autenticación JWT](autenticacion-jwt.md)

### 4.2. Endpoints de Instancias (`/instances`)

| Método | Endpoint | Descripción | Payload |
| :--- | :--- | :--- | :--- |
| `POST` | `/instances` | Crear instancia | `{"instance_id": "string"}` |
| `GET` | `/instances` | Listar instancias | - |
| `GET` | `/instances/{id}` | Detalles de instancia | - |
| `DELETE` | `/instances/{id}` | Eliminar instancia | - |
| `POST` | `/instances/{id}/connect` | Iniciar conexión | - |
| `POST` | `/instances/{id}/disconnect` | Cerrar conexión | - |
| `GET` | `/instances/{id}/qr` | Obtener QR | - |
| `GET` | `/instances/{id}/status` | Estado de conexión | - |

### 4.3. Endpoints de Mensajería (`/instances/{id}/messages`)

| Método | Endpoint | Descripción | Payload (Resumido) |
| :--- | :--- | :--- | :--- |
| `POST` | `/text` | Enviar texto | `{"phone": "...", "message": "..."}` |
| `POST` | `/image` | Enviar imagen | `{"phone": "...", "image_url": "...", "caption": "..."}` |
| `POST` | `/video` | Enviar video | `{"phone": "...", "video_url": "...", "caption": "..."}` |
| `POST` | `/audio` | Enviar audio | `{"phone": "...", "audio_url": "..."}` |
| `POST` | `/document` | Enviar documento | `{"phone": "...", "document_url": "...", "filename": "..."}` |
| `POST` | `/react` | Reaccionar | `{"phone": "...", "message_id": "...", "emoji": "..."}` |
| `POST` | `/revoke` | Eliminar mensaje | `{"phone": "...", "message_id": "..."}` |
| `POST` | `/poll` | Crear encuesta | `{"phone": "...", "question": "...", "options": [...]}` |

### 4.4. Endpoints de Grupos (`/instances/{id}/groups`)

| Método | Endpoint | Descripción |
| :--- | :--- | :--- |
| `POST` | `/` | Crear grupo |
| `GET` | `/` | Listar grupos |
| `GET` | `/{groupID}` | Info del grupo |
| `PATCH` | `/{groupID}/participants` | Añadir/Remover participantes |
| `POST` | `/{groupID}/leave` | Salir del grupo |

### 4.5. Endpoints de Contactos (`/instances/{id}/contacts`)

| Método | Endpoint | Descripción |
| :--- | :--- | :--- |
| `POST` | `/check` | Verificar si números tienen WhatsApp |
| `GET` | `/` | Listar contactos sincronizados |
| `GET` | `/profile-picture` | Obtener URL de foto de perfil |
| `POST` | `/block` | Bloquear contacto |
| `POST` | `/unblock` | Desbloquear contacto |

### 4.6. Webhooks (`/instances/{id}/webhook`)

Permite configurar una URL para recibir eventos.

**Eventos soportados:**
*   `message`: Nuevo mensaje entrante.
*   `status`: Cambio de estado de conexión.
*   `receipt`: Confirmación de entrega/lectura.

**Payload de Webhook (Ejemplo Message):**
```json
{
  "event": "message",
  "instance_id": "test-1",
  "data": {
    "id": "MSG_ID",
    "from": "1234567890@s.whatsapp.net",
    "type": "text",
    "content": "Hola mundo",
    "timestamp": 1678900000
  }
}
```

---

## 5. Configuración y Despliegue

### 5.1. Variables de Entorno (.env)

| Variable | Descripción | Default |
| :--- | :--- | :--- |
| `API_PORT` | Puerto del servidor | `8080` |
| `DB_HOST` | Host de PostgreSQL | `localhost` |
| `DB_PORT` | Puerto de PostgreSQL | `5432` |
| `DB_USER` | Usuario de BD | `kero` |
| `DB_PASSWORD` | Contraseña de BD | `kero` |
| `DB_NAME` | Nombre de BD | `kero` |
| `REDIS_ADDR` | Dirección de Redis | `localhost:6379` |
| `REDIS_PASSWORD` | Contraseña de Redis | `` |
| `API_KEY` | Clave maestra para autenticación directa | - |
| `JWT_SECRET` | Secreto para firmar tokens JWT (dashboard) | - |

### 5.2. Despliegue con Docker

El proyecto incluye un `Dockerfile` optimizado (multi-stage build) y un `docker-compose.yml`.

**Comandos:**
```bash
# Construir y levantar
docker-compose up -d --build

# Ver logs
docker-compose logs -f

# Detener
docker-compose down
```

### 5.3. Ejecución Local (Desarrollo)

Requisitos: Go 1.21+, PostgreSQL, Redis.

```bash
# Instalar dependencias
go mod tidy

# Ejecutar migraciones (si aplica)
make migrate

# Iniciar servidor
make run
```

---

## 6. Testing

El proyecto cuenta con tests unitarios y de integración.

```bash
# Ejecutar todos los tests
make test

# Ver cobertura
go test -cover ./...
```

## 7. Seguridad

*   **Autenticación Dual:**
    *   **API Key**: Para acceso directo de máquina a máquina
    *   **JWT**: Para aplicaciones web con sesiones (tokens firmados con HMAC-SHA256)
*   **Validación:** Se validan todos los inputs JSON.
*   **Firma de Webhooks:** Los payloads de webhooks incluyen una firma HMAC-SHA256 para verificar su autenticidad (si se configura `WEBHOOK_SECRET`).
*   **Expiración de Tokens**: Los tokens JWT expiran automáticamente después de 24 horas.
*   **Rate Limiting**: Control de frecuencia de peticiones para prevenir abuso.

**Recomendaciones de Producción:**
- Usar `JWT_SECRET` aleatorio y seguro (generar con `openssl rand -base64 32`)
- Habilitar HTTPS para proteger tokens en tránsito
- Configurar CORS con dominios específicos
- Rotar secretos periódicamente
