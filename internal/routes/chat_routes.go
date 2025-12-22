package routes

import (
	"kero-kero/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupChatRoutes(r chi.Router, handler *handlers.ChatHandler) {
	r.Route("/instances/{instanceID}/chats", func(r chi.Router) {
		r.Get("/", handler.List)
		r.Get("/{jid}/messages", handler.GetHistory)
		r.Delete("/{jid}", handler.Delete)
		r.Post("/archive", handler.Archive)
		r.Post("/status", handler.UpdateStatus)
		r.Post("/{jid}/read", handler.MarkAsRead)
		r.Post("/mute", handler.Mute)
		r.Post("/pin", handler.Pin)
	})
}
