package models

// Message representa un mensaje de WhatsApp
// Message representa un mensaje de WhatsApp
type Message struct {
	ID         string `json:"id"`
	InstanceID string `json:"instance_id"`
	From       string `json:"from"`
	To         string `json:"to"`
	Sender     string `json:"sender,omitempty"` // Quien lo envió realmente (en grupos)
	Timestamp  int64  `json:"timestamp"`        // Unix timestamp
	Type       string `json:"type"`
	Content    string `json:"content"`             // Texto o descripción
	PushName   string `json:"push_name,omitempty"` // Nombre público del usuario
	IsFromMe   bool   `json:"is_from_me"`
	Status     string `json:"status"`          // sent, delivered, read
	Error      string `json:"error,omitempty"` // Para errores de envío
}

// MessageType representa los tipos de mensajes soportados
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeVideo    MessageType = "video"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeDocument MessageType = "document"
	MessageTypeLocation MessageType = "location"
	MessageTypeContact  MessageType = "contact"
	MessageTypeReaction MessageType = "reaction"
)

// MessageStatus representa los estados de un mensaje
type MessageStatus string

const (
	MessageStatusPending   MessageStatus = "pending"
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

// SendTextRequest representa la solicitud para enviar mensaje de texto
type SendTextRequest struct {
	Phone   string `json:"phone" validate:"required"`
	Message string `json:"message" validate:"required"`
}

// SendMediaRequest representa la solicitud para enviar medios
type SendMediaRequest struct {
	Phone    string `json:"phone" validate:"required"`
	MediaURL string `json:"media_url,omitempty"`
	Caption  string `json:"caption,omitempty"`
	FileName string `json:"file_name,omitempty"`
}

// SendLocationRequest representa la solicitud para enviar ubicación
type SendLocationRequest struct {
	Phone     string  `json:"phone" validate:"required"`
	Latitude  float64 `json:"latitude" validate:"required"`
	Longitude float64 `json:"longitude" validate:"required"`
	Name      string  `json:"name,omitempty"`
	Address   string  `json:"address,omitempty"`
}

// SendContactRequest representa la solicitud para enviar un contacto
type SendContactRequest struct {
	Phone       string `json:"phone" validate:"required"`
	DisplayName string `json:"display_name" validate:"required"`
	VCard       string `json:"vcard" validate:"required"`
}

// ReactionRequest representa la solicitud para reaccionar a un mensaje
type ReactionRequest struct {
	Phone     string `json:"phone" validate:"required"`
	MessageID string `json:"message_id" validate:"required"`
	Emoji     string `json:"emoji" validate:"required"` // Emoji o string vacío para eliminar reacción
}

// RevokeRequest representa la solicitud para eliminar un mensaje
type RevokeRequest struct {
	Phone     string `json:"phone" validate:"required"`
	MessageID string `json:"message_id" validate:"required"`
}

// DownloadMediaRequest representa la solicitud para descargar multimedia
type DownloadMediaRequest struct {
	Type          string `json:"type" validate:"required"` // image, video, audio, document, sticker
	DirectPath    string `json:"direct_path"`
	Url           string `json:"url"`
	MediaKey      string `json:"media_key"`       // Base64
	FileEncSha256 string `json:"file_enc_sha256"` // Base64
	FileSha256    string `json:"file_sha256"`     // Base64
	FileLength    uint64 `json:"file_length"`
	Mimetype      string `json:"mimetype"`
}

// MessageResponse representa la respuesta al enviar un mensaje
type MessageResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Status    string `json:"status,omitempty"`
	Error     string `json:"error,omitempty"`
}

// QueuedResponse es la respuesta para una operación de mensaje que ha sido encolada.
type QueuedResponse struct {
	Status        string `json:"status"`
	CorrelationID string `json:"correlation_id"`
}

// SendTextWithTypingRequest solicitud para enviar texto con simulación de escritura
type SendTextWithTypingRequest struct {
	Phone          string `json:"phone" validate:"required"`
	Message        string `json:"message" validate:"required"`
	TypingDuration *int   `json:"typing_duration,omitempty"` // En milisegundos, opcional (se calcula automáticamente si no se proporciona)
}

// MarkAsReadRequest solicitud para marcar mensajes como leídos
type MarkAsReadRequest struct {
	MessageID string `json:"message_id" validate:"required"`
	ChatJID   string `json:"chat_jid" validate:"required"`   // JID del chat (ej: 5215512345678@s.whatsapp.net)
	SenderJID string `json:"sender_jid" validate:"required"` // JID del remitente
	Timestamp int64  `json:"timestamp" validate:"required"`  // Unix timestamp del mensaje
}
