package models

// CreatePollRequest representa la solicitud para crear una encuesta
type CreatePollRequest struct {
	Phone           string   `json:"phone" validate:"required"`
	Question        string   `json:"question" validate:"required"`
	Options         []string `json:"options" validate:"required,min=2,max=12"` // WhatsApp permite 2-12 opciones
	SelectableCount uint32   `json:"selectable_count,omitempty"`               // 0 = selección única, >0 = múltiple
}

// VotePollRequest representa la solicitud para votar en una encuesta
type VotePollRequest struct {
	Phone       string   `json:"phone" validate:"required"`
	MessageID   string   `json:"message_id" validate:"required"`
	OptionNames []string `json:"option_names" validate:"required"` // Nombres de las opciones a votar
}

// PollResponse representa la respuesta al crear una encuesta
type PollResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}
