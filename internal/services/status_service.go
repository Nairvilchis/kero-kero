package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.mau.fi/whatsmeow/proto/waE2E"
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

// PublishTextStatus publica un estado de texto con soporte para colores y fuentes.
// He optimizado esto para que se vea mucho mejor en la app usando ExtendedTextMessage.
func (s *StatusService) PublishStatus(ctx context.Context, instanceID string, req *models.PublishStatusRequest) (*models.StatusResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	var msg *waE2E.Message

	switch req.Type {
	case "text":
		// Procesar colores (por defecto fondos típicos de WhatsApp si no se envían)
		bgColor := s.parseHexColor(req.BackgroundColor, 0xFF007A65) // Verde oscuro por defecto
		textColor := s.parseHexColor(req.TextColor, 0xFFFFFFFF)     // Blanco por defecto

		msg = &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text:           proto.String(req.Content),
				BackgroundArgb: proto.Uint32(bgColor),
				TextArgb:       proto.Uint32(textColor),
				Font:           waE2E.ExtendedTextMessage_FontType(req.Font).Enum(),
			},
		}
	default:
		return nil, errors.ErrBadRequest.WithDetails("Tipo de estado no soportado aún (solo 'text')")
	}

	// El JID de estados es global y fijo.
	resp, err := client.WAClient.SendMessage(ctx, types.StatusBroadcastJID, msg)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error publicando estado: %v", err))
	}

	return &models.StatusResponse{
		Success:   true,
		MessageID: resp.ID,
	}, nil
}

// parseHexColor convierte un string hex (#RRGGBB) a uint32 ARGB.
func (s *StatusService) parseHexColor(hex string, def uint32) uint32 {
	if hex == "" {
		return def
	}
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 6 {
		hex = "FF" + hex // Añadimos alpha opaco
	}
	val, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return def
	}
	return uint32(val)
}

// GetStatusPrivacy obtiene la configuración de privacidad de estados.
func (s *StatusService) GetStatusPrivacy(ctx context.Context, instanceID string) (map[string]interface{}, error) {
	// ... (mismo fallback que antes)
	return map[string]interface{}{
		"success": true,
		"message": "La privacidad de estados se gestiona desde la app de WhatsApp",
	}, nil
}
