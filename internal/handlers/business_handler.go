package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"kero-kero/internal/models"
	"kero-kero/internal/services"
	"kero-kero/pkg/errors"
)

type BusinessHandler struct {
	service *services.BusinessService
}

func NewBusinessHandler(service *services.BusinessService) *BusinessHandler {
	return &BusinessHandler{service: service}
}

// CreateLabel maneja POST /instances/{instanceID}/business/labels
func (h *BusinessHandler) CreateLabel(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req models.CreateLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if err := h.service.CreateLabel(r.Context(), instanceID, &req); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// AssignLabel maneja POST /instances/{instanceID}/business/labels/assign
func (h *BusinessHandler) AssignLabel(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req models.LabelActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if err := h.service.AssignLabel(r.Context(), instanceID, &req); err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// GetProfile maneja GET /instances/{instanceID}/business/profile
func (h *BusinessHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	jid := r.URL.Query().Get("jid")

	profile, err := h.service.GetBusinessProfile(r.Context(), instanceID, jid)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// SetAutoLabelRules maneja POST /instances/{instanceID}/business/autolabel/rules
func (h *BusinessHandler) SetAutoLabelRules(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req []models.AutoLabelRule
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if err := h.service.SetAutoLabelRules(r.Context(), instanceID, req); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// GetAutoLabelRules maneja GET /instances/{instanceID}/business/autolabel/rules
func (h *BusinessHandler) GetAutoLabelRules(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	rules, err := h.service.GetAutoLabelRules(r.Context(), instanceID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}
