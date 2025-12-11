package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"kero-kero/internal/models"
	"kero-kero/internal/repository"
)

// QueueService gestiona la cola de mensajes
type QueueService struct {
	redisClient *repository.RedisClient
	msgService  *MessageService
	workers     int
	stopChan    chan struct{}
}

// NewQueueService crea un nuevo servicio de colas
func NewQueueService(redisClient *repository.RedisClient, msgService *MessageService) *QueueService {
	return &QueueService{
		redisClient: redisClient,
		msgService:  msgService,
		workers:     3, // Default 3 workers
		stopChan:    make(chan struct{}),
	}
}

// Start inicia los workers
func (s *QueueService) Start() {
	log.Info().Int("workers", s.workers).Msg("Iniciando workers de cola de mensajes")
	for i := 0; i < s.workers; i++ {
		go s.workerLoop(i)
	}
}

// Stop detiene los workers
func (s *QueueService) Stop() {
	close(s.stopChan)
}

// EnqueueMessage encola un mensaje para envío asíncrono
func (s *QueueService) EnqueueMessage(ctx context.Context, instanceID string, msgType models.MessageType, payload interface{}) (string, error) {
	msgID := fmt.Sprintf("msg_%d", time.Now().UnixNano())
	queuedMsg := &models.QueuedMessage{
		ID:         msgID,
		InstanceID: instanceID,
		Type:       msgType,
		Payload:    payload,
		CreatedAt:  time.Now().Unix(),
		Attempts:   0,
	}

	jsonBytes, err := json.Marshal(queuedMsg)
	if err != nil {
		return "", err
	}

	queueKey := "queue:messages"
	if err := s.redisClient.EnqueueMessage(ctx, queueKey, string(jsonBytes)); err != nil {
		return "", err
	}

	return msgID, nil
}

func (s *QueueService) workerLoop(id int) {
	queueKey := "queue:messages"
	log.Debug().Int("worker_id", id).Msg("Worker iniciado")

	for {
		select {
		case <-s.stopChan:
			log.Debug().Int("worker_id", id).Msg("Worker detenido")
			return
		default:
			// Bloqueante con timeout de 2 segundos para permitir shutdown limpio
			data, err := s.redisClient.DequeueMessage(context.Background(), queueKey, 2*time.Second)
			if err != nil {
				// Redis timeout es normal
				if err.Error() == "redis: nil" {
					continue
				}
				// Otros errores (conexión, etc)
				// Si no es timeout, hacer un pequeño sleep para no saturar CPU en caso de error loop
				time.Sleep(1 * time.Second)
				continue
			}

			if data == "" {
				continue
			}

			s.processMessage(data)
		}
	}
}

func (s *QueueService) processMessage(data string) {
	var msg models.QueuedMessage
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		log.Error().Err(err).Str("data", data).Msg("Error deserializando mensaje de cola")
		return
	}

	ctx := context.Background()
	log.Debug().Str("msg_id", msg.ID).Str("type", string(msg.Type)).Msg("Procesando mensaje de cola")

	var err error

	// Convertir payload map[string]interface{} al struct correcto
	// JSON unmarshal lo deja como map, necesitamos re-marshalear o usar mapstructure
	// Para simplicidad, usamos Marshal/Unmarshal de nuevo
	payloadBytes, _ := json.Marshal(msg.Payload)

	switch msg.Type {
	case models.MessageTypeText:
		var req models.SendTextRequest
		if err = json.Unmarshal(payloadBytes, &req); err == nil {
			_, err = s.msgService.SendText(ctx, msg.InstanceID, &req)
		}
	case models.MessageTypeImage:
		var req models.SendMediaRequest
		if err = json.Unmarshal(payloadBytes, &req); err == nil {
			_, err = s.msgService.SendImage(ctx, msg.InstanceID, &req)
		}
	case models.MessageTypeVideo:
		var req models.SendMediaRequest
		if err = json.Unmarshal(payloadBytes, &req); err == nil {
			_, err = s.msgService.SendVideo(ctx, msg.InstanceID, &req)
		}
	case models.MessageTypeAudio:
		var req models.SendMediaRequest
		if err = json.Unmarshal(payloadBytes, &req); err == nil {
			_, err = s.msgService.SendAudio(ctx, msg.InstanceID, &req)
		}
	case models.MessageTypeDocument:
		var req models.SendMediaRequest
		if err = json.Unmarshal(payloadBytes, &req); err == nil {
			_, err = s.msgService.SendDocument(ctx, msg.InstanceID, &req)
		}
	case models.MessageTypeLocation:
		var req models.SendLocationRequest
		if err = json.Unmarshal(payloadBytes, &req); err == nil {
			_, err = s.msgService.SendLocation(ctx, msg.InstanceID, &req)
		}
	default:
		log.Warn().Str("type", string(msg.Type)).Msg("Tipo de mensaje desconocido en cola")
		return
	}

	if err != nil {
		log.Error().Err(err).Str("msg_id", msg.ID).Msg("Error enviando mensaje desde cola")
		// Aquí se podría implementar lógica de reintentos (incrementar msg.Attempts y re-encolar si < MaxAttempts)
		// Por ahora lo descartamos para evitar bucles infinitos
	}
}
