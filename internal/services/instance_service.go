package services

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/skip2/go-qrcode"

	"kero-kero/internal/models"
	"kero-kero/internal/repository"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
)

// InstanceService maneja la lógica de negocio de instancias
type InstanceService struct {
	waManager      *whatsapp.Manager
	instanceRepo   *repository.InstanceRepository
	redisClient    *repository.RedisClient
	webhookService *WebhookService
}

// NewInstanceService crea un nuevo servicio de instancias
func NewInstanceService(
	waManager *whatsapp.Manager,
	instanceRepo *repository.InstanceRepository,
	redisClient *repository.RedisClient,
	webhookService *WebhookService,
) *InstanceService {
	return &InstanceService{
		waManager:      waManager,
		instanceRepo:   instanceRepo,
		redisClient:    redisClient,
		webhookService: webhookService,
	}
}

// CreateInstance crea una nueva instancia
func (s *InstanceService) CreateInstance(ctx context.Context, req *models.CreateInstanceRequest) (*models.Instance, error) {
	// Validar que no exista
	exists, err := s.instanceRepo.Exists(ctx, req.InstanceID)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(err.Error())
	}
	if exists {
		return nil, errors.ErrInstanceExists
	}

	// Crear en WhatsApp Manager
	_, err = s.waManager.CreateInstance(ctx, req.InstanceID, req.SyncHistory)
	if err != nil {
		log.Error().Err(err).Str("instance_id", req.InstanceID).Msg("Error creando instancia")
		return nil, errors.ErrInternalServer.WithDetails(err.Error())
	}

	// Obtener instancia creada
	instance, err := s.instanceRepo.GetByID(ctx, req.InstanceID)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(err.Error())
	}

	log.Info().Str("instance_id", req.InstanceID).Msg("Instancia creada exitosamente")
	return instance, nil
}

// GetInstance obtiene una instancia por ID
func (s *InstanceService) GetInstance(ctx context.Context, instanceID string) (*models.Instance, error) {
	instance, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(err.Error())
	}
	if instance == nil {
		return nil, errors.ErrInstanceNotFound
	}

	return instance, nil
}

// GetInstanceWithWebhook obtiene una instancia con su configuración de webhook
func (s *InstanceService) GetInstanceWithWebhook(ctx context.Context, instanceID string) (map[string]interface{}, error) {
	instance, err := s.GetInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"instance_id":       instance.InstanceID,
		"jid":               instance.JID,
		"name":              instance.Name,
		"webhook_url":       instance.WebhookURL,
		"status":            instance.Status,
		"sync_history":      instance.SyncHistory,
		"last_connected_at": instance.LastConnectedAt,
		"created_at":        instance.CreatedAt,
		"updated_at":        instance.UpdatedAt,
	}

	// Intentar cargar configuración de webhook desde Redis
	webhookConfig, err := s.webhookService.GetWebhook(ctx, instanceID)
	if err == nil && webhookConfig != nil {
		result["webhook_config"] = map[string]interface{}{
			"url":     webhookConfig.URL,
			"events":  webhookConfig.Events,
			"secret":  webhookConfig.Secret,
			"enabled": webhookConfig.Enabled,
		}
	}

	return result, nil
}

// ListInstances lista todas las instancias
func (s *InstanceService) ListInstances(ctx context.Context) ([]*models.Instance, error) {
	instances, err := s.instanceRepo.GetAll(ctx)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(err.Error())
	}

	return instances, nil
}

// DeleteInstance elimina una instancia
func (s *InstanceService) DeleteInstance(ctx context.Context, instanceID string) error {
	// Verificar que existe
	exists, err := s.instanceRepo.Exists(ctx, instanceID)
	if err != nil {
		return errors.ErrInternalServer.WithDetails(err.Error())
	}
	if !exists {
		return errors.ErrInstanceNotFound
	}

	// Eliminar del manager
	if err := s.waManager.DeleteInstance(ctx, instanceID); err != nil {
		log.Error().Err(err).Str("instance_id", instanceID).Msg("Error eliminando instancia")
		return errors.ErrInternalServer.WithDetails(err.Error())
	}

	log.Info().Str("instance_id", instanceID).Msg("Instancia eliminada exitosamente")
	return nil
}

// ConnectInstance inicia la conexión de una instancia
func (s *InstanceService) ConnectInstance(ctx context.Context, instanceID string) error {
	// Verificar que la instancia existe en BD
	exists, err := s.instanceRepo.Exists(ctx, instanceID)
	if err != nil {
		return errors.ErrInternalServer.WithDetails(err.Error())
	}
	if !exists {
		return errors.ErrInstanceNotFound
	}

	// Obtener o crear el cliente en el Manager (sin tocar la BD)
	client, err := s.waManager.GetOrCreateClient(ctx, instanceID)
	if err != nil {
		log.Error().Err(err).Str("instance_id", instanceID).Msg("Error obteniendo/creando cliente")
		return errors.ErrInternalServer.WithDetails(err.Error())
	}

	// Verificar si ya está conectado
	if client.WAClient.IsConnected() {
		return nil // Ya está conectado
	}

	// Actualizar estado
	if err := s.instanceRepo.UpdateStatus(ctx, instanceID, models.StatusConnecting); err != nil {
		log.Error().Err(err).Msg("Error actualizando estado")
	}

	// Conectar en goroutine
	go func() {
		if err := client.Connect(); err != nil {
			log.Error().Err(err).Str("instance_id", instanceID).Msg("Error conectando instancia")
			s.instanceRepo.UpdateStatus(context.Background(), instanceID, models.StatusFailed)
		}
	}()

	log.Info().Str("instance_id", instanceID).Msg("Conexión iniciada")
	return nil
}

// DisconnectInstance desconecta una instancia
func (s *InstanceService) DisconnectInstance(ctx context.Context, instanceID string) error {
	// Verificar que la instancia existe en BD
	exists, err := s.instanceRepo.Exists(ctx, instanceID)
	if err != nil {
		return errors.ErrInternalServer.WithDetails(err.Error())
	}
	if !exists {
		return errors.ErrInstanceNotFound
	}

	// Intentar obtener el cliente del Manager
	client := s.waManager.GetClient(instanceID)
	if client != nil {
		// Si existe en el Manager, desconectarlo
		client.Disconnect()
	}

	// Actualizar estado en BD (siempre, incluso si no estaba en el Manager)
	if err := s.instanceRepo.UpdateStatus(ctx, instanceID, models.StatusDisconnected); err != nil {
		log.Error().Err(err).Msg("Error actualizando estado")
		return errors.ErrInternalServer.WithDetails(err.Error())
	}

	log.Info().Str("instance_id", instanceID).Msg("Instancia desconectada")
	return nil
}

// GetQRCode obtiene el código QR de una instancia
func (s *InstanceService) GetQRCode(ctx context.Context, instanceID string) ([]byte, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	if client.WAClient.IsLoggedIn() {
		return nil, errors.New(400, "La instancia ya está autenticada")
	}

	// Intentar obtener de Redis con reintentos (el QR puede tardar en guardarse)
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		qrCode, err := s.redisClient.GetQRCode(ctx, instanceID)
		if err == nil && qrCode != "" {
			log.Debug().Str("instance_id", instanceID).Int("attempt", i+1).Msg("QR obtenido de Redis")
			// Generar imagen PNG
			png, err := qrcode.Encode(qrCode, qrcode.Medium, 256)
			if err != nil {
				return nil, errors.ErrInternalServer.WithDetails("Error generando imagen QR")
			}
			return png, nil
		}

		// Si no es el último intento, esperar un poco
		if i < maxAttempts-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	// Si no hay en Redis después de reintentos, esperar uno nuevo con timeout
	log.Debug().Str("instance_id", instanceID).Msg("QR no encontrado en Redis, esperando nuevo QR")
	qrChan, _ := client.WAClient.GetQRChannel(ctx)

	select {
	case evt := <-qrChan:
		if evt.Event == "code" {
			png, err := qrcode.Encode(evt.Code, qrcode.Medium, 256)
			if err != nil {
				return nil, errors.ErrInternalServer.WithDetails("Error generando imagen QR")
			}
			return png, nil
		}
		return nil, errors.New(503, "QR no disponible: "+evt.Event)

	case <-time.After(25 * time.Second): // Reducido a 25s para dar tiempo a los reintentos
		return nil, errors.New(408, "Timeout esperando código QR")

	case <-ctx.Done():
		return nil, errors.New(499, "Request cancelado")
	}
}

// GetStatus obtiene el estado de una instancia
func (s *InstanceService) GetStatus(ctx context.Context, instanceID string) (string, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return "", errors.ErrInstanceNotFound
	}

	status := string(models.StatusDisconnected)
	if client.WAClient.IsConnected() {
		status = string(models.StatusConnected)
	}
	if client.WAClient.IsLoggedIn() {
		status = string(models.StatusAuthenticated)
	}

	// Actualizar en base de datos si es diferente
	instance, _ := s.instanceRepo.GetByID(ctx, instanceID)
	if instance != nil && instance.Status != status {
		s.instanceRepo.UpdateStatus(ctx, instanceID, models.InstanceStatus(status))
	}

	return status, nil
}

// UpdateInstance actualiza la configuración de una instancia
func (s *InstanceService) UpdateInstance(ctx context.Context, instanceID string, req *models.UpdateInstanceRequest) error {
	// Verificar que existe
	exists, err := s.instanceRepo.Exists(ctx, instanceID)
	if err != nil {
		return errors.ErrInternalServer.WithDetails(err.Error())
	}
	if !exists {
		return errors.ErrInstanceNotFound
	}

	// Actualizar en BD
	if err := s.instanceRepo.Update(ctx, instanceID, req); err != nil {
		log.Error().Err(err).Str("instance_id", instanceID).Msg("Error actualizando instancia")
		return errors.ErrInternalServer.WithDetails(err.Error())
	}

	// Si hay cambios en webhook, actualizar en manager si es necesario
	if req.WebhookURL != nil && *req.WebhookURL != "" {
		webhookConfig := &models.WebhookConfig{
			InstanceID: instanceID,
			URL:        *req.WebhookURL,
			Enabled:    true,
			Events:     []string{"message", "status", "receipt"}, // Default events
		}
		if err := s.webhookService.SetWebhook(ctx, webhookConfig); err != nil {
			log.Error().Err(err).Str("instance_id", instanceID).Msg("Error actualizando webhook en Redis")
			// No fallamos la request completa, pero logueamos el error
		} else {
			log.Info().Str("instance_id", instanceID).Str("url", *req.WebhookURL).Msg("Webhook actualizado en Redis")
		}
	}

	log.Info().Str("instance_id", instanceID).Msg("Instancia actualizada exitosamente")
	return nil
}
