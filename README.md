# 🐸 Kero-Kero API

**Tu puerta de entrada profesional a la automatización de WhatsApp.**

Kero-Kero es una API REST potente, escalable y fácil de usar que te permite integrar WhatsApp en tus aplicaciones, CRMs y sistemas de soporte. Construida con tecnología de vanguardia en Go, ofrece un rendimiento excepcional y una gestión robusta de múltiples sesiones.

---

## ✨ Características Principales

*   **🚀 Multi-Instancia:** Gestiona cientos de cuentas de WhatsApp desde un solo servidor.
*   **💬 Mensajería Completa:** Envía texto, imágenes, videos, audios, documentos, ubicaciones y reacciones.
*   **🤖 Automatización:** Webhooks en tiempo real para mensajes entrantes y eventos de estado.
*   **👥 Gestión de Grupos:** Crea grupos, añade participantes y administra comunidades programáticamente.
*   **🔒 Privacidad y Seguridad:** Control total sobre la configuración de privacidad y bloqueo de contactos.
*   **📊 Encuestas:** Crea y gestiona encuestas nativas de WhatsApp.
*   **🐳 Docker Ready:** Despliegue instantáneo con contenedores optimizados.

---

## 🚀 Inicio Rápido

La forma más sencilla de empezar es usando Docker Compose.

### Requisitos
*   Docker y Docker Compose instalados.

### Pasos

1.  **Clona el repositorio:**
    ```bash
    git clone https://github.com/tu-usuario/kero-kero.git
    cd kero-kero
    ```

2.  **Inicia los servicios:**
    ```bash
    docker-compose up -d
    ```

3.  **¡Listo!** La API estará disponible en `http://localhost:8080`.

### Autenticación

Kero-Kero usa un sistema de autenticación dual:

- **API Key**: Para acceso directo a la API (configurar en `.env`)
- **JWT**: Para el dashboard (el usuario se autentica con API Key y recibe un token JWT)

**Configurar las claves secretas** en tu archivo `.env.local`:

```bash
# API Key para autenticación directa
API_KEY=tu-clave-secreta-aqui

# JWT Secret para tokens del dashboard
JWT_SECRET=tu-secreto-jwt-aqui  # Generar con: openssl rand -base64 32
```

⚠️ **Importante**: Cambia estas claves en producción por valores aleatorios y seguros.

---

## 📖 Uso Básico

### 1. Crear una Instancia
```bash
curl -X POST http://localhost:8080/instances \
  -H "Content-Type: application/json" \
  -d '{"instance_id": "mi-empresa"}'
```

### 2. Obtener el QR para conectar
```bash
curl http://localhost:8080/instances/mi-empresa/qr --output qr.png
```
*Escanea el código QR generado con tu aplicación de WhatsApp.*

### 3. Enviar un Mensaje
```bash
curl -X POST http://localhost:8080/instances/mi-empresa/messages/text \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5215512345678",
    "message": "¡Hola desde Kero-Kero! 🐸"
  }'
```

---

## 📚 Documentación Técnica

Para una guía profunda sobre la arquitectura, configuración avanzada, referencia completa de endpoints y esquemas de base de datos, consulta nuestra documentación técnica:

👉 **[Documentación Técnica Completa](docs/TECHNICAL_DOCUMENTATION.md)**

### Guías Específicas

- **[Sistema de Autenticación JWT](docs/autenticacion-jwt.md)** - Cómo funciona el login y la seguridad

---

## 🛠️ Stack Tecnológico

*   **Lenguaje:** Go (Golang)
*   **Core:** whatsmeow
*   **Base de Datos:** PostgreSQL
*   **Caché:** Redis

---

## 📄 Licencia

Este proyecto está bajo la licencia MIT.
