package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"kero-kero/internal/config"
)

// RedisClient representa el cliente de Redis
type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient crea una nueva conexión a Redis
func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	log.Info().
		Str("host", cfg.Redis.Host).
		Int("port", cfg.Redis.Port).
		Int("db", cfg.Redis.DB).
		Msg("Conectando a Redis")

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.GetRedisAddr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("error conectando a Redis: %w", err)
	}

	log.Info().Msg("Redis conectado exitosamente")
	return &RedisClient{Client: client}, nil
}

// Close cierra la conexión a Redis
func (r *RedisClient) Close() error {
	log.Info().Msg("Cerrando conexión a Redis")
	return r.Client.Close()
}

// Health verifica el estado de Redis
func (r *RedisClient) Health(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

// --- Métodos de Caché ---

// Set almacena un valor en caché
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Get obtiene un valor del caché
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Delete elimina una clave del caché
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists verifica si una clave existe
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.Client.Exists(ctx, key).Result()
	return result > 0, err
}

// --- Métodos para QR Codes ---

// SetQRCode almacena un código QR para una instancia
func (r *RedisClient) SetQRCode(ctx context.Context, instanceID, qrCode string) error {
	key := fmt.Sprintf("qr:%s", instanceID)
	return r.Set(ctx, key, qrCode, 2*time.Minute) // QR expira en 2 minutos
}

// GetQRCode obtiene el código QR de una instancia
func (r *RedisClient) GetQRCode(ctx context.Context, instanceID string) (string, error) {
	key := fmt.Sprintf("qr:%s", instanceID)
	return r.Get(ctx, key)
}

// DeleteQRCode elimina el código QR de una instancia
func (r *RedisClient) DeleteQRCode(ctx context.Context, instanceID string) error {
	key := fmt.Sprintf("qr:%s", instanceID)
	return r.Delete(ctx, key)
}

// --- Métodos para Sesiones ---

// SetSession almacena información de sesión
func (r *RedisClient) SetSession(ctx context.Context, instanceID string, data interface{}) error {
	key := fmt.Sprintf("session:%s", instanceID)
	return r.Set(ctx, key, data, 24*time.Hour) // Sesión expira en 24 horas
}

// GetSession obtiene información de sesión
func (r *RedisClient) GetSession(ctx context.Context, instanceID string) (string, error) {
	key := fmt.Sprintf("session:%s", instanceID)
	return r.Get(ctx, key)
}

// DeleteSession elimina una sesión
func (r *RedisClient) DeleteSession(ctx context.Context, instanceID string) error {
	key := fmt.Sprintf("session:%s", instanceID)
	return r.Delete(ctx, key)
}

// --- Métodos para Llamadas ---

// SetCallSettings almacena la configuración de llamadas
func (r *RedisClient) SetCallSettings(ctx context.Context, instanceID string, settings interface{}) error {
	key := fmt.Sprintf("call_settings:%s", instanceID)
	return r.Set(ctx, key, settings, 0) // No expira
}

// GetCallSettings obtiene la configuración de llamadas
func (r *RedisClient) GetCallSettings(ctx context.Context, instanceID string) (string, error) {
	key := fmt.Sprintf("call_settings:%s", instanceID)
	return r.Get(ctx, key)
}

// --- Métodos para Cola de Mensajes ---

// EnqueueMessage añade un mensaje a la cola
func (r *RedisClient) EnqueueMessage(ctx context.Context, queueName string, message interface{}) error {
	return r.Client.RPush(ctx, queueName, message).Err()
}

// DequeueMessage obtiene un mensaje de la cola (bloqueante)
func (r *RedisClient) DequeueMessage(ctx context.Context, queueName string, timeout time.Duration) (string, error) {
	result, err := r.Client.BLPop(ctx, timeout, queueName).Result()
	if err != nil {
		return "", err
	}
	if len(result) < 2 {
		return "", fmt.Errorf("respuesta inválida de cola")
	}
	return result[1], nil
}

// DequeueMessageReliable obtiene un mensaje de la cola principal y lo mueve a una cola de procesamiento de forma atómica
func (r *RedisClient) DequeueMessageReliable(ctx context.Context, queueName, processingQueue string, timeout time.Duration) (string, error) {
	// BRPopLPush mueve un elemento de la cola de origen a la cola de destino de forma atómica
	result, err := r.Client.BRPopLPush(ctx, queueName, processingQueue, timeout).Result()
	if err != nil {
		return "", err
	}
	return result, nil
}

// AckMessage elimina un mensaje de la cola de procesamiento una vez que ha sido procesado con éxito
func (r *RedisClient) AckMessage(ctx context.Context, processingQueue, message string) error {
	return r.Client.LRem(ctx, processingQueue, 1, message).Err()
}

// GetQueueLength obtiene la longitud de una cola
func (r *RedisClient) GetQueueLength(ctx context.Context, queueName string) (int64, error) {
	return r.Client.LLen(ctx, queueName).Result()
}

// --- Métodos para Status ---

// SetInstanceStatus almacena el estado de conexión de una instancia
func (r *RedisClient) SetInstanceStatus(ctx context.Context, instanceID, status string) error {
	key := fmt.Sprintf("instance:%s:status", instanceID)
	// Sin expiración, ya que representa el estado actual.
	// Se debe actualizar explícitamente a "disconnected" cuando sea necesario.
	return r.Set(ctx, key, status, 0)
}

// GetInstanceStatus obtiene el estado de conexión de una instancia
func (r *RedisClient) GetInstanceStatus(ctx context.Context, instanceID string) (string, error) {
	key := fmt.Sprintf("instance:%s:status", instanceID)
	val, err := r.Get(ctx, key)
	if err == redis.Nil {
		return "disconnected", nil // Default si no hay clave
	}
	return val, err
}

// CheckRateLimit verifica si una instancia ha superado el límite de mensajes en una ventana de tiempo.
// He usado un script Lua para garantizar que el incremento y la expiración sean atómicos.
func (r *RedisClient) CheckRateLimit(ctx context.Context, instanceID string, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("ratelimit:%s", instanceID)

	script := `
        local current = redis.call("INCR", KEYS[1])
        if current == 1 then
            redis.call("EXPIRE", KEYS[1], ARGV[1])
        end
        return current
    `

	val, err := r.Client.Eval(ctx, script, []string{key}, int(window.Seconds())).Int64()
	if err != nil {
		return false, err
	}

	return val <= int64(limit), nil
}

// SetAutoLabelRules almacena las reglas de etiquetado automático
func (r *RedisClient) SetAutoLabelRules(ctx context.Context, instanceID string, rulesJSON string) error {
	key := fmt.Sprintf("autolabel_rules:%s", instanceID)
	return r.Set(ctx, key, rulesJSON, 0)
}

// GetAutoLabelRules obtiene las reglas de etiquetado automático
func (r *RedisClient) GetAutoLabelRules(ctx context.Context, instanceID string) (string, error) {
	key := fmt.Sprintf("autolabel_rules:%s", instanceID)
	return r.Get(ctx, key)
}
