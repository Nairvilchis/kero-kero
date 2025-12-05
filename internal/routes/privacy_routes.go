package routes

import (
	"kero-kero/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupPrivacyRoutes(r chi.Router, handler *handlers.PrivacyHandler) {
	r.Route("/instances/{instanceID}/privacy", func(r chi.Router) {
		r.Get("/", handler.GetPrivacySettings)
		r.Patch("/", handler.UpdatePrivacySettings)
		r.Put("/", handler.UpdatePrivacySettings) // Alias para PUT
	})
}
