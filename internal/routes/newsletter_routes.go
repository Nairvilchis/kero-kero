package routes

import (
	"kero-kero/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupNewsletterRoutes(r chi.Router, handler *handlers.NewsletterHandler) {
	r.Route("/instances/{instanceID}/newsletters", func(r chi.Router) {
		r.Post("/", handler.Create)          // Crear canal
		r.Get("/", handler.ListSubscribed)   // Listar canales suscritos
		r.Post("/send", handler.SendMessage) // Enviar mensaje a un canal (si eres admin)

		r.Route("/{jid}", func(r chi.Router) {
			r.Get("/", handler.GetInfo)           // Info del canal
			r.Post("/follow", handler.Follow)     // Seguir canal
			r.Post("/unfollow", handler.Unfollow) // Dejar de seguir canal
		})
	})
}
