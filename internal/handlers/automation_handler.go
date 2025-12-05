package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"kero-kero/internal/models"
	"kero-kero/internal/services"
	"kero-kero/pkg/errors"
)

type AutomationHandler struct {
	service *services.AutomationService
}

func NewAutomationHandler(service *services.AutomationService) *AutomationHandler {
	return &AutomationHandler{service: service}
}

// SendBulkMessage maneja POST /instances/{instanceID}/automation/bulk-message
func (h *AutomationHandler) SendBulkMessage(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req models.BulkMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	resp, err := h.service.SendBulkMessage(r.Context(), instanceID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ScheduleMessage maneja POST /instances/{instanceID}/automation/schedule-message
func (h *AutomationHandler) ScheduleMessage(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req models.ScheduleMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if err := h.service.ScheduleMessage(r.Context(), instanceID, &req); err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// SetAutoReply maneja POST /instances/{instanceID}/automation/auto-reply
func (h *AutomationHandler) SetAutoReply(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req models.AutoReplyConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if err := h.service.SetAutoReply(r.Context(), instanceID, &req); err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// GetAutoReply maneja GET /instances/{instanceID}/automation/auto-reply
func (h *AutomationHandler) GetAutoReply(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	resp, err := h.service.GetAutoReply(r.Context(), instanceID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
