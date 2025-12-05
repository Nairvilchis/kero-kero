package models

// PublishStatusRequest representa la solicitud para publicar un estado
type PublishStatusRequest struct {
	Type     string `json:"type" validate:"required"` // text, image, video
	Content  string `json:"content,omitempty"`        // Para texto
	MediaURL string `json:"media_url,omitempty"`      // Para imagen/video
	Caption  string `json:"caption,omitempty"`
}

// StatusResponse representa la respuesta al publicar un estado
type StatusResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// StatusInfo representa información de un estado
type StatusInfo struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"`
	Caption   string `json:"caption,omitempty"`
}
