package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"kero-kero/internal/config"
	"kero-kero/internal/models"
	"kero-kero/internal/repository"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
)

type MessageService struct {
	waManager   *whatsapp.Manager
	msgRepo     *repository.MessageRepository
	redisClient *repository.RedisClient
}

func NewMessageService(waManager *whatsapp.Manager, msgRepo *repository.MessageRepository, redisClient *repository.RedisClient) *MessageService {
	return &MessageService{
		waManager:   waManager,
		msgRepo:     msgRepo,
		redisClient: redisClient,
	}
}

// SendText envía un mensaje de texto de forma síncrona o asíncrona.
func (s *MessageService) SendText(ctx context.Context, instanceID string, req *models.SendTextRequest, isAsync bool) (interface{}, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}
	if isAsync {
		return s.enqueueMessage(ctx, instanceID, models.MessageJobTypeText, req)
	}
	return s.sendSyncTextMessage(ctx, client, req)
}

func (s *MessageService) sendSyncTextMessage(ctx context.Context, client *whatsapp.Client, req *models.SendTextRequest) (*models.MessageResponse, error) {
	recipientJID, err := types.ParseJID(req.Phone)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("Número de teléfono inválido")
	}
	msg := &waProto.Message{
		Conversation: proto.String(req.Message),
	}
	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		log.Error().Err(err).Str("recipient", req.Phone).Msg("Error enviando mensaje síncrono")
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando mensaje: %v", err))
	}

	dbMsg := &models.Message{
		ID:         resp.ID,
		InstanceID: client.WAClient.Store.ID.User,
		From:       client.WAClient.Store.ID.User,
		To:         recipientJID.User,
		Type:       "text",
		Content:    req.Message,
		Timestamp:  resp.Timestamp.Unix(),
		IsFromMe:   true,
		Status:     "sent",
	}
	if err := s.msgRepo.Create(ctx, dbMsg); err != nil {
		log.Error().Err(err).Msg("Error guardando mensaje síncrono enviado en DB")
	}

	return &models.MessageResponse{Success: true, MessageID: resp.ID, Status: string(models.MessageStatusSent)}, nil
}

// SendImage envía una imagen de forma síncrona o asíncrona.
func (s *MessageService) SendImage(ctx context.Context, instanceID string, req *models.SendMediaRequest, isAsync bool) (interface{}, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}
	if isAsync {
		return s.enqueueMessage(ctx, instanceID, models.MessageJobTypeImage, req)
	}
	return s.sendSyncImageMessage(ctx, client, req)
}

func (s *MessageService) sendSyncImageMessage(ctx context.Context, client *whatsapp.Client, req *models.SendMediaRequest) (*models.MessageResponse, error) {
	data, mimeType, err := s.helperDownloadMediaBytes(req.MediaURL)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails(err.Error())
	}
	uploaded, err := client.WAClient.Upload(ctx, data, whatsmeow.MediaImage)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error subiendo imagen: %v", err))
	}
	recipientJID, err := types.ParseJID(req.Phone)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("Número de teléfono inválido")
	}
	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			Caption:       proto.String(req.Caption),
		},
	}
	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando imagen: %v", err))
	}
	return &models.MessageResponse{Success: true, MessageID: resp.ID, Status: "sent"}, nil
}

// SendVideo envía un video de forma síncrona o asíncrona.
func (s *MessageService) SendVideo(ctx context.Context, instanceID string, req *models.SendMediaRequest, isAsync bool) (interface{}, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}
	if isAsync {
		return s.enqueueMessage(ctx, instanceID, models.MessageJobTypeVideo, req)
	}
	return s.sendSyncVideoMessage(ctx, client, req)
}

func (s *MessageService) sendSyncVideoMessage(ctx context.Context, client *whatsapp.Client, req *models.SendMediaRequest) (*models.MessageResponse, error) {
	data, mimeType, err := s.helperDownloadMediaBytes(req.MediaURL)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails(err.Error())
	}
	uploaded, err := client.WAClient.Upload(ctx, data, whatsmeow.MediaVideo)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error subiendo video: %v", err))
	}
	recipientJID, err := types.ParseJID(req.Phone)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("Número de teléfono inválido")
	}
	msg := &waProto.Message{
		VideoMessage: &waProto.VideoMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			Caption:       proto.String(req.Caption),
		},
	}
	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando video: %v", err))
	}
	return &models.MessageResponse{Success: true, MessageID: resp.ID, Status: "sent"}, nil
}

// SendAudio envía un audio de forma síncrona o asíncrona.
func (s *MessageService) SendAudio(ctx context.Context, instanceID string, req *models.SendMediaRequest, isAsync bool) (interface{}, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}
	if isAsync {
		return s.enqueueMessage(ctx, instanceID, models.MessageJobTypeAudio, req)
	}
	return s.sendSyncAudioMessage(ctx, client, req)
}

func (s *MessageService) sendSyncAudioMessage(ctx context.Context, client *whatsapp.Client, req *models.SendMediaRequest) (*models.MessageResponse, error) {
	data, mimeType, err := s.helperDownloadMediaBytes(req.MediaURL)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails(err.Error())
	}
	uploaded, err := client.WAClient.Upload(ctx, data, whatsmeow.MediaAudio)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error subiendo audio: %v", err))
	}
	recipientJID, err := types.ParseJID(req.Phone)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("Número de teléfono inválido")
	}
	msg := &waProto.Message{
		AudioMessage: &waProto.AudioMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			PTT:           proto.Bool(true),
		},
	}
	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando audio: %v", err))
	}
	return &models.MessageResponse{Success: true, MessageID: resp.ID, Status: "sent"}, nil
}

// SendDocument envía un documento de forma síncrona o asíncrona.
func (s *MessageService) SendDocument(ctx context.Context, instanceID string, req *models.SendMediaRequest, isAsync bool) (interface{}, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}
	if isAsync {
		return s.enqueueMessage(ctx, instanceID, models.MessageJobTypeDocument, req)
	}
	return s.sendSyncDocumentMessage(ctx, client, req)
}

func (s *MessageService) sendSyncDocumentMessage(ctx context.Context, client *whatsapp.Client, req *models.SendMediaRequest) (*models.MessageResponse, error) {
	data, mimeType, err := s.helperDownloadMediaBytes(req.MediaURL)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails(err.Error())
	}
	uploaded, err := client.WAClient.Upload(ctx, data, whatsmeow.MediaDocument)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error subiendo documento: %v", err))
	}
	recipientJID, err := types.ParseJID(req.Phone)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("Número de teléfono inválido")
	}
	msg := &waProto.Message{
		DocumentMessage: &waProto.DocumentMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			Caption:       proto.String(req.Caption),
			FileName:      proto.String(req.FileName),
		},
	}
	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando documento: %v", err))
	}
	return &models.MessageResponse{Success: true, MessageID: resp.ID, Status: "sent"}, nil
}

func (s *MessageService) enqueueMessage(ctx context.Context, instanceID string, jobType models.MessageJobType, payload interface{}) (*models.QueuedResponse, error) {
	correlationID := uuid.NewString()
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails("Error al serializar el payload del trabajo")
	}
	job := &models.MessageJob{
		InstanceID:    instanceID,
		CorrelationID: correlationID,
		Type:          jobType,
		Payload:       payloadBytes,
		RetryCount:    0,
	}
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails("Error al serializar el trabajo")
	}
	if err := s.redisClient.Client.LPush(ctx, config.QueueName, jobBytes).Err(); err != nil {
		log.Error().Err(err).Str("instance_id", instanceID).Msg("Error al encolar el trabajo en Redis")
		return nil, errors.ErrInternalServer.WithDetails("Error al encolar el mensaje")
	}
	log.Info().Str("instance_id", instanceID).Str("correlation_id", correlationID).Str("type", string(jobType)).Msg("Mensaje encolado para envío asíncrono")
	return &models.QueuedResponse{Status: "queued", CorrelationID: correlationID}, nil
}

func (s *MessageService) helperDownloadMediaBytes(mediaSource string) ([]byte, string, error) {
	if strings.HasPrefix(mediaSource, "data:") {
		parts := strings.Split(mediaSource, ",")
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("data uri inválida")
		}
		mimeParts := strings.Split(parts[0], ";")
		mimeType := strings.TrimPrefix(mimeParts[0], "data:")
		data, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, "", fmt.Errorf("error decodificando base64: %v", err)
		}
		return data, mimeType, nil
	}
	if _, err := url.ParseRequestURI(mediaSource); err != nil {
		return nil, "", fmt.Errorf("url inválida y no es data uri")
	}
	resp, err := http.Get(mediaSource)
	if err != nil {
		return nil, "", fmt.Errorf("error descargando media: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("error descarga status: %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("error leyendo body: %v", err)
	}
	mimeType := resp.Header.Get("Content-Type")
	return data, mimeType, nil
}

func (s *MessageService) SendLocation(ctx context.Context, instanceID string, req *models.SendLocationRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() { return nil, errors.ErrNotAuthenticated }
	recipientJID, err := types.ParseJID(req.Phone)
	if err != nil { return nil, errors.ErrBadRequest.WithDetails("Número de teléfono inválido") }
	msg := &waProto.Message{
		LocationMessage: &waProto.LocationMessage{
			DegreesLatitude:  proto.Float64(req.Latitude),
			DegreesLongitude: proto.Float64(req.Longitude),
			Name:             proto.String(req.Name),
			Address:          proto.String(req.Address),
		},
	}
	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil { return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando ubicación: %v", err)) }
	return &models.MessageResponse{Success: true, MessageID: resp.ID, Status: "sent"}, nil
}

func (s *MessageService) SendContact(ctx context.Context, instanceID string, req *models.SendContactRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() { return nil, errors.ErrNotAuthenticated }
	recipientJID, err := types.ParseJID(req.Phone)
	if err != nil { return nil, errors.ErrBadRequest.WithDetails("Número de teléfono inválido") }
	msg := &waProto.Message{
		ContactMessage: &waProto.ContactMessage{
			DisplayName: proto.String(req.DisplayName),
			Vcard:       proto.String(req.VCard),
		},
	}
	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil { return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando contacto: %v", err)) }
	return &models.MessageResponse{Success: true, MessageID: resp.ID, Status: "sent"}, nil
}

func (s *MessageService) ReactToMessage(ctx context.Context, instanceID string, req *models.ReactionRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() { return nil, errors.ErrNotAuthenticated }
	recipientJID, err := types.ParseJID(req.Phone)
	if err != nil { return nil, errors.ErrBadRequest.WithDetails("Número de teléfono inválido") }
	msg := &waProto.Message{
		ReactionMessage: &waProto.ReactionMessage{
			Key: &waProto.MessageKey{
				RemoteJID: proto.String(recipientJID.String()),
				FromMe:    proto.Bool(true),
				ID:        proto.String(req.MessageID),
			},
			Text:              proto.String(req.Emoji),
			SenderTimestampMS: proto.Int64(time.Now().UnixMilli()),
		},
	}
	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil { return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando reacción: %v", err)) }
	return &models.MessageResponse{Success: true, MessageID: resp.ID, Status: "sent"}, nil
}

func (s *MessageService) RevokeMessage(ctx context.Context, instanceID string, req *models.RevokeRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() { return nil, errors.ErrNotAuthenticated }
	recipientJID, err := types.ParseJID(req.Phone)
	if err != nil { return nil, errors.ErrBadRequest.WithDetails("Número de teléfono inválido") }
	msg := &waProto.Message{
		ProtocolMessage: &waProto.ProtocolMessage{
			Key: &waProto.MessageKey{
				RemoteJID: proto.String(recipientJID.String()),
				FromMe:    proto.Bool(true),
				ID:        proto.String(req.MessageID),
			},
			Type: waProto.ProtocolMessage_REVOKE.Enum(),
		},
	}
	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil { return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error revocando mensaje: %v", err)) }
	return &models.MessageResponse{Success: true, MessageID: resp.ID, Status: "sent"}, nil
}

func (s *MessageService) DownloadMedia(ctx context.Context, instanceID string, messageID string) ([]byte, string, string, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() {
		return nil, "", "", errors.ErrNotAuthenticated
	}
	return nil, "", "", errors.New(501, "Not Implemented")
}

func (s *MessageService) CreatePoll(ctx context.Context, instanceID string, req *models.CreatePollRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() { return nil, errors.ErrNotAuthenticated }
	// ... (poll logic)
	return nil, errors.New(501, "Not Implemented")
}

func (s *MessageService) VotePoll(ctx context.Context, instanceID string, req *models.VotePollRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() { return nil, errors.ErrNotAuthenticated }
	// ... (vote logic)
	return nil, errors.New(501, "Not Implemented")
}

func (s *MessageService) SendTextWithTyping(ctx context.Context, instanceID string, req *models.SendTextWithTypingRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() { return nil, errors.ErrNotAuthenticated }
	// ... (typing logic)
	return nil, errors.New(501, "Not Implemented")
}

func (s *MessageService) MarkAsRead(ctx context.Context, instanceID string, req *models.MarkAsReadRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() { return nil, errors.ErrNotAuthenticated }
	// ... (mark as read logic)
	return nil, errors.New(501, "Not Implemented")
}
