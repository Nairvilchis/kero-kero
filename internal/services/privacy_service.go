package services

import (
	"context"
	"fmt"
	"strconv"

	"go.mau.fi/whatsmeow/types"

	"kero-kero/internal/models"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
)

type PrivacyService struct {
	waManager *whatsapp.Manager
}

func NewPrivacyService(waManager *whatsapp.Manager) *PrivacyService {
	return &PrivacyService{waManager: waManager}
}

// GetPrivacySettings obtiene la configuración actual de privacidad
func (s *PrivacyService) GetPrivacySettings(ctx context.Context, instanceID string) (*models.PrivacySettings, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	settings, err := client.WAClient.TryFetchPrivacySettings(ctx, false)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error obteniendo privacidad: %v", err))
	}

	// Convertir ReadReceipts de PrivacySetting a bool
	readReceipts := settings.ReadReceipts == types.PrivacySettingAll

	return &models.PrivacySettings{
		LastSeen:     string(settings.LastSeen),
		ProfilePhoto: string(settings.Profile),
		Status:       string(settings.Status),
		ReadReceipts: &readReceipts,
		Groups:       string(settings.GroupAdd),
		DefaultTimer: "0", // Campo no disponible en la versión actual de whatsmeow
	}, nil
}

// UpdatePrivacySettings actualiza una configuración de privacidad
func (s *PrivacyService) UpdatePrivacySettings(ctx context.Context, instanceID string, req *models.PrivacyUpdateRequest) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	var err error
	switch req.Category {
	case "last_seen":
		_, err = client.WAClient.SetPrivacySetting(ctx, types.PrivacySettingTypeLastSeen, types.PrivacySetting(req.Value))
	case "profile_photo":
		_, err = client.WAClient.SetPrivacySetting(ctx, types.PrivacySettingTypeProfile, types.PrivacySetting(req.Value))
	case "status":
		_, err = client.WAClient.SetPrivacySetting(ctx, types.PrivacySettingTypeStatus, types.PrivacySetting(req.Value))
	case "read_receipts":
		val, _ := strconv.ParseBool(req.Value)
		if val {
			_, err = client.WAClient.SetPrivacySetting(ctx, types.PrivacySettingTypeReadReceipts, types.PrivacySettingAll)
		} else {
			_, err = client.WAClient.SetPrivacySetting(ctx, types.PrivacySettingTypeReadReceipts, types.PrivacySettingNone)
		}
	case "groups":
		_, err = client.WAClient.SetPrivacySetting(ctx, types.PrivacySettingTypeGroupAdd, types.PrivacySetting(req.Value))
	case "default_timer":
		return errors.New(501, "Actualización de timer no implementada")
	default:
		return errors.ErrBadRequest.WithDetails("Categoría inválida")
	}

	if err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error actualizando privacidad: %v", err))
	}

	return nil
}
