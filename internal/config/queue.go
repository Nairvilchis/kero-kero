package config

const (
	// QueueName es el nombre de la lista en Redis que se usará como cola de mensajes.
	QueueName = "kero_message_queue"

	// MaxRetries es el número máximo de veces que un trabajo de la cola se reintentará antes de moverlo a la dead-letter queue.
	MaxRetries = 3

	// DeadLetterQueueName es el nombre de la cola de "mensajes muertos" donde se almacenan los trabajos que fallaron repetidamente.
	DeadLetterQueueName = "kero_dead_letter_queue"
)
