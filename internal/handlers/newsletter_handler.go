package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"kero-kero/internal/models"
	"kero-kero/internal/services"
	"kero-kero/pkg/errors"
)

type NewsletterHandler struct {
	service *services.NewsletterService
}

func NewNewsletterHandler(service *services.NewsletterService) *NewsletterHandler {
	return &NewsletterHandler{service: service}
}

// Create maneja POST /instances/{instanceID}/newsletters
func (h *NewsletterHandler) Create(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req models.CreateNewsletterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	resp, err := h.service.CreateNewsletter(r.Context(), instanceID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListSubscribed maneja GET /instances/{instanceID}/newsletters
func (h *NewsletterHandler) ListSubscribed(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	resp, err := h.service.GetSubscribedNewsletters(r.Context(), instanceID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetInfo maneja GET /instances/{instanceID}/newsletters/{jid}
func (h *NewsletterHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	jid := chi.URLParam(r, "jid")

	resp, err := h.service.GetNewsletterInfo(r.Context(), instanceID, jid)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Follow maneja POST /instances/{instanceID}/newsletters/{jid}/follow
func (h *NewsletterHandler) Follow(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	jid := chi.URLParam(r, "jid")

	if err := h.service.FollowNewsletter(r.Context(), instanceID, jid); err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Unfollow maneja POST /instances/{instanceID}/newsletters/{jid}/unfollow
func (h *NewsletterHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	jid := chi.URLParam(r, "jid")

	if err := h.service.UnfollowNewsletter(r.Context(), instanceID, jid); err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// SendMessage maneja POST /instances/{instanceID}/newsletters/send
func (h *NewsletterHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req models.SendNewsletterMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	resp, err := h.service.SendMessage(r.Context(), instanceID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
