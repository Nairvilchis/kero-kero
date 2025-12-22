# Plan de Implementación: Script de Pruebas CLI para Kero-Kero

Este plan describe la creación de una herramienta de línea de comandos (CLI) interactiva en Python que permitirá levantar el servidor Kero-Kero, realizar pruebas de integración reales contra una instancia de WhatsApp y gestionar el ciclo de vida de la aplicación desde una única terminal.

## Objetivos
1.  **Automatización**: Levantar y detener el servidor Go automáticamente.
2.  **Interactividad**: Menú amigable para ejecutar acciones sin usar herramientas externas como Postman.
3.  **Realismo**: Pruebas con conexión real (sin mocks) usando una instancia de WhatsApp.
4.  **Feedback**: Visualización de respuestas JSON y logs del servidor.

## Stack Tecnológico
*   **Lenguaje**: Python 3 (por su facilidad para scripting y manejo de HTTP/JSON).
*   **Librerías**:
    *   `requests`: Para realizar peticiones HTTP a la API.
    *   `subprocess`: Para ejecutar el servidor Go en segundo plano.
    *   `time`, `json`, `sys`: Utilidades del sistema.
    *   `qrcode` (opcional): Para renderizar el código QR en la terminal si la API devuelve el string del QR.

## Arquitectura del Script (`cli_tester.py`)

El script constará de tres componentes principales:

### 1. Gestor del Servidor (`ServerManager`)
Clase responsable de:
*   Iniciar el servidor (`go run cmd/server/main.go`) en un subproceso.
*   Monitorear la salida estándar (logs) en tiempo real (opcional) o bajo demanda.
*   Verificar la disponibilidad del servidor (Health Check) antes de iniciar el menú.
*   Asegurar el cierre limpio del proceso al salir.

### 2. Cliente API (`ApiClient`)
Clase que encapsula los endpoints de Kero-Kero:
*   `create_instance()`: POST /instances
*   `get_qr()`: GET /instances/{id}/qr (y renderizado en terminal)
*   `send_message(type, data)`: POST /messages/{type}
*   `get_contacts()`: GET /contacts
*   Manejo de headers (API Key) y errores HTTP.

### 3. Interfaz de Menú (`Menu`)
Bucle principal que ofrece las siguientes opciones:

*   **1. Gestión del Servidor**
    *   Ver estado (Health)
    *   Ver últimos logs
*   **2. Gestión de Instancias**
    *   Crear nueva instancia
    *   **Conectar (Ver QR)**: Esta es la función clave. Si la API devuelve el código QR en base64 o texto, intentaremos renderizarlo en la terminal o guardar la imagen temporalmente.
    *   Ver estado de instancia
    *   Desconectar/Eliminar
*   **3. Mensajería (Tests Reales)**
    *   Enviar Texto
    *   Enviar Imagen (pidiendo path local)
    *   Enviar Video/Audio
*   **4. Contactos y Grupos**
    *   Listar contactos
    *   Listar grupos
*   **0. Salir**: Apaga el servidor y cierra el script.

## Flujo de Trabajo

1.  El usuario ejecuta `python3 cli_tester.py`.
2.  El script valida variables de entorno (ej. puerto 8080).
3.  Se lanza el servidor Go en background. Se muestra un spinner "Esperando servidor...".
4.  Una vez el servidor responde OK a `/health`, aparece el Menú Principal.
5.  El usuario selecciona "Crear Instancia" -> Se muestra el ID.
6.  El usuario selecciona "Conectar" -> Se muestra el QR en terminal. El usuario lo escanea con su celular real.
7.  El usuario selecciona "Enviar Texto" -> Ingresa número y mensaje -> Se envía realmente.
8.  Al finalizar, selecciona "Salir" y todo se limpia.

## Próximos Pasos
1.  Crear el archivo `cli_tester.py` con la estructura base.
2.  Implementar el `ServerManager` para correr el backend.
3.  Implementar las funciones de `ApiClient` conectadas a los endpoints reales analizados en `internal/routes`.
4.  Construir el menú interactivo.
