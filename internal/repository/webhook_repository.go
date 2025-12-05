package repository

import (
	"context"
	"encoding/json"
	"time"

	"kero-kero/internal/models"
)

type WebhookRepository struct {
	redis *RedisClient
}

func NewWebhookRepository(redis *RedisClient) *WebhookRepository {
	return &WebhookRepository{redis: redis}
}

// Set guarda la configuración de webhook para una instancia
func (r *WebhookRepository) Set(ctx context.Context, config *models.WebhookConfig) error {
	config.UpdatedAt = time.Now()
	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now()
	}

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	key := "webhook:" + config.InstanceID
	return r.redis.Client.Set(ctx, key, data, 0).Err()
}

// Get obtiene la configuración de webhook de una instancia
func (r *WebhookRepository) Get(ctx context.Context, instanceID string) (*models.WebhookConfig, error) {
	key := "webhook:" + instanceID
	data, err := r.redis.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var config models.WebhookConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Delete elimina la configuración de webhook
func (r *WebhookRepository) Delete(ctx context.Context, instanceID string) error {
	key := "webhook:" + instanceID
	return r.redis.Client.Del(ctx, key).Err()
}
