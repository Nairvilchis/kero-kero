package models

// BulkMessageRequest solicitud para envío masivo
type BulkMessageRequest struct {
	Phones   []string `json:"phones" validate:"required,min=1"`
	Message  string   `json:"message" validate:"required"`
	MediaURL string   `json:"media_url,omitempty"`
	MinDelay int      `json:"min_delay,omitempty"` // Milisegundos
	MaxDelay int      `json:"max_delay,omitempty"` // Milisegundos
}

// BulkMessageResponse respuesta de envío masivo
type BulkMessageResponse struct {
	Success         bool   `json:"success"`
	JobID           string `json:"job_id"`
	TotalRecipients int    `json:"total_recipients"`
	Status          string `json:"status"`
}

// ScheduleMessageRequest solicitud para programar mensaje
type ScheduleMessageRequest struct {
	Phone     string `json:"phone" validate:"required"`
	Message   string `json:"message" validate:"required"`
	ExecuteAt int64  `json:"execute_at" validate:"required"` // Unix Timestamp
}

// AutoReplyConfig configuración de respuesta automática
type AutoReplyConfig struct {
	Enabled         bool     `json:"enabled"`
	Message         string   `json:"message"`
	TriggerKeywords []string `json:"trigger_keywords,omitempty"`
	MatchType       string   `json:"match_type,omitempty"` // contains, exact, startswith
}

// ScheduledMessage estructura interna para guardar en Redis
type ScheduledMessage struct {
	ID        string `json:"id"`
	Phone     string `json:"phone"`
	Message   string `json:"message"`
	ExecuteAt int64  `json:"execute_at"`
}
