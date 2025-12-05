package routes

import (
	"github.com/go-chi/chi/v5"

	"kero-kero/internal/handlers"
)

// SetupInstanceRoutes configura las rutas de instancias
func SetupInstanceRoutes(r chi.Router, handler *handlers.InstanceHandler) {
	r.Route("/instances", func(r chi.Router) {
		r.Post("/", handler.Create)
		r.Get("/", handler.List)

		r.Route("/{instanceID}", func(r chi.Router) {
			r.Get("/", handler.Get)
			r.Put("/", handler.Update)
			r.Delete("/", handler.Delete)

			// Acciones
			r.Post("/connect", handler.Connect)
			r.Post("/disconnect", handler.Disconnect)

			// Estado y QR
			r.Get("/qr", handler.GetQR)
			r.Get("/status", handler.GetStatus)
		})
	})
}
