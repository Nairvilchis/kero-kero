package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"kero-kero/internal/models"
	"kero-kero/internal/services"
	"kero-kero/pkg/errors"
)

type CallHandler struct {
	service *services.CallService
}

func NewCallHandler(service *services.CallService) *CallHandler {
	return &CallHandler{service: service}
}

// RejectCall maneja POST /instances/{instanceID}/calls/reject
func (h *CallHandler) RejectCall(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req struct {
		CallID string `json:"call_id"`
		From   string `json:"from"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if err := h.service.RejectCall(r.Context(), instanceID, req.CallID, req.From); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// GetSettings maneja GET /instances/{instanceID}/calls/settings
func (h *CallHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	settings, err := h.service.GetCallSettings(r.Context(), instanceID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// UpdateSettings maneja PUT /instances/{instanceID}/calls/settings
func (h *CallHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var settings models.CallSettings

	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if err := h.service.SetCallSettings(r.Context(), instanceID, &settings); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
