# 📟 Guía de Uso: Kero-Kero CLI Tester

Esta herramienta de línea de comandos (`cli_tester.py`) permite levantar el servidor Kero-Kero, gestionar instancias de WhatsApp y probar **el 100% de los endpoints** de la API sin necesidad de instalar Postman ni configurar entornos complejos.

## 📋 Requisitos Previos

Asegúrate de tener instalado lo siguiente en tu sistema:

1.  **Go** (v1.22+): Para ejecutar el servidor backend.
2.  **Python 3** (v3.8+): Para ejecutar el script de pruebas.
3.  **Librería `requests`**:
    ```bash
    pip install requests
    ```

## 🚀 Inicio Rápido

1.  Abre una terminal en la raíz del proyecto `kero-kero-server`.
2.  Ejecuta el script:
    ```bash
    ./cli_tester.py
    # O alternativamente:
    python3 cli_tester.py
    ```
3.  El script iniciará automáticamente el servidor (`go run ...`) y esperará a que esté listo (`Health Check OK`).

---

## 🎮 Menú Principal

El CLI está organizado en módulos temáticos:

### 1. Instancias & Privacidad
*   **Gestión Básica**: Crear (`Create`), Listar, Eliminar instancias.
*   **Conexión**:
    *   Selecciona "Conectar / Ver QR".
    *   El script descargará el QR y lo abrirá automáticamente con tu visor de imágenes.
    *   Escanéalo con WhatsApp.
*   **Configuración**: Ajustar Webhooks, Privacidad (Last Seen), y Rechazo de Llamadas.
*   **Sincronización**: Forzar la descarga del historial de chats.

### 2. Mensajería
*   **Envío Básico**: Texto plano o con simulación de "Escribiendo...".
*   **Multimedia**: Enviar Imágenes, Videos, Audios o Documentos (usando URLs públicas).
*   **Interactivo**:
    *   **Encuestas (Polls)**: Crea encuestas con opciones múltiples.
    *   **Reacciones**: Reacciona con emojis a mensajes (necesitas el ID del mensaje).
    *   **Edición**: Corrige mensajes de texto ya enviados.

### 3. Automatización & Business
*   **Mensajes Masivos**: Envía un mismo mensaje a múltiples números con un *delay* de seguridad.
*   **Auto-Respuestas**: Configura respuestas simples basadas en palabras clave.
*   **Etiquetas**: Crea y gestiona etiquetas para organizar chats (Business API).

### 4. Grupos Avanzado
*   Crear grupos, obtener enlaces de invitación, unirse mediante código, y gestionar permisos de administración.

### 5. Extras
*   **Estados**: Publica "Historias" de color.
*   **Canales**: Crea Newsletters/Canales de difusión.

---

## 💡 Flujos de Prueba Comunes

### A. Crear y Conectar una Instancia
1.  Ve a `Instancias` > `Crear Nueva`.
2.  Ingresa un ID simple (ej: `test1`).
3.  Ve a `Conectar / Ver QR`.
4.  Ingresa el ID `test1`.
5.  Escanea el QR que aparecerá.
6.  Regresa al Menú Principal y selecciona la opción `7` para fijar `test1` como tu instancia activa.

### B. Enviar un Mensaje de Prueba
1.  Asegúrate de tener una instancia activa (se ve en la cabecera del menú).
2.  Ve a `Mensajería` > `Texto Plano`.
3.  Ingresa el número destino en formato internacional (ej: `5215512345678`).
4.  Escribe el mensaje y pulsa Enter.
5.  Verás la respuesta JSON de la API confirmando el envío.

### C. Crear una Encuesta
1.  Ve a `Mensajería` > `Crear ENCUESTA`.
2.  Destinatario: Tu número o un grupo.
3.  Título: "¿Qué cenamos hoy?".
4.  Opciones: `Pizza,Tacos,Sushi` (separadas por coma).
5.  Max respuestas: `1`.

---

## ⚠️ Solución de Problemas

*   **Error "Connection Refused"**: El servidor Go no pudo iniciar. Revisa si ya tienes algo corriendo en el puerto 8080.
*   **El QR no se abre**: Busca el archivo `current_qr.png` en la carpeta del proyecto y ábrelo manualmente.
*   **No puedo enviar mensajes**: Verifica en `Ver Estado Detallado` que el status sea `connected`.

---

**Nota**: Al seleccionar "Salir (0)", el script detendrá automáticamente el proceso del servidor Go para liberar recursos.
