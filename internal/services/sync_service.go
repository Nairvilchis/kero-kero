package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow/types"

	"kero-kero/internal/repository"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/errors"
)

type SyncService struct {
	waManager *whatsapp.Manager
	msgRepo   *repository.MessageRepository
	chatSvc   *ChatService
	mu        sync.Mutex
	progress  map[string]*SyncProgress
}

type SyncProgress struct {
	InstanceID     string     `json:"instance_id"`
	Status         string     `json:"status"` // running, completed, failed
	TotalChats     int        `json:"total_chats"`
	ProcessedChats int        `json:"processed_chats"`
	TotalMessages  int        `json:"total_messages"`
	SyncedMessages int        `json:"synced_messages"`
	CurrentChat    string     `json:"current_chat"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	Error          string     `json:"error,omitempty"`
}

type SyncOptions struct {
	MessagesPerChat int  `json:"messages_per_chat"` // Cantidad de mensajes por chat (default: 50)
	MaxChats        int  `json:"max_chats"`         // Máximo de chats a sincronizar (0 = todos)
	Advanced        bool `json:"advanced"`          // Modo avanzado
}

func NewSyncService(waManager *whatsapp.Manager, msgRepo *repository.MessageRepository, chatSvc *ChatService) *SyncService {
	return &SyncService{
		waManager: waManager,
		msgRepo:   msgRepo,
		chatSvc:   chatSvc,
		progress:  make(map[string]*SyncProgress),
	}
}

// SyncChatHistory sincroniza el historial de mensajes de una instancia
func (s *SyncService) SyncChatHistory(ctx context.Context, instanceID string, opts SyncOptions) error {
	client := s.waManager.GetClient(instanceID)
	if client == nil {
		return errors.ErrInstanceNotFound
	}
	if !client.WAClient.IsLoggedIn() {
		return errors.ErrNotAuthenticated
	}

	// Verificar si ya hay una sincronización en progreso
	s.mu.Lock()
	if prog, exists := s.progress[instanceID]; exists && prog.Status == "running" {
		s.mu.Unlock()
		return errors.New(409, "Ya hay una sincronización en progreso para esta instancia")
	}

	// Inicializar progreso
	progress := &SyncProgress{
		InstanceID: instanceID,
		Status:     "running",
		StartedAt:  time.Now(),
	}
	s.progress[instanceID] = progress
	s.mu.Unlock()

	// Ejecutar sincronización en goroutine
	go s.performSync(context.Background(), instanceID, opts, progress)

	return nil
}

// performSync realiza la sincronización en background
func (s *SyncService) performSync(ctx context.Context, instanceID string, opts SyncOptions, progress *SyncProgress) {
	defer func() {
		now := time.Now()
		progress.CompletedAt = &now
		if progress.Status == "running" {
			progress.Status = "completed"
		}
	}()

	client := s.waManager.GetClient(instanceID)
	if client == nil {
		progress.Status = "failed"
		progress.Error = "Cliente no encontrado"
		return
	}

	// Configurar opciones por defecto
	if opts.MessagesPerChat == 0 {
		opts.MessagesPerChat = 50
	}
	if opts.MaxChats == 0 && !opts.Advanced {
		opts.MaxChats = 20 // Modo básico: solo 20 chats más recientes
	}

	log.Info().
		Str("instance_id", instanceID).
		Int("messages_per_chat", opts.MessagesPerChat).
		Int("max_chats", opts.MaxChats).
		Bool("advanced", opts.Advanced).
		Msg("Iniciando sincronización de historial")

	// Obtener lista de contactos/chats
	contacts, err := client.WAClient.Store.Contacts.GetAllContacts(ctx)
	if err != nil {
		progress.Status = "failed"
		progress.Error = fmt.Sprintf("Error obteniendo contactos: %v", err)
		log.Error().Err(err).Msg("Error obteniendo contactos para sincronización")
		return
	}

	// Convertir a slice para poder limitar
	var chatJIDs []types.JID
	for jid, contact := range contacts {
		if contact.Found {
			chatJIDs = append(chatJIDs, jid)
		}
	}

	// Limitar cantidad de chats si está configurado
	if opts.MaxChats > 0 && len(chatJIDs) > opts.MaxChats {
		chatJIDs = chatJIDs[:opts.MaxChats]
	}

	progress.TotalChats = len(chatJIDs)

	// Sincronizar cada chat
	for _, jid := range chatJIDs {
		progress.CurrentChat = jid.String()

		log.Debug().
			Str("instance_id", instanceID).
			Str("jid", jid.String()).
			Msg("Sincronizando chat")

		// Intentar obtener historial usando whatsmeow
		// Nota: whatsmeow no tiene un método directo GetMessageHistory en todas las versiones
		// Vamos a usar un enfoque alternativo: obtener mensajes del store si están disponibles

		// Por ahora, simulamos que no hay historial disponible desde whatsmeow
		// En una implementación real, necesitarías acceder al store de whatsmeow directamente
		// o usar métodos específicos de la versión de whatsmeow que estés usando

		log.Warn().
			Str("jid", jid.String()).
			Msg("Sincronización de historial desde WhatsApp no implementada - whatsmeow no expone API pública para esto")

		progress.ProcessedChats++
	}

	log.Info().
		Str("instance_id", instanceID).
		Int("chats_processed", progress.ProcessedChats).
		Int("messages_synced", progress.SyncedMessages).
		Msg("Sincronización completada")
}

// GetSyncProgress obtiene el progreso de sincronización de una instancia
func (s *SyncService) GetSyncProgress(instanceID string) (*SyncProgress, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	progress, exists := s.progress[instanceID]
	if !exists {
		return nil, errors.New(404, "No se encontró información de sincronización para esta instancia")
	}

	return progress, nil
}

// CancelSync cancela una sincronización en progreso
func (s *SyncService) CancelSync(instanceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	progress, exists := s.progress[instanceID]
	if !exists {
		return errors.New(404, "No se encontró sincronización para esta instancia")
	}

	if progress.Status != "running" {
		return errors.New(400, "La sincronización no está en progreso")
	}

	progress.Status = "cancelled"
	now := time.Now()
	progress.CompletedAt = &now

	return nil
}
