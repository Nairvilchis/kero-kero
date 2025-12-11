package services

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"kero-kero/internal/config"
	"kero-kero/internal/models"
	"kero-kero/internal/repository"
	"kero-kero/pkg/errors"
)

// QueueWorker procesa trabajos de la cola de mensajes de Redis.
type QueueWorker struct {
	redisClient    *repository.RedisClient
	messageService *MessageService
	webhookService *WebhookService
}

// NewQueueWorker crea un nuevo consumidor de la cola.
func NewQueueWorker(redisClient *repository.RedisClient, messageService *MessageService, webhookService *WebhookService) *QueueWorker {
	return &QueueWorker{
		redisClient:    redisClient,
		messageService: messageService,
		webhookService: webhookService,
	}
}

// Start inicia el worker en una goroutine para que escuche la cola.
func (w *QueueWorker) Start() {
	log.Info().Msg("Iniciando el Queue Worker para procesar mensajes asíncronos...")
	go w.listenForJobs()
}

func (w *QueueWorker) listenForJobs() {
	for {
		result, err := w.redisClient.Client.BRPop(context.Background(), 0, config.QueueName).Result()
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
			w.moveToDeadLetterQueue(jobData, "deserialization_failed")
			continue
		}

		log.Info().Str("correlation_id", job.CorrelationID).Msg("Procesando nuevo trabajo de la cola")
		w.processJob(&job, jobData)
	}
}

func (w *QueueWorker) processJob(job *models.MessageJob, originalJobData string) {
	ctx := context.Background()
	var err error
	var response *models.MessageResponse

	switch job.Type {
	case models.MessageJobTypeText:
		var payload models.SendTextRequest
		if err = json.Unmarshal(job.Payload, &payload); err == nil {
			response, err = w.messageService.sendSyncTextMessage(ctx, w.messageService.waManager.GetClient(job.InstanceID), &payload)
		}
	case models.MessageJobTypeImage:
		var payload models.SendMediaRequest
		if err = json.Unmarshal(job.Payload, &payload); err == nil {
			response, err = w.messageService.sendSyncImageMessage(ctx, w.messageService.waManager.GetClient(job.InstanceID), &payload)
		}
	case models.MessageJobTypeVideo:
		var payload models.SendMediaRequest
		if err = json.Unmarshal(job.Payload, &payload); err == nil {
			response, err = w.messageService.sendSyncVideoMessage(ctx, w.messageService.waManager.GetClient(job.InstanceID), &payload)
		}
	case models.MessageJobTypeAudio:
		var payload models.SendMediaRequest
		if err = json.Unmarshal(job.Payload, &payload); err == nil {
			response, err = w.messageService.sendSyncAudioMessage(ctx, w.messageService.waManager.GetClient(job.InstanceID), &payload)
		}
	case models.MessageJobTypeDocument:
		var payload models.SendMediaRequest
		if err = json.Unmarshal(job.Payload, &payload); err == nil {
			response, err = w.messageService.sendSyncDocumentMessage(ctx, w.messageService.waManager.GetClient(job.InstanceID), &payload)
		}
	default:
		err = errors.New(http.StatusBadRequest, "Tipo de trabajo de mensaje no soportado")
	}

	if err != nil {
		log.Error().Err(err).Str("correlation_id", job.CorrelationID).Msg("Fallo al procesar el trabajo")
		w.handleJobFailure(ctx, job, originalJobData, err)
		return
	}

	log.Info().Str("correlation_id", job.CorrelationID).Str("message_id", response.MessageID).Msg("Trabajo procesado con éxito")
	w.sendSuccessWebhook(ctx, job, response.MessageID)
}

func (w *QueueWorker) handleJobFailure(ctx context.Context, job *models.MessageJob, originalJobData string, processingError error) {
	if job.RetryCount < config.MaxRetries {
		job.RetryCount++
		log.Warn().Str("correlation_id", job.CorrelationID).Int("retry_count", job.RetryCount).Msg("Reintentando trabajo...")
		jobBytes, err := json.Marshal(job)
		if err != nil {
			log.Error().Err(err).Msg("Error al serializar trabajo para reintento. Moviendo a dead-letter queue.")
			w.moveToDeadLetterQueue(originalJobData, "retry_serialization_failed")
			w.sendFailureWebhook(ctx, job, "retry_serialization_failed")
			return
		}
		time.Sleep(time.Duration(job.RetryCount*2) * time.Second)
		if err := w.redisClient.Client.LPush(ctx, config.QueueName, jobBytes).Err(); err != nil {
			log.Error().Err(err).Msg("Error al re-encolar trabajo. Moviendo a dead-letter queue.")
			w.moveToDeadLetterQueue(originalJobData, "requeue_failed")
			w.sendFailureWebhook(ctx, job, "requeue_failed")
		}
	} else {
		log.Error().Str("correlation_id", job.CorrelationID).Msg("El trabajo ha alcanzado el número máximo de reintentos. Moviendo a dead-letter queue.")
		w.moveToDeadLetterQueue(originalJobData, processingError.Error())
		w.sendFailureWebhook(ctx, job, processingError.Error())
	}
}

func (w *QueueWorker) moveToDeadLetterQueue(jobData string, reason string) {
	err := w.redisClient.Client.LPush(context.Background(), config.DeadLetterQueueName, jobData).Err()
	if err != nil {
		log.Error().Err(err).Msg("¡Fallo CRÍTICO! No se pudo mover el trabajo a la dead-letter queue.")
	}
}

func (w *QueueWorker) sendSuccessWebhook(ctx context.Context, job *models.MessageJob, messageID string) {
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
	if err := w.webhookService.SendEvent(ctx, job.InstanceID, webhookEvent); err != nil {
		log.Error().Err(err).Str("correlation_id", job.CorrelationID).Msg("Error enviando webhook de éxito")
	}
}

func (w *QueueWorker) sendFailureWebhook(ctx context.Context, job *models.MessageJob, errorReason string) {
	ackEvent := models.MessageAckEvent{
		Status:        "failed",
		CorrelationID: job.CorrelationID,
		Error:         errorReason,
	}
	webhookEvent := &models.WebhookEvent{
		InstanceID: job.InstanceID,
		Event:      "message_ack",
		Timestamp:  time.Now().Unix(),
		Data:       ackEvent,
	}
	if err := w.webhookService.SendEvent(ctx, job.InstanceID, webhookEvent); err != nil {
		log.Error().Err(err).Str("correlation_id", job.CorrelationID).Msg("Error enviando webhook de fallo")
	}
}
