# Registro de Cambios

Este documento registra los cambios, correcciones y mejoras realizadas en el servidor Kero-Kero.

## [En Desarrollo] - 2025-12-10

### 🐛 Corregido
- **Gestión de Conexiones Zombie**: Se solucionó un error crítico donde las instancias quedaban atrapadas en un estado "conectado pero no autenticado" si el código QR no se escaneaba a tiempo.
  - Se modificó `InstanceService.ConnectInstance` (internal/services/instance_service.go) para verificar tanto `IsConnected()` como `IsLoggedIn()`.
  - Se implementó una desconexión forzada si la instancia está conectada al socket pero no logueada al intentar conectar de nuevo. Esto limpia el estado y permite generar un nuevo código QR correctamente.

### 🔧 Mejoras
- **Soporte de PostgreSQL para whatsmeow**: Ahora el almacenamiento de sesiones de WhatsApp (`DeviceStore`) puede utilizar PostgreSQL si la aplicación está configurada con este driver. Esto centraliza la identidad de las sesiones, facilitando la escalabilidad y la recuperación ante fallos del contenedor.

### ⚡ Arquitectura
- **Estado de Conexión en Redis**: Se implementó la sincronización del estado de las instancias (`connected`, `authenticated`, `disconnected`) en Redis.
  - Esto permite que cualquier nodo del cluster conozca el estado real de una instancia, independientemente de qué nodo la esté gestionando inicialmente.
  - Se refactorizó `InstanceService.GetStatus` para priorizar la consulta a Redis sobre la memoria local, paso fundamental para escalar horizontalmente.
- **Cola de Mensajería Asíncrona**: Se implementó un sistema de workers para el envío de mensajes.
  - Los mensajes ahora pueden ser procesados en segundo plano `QueueService`, desacoplando la recepción de la petición HTTP del proceso de envío real en WhatsApp.
  - Esto mejora la resiliencia y evita bloqueos en el servidor web si la red de WhatsApp está lenta.
