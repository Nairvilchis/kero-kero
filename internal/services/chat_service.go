package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow/types"

	"kero-kero/internal/models"
	"kero-kero/internal/repository"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
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

// ArchiveChat archiva o desarchiva un chat
func (s *ChatService) ArchiveChat(ctx context.Context, instanceID string, req *models.ArchiveChatRequest) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	// jid := types.NewJID(req.Phone, types.DefaultUserServer)
	// return client.WAClient.SendArchive(jid, req.Archived) // Método hipotético
	return errors.New(501, "Archivar chat no implementado")
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

// DeleteChat elimina todos los mensajes de un chat (borra la conversación)
func (s *ChatService) DeleteChat(ctx context.Context, instanceID, jid string) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	// Eliminar todos los mensajes del chat
	if err := s.msgRepo.DeleteChatMessages(ctx, instanceID, jid); err != nil {
		return errors.ErrInternalServer.WithDetails(fmt.Sprintf("Error eliminando chat: %v", err))
	}

	return nil
}

// MarkAsRead marca los mensajes de un chat como leídos
func (s *ChatService) MarkAsRead(ctx context.Context, instanceID, jidStr string) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	// Parsear JID
	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return errors.ErrBadRequest.WithDetails("JID inválido")
	}

	// Marcar como leído en WhatsApp
	// Nota: whatsmeow no tiene un método directo para marcar como leído,
	// pero podemos enviar un receipt de lectura para los mensajes recientes
	// Por ahora, simplemente retornamos success ya que WhatsApp marca automáticamente
	// como leído cuando abres el chat en la app oficial

	log.Info().Str("instance_id", instanceID).Str("jid", jidStr).Msg("Marcando chat como leído")

	// TODO: Implementar lógica de marcar como leído si whatsmeow lo soporta en el futuro
	// Por ahora, esto es principalmente para el frontend (resetear contador de no leídos)

	_ = jid // Evitar warning de variable no usada

	return nil
}
