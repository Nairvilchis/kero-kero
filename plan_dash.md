# 📘 Guía Maestra de Desarrollo: Dashboard Kero-Kero (Next.js + App Router)

Esta guía detalla la implementación de un Dashboard profesional para la gestión de instancias de WhatsApp utilizando la API Server de Kero-Kero.

## 🛠️ Stack Tecnológico

*   **Framework**: Next.js 14+ (App Router)
*   **UI Library**: shadcn/ui (basado en Radix UI + Tailwind CSS)
*   **Styling**: Tailwind CSS
*   **State Management**:
    *   **Server State (Data Fetching)**: TanStack Query (React Query) v5
    *   **Client State (UI/Global)**: Zustand
*   **Formularios**: React Hook Form + Zod
*   **HTTP Client**: Axios (configurado con interceptors)
*   **Iconos**: Lucide React

---

## 📡 Cliente API y Autenticación

La API utiliza autenticación vía Header `X-Api-Key` o JWT bearer (según configuración). El dashboard debe configurarse para manejar ambas estrategias, priorizando la seguridad.

### Configuración Axios (`lib/api.ts`)
```typescript
import axios from 'axios';

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use((config) => {
  // Opción A: API Key global desde env (modo admin único)
  if (process.env.NEXT_PUBLIC_API_KEY) {
     config.headers['X-Api-Key'] = process.env.NEXT_PUBLIC_API_KEY;
  }
  
  // Opción B: JWT desde sesión (modo multi-usuario)
  // const token = useAuthStore.getState().token;
  // if (token) config.headers['Authorization'] = `Bearer ${token}`;
  
  return config;
});

export default api;
```

---

## 🗺️ Mapa Completo de Endpoints API

A continuación, se listan **todas** las rutas disponibles extraídas del código fuente del servidor.
*Nota: La mayoría de las rutas requieren el prefijo `/instances/{instanceID}`.*

### 1. Gestión de Instancias (`/instances`)
| Método | Ruta | Descripción | Payload / Params |
| :--- | :--- | :--- | :--- |
| `POST` | `/instances` | Crear nueva instancia | `{ "instance_id": "nombre", "sync_history": bool }` |
| `GET` | `/instances` | Listar todas | - |
| `GET` | `/{id}` | Detalles instancia | - |
| `PUT` | `/{id}` | Actualizar config (webhook) | `{ "webhook_url": "...", "events": [...] }` |
| `DELETE` | `/{id}` | Eliminar instancia | - |
| `POST` | `/{id}/connect` | Iniciar conexión | - |
| `POST` | `/{id}/disconnect` | Cerrar sesión | - |
| `GET` | `/{id}/qr` | Obtener QR (Base64) | - |
| `GET` | `/{id}/status` | Estado actual | Response: `{ "status": "connected"|"disconnected" }` |

### 2. Mensajería (`/instances/{id}/messages`)
Todos los endpoints son POST.

| Ruta Suffix | Descripción | Payload JSON |
| :--- | :--- | :--- |
| `/text` | Texto simple | `{ "to": "521...", "message": "Hola" }` |
| `/text-with-typing` | Texto con simulación | `{ "to": "...", "message": "...", "duration": 2 }` |
| `/image` | Enviar Imagen | `{ "to": "...", "url": "http...", "caption": "..." }` |
| `/video` | Enviar Video | `{ "to": "...", "url": "...", "caption": "..." }` |
| `/audio` | Enviar Audio (PTT) | `{ "to": "...", "url": "..." }` |
| `/document` | Enviar Documento | `{ "to": "...", "url": "...", "filename": "doc.pdf" }` |
| `/location` | Enviar Ubicación | `{ "to": "...", "latitude": 0.0, "longitude": 0.0 }` |
| `/contact` | Enviar VCard | `{ "to": "...", "vcard": "BEGIN:VCARD..." }` |
| `/react` | Reaccionar | `{ "message_id": "...", "reaction": "👍" }` |
| `/revoke` | Eliminar para todos | `{ "message_id": "..." }` |
| `/edit` | Editar mensaje | `{ "message_id": "...", "new_text": "..." }` |
| `/mark-read` | Marcar leído | `{ "chat_jid": "...", "message_id": "..." }` |

### 3. Chats e Historial (`/instances/{id}/chats`)
Fundamental para la vista tipo "WhatsApp Web".

| Método | Ruta | Descripción |
| :--- | :--- | :--- |
| `GET` | `/` | Listar chats recientes (Inbox) |
| `GET` | `/{jid}/messages` | Obtener historial de mensajes de un chat |
| `DELETE` | `/{jid}` | Eliminar chat |
| `POST` | `/{jid}/read` | Marcar chat completo como leído |
| `POST` | `/archive` | Archivar chat |
| `POST` | `/pin` | Fijar chat |
| `POST` | `/mute` | Silenciar chat |

### 4. Contactos (`/instances/{id}/contacts`)
| Método | Ruta | Descripción | Payload |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | Listar contactos guardados | `?page=1&limit=50` |
| `POST` | `/check` | Verificar si tienen WhatsApp | `{ "phones": ["..."] }` |
| `GET` | `/{phone}` | Info detallada | - |
| `GET` | `/{phone}/profile-picture` | URL foto perfil | - |
| `POST` | `/block` | Bloquear usuario | `{ "phone": "..." }` |
| `POST` | `/unblock` | Desbloquear | `{ "phone": "..." }` |

### 5. Grupos (`/instances/{id}/groups`)
| Método | Ruta | Descripción |
| :--- | :--- | :--- |
| `GET` | `/` | Listar grupos |
| `POST` | `/` | Crear grupo `{ "subject": "...", "participants": [] }` |
| `GET` | `/{gid}` | Info grupo (metadatos) |
| `GET` | `/{gid}/invite` | Obtener enlace invitación |
| `POST` | `/join` | Unirse vía enlace |
| `POST` | `/{gid}/participants` | Agregar participantes |
| `POST` | `/{gid}/leave` | Salir del grupo |

### 6. Automatización y Negocio
*   **Automation** (`/automation`):
    *   `POST /bulk-message`: Envío masivo.
    *   `POST /auto-reply`: Configurar autorespuesta simple.
*   **Business** (`/business`):
    *   `GET /profile`: Perfil de negocio.
    *   `POST /labels`: Gestión de etiquetas.
    *   `GET/POST /autolabel/rules`: Reglas para etiquetar chats automáticamente.

---

## 🏗️ Estructura del Proyecto Next.js

```
app/
├── (auth)/                 # Layout de autenticación (si aplica)
│   └── login/
├── (dashboard)/            # Layout principal con Sidebar
│   ├── layout.tsx          # Provider de estado, Sidebar, Header
│   ├── page.tsx            # Dashboard Home (Vista General)
│   ├── instances/
│   │   ├── page.tsx        # Lista de instancias (Cards)
│   │   └── new/            # Crear instancia
│   └── [instanceId]/       # Rutas dependientes de instancia
│       ├── chat/           # 💬 CLAVE: Interfaz de Chat
│       │   └── page.tsx
│       ├── contacts/       # Agenda
│       │   └── page.tsx
│       ├── campaigns/      # Envíos masivos
│       │   └── page.tsx
│       ├── automation/     # Autorespuestas y Reglas
│       │   └── page.tsx
│       └── settings/       # Webhooks, Perfil, Privacidad
│           └── page.tsx
├── layout.tsx              # Root Layout
└── globals.css
```

---

## 🧩 Componentes Clave Sugeridos

### 1. `InstanceGuard` (Layout)
Componente que envuelve `app/[instanceId]/...` para:
*   Validar que la instancia existe.
*   Verificar su estado (`/status`).
*   Mostrar un "DisconnectedOverlay" si la instancia no está conectada, impidiendo interactuar con módulos que requieren conexión (chats, mensajes).

### 2. `ChatInterface` (Compositor Complejo)
Ubicado en `/chat`. Debe replicar la experiencia de WhatsApp Web:
*   **Sidebar Izquierdo**: Lista virtualizada de Chats (`GET /chats`).
    *   Buscador.
    *   Filtros (No leídos, Grupos).
*   **Panel Derecho**: Lista de mensajes (`GET /chats/{jid}/messages`).
    *   Scroll infinito inverso.
    *   **WebSocket Listener**: Escuchar eventos `message` entrantes para hacer append real-time sin re-fetch.
*   **Input Area**:
    *   Soporte para emoji picker.
    *   Upload de archivos (Drag & Drop) -> Llama a endpoints `/image`, `/document`, etc.
    *   Grabadora de voz -> endpoint `/audio`.

### 3. `QRCodeScanner`
Componente que hace polling a `GET /{id}/qr` o usa WebSocket (si disponible) para mostrar el código QR. Debe manejar expiración y autoreload.

### 4. `CampaignWizard`
Formulario por pasos para `/automation/bulk-message`:
1.  **Selección**: Elegir contactos (desde lista o CSV upload).
2.  **Composición**: Escribir mensaje / media.
3.  **Programación**: Definir delay aleatorio (importante para evitar bloqueos).
4.  **Resumen**: Confirmar envío.

---

## ⚡ Estrategia de Sincronización (Real-Time)

El dashboard debe sentirse "vivo".
1.  **WebSockets**: Si el servidor expone WS en `/ws`:
    *   Conectar al abrir el dashboard.
    *   Escuchar eventos:
        *   `connection.update`: Actualizar estado de instancia (QR -> Connecting -> Connected).
        *   `messages.upsert`: Nuevo mensaje -> Actualizar caché de React Query (`["chats", jid]`) e insertar en la UI.
        *   `presence.update`: Mostrar "Escribiendo..." en la UI del chat.

2.  **React Query**:
    *   Usar `staleTime: Infinity` para chats históricos.
    *   Invalidar queries manualmente al recibir eventos WS.

---

## 📝 Próximos Pasos para Desarrollo

1.  **Fase 1: Core & Conexión**
    *   Setup Next.js dashboard layout.
    *   CRUD Instancias.
    *   Vista de QR y Conexión.
2.  **Fase 2: Mensajería Básica**
    *   Implementar `ChatInterface` básico (solo texto).
    *   Listado de Chats.
3.  **Fase 3: Mensajería Avanzada y Contactos**
    *   Soporte Multimedia.
    *   Gestión de Contactos.
4.  **Fase 4: Automatización**
    *   Campañas y Autorespuestas.
