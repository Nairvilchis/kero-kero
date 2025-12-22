package routes

import (
	"kero-kero/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupContactRoutes(r chi.Router, handler *handlers.ContactHandler) {
	r.Route("/instances/{instanceID}/contacts", func(r chi.Router) {
		// Verificación y listado
		r.Post("/check", handler.CheckContacts)
		r.Post("/check-numbers", handler.CheckNumbers)
		r.Get("/", handler.GetContacts)
		r.Get("/blocklist", handler.GetBlocklist)

		// Información específica
		r.Get("/{phone}", handler.GetContactInfo)                    // Info detallada
		r.Get("/{phone}/about", handler.GetAbout)                    // Estado/About
		r.Get("/profile-picture", handler.GetProfilePicture)         // Query param ?phone=... (Legacy)
		r.Get("/{phone}/profile-picture", handler.GetProfilePicture) // Nuevo estilo REST

		// Acciones
		r.Post("/presence/subscribe", handler.SubscribePresence)
		r.Post("/block", handler.Block)     // Body {phone: ...}
		r.Post("/unblock", handler.Unblock) // Body {phone: ...}

		// Acciones RESTful
		r.Post("/{phone}/block", handler.Block)
		r.Post("/{phone}/unblock", handler.Unblock)
	})
}
