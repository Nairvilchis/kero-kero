package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"kero-kero/internal/models"
	"kero-kero/internal/services"
	"kero-kero/pkg/errors"
)

type PrivacyHandler struct {
	service *services.PrivacyService
}

func NewPrivacyHandler(service *services.PrivacyService) *PrivacyHandler {
	return &PrivacyHandler{service: service}
}

func (h *PrivacyHandler) GetPrivacySettings(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	resp, err := h.service.GetPrivacySettings(r.Context(), instanceID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *PrivacyHandler) UpdatePrivacySettings(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req models.PrivacyUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if err := h.service.UpdatePrivacySettings(r.Context(), instanceID, &req); err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
