package routes

import (
	"kero-kero/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupSyncRoutes(r chi.Router, handler *handlers.SyncHandler) {
	r.Route("/instances/{instanceID}/sync", func(r chi.Router) {
		r.Post("/", handler.Start)              // Iniciar sincronización
		r.Get("/progress", handler.GetProgress) // Obtener progreso
		r.Delete("/", handler.Cancel)           // Cancelar sincronización
	})
}
