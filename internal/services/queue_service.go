package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"kero-kero/internal/models"
	"kero-kero/internal/repository"
	"kero-kero/pkg/errors"
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
	processingKey := fmt.Sprintf("queue:processing:%d", id)
	log.Debug().Int("worker_id", id).Msg("Worker iniciado")

	for {
		select {
		case <-s.stopChan:
			log.Debug().Int("worker_id", id).Msg("Worker detenido")
			return
		default:
			// Usar pop confiable para evitar pérdida de mensajes en crashes
			data, err := s.redisClient.DequeueMessageReliable(context.Background(), queueKey, processingKey, 2*time.Second)
			if err != nil {
				// redis.Nil es normal cuando el timeout ocurre y la cola está vacía
				if err.Error() == "redis: nil" {
					continue
				}
				log.Error().Err(err).Int("worker_id", id).Msg("Error extrayendo de la cola")
				time.Sleep(1 * time.Second)
				continue
			}

			if data == "" {
				continue
			}

			// Procesar el mensaje
			if err := s.processMessage(data); err != nil {
				if err == errors.ErrRateLimitReached {
					log.Warn().Int("worker_id", id).Msg("Rate limit alcanzado para la instancia. Re-encolando con delay.")
					s.handleRateLimitRetry(data)
				} else {
					log.Error().Err(err).Int("worker_id", id).Msg("Error procesando mensaje, re-encolando si es posible")
					s.handleRetry(data)
				}
			}

			// Una vez procesado (con éxito o fallido tras reintentos), quitar de la cola de procesamiento
			if err := s.redisClient.AckMessage(context.Background(), processingKey, data); err != nil {
				log.Error().Err(err).Int("worker_id", id).Msg("Error haciendo ACK de mensaje")
			}
		}
	}
}

func (s *QueueService) processMessage(data string) error {
	var msg models.QueuedMessage
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		log.Error().Err(err).Str("data", data).Msg("Error deserializando mensaje de cola")
		return nil // No reintentamos error de formato
	}

	ctx := context.Background()

	// Verificar Rate Limit (20 mensajes por minuto por instancia)
	// He implementado esto aquí para que sea la primera línea de defensa antes de tocar el cliente WA.
	allowed, err := s.redisClient.CheckRateLimit(ctx, msg.InstanceID, 20, 60*time.Second)
	if err != nil {
		log.Error().Err(err).Str("instance_id", msg.InstanceID).Msg("Error verificando rate limit")
	} else if !allowed {
		return errors.ErrRateLimitReached
	}

	log.Debug().Str("msg_id", msg.ID).Str("type", string(msg.Type)).Msg("Procesando mensaje de cola")

	// Convertir payload map[string]interface{} al struct correcto
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
		return nil
	}

	if err != nil {
		log.Error().Err(err).Str("msg_id", msg.ID).Msg("Error enviando mensaje desde cola")
		return err
	}

	return nil
}

func (s *QueueService) handleRetry(data string) {
	var msg models.QueuedMessage
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		return
	}

	if msg.Attempts >= 3 {
		log.Error().Str("msg_id", msg.ID).Int("attempts", msg.Attempts).Msg("Mensaje fallido tras máximo de reintentos")
		return
	}

	msg.Attempts++
	// Esperar un poco antes de re-encolar (backoff simple)
	time.Sleep(time.Duration(msg.Attempts) * 2 * time.Second)

	jsonBytes, _ := json.Marshal(msg)
	s.redisClient.EnqueueMessage(context.Background(), "queue:messages", string(jsonBytes))
	log.Info().Str("msg_id", msg.ID).Int("attempt", msg.Attempts).Msg("Mensaje re-encolado para reintento")
}

func (s *QueueService) handleRateLimitRetry(data string) {
	// En caso de rate limit, re-encolamos sin penalizar "Attempts"
	// Pero esperamos un poco para dejar que la ventana de tiempo se limpie.
	time.Sleep(5 * time.Second)
	s.redisClient.EnqueueMessage(context.Background(), "queue:messages", data)
}
