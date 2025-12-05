package services

import (
	"context"
	"fmt"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"kero-kero/internal/models"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
)

type StatusService struct {
	waManager *whatsapp.Manager
}

func NewStatusService(waManager *whatsapp.Manager) *StatusService {
	return &StatusService{waManager: waManager}
}

// PublishTextStatus publica un estado de texto
func (s *StatusService) PublishTextStatus(ctx context.Context, instanceID string, req *models.PublishStatusRequest) (*models.StatusResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Los estados se envían al JID especial de status
	statusJID := types.NewJID("status", types.BroadcastServer)

	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(req.Content),
		},
	}

	resp, err := client.WAClient.SendMessage(ctx, statusJID, msg)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error publicando estado: %v", err))
	}

	return &models.StatusResponse{
		Success:   true,
		MessageID: resp.ID,
	}, nil
}

// GetStatusPrivacy obtiene la configuración de privacidad de estados
func (s *StatusService) GetStatusPrivacy(ctx context.Context, instanceID string) (map[string]interface{}, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Nota: whatsmeow no expone directamente la privacidad de estados
	// Esta es una implementación básica que devuelve info genérica
	return map[string]interface{}{
		"success": true,
		"message": "La privacidad de estados se gestiona desde la app de WhatsApp",
	}, nil
}
