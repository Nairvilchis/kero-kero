package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"kero-kero/internal/models"
	"kero-kero/internal/services"
	"kero-kero/pkg/errors"
)

type ContactHandler struct {
	service *services.ContactService
}

func NewContactHandler(service *services.ContactService) *ContactHandler {
	return &ContactHandler{service: service}
}

func (h *ContactHandler) CheckContacts(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req models.CheckContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	resp, err := h.service.CheckContacts(r.Context(), instanceID, req.Phones)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ContactHandler) GetProfilePicture(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	phone := chi.URLParam(r, "phone") // Intentar obtener de URL param

	if phone == "" {
		phone = r.URL.Query().Get("phone") // Si no, obtener de query param
	}

	if phone == "" {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("phone requerido (en URL o query param)"))
		return
	}

	resp, err := h.service.GetProfilePicture(r.Context(), instanceID, phone)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ContactHandler) GetContacts(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	resp, err := h.service.GetContacts(r.Context(), instanceID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ContactHandler) SubscribePresence(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req models.PresenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if err := h.service.SubscribePresence(r.Context(), instanceID, req.Phone); err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *ContactHandler) Block(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	phone := chi.URLParam(r, "phone") // Intentar obtener de URL param

	var req models.BlockRequest
	// Si no está en URL, intentar decodificar del body
	if phone == "" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
			return
		}
		phone = req.Phone
	}

	if phone == "" {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("phone requerido"))
		return
	}

	if err := h.service.BlockContact(r.Context(), instanceID, phone); err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *ContactHandler) Unblock(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	phone := chi.URLParam(r, "phone") // Intentar obtener de URL param

	var req models.BlockRequest
	// Si no está en URL, intentar decodificar del body
	if phone == "" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
			return
		}
		phone = req.Phone
	}

	if phone == "" {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("phone requerido"))
		return
	}

	if err := h.service.UnblockContact(r.Context(), instanceID, phone); err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// CheckNumbers maneja POST /instances/{instanceID}/contacts/check
func (h *ContactHandler) CheckNumbers(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	var req models.CheckNumbersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	response, err := h.service.CheckNumbers(r.Context(), instanceID, &req)
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

// GetContactInfo maneja GET /instances/{instanceID}/contacts/{phone}
func (h *ContactHandler) GetContactInfo(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	phone := chi.URLParam(r, "phone")

	resp, err := h.service.GetContactInfo(r.Context(), instanceID, phone)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetAbout maneja GET /instances/{instanceID}/contacts/{phone}/about
func (h *ContactHandler) GetAbout(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	phone := chi.URLParam(r, "phone")

	resp, err := h.service.GetAbout(r.Context(), instanceID, phone)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
