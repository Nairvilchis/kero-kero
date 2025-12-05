package routes

import (
	"kero-kero/internal/handlers"

	"github.com/go-chi/chi/v5"
)

// SetupMessageRoutes configura las rutas de mensajes
func SetupMessageRoutes(r chi.Router, handler *handlers.MessageHandler) {
	r.Route("/instances/{instanceID}/messages", func(r chi.Router) {
		// Mensajes básicos
		r.Post("/text", handler.SendText)
		r.Post("/text-with-typing", handler.SendTextWithTyping) // Nuevo: con simulación de escritura
		r.Post("/image", handler.SendImage)
		r.Post("/video", handler.SendVideo)
		r.Post("/audio", handler.SendAudio)
		r.Post("/document", handler.SendDocument)
		r.Post("/location", handler.SendLocation)

		// Interacciones
		r.Post("/react", handler.React)
		r.Post("/revoke", handler.Revoke)
		r.Post("/mark-read", handler.MarkAsRead) // Nuevo: marcar como leído
		r.Post("/download", handler.DownloadMedia)

		// Encuestas
		r.Post("/poll", handler.CreatePoll)
		r.Post("/poll/vote", handler.VotePoll)
	})
}
