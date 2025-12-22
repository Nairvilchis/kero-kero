package services

import (
	"context"
	"encoding/base64"
	"fmt"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"kero-kero/internal/models"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
)

type NewsletterService struct {
	waManager  *whatsapp.Manager
	msgService *MessageService
}

func NewNewsletterService(waManager *whatsapp.Manager, msgService *MessageService) *NewsletterService {
	return &NewsletterService{
		waManager:  waManager,
		msgService: msgService,
	}
}

// CreateNewsletter crea un nuevo canal de WhatsApp.
func (s *NewsletterService) CreateNewsletter(ctx context.Context, instanceID string, req *models.CreateNewsletterRequest) (*models.NewsletterMetadata, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	var picture []byte
	if req.Picture != "" {
		data, err := base64.StdEncoding.DecodeString(req.Picture)
		if err == nil {
			picture = data
		}
	}

	meta, err := client.WAClient.CreateNewsletter(ctx, whatsmeow.CreateNewsletterParams{
		Name:        req.Name,
		Description: req.Description,
		Picture:     picture,
	})
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error de WhatsApp al crear canal: %v", err))
	}

	return s.mapMetadata(meta), nil
}

// GetSubscribedNewsletters lista los canales a los que la instancia está unida.
func (s *NewsletterService) GetSubscribedNewsletters(ctx context.Context, instanceID string) ([]*models.NewsletterMetadata, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	newsletters, err := client.WAClient.GetSubscribedNewsletters(ctx)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error obteniendo canales suscritos: %v", err))
	}

	var result []*models.NewsletterMetadata
	for _, n := range newsletters {
		result = append(result, s.mapMetadata(n))
	}
	return result, nil
}

// FollowNewsletter se suscribe a un canal mediante su JID.
func (s *NewsletterService) FollowNewsletter(ctx context.Context, instanceID, jidStr string) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}

	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return errors.ErrBadRequest.WithDetails("JID de canal inválido")
	}

	return client.WAClient.FollowNewsletter(ctx, jid)
}

// UnfollowNewsletter cancela la suscripción a un canal.
func (s *NewsletterService) UnfollowNewsletter(ctx context.Context, instanceID, jidStr string) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}

	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return errors.ErrBadRequest.WithDetails("JID de canal inválido")
	}

	return client.WAClient.UnfollowNewsletter(ctx, jid)
}

// SendMessage envía un mensaje (texto o multimedia) al canal.
// Nota: Solo funciona si la instancia es administradora del canal.
func (s *NewsletterService) SendMessage(ctx context.Context, instanceID string, req *models.SendNewsletterMessageRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	jid, err := types.ParseJID(req.JID)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("JID de canal inválido")
	}

	var msg waE2E.Message

	switch req.Type {
	case "image", "video":
		// Para multimedia, descargamos y subimos específicamente al servidor de Canales.
		mediaSource := req.MediaURL
		if req.Payload != "" {
			mediaSource = req.Payload
		}
		if mediaSource == "" {
			return nil, errors.ErrBadRequest.WithDetails("MediaURL o Payload es requerido para este tipo")
		}

		data, mimeType, err := s.msgService.HelperDownloadMediaBytes(mediaSource)
		if err != nil {
			return nil, errors.ErrBadRequest.WithDetails(err.Error())
		}

		mediaType := whatsmeow.MediaImage
		if req.Type == "video" {
			mediaType = whatsmeow.MediaVideo
		}

		// IMPORTANTE: Los canales usan un upload distinto.
		uploaded, err := client.WAClient.UploadNewsletter(ctx, data, mediaType)
		if err != nil {
			return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error subiendo media al canal: %v", err))
		}

		if req.Type == "image" {
			msg.ImageMessage = &waE2E.ImageMessage{
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				Mimetype:      proto.String(mimeType),
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(data))),
				Caption:       proto.String(req.Message),
			}
		} else {
			msg.VideoMessage = &waE2E.VideoMessage{
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				Mimetype:      proto.String(mimeType),
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(data))),
				Caption:       proto.String(req.Message),
			}
		}

	default: // "text"
		if req.Message == "" {
			return nil, errors.ErrBadRequest.WithDetails("Mensaje es requerido")
		}
		msg.Conversation = proto.String(req.Message)
	}

	resp, err := client.WAClient.SendMessage(ctx, jid, &msg)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando mensaje al canal: %v", err))
	}

	return &models.MessageResponse{
		Success:   true,
		MessageID: resp.ID,
	}, nil
}

// GetNewsletterInfo obtiene información detallada de un canal.
func (s *NewsletterService) GetNewsletterInfo(ctx context.Context, instanceID, jidStr string) (*models.NewsletterMetadata, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("JID de canal inválido")
	}

	meta, err := client.WAClient.GetNewsletterInfo(ctx, jid)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error obteniendo info del canal: %v", err))
	}

	return s.mapMetadata(meta), nil
}

// mapMetadata convierte los metadatos de whatsmeow a nuestro modelo interno.
func (s *NewsletterService) mapMetadata(m *types.NewsletterMetadata) *models.NewsletterMetadata {
	inviteLink := ""
	if m.ThreadMeta.InviteCode != "" {
		inviteLink = "https://whatsapp.com/channel/" + m.ThreadMeta.InviteCode
	}

	role := ""
	if m.ViewerMeta != nil {
		role = string(m.ViewerMeta.Role)
	}

	return &models.NewsletterMetadata{
		ID:              m.ID.String(),
		Name:            m.ThreadMeta.Name.Text,
		Description:     m.ThreadMeta.Description.Text,
		InviteCode:      m.ThreadMeta.InviteCode,
		InviteLink:      inviteLink,
		SubscriberCount: m.ThreadMeta.SubscriberCount,
		Role:            role,
	}
}
