# Documentación de la API de WhatsApp Kero-Kero

## 1. Visión General de la Arquitectura

Esta API está diseñada como una puerta de enlace para interactuar con WhatsApp a través de la librería `whatsmeow`. Su arquitectura sigue un patrón clásico de aplicación web en Go, separando las responsabilidades en capas bien definidas:

-   **Handlers (Manejadores):** Ubicados en `internal/handlers`, son responsables de recibir las peticiones HTTP, validar las entradas y llamar a los servicios correspondientes.
-   **Services (Servicios):** En `internal/services`, contienen la lógica de negocio principal de la aplicación. Orquestan las operaciones entre los repositorios y el `whatsapp.Manager`.
-   **Repositories (Repositorios):** En `internal/repository`, abstraen el acceso a las bases de datos (PostgreSQL/SQLite y Redis).
-   **WhatsApp Manager:** `internal/whatsapp/manager.go` es el núcleo de la aplicación. Es un componente centralizado que gestiona el ciclo de vida de todas las instancias (clientes) de `whatsmeow` en memoria.
-   **Queue Worker:** En `internal/services/queue_worker.go`, es un consumidor en segundo plano que procesa trabajos de una cola de Redis para manejar tareas asíncronas como el envío de mensajes.

---

## 2. Ciclo de Vida de una Instancia

El flujo para crear, conectar y autenticar una instancia no ha cambiado. Sigue el proceso de `POST /instances`, `POST /instances/{id}/connect`, y `GET /instances/{id}/qr`. La corrección del bug del "cliente zombie" asegura que este proceso sea robusto.

---

## 3. Flujo de Mensajes

El envío de mensajes ahora soporta dos modos: **Síncrono** (por defecto) y **Asíncrono** (recomendado para alto rendimiento).

### 3.1. Envío Síncrono

Este es el comportamiento por defecto si no se especifica lo contrario. Es más simple pero más lento.

1.  **Petición a la API:** Un cliente envía `POST /messages/text` sin ningún encabezado especial.
2.  **`MessageHandler`:** Recibe la petición.
3.  **`MessageService`:** Llama directamente a la función `client.WAClient.SendMessage`. La petición HTTP queda bloqueada esperando la respuesta de los servidores de WhatsApp.
4.  **Respuesta:** Una vez que WhatsApp confirma que el mensaje ha sido enviado, el `MessageService` responde con un código `200 OK` y el `message_id` final.

    ```json
    {
      "success": true,
      "message_id": "1A2B3C4D5E6F7G8H",
      "status": "sent"
    }
    ```

### 3.2. Envío Asíncrono (Nuevo)

Este modo está diseñado para alto rendimiento. La API responde inmediatamente y notifica el resultado final más tarde a través de un webhook.

1.  **Petición a la API:** Un cliente envía `POST /messages/text` incluyendo el encabezado **`X-Async: true`**.
2.  **`MessageHandler`:** Detecta el encabezado `X-Async`.
3.  **`MessageService`:**
    -   Valida la petición.
    -   Genera un **`correlation_id`** único.
    -   Crea un `MessageJob` que contiene toda la información de la petición y el `correlation_id`.
    -   Encola el trabajo en la cola de **Redis** (`kero_message_queue`).
4.  **Respuesta Inmediata:** La API responde inmediatamente con un código `202 Accepted` y el `correlation_id`. Esto libera al cliente sin tener que esperar a WhatsApp.

    ```json
    {
      "status": "queued",
      "correlation_id": "b7e7a8c8-f3b1-4f6e-a5b5-ae6f2f2f8a4e"
    }
    ```

5.  **Procesamiento en Segundo Plano (`QueueWorker`):**
    -   El worker, que se ejecuta en segundo plano, toma el trabajo de la cola de Redis.
    -   Intenta enviar el mensaje usando la misma lógica síncrona.
    -   Si el envío falla, lo reintenta (`MaxRetries` veces) con una espera exponencial. Si sigue fallando, lo mueve a una cola de "mensajes muertos" (`kero_dead_letter_queue`).

6.  **Notificación por Webhook (`message_ack`):**
    -   Una vez que el worker determina el resultado final (ya sea éxito o fallo definitivo), envía un nuevo evento de webhook a la URL configurada por el cliente.

    **Webhook en caso de ÉXITO:**
    ```json
    {
      "event": "message_ack",
      "instance_id": "your_instance_id",
      "timestamp": 1678886400,
      "data": {
        "status": "sent",
        "correlation_id": "b7e7a8c8-f3b1-4f6e-a5b5-ae6f2f2f8a4e",
        "message_id": "1A2B3C4D5E6F7G8H"
      }
    }
    ```

    **Webhook en caso de FALLO:**
    ```json
    {
      "event": "message_ack",
      "instance_id": "your_instance_id",
      "timestamp": 1678886400,
      "data": {
        "status": "failed",
        "correlation_id": "b7e7a8c8-f3b1-4f6e-a5b5-ae6f2f2f8a4e",
        "error": "Recipient is not a valid WhatsApp user"
      }
    }
    ```

El cliente utiliza el `correlation_id` para asociar esta notificación asíncrona con su petición original.
