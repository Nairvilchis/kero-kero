package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"kero-kero/internal/models"
	"kero-kero/internal/services"
	"kero-kero/pkg/errors"
)

// InstanceHandler maneja las peticiones HTTP de instancias
type InstanceHandler struct {
	service *services.InstanceService
}

// NewInstanceHandler crea un nuevo handler de instancias
func NewInstanceHandler(service *services.InstanceService) *InstanceHandler {
	return &InstanceHandler{service: service}
}

// Create maneja POST /instances
func (h *InstanceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if req.InstanceID == "" {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("instance_id es requerido"))
		return
	}

	instance, err := h.service.CreateInstance(r.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.InstanceResponse{
		Success: true,
		Data:    instance,
		Message: "Instancia creada exitosamente",
	})
}

// List maneja GET /instances
func (h *InstanceHandler) List(w http.ResponseWriter, r *http.Request) {
	instances, err := h.service.ListInstances(r.Context())
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.InstanceListResponse{
		Success: true,
		Data:    instances,
		Total:   len(instances),
	})
}

// Get maneja GET /instances/{instanceID}
func (h *InstanceHandler) Get(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	instance, err := h.service.GetInstance(r.Context(), instanceID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.InstanceResponse{
		Success: true,
		Data:    instance,
	})
}

// Delete maneja DELETE /instances/{instanceID}
func (h *InstanceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	if err := h.service.DeleteInstance(r.Context(), instanceID); err != nil {
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
		"message": "Instancia eliminada exitosamente",
	})
}

// Connect maneja POST /instances/{instanceID}/connect
func (h *InstanceHandler) Connect(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	if err := h.service.ConnectInstance(r.Context(), instanceID); err != nil {
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
		"message": "Conexión iniciada",
	})
}

// Disconnect maneja POST /instances/{instanceID}/disconnect
func (h *InstanceHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	if err := h.service.DisconnectInstance(r.Context(), instanceID); err != nil {
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
		"message": "Instancia desconectada",
	})
}

// GetQR maneja GET /instances/{instanceID}/qr
func (h *InstanceHandler) GetQR(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	qrImage, err := h.service.GetQRCode(r.Context(), instanceID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.WriteJSON(w, appErr)
		} else {
			errors.WriteJSON(w, errors.ErrInternalServer.WithDetails(err.Error()))
		}
		return
	}

	// Verificar si el cliente quiere PNG directo (para compatibilidad)
	acceptHeader := r.Header.Get("Accept")
	if acceptHeader == "image/png" {
		// Devolver imagen PNG directa
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Write(qrImage)
		return
	}

	// Por defecto, devolver JSON con base64
	qrBase64 := base64.StdEncoding.EncodeToString(qrImage)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"instance_id": instanceID,
		"qr_code":     qrBase64,
		"format":      "base64",
		"image_type":  "png",
	})
}

// GetStatus maneja GET /instances/{instanceID}/status
func (h *InstanceHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	status, err := h.service.GetStatus(r.Context(), instanceID)
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
		"success":     true,
		"instance_id": instanceID,
		"status":      status,
	})
}

// Update maneja PUT /instances/{instanceID}
func (h *InstanceHandler) Update(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")
	if instanceID == "" {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("ID de instancia requerido"))
		return
	}

	var req models.UpdateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("JSON inválido"))
		return
	}

	if err := h.service.UpdateInstance(r.Context(), instanceID, &req); err != nil {
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
