package routes

import (
	"kero-kero/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupAutomationRoutes(r chi.Router, handler *handlers.AutomationHandler) {
	r.Route("/instances/{instanceID}/automation", func(r chi.Router) {
		r.Post("/bulk-message", handler.SendBulkMessage)
		r.Post("/schedule-message", handler.ScheduleMessage)
		r.Post("/auto-reply", handler.SetAutoReply)
		r.Get("/auto-reply", handler.GetAutoReply)
	})
}
