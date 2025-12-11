package models

import "time"

// WebhookConfig configuración de webhook para una instancia
type WebhookConfig struct {
	InstanceID string    `json:"instance_id"`
	URL        string    `json:"url" validate:"required,url"`
	Events     []string  `json:"events" validate:"required"` // message, status, receipt, etc.
	Secret     string    `json:"secret,omitempty"`           // Para firmar requests
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// WebhookEvent evento que se envía al webhook
type WebhookEvent struct {
	InstanceID string      `json:"instance_id"`
	Event      string      `json:"event"`
	Timestamp  int64       `json:"timestamp"`
	Data       interface{} `json:"data"`
}

// MessageEvent datos de un mensaje recibido
type MessageEvent struct {
	MessageID       string `json:"message_id"`
	From            string `json:"from"`
	FromName        string `json:"from_name,omitempty"`
	To              string `json:"to"`
	IsGroup         bool   `json:"is_group"`
	MessageType     string `json:"message_type"` // text, image, video, audio, document, location
	Text            string `json:"text,omitempty"`
	MediaURL        string `json:"media_url,omitempty"`  // Deprecated: usar MediaData
	MediaData       string `json:"media_data,omitempty"` // Base64 del archivo (si < 5MB)
	Caption         string `json:"caption,omitempty"`
	FileName        string `json:"file_name,omitempty"`   // Nombre del archivo
	MimeType        string `json:"mime_type,omitempty"`   // Tipo MIME del archivo
	FileSize        int64  `json:"file_size,omitempty"`   // Tamaño en bytes
	MediaError      string `json:"media_error,omitempty"` // Error al descargar media
	Latitude        string `json:"latitude,omitempty"`    // Coordenadas de ubicación
	Longitude       string `json:"longitude,omitempty"`
	LocationName    string `json:"location_name,omitempty"`    // Nombre del lugar
	LocationAddress string `json:"location_address,omitempty"` // Dirección del lugar
	Timestamp       int64  `json:"timestamp"`
	IsFromMe        bool   `json:"is_from_me"`
}

// StatusEvent evento de cambio de estado
type StatusEvent struct {
	Status string `json:"status"` // connected, disconnected, authenticated
}

// ReceiptEvent evento de confirmación de lectura/entrega
type ReceiptEvent struct {
	MessageID string   `json:"message_id"`
	From      string   `json:"from"`
	Type      string   `json:"type"` // read, delivered
	Timestamp int64    `json:"timestamp"`
	IDs       []string `json:"ids,omitempty"`
}

// MessageAckEvent es el evento de confirmación para un mensaje enviado de forma asíncrona.
type MessageAckEvent struct {
	Status        string `json:"status"` // "sent" o "failed"
	CorrelationID string `json:"correlation_id"`
	MessageID     string `json:"message_id,omitempty"` // ID de WhatsApp, solo en caso de éxito
	Error         string `json:"error,omitempty"`      // Razón del fallo, solo en caso de error
}
