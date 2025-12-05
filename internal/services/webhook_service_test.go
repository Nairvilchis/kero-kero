package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kero-kero/internal/models"
	"kero-kero/internal/repository"
	"kero-kero/internal/testutil"
)

func TestWebhookService_SetWebhook(t *testing.T) {
	// Setup
	mr, redisClient := testutil.NewMockRedis(t)
	defer testutil.CleanupRedis(t, mr, redisClient)

	redisRepo := &repository.RedisClient{Client: redisClient}
	webhookRepo := repository.NewWebhookRepository(redisRepo)
	service := NewWebhookService(webhookRepo)

	ctx := context.Background()

	t.Run("crear webhook exitosamente", func(t *testing.T) {
		config := &models.WebhookConfig{
			InstanceID: "test-instance",
			URL:        "https://example.com/webhook",
			Events:     []string{"message", "status"},
			Secret:     "test-secret",
		}

		err := service.SetWebhook(ctx, config)
		require.NoError(t, err)

		// Verificar que se guardó
		saved, err := service.GetWebhook(ctx, "test-instance")
		require.NoError(t, err)
		assert.Equal(t, config.URL, saved.URL)
		assert.Equal(t, config.Events, saved.Events)
		assert.True(t, saved.Enabled)
	})

	t.Run("error cuando falta instance_id", func(t *testing.T) {
		config := &models.WebhookConfig{
			URL:    "https://example.com/webhook",
			Events: []string{"message"},
		}

		err := service.SetWebhook(ctx, config)
		assert.Error(t, err)
	})

	t.Run("usar eventos por defecto si no se especifican", func(t *testing.T) {
		config := &models.WebhookConfig{
			InstanceID: "test-instance-2",
			URL:        "https://example.com/webhook",
		}

		err := service.SetWebhook(ctx, config)
		require.NoError(t, err)

		saved, err := service.GetWebhook(ctx, "test-instance-2")
		require.NoError(t, err)
		assert.Contains(t, saved.Events, "message")
		assert.Contains(t, saved.Events, "status")
		assert.Contains(t, saved.Events, "receipt")
	})
}

func TestWebhookService_SendEvent(t *testing.T) {
	// Setup
	mr, redisClient := testutil.NewMockRedis(t)
	defer testutil.CleanupRedis(t, mr, redisClient)

	redisRepo := &repository.RedisClient{Client: redisClient}
	webhookRepo := repository.NewWebhookRepository(redisRepo)
	service := NewWebhookService(webhookRepo)

	ctx := context.Background()

	t.Run("enviar evento exitosamente", func(t *testing.T) {
		// Crear servidor HTTP mock
		received := false
		var receivedEvent models.WebhookEvent

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			received = true
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.NotEmpty(t, r.Header.Get("X-Webhook-Signature"))

			err := json.NewDecoder(r.Body).Decode(&receivedEvent)
			require.NoError(t, err)

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Configurar webhook
		config := &models.WebhookConfig{
			InstanceID: "test-instance",
			URL:        server.URL,
			Events:     []string{"message"},
			Secret:     "test-secret",
			Enabled:    true,
		}
		err := webhookRepo.Set(ctx, config)
		require.NoError(t, err)

		// Enviar evento
		event := &models.WebhookEvent{
			Event: "message",
			Data: models.MessageEvent{
				MessageID:   "msg-123",
				From:        "5215512345678",
				MessageType: "text",
				Text:        "Hola",
			},
		}

		err = service.SendEvent(ctx, "test-instance", event)
		require.NoError(t, err)

		// Esperar un poco para que se procese
		time.Sleep(100 * time.Millisecond)

		assert.True(t, received)
		assert.Equal(t, "message", receivedEvent.Event)
		assert.Equal(t, "test-instance", receivedEvent.InstanceID)
	})

	t.Run("no enviar si evento no está en la lista", func(t *testing.T) {
		received := false

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			received = true
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := &models.WebhookConfig{
			InstanceID: "test-instance-2",
			URL:        server.URL,
			Events:     []string{"status"}, // Solo status
			Enabled:    true,
		}
		err := webhookRepo.Set(ctx, config)
		require.NoError(t, err)

		event := &models.WebhookEvent{
			Event: "message", // Intentar enviar message
			Data:  models.MessageEvent{},
		}

		err = service.SendEvent(ctx, "test-instance-2", event)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		assert.False(t, received)
	})

	t.Run("no enviar si webhook está deshabilitado", func(t *testing.T) {
		received := false

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			received = true
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := &models.WebhookConfig{
			InstanceID: "test-instance-3",
			URL:        server.URL,
			Events:     []string{"message"},
			Enabled:    false, // Deshabilitado
		}
		err := webhookRepo.Set(ctx, config)
		require.NoError(t, err)

		event := &models.WebhookEvent{
			Event: "message",
			Data:  models.MessageEvent{},
		}

		err = service.SendEvent(ctx, "test-instance-3", event)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		assert.False(t, received)
	})
}

func TestWebhookService_DeleteWebhook(t *testing.T) {
	mr, redisClient := testutil.NewMockRedis(t)
	defer testutil.CleanupRedis(t, mr, redisClient)

	redisRepo := &repository.RedisClient{Client: redisClient}
	webhookRepo := repository.NewWebhookRepository(redisRepo)
	service := NewWebhookService(webhookRepo)

	ctx := context.Background()

	// Crear webhook
	config := &models.WebhookConfig{
		InstanceID: "test-instance",
		URL:        "https://example.com/webhook",
		Events:     []string{"message"},
	}
	err := service.SetWebhook(ctx, config)
	require.NoError(t, err)

	// Eliminar
	err = service.DeleteWebhook(ctx, "test-instance")
	require.NoError(t, err)

	// Verificar que no existe
	_, err = service.GetWebhook(ctx, "test-instance")
	assert.Error(t, err)
}
