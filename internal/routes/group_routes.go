package routes

import (
	"kero-kero/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupGroupRoutes(r chi.Router, handler *handlers.GroupHandler) {
	r.Route("/instances/{instanceID}/groups", func(r chi.Router) {
		// Gestión básica de grupos
		r.Post("/", handler.CreateGroup)               // Crear grupo
		r.Get("/", handler.ListGroups)                 // Listar grupos
		r.Get("/{groupID}", handler.GetGroupInfo)      // Info del grupo
		r.Put("/{groupID}", handler.UpdateGroupInfo)   // Actualizar nombre/descripción
		r.Post("/{groupID}/leave", handler.LeaveGroup) // Salir del grupo

		// Gestión de participantes
		r.Post("/{groupID}/participants", handler.AddParticipants)      // Agregar participantes
		r.Delete("/{groupID}/participants", handler.RemoveParticipants) // Remover participantes
		r.Patch("/{groupID}/participants", handler.UpdateParticipants)  // Actualizar (método genérico legacy)

		// Gestión de administradores
		r.Post("/{groupID}/admins", handler.PromoteToAdmin) // Promover a admin
		r.Delete("/{groupID}/admins", handler.DemoteAdmin)  // Degradar admin

		// Configuración del grupo
		r.Put("/{groupID}/settings", handler.UpdateGroupSettings) // Actualizar configuración

		// Links de invitación
		r.Get("/{groupID}/invite", handler.GetInviteLink)            // Obtener link
		r.Post("/{groupID}/invite/revoke", handler.RevokeInviteLink) // Revocar link
	})
}
