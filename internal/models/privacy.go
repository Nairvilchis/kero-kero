package models

// PrivacySettings configuración de privacidad
type PrivacySettings struct {
	LastSeen     string `json:"last_seen" validate:"omitempty,oneof=all contacts contact_blacklist none"`
	ProfilePhoto string `json:"profile_photo" validate:"omitempty,oneof=all contacts contact_blacklist none"`
	Status       string `json:"status" validate:"omitempty,oneof=all contacts contact_blacklist none"`
	ReadReceipts *bool  `json:"read_receipts,omitempty"`
	Groups       string `json:"groups" validate:"omitempty,oneof=all contacts contact_blacklist none"`
	DefaultTimer string `json:"default_timer,omitempty"` // Duración de mensajes temporales
}

// PrivacyUpdateRequest solicitud para actualizar privacidad
type PrivacyUpdateRequest struct {
	Category string `json:"category" validate:"required,oneof=last_seen profile_photo status read_receipts groups default_timer"`
	Value    string `json:"value" validate:"required"` // Para read_receipts usar "true"/"false"
}
