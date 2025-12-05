# Kero-Kero WhatsApp API - Endpoints

## 📋 Endpoints Públicos

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/` | Información del servicio |
| `GET` | `/health` | Estado de salud del servidor |

---

## 📱 Gestión de Instancias

| Método | Ruta | Descripción |
|--------|------|-------------|
| `POST` | `/instances` | Crear nueva instancia |
| `GET` | `/instances` | Listar todas las instancias |
| `GET` | `/instances/{id}` | Obtener detalles de instancia |
| `DELETE` | `/instances/{id}` | Eliminar instancia |
| `POST` | `/instances/{id}/connect` | Conectar instancia |
| `POST` | `/instances/{id}/disconnect` | Desconectar instancia |
| `GET` | `/instances/{id}/qr` | Obtener código QR (PNG) |
| `GET` | `/instances/{id}/status` | Consultar estado |

---

## 💬 Mensajería

| Método | Ruta | Descripción |
|--------|------|-------------|
| `POST` | `/instances/{id}/messages/text` | Enviar mensaje de texto |
| `POST` | `/instances/{id}/messages/image` | Enviar imagen |
| `POST` | `/instances/{id}/messages/video` | Enviar video |
| `POST` | `/instances/{id}/messages/audio` | Enviar audio |
| `POST` | `/instances/{id}/messages/document` | Enviar documento |
| `POST` | `/instances/{id}/messages/location` | Enviar ubicación |
| `POST` | `/instances/{id}/messages/react` | Reaccionar a mensaje |
| `POST` | `/instances/{id}/messages/revoke` | Eliminar mensaje (para todos) |
| `POST` | `/instances/{id}/messages/download` | Descargar archivo multimedia |
| `POST` | `/instances/{id}/messages/poll` | Crear encuesta |
| `POST` | `/instances/{id}/messages/poll/vote` | Votar en encuesta |

---

## 👥 Grupos

| Método | Ruta | Descripción |
|--------|------|-------------|
| `POST` | `/instances/{id}/groups` | Crear grupo |
| `GET` | `/instances/{id}/groups` | Listar grupos |
| `GET` | `/instances/{id}/groups/{groupID}` | Obtener info del grupo |
| `PATCH` | `/instances/{id}/groups/{groupID}/participants` | Gestionar participantes |
| `POST` | `/instances/{id}/groups/{groupID}/leave` | Salir del grupo |

---

## 📇 Contactos y Presencia

| Método | Ruta | Descripción |
|--------|------|-------------|
| `POST` | `/instances/{id}/contacts/check` | Verificar números en WhatsApp |
| `GET` | `/instances/{id}/contacts` | Listar contactos sincronizados |
| `GET` | `/instances/{id}/contacts/profile-picture` | Obtener foto de perfil |
| `POST` | `/instances/{id}/contacts/presence/subscribe` | Suscribirse a presencia |
| `POST` | `/instances/{id}/contacts/block` | Bloquear contacto |
| `POST` | `/instances/{id}/contacts/unblock` | Desbloquear contacto |

---

## 💬 Chats y Estado

| Método | Ruta | Descripción |
|--------|------|-------------|
| `POST` | `/instances/{id}/chats/status` | Actualizar estado (About) |
| `POST` | `/instances/{id}/chats/archive` | Archivar chat (WIP) |

---

## 📱 Estados (Historias)

| Método | Ruta | Descripción |
|--------|------|-------------|
| `POST` | `/instances/{id}/status` | Publicar estado de texto |
| `GET` | `/instances/{id}/status/privacy` | Obtener privacidad de estados |

---

## 📞 Llamadas

| Método | Ruta | Descripción |
|--------|------|-------------|
| `POST` | `/instances/{id}/calls/reject` | Rechazar llamada (WIP) |
| `GET` | `/instances/{id}/calls/settings` | Obtener configuración de llamadas |
| `PUT` | `/instances/{id}/calls/settings` | Actualizar configuración de llamadas |

---

## 🔒 Privacidad

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/instances/{id}/privacy` | Obtener configuración de privacidad |
| `PATCH` | `/instances/{id}/privacy` | Actualizar configuración |

---

## 🔔 Webhooks

| Método | Ruta | Descripción |
|--------|------|-------------|
| `POST` | `/instances/{id}/webhook` | Configurar webhook |
| `GET` | `/instances/{id}/webhook` | Obtener configuración de webhook |
| `DELETE` | `/instances/{id}/webhook` | Eliminar webhook |

### Eventos de Webhook

Los webhooks pueden recibir los siguientes eventos:

- **message**: Mensaje recibido (texto, imagen, video, audio, documento, ubicación)
- **status**: Cambio de estado (connected, disconnected, logged_out)
- **receipt**: Confirmación de lectura/entrega

---

## 🔐 Autenticación

Todos los endpoints (excepto `/` y `/health`) requieren autenticación mediante API Key:

```bash
# Header
X-API-Key: tu_api_key_aqui

# O Authorization Bearer
Authorization: Bearer tu_api_key_aqui
```

---

## 📝 Ejemplos de Uso

### Crear Instancia
```bash
curl -X POST http://localhost:8080/instances \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"instance_id": "mi-instancia"}'
```

### Obtener QR
```bash
curl http://localhost:8080/instances/mi-instancia/qr \
  -H "X-API-Key: your-api-key" \
  --output qr.png
```

### Enviar Mensaje
```bash
curl -X POST http://localhost:8080/instances/mi-instancia/messages/text \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5215512345678",
    "message": "Hola desde Kero-Kero!"
  }'
```

### Configurar Webhook
```bash
curl -X POST http://localhost:8080/instances/mi-instancia/webhook \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://tu-servidor.com/webhook",
    "events": ["message", "status", "receipt"],
    "secret": "tu-secreto-para-firmar"
  }'
```

### Reaccionar a un mensaje
```bash
curl -X POST http://localhost:8080/instances/mi-instancia/messages/react \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5215512345678",
    "message_id": "ID_DEL_MENSAJE",
    "emoji": "👍"
  }'
```

### Eliminar un mensaje
```bash
curl -X POST http://localhost:8080/instances/mi-instancia/messages/revoke \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5215512345678",
    "message_id": "ID_DEL_MENSAJE"
  }'
```

### Descargar archivo multimedia
```bash
curl -X POST http://localhost:8080/instances/mi-instancia/messages/download \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "image",
    "url": "https://mmg.whatsapp.net/...",
    "direct_path": "/v/...",
    "media_key": "BASE64_ENCODED_KEY",
    "file_enc_sha256": "BASE64_ENCODED_SHA",
    "file_sha256": "BASE64_ENCODED_SHA",
    "file_length": 12345,
    "mimetype": "image/jpeg"
  }' \
  --output imagen.jpg
```

> **Nota**: Los datos de descarga (url, media_key, etc.) se obtienen del webhook cuando recibes un mensaje con multimedia.

### Crear una encuesta
```bash
curl -X POST http://localhost:8080/instances/mi-instancia/messages/poll \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5215512345678",
    "question": "¿Cuál es tu lenguaje favorito?",
    "options": ["Go", "Python", "JavaScript", "Rust"],
    "selectable_count": 1
  }'
```

### Votar en una encuesta
```bash
curl -X POST http://localhost:8080/instances/mi-instancia/messages/poll/vote \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5215512345678",
    "message_id": "ID_DE_LA_ENCUESTA",
    "option_names": ["Go"]
  }'
```

### Publicar un estado de texto
```bash
curl -X POST http://localhost:8080/instances/mi-instancia/status \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "text",
    "content": "¡Hola desde Kero-Kero! 🚀"
  }'
```

### Configurar auto-rechazo de llamadas
```bash
curl -X PUT http://localhost:8080/instances/mi-instancia/calls/settings \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "auto_reject": true
  }'
```

---

## 🚀 Características Implementadas

✅ Gestión completa de instancias  
✅ Mensajería multimedia (texto, imagen, video, audio, documento, ubicación)  
✅ Gestión de grupos (crear, listar, participantes)  
✅ Contactos y presencia  
✅ Configuración de privacidad  
✅ Sistema de webhooks con firma HMAC-SHA256  
✅ Rate limiting  
✅ CORS configurable  
✅ Logging estructurado  
✅ Manejo de errores estandarizado  

---

## 📦 Próximas Mejoras

- [ ] Documentación Swagger/OpenAPI
- [ ] Tests unitarios y de integración
- [ ] Métricas y monitoreo (Prometheus)
- [ ] Soporte para stickers
- [ ] Envío de mensajes programados
- [ ] Respuestas automáticas
- [ ] Dockerización completa
