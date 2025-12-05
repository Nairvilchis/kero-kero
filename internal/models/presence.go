package models

// PresenceType tipos de presencia en chat
type PresenceType string

const (
	PresenceTyping    PresenceType = "typing"    // Escribiendo...
	PresenceRecording PresenceType = "recording" // Grabando audio...
)

// PresenceStatus estado de presencia general
type PresenceStatus string

const (
	PresenceAvailable   PresenceStatus = "available"   // En línea
	PresenceUnavailable PresenceStatus = "unavailable" // Desconectado
)

// StartPresenceRequest solicitud para activar presencia en un chat
type StartPresenceRequest struct {
	Phone string       `json:"phone" validate:"required"`
	Type  PresenceType `json:"type" validate:"required,oneof=typing recording"`
}

// StopPresenceRequest solicitud para detener presencia en un chat
type StopPresenceRequest struct {
	Phone string `json:"phone" validate:"required"`
}

// TimedPresenceRequest solicitud para presencia temporal
type TimedPresenceRequest struct {
	Phone    string       `json:"phone" validate:"required"`
	Type     PresenceType `json:"type" validate:"required,oneof=typing recording"`
	Duration int          `json:"duration" validate:"required,min=100,max=120000"` // 100ms a 2 minutos
}

// SetStatusRequest solicitud para cambiar estado en línea/desconectado
type SetStatusRequest struct {
	Status PresenceStatus `json:"status" validate:"required,oneof=available unavailable"`
}

// PresenceResponse respuesta de operaciones de presencia
type PresenceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
