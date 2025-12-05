package routes

import (
	"kero-kero/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupStatusRoutes(r chi.Router, handler *handlers.StatusHandler) {
	r.Route("/instances/{instanceID}/status", func(r chi.Router) {
		r.Post("/", handler.PublishStatus)
		r.Get("/privacy", handler.GetStatusPrivacy)
	})
}
