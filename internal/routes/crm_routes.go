package routes

import (
	"kero-kero/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupCRMRoutes(r chi.Router, handler *handlers.CRMHandler) {
	r.Route("/instances/{instanceID}/crm", func(r chi.Router) {
		r.Get("/contacts", handler.ListContacts)
		r.Get("/contacts/{jid}", handler.GetContact)
		r.Put("/contacts/{jid}", handler.UpdateContact)
	})
}
