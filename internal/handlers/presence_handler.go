package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"kero-kero/internal/models"
	"kero-kero/internal/services"
	"kero-kero/pkg/errors"
)

// PresenceHandler maneja las peticiones HTTP de presencia
type PresenceHandler struct {
	service *services.PresenceService
}

// NewPresenceHandler crea un nuevo handler de presencia
func NewPresenceHandler(service *services.PresenceService) *PresenceHandler {
	return &PresenceHandler{service: service}
}

// Start maneja POST /instances/{instanceID}/presence/start
func (h *PresenceHandler) Start(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req models.StartPresenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	response, err := h.service.StartPresence(r.Context(), instanceID, &req)
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

// Stop maneja POST /instances/{instanceID}/presence/stop
func (h *PresenceHandler) Stop(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req models.StopPresenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	response, err := h.service.StopPresence(r.Context(), instanceID, &req)
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

// Timed maneja POST /instances/{instanceID}/presence/timed
func (h *PresenceHandler) Timed(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req models.TimedPresenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	response, err := h.service.TimedPresence(r.Context(), instanceID, &req)
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

// SetStatus maneja POST /instances/{instanceID}/presence/status
func (h *PresenceHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req models.SetStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	response, err := h.service.SetOnlineStatus(r.Context(), instanceID, &req)
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
