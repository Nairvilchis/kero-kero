package handlers

import (
	"encoding/json"
	"net/http"

	"kero-kero/internal/services"
	"kero-kero/pkg/errors"
)

// AuthHandler maneja las operaciones de autenticación
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler crea una nueva instancia del handler de autenticación
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginRequest representa la solicitud de login desde el cliente
type LoginRequest struct {
	APIKey string `json:"api_key"`
}

// Login maneja la autenticación y genera un token JWT
// POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	// Decodificar el cuerpo de la solicitud
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("Error decodificando solicitud"))
		return
	}

	// Validar que se proporcionó la API key
	if req.APIKey == "" {
		errors.WriteJSON(w, errors.ErrBadRequest.WithDetails("API key es requerida"))
		return
	}

	// Realizar login
	response, err := h.authService.Login(req.APIKey)
	if err != nil {
		errors.WriteJSON(w, errors.ErrUnauthorized.WithDetails("Credenciales inválidas"))
		return
	}

	// Responder con el token
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ValidateToken valida un token JWT (útil para debugging o verificación del frontend)
// GET /auth/validate
func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	// Obtener el token del header Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		errors.WriteJSON(w, errors.ErrUnauthorized.WithDetails("Token no proporcionado"))
		return
	}

	// Validar el token
	claims, err := h.authService.ValidateToken(authHeader)
	if err != nil {
		errors.WriteJSON(w, errors.ErrUnauthorized.WithDetails("Token inválido: "+err.Error()))
		return
	}

	// Responder con los claims
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":  true,
		"claims": claims,
	})
}
