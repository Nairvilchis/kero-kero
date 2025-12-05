package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"kero-kero/internal/models"
	"kero-kero/internal/services"
	"kero-kero/pkg/errors"
)

type GroupHandler struct {
	service *services.GroupService
}

func NewGroupHandler(service *services.GroupService) *GroupHandler {
	return &GroupHandler{service: service}
}

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	var req models.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	resp, err := h.service.CreateGroup(r.Context(), instanceID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *GroupHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	resp, err := h.service.ListGroups(r.Context(), instanceID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *GroupHandler) GetGroupInfo(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	groupID := chi.URLParam(r, "groupID")

	resp, err := h.service.GetGroupInfo(r.Context(), instanceID, groupID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *GroupHandler) UpdateParticipants(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	groupID := chi.URLParam(r, "groupID")

	var req models.GroupParticipantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if err := h.service.UpdateParticipants(r.Context(), instanceID, groupID, &req); err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *GroupHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	groupID := chi.URLParam(r, "groupID")

	if err := h.service.LeaveGroup(r.Context(), instanceID, groupID); err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func handleError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		errors.WriteJSON(w, appErr)
	} else {
		errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
	}
}

// AddParticipants maneja POST /instances/{instanceID}/groups/{groupID}/participants
func (h *GroupHandler) AddParticipants(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	groupID := chi.URLParam(r, "groupID")

	var req models.AddParticipantsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	resp, err := h.service.AddParticipants(r.Context(), instanceID, groupID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RemoveParticipants maneja DELETE /instances/{instanceID}/groups/{groupID}/participants
func (h *GroupHandler) RemoveParticipants(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	groupID := chi.URLParam(r, "groupID")

	var req models.RemoveParticipantsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	resp, err := h.service.RemoveParticipants(r.Context(), instanceID, groupID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// PromoteToAdmin maneja POST /instances/{instanceID}/groups/{groupID}/admins
func (h *GroupHandler) PromoteToAdmin(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	groupID := chi.URLParam(r, "groupID")

	var req models.PromoteAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	resp, err := h.service.PromoteToAdmin(r.Context(), instanceID, groupID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DemoteAdmin maneja DELETE /instances/{instanceID}/groups/{groupID}/admins
func (h *GroupHandler) DemoteAdmin(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	groupID := chi.URLParam(r, "groupID")

	var req models.DemoteAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	resp, err := h.service.DemoteAdmin(r.Context(), instanceID, groupID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateGroupInfo maneja PUT /instances/{instanceID}/groups/{groupID}
func (h *GroupHandler) UpdateGroupInfo(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	groupID := chi.URLParam(r, "groupID")

	var req models.UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	resp, err := h.service.UpdateGroupInfo(r.Context(), instanceID, groupID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateGroupSettings maneja PUT /instances/{instanceID}/groups/{groupID}/settings
func (h *GroupHandler) UpdateGroupSettings(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	groupID := chi.URLParam(r, "groupID")

	var req models.UpdateGroupSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	resp, err := h.service.UpdateGroupSettings(r.Context(), instanceID, groupID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetInviteLink maneja GET /instances/{instanceID}/groups/{groupID}/invite
func (h *GroupHandler) GetInviteLink(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	groupID := chi.URLParam(r, "groupID")

	resp, err := h.service.GetInviteLink(r.Context(), instanceID, groupID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RevokeInviteLink maneja POST /instances/{instanceID}/groups/{groupID}/invite/revoke
func (h *GroupHandler) RevokeInviteLink(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	groupID := chi.URLParam(r, "groupID")

	resp, err := h.service.RevokeInviteLink(r.Context(), instanceID, groupID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
