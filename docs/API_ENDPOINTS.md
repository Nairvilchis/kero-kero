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

Todos los endpoints de envío de mensajes soportan el encabezado `X-Async: true` para un comportamiento asíncrono.

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
- **message_ack**: **(NUEVO)** Confirmación de un mensaje enviado de forma asíncrona. Indica si el envío fue exitoso (`sent`) o falló (`failed`).

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

### Enviar Mensaje (Modo Síncrono - por defecto)
```bash
curl -X POST http://localhost:8080/instances/mi-instancia/messages/text \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5215512345678",
    "message": "Hola desde Kero-Kero!"
  }'
```
**Respuesta (200 OK):**
```json
{
  "success": true,
  "message_id": "1A2B3C4D5E6F7G8H",
  "status": "sent"
}
```

### Enviar Mensaje (Modo Asíncrono - NUEVO)
```bash
curl -X POST http://localhost:8080/instances/mi-instancia/messages/text \
  -H "X-API-Key: your-api-key" \
  -H "X-Async: true" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5215512345678",
    "message": "Hola desde Kero-Kero!"
  }'
```
**Respuesta Inmediata (202 Accepted):**
```json
{
  "status": "queued",
  "correlation_id": "b7e7a8c8-f3b1-4f6e-a5b5-ae6f2f2f8a4e"
}
```
Más tarde, recibirás un webhook con el evento `message_ack` y este `correlation_id`.

---
... (resto de ejemplos sin cambios) ...
---

## 🚀 Características Implementadas

✅ Gestión completa de instancias  
✅ Mensajería multimedia (texto, imagen, video, audio, documento, ubicación)  
✅ **Envío de mensajes asíncrono y síncrono**
✅ Gestión de grupos (crear, listar, participantes)  
✅ Contactos y presencia  
✅ Configuración de privacidad  
✅ Sistema de webhooks con firma HMAC-SHA256  
✅ Rate limiting  
✅ CORS configurable  
✅ Logging estructurado  
✅ Manejo de errores estandarizado
