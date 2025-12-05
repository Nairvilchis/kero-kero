package services

import (
	"context"
	"fmt"
	"time"

	"go.mau.fi/whatsmeow/types"

	"kero-kero/internal/models"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
)

// PresenceService maneja operaciones de presencia
type PresenceService struct {
	waManager *whatsapp.Manager
}

// NewPresenceService crea una nueva instancia del servicio de presencia
func NewPresenceService(waManager *whatsapp.Manager) *PresenceService {
	return &PresenceService{
		waManager: waManager,
	}
}

// StartPresence activa presencia en un chat (typing o recording)
func (s *PresenceService) StartPresence(ctx context.Context, instanceID string, req *models.StartPresenceRequest) (*models.PresenceResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Construir JID del chat
	jid, err := types.ParseJID(req.Phone)
	if err != nil {
		// Intentar agregar sufijo @s.whatsapp.net si no lo tiene
		jid, err = types.ParseJID(req.Phone + "@s.whatsapp.net")
		if err != nil {
			return nil, errors.ErrBadRequest.WithDetails("Número de teléfono inválido")
		}
	}

	// Determinar el tipo de presencia
	var chatPresence types.ChatPresence
	var media types.ChatPresenceMedia

	switch req.Type {
	case models.PresenceTyping:
		chatPresence = types.ChatPresenceComposing
		media = types.ChatPresenceMediaText
	case models.PresenceRecording:
		chatPresence = types.ChatPresenceComposing
		media = types.ChatPresenceMediaAudio
	default:
		return nil, errors.ErrBadRequest.WithDetails("Tipo de presencia inválido")
	}

	// Enviar presencia
	err = client.WAClient.SendChatPresence(ctx, jid, chatPresence, media)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails("Error enviando presencia: " + err.Error())
	}

	return &models.PresenceResponse{
		Success: true,
		Message: fmt.Sprintf("Presencia '%s' activada en %s", req.Type, req.Phone),
	}, nil
}

// StopPresence detiene cualquier presencia activa en un chat
func (s *PresenceService) StopPresence(ctx context.Context, instanceID string, req *models.StopPresenceRequest) (*models.PresenceResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Construir JID del chat
	jid, err := types.ParseJID(req.Phone)
	if err != nil {
		jid, err = types.ParseJID(req.Phone + "@s.whatsapp.net")
		if err != nil {
			return nil, errors.ErrBadRequest.WithDetails("Número de teléfono inválido")
		}
	}

	// Enviar presencia "paused" para detener
	err = client.WAClient.SendChatPresence(ctx, jid, types.ChatPresencePaused, types.ChatPresenceMediaText)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails("Error deteniendo presencia: " + err.Error())
	}

	return &models.PresenceResponse{
		Success: true,
		Message: fmt.Sprintf("Presencia detenida en %s", req.Phone),
	}, nil
}

// TimedPresence activa presencia temporal que se detiene automáticamente
func (s *PresenceService) TimedPresence(ctx context.Context, instanceID string, req *models.TimedPresenceRequest) (*models.PresenceResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Construir JID del chat
	jid, err := types.ParseJID(req.Phone)
	if err != nil {
		jid, err = types.ParseJID(req.Phone + "@s.whatsapp.net")
		if err != nil {
			return nil, errors.ErrBadRequest.WithDetails("Número de teléfono inválido")
		}
	}

	// Determinar el tipo de presencia
	var chatPresence types.ChatPresence
	var media types.ChatPresenceMedia

	switch req.Type {
	case models.PresenceTyping:
		chatPresence = types.ChatPresenceComposing
		media = types.ChatPresenceMediaText
	case models.PresenceRecording:
		chatPresence = types.ChatPresenceComposing
		media = types.ChatPresenceMediaAudio
	default:
		return nil, errors.ErrBadRequest.WithDetails("Tipo de presencia inválido")
	}

	// Activar presencia
	err = client.WAClient.SendChatPresence(ctx, jid, chatPresence, media)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails("Error enviando presencia: " + err.Error())
	}

	// Programar detención automática usando goroutine
	go func() {
		time.Sleep(time.Duration(req.Duration) * time.Millisecond)
		// Detener presencia
		_ = client.WAClient.SendChatPresence(context.Background(), jid, types.ChatPresencePaused, types.ChatPresenceMediaText)
	}()

	return &models.PresenceResponse{
		Success: true,
		Message: fmt.Sprintf("Presencia '%s' activada por %dms en %s", req.Type, req.Duration, req.Phone),
	}, nil
}

// SetOnlineStatus cambia el estado de presencia general (en línea/desconectado)
func (s *PresenceService) SetOnlineStatus(ctx context.Context, instanceID string, req *models.SetStatusRequest) (*models.PresenceResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Determinar el estado
	var presence types.Presence
	switch req.Status {
	case models.PresenceAvailable:
		presence = types.PresenceAvailable
	case models.PresenceUnavailable:
		presence = types.PresenceUnavailable
	default:
		return nil, errors.ErrBadRequest.WithDetails("Estado de presencia inválido")
	}

	// Enviar presencia global
	err := client.WAClient.SendPresence(ctx, presence)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails("Error cambiando estado: " + err.Error())
	}

	return &models.PresenceResponse{
		Success: true,
		Message: fmt.Sprintf("Estado cambiado a '%s'", req.Status),
	}, nil
}
