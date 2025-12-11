package models

// QueuedMessage representa un mensaje encolado en Redis
type QueuedMessage struct {
	ID         string      `json:"id"`
	InstanceID string      `json:"instance_id"`
	Type       MessageType `json:"type"`
	Payload    interface{} `json:"payload"`
	CreatedAt  int64       `json:"created_at"`
	Attempts   int         `json:"attempts"`
}

// QueueMessagePayloads wrappers para serializar diferentes tipos de request
// Nota: Usamos interface{} en QueuedMessage.Payload, y al deserializar
// checkeamos el Type para convertir al request concreto.
