# Sistema de Autenticación JWT

## Descripción General

Kero-Kero ahora implementa un sistema de autenticación basado en **JSON Web Tokens (JWT)** para el dashboard. Este sistema proporciona una capa adicional de seguridad y permite un mejor control de sesiones.

## Flujo de Autenticación

### 1. Login del Usuario

Cuando un usuario accede al dashboard por primera vez:

1. **Página de Login**: El usuario ingresa:
   - URL del servidor (ejemplo: `http://localhost:8080`)
   - API Key (la clave configurada en `.env.local` del servidor)

2. **Validación**: El dashboard envía una petición `POST` al endpoint `/auth/login`:
   ```json
   {
     "api_key": "dev-api-key-12345"
   }
   ```

3. **Generación de Token**: Si la API Key es válida, el servidor:
   - Valida que la API Key coincida con la configurada en `JWT_SECRET`
   - Genera un token JWT firmado con `JWT_SECRET`
   - El token incluye:
     - `sub` (subject): Identificador (actualmente "dashboard")
     - `iat` (issued at): Timestamp de emisión
     - `exp` (expires at): Timestamp de expiración (24 horas)
     - `type`: Tipo de token ("dashboard")

4. **Respuesta**: El servidor responde con:
   ```json
   {
     "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
     "expires_at": 1733444123,
     "type": "Bearer"
   }
   ```

5. **Almacenamiento**: El dashboard guarda en `localStorage`:
   - `kero_jwt_token`: El token JWT
   - `kero_jwt_expires`: Timestamp de expiración
   - `kero_api_url`: URL del servidor
   - `kero_api_key`: API Key (backup)

### 2. Peticiones Autenticadas

Para todas las peticiones subsecuentes a la API:

1. El cliente API (Axios) automáticamente intercepta cada petición
2. Obtiene el token JWT de `localStorage`
3. Lo incluye en el header `Authorization`:
   ```
   Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ```

### 3. Validación en el Servidor

El middleware de autenticación del servidor:

1. **Extrae el token** del header `Authorization`
2. **Valida la firma** usando `JWT_SECRET`
3. **Verifica la expiración** del token
4. Si todo es válido, permite el acceso al endpoint
5. Si falla, responde con `401 Unauthorized`

### 4. Manejo de Tokens Expirados

**En el Cliente (Dashboard)**:
- `AuthGuard` verifica la expiración del token antes de cada navegación
- Si el token ha expirado, limpia el `localStorage` y redirige al login
- El interceptor de Axios detecta respuestas `401` y redirige al login

**En el Servidor**:
- El servicio de autenticación valida el timestamp `exp` del token
- Rechaza tokens expirados con error específico

## Configuración

### Variables de Entorno del Servidor

En `server-kero-kero/.env.local`:

```bash
# API Key para login inicial
API_KEY=dev-api-key-12345

# Secreto para firmar tokens JWT
# ⚠️ IMPORTANTE: Debe ser una cadena aleatoria y segura en producción
# Generar con: openssl rand -base64 32
JWT_SECRET=dev-jwt-secret-67890
```

### Variables de Entorno del Dashboard

En `kero-kero-dashboard/.env.local`:

```bash
# URL del servidor
NEXT_PUBLIC_API_URL=http://localhost:8080

# API Key (solo para login, no para peticiones)
NEXT_PUBLIC_API_KEY=dev-api-key-12345
```

## Implementación Técnica

### Backend (Go)

**Archivos Clave**:

1. **`internal/services/auth_service.go`**
   - Servicio principal de autenticación
   - Genera y valida tokens JWT
   - Usa HMAC-SHA256 para firmar tokens
   - No requiere dependencias externas

2. **`internal/handlers/auth_handler.go`**
   - Endpoints HTTP de autenticación
   - `POST /auth/login`: Autenticación y generación de token
   - `GET /auth/validate`: Validación de token (debugging)

3. **`internal/routes/auth_routes.go`**
   - Registra las rutas de autenticación

4. **`internal/server/middleware/auth.go`**
   - Middleware `Auth()`: Valida JWT o API Key
   - Soporta ambos métodos de autenticación para compatibilidad
   - Permite endpoints públicos como `/health` y `/auth/login`

### Frontend (Next.js/React)

**Archivos Clave**:

1. **`app/login/page.tsx`**
   - Página de login
   - Envía API Key al endpoint `/auth/login`
   - Guarda token JWT en `localStorage`

2. **`components/AuthGuard.tsx`**
   - Componente que protege rutas
   - Verifica existencia y validez del token
   - Redirige al login si el token está ausente o expirado

3. **`lib/api.ts`**
   - Cliente Axios con interceptores
   - Agrega automáticamente el token a cada petición
   - Maneja errores 401 y redirige al login

## Seguridad

### Mejores Prácticas Implementadas

1. **Firma HMAC-SHA256**: Los tokens están firmados criptográficamente
2. **Expiración de 24 horas**: Los tokens tienen vida limitada
3. **Validación estricta**: Se verifica firma, estructura y expiración
4. **Sin almacenamiento en servidor**: Tokens stateless (sin base de datos)
5. **Fallback a API Key**: Mantiene compatibilidad con métodos antiguos

### Consideraciones de Producción

⚠️ **IMPORTANTE**:

1. **JWT_SECRET debe ser aleatorio y seguro**:
   ```bash
   openssl rand -base64 32
   ```

2. **Usar HTTPS** en producción para proteger tokens en tránsito

3. **CORS configurado correctamente**: Limitar orígenes permitidos

4. **Rate limiting**: Ya implementado para prevenir ataques de fuerza bruta

## Endpoints de Autenticación

### POST /auth/login

Autentica con API Key y obtiene un token JWT.

**Request**:
```json
{
  "api_key": "dev-api-key-12345"
}
```

**Response (200 OK)**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJkYXNoYm9hcmQiLCJpYXQiOjE3MzMzNTc3MjMsImV4cCI6MTczMzQ0NDEyMywidHlwZSI6ImRhc2hib2FyZCJ9.signature",
  "expires_at": 1733444123,
  "type": "Bearer"
}
```

**Response (401 Unauthorized)**:
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "No autorizado",
    "details": "Credenciales inválidas"
  }
}
```

### GET /auth/validate

Valida un token JWT (útil para debugging).

**Request Headers**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response (200 OK)**:
```json
{
  "valid": true,
  "claims": {
    "sub": "dashboard",
    "iat": 1733357723,
    "exp": 1733444123,
    "type": "dashboard"
  }
}
```

## Retrocompatibilidad

El sistema mantiene compatibilidad con el método anterior (API Key directa):

- **Middleware `Auth()`**: Intenta validar JWT primero, luego API Key
- **Cliente API**: Usa JWT si está disponible, sino API Key
- **Migración gradual**: Usuarios existentes pueden seguir usando API Key

## Testing

### Probar el Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"api_key":"dev-api-key-12345"}'
```

### Probar Token JWT

```bash
# 1. Obtener token
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"api_key":"dev-api-key-12345"}' | jq -r '.token')

# 2. Usar token
curl http://localhost:8080/instances \
  -H "Authorization: Bearer $TOKEN"
```

### Validar Token

```bash
curl http://localhost:8080/auth/validate \
  -H "Authorization: Bearer $TOKEN"
```

## Troubleshooting

### Error: "Token expirado"

**Causa**: El token JWT ha superado su tiempo de vida (24 horas).

**Solución**: Volver a iniciar sesión en el dashboard.

### Error: "Firma de token inválida"

**Causa**: El `JWT_SECRET` en el servidor ha cambiado.

**Solución**: 
1. Verificar que `JWT_SECRET` esté configurado correctamente
2. Limpiar tokens antiguos y volver a iniciar sesión

### Error: "Autenticación inválida o faltante"

**Causa**: No se está enviando el token o API Key.

**Solución**:
1. Verificar que el token esté en `localStorage`
2. Verificar que el interceptor de Axios esté funcionando
3. Ver la consola del navegador para errores

## Futuras Mejoras

1. **Refresh Tokens**: Implementar tokens de actualización de larga duración
2. **Roles y Permisos**: Agregar claims de rol al JWT
3. **Múltiples Usuarios**: Sistema de usuarios con credenciales individuales
4. **Revocación de Tokens**: Blacklist de tokens en Redis
5. **2FA**: Autenticación de dos factores
