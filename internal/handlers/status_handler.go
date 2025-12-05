package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"kero-kero/internal/models"
	"kero-kero/internal/services"
	"kero-kero/pkg/errors"
)

type StatusHandler struct {
	service *services.StatusService
}

func NewStatusHandler(service *services.StatusService) *StatusHandler {
	return &StatusHandler{service: service}
}

// PublishStatus maneja POST /instances/{instanceID}/status
func (h *StatusHandler) PublishStatus(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req models.PublishStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	var resp *models.StatusResponse
	var err error

	switch req.Type {
	case "text":
		resp, err = h.service.PublishTextStatus(r.Context(), instanceID, &req)
	default:
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("Tipo de estado no soportado. Use: text"))
		return
	}

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

// GetStatusPrivacy maneja GET /instances/{instanceID}/status/privacy
func (h *StatusHandler) GetStatusPrivacy(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	resp, err := h.service.GetStatusPrivacy(r.Context(), instanceID)
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
