package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"kero-kero/internal/models"
	"kero-kero/internal/repository"
	"kero-kero/pkg/errors"
)

type WebhookService struct {
	webhookRepo *repository.WebhookRepository
	httpClient  *http.Client
}

func NewWebhookService(webhookRepo *repository.WebhookRepository) *WebhookService {
	return &WebhookService{
		webhookRepo: webhookRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetWebhook configura el webhook para una instancia
func (s *WebhookService) SetWebhook(ctx context.Context, config *models.WebhookConfig) error {
	if config.InstanceID == "" {
		return errors.ErrBadRequest.WithDetails("instance_id requerido")
	}

	if len(config.Events) == 0 {
		config.Events = []string{"message", "status", "receipt"}
	}

	config.Enabled = true
	return s.webhookRepo.Set(ctx, config)
}

// GetWebhook obtiene la configuración de webhook
func (s *WebhookService) GetWebhook(ctx context.Context, instanceID string) (*models.WebhookConfig, error) {
	return s.webhookRepo.Get(ctx, instanceID)
}

// DeleteWebhook elimina la configuración de webhook
func (s *WebhookService) DeleteWebhook(ctx context.Context, instanceID string) error {
	return s.webhookRepo.Delete(ctx, instanceID)
}

// SendEvent envía un evento al webhook configurado
func (s *WebhookService) SendEvent(ctx context.Context, instanceID string, event *models.WebhookEvent) error {
	config, err := s.webhookRepo.Get(ctx, instanceID)
	if err != nil {
		log.Debug().Err(err).Str("instance_id", instanceID).Msg("No hay webhook configurado")
		return nil // No es un error crítico
	}

	if !config.Enabled {
		return nil
	}

	// Verificar si el evento está en la lista de eventos configurados
	eventEnabled := false
	for _, e := range config.Events {
		if e == event.Event || e == "all" {
			eventEnabled = true
			break
		}
	}

	if !eventEnabled {
		return nil
	}

	event.InstanceID = instanceID
	event.Timestamp = time.Now().Unix()

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshaling event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", config.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Kero-Kero-Webhook/2.0")

	// Firmar el payload si hay secret configurado
	if config.Secret != "" {
		signature := s.signPayload(payload, config.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Str("instance_id", instanceID).Str("url", config.URL).Msg("Error enviando webhook")
		return fmt.Errorf("error sending webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Warn().
			Int("status", resp.StatusCode).
			Str("instance_id", instanceID).
			Str("url", config.URL).
			Msg("Webhook respondió con error")
	}

	return nil
}

// signPayload firma el payload con HMAC-SHA256
func (s *WebhookService) signPayload(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
