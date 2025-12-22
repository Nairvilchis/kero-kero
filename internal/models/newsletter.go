package models

// CreateNewsletterRequest representa la solicitud para crear un canal
type CreateNewsletterRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Picture     string `json:"picture,omitempty"` // Base64
}

// NewsletterMetadata representa la información de un canal
type NewsletterMetadata struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	InviteCode      string `json:"invite_code,omitempty"`
	InviteLink      string `json:"invite_link,omitempty"`
	SubscriberCount int    `json:"subscriber_count"`
	Role            string `json:"role,omitempty"` // admin, member, etc.
}

// SendNewsletterMessageRequest solicitud para enviar mensaje a un canal
type SendNewsletterMessageRequest struct {
	JID      string `json:"jid" validate:"required"`
	Message  string `json:"message"`           // Texto o subtítulo (caption)
	Type     string `json:"type"`              // text (default), image, video
	MediaURL string `json:"media_url"`         // URL del medio a enviar
	Payload  string `json:"payload,omitempty"` // Base64 del medio (opcional si hay media_url)
}
