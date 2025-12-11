package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"kero-kero/internal/config"
	"kero-kero/internal/models"
	"kero-kero/internal/repository"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
)

// QueueService gestiona la cola de mensajes asíncronos y procesa los trabajos en segundo plano.
type QueueService struct {
	redisClient    *repository.RedisClient
	waManager      *whatsapp.Manager
	webhookService *WebhookService
	messageService *MessageService
}

// NewQueueService crea un nuevo servicio de cola y consumidor.
func NewQueueService(redisClient *repository.RedisClient, waManager *whatsapp.Manager, webhookService *WebhookService, messageService *MessageService) *QueueService {
	return &QueueService{
		redisClient:    redisClient,
		waManager:      waManager,
		webhookService: webhookService,
		messageService: messageService,
	}
}

// EnqueueMessage encola un nuevo trabajo de envío de mensaje.
func (s *QueueService) EnqueueMessage(ctx context.Context, instanceID string, jobType models.MessageJobType, payload interface{}) (*models.QueuedResponse, error) {
	correlationID := uuid.NewString()
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails("Error al serializar el payload del trabajo")
	}
	job := &models.MessageJob{
		InstanceID:    instanceID,
		CorrelationID: correlationID,
		Type:          jobType,
		Payload:       payloadBytes,
		RetryCount:    0,
	}
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails("Error al serializar el trabajo")
	}
	if err := s.redisClient.Client.LPush(ctx, config.QueueName, jobBytes).Err(); err != nil {
		log.Error().Err(err).Str("instance_id", instanceID).Msg("Error al encolar el trabajo en Redis")
		return nil, errors.ErrInternalServer.WithDetails("Error al encolar el mensaje")
	}
	log.Info().Str("instance_id", instanceID).Str("correlation_id", correlationID).Str("type", string(jobType)).Msg("Mensaje encolado para envío asíncrono")
	return &models.QueuedResponse{Status: "queued", CorrelationID: correlationID}, nil
}

// Start inicia el worker en una goroutine para que escuche la cola.
func (s *QueueService) Start() {
	log.Info().Msg("Iniciando el Queue Service Worker para procesar mensajes asíncronos...")
	go s.listenForJobs()
}

func (s *QueueService) listenForJobs() {
	for {
		result, err := s.redisClient.Client.BRPop(context.Background(), 0, config.QueueName).Result()
		if err != nil {
			log.Error().Err(err).Msg("Error obteniendo trabajo de la cola de Redis. Reintentando en 5 segundos...")
			time.Sleep(5 * time.Second)
			continue
		}
		if len(result) < 2 {
			continue
		}
		jobData := result[1]
		var job models.MessageJob
		if err := json.Unmarshal([]byte(jobData), &job); err != nil {
			log.Error().Err(err).Msg("Error al deserializar el trabajo de la cola. Moviendo a dead-letter queue.")
			s.moveToDeadLetterQueue(jobData, "deserialization_failed")
			continue
		}
		log.Info().Str("correlation_id", job.CorrelationID).Msg("Procesando nuevo trabajo de la cola")
		s.processJob(&job, jobData)
	}
}

func (s *QueueService) processJob(job *models.MessageJob, originalJobData string) {
	ctx := context.Background()
	var err error
	var response *models.MessageResponse

	client := s.waManager.GetClient(job.InstanceID)
	if client == nil || !client.WAClient.IsLoggedIn() {
		err = errors.ErrNotAuthenticated
	} else {
		switch job.Type {
		case models.MessageJobTypeText:
			var payload models.SendTextRequest
			if err = json.Unmarshal(job.Payload, &payload); err == nil {
				response, err = s.messageService.SendText(ctx, job.InstanceID, &payload)
			}
		// Agregaremos los otros tipos de mensajes aquí
		default:
			err = errors.New(http.StatusBadRequest, "Tipo de trabajo de mensaje no soportado")
		}
	}

	if err != nil {
		log.Error().Err(err).Str("correlation_id", job.CorrelationID).Msg("Fallo al procesar el trabajo")
		s.handleJobFailure(ctx, job, originalJobData, err)
		return
	}

	log.Info().Str("correlation_id", job.CorrelationID).Str("message_id", response.MessageID).Msg("Trabajo procesado con éxito")
	s.sendSuccessWebhook(ctx, job, response.MessageID)
}

func (s *QueueService) handleJobFailure(ctx context.Context, job *models.MessageJob, originalJobData string, processingError error) {
	if job.RetryCount < config.MaxRetries {
		job.RetryCount++
		log.Warn().Str("correlation_id", job.CorrelationID).Int("retry_count", job.RetryCount).Msg("Reintentando trabajo...")
		jobBytes, err := json.Marshal(job)
		if err != nil {
			log.Error().Err(err).Msg("Error al serializar trabajo para reintento. Moviendo a dead-letter queue.")
			s.moveToDeadLetterQueue(originalJobData, "retry_serialization_failed")
			s.sendFailureWebhook(ctx, job, "retry_serialization_failed")
			return
		}
		time.Sleep(time.Duration(job.RetryCount*2) * time.Second)
		if err := s.redisClient.Client.LPush(ctx, config.QueueName, jobBytes).Err(); err != nil {
			log.Error().Err(err).Msg("Error al re-encolar trabajo. Moviendo a dead-letter queue.")
			s.moveToDeadLetterQueue(originalJobData, "requeue_failed")
			s.sendFailureWebhook(ctx, job, "requeue_failed")
		}
	} else {
		log.Error().Str("correlation_id", job.CorrelationID).Msg("El trabajo ha alcanzado el número máximo de reintentos. Moviendo a dead-letter queue.")
		s.moveToDeadLetterQueue(originalJobData, processingError.Error())
		s.sendFailureWebhook(ctx, job, processingError.Error())
	}
}

func (s *QueueService) moveToDeadLetterQueue(jobData string, reason string) {
	err := s.redisClient.Client.LPush(context.Background(), config.DeadLetterQueueName, jobData).Err()
	if err != nil {
		log.Error().Err(err).Msg("¡Fallo CRÍTICO! No se pudo mover el trabajo a la dead-letter queue.")
	}
}

func (s *QueueService) sendSuccessWebhook(ctx context.Context, job *models.MessageJob, messageID string) {
	ackEvent := models.MessageAckEvent{
		Status:        "sent",
		CorrelationID: job.CorrelationID,
		MessageID:     messageID,
	}
	webhookEvent := &models.WebhookEvent{
		InstanceID: job.InstanceID,
		Event:      "message_ack",
		Timestamp:  time.Now().Unix(),
		Data:       ackEvent,
	}
	if err := s.webhookService.SendEvent(ctx, job.InstanceID, webhookEvent); err != nil {
		log.Error().Err(err).Str("correlation_id", job.CorrelationID).Msg("Error enviando webhook de éxito")
	}
}

func (s *QueueService) sendFailureWebhook(ctx context.Context, job *models.MessageJob, errorReason error) {
	ackEvent := models.MessageAckEvent{
		Status:        "failed",
		CorrelationID: job.CorrelationID,
		Error:         errorReason.Error(),
	}
	webhookEvent := &models.WebhookEvent{
		InstanceID: job.InstanceID,
		Event:      "message_ack",
		Timestamp:  time.Now().Unix(),
		Data:       ackEvent,
	}
	if err := s.webhookService.SendEvent(ctx, job.InstanceID, webhookEvent); err != nil {
		log.Error().Err(err).Str("correlation_id", job.CorrelationID).Msg("Error enviando webhook de fallo")
	}
}
