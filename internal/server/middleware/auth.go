package middleware

import (
	"net/http"
	"strings"

	"kero-kero/internal/services"
	"kero-kero/pkg/errors"
)

// Auth middleware para autenticación con API Key o JWT
func Auth(apiKey string, authService *services.AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Permitir health check, endpoints de autenticación y system info sin validación
			if r.URL.Path == "/health" ||
				r.URL.Path == "/" ||
				r.URL.Path == "/system/info" ||
				strings.HasPrefix(r.URL.Path, "/auth/login") {
				next.ServeHTTP(w, r)
				return
			}

			// Primero intentar validar con JWT (si existe Authorization header)
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				// Extraer el token
				token := strings.TrimPrefix(authHeader, "Bearer ")

				// Validar token JWT
				_, err := authService.ValidateToken(token)
				if err == nil {
					// Token JWT válido
					next.ServeHTTP(w, r)
					return
				}
				// Si el token JWT falló, continuamos para intentar con API Key
			}

			// Validación con API Key (legacy/alternativa)
			providedKey := r.Header.Get("X-API-Key")
			if providedKey == "" && authHeader != "" {
				// Intentar usar el valor del Authorization header como API key
				providedKey = strings.TrimPrefix(authHeader, "Bearer ")
			}

			// Validar API key
			if providedKey != apiKey {
				errors.WriteJSON(w, errors.ErrUnauthorized.WithDetails("Autenticación inválida o faltante"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// APIKeyAuth es un alias para mantener compatibilidad con código existente
// DEPRECATED: Usar Auth en su lugar
func APIKeyAuth(apiKey string) func(next http.Handler) http.Handler {
	// Para usar esta función sin el authService, creamos un middleware simplificado
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Permitir health check, endpoints de autenticación y system info sin validación
			if r.URL.Path == "/health" ||
				r.URL.Path == "/" ||
				r.URL.Path == "/system/info" ||
				strings.HasPrefix(r.URL.Path, "/auth/login") {
				next.ServeHTTP(w, r)
				return
			}

			// Obtener API key del header
			providedKey := r.Header.Get("X-API-Key")
			if providedKey == "" {
				// Intentar obtener del header Authorization
				auth := r.Header.Get("Authorization")
				if strings.HasPrefix(auth, "Bearer ") {
					providedKey = strings.TrimPrefix(auth, "Bearer ")
				}
			}

			// Validar API key
			if providedKey != apiKey {
				errors.WriteJSON(w, errors.ErrUnauthorized.WithDetails("API key inválida o faltante"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
