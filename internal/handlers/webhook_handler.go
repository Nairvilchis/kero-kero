package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"kero-kero/internal/models"
	"kero-kero/internal/services"
	"kero-kero/pkg/errors"
)

type WebhookHandler struct {
	service *services.WebhookService
}

func NewWebhookHandler(service *services.WebhookService) *WebhookHandler {
	return &WebhookHandler{service: service}
}

func (h *WebhookHandler) SetWebhook(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var config models.WebhookConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	config.InstanceID = instanceID
	if err := h.service.SetWebhook(r.Context(), &config); err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *WebhookHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	config, err := h.service.GetWebhook(r.Context(), instanceID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func (h *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	if err := h.service.DeleteWebhook(r.Context(), instanceID); err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
