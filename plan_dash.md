# 📘 Guía Maestra de Desarrollo: Dashboard Kero-Kero (Api Server)

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

La API utiliza autenticación vía Header `X-Api-Key` o JWT bearer.

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
  // Configuración para usar API Key o JWT según corresponda
  if (process.env.NEXT_PUBLIC_API_KEY) {
     config.headers['X-Api-Key'] = process.env.NEXT_PUBLIC_API_KEY;
  }
  return config;
});

export default api;
```

---

## 🗺️ Mapa Completo de Endpoints API

A continuación, se listan **todas** las rutas disponibles extraídas del código fuente del servidor.

### 1. Autenticación (`/auth`)
| Método | Ruta | Descripción | Payload |
| :--- | :--- | :--- | :--- |
| `POST` | `/auth/login` | Login usuario | `{ "username": "...", "password": "..." }` (según implementación) |
| `GET` | `/auth/validate` | Validar token actual | - |

### 2. Gestión de Instancias (`/instances`)
Rutas base para manejo de sesiones.
| Método | Ruta | Descripción | Payload / Params |
| :--- | :--- | :--- | :--- |
| `POST` | `/` | Crear nueva instancia | `{ "instance_id": "nombre", "sync_history": bool }` |
| `GET` | `/` | Listar todas | - |
| `GET` | `/{id}` | Detalles instancia | - |
| `PUT` | `/{id}` | Actualizar config | `{ "webhook_url": "...", "events": [...] }` |
| `DELETE` | `/{id}` | Eliminar instancia | - |
| `POST` | `/{id}/connect` | Iniciar conexión | - |
| `POST` | `/{id}/disconnect` | Cerrar sesión | - |
| `GET` | `/{id}/qr` | Obtener QR (Base64) | - |
| `GET` | `/{id}/status` | Estado actual | Response: `{ "status": "..." }` |

### 3. Mensajería (`/instances/{id}/messages`)
Todos los endpoints son POST. Usados para enviar mensajes.

| Ruta Suffix | Descripción | Payload JSON |
| :--- | :--- | :--- |
| `/text` | Texto simple | `{ "to": "521...", "message": "Hola" }` |
| `/text-with-typing` | Texto con simulación | `{ "to": "...", "message": "...", "duration": 2 }` |
| `/image` | Enviar Imagen | `{ "to": "...", "url": "...", "caption": "..." }` |
| `/video` | Enviar Video | `{ "to": "...", "url": "...", "caption": "..." }` |
| `/audio` | Enviar Audio (PTT) | `{ "to": "...", "url": "..." }` |
| `/document` | Enviar Documento | `{ "to": "...", "url": "...", "filename": "doc.pdf" }` |
| `/location` | Enviar Ubicación | `{ "to": "...", "latitude": 0.0, "longitude": 0.0, "name": "...", "address": "..." }` |
| `/contact` | Enviar VCard | `{ "to": "...", "vcard": "BEGIN:VCARD..." }` |
| `/react` | Reaccionar | `{ "message_id": "...", "reaction": "👍" }` |
| `/revoke` | Eliminar mensaje | `{ "message_id": "..." }` |
| `/edit` | Editar mensaje | `{ "message_id": "...", "new_text": "..." }` |
| `/mark-read` | Marcar leído | `{ "chat_jid": "...", "message_id": "...", "sender_jid": "..." }` |
| `/download` | Descargar multimedia | `{ "message_id": "...", "type": "image" }` |
| `/poll` | Crear Encuesta | `{ "to": "...", "name": "...", "options": ["..."], "selectable_count": 1 }` |
| `/poll/vote` | Votar Encuesta | `{ "to": "...", "poll_id": "...", "option_ids": ["..."] }` |

### 4. Chats (`/instances/{id}/chats`)
Gestión del listado de conversaciones.

| Método | Ruta | Descripción | Params |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | Listar chats | `?page=1` |
| `GET` | `/{jid}/messages` | Historial de mensajes | `?limit=50&offset=0` |
| `DELETE` | `/{jid}` | Eliminar chat | - |
| `POST` | `/archive` | Archivar chat | `{ "jid": "...", "archived": true }` |
| `POST` | `/status` | Actualizar estado | - |
| `POST` | `/{jid}/read` | Marcar como leído | - |
| `POST` | `/mute` | Silenciar chat | `{ "jid": "...", "duration": 8h }` |
| `POST` | `/pin` | Fijar chat | `{ "jid": "...", "pinned": true }` |

### 5. Contactos (`/instances/{id}/contacts`)
| Método | Ruta | Descripción | Payload |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | Listar contactos | `?page=1&limit=50` |
| `GET` | `/blocklist` | Listar bloqueados | - |
| `POST` | `/check` | Verificar si tienen WA | `{ "phones": ["..."] }` |
| `POST` | `/check-numbers` | Igual a check | - |
| `GET` | `/{phone}` | Info detallada | - |
| `GET` | `/{phone}/about` | Estado (About) | - |
| `GET` | `/{phone}/profile-picture` | URL foto perfil | - |
| `POST` | `/presence/subscribe` | Suscribirse a presencia | `{ "phones": [...] }` |
| `POST` | `/block` | Bloquear usuario | `{ "phone": "..." }` |
| `POST` | `/unblock` | Desbloquear | `{ "phone": "..." }` |

### 6. Grupos (`/instances/{id}/groups`)
| Método | Ruta | Descripción | Payload |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | Listar grupos | - |
| `POST` | `/` | Crear grupo | `{ "subject": "...", "participants": [] }` |
| `POST` | `/join` | Unirse vía link | `{ "code": "..." }` |
| `GET` | `/{gid}` | Info grupo | - |
| `PUT` | `/{gid}` | Actualizar Info | `{ "subject": "..." }` |
| `PUT` | `/{gid}/picture` | Actualizar Foto | `{ "url": "..." }` |
| `POST` | `/{gid}/leave` | Salir del grupo | - |
| `GET` | `/{gid}/invite` | Obtener link | - |
| `POST` | `/{gid}/invite/revoke` | Revocar link | - |
| `POST` | `/{gid}/participants` | Añadir participantes | `{ "participants": ["..."] }` |
| `DELETE` | `/{gid}/participants` | Remover participantes | `{ "participants": ["..."] }` |
| `PATCH` | `/{gid}/participants` | Actualizar participantes | `{ "action": "add/remove", ... }` |
| `POST` | `/{gid}/admins` | Promover a admin | `{ "participants": ["..."] }` |
| `DELETE` | `/{gid}/admins` | Degradar admin | `{ "participants": ["..."] }` |
| `PUT` | `/{gid}/settings` | Configuración | `{ "announce": bool, "locked": bool }` |

### 7. Automatización (`/instances/{id}/automation`)
| Método | Ruta | Descripción | Payload |
| :--- | :--- | :--- | :--- |
| `POST` | `/bulk-message` | Envío masivo | `{ "numbers": [...], "message": ... }` |
| `POST` | `/schedule-message` | Programar mensaje | `{ "to": "...", "message": "...", "send_at": "timestamp" }` |
| `POST` | `/auto-reply` | Configurar autorespuesta | `{ "match": "...", "response": "..." }` |
| `GET` | `/auto-reply` | Ver config actual | - |

### 8. Gestión de Negocio (`/instances/{id}/business`)
| Método | Ruta | Descripción | Payload |
| :--- | :--- | :--- | :--- |
| `GET` | `/profile` | Perfil de negocio | - |
| `POST` | `/labels` | Crear etiqueta | `{ "name": "...", "color": "..." }` |
| `POST` | `/labels/assign` | Asignar etiqueta | `{ "chat_jid": "...", "label_id": "..." }` |
| `GET` | `/autolabel/rules` | Ver reglas autolabel | - |
| `POST` | `/autolabel/rules` | Crear reglas | `{ "keywords": [...], "label_id": "..." }` |

### 9. Canales / Newsletters (`/instances/{id}/newsletters`)
| Método | Ruta | Descripción | Payload |
| :--- | :--- | :--- | :--- |
| `POST` | `/` | Crear Canal | `{ "name": "...", "description": "..." }` |
| `GET` | `/` | Listar Suscritos | - |
| `POST` | `/send` | Enviar a Canal | `{ "jid": "...", "content": "..." }` |
| `GET` | `/{jid}` | Info Canal | - |
| `POST` | `/{jid}/follow` | Seguir | - |
| `POST` | `/{jid}/unfollow` | Dejar de seguir | - |

### 10. Presencia (`/instances/{id}/presence`)
| Método | Ruta | Descripción | Payload |
| :--- | :--- | :--- | :--- |
| `POST` | `/start` | Iniciar (escribiendo/grabando) | `{ "to": "...", "state": "composing"|"recording"|"paused" }` |
| `POST` | `/stop` | Detener | `{ "to": "..." }` |
| `POST` | `/timed` | Presencia temporal | `{ "to": "...", "duration": 5 }` |
| `POST` | `/status` | Configurar estado online | `{ "status": "available"|"unavailable" }` |

### 11. Estados de WhatsApp (`/instances/{id}/status`)
Rutas para publicar "Historias/Estados". No confundir con el status de conexión.
| Método | Ruta | Descripción | Payload |
| :--- | :--- | :--- | :--- |
| `POST` | `/` | Publicar Estado | `{ "message": "...", "background_color": "#..." }` |
| `GET` | `/privacy` | Ver privacidad estados | - |

### 12. Sincronización Histórica (`/instances/{id}/sync`)
| Método | Ruta | Descripción | Payload |
| :--- | :--- | :--- | :--- |
| `POST` | `/` | Forzar Sync Historial | `{ "full": true|false }` |
| `GET` | `/progress` | Ver progreso % | - |
| `DELETE` | `/` | Cancelar Sync | - |

### 13. Llamadas (`/instances/{id}/calls`)
| Método | Ruta | Descripción | Payload |
| :--- | :--- | :--- | :--- |
| `POST` | `/reject` | Rechazar llamada entrante | `{ "call_id": "...", "call_from": "..." }` |
| `GET` | `/settings` | Configuración llamadas | - |
| `PUT` | `/settings` | Actualizar config | `{ "reject_all": true, "reject_message": "..." }` |

### 14. Privacidad (`/instances/{id}/privacy`)
| Método | Ruta | Descripción | Payload |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | Ver Config de Privacidad | (LastSeen, Profile, Status, ReadReceipts, Groups) |
| `PUT` | `/` | Actualizar Privacidad | `{ "last_seen": "all"|"contacts"|"none", ... }` |

### 15. Webhooks (`/instances/{id}/webhook`)
Configuración específica por instancia.
| Método | Ruta | Descripción | Payload |
| :--- | :--- | :--- | :--- |
| `POST` | `/` | Configurar Webhook | `{ "url": "...", "events": ["message", "status"] }` |
| `GET` | `/` | Ver Webhook actual | - |
| `DELETE` | `/` | Eliminar Webhook | - |

---

## 🏗️ Estructura del Frontend (Sugerida)

```
app/
├── (dashboard)/
│   ├── [instanceId]/
│   │   ├── chat/           # Usa /chats y /messages
│   │   ├── contacts/       # Usa /contacts
│   │   ├── groups/         # Usa /groups
│   │   ├── status/         # Usa /status (Stories)
│   │   ├── channels/       # Usa /newsletters
│   │   ├── automation/     # Usa /automation
│   │   └── settings/       # Usa /privacy, /calls, /business, /webhook
```
