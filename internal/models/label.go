package models

// Label representa una etiqueta de WhatsApp Business
type Label struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color int32  `json:"color"` // Índice de color de WhatsApp
	Count int    `json:"count,omitempty"`
}

// LabelActionRequest para asignar o remover una etiqueta de un chat
type LabelActionRequest struct {
	LabelID string `json:"label_id" validate:"required"`
	ChatJID string `json:"chat_jid" validate:"required"`
	Action  string `json:"action" validate:"required,oneof=add remove"`
}

// CreateLabelRequest para crear o editar una etiqueta
type CreateLabelRequest struct {
	Name  string `json:"name" validate:"required"`
	Color int32  `json:"color"` // 0-19 usualmente
}

// AutoLabelRule define una regla para etiquetado automático
type AutoLabelRule struct {
	Keywords []string `json:"keywords"` // Palabras que activan la regla
	LabelID  string   `json:"label_id"` // ID de la etiqueta a aplicar
}
