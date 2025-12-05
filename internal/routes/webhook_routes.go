package routes

import (
	"kero-kero/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupWebhookRoutes(r chi.Router, handler *handlers.WebhookHandler) {
	r.Route("/instances/{instanceID}/webhook", func(r chi.Router) {
		r.Post("/", handler.SetWebhook)
		r.Get("/", handler.GetWebhook)
		r.Delete("/", handler.DeleteWebhook)
	})
}
