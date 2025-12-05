package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"kero-kero/internal/services"
	"kero-kero/pkg/errors"
)

type SyncHandler struct {
	service *services.SyncService
}

func NewSyncHandler(service *services.SyncService) *SyncHandler {
	return &SyncHandler{service: service}
}

func (h *SyncHandler) Start(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var opts services.SyncOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		// Si no hay body, usar opciones por defecto (modo básico)
		opts = services.SyncOptions{
			MessagesPerChat: 50,
			MaxChats:        20,
			Advanced:        false,
		}
	}

	if err := h.service.SyncChatHistory(r.Context(), instanceID, opts); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Sincronización iniciada en segundo plano",
	})
}

func (h *SyncHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	progress, err := h.service.GetSyncProgress(instanceID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    progress,
	})
}

func (h *SyncHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	if err := h.service.CancelSync(instanceID); err != nil {
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
