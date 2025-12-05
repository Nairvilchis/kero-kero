package models

// CreateGroupRequest solicitud para crear un grupo
type CreateGroupRequest struct {
	Name         string   `json:"name" validate:"required"`
	Participants []string `json:"participants" validate:"required"` // Lista de números de teléfono
}

// GroupParticipantRequest solicitud para modificar participantes
type GroupParticipantRequest struct {
	Participants []string `json:"participants" validate:"required"`
	Action       string   `json:"action" validate:"required,oneof=add remove promote demote"`
}

// GroupResponse respuesta con información del grupo
type GroupResponse struct {
	JID              string             `json:"jid"`
	Name             string             `json:"name"`
	OwnerJID         string             `json:"owner_jid,omitempty"`
	Topic            string             `json:"topic,omitempty"`
	CreationTime     int64              `json:"creation_time,omitempty"`
	ParticipantCount int                `json:"participant_count"`
	Participants     []GroupParticipant `json:"participants,omitempty"`
}

// GroupParticipant información de un participante
type GroupParticipant struct {
	JID          string `json:"jid"`
	IsAdmin      bool   `json:"is_admin,omitempty"`
	IsSuperAdmin bool   `json:"is_super_admin,omitempty"`
}

// CreateGroupResponse respuesta al crear un grupo
type CreateGroupResponse struct {
	Success            bool     `json:"success"`
	GroupJID           string   `json:"group_jid"`
	Name               string   `json:"name"`
	ParticipantsAdded  int      `json:"participants_added"`
	FailedParticipants []string `json:"failed_participants,omitempty"`
}

// AddParticipantsRequest solicitud para agregar participantes
type AddParticipantsRequest struct {
	Participants []string `json:"participants" validate:"required,min=1"`
}

// RemoveParticipantsRequest solicitud para remover participantes
type RemoveParticipantsRequest struct {
	Participants []string `json:"participants" validate:"required,min=1"`
}

// PromoteAdminRequest solicitud para promover a admin
type PromoteAdminRequest struct {
	Participants []string `json:"participants" validate:"required,min=1"`
}

// DemoteAdminRequest solicitud para degradar admin
type DemoteAdminRequest struct {
	Participants []string `json:"participants" validate:"required,min=1"`
}

// UpdateGroupRequest solicitud para actualizar información del grupo
type UpdateGroupRequest struct {
	Name        *string `json:"name,omitempty"`        // Nombre del grupo
	Description *string `json:"description,omitempty"` // Descripción/topic del grupo
}

// UpdateGroupSettingsRequest solicitud para actualizar configuración del grupo
type UpdateGroupSettingsRequest struct {
	IsLocked          *bool  `json:"is_locked,omitempty"`          // Solo admins pueden editar info
	IsAnnounce        *bool  `json:"is_announce,omitempty"`        // Solo admins pueden enviar mensajes
	IsEphemeral       *bool  `json:"is_ephemeral,omitempty"`       // Mensajes temporales activados
	DisappearingTimer *int64 `json:"disappearing_timer,omitempty"` // Tiempo en segundos (0 = desactivado)
}

// UpdateGroupPictureRequest solicitud para actualizar foto del grupo
type UpdateGroupPictureRequest struct {
	ImageURL *string `json:"image_url,omitempty"` // URL de la imagen
	Image    *string `json:"image,omitempty"`     // Imagen en base64
}

// GroupInfoResponse respuesta detallada con información del grupo
type GroupInfoResponse struct {
	Success bool       `json:"success"`
	Data    *GroupInfo `json:"data,omitempty"`
}

// GroupInfo información completa de un grupo
type GroupInfo struct {
	JID               string             `json:"jid"`
	Name              string             `json:"name"`
	Topic             string             `json:"topic,omitempty"`   // Descripción
	Owner             string             `json:"owner,omitempty"`   // JID del creador
	Created           int64              `json:"created,omitempty"` // Unix timestamp
	Participants      []GroupParticipant `json:"participants,omitempty"`
	ParticipantCount  int                `json:"participant_count"`
	IsLocked          bool               `json:"is_locked"`          // Solo admins editan info
	IsAnnounce        bool               `json:"is_announce"`        // Solo admins envían mensajes
	IsEphemeral       bool               `json:"is_ephemeral"`       // Mensajes temporales
	DisappearingTimer int64              `json:"disappearing_timer"` // Segundos
}

// InviteLinkResponse respuesta con link de invitación
type InviteLinkResponse struct {
	Success    bool   `json:"success"`
	InviteLink string `json:"invite_link,omitempty"`
	InviteCode string `json:"invite_code,omitempty"` // Solo el código
}

// GroupActionResponse respuesta genérica para acciones de grupo
type GroupActionResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message,omitempty"`
	Failed  []string `json:"failed,omitempty"` // Participantes que fallaron
}
