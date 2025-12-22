package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
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

	// Validar y limpiar número de teléfono
	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return nil, err
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)

	msg := &waE2E.Message{
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

// HelperDownloadMediaBytes descarga o decodifica los bytes del medio
// Incluye validaciones de seguridad: límite de tamaño y prevención de SSRF
func (s *MessageService) HelperDownloadMediaBytes(mediaSource string) ([]byte, string, error) {
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
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	// Validar y limpiar número de teléfono
	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return nil, err
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	data, mimeType, err := s.HelperDownloadMediaBytes(req.MediaURL)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails(err.Error())
	}

	uploaded, err := client.WAClient.Upload(ctx, data, whatsmeow.MediaImage)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error subiendo imagen: %v", err))
	}

	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)
	msg := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
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

	// Guardar en DB
	message := &models.Message{
		ID:         resp.ID,
		InstanceID: instanceID,
		To:         recipientJID.String(),
		From:       "me",
		Content:    req.Caption,
		Timestamp:  resp.Timestamp.Unix(),
		Type:       string(models.MessageTypeImage),
		IsFromMe:   true,
		Status:     string(models.MessageStatusSent),
	}
	if err := s.msgRepo.Create(ctx, message); err != nil {
		log.Error().Err(err).Msg("Error guardando imagen enviada en DB")
	}

	log.Info().
		Str("instance_id", instanceID).
		Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
		Str("message_id", resp.ID).
		Msg("Imagen enviada exitosamente")

	return &models.MessageResponse{
		Success:   true,
		MessageID: resp.ID,
		Status:    string(models.MessageStatusSent),
	}, nil
}

// SendVideo envía un video
func (s *MessageService) SendVideo(ctx context.Context, instanceID string, req *models.SendMediaRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	// Validar y limpiar número de teléfono
	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return nil, err
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	data, mimeType, err := s.HelperDownloadMediaBytes(req.MediaURL)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails(err.Error())
	}

	uploaded, err := client.WAClient.Upload(ctx, data, whatsmeow.MediaVideo)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error subiendo video: %v", err))
	}

	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)
	msg := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
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

	// Guardar en DB
	message := &models.Message{
		ID:         resp.ID,
		InstanceID: instanceID,
		To:         recipientJID.String(),
		From:       "me",
		Content:    req.Caption,
		Timestamp:  resp.Timestamp.Unix(),
		Type:       string(models.MessageTypeVideo),
		IsFromMe:   true,
		Status:     string(models.MessageStatusSent),
	}
	if err := s.msgRepo.Create(ctx, message); err != nil {
		log.Error().Err(err).Msg("Error guardando video enviado en DB")
	}

	log.Info().
		Str("instance_id", instanceID).
		Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
		Str("message_id", resp.ID).
		Msg("Video enviado exitosamente")

	return &models.MessageResponse{
		Success:   true,
		MessageID: resp.ID,
		Status:    string(models.MessageStatusSent),
	}, nil
}

// SendAudio envía un audio (nota de voz o audio general)
func (s *MessageService) SendAudio(ctx context.Context, instanceID string, req *models.SendMediaRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	// Validar y limpiar número de teléfono
	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return nil, err
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	data, mimeType, err := s.HelperDownloadMediaBytes(req.MediaURL)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails(err.Error())
	}

	uploaded, err := client.WAClient.Upload(ctx, data, whatsmeow.MediaAudio)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error subiendo audio: %v", err))
	}

	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)
	msg := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
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
		log.Error().
			Err(err).
			Str("instance_id", instanceID).
			Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
			Msg("Error enviando audio")

		if helpers.IsDatabaseLockedError(err) {
			return nil, errors.ErrDatabaseLocked
		}

		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando audio: %v", err))
	}

	// Guardar en DB
	message := &models.Message{
		ID:         resp.ID,
		InstanceID: instanceID,
		To:         recipientJID.String(),
		From:       "me",
		Content:    "Audio Message",
		Timestamp:  resp.Timestamp.Unix(),
		Type:       string(models.MessageTypeAudio),
		IsFromMe:   true,
		Status:     string(models.MessageStatusSent),
	}
	if err := s.msgRepo.Create(ctx, message); err != nil {
		log.Error().Err(err).Msg("Error guardando audio enviado en DB")
	}

	log.Info().
		Str("instance_id", instanceID).
		Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
		Str("message_id", resp.ID).
		Msg("Audio enviado exitosamente")

	return &models.MessageResponse{
		Success:   true,
		MessageID: resp.ID,
		Status:    string(models.MessageStatusSent),
	}, nil
}

// SendDocument envía un documento
func (s *MessageService) SendDocument(ctx context.Context, instanceID string, req *models.SendMediaRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	// Validar y limpiar número de teléfono
	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return nil, err
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	data, mimeType, err := s.HelperDownloadMediaBytes(req.MediaURL)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails(err.Error())
	}

	uploaded, err := client.WAClient.Upload(ctx, data, whatsmeow.MediaDocument)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error subiendo documento: %v", err))
	}

	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)
	msg := &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
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
		log.Error().
			Err(err).
			Str("instance_id", instanceID).
			Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
			Msg("Error enviando documento")

		if helpers.IsDatabaseLockedError(err) {
			return nil, errors.ErrDatabaseLocked
		}

		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando documento: %v", err))
	}

	// Guardar en DB
	message := &models.Message{
		ID:         resp.ID,
		InstanceID: instanceID,
		To:         recipientJID.String(),
		From:       "me",
		Content:    req.FileName,
		Timestamp:  resp.Timestamp.Unix(),
		Type:       string(models.MessageTypeDocument),
		IsFromMe:   true,
		Status:     string(models.MessageStatusSent),
	}
	if err := s.msgRepo.Create(ctx, message); err != nil {
		log.Error().Err(err).Msg("Error guardando documento enviado en DB")
	}

	log.Info().
		Str("instance_id", instanceID).
		Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
		Str("message_id", resp.ID).
		Msg("Documento enviado exitosamente")

	return &models.MessageResponse{
		Success:   true,
		MessageID: resp.ID,
		Status:    string(models.MessageStatusSent),
	}, nil
}

// SendLocation envía una ubicación
func (s *MessageService) SendLocation(ctx context.Context, instanceID string, req *models.SendLocationRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	// Validar y limpiar número de teléfono
	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return nil, err
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)

	msg := &waE2E.Message{
		LocationMessage: &waE2E.LocationMessage{
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
			Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
			Msg("Error enviando ubicación")

		if helpers.IsDatabaseLockedError(err) {
			return nil, errors.ErrDatabaseLocked
		}

		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando ubicación: %v", err))
	}

	// Guardar en DB
	message := &models.Message{
		ID:         resp.ID,
		InstanceID: instanceID,
		To:         recipientJID.String(),
		From:       "me",
		Content:    fmt.Sprintf("Lat: %f, Long: %f", req.Latitude, req.Longitude),
		Timestamp:  resp.Timestamp.Unix(),
		Type:       string(models.MessageTypeLocation),
		IsFromMe:   true,
		Status:     string(models.MessageStatusSent),
	}
	if err := s.msgRepo.Create(ctx, message); err != nil {
		log.Error().Err(err).Msg("Error guardando ubicación enviada en DB")
	}

	log.Info().
		Str("instance_id", instanceID).
		Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
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
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	// Validar y limpiar número de teléfono
	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return nil, err
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)

	msg := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: proto.String(req.DisplayName),
			Vcard:       proto.String(req.VCard),
		},
	}

	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		log.Error().
			Err(err).
			Str("instance_id", instanceID).
			Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
			Msg("Error enviando contacto")

		if helpers.IsDatabaseLockedError(err) {
			return nil, errors.ErrDatabaseLocked
		}
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando contacto: %v", err))
	}

	// Guardar en DB
	message := &models.Message{
		ID:         resp.ID,
		InstanceID: instanceID,
		To:         recipientJID.String(),
		From:       "me",
		Content:    fmt.Sprintf("Contact: %s", req.DisplayName),
		Timestamp:  resp.Timestamp.Unix(),
		Type:       string(models.MessageTypeContact),
		IsFromMe:   true,
		Status:     string(models.MessageStatusSent),
	}
	if err := s.msgRepo.Create(ctx, message); err != nil {
		log.Error().Err(err).Msg("Error guardando contacto enviado en DB")
	}

	return &models.MessageResponse{
		Success:   true,
		MessageID: resp.ID,
		Status:    string(models.MessageStatusSent),
	}, nil
}

// ReactToMessage es mi implementación para reaccionar a mensajes existentes.
// Aquí construyo manualmente el mensaje de reacción usando la nueva estructura de waE2E,
// asegurándome de incluir el timestamp actual para que WhatsApp lo procese correctamente.
func (s *MessageService) ReactToMessage(ctx context.Context, instanceID string, req *models.ReactionRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	// Primero valido el número para no intentar enviar nada a un destino inválido.
	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return nil, err
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)

	// Armo el mensaje de reacción. Me ha tocado usar waCommon.MessageKey porque
	// en la nueva versión de whatsmeow la llave del mensaje se movió de paquete.
	msg := &waE2E.Message{
		ReactionMessage: &waE2E.ReactionMessage{
			Key: &waCommon.MessageKey{
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
		log.Error().
			Err(err).
			Str("instance_id", instanceID).
			Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
			Msg("Tuve un fallo al enviar la reacción")

		if helpers.IsDatabaseLockedError(err) {
			return nil, errors.ErrDatabaseLocked
		}
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando reacción: %v", err))
	}

	return &models.MessageResponse{
		Success:   true,
		MessageID: resp.ID,
		Status:    "sent",
	}, nil
}

// RevokeMessage es mi método para eliminar mensajes ya enviados.
// Básicamente envío un ProtocolMessage de tipo REVOKE. Al igual que con las reacciones,
// aquí también uso waCommon.MessageKey para apuntar al mensaje original.
func (s *MessageService) RevokeMessage(ctx context.Context, instanceID string, req *models.RevokeRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return nil, err
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)

	// Construyo el mensaje de protocolo para la revocación. Es importante que el ID
	// coincida exactamente con el que queremos borrar.
	msg := &waE2E.Message{
		ProtocolMessage: &waE2E.ProtocolMessage{
			Key: &waCommon.MessageKey{
				RemoteJID: proto.String(recipientJID.String()),
				FromMe:    proto.Bool(true),
				ID:        proto.String(req.MessageID),
			},
			Type: waE2E.ProtocolMessage_REVOKE.Enum(),
		},
	}

	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		log.Error().
			Err(err).
			Str("instance_id", instanceID).
			Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
			Msg("Fallo al intentar revocar el mensaje")

		if helpers.IsDatabaseLockedError(err) {
			return nil, errors.ErrDatabaseLocked
		}
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error revocando mensaje: %v", err))
	}

	return &models.MessageResponse{
		Success:   true,
		MessageID: resp.ID,
		Status:    "revoked",
	}, nil
}

// CreatePoll es mi forma de crear nuevas encuestas.
// Me apoyo en el helper oficial de whatsmeow (BuildPollCreation) que me facilita
// mucho la vida al no tener que armar el mensaje bit a bit.
func (s *MessageService) CreatePoll(ctx context.Context, instanceID string, req *models.CreatePollRequest) (*models.PollResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	// Limpio el número como siempre para evitar líos con el formato.
	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return nil, err
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)

	// Aquí es donde el SDK se encarga de todo lo pesado de construir la encuesta.
	// Solo le paso la pregunta, las opciones y cuántas se pueden elegir.
	msg := client.WAClient.BuildPollCreation(req.Question, req.Options, int(req.SelectableCount))

	// Finalmente lo mando y capturo cualquier error, especialmente los de base de datos bloqueada.
	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		log.Error().
			Err(err).
			Str("instance_id", instanceID).
			Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
			Msg("No pude enviar la encuesta")

		if helpers.IsDatabaseLockedError(err) {
			return nil, errors.ErrDatabaseLocked
		}
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando encuesta: %v", err))
	}

	return &models.PollResponse{
		Success:   true,
		MessageID: resp.ID,
	}, nil
}

// VotePoll permite votar en una encuesta.
// Para esto, necesito recrear un MessageInfo básico del mensaje original.
func (s *MessageService) VotePoll(ctx context.Context, instanceID string, req *models.VotePollRequest) (*models.PollResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return nil, err
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)

	// Como no guardo todo el objeto MessageInfo en mi DB, armo este "esqueleto"
	// que es lo mínimo que me pide whatsmeow para identificar a qué encuesta estamos votando.
	pollInfo := &types.MessageInfo{
		MessageSource: types.MessageSource{
			Chat: recipientJID,
		},
		ID: req.MessageID,
	}

	// Llamo al helper para construir el mensaje de voto. Es mucho más limpio así.
	msg, err := client.WAClient.BuildPollVote(ctx, pollInfo, req.OptionNames)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Me fue imposible construir el voto: %v", err))
	}

	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		log.Error().
			Err(err).
			Str("instance_id", instanceID).
			Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
			Msg("Tuve un error al votar")

		if helpers.IsDatabaseLockedError(err) {
			return nil, errors.ErrDatabaseLocked
		}
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error votando en encuesta: %v", err))
	}

	return &models.PollResponse{
		Success:   true,
		MessageID: resp.ID,
	}, nil
}

// DownloadMedia descarga un archivo multimedia
func (s *MessageService) DownloadMedia(ctx context.Context, instanceID string, req *models.DownloadMediaRequest) ([]byte, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Decodificar Base64 de las llaves
	mediaKey, err := base64.StdEncoding.DecodeString(req.MediaKey)
	if err != nil {
		return nil, fmt.Errorf("error decodificando media_key: %v", err)
	}
	fileEncSha256, err := base64.StdEncoding.DecodeString(req.FileEncSha256)
	if err != nil {
		return nil, fmt.Errorf("error decodificando file_enc_sha256: %v", err)
	}
	fileSha256, err := base64.StdEncoding.DecodeString(req.FileSha256)
	if err != nil {
		return nil, fmt.Errorf("error decodificando file_sha256: %v", err)
	}

	// Reconstruir metadatos del mensaje para la descarga
	// WhatsApp usa diferentes estructuras según el tipo
	var downloadable whatsmeow.DownloadableMessage

	switch req.Type {
	case "image":
		downloadable = &waE2E.ImageMessage{
			DirectPath:    proto.String(req.DirectPath),
			URL:           proto.String(req.Url),
			MediaKey:      mediaKey,
			FileEncSHA256: fileEncSha256,
			FileSHA256:    fileSha256,
			FileLength:    proto.Uint64(req.FileLength),
			Mimetype:      proto.String(req.Mimetype),
		}
	case "video":
		downloadable = &waE2E.VideoMessage{
			DirectPath:    proto.String(req.DirectPath),
			URL:           proto.String(req.Url),
			MediaKey:      mediaKey,
			FileEncSHA256: fileEncSha256,
			FileSHA256:    fileSha256,
			FileLength:    proto.Uint64(req.FileLength),
			Mimetype:      proto.String(req.Mimetype),
		}
	case "audio":
		downloadable = &waE2E.AudioMessage{
			DirectPath:    proto.String(req.DirectPath),
			URL:           proto.String(req.Url),
			MediaKey:      mediaKey,
			FileEncSHA256: fileEncSha256,
			FileSHA256:    fileSha256,
			FileLength:    proto.Uint64(req.FileLength),
			Mimetype:      proto.String(req.Mimetype),
		}
	case "document":
		downloadable = &waE2E.DocumentMessage{
			DirectPath:    proto.String(req.DirectPath),
			URL:           proto.String(req.Url),
			MediaKey:      mediaKey,
			FileEncSHA256: fileEncSha256,
			FileSHA256:    fileSha256,
			FileLength:    proto.Uint64(req.FileLength),
			Mimetype:      proto.String(req.Mimetype),
		}
	case "sticker":
		downloadable = &waE2E.StickerMessage{
			DirectPath:    proto.String(req.DirectPath),
			URL:           proto.String(req.Url),
			MediaKey:      mediaKey,
			FileEncSHA256: fileEncSha256,
			FileSHA256:    fileSha256,
			FileLength:    proto.Uint64(req.FileLength),
			Mimetype:      proto.String(req.Mimetype),
		}
	default:
		return nil, fmt.Errorf("tipo de media no soportado: %s", req.Type)
	}

	// Realizar la descarga y desencriptación
	data, err := client.WAClient.Download(ctx, downloadable)
	if err != nil {
		return nil, fmt.Errorf("error descargando media: %v", err)
	}

	return data, nil
}

// SendTextWithTyping envía un mensaje de texto con simulación de escritura
func (s *MessageService) SendTextWithTyping(ctx context.Context, instanceID string, req *models.SendTextWithTypingRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	// Validar y limpiar número de teléfono
	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return nil, err
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)

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
	msg := &waE2E.Message{
		Conversation: proto.String(req.Message),
	}

	resp, err := client.WAClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		log.Error().
			Err(err).
			Str("instance_id", instanceID).
			Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
			Msg("Error enviando mensaje con typing")

		if helpers.IsDatabaseLockedError(err) {
			return nil, errors.ErrDatabaseLocked
		}
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error enviando mensaje: %v", err))
	}

	// Guardar mensaje en la base de datos
	dbMsg := &models.Message{
		ID:         resp.ID,
		InstanceID: instanceID,
		From:       "me",
		To:         recipientJID.String(),
		Type:       string(models.MessageTypeText),
		Content:    req.Message,
		Timestamp:  resp.Timestamp.Unix(),
		IsFromMe:   true,
		Status:     string(models.MessageStatusSent),
	}

	if err := s.msgRepo.Create(ctx, dbMsg); err != nil {
		log.Error().Err(err).Msg("Error guardando mensaje en DB")
	}

	log.Info().
		Str("instance_id", instanceID).
		Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
		Str("message_id", resp.ID).
		Msg("Mensaje con typing enviado exitosamente")

	return &models.MessageResponse{
		Success:   true,
		MessageID: resp.ID,
		Status:    string(models.MessageStatusSent),
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

// EditMessage edita un mensaje de texto enviado previamente.
// Solo se pueden editar mensajes propios dentro de un periodo de tiempo (usualmente 15 min).
// He implementado esto usando el helper BuildEdit de whatsmeow, que encapsula correctamente
// la estructura de ProtocolMessage necesaria.
func (s *MessageService) EditMessage(ctx context.Context, instanceID string, req *models.EditMessageRequest) (*models.MessageResponse, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}

	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return nil, err
	}

	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	recipientJID := types.NewJID(cleanPhone, types.DefaultUserServer)

	log.Info().
		Str("instance_id", instanceID).
		Str("recipient", validators.MaskPhoneNumber(cleanPhone)).
		Str("original_id", req.MessageID).
		Msg("Editando mensaje")

	// Construimos el mensaje de edición.
	// whatsmeow requiere el JID, el ID original y el nuevo contenido.
	editMsg := client.WAClient.BuildEdit(recipientJID, req.MessageID, &waE2E.Message{
		Conversation: proto.String(req.NewText),
	})

	resp, err := client.WAClient.SendMessage(ctx, recipientJID, editMsg)
	if err != nil {
		log.Error().Err(err).Str("instance_id", instanceID).Str("msg_id", req.MessageID).Msg("Error editando mensaje")
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error editando mensaje: %v", err))
	}

	return &models.MessageResponse{
		Success:   true,
		MessageID: resp.ID,
		Status:    "edited",
	}, nil
}
