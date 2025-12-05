package models

import "time"

// Instance representa una instancia de WhatsApp
type Instance struct {
	InstanceID      string     `json:"instance_id"`
	JID             string     `json:"jid,omitempty"`
	Name            string     `json:"name,omitempty"`
	WebhookURL      string     `json:"webhook_url,omitempty"`
	Status          string     `json:"status"`
	SyncHistory     bool       `json:"sync_history"` // Nuevo campo
	LastConnectedAt *time.Time `json:"last_connected_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// InstanceStatus representa los posibles estados de una instancia
type InstanceStatus string

const (
	StatusDisconnected  InstanceStatus = "disconnected"
	StatusConnecting    InstanceStatus = "connecting"
	StatusConnected     InstanceStatus = "connected"
	StatusAuthenticated InstanceStatus = "authenticated"
	StatusFailed        InstanceStatus = "failed"
)

// CreateInstanceRequest representa la solicitud para crear una instancia
type CreateInstanceRequest struct {
	InstanceID  string `json:"instance_id" validate:"required"`
	WebhookURL  string `json:"webhook_url,omitempty"`
	SyncHistory bool   `json:"sync_history"` // Nuevo campo
}

// UpdateInstanceRequest representa la solicitud para actualizar una instancia
type UpdateInstanceRequest struct {
	Name        *string `json:"name,omitempty"`
	WebhookURL  *string `json:"webhook_url,omitempty"`
	SyncHistory *bool   `json:"sync_history,omitempty"` // Nuevo campo
}

// InstanceResponse representa la respuesta con información de instancia
type InstanceResponse struct {
	Success bool      `json:"success"`
	Data    *Instance `json:"data,omitempty"`
	Message string    `json:"message,omitempty"`
}

// InstanceListResponse representa la respuesta con lista de instancias
type InstanceListResponse struct {
	Success bool        `json:"success"`
	Data    []*Instance `json:"data"`
	Total   int         `json:"total"`
}
