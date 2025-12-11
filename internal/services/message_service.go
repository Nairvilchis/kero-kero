package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"kero-kero/internal/models"
	"kero-kero/internal/repository"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
)

type MessageService struct {
	waManager *whatsapp.Manager
	msgRepo   *repository.MessageRepository
}

func NewMessageService(waManager *whatsapp.Manager, msgRepo *repository.MessageRepository) *MessageService {
	return &MessageService{
		waManager: waManager,
		msgRepo:   msgRepo,
	}
}

func (s *MessageService) SendText(ctx context.Context, instanceID string, req *models.SendTextRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}
	recipientJID, err := types.ParseJID(req.Phone)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("Número de teléfono inválido")
	}
	msg := &waProto.Message{
		Conversation: proto.String(req.Message),
	}
	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando mensaje: %v", err))
	}
	return &models.MessageResponse{Success: true, MessageID: resp.ID, Status: "sent"}, nil
}
// ... (El resto de los métodos de envío síncrono se mantienen igual)
