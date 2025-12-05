package routes

import (
	"kero-kero/internal/handlers"

	"github.com/go-chi/chi/v5"
)

// RegisterAuthRoutes registra las rutas de autenticación
func RegisterAuthRoutes(r chi.Router, handler *handlers.AuthHandler) {
	r.Route("/auth", func(r chi.Router) {
		// POST /auth/login - Iniciar sesión con API key y obtener JWT
		r.Post("/login", handler.Login)

		// GET /auth/validate - Validar token JWT (protegido con el middleware de auth)
		r.Get("/validate", handler.ValidateToken)
	})
}
