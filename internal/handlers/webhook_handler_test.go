package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kero-kero/internal/models"
	"kero-kero/internal/repository"
	"kero-kero/internal/services"
	"kero-kero/internal/testutil"
)

func TestWebhookHandler_SetWebhook(t *testing.T) {
	// Setup
	mr, redisClient := testutil.NewMockRedis(t)
	defer testutil.CleanupRedis(t, mr, redisClient)

	redisRepo := &repository.RedisClient{Client: redisClient}
	webhookRepo := repository.NewWebhookRepository(redisRepo)
	service := services.NewWebhookService(webhookRepo)
	handler := NewWebhookHandler(service)

	t.Run("configurar webhook exitosamente", func(t *testing.T) {
		config := models.WebhookConfig{
			URL:    "https://example.com/webhook",
			Events: []string{"message", "status"},
			Secret: "test-secret",
		}

		body, err := json.Marshal(config)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/instances/test-instance/webhook", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Configurar chi context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("instanceID", "test-instance")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.SetWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]bool
		err = json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response["success"])
	})

	t.Run("error con JSON inválido", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/instances/test-instance/webhook", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("instanceID", "test-instance")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.SetWebhook(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestWebhookHandler_GetWebhook(t *testing.T) {
	mr, redisClient := testutil.NewMockRedis(t)
	defer testutil.CleanupRedis(t, mr, redisClient)

	redisRepo := &repository.RedisClient{Client: redisClient}
	webhookRepo := repository.NewWebhookRepository(redisRepo)
	service := services.NewWebhookService(webhookRepo)
	handler := NewWebhookHandler(service)

	// Crear webhook primero
	config := &models.WebhookConfig{
		InstanceID: "test-instance",
		URL:        "https://example.com/webhook",
		Events:     []string{"message"},
	}
	err := service.SetWebhook(context.Background(), config)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/instances/test-instance/webhook", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("instanceID", "test-instance")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetWebhook(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.WebhookConfig
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, config.URL, response.URL)
}

func TestWebhookHandler_DeleteWebhook(t *testing.T) {
	mr, redisClient := testutil.NewMockRedis(t)
	defer testutil.CleanupRedis(t, mr, redisClient)

	redisRepo := &repository.RedisClient{Client: redisClient}
	webhookRepo := repository.NewWebhookRepository(redisRepo)
	service := services.NewWebhookService(webhookRepo)
	handler := NewWebhookHandler(service)

	// Crear webhook primero
	config := &models.WebhookConfig{
		InstanceID: "test-instance",
		URL:        "https://example.com/webhook",
		Events:     []string{"message"},
	}
	err := service.SetWebhook(context.Background(), config)
	require.NoError(t, err)

	req := httptest.NewRequest("DELETE", "/instances/test-instance/webhook", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("instanceID", "test-instance")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.DeleteWebhook(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]bool
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.True(t, response["success"])
}
