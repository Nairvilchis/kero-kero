package models

// CallSettings representa la configuración de llamadas
type CallSettings struct {
	AutoReject bool `json:"auto_reject"` // Rechazar llamadas automáticamente
}

// CallEvent representa un evento de llamada
type CallEvent struct {
	From      string `json:"from"`
	Timestamp int64  `json:"timestamp"`
	IsVideo   bool   `json:"is_video"`
	Status    string `json:"status"` // incoming, rejected, missed
}
