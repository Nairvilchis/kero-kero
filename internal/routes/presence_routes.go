package routes

import (
	"github.com/go-chi/chi/v5"

	"kero-kero/internal/handlers"
)

// RegisterPresenceRoutes registra las rutas de presencia
func RegisterPresenceRoutes(r chi.Router, handler *handlers.PresenceHandler) {
	r.Route("/instances/{instanceID}/presence", func(r chi.Router) {
		// POST /instances/{instanceID}/presence/start - Activar presencia (typing/recording)
		r.Post("/start", handler.Start)

		// POST /instances/{instanceID}/presence/stop - Detener presencia
		r.Post("/stop", handler.Stop)

		// POST /instances/{instanceID}/presence/timed - Presencia temporal con auto-stop
		r.Post("/timed", handler.Timed)

		// POST /instances/{instanceID}/presence/status - Cambiar estado online/offline
		r.Post("/status", handler.SetStatus)
	})
}
