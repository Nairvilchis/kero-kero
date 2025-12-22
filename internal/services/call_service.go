package services

import (
	"context"
	"encoding/json"
	"fmt"

	"kero-kero/internal/models"
	"kero-kero/internal/repository"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
)

type CallService struct {
	waManager   *whatsapp.Manager
	redisClient *repository.RedisClient
}

func NewCallService(waManager *whatsapp.Manager, redisClient *repository.RedisClient) *CallService {
	return &CallService{
		waManager:   waManager,
		redisClient: redisClient,
	}
}

// RejectCall rechaza una llamada entrante (manualmente vía API)
func (s *CallService) RejectCall(ctx context.Context, instanceID, callID, from string) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	// whatsmeow requiere que estemos escuchando el evento CallOffer para rechazarla.
	// El rechazo manual vía ID es complejo si no tenemos el estado de la llamada.
	// Por ahora devolvemos 501.
	return errors.New(501, "Rechazo manual de llamadas por ID no implementado directamente en whatsmeow. Use auto_reject.")
}

// GetCallSettings obtiene la configuración de llamadas desde Redis
func (s *CallService) GetCallSettings(ctx context.Context, instanceID string) (*models.CallSettings, error) {
	data, err := s.redisClient.GetCallSettings(ctx, instanceID)
	if err != nil {
		// Si no hay configuración, devolvemos los valores por defecto
		return &models.CallSettings{
			AutoReject:       false,
			AutoReplyEnabled: false,
		}, nil
	}

	var settings models.CallSettings
	if err := json.Unmarshal([]byte(data), &settings); err != nil {
		return nil, errors.ErrInternalServer.WithDetails("Error parseando configuración de llamadas")
	}

	return &settings, nil
}

// SetCallSettings actualiza la configuración de llamadas en Redis
func (s *CallService) SetCallSettings(ctx context.Context, instanceID string, settings *models.CallSettings) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}

	if err := s.redisClient.SetCallSettings(ctx, instanceID, settings); err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error guardando settings en Redis: %v", err))
	}

	return nil
}
