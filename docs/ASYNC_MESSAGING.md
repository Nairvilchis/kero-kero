# Guía de Uso: Envío Asíncrono de Mensajes

## Introducción

El servidor Kero-Kero ahora soporta dos modos de envío de mensajes:

1. **Modo Síncrono (Por defecto):** El servidor espera la confirmación de WhatsApp antes de responder.
2. **Modo Asíncrono (Opcional):** El mensaje se encola y se procesa en segundo plano.

## ¿Cuándo usar cada modo?

### Modo Síncrono (Recomendado para la mayoría de casos)
✅ **Usa este modo cuando:**
- Necesitas certeza inmediata del envío (ej. confirmaciones de pago)
- Estás usando N8N u otras herramientas de automatización
- El flujo depende del `message_id` de WhatsApp
- Envías mensajes transaccionales críticos

### Modo Asíncrono (Para casos específicos)
✅ **Usa este modo cuando:**
- Envías mensajes masivos (newsletters, avisos grupales)
- La velocidad de respuesta es más importante que la confirmación inmediata
- Tienes un sistema de webhooks configurado para recibir confirmaciones
- Quieres evitar timeouts en envíos de medios pesados

## Cómo activar el Modo Asíncrono

Simplemente añade el header `X-Async: true` a tu petición HTTP:

### Ejemplo con cURL

```bash
# Modo Síncrono (comportamiento normal)
curl -X POST http://localhost:8080/instances/mi-instancia/messages/text \
  -H "X-API-Key: tu-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5215512345678",
    "message": "Hola desde Kero-Kero"
  }'

# Modo Asíncrono (con header X-Async)
curl -X POST http://localhost:8080/instances/mi-instancia/messages/text \
  -H "X-API-Key: tu-api-key" \
  -H "X-Async: true" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5215512345678",
    "message": "Hola desde Kero-Kero"
  }'
```

### Ejemplo con N8N

En el nodo HTTP Request de N8N:
1. Configura tu endpoint normalmente
2. En la sección "Headers":
   - Name: `X-Async`
   - Value: `true`

### Ejemplo con JavaScript/Fetch

```javascript
// Modo Síncrono
const response = await fetch('http://localhost:8080/instances/mi-instancia/messages/text', {
  method: 'POST',
  headers: {
    'X-API-Key': 'tu-api-key',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    phone: '5215512345678',
    message: 'Hola desde Kero-Kero'
  })
});

// Modo Asíncrono
const response = await fetch('http://localhost:8080/instances/mi-instancia/messages/text', {
  method: 'POST',
  headers: {
    'X-API-Key': 'tu-api-key',
    'X-Async': 'true',  // ← Añade este header
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    phone: '5215512345678',
    message: 'Hola desde Kero-Kero'
  })
});
```

## Respuestas del Servidor

### Modo Síncrono
```json
{
  "success": true,
  "message_id": "3EB0C2FE7A4B5C8D9E0F",
  "status": "sent"
}
```
**Status Code:** `200 OK`

### Modo Asíncrono
```json
{
  "success": true,
  "message_id": "msg_1702345678901234567",
  "status": "queued"
}
```
**Status Code:** `202 Accepted`

⚠️ **Nota:** El `message_id` en modo asíncrono es un ID de cola, NO el ID de WhatsApp. El ID real de WhatsApp se generará cuando el worker procese el mensaje.

## Endpoints Soportados

Los siguientes endpoints soportan el header `X-Async`:

- ✅ `POST /instances/{id}/messages/text`
- ✅ `POST /instances/{id}/messages/image`
- ✅ `POST /instances/{id}/messages/video`
- ✅ `POST /instances/{id}/messages/audio`
- ✅ `POST /instances/{id}/messages/document`
- ✅ `POST /instances/{id}/messages/location`

## Monitoreo de la Cola

Para verificar el estado de la cola de mensajes, puedes consultar Redis directamente:

```bash
# Longitud de la cola
redis-cli LLEN queue:messages

# Ver mensajes pendientes (sin removerlos)
redis-cli LRANGE queue:messages 0 -1
```

## Preguntas Frecuentes

**P: ¿Qué pasa si el envío falla en modo asíncrono?**  
R: Actualmente, el error se registra en los logs del servidor. En futuras versiones se implementará un sistema de reintentos y notificaciones vía webhook.

**P: ¿Cuántos workers procesan la cola?**  
R: Por defecto, 3 workers. Esto se puede configurar en el código (`queue_service.go`).

**P: ¿Puedo usar webhooks para saber cuándo se envió el mensaje?**  
R: Sí, configura un webhook en tu instancia. Recibirás el evento cuando el mensaje se envíe efectivamente.

**P: ¿El modo asíncrono es más rápido?**  
R: Sí, la respuesta HTTP es inmediata (milisegundos). El envío real ocurre en segundo plano.
