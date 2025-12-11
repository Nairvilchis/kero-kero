package models

import "encoding/json"

// MessageJobType define el tipo de mensaje a ser enviado.
type MessageJobType string

const (
	MessageJobTypeText     MessageJobType = "text"
	MessageJobTypeImage    MessageJobType = "image"
	MessageJobTypeAudio    MessageJobType = "audio"
	MessageJobTypeVideo    MessageJobType = "video"
	MessageJobTypeDocument MessageJobType = "document"
	// ... otros tipos de mensaje que se quieran encolar
)

// MessageJob representa un trabajo de envío de mensaje que se encola en Redis.
type MessageJob struct {
	InstanceID    string          `json:"instance_id"`
	CorrelationID string          `json:"correlation_id"`
	Type          MessageJobType  `json:"type"`
	Payload       json.RawMessage `json:"payload"` // Contiene el JSON del request original (ej. SendTextRequest)
	RetryCount    int             `json:"retry_count"`
}
