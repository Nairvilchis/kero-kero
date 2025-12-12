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
	"kero-kero/pkg/helpers"
	"kero-kero/pkg/validators"
)

// ... (resto del código)

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

// SendText envía un mensaje de texto
func (s *MessageService) SendText(ctx context.Context, instanceID string, req *models.SendTextRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Validar y limpiar número de teléfono
	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return nil, err
	}

	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)

	msg := &waProto.Message{
		Conversation: proto.String(req.Message),
	}

	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		log.Error().
			Err(err).
			Str("instance_id", instanceID).
			Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
			Msg("Error enviando mensaje")

		// Usar helper para detectar errores de database locked
		if helpers.IsDatabaseLockedError(err) {
			return nil, errors.ErrDatabaseLocked
		}

		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando mensaje: %v", err))
	}

	// Guardar mensaje en base de datos
	message := &models.Message{
		ID:         resp.ID,
		InstanceID: instanceID,
		To:         recipientJID.String(),
		From:       "me",
		Content:    req.Message,
		Timestamp:  resp.Timestamp.Unix(),
		Type:       "text",
		IsFromMe:   true,
		Status:     string(models.MessageStatusSent),
	}

	if err := s.msgRepo.Create(ctx, message); err != nil {
		log.Error().Err(err).Msg("Error guardando mensaje enviado en DB")
		// No fallamos el request si falla guardar en DB, pero logueamos
	}

	log.Info().
		Str("instance_id", instanceID).
		Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
		Str("message_id", resp.ID).
		Msg("Mensaje enviado exitosamente")

	return &models.MessageResponse{
		Success:   true,
		MessageID: resp.ID,
		Status:    string(models.MessageStatusSent),
	}, nil
}

// helperDownloadMediaBytes descarga o decodifica los bytes del medio
// Incluye validaciones de seguridad: límite de tamaño y prevención de SSRF
func (s *MessageService) helperDownloadMediaBytes(mediaSource string) ([]byte, string, error) {
	// Límite de tamaño: 50MB (ajustable según necesidades)
	const maxSize = 50 * 1024 * 1024 // 50MB

	// 1. Verificar si es Data URI (Base64)
	if strings.HasPrefix(mediaSource, "data:") {
		parts := strings.Split(mediaSource, ",")
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("data uri inválida")
		}
		// Extraer mime type: data:image/jpeg;base64
		mimeParts := strings.Split(parts[0], ";")
		mimeType := strings.TrimPrefix(mimeParts[0], "data:")

		data, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, "", fmt.Errorf("error decodificando base64: %v", err)
		}

		// Verificar tamaño del archivo decodificado
		if len(data) > maxSize {
			return nil, "", errors.New(413, fmt.Sprintf("Archivo demasiado grande (máximo %dMB)", maxSize/(1024*1024)))
		}

		return data, mimeType, nil
	}

	// 2. Verificar si es URL válida
	if _, err := url.ParseRequestURI(mediaSource); err != nil {
		return nil, "", fmt.Errorf("url inválida y no es data uri")
	}

	// 3. Validar URL para prevenir SSRF
	// Importar: "kero-kero/pkg/validators"
	if err := validators.ValidateMediaURL(mediaSource); err != nil {
		return nil, "", err
	}

	// 4. Descargar desde URL con timeout
	client := &http.Client{
		Timeout: 30 * time.Second, // Timeout de 30 segundos
	}

	resp, err := client.Get(mediaSource)
	if err != nil {
		return nil, "", fmt.Errorf("error descargando media: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("error descarga status: %d", resp.StatusCode)
	}

	// 5. Verificar Content-Length si está disponible
	if resp.ContentLength > maxSize {
		return nil, "", errors.New(413, fmt.Sprintf("Archivo demasiado grande (máximo %dMB)", maxSize/(1024*1024)))
	}

	// 6. Usar LimitReader para prevenir lecturas excesivas
	limitedReader := io.LimitReader(resp.Body, maxSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, "", fmt.Errorf("error leyendo body: %v", err)
	}

	// 7. Verificar que no excedió el límite
	if len(data) > maxSize {
		return nil, "", errors.New(413, fmt.Sprintf("Archivo demasiado grande (máximo %dMB)", maxSize/(1024*1024)))
	}

	mimeType := resp.Header.Get("Content-Type")
	return data, mimeType, nil
}

// SendImage envía una imagen
func (s *MessageService) SendImage(ctx context.Context, instanceID string, req *models.SendMediaRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Validar y limpiar número de teléfono
	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return nil, err
	}

	data, mimeType, err := s.helperDownloadMediaBytes(req.MediaURL)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails(err.Error())
	}

	uploaded, err := client.WAClient.Upload(ctx, data, whatsmeow.MediaImage)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error subiendo imagen: %v", err))
	}

	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)
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
		log.Error().
			Err(err).
			Str("instance_id", instanceID).
			Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
			Msg("Error enviando imagen")

		if helpers.IsDatabaseLockedError(err) {
			return nil, errors.ErrDatabaseLocked
		}

		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando imagen: %v", err))
	}

	log.Info().
		Str("instance_id", instanceID).
		Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
		Str("message_id", resp.ID).
		Msg("Imagen enviada exitosamente")

	return &models.MessageResponse{Success: true, MessageID: resp.ID, Status: "sent"}, nil
}

// SendVideo envía un video
func (s *MessageService) SendVideo(ctx context.Context, instanceID string, req *models.SendMediaRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Validar y limpiar número de teléfono
	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return nil, err
	}

	data, mimeType, err := s.helperDownloadMediaBytes(req.MediaURL)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails(err.Error())
	}

	uploaded, err := client.WAClient.Upload(ctx, data, whatsmeow.MediaVideo)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error subiendo video: %v", err))
	}

	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)
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
		log.Error().
			Err(err).
			Str("instance_id", instanceID).
			Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
			Msg("Error enviando video")

		if helpers.IsDatabaseLockedError(err) {
			return nil, errors.ErrDatabaseLocked
		}

		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando video: %v", err))
	}

	log.Info().
		Str("instance_id", instanceID).
		Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
		Str("message_id", resp.ID).
		Msg("Video enviado exitosamente")

	return &models.MessageResponse{Success: true, MessageID: resp.ID, Status: "sent"}, nil
}

// SendAudio envía un audio (nota de voz o audio general)
func (s *MessageService) SendAudio(ctx context.Context, instanceID string, req *models.SendMediaRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	data, mimeType, err := s.helperDownloadMediaBytes(req.MediaURL)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails(err.Error())
	}

	uploaded, err := client.WAClient.Upload(ctx, data, whatsmeow.MediaAudio)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error subiendo audio: %v", err))
	}

	recipientJID := types.NewJID(req.Phone, types.DefaultUserServer)
	msg := &waProto.Message{
		AudioMessage: &waProto.AudioMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			PTT:           proto.Bool(true), // Por defecto como nota de voz
		},
	}

	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando audio: %v", err))
	}

	return &models.MessageResponse{Success: true, MessageID: resp.ID, Status: "sent"}, nil
}

// SendDocument envía un documento
func (s *MessageService) SendDocument(ctx context.Context, instanceID string, req *models.SendMediaRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	data, mimeType, err := s.helperDownloadMediaBytes(req.MediaURL)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails(err.Error())
	}

	uploaded, err := client.WAClient.Upload(ctx, data, whatsmeow.MediaDocument)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error subiendo documento: %v", err))
	}

	recipientJID := types.NewJID(req.Phone, types.DefaultUserServer)
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

// SendLocation envía una ubicación
func (s *MessageService) SendLocation(ctx context.Context, instanceID string, req *models.SendLocationRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	recipientJID := types.NewJID(req.Phone, types.DefaultUserServer)

	msg := &waProto.Message{
		LocationMessage: &waProto.LocationMessage{
			DegreesLatitude:  proto.Float64(req.Latitude),
			DegreesLongitude: proto.Float64(req.Longitude),
			Name:             proto.String(req.Name),
			Address:          proto.String(req.Address),
		},
	}

	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		log.Error().
			Err(err).
			Str("instance_id", instanceID).
			Str("recipient", req.Phone).
			Msg("Error enviando ubicación")

		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando ubicación: %v", err))
	}

	log.Info().
		Str("instance_id", instanceID).
		Str("recipient", req.Phone).
		Str("message_id", resp.ID).
		Msg("Ubicación enviada exitosamente")

	return &models.MessageResponse{
		Success:   true,
		MessageID: resp.ID,
		Status:    string(models.MessageStatusSent),
	}, nil
}

// SendContact envía un contacto (vCard)
func (s *MessageService) SendContact(ctx context.Context, instanceID string, req *models.SendContactRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil || !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	recipientJID := types.NewJID(req.Phone, types.DefaultUserServer)

	msg := &waProto.Message{
		ContactMessage: &waProto.ContactMessage{
			DisplayName: proto.String(req.DisplayName),
			Vcard:       proto.String(req.VCard),
		},
	}

	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando contacto: %v", err))
	}

	return &models.MessageResponse{Success: true, MessageID: resp.ID, Status: "sent"}, nil
}

// ReactToMessage reacciona a un mensaje
func (s *MessageService) ReactToMessage(ctx context.Context, instanceID string, req *models.ReactionRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	recipientJID := types.NewJID(req.Phone, types.DefaultUserServer)

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
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando reacción: %v", err))
	}

	return &models.MessageResponse{
		Success:   true,
		MessageID: resp.ID,
		Status:    string(models.MessageStatusSent),
	}, nil
}

// RevokeMessage elimina un mensaje enviado (para todos)
func (s *MessageService) RevokeMessage(ctx context.Context, instanceID string, req *models.RevokeRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	recipientJID := types.NewJID(req.Phone, types.DefaultUserServer)

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
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error revocando mensaje: %v", err))
	}

	return &models.MessageResponse{
		Success:   true,
		MessageID: resp.ID,
		Status:    string(models.MessageStatusSent),
	}, nil
}

// CreatePoll crea una encuesta
func (s *MessageService) CreatePoll(ctx context.Context, instanceID string, req *models.CreatePollRequest) (*models.PollResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	recipientJID := types.NewJID(req.Phone, types.DefaultUserServer)

	// Construir opciones de la encuesta
	pollOptions := make([]*waProto.PollCreationMessage_Option, len(req.Options))
	for i, option := range req.Options {
		pollOptions[i] = &waProto.PollCreationMessage_Option{
			OptionName: proto.String(option),
		}
	}

	// Crear mensaje de encuesta
	pollCreation := &waProto.PollCreationMessage{
		Name:                   proto.String(req.Question),
		Options:                pollOptions,
		SelectableOptionsCount: proto.Uint32(req.SelectableCount),
	}

	msg := &waProto.Message{
		PollCreationMessage: pollCreation,
	}

	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando encuesta: %v", err))
	}

	return &models.PollResponse{
		Success:   true,
		MessageID: resp.ID,
	}, nil
}

// VotePoll vota en una encuesta
func (s *MessageService) VotePoll(ctx context.Context, instanceID string, req *models.VotePollRequest) (*models.PollResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	recipientJID := types.NewJID(req.Phone, types.DefaultUserServer)

	// Construir opciones seleccionadas (hashes SHA256 de los nombres)
	// WhatsApp usa SHA256 de los nombres de opciones como identificadores
	selectedOptions := make([][]byte, len(req.OptionNames))
	for i, optionName := range req.OptionNames {
		hash := sha256.Sum256([]byte(optionName))
		selectedOptions[i] = hash[:]
	}

	// Crear mensaje de voto
	pollUpdate := &waProto.PollUpdateMessage{
		PollCreationMessageKey: &waProto.MessageKey{
			RemoteJID: proto.String(recipientJID.String()),
			FromMe:    proto.Bool(false), // La encuesta fue creada por otro usuario
			ID:        proto.String(req.MessageID),
		},
		Vote: &waProto.PollEncValue{
			EncPayload: selectedOptions[0], // Simplificado: solo primera opción
		},
	}

	msg := &waProto.Message{
		PollUpdateMessage: pollUpdate,
	}

	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error votando en encuesta: %v", err))
	}

	return &models.PollResponse{
		Success:   true,
		MessageID: resp.ID,
	}, nil
}

// DownloadMedia descarga un archivo multimedia bajo demanda
func (s *MessageService) DownloadMedia(ctx context.Context, instanceID string, messageID string) ([]byte, string, string, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, "", "", errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, "", "", errors.ErrNotAuthenticated
	}

	// NOTA IMPORTANTE: La descarga bajo demanda requiere guardar los metadatos
	// completos del mensaje de WhatsApp (DirectPath, MediaKey, FileEncSha256, etc.)
	// que contienen la información necesaria para descargar y desencriptar el archivo.
	//
	// Actualmente, solo guardamos el contenido del mensaje en texto, no los metadatos
	// binarios necesarios para la descarga.
	//
	// SOLUCIÓN RECOMENDADA: Aumentar los límites de inline (16MB/50MB) para que
	// la mayoría de archivos se envíen directamente en el webhook en Base64.
	//
	// Para implementar descarga bajo demanda en el futuro, se necesitaría:
	// 1. Guardar los metadatos del mensaje original en la BD
	// 2. Reconstruir el DownloadableMessage desde esos metadatos
	// 3. Llamar a client.WAClient.Download() con ese mensaje

	return nil, "", "", errors.New(501, "Descarga bajo demanda no implementada. Los archivos multimedia se envían inline en el webhook (hasta 16MB para imágenes/videos/audio, 50MB para documentos).")
}

// SendTextWithTyping envía un mensaje de texto con simulación de escritura
func (s *MessageService) SendTextWithTyping(ctx context.Context, instanceID string, req *models.SendTextWithTypingRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Construir JID del destinatario
	recipientJID, err := types.ParseJID(req.Phone)
	if err != nil {
		recipientJID, err = types.ParseJID(req.Phone + "@s.whatsapp.net")
		if err != nil {
			return nil, errors.ErrBadRequest.WithDetails("Número de teléfono inválido")
		}
	}

	// Calcular duración de typing si no se proporcionó
	typingDuration := 0
	if req.TypingDuration != nil {
		typingDuration = *req.TypingDuration
	} else {
		// Fórmula: ~50ms por carácter, mínimo 500ms, máximo 5000ms
		typingDuration = len(req.Message) * 50
		if typingDuration < 500 {
			typingDuration = 500
		}
		if typingDuration > 5000 {
			typingDuration = 5000
		}
	}

	// 1. Activar presencia "typing"
	err = client.WAClient.SendChatPresence(ctx, recipientJID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	if err != nil {
		log.Warn().Err(err).Msg("Error enviando presencia typing (continuando de todos modos)")
	}

	// 2. Esperar la duración calculada
	time.Sleep(time.Duration(typingDuration) * time.Millisecond)

	// 3. Detener presencia
	err = client.WAClient.SendChatPresence(ctx, recipientJID, types.ChatPresencePaused, types.ChatPresenceMediaText)
	if err != nil {
		log.Warn().Err(err).Msg("Error deteniendo presencia (continuando de todos modos)")
	}

	// 4. Enviar el mensaje de texto
	msg := &waProto.Message{
		Conversation: proto.String(req.Message),
	}

	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando mensaje: %v", err))
	}

	// Guardar mensaje en la base de datos
	dbMsg := &models.Message{
		ID:         resp.ID,
		InstanceID: instanceID,
		From:       client.WAClient.Store.ID.User,
		To:         recipientJID.User,
		Type:       "text",
		Content:    req.Message,
		Timestamp:  resp.Timestamp.Unix(),
		IsFromMe:   true,
		Status:     "sent",
	}

	if err := s.msgRepo.Create(ctx, dbMsg); err != nil {
		log.Error().Err(err).Msg("Error guardando mensaje en DB")
	}

	return &models.MessageResponse{
		Success:   true,
		MessageID: resp.ID,
		Status:    "sent",
	}, nil
}

// MarkAsRead marca uno o más mensajes como leídos
func (s *MessageService) MarkAsRead(ctx context.Context, instanceID string, req *models.MarkAsReadRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Parsear JIDs
	chatJID, err := types.ParseJID(req.ChatJID)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("Chat JID inválido")
	}

	senderJID, err := types.ParseJID(req.SenderJID)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("Sender JID inválido")
	}

	// Convertir timestamp a time.Time
	timestamp := time.Unix(req.Timestamp, 0)

	// Marcar como leído
	ids := []types.MessageID{req.MessageID}
	err = client.WAClient.MarkRead(ctx, ids, timestamp, chatJID, senderJID)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error marcando mensaje como leído: %v", err))
	}

	return &models.MessageResponse{
		Success:   true,
		MessageID: req.MessageID,
		Status:    "read",
	}, nil
}
