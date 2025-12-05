package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"kero-kero/internal/models"
	"kero-kero/internal/services"
	"kero-kero/pkg/errors"
)

type CRMHandler struct {
	service *services.CRMService
}

func NewCRMHandler(service *services.CRMService) *CRMHandler {
	return &CRMHandler{service: service}
}

// ListContacts maneja GET /instances/{instanceID}/crm/contacts
func (h *CRMHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	contacts, err := h.service.ListContacts(r.Context(), instanceID)
	if err != nil {
		errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contacts)
}

// UpdateContact maneja PUT /instances/{instanceID}/crm/contacts/{jid}
func (h *CRMHandler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	jid := chi.URLParam(r, "jid")

	var req models.UpdateCRMContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	contact, err := h.service.UpdateContact(r.Context(), instanceID, jid, &req)
	if err != nil {
		errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contact)
}

// GetContact maneja GET /instances/{instanceID}/crm/contacts/{jid}
func (h *CRMHandler) GetContact(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	jid := chi.URLParam(r, "jid")

	contact, err := h.service.GetContact(r.Context(), instanceID, jid)
	if err != nil {
		if err == errors.ErrNotFound {
			errors.WriteJSON(w, errors.ErrNotFound)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contact)
}
