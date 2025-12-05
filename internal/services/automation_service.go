package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"kero-kero/internal/models"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
)

type AutomationService struct {
	waManager *whatsapp.Manager
	redis     *redis.Client
}

func NewAutomationService(waManager *whatsapp.Manager, redis *redis.Client) *AutomationService {
	return &AutomationService{
		waManager: waManager,
		redis:     redis,
	}
}

// SendBulkMessage inicia un proceso de envío masivo
func (s *AutomationService) SendBulkMessage(ctx context.Context, instanceID string, req *models.BulkMessageRequest) (*models.BulkMessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Validar delays
	if req.MinDelay <= 0 {
		req.MinDelay = 2000 // 2 segundos por defecto
	}
	if req.MaxDelay <= 0 {
		req.MaxDelay = 5000 // 5 segundos por defecto
	}
	if req.MaxDelay < req.MinDelay {
		req.MaxDelay = req.MinDelay
	}

	jobID := uuid.New().String()

	// Ejecutar en goroutine (fire and forget)
	go func() {
		// Aquí podríamos guardar estado del job en Redis si quisiéramos tracking detallado
		for _, phone := range req.Phones {
			// Simular delay humano
			delay := rand.Intn(req.MaxDelay-req.MinDelay+1) + req.MinDelay
			time.Sleep(time.Duration(delay) * time.Millisecond)

			jid := types.NewJID(phone, types.DefaultUserServer)

			// Enviar mensaje
			// Nota: Si hay media_url, la lógica sería más compleja (descargar, subir, etc.)
			// Por simplicidad en esta fase, asumimos texto simple.
			// Si se requiere media, se puede reutilizar MessageService.

			_, err := client.WAClient.SendMessage(context.Background(), jid, &waProto.Message{
				Conversation: proto.String(req.Message),
			})

			if err != nil {
				fmt.Printf("Error enviando bulk a %s: %v\n", phone, err)
				// Continuar con el siguiente
			}
		}
	}()

	return &models.BulkMessageResponse{
		Success:         true,
		JobID:           jobID,
		TotalRecipients: len(req.Phones),
		Status:          "processing",
	}, nil
}

// ScheduleMessage programa un mensaje para envío futuro
func (s *AutomationService) ScheduleMessage(ctx context.Context, instanceID string, req *models.ScheduleMessageRequest) error {
	// Validar fecha futura
	if req.ExecuteAt <= time.Now().Unix() {
		return errors.ErrBadRequest.WithDetails("La fecha de ejecución debe ser futura")
	}

	msg := models.ScheduledMessage{
		ID:        uuid.New().String(),
		Phone:     req.Phone,
		Message:   req.Message,
		ExecuteAt: req.ExecuteAt,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return errors.ErrInternalServer.WithDetails("Error codificando mensaje")
	}

	// Guardar en Redis ZSET: scheduled_messages:{instanceID}
	// Score = ExecuteAt
	key := fmt.Sprintf("scheduled_messages:%s", instanceID)
	err = s.redis.ZAdd(ctx, key, redis.Z{
		Score:  float64(req.ExecuteAt),
		Member: data,
	}).Err()

	if err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error programando mensaje: %v", err))
	}

	return nil
}

// SetAutoReply configura la respuesta automática
func (s *AutomationService) SetAutoReply(ctx context.Context, instanceID string, config *models.AutoReplyConfig) error {
	key := fmt.Sprintf("autoreply:%s", instanceID)

	data, err := json.Marshal(config)
	if err != nil {
		return errors.ErrInternalServer.WithDetails("Error codificando configuración")
	}

	err = s.redis.Set(ctx, key, data, 0).Err()
	if err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error guardando auto-reply: %v", err))
	}

	return nil
}

// GetAutoReply obtiene la configuración de respuesta automática
func (s *AutomationService) GetAutoReply(ctx context.Context, instanceID string) (*models.AutoReplyConfig, error) {
	key := fmt.Sprintf("autoreply:%s", instanceID)

	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		// No config found, return default disabled
		return &models.AutoReplyConfig{Enabled: false}, nil
	}
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error obteniendo auto-reply: %v", err))
	}

	var config models.AutoReplyConfig
	if err := json.Unmarshal([]byte(val), &config); err != nil {
		return nil, errors.ErrInternalServer.WithDetails("Error decodificando configuración")
	}

	return &config, nil
}

// StartScheduler inicia el worker que procesa mensajes programados
// Debe llamarse una vez al inicio de la aplicación (en main.go)
func (s *AutomationService) StartScheduler() {
	ticker := time.NewTicker(30 * time.Second) // Revisar cada 30 segundos
	go func() {
		for range ticker.C {
			s.processScheduledMessages()
		}
	}()
}

func (s *AutomationService) processScheduledMessages() {
	ctx := context.Background()
	now := time.Now().Unix()

	// Obtener todas las claves de scheduled_messages
	// Esto es ineficiente si hay muchas instancias.
	// Idealmente tendríamos un set global de instancias activas o iteraríamos sobre las conocidas.
	// Por simplicidad, usaremos SCAN o asumiremos que el manager tiene las instancias cargadas.

	// Mejor enfoque: Iterar sobre las instancias cargadas en el manager
	// Como no tengo GetLoadedInstances expuesto aún, voy a usar SCAN en Redis para encontrar keys
	iter := s.redis.Scan(ctx, 0, "scheduled_messages:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		// Extraer instanceID de la key "scheduled_messages:{instanceID}"
		var instanceID string
		fmt.Sscanf(key, "scheduled_messages:%s", &instanceID)

		// Obtener mensajes vencidos (Score <= Now)
		vals, err := s.redis.ZRangeByScore(ctx, key, &redis.ZRangeBy{
			Min: "-inf",
			Max: fmt.Sprintf("%d", now),
		}).Result()

		if err != nil {
			continue
		}

		if len(vals) == 0 {
			continue
		}

		client := s.waManager.GetClient(instanceID)
		if client == nil || !client.WAClient.IsLoggedIn() {
			// Si la instancia no está lista, no podemos enviar.
			// Opción: Dejar en cola para después o borrar.
			// Por ahora, dejaremos en cola (no borramos).
			continue
		}

		for _, val := range vals {
			var msg models.ScheduledMessage
			if err := json.Unmarshal([]byte(val), &msg); err != nil {
				// Mensaje corrupto, borrar
				s.redis.ZRem(ctx, key, val)
				continue
			}

			// Enviar mensaje
			jid := types.NewJID(msg.Phone, types.DefaultUserServer)
			_, err := client.WAClient.SendMessage(ctx, jid, &waProto.Message{
				Conversation: proto.String(msg.Message),
			})

			if err != nil {
				fmt.Printf("Error enviando mensaje programado %s: %v\n", msg.ID, err)
				// Podríamos reintentar o marcar error. Por ahora borramos para no bloquear.
			} else {
				fmt.Printf("Mensaje programado %s enviado exitosamente\n", msg.ID)
			}

			// Borrar de Redis
			s.redis.ZRem(ctx, key, val)
		}
	}
}
