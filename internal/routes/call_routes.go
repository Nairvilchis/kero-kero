package routes

import (
	"kero-kero/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupCallRoutes(r chi.Router, handler *handlers.CallHandler) {
	r.Route("/instances/{instanceID}/calls", func(r chi.Router) {
		r.Post("/reject", handler.RejectCall)
		r.Get("/settings", handler.GetSettings)
		r.Put("/settings", handler.UpdateSettings)
	})
}
