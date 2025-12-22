package routes

import (
	"kero-kero/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupBusinessRoutes(r chi.Router, handler *handlers.BusinessHandler) {
	r.Route("/instances/{instanceID}/business", func(r chi.Router) {
		r.Get("/profile", handler.GetProfile)

		r.Route("/labels", func(r chi.Router) {
			r.Post("/", handler.CreateLabel)
			r.Post("/assign", handler.AssignLabel)
		})

		r.Route("/autolabel", func(r chi.Router) {
			r.Get("/rules", handler.GetAutoLabelRules)
			r.Post("/rules", handler.SetAutoLabelRules)
		})
	})
}
