package whatsapp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow/appstate"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	"kero-kero/internal/models"
	"kero-kero/internal/repository"
)

// Manager gestiona múltiples instancias de WhatsApp
type Manager struct {
	Container     *sqlstore.Container
	Clients       map[string]*Client
	mu            sync.RWMutex
	instanceRepo  *repository.InstanceRepository
	msgRepo       *repository.MessageRepository
	redisClient   *repository.RedisClient
	webhookSvc    WebhookServiceInterface
	wsService     WebSocketServiceInterface
	automationSvc AutomationServiceInterface
}

// WebhookServiceInterface interfaz para evitar dependencia circular
type WebhookServiceInterface interface {
	SendEvent(ctx context.Context, instanceID string, event *models.WebhookEvent) error
}

// WebSocketServiceInterface interfaz para evitar dependencia circular
type WebSocketServiceInterface interface {
	BroadcastEvent(eventType string, payload interface{})
}

// AutomationServiceInterface interfaz para evitar dependencia circular
type AutomationServiceInterface interface {
	GetAutoReply(ctx context.Context, instanceID string) (*models.AutoReplyConfig, error)
}

// NewManager crea un nuevo gestor de instancias
func NewManager(
	container *sqlstore.Container,
	instanceRepo *repository.InstanceRepository,
	msgRepo *repository.MessageRepository,
	redisClient *repository.RedisClient,
) *Manager {
	return &Manager{
		Container:    container,
		Clients:      make(map[string]*Client),
		instanceRepo: instanceRepo,
		msgRepo:      msgRepo,
		redisClient:  redisClient,
	}
}

// SetWebhookService configura el servicio de webhooks
func (m *Manager) SetWebhookService(svc WebhookServiceInterface) {
	m.webhookSvc = svc
}

// SetWebSocketService configura el servicio de WebSocket
func (m *Manager) SetWebSocketService(svc WebSocketServiceInterface) {
	m.wsService = svc
}

// SetAutomationService configura el servicio de automatización
func (m *Manager) SetAutomationService(svc AutomationServiceInterface) {
	m.automationSvc = svc
}

// LoadInstances carga las instancias existentes desde la base de datos
func (m *Manager) LoadInstances(ctx context.Context) error {
	log.Info().Msg("Cargando instancias existentes")

	instances, err := m.instanceRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("error cargando instancias: %w", err)
	}

	for _, instance := range instances {
		if instance.JID == "" {
			log.Debug().Str("instance_id", instance.InstanceID).Msg("Instancia sin JID, omitiendo")
			continue
		}

		jid, err := types.ParseJID(instance.JID)
		if err != nil {
			log.Error().Err(err).Str("instance_id", instance.InstanceID).Msg("JID inválido")
			continue
		}

		device, err := m.Container.GetDevice(ctx, jid)
		if err != nil {
			log.Error().Err(err).Str("instance_id", instance.InstanceID).Msg("Error obteniendo dispositivo")
			continue
		}

		if device == nil {
			log.Warn().Str("instance_id", instance.InstanceID).Msg("Dispositivo no encontrado, reseteando estado")
			// No borrar la instancia, solo resetear su estado para permitir re-escaneo
			m.instanceRepo.UpdateStatus(ctx, instance.InstanceID, models.StatusDisconnected)
			m.instanceRepo.UpdateJID(ctx, instance.InstanceID, "")
			continue
		}

		client := NewClient(device, waLog.Stdout("Client-"+instance.InstanceID, "INFO", true))
		client.SyncHistory = instance.SyncHistory
		m.Clients[instance.InstanceID] = client

		// Registrar handler de eventos
		client.WAClient.AddEventHandler(func(evt interface{}) {
			m.handleEvent(ctx, instance.InstanceID, evt)
		})

		log.Info().Str("instance_id", instance.InstanceID).Msg("Instancia cargada")

		// Reconectar automáticamente si estaba conectada
		if instance.Status == string(models.StatusConnected) || instance.Status == string(models.StatusAuthenticated) {
			log.Info().Str("instance_id", instance.InstanceID).Msg("Reconectando instancia automáticamente")

			// Actualizar estado a "connecting"
			bgCtx := context.Background()
			m.instanceRepo.UpdateStatus(bgCtx, instance.InstanceID, models.StatusConnecting)

			// Conectar en goroutine para no bloquear el inicio
			go func(instanceID string, c *Client) {
				if err := c.Connect(); err != nil {
					log.Error().Err(err).Str("instance_id", instanceID).Msg("Error reconectando instancia")
					m.instanceRepo.UpdateStatus(context.Background(), instanceID, models.StatusDisconnected)
				} else {
					log.Info().Str("instance_id", instanceID).Msg("Instancia reconectada exitosamente")
				}
			}(instance.InstanceID, client)
		} else {
			// Si no estaba conectada, asegurar que el estado sea "disconnected"
			bgCtx := context.Background()
			m.instanceRepo.UpdateStatus(bgCtx, instance.InstanceID, models.StatusDisconnected)
		}
	}

	log.Info().Int("count", len(m.Clients)).Msg("Instancias cargadas exitosamente")
	return nil
}

// CreateInstance crea una nueva instancia
func (m *Manager) CreateInstance(ctx context.Context, instanceID string, syncHistory bool) (*Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verificar si ya existe
	if _, exists := m.Clients[instanceID]; exists {
		return nil, fmt.Errorf("instancia ya existe")
	}

	// Crear en base de datos
	instance := &models.Instance{
		InstanceID:  instanceID,
		Status:      string(models.StatusDisconnected),
		SyncHistory: syncHistory,
	}

	if err := m.instanceRepo.Create(ctx, instance); err != nil {
		return nil, fmt.Errorf("error creando instancia en DB: %w", err)
	}

	// Crear dispositivo de WhatsApp
	device := m.Container.NewDevice()
	client := NewClient(device, waLog.Stdout("Client-"+instanceID, "INFO", true))
	client.SyncHistory = syncHistory

	m.Clients[instanceID] = client

	// Registrar handler de eventos
	client.WAClient.AddEventHandler(func(evt interface{}) {
		m.handleEvent(ctx, instanceID, evt)
	})

	log.Info().Str("instance_id", instanceID).Msg("Instancia creada")
	return client, nil
}

// GetClient obtiene un cliente por ID
func (m *Manager) GetClient(instanceID string) *Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Clients[instanceID]
}

// GetOrCreateClient obtiene un cliente existente o crea uno nuevo (sin tocar la BD)
// Este método es útil cuando la instancia ya existe en BD pero no tiene cliente en el Manager
func (m *Manager) GetOrCreateClient(ctx context.Context, instanceID string) (*Client, error) {
	// Primero intentar obtener el cliente existente (sin lock de escritura)
	m.mu.RLock()
	if client, exists := m.Clients[instanceID]; exists {
		m.mu.RUnlock()
		return client, nil
	}
	m.mu.RUnlock()

	// Si no existe, crear uno nuevo (con lock de escritura)
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verificar de nuevo por si otro goroutine lo creó mientras esperábamos el lock
	if client, exists := m.Clients[instanceID]; exists {
		return client, nil
	}

	// Crear dispositivo de WhatsApp
	device := m.Container.NewDevice()
	client := NewClient(device, waLog.Stdout("Client-"+instanceID, "INFO", true))

	m.Clients[instanceID] = client

	// Registrar handler de eventos
	client.WAClient.AddEventHandler(func(evt interface{}) {
		m.handleEvent(ctx, instanceID, evt)
	})

	log.Info().Str("instance_id", instanceID).Msg("Cliente creado en Manager")
	return client, nil
}

// DeleteInstance elimina una instancia
func (m *Manager) DeleteInstance(ctx context.Context, instanceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.Clients[instanceID]
	if exists {
		client.Disconnect()
		delete(m.Clients, instanceID)
	}

	// Eliminar de base de datos
	if err := m.instanceRepo.Delete(ctx, instanceID); err != nil {
		return fmt.Errorf("error eliminando instancia: %w", err)
	}

	// Limpiar Redis
	m.redisClient.DeleteQRCode(ctx, instanceID)
	m.redisClient.DeleteSession(ctx, instanceID)

	log.Info().Str("instance_id", instanceID).Msg("Instancia eliminada")
	return nil
}

// handleEvent maneja eventos de WhatsApp
func (m *Manager) handleEvent(ctx context.Context, instanceID string, evt interface{}) {
	switch v := evt.(type) {
	case *events.QR:
		// Guardar QR en Redis
		if len(v.Codes) > 0 {
			qrCode := v.Codes[0]
			// Usar context.Background() porque esta operación debe completarse
			// independientemente del contexto de la request original
			bgCtx := context.Background()
			if err := m.redisClient.SetQRCode(bgCtx, instanceID, qrCode); err != nil {
				log.Error().Err(err).Str("instance_id", instanceID).Msg("Error guardando QR en Redis")
			} else {
				log.Debug().Str("instance_id", instanceID).Msg("QR code guardado en Redis exitosamente")
			}

			if m.wsService != nil {
				m.wsService.BroadcastEvent("qr", map[string]interface{}{
					"instance_id": instanceID,
					"qr":          qrCode,
				})
			}
		}

	case *events.PairSuccess:
		// Usar context.Background() para operaciones de BD
		bgCtx := context.Background()

		// Actualizar JID en base de datos
		if err := m.instanceRepo.UpdateJID(bgCtx, instanceID, v.ID.String()); err != nil {
			log.Error().Err(err).Str("instance_id", instanceID).Msg("Error actualizando JID")
		}

		// Actualizar estado en DB
		if err := m.instanceRepo.UpdateStatus(bgCtx, instanceID, models.StatusAuthenticated); err != nil {
			log.Error().Err(err).Str("instance_id", instanceID).Msg("Error actualizando estado")
		}
		// Sincronizar estado en Redis
		m.redisClient.SetInstanceStatus(bgCtx, instanceID, string(models.StatusAuthenticated))

		// Actualizar última conexión
		if err := m.instanceRepo.UpdateLastConnected(bgCtx, instanceID); err != nil {
			log.Error().Err(err).Str("instance_id", instanceID).Msg("Error actualizando última conexión")
		}

		// Limpiar QR de Redis
		m.redisClient.DeleteQRCode(bgCtx, instanceID)

		log.Info().
			Str("instance_id", instanceID).
			Str("jid", v.ID.String()).
			Msg("Instancia autenticada exitosamente")

	case *events.Connected:
		bgCtx := context.Background()
		if err := m.instanceRepo.UpdateStatus(bgCtx, instanceID, models.StatusConnected); err != nil {
			log.Error().Err(err).Str("instance_id", instanceID).Msg("Error actualizando estado en DB")
		}
		// Sincronizar estado en Redis
		if err := m.redisClient.SetInstanceStatus(bgCtx, instanceID, string(models.StatusConnected)); err != nil {
			log.Error().Err(err).Str("instance_id", instanceID).Msg("Error actualizando estado en Redis")
		}
		log.Info().Str("instance_id", instanceID).Msg("Instancia conectada")

		// Enviar webhook de estado
		if m.webhookSvc != nil {
			m.webhookSvc.SendEvent(ctx, instanceID, &models.WebhookEvent{
				Event: "status",
				Data: models.StatusEvent{
					Status: "connected",
				},
			})
		}

		if m.wsService != nil {
			m.wsService.BroadcastEvent("status", map[string]interface{}{
				"instance_id": instanceID,
				"status":      "connected",
			})
		}

	case *events.Disconnected:
		bgCtx := context.Background()
		if err := m.instanceRepo.UpdateStatus(bgCtx, instanceID, models.StatusDisconnected); err != nil {
			log.Error().Err(err).Str("instance_id", instanceID).Msg("Error actualizando estado en DB")
		}
		// Sincronizar estado en Redis
		m.redisClient.SetInstanceStatus(bgCtx, instanceID, string(models.StatusDisconnected))
		log.Info().Str("instance_id", instanceID).Msg("Instancia desconectada")

		// Enviar webhook de estado
		if m.webhookSvc != nil {
			m.webhookSvc.SendEvent(ctx, instanceID, &models.WebhookEvent{
				Event: "status",
				Data: models.StatusEvent{
					Status: "disconnected",
				},
			})
		}

		if m.wsService != nil {
			m.wsService.BroadcastEvent("status", map[string]interface{}{
				"instance_id": instanceID,
				"status":      "disconnected",
			})
		}

	case *events.LoggedOut:
		if err := m.instanceRepo.UpdateStatus(ctx, instanceID, models.StatusDisconnected); err != nil {
			log.Error().Err(err).Str("instance_id", instanceID).Msg("Error actualizando estado")
		}
		// Sincronizar estado en Redis
		m.redisClient.SetInstanceStatus(ctx, instanceID, string(models.StatusDisconnected))
		log.Info().Str("instance_id", instanceID).Msg("Instancia cerró sesión")

		// Enviar webhook de estado
		if m.webhookSvc != nil {
			m.webhookSvc.SendEvent(ctx, instanceID, &models.WebhookEvent{
				Event: "status",
				Data: models.StatusEvent{
					Status: "logged_out",
				},
			})
		}

	case *events.HistorySync:
		// Verificar configuración de sincronización
		client := m.GetClient(instanceID)
		if client == nil || !client.SyncHistory {
			log.Info().Str("instance_id", instanceID).Msg("Omitiendo sincronización de historial (desactivada)")
			return
		}

		// Procesar historial en una goroutine para no bloquear
		go func() {
			bgCtx := context.Background()
			log.Info().Str("instance_id", instanceID).Msg("Procesando sincronización de historial")

			// Enviar primer webhook de progreso
			if m.webhookSvc != nil && v.Data.Progress != nil {
				syncType := "unknown"
				if v.Data.SyncType != nil {
					syncType = v.Data.SyncType.String()
				}
				m.webhookSvc.SendEvent(bgCtx, instanceID, &models.WebhookEvent{
					Event: "sync_progress",
					Data: models.SyncEvent{
						Percentage: int(*v.Data.Progress),
						SyncType:   strings.ToLower(syncType),
					},
				})
			}

			count := 0
			for _, conv := range v.Data.GetConversations() {

				for _, historyMsg := range conv.GetMessages() {
					webMsgInfo := historyMsg.GetMessage()
					if webMsgInfo == nil {
						continue
					}

					msgContent := webMsgInfo.GetMessage()
					if msgContent == nil {
						continue
					}

					// Determinar JID del chat
					chatJID, _ := types.ParseJID(conv.GetID())
					chatJID = m.GetClient(instanceID).ResolveJID(chatJID)

					// Obtener Key
					msgKey := webMsgInfo.GetKey()

					// Determinar Sender
					senderJID := chatJID // Por defecto en chats privados
					if chatJID.Server == "g.us" {
						if msgKey != nil && msgKey.Participant != nil {
							senderJID, _ = types.ParseJID(*msgKey.Participant)
							senderJID = m.GetClient(instanceID).ResolveJID(senderJID)
						}
					} else if msgKey != nil && msgKey.FromMe != nil && *msgKey.FromMe {
						// Si es enviado por mí
						senderJID = m.GetClient(instanceID).WAClient.Store.ID.ToNonAD()
					} else {
						// Si es chat privado y no soy yo, el sender es el chatJID (ya resuelto)
						senderJID = chatJID
					}

					// Determinar contenido y tipo
					var content string
					var msgType string = "unknown"

					if msgContent.Conversation != nil {
						msgType = "text"
						content = *msgContent.Conversation
					} else if msgContent.ExtendedTextMessage != nil {
						msgType = "text"
						content = *msgContent.ExtendedTextMessage.Text
					} else if msgContent.ImageMessage != nil {
						msgType = "image"
						content = "[Imagen]"
						if msgContent.ImageMessage.Caption != nil {
							content = *msgContent.ImageMessage.Caption
						}
					} else if msgContent.VideoMessage != nil {
						msgType = "video"
						content = "[Video]"
						if msgContent.VideoMessage.Caption != nil {
							content = *msgContent.VideoMessage.Caption
						}
					} else if msgContent.AudioMessage != nil {
						msgType = "audio"
						content = "[Audio]"
					} else if msgContent.DocumentMessage != nil {
						msgType = "document"
						content = "[Documento]"
					} else if msgContent.LocationMessage != nil {
						msgType = "location"
						content = "[Ubicación]"
					}

					// Timestamp
					ts := int64(0)
					if webMsgInfo.MessageTimestamp != nil {
						ts = int64(*webMsgInfo.MessageTimestamp)
					} else {
						ts = time.Now().Unix()
					}

					// ID
					id := ""
					if msgKey != nil && msgKey.ID != nil {
						id = *msgKey.ID
					}

					// IsFromMe
					isFromMe := false
					if msgKey != nil && msgKey.FromMe != nil {
						isFromMe = *msgKey.FromMe
					}

					msg := &models.Message{
						ID:         id,
						InstanceID: instanceID,
						From:       senderJID.String(),
						To:         chatJID.String(),
						Sender:     senderJID.String(),
						Timestamp:  ts,
						Type:       msgType,
						Content:    content,
						PushName:   webMsgInfo.GetPushName(),
						IsFromMe:   isFromMe,
						Status:     "history", // Estado especial para historial
					}

					if err := m.msgRepo.Create(bgCtx, msg); err == nil {
						count++
					}
				}
			}
			log.Info().Str("instance_id", instanceID).Int("msgs_synced", count).Msg("Sincronización de historial completada")
		}()

	case *events.Message:
		// Guardar mensaje en base de datos
		bgCtx := context.Background()

		// Determinar contenido y tipo
		var content string
		var msgType string = "unknown"

		if v.Message.Conversation != nil {
			msgType = "text"
			content = *v.Message.Conversation
		} else if v.Message.ExtendedTextMessage != nil {
			msgType = "text"
			content = *v.Message.ExtendedTextMessage.Text
		} else if v.Message.ImageMessage != nil {
			msgType = "image"
			if v.Message.ImageMessage.Caption != nil {
				content = *v.Message.ImageMessage.Caption
			} else {
				content = "[Imagen]"
			}
		} else if v.Message.VideoMessage != nil {
			msgType = "video"
			if v.Message.VideoMessage.Caption != nil {
				content = *v.Message.VideoMessage.Caption
			} else {
				content = "[Video]"
			}
		} else if v.Message.AudioMessage != nil {
			msgType = "audio"
			content = "[Audio]"
		} else if v.Message.DocumentMessage != nil {
			msgType = "document"
			content = "[Documento]"
		} else if v.Message.LocationMessage != nil {
			msgType = "location"
			content = "[Ubicación]"
		}

		// Resolver JIDs para evitar duplicados por LID
		client := m.GetClient(instanceID)
		chatJID := v.Info.Chat
		senderJID := v.Info.Sender
		if client != nil {
			chatJID = client.ResolveJID(v.Info.Chat)
			senderJID = client.ResolveJID(v.Info.Sender)
		}

		msg := &models.Message{
			ID:         v.Info.ID,
			InstanceID: instanceID,
			From:       senderJID.String(),
			To:         chatJID.String(),
			Sender:     senderJID.String(),
			Timestamp:  v.Info.Timestamp.Unix(),
			Type:       msgType,
			Content:    content,
			PushName:   v.Info.PushName,
			IsFromMe:   v.Info.IsFromMe,
			Status:     "received",
		}

		if err := m.msgRepo.Create(bgCtx, msg); err != nil {
			log.Error().Err(err).Str("instance_id", instanceID).Msg("Error guardando mensaje entrante en DB")
		}

		// Enviar webhook de mensaje recibido
		if m.webhookSvc != nil {
			messageEvent := models.MessageEvent{
				MessageID:   v.Info.ID,
				From:        senderJID.String(),
				To:          chatJID.String(),
				IsGroup:     v.Info.IsGroup,
				MessageType: msgType,
				Text:        content,
				Timestamp:   v.Info.Timestamp.Unix(),
				IsFromMe:    v.Info.IsFromMe,
				SenderName:  v.Info.PushName,
			}

			// Intentar obtener el nombre del chat (especialmente si es grupo o contacto guardado)
			if client != nil {
				contact, err := client.WAClient.Store.Contacts.GetContact(bgCtx, chatJID)
				if err == nil && (contact.FullName != "" || contact.BusinessName != "") {
					if contact.BusinessName != "" {
						messageEvent.ChatName = contact.BusinessName
					} else {
						messageEvent.ChatName = contact.FullName
					}
				}
			}

			// Extraer caption si existe
			if v.Message.ImageMessage != nil && v.Message.ImageMessage.Caption != nil {
				messageEvent.Caption = *v.Message.ImageMessage.Caption
			} else if v.Message.VideoMessage != nil && v.Message.VideoMessage.Caption != nil {
				messageEvent.Caption = *v.Message.VideoMessage.Caption
			} else if v.Message.DocumentMessage != nil && v.Message.DocumentMessage.Caption != nil {
				messageEvent.Caption = *v.Message.DocumentMessage.Caption
			}

			// Extraer ubicación si existe
			if v.Message.LocationMessage != nil {
				if v.Message.LocationMessage.DegreesLatitude != nil {
					lat := *v.Message.LocationMessage.DegreesLatitude
					messageEvent.Latitude = fmt.Sprintf("%.6f", lat)
				}
				if v.Message.LocationMessage.DegreesLongitude != nil {
					lon := *v.Message.LocationMessage.DegreesLongitude
					messageEvent.Longitude = fmt.Sprintf("%.6f", lon)
				}
				if v.Message.LocationMessage.Name != nil {
					messageEvent.LocationName = *v.Message.LocationMessage.Name
				}
				if v.Message.LocationMessage.Address != nil {
					messageEvent.LocationAddress = *v.Message.LocationMessage.Address
				}
			}

			// Descargar y adjuntar medios (imágenes, videos, audios, documentos)
			// Límites según tipo de archivo para envío inline en webhook
			var maxMediaSize uint64
			shouldDownloadMedia := false
			var mediaSize uint64 = 0

			if v.Message.ImageMessage != nil && v.Message.ImageMessage.FileLength != nil {
				mediaSize = *v.Message.ImageMessage.FileLength
				maxMediaSize = 16 * 1024 * 1024 // 16MB para imágenes
				shouldDownloadMedia = mediaSize > 0 && mediaSize < maxMediaSize
				messageEvent.FileName = "image.jpg"
				if v.Message.ImageMessage.Mimetype != nil {
					messageEvent.MimeType = *v.Message.ImageMessage.Mimetype
				}
			} else if v.Message.VideoMessage != nil && v.Message.VideoMessage.FileLength != nil {
				mediaSize = *v.Message.VideoMessage.FileLength
				maxMediaSize = 16 * 1024 * 1024 // 16MB para videos
				shouldDownloadMedia = mediaSize > 0 && mediaSize < maxMediaSize
				messageEvent.FileName = "video.mp4"
				if v.Message.VideoMessage.Mimetype != nil {
					messageEvent.MimeType = *v.Message.VideoMessage.Mimetype
				}
			} else if v.Message.AudioMessage != nil && v.Message.AudioMessage.FileLength != nil {
				mediaSize = *v.Message.AudioMessage.FileLength
				maxMediaSize = 16 * 1024 * 1024 // 16MB para audios
				shouldDownloadMedia = mediaSize > 0 && mediaSize < maxMediaSize
				messageEvent.FileName = "audio.ogg"
				if v.Message.AudioMessage.Mimetype != nil {
					messageEvent.MimeType = *v.Message.AudioMessage.Mimetype
				}
			} else if v.Message.DocumentMessage != nil && v.Message.DocumentMessage.FileLength != nil {
				mediaSize = *v.Message.DocumentMessage.FileLength
				maxMediaSize = 50 * 1024 * 1024 // 50MB para documentos
				shouldDownloadMedia = mediaSize > 0 && mediaSize < maxMediaSize
				if v.Message.DocumentMessage.FileName != nil {
					messageEvent.FileName = *v.Message.DocumentMessage.FileName
				}
				if v.Message.DocumentMessage.Mimetype != nil {
					messageEvent.MimeType = *v.Message.DocumentMessage.Mimetype
				}
			}

			messageEvent.FileSize = int64(mediaSize)

			// Descargar media si cumple los requisitos
			if shouldDownloadMedia && client != nil {
				var mediaData []byte
				var err error

				if v.Message.ImageMessage != nil {
					mediaData, err = client.WAClient.Download(bgCtx, v.Message.ImageMessage)
				} else if v.Message.VideoMessage != nil {
					mediaData, err = client.WAClient.Download(bgCtx, v.Message.VideoMessage)
				} else if v.Message.AudioMessage != nil {
					mediaData, err = client.WAClient.Download(bgCtx, v.Message.AudioMessage)
				} else if v.Message.DocumentMessage != nil {
					mediaData, err = client.WAClient.Download(bgCtx, v.Message.DocumentMessage)
				}

				if err != nil {
					log.Error().Err(err).Str("instance_id", instanceID).Str("msg_id", v.Info.ID).Msg("Error descargando media")
					messageEvent.MediaError = "Error descargando archivo: " + err.Error()
					// Ofrecer URL de descarga bajo demanda como alternativa
					messageEvent.MediaURL = fmt.Sprintf("/instances/%s/messages/%s/media", instanceID, v.Info.ID)
				} else if mediaData != nil {
					// Convertir a base64
					messageEvent.MediaData = base64.StdEncoding.EncodeToString(mediaData)
					log.Debug().
						Str("instance_id", instanceID).
						Str("msg_id", v.Info.ID).
						Int("size", len(mediaData)).
						Msg("Media descargado y adjuntado al webhook")
				}
			} else if mediaSize >= maxMediaSize {
				// Archivo muy grande para inline, ofrecer URL de descarga bajo demanda
				messageEvent.MediaURL = fmt.Sprintf("/instances/%s/messages/%s/media", instanceID, v.Info.ID)
				messageEvent.MediaError = fmt.Sprintf("Archivo muy grande para envío inline (%d bytes). Usa media_url para descarga bajo demanda", mediaSize)
				log.Debug().
					Str("instance_id", instanceID).
					Str("msg_id", v.Info.ID).
					Uint64("size", mediaSize).
					Msg("Media no descargado: excede tamaño máximo para inline, disponible vía URL")
			}

			m.webhookSvc.SendEvent(ctx, instanceID, &models.WebhookEvent{
				Event: "message",
				Data:  messageEvent,
			})

			// Lógica de Autolabeling (Etiquetas automáticas)
			if content != "" && !v.Info.IsFromMe {
				go m.handleAutoLabeling(instanceID, v.Info.Chat, content)
			}

			// Auto Reply Logic
			if !v.Info.IsFromMe && m.automationSvc != nil {
				go func() {
					config, err := m.automationSvc.GetAutoReply(bgCtx, instanceID)
					if err != nil || !config.Enabled || config.Message == "" {
						return
					}

					shouldReply := false
					if len(config.TriggerKeywords) == 0 {
						shouldReply = true
					} else {
						msgText := strings.ToLower(content)
						for _, keyword := range config.TriggerKeywords {
							keyword = strings.ToLower(keyword)
							switch config.MatchType {
							case "exact":
								if msgText == keyword {
									shouldReply = true
								}
							case "startswith":
								if strings.HasPrefix(msgText, keyword) {
									shouldReply = true
								}
							default: // contains
								if strings.Contains(msgText, keyword) {
									shouldReply = true
								}
							}
							if shouldReply {
								break
							}
						}
					}

					if shouldReply {
						// Simular un pequeño delay humano
						time.Sleep(2 * time.Second)

						// Enviar respuesta
						_, err := client.WAClient.SendMessage(bgCtx, chatJID, &waProto.Message{
							Conversation: proto.String(config.Message),
						})
						if err != nil {
							log.Error().Err(err).Str("instance_id", instanceID).Msg("Error enviando auto-reply")
						} else {
							log.Info().Str("instance_id", instanceID).Str("to", chatJID.String()).Msg("Auto-reply enviado")
						}
					}
				}()
			}

			if m.wsService != nil {
				m.wsService.BroadcastEvent("message", map[string]interface{}{
					"instance_id": instanceID,
					"data":        messageEvent,
				})
			}
		}

	case *events.Receipt:
		// Enviar webhook de confirmación de lectura/entrega
		if m.webhookSvc != nil {
			m.webhookSvc.SendEvent(ctx, instanceID, &models.WebhookEvent{
				Event: "receipt",
				Data: models.ReceiptEvent{
					From:      v.Sender.String(),
					Type:      string(v.Type),
					Timestamp: v.Timestamp.Unix(),
				},
			})
		}

	case *events.ChatPresence:
		// Enviar evento de presencia (escribiendo, grabando audio, etc.)
		if m.wsService != nil {
			presenceType := "available"
			if v.State == types.ChatPresenceComposing {
				presenceType = "composing"
			} else if v.State == types.ChatPresencePaused {
				presenceType = "paused"
			}

			m.wsService.BroadcastEvent("presence", map[string]interface{}{
				"instance_id": instanceID,
				"data": map[string]interface{}{
					"from": v.MessageSource.Chat.String(),
					"type": presenceType,
				},
			})
		}

		// Enviar webhook de presencia
		if m.webhookSvc != nil {
			m.webhookSvc.SendEvent(ctx, instanceID, &models.WebhookEvent{
				Event: "presence",
				Data: map[string]interface{}{
					"from":  v.MessageSource.Chat.String(),
					"state": v.State,
					"media": v.Media,
				},
			})
		}

	case *events.CallOffer:
		// Lógica de rechazo automático inteligente.
		// He implementado esto en una goroutine para no bloquear el procesamiento de otros eventos.
		go func() {
			bgCtx := context.Background()
			data, err := m.redisClient.GetCallSettings(bgCtx, instanceID)
			if err != nil {
				return // Si no hay configuración en Redis, ignoramos el auto-rechazo.
			}

			var settings models.CallSettings
			if err := json.Unmarshal([]byte(data), &settings); err != nil {
				log.Error().Err(err).Msg("Error al decodificar settings de llamadas en el handler")
				return
			}

			if settings.AutoReject {
				// Aplicamos el delay de rechazo si se ha configurado
				if settings.RejectDelay > 0 {
					log.Debug().
						Str("instance_id", instanceID).
						Int("delay_seconds", settings.RejectDelay).
						Msg("Esperando antes de rechazar la llamada")
					time.Sleep(time.Duration(settings.RejectDelay) * time.Second)
				}

				log.Warn().
					Str("instance_id", instanceID).
					Str("from", v.From.String()).
					Str("call_id", v.CallID).
					Msg("Llamada detectada: Ejecutando rechazo automático")

				client := m.GetClient(instanceID)
				if client != nil {
					// Rechazamos la llamada usando el cliente WA y el contexto.
					client.WAClient.RejectCall(bgCtx, v.From, v.CallID)

					// Si el usuario activó la respuesta inteligente, enviamos el mensaje.
					if settings.AutoReplyEnabled && settings.AutoReplyMessage != "" {
						// Un pequeño delay para que parezca que el usuario colgó y luego escribió.
						time.Sleep(1500 * time.Millisecond)

						replyMsg := &waProto.Message{
							Conversation: proto.String(settings.AutoReplyMessage),
						}
						_, err := client.WAClient.SendMessage(bgCtx, v.From, replyMsg)
						if err != nil {
							log.Error().Err(err).Msg("Error enviando mensaje de auto-rechazo")
						}
					}
				}
			}

			// Notificamos vía Webhook para que el CRM del cliente sepa qué pasó.
			if m.webhookSvc != nil {
				status := "incoming"
				if settings.AutoReject {
					status = "rejected"
				}

				m.webhookSvc.SendEvent(bgCtx, instanceID, &models.WebhookEvent{
					Event: "call",
					Data: models.CallEvent{
						From:      v.From.String(),
						Timestamp: time.Now().Unix(),
						IsVideo:   false, // No disponible directamente en CallOffer básico
						Status:    status,
					},
				})
			}
		}()

	}
}

// handleAutoLabeling procesa el contenido de un mensaje y aplica etiquetas si coincide con las reglas.
func (m *Manager) handleAutoLabeling(instanceID string, chatJID types.JID, content string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rulesJSON, err := m.redisClient.GetAutoLabelRules(ctx, instanceID)
	if err != nil || rulesJSON == "" {
		return
	}

	var rules []models.AutoLabelRule
	if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
		log.Error().Err(err).Msg("Error decodificando reglas de autolabeling")
		return
	}

	contentLower := strings.ToLower(content)
	for _, rule := range rules {
		matched := false
		for _, keyword := range rule.Keywords {
			if strings.Contains(contentLower, strings.ToLower(keyword)) {
				matched = true
				break
			}
		}

		if matched {
			log.Info().
				Str("instance_id", instanceID).
				Str("label_id", rule.LabelID).
				Str("chat", chatJID.String()).
				Msg("Auto-labeling: Regla coincidente detectada")

			client := m.GetClient(instanceID)
			if client != nil {
				// whatsmeow usa BuildLabelChat para crear el patch de sincronización.
				// El orden correcto es: (JID del chat, ID de la etiqueta, si se añade o quita)
				patch := appstate.BuildLabelChat(chatJID, rule.LabelID, true)
				err := client.WAClient.SendAppState(ctx, patch)
				if err != nil {
					log.Error().Err(err).Msg("Error aplicando etiqueta automática vía App State")
				}
			}

			// Una vez que aplicamos una etiqueta, podríamos querer detenernos o seguir con otras reglas.
			// Por ahora seguimos, permitiendo múltiples etiquetas.
		}
	}
}

// Close cierra el gestor y desconecta todos los clientes
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, client := range m.Clients {
		client.Disconnect()
	}
	// No cerramos el container aquí porque se gestiona externamente o no tiene Close explícito necesario
}
