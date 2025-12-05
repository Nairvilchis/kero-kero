package models

import "time"

// CRMContact representa un contacto con información extendida de CRM
type CRMContact struct {
	JID           string    `json:"jid" db:"jid"`
	InstanceID    string    `json:"instance_id" db:"instance_id"`
	Name          string    `json:"name" db:"name"`
	Phone         string    `json:"phone" db:"phone"`
	Email         string    `json:"email,omitempty" db:"email"`
	Notes         string    `json:"notes,omitempty" db:"notes"`
	Tags          []string  `json:"tags,omitempty" db:"tags"` // Almacenado como JSON o array en DB
	Status        string    `json:"status" db:"status"`       // lead, customer, blocked, etc.
	LastContactAt time.Time `json:"last_contact_at" db:"last_contact_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// CreateContactRequest solicitud para crear/actualizar un contacto en CRM
type CreateCRMContactRequest struct {
	Phone  string   `json:"phone" validate:"required"`
	Name   string   `json:"name"`
	Email  string   `json:"email"`
	Notes  string   `json:"notes"`
	Tags   []string `json:"tags"`
	Status string   `json:"status"` // lead, customer, etc.
}

// UpdateCRMContactRequest solicitud para actualizar un contacto
type UpdateCRMContactRequest struct {
	Name   string   `json:"name"`
	Email  string   `json:"email"`
	Notes  string   `json:"notes"`
	Tags   []string `json:"tags"`
	Status string   `json:"status"`
}
