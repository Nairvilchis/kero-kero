package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"kero-kero/internal/models"
	"kero-kero/internal/repository"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
	"kero-kero/pkg/validators"
)

type ChatService struct {
	waManager *whatsapp.Manager
	msgRepo   *repository.MessageRepository
}

func NewChatService(waManager *whatsapp.Manager, msgRepo *repository.MessageRepository) *ChatService {
	return &ChatService{
		waManager: waManager,
		msgRepo:   msgRepo,
	}
}

// ListChats lista los chats activos (solo aquellos con mensajes)
func (s *ChatService) ListChats(ctx context.Context, instanceID string) ([]models.Chat, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Obtener chats únicos desde la tabla de mensajes
	// Solo mostramos chats que tienen al menos un mensaje
	chats, err := s.msgRepo.GetChatsWithMessages(ctx, instanceID)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error obteniendo chats: %v", err))
	}

	// Enriquecer con información de contactos
	contacts, err := client.WAClient.Store.Contacts.GetAllContacts(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Error obteniendo contactos, usando solo JIDs")
	}

	// Filtrar y clasificar chats
	var filteredChats []models.Chat
	for i := range chats {
		// Clasificar tipo de chat
		chats[i].ChatType = whatsapp.GetChatType(chats[i].JID)

		// Filtrar estados (status@broadcast)
		if chats[i].ChatType == "status" {
			continue
		}

		// Enriquecer con nombres de contactos
		if contacts != nil {
			jid, err := types.ParseJID(chats[i].JID)
			if err == nil {
				if contact, ok := contacts[jid]; ok && contact.Found {
					if contact.FullName != "" {
						chats[i].Name = contact.FullName
					} else if contact.PushName != "" {
						chats[i].Name = contact.PushName
					}
				}
			}
		}

		// Si no hay nombre, usar el número del JID
		if chats[i].Name == "" {
			parts := strings.Split(chats[i].JID, "@")
			if len(parts) > 0 {
				chats[i].Name = parts[0]
			}
		}

		filteredChats = append(filteredChats, chats[i])
	}

	return filteredChats, nil
}

// GetChatHistory obtiene el historial de mensajes de un chat
func (s *ChatService) GetChatHistory(ctx context.Context, instanceID, jidStr string, limit int) ([]models.Message, error) {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return nil, errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return nil, errors.ErrNotAuthenticated
	}

	// Validar JID
	_, err := types.ParseJID(jidStr)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("JID inválido")
	}

	// Obtener mensajes de la base de datos
	messages, err := s.msgRepo.GetByJID(ctx, instanceID, jidStr, limit)
	if err != nil {
		return nil, errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error obteniendo historial: %v", err))
	}

	return messages, nil
}

// ArchiveChat es mi implementación para archivar o desarchivar un chat.
// Me he dado cuenta de que whatsmeow lo hace a través de parches de App State,
// así que aquí construyo el parche y lo envío de forma atómica.
func (s *ChatService) ArchiveChat(ctx context.Context, instanceID string, req *models.ArchiveChatRequest) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}

	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return err
	}

	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	jid := types.NewJID(cleanPhone, types.DefaultUserServer)

	// Para archivar correctamente, WhatsApp me pide saber qué mensaje es el último.
	// Intento obtenerlo de mi repositorio local.
	lastMsg, err := s.msgRepo.GetLastMessage(ctx, instanceID, jid.String())
	var lastTs time.Time
	var lastKey *waCommon.MessageKey

	if lastMsg != nil {
		lastTs = time.Unix(lastMsg.Timestamp, 0)
		lastKey = &waCommon.MessageKey{
			RemoteJID: proto.String(jid.String()),
			FromMe:    proto.Bool(lastMsg.IsFromMe),
			ID:        proto.String(lastMsg.ID),
		}
	} else {
		// Si no hay mensajes, uso el tiempo actual; WhatsApp suele aceptar esto para chats vacíos.
		lastTs = time.Now()
	}

	patch := appstate.BuildArchive(jid, req.Archived, lastTs, lastKey)
	if err := client.WAClient.SendAppState(ctx, patch); err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Fallo al enviar parche de archivo: %v", err))
	}

	return nil
}

// UpdateStatus actualiza el estado de texto (About)
func (s *ChatService) UpdateStatus(ctx context.Context, instanceID, status string) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	if err := client.WAClient.SetStatusMessage(ctx, status); err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error actualizando estado: %v", err))
	}

	return nil
}

// DeleteChat es mi método para borrar permanentemente un chat.
// No solo limpio mi base de datos, sino que también le digo a WhatsApp que lo borre
// para que desaparezca de todos los dispositivos sincronizados.
func (s *ChatService) DeleteChat(ctx context.Context, instanceID, jidStr string) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return errors.ErrBadRequest.WithDetails("JID inválido")
	}

	// Primero obtengo el último mensaje para que el comando de borrado sea preciso.
	lastMsg, _ := s.msgRepo.GetLastMessage(ctx, instanceID, jidStr)
	var lastTs time.Time
	var lastKey *waCommon.MessageKey

	if lastMsg != nil {
		lastTs = time.Unix(lastMsg.Timestamp, 0)
		lastKey = &waCommon.MessageKey{
			RemoteJID: proto.String(jid.String()),
			FromMe:    proto.Bool(lastMsg.IsFromMe),
			ID:        proto.String(lastMsg.ID),
		}
	} else {
		lastTs = time.Now()
	}

	// Envío la señal a WhatsApp.
	patch := appstate.BuildDeleteChat(jid, lastTs, lastKey)
	if err := client.WAClient.SendAppState(ctx, patch); err != nil {
		log.Warn().Err(err).Msg("No pude sincronizar el borrado del chat con WhatsApp, borrando solo local")
	}

	// Y limpio los mensajes de mi base de datos local.
	if err := s.msgRepo.DeleteChatMessages(ctx, instanceID, jidStr); err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error eliminando chat local: %v", err))
	}

	return nil
}

// MarkAsRead marca todos los mensajes pendientes de un chat como leídos.
// He implementado esto usando BuildMarkChatAsRead para que el cambio se sincronice
// correctamente en el teléfono y otros clientes web/desktop.
func (s *ChatService) MarkAsRead(ctx context.Context, instanceID, jidStr string) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return errors.ErrBadRequest.WithDetails("JID inválido")
	}

	// Necesitamos identificar hasta qué mensaje marcamos como leído.
	lastMsg, _ := s.msgRepo.GetLastMessage(ctx, instanceID, jidStr)
	var lastTs time.Time
	var lastKey *waCommon.MessageKey

	if lastMsg != nil {
		lastTs = time.Unix(lastMsg.Timestamp, 0)
		lastKey = &waCommon.MessageKey{
			RemoteJID: proto.String(jid.String()),
			FromMe:    proto.Bool(lastMsg.IsFromMe),
			ID:        proto.String(lastMsg.ID),
		}
	} else {
		lastTs = time.Now()
	}

	log.Info().Str("instance_id", instanceID).Str("jid", jidStr).Msg("Sincronizando marca de lectura en WhatsApp")

	patch := appstate.BuildMarkChatAsRead(jid, true, lastTs, lastKey)
	if err := client.WAClient.SendAppState(ctx, patch); err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Fallo al marcar chat como leído: %v", err))
	}

	return nil
}

// MuteChat silencia o quita el silencio de un chat.
// He añadido esto porque whatsmeow permite sincronizar el silencio entre dispositivos.
func (s *ChatService) MuteChat(ctx context.Context, instanceID string, req *models.MuteChatRequest) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}

	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return err
	}

	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	jid := types.NewJID(cleanPhone, types.DefaultUserServer)

	log.Info().Str("instance_id", instanceID).Str("jid", jid.String()).Bool("muted", req.Muted).Msg("Cambiando estado de silencio del chat")

	var duration time.Duration
	if req.Muted && req.Duration > 0 {
		duration = time.Duration(req.Duration) * time.Second
	} else if req.Muted {
		duration = 24 * 365 * 10 * time.Hour // "Para siempre" (aprox 10 años)
	}

	patch := appstate.BuildMute(jid, req.Muted, duration)
	if err := client.WAClient.SendAppState(ctx, patch); err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Fallo al silenciar chat: %v", err))
	}

	return nil
}

// PinChat fija o desfija un chat en la parte superior.
// Muy útil para mantener chats importantes siempre a la vista, sincronizado con WhatsApp.
func (s *ChatService) PinChat(ctx context.Context, instanceID string, req *models.PinChatRequest) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}

	cleanPhone, err := validators.ValidatePhoneNumber(req.Phone)
	if err != nil {
		return err
	}

	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	jid := types.NewJID(cleanPhone, types.DefaultUserServer)

	log.Info().Str("instance_id", instanceID).Str("jid", jid.String()).Bool("pinned", req.Pinned).Msg("Cambiando pin del chat")

	patch := appstate.BuildPin(jid, req.Pinned)
	if err := client.WAClient.SendAppState(ctx, patch); err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Fallo al fijar chat: %v", err))
	}

	return nil
}
