package services

import (
	"context"

	"kero-kero/internal/models"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
)

type CallService struct {
	waManager *whatsapp.Manager
}

func NewCallService(waManager *whatsapp.Manager) *CallService {
	return &CallService{waManager: waManager}
}

// RejectCall rechaza una llamada entrante
func (s *CallService) RejectCall(ctx context.Context, instanceID, callID, from string) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	// Nota: El rechazo de llamadas en whatsmeow requiere manejar eventos de llamada
	// y responder apropiadamente. La API de WhatsApp no expone un método directo
	// para rechazar llamadas vía mensaje.
	// Esta funcionalidad se implementa mejor en el event handler del Manager.

	return errors.New(501, "Rechazo manual de llamadas no implementado. Use auto_reject en settings")
}

// GetCallSettings obtiene la configuración de llamadas
func (s *CallService) GetCallSettings(ctx context.Context, instanceID string) (*models.CallSettings, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Nota: La configuración de auto-rechazo se gestiona en el Manager
	// Esta es una implementación básica
	return &models.CallSettings{
		AutoReject: false, // Por defecto desactivado
	}, nil
}

// SetCallSettings actualiza la configuración de llamadas
func (s *CallService) SetCallSettings(ctx context.Context, instanceID string, settings *models.CallSettings) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	// Nota: La lógica de auto-rechazo debería implementarse en el event handler del Manager
	// Aquí solo validamos que el cliente existe
	return nil
}
