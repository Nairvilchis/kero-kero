package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"kero-kero/internal/models"
	"kero-kero/internal/services"
	"kero-kero/pkg/errors"
)

// MessageHandler maneja las peticiones HTTP de mensajes
type MessageHandler struct {
	service      *services.MessageService
	queueService *services.QueueService
}

// NewMessageHandler crea un nuevo handler de mensajes
func NewMessageHandler(service *services.MessageService, queueService *services.QueueService) *MessageHandler {
	return &MessageHandler{
		service:      service,
		queueService: queueService,
	}
}

// SendText maneja POST /instances/{instanceID}/messages/text
func (h *MessageHandler) SendText(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req models.SendTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if req.Phone == "" || req.Message == "" {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("phone y message son requeridos"))
		return
	}

	// Verificar si se solicitó envío asíncrono
	if r.Header.Get("X-Async") == "true" {
		msgID, err := h.queueService.EnqueueMessage(r.Context(), instanceID, models.MessageTypeText, req)
		if err != nil {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails("Error encolando mensaje: "+err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"message_id": msgID,
			"status":     "queued",
		})
		return
	}

	response, err := h.service.SendText(r.Context(), instanceID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SendImage maneja POST /instances/{instanceID}/messages/image
func (h *MessageHandler) SendImage(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req models.SendMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	// Verificar si se solicitó envío asíncrono
	if r.Header.Get("X-Async") == "true" {
		msgID, err := h.queueService.EnqueueMessage(r.Context(), instanceID, models.MessageTypeImage, req)
		if err != nil {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails("Error encolando mensaje: "+err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"message_id": msgID,
			"status":     "queued",
		})
		return
	}

	response, err := h.service.SendImage(r.Context(), instanceID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SendVideo maneja POST /instances/{instanceID}/messages/video
func (h *MessageHandler) SendVideo(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req models.SendMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	// Verificar si se solicitó envío asíncrono
	if r.Header.Get("X-Async") == "true" {
		msgID, err := h.queueService.EnqueueMessage(r.Context(), instanceID, models.MessageTypeVideo, req)
		if err != nil {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails("Error encolando mensaje: "+err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"message_id": msgID,
			"status":     "queued",
		})
		return
	}

	response, err := h.service.SendVideo(r.Context(), instanceID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SendAudio maneja POST /instances/{instanceID}/messages/audio
func (h *MessageHandler) SendAudio(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req models.SendMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	// Verificar si se solicitó envío asíncrono
	if r.Header.Get("X-Async") == "true" {
		msgID, err := h.queueService.EnqueueMessage(r.Context(), instanceID, models.MessageTypeAudio, req)
		if err != nil {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails("Error encolando mensaje: "+err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"message_id": msgID,
			"status":     "queued",
		})
		return
	}

	response, err := h.service.SendAudio(r.Context(), instanceID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SendDocument maneja POST /instances/{instanceID}/messages/document
func (h *MessageHandler) SendDocument(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req models.SendMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	// Verificar si se solicitó envío asíncrono
	if r.Header.Get("X-Async") == "true" {
		msgID, err := h.queueService.EnqueueMessage(r.Context(), instanceID, models.MessageTypeDocument, req)
		if err != nil {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails("Error encolando mensaje: "+err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"message_id": msgID,
			"status":     "queued",
		})
		return
	}

	response, err := h.service.SendDocument(r.Context(), instanceID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SendLocation maneja POST /instances/{instanceID}/messages/location
func (h *MessageHandler) SendLocation(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req models.SendLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if req.Phone == "" {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("phone es requerido"))
		return
	}

	// Verificar si se solicitó envío asíncrono
	if r.Header.Get("X-Async") == "true" {
		msgID, err := h.queueService.EnqueueMessage(r.Context(), instanceID, models.MessageTypeLocation, req)
		if err != nil {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails("Error encolando mensaje: "+err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"message_id": msgID,
			"status":     "queued",
		})
		return
	}

	response, err := h.service.SendLocation(r.Context(), instanceID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SendContact maneja POST /instances/{instanceID}/messages/contact
func (h *MessageHandler) SendContact(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req models.SendContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if req.Phone == "" || req.VCard == "" {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("phone y vcard son requeridos"))
		return
	}

	response, err := h.service.SendContact(r.Context(), instanceID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// React maneja POST /instances/{instanceID}/messages/react
func (h *MessageHandler) React(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req models.ReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	resp, err := h.service.ReactToMessage(r.Context(), instanceID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Revoke maneja POST /instances/{instanceID}/messages/revoke
func (h *MessageHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req models.RevokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	resp, err := h.service.RevokeMessage(r.Context(), instanceID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DownloadMedia maneja POST /instances/{instanceID}/messages/download
func (h *MessageHandler) DownloadMedia(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req models.DownloadMediaRequest
	if err := json.NewDecoder(r.Context().Value("body").(io.ReadCloser)).Decode(&req); err != nil {
		// Si el middleware no inyectó el body, lo leemos directamente
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("Cuerpo de solicitud inválido: "+err.Error()))
			return
		}
	}

	data, err := h.service.DownloadMedia(r.Context(), instanceID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	filename := "media"
	if req.Type != "" {
		filename += "." + req.Type
	}

	w.Header().Set("Content-Type", req.Mimetype)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Write(data)
}

// CreatePoll maneja POST /instances/{instanceID}/messages/poll
func (h *MessageHandler) CreatePoll(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req models.CreatePollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	resp, err := h.service.CreatePoll(r.Context(), instanceID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// VotePoll maneja POST /instances/{instanceID}/messages/poll/vote
func (h *MessageHandler) VotePoll(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req models.VotePollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	resp, err := h.service.VotePoll(r.Context(), instanceID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// SendTextWithTyping maneja POST /instances/{instanceID}/messages/text-with-typing
func (h *MessageHandler) SendTextWithTyping(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req models.SendTextWithTypingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if req.Phone == "" || req.Message == "" {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("phone y message son requeridos"))
		return
	}

	response, err := h.service.SendTextWithTyping(r.Context(), instanceID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// MarkAsRead maneja POST /instances/{instanceID}/messages/mark-read
func (h *MessageHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req models.MarkAsReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if req.MessageID == "" || req.ChatJID == "" || req.SenderJID == "" {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("message_id, chat_jid y sender_jid son requeridos"))
		return
	}

	response, err := h.service.MarkAsRead(r.Context(), instanceID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Edit maneja POST /instances/{instanceID}/messages/edit
func (h *MessageHandler) Edit(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req models.EditMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if req.Phone == "" || req.MessageID == "" || req.NewText == "" {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("phone, message_id y new_text son requeridos"))
		return
	}

	response, err := h.service.EditMessage(r.Context(), instanceID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
