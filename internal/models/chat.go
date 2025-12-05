package models

// ArchiveChatRequest solicitud para archivar/desarchivar chat
type ArchiveChatRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Archived bool   `json:"archived"`
}

// UpdateStatusRequest solicitud para actualizar el estado de texto (About)
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

// Chat representa un chat de WhatsApp
type Chat struct {
	JID             string `json:"jid"`
	Name            string `json:"name"`
	ChatType        string `json:"chat_type"` // "private", "group", "channel", "status"
	UnreadCount     int    `json:"unread_count"`
	LastMessageTime int64  `json:"last_message_time"`
	LastMessage     string `json:"last_message"`
	IsArchived      bool   `json:"is_archived"`
	IsPinned        bool   `json:"is_pinned"`
	IsMuted         bool   `json:"is_muted"`
}

// ChatHistoryResponse respuesta con historial de mensajes
type ChatHistoryResponse struct {
	Messages []Message `json:"messages"`
}
