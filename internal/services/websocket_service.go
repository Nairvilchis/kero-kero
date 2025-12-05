package services

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// WebSocketMessage representa un mensaje enviado por WS
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// ClientRegistration representa un cliente registrándose a un room
type ClientRegistration struct {
	Conn       *websocket.Conn
	InstanceID string
}

// BroadcastMessage representa un mensaje a enviar a un room específico
type BroadcastMessage struct {
	InstanceID string
	Data       []byte
}

// WebSocketService maneja las conexiones en tiempo real con soporte de rooms
type WebSocketService struct {
	rooms      map[string]map[*websocket.Conn]bool // instanceID -> set of connections
	broadcast  chan BroadcastMessage
	register   chan ClientRegistration
	unregister chan ClientRegistration
	mutex      sync.RWMutex
	upgrader   websocket.Upgrader
}

func NewWebSocketService() *WebSocketService {
	return &WebSocketService{
		rooms:      make(map[string]map[*websocket.Conn]bool),
		broadcast:  make(chan BroadcastMessage),
		register:   make(chan ClientRegistration),
		unregister: make(chan ClientRegistration),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Permitir todas las conexiones por ahora (CORS)
			},
		},
	}
}

// Run inicia el bucle de eventos del WS
func (s *WebSocketService) Run() {
	for {
		select {
		case registration := <-s.register:
			s.mutex.Lock()
			if _, ok := s.rooms[registration.InstanceID]; !ok {
				s.rooms[registration.InstanceID] = make(map[*websocket.Conn]bool)
			}
			s.rooms[registration.InstanceID][registration.Conn] = true
			s.mutex.Unlock()
			log.Info().Str("instance_id", registration.InstanceID).Msg("Cliente WebSocket conectado a room")

		case registration := <-s.unregister:
			s.mutex.Lock()
			if room, ok := s.rooms[registration.InstanceID]; ok {
				if _, ok := room[registration.Conn]; ok {
					delete(room, registration.Conn)
					registration.Conn.Close()
					// Si el room está vacío, eliminarlo
					if len(room) == 0 {
						delete(s.rooms, registration.InstanceID)
					}
					log.Info().Str("instance_id", registration.InstanceID).Msg("Cliente WebSocket desconectado de room")
				}
			}
			s.mutex.Unlock()

		case message := <-s.broadcast:
			s.mutex.RLock()
			if room, ok := s.rooms[message.InstanceID]; ok {
				for client := range room {
					err := client.WriteMessage(websocket.TextMessage, message.Data)
					if err != nil {
						log.Error().Err(err).Str("instance_id", message.InstanceID).Msg("Error enviando mensaje WS")
						client.Close()
						delete(room, client)
					}
				}
			}
			s.mutex.RUnlock()
		}
	}
}

// HandleConnection maneja la solicitud HTTP de upgrade a WS para una instancia específica
func (s *WebSocketService) HandleConnection(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	if instanceID == "" {
		http.Error(w, "Instance ID required", http.StatusBadRequest)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Error al actualizar a WebSocket")
		return
	}

	registration := ClientRegistration{
		Conn:       conn,
		InstanceID: instanceID,
	}
	s.register <- registration

	// Mantener conexión viva y leer mensajes (ping/pong)
	go s.readPump(conn, instanceID)
}

func (s *WebSocketService) readPump(conn *websocket.Conn, instanceID string) {
	defer func() {
		s.unregister <- ClientRegistration{
			Conn:       conn,
			InstanceID: instanceID,
		}
	}()

	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Msg("Error leyendo mensaje WS")
			}
			break
		}
	}
}

// BroadcastEvent envía un evento a todos los clientes conectados a un room específico
func (s *WebSocketService) BroadcastEvent(eventType string, payload interface{}) {
	msg := WebSocketMessage{
		Type:    eventType,
		Payload: payload,
	}

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("Error serializando mensaje WS")
		return
	}

	// Extraer instance_id del payload
	instanceID := ""
	if payloadMap, ok := payload.(map[string]interface{}); ok {
		if id, ok := payloadMap["instance_id"].(string); ok {
			instanceID = id
		}
	}

	if instanceID == "" {
		log.Warn().Msg("No se pudo extraer instance_id del payload, no se enviará mensaje WS")
		return
	}

	s.broadcast <- BroadcastMessage{
		InstanceID: instanceID,
		Data:       jsonMsg,
	}
}
