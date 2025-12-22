package errors

import (
	"encoding/json"
	"net/http"
)

// AppError representa un error de la aplicación
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error implementa la interfaz error
func (e *AppError) Error() string {
	return e.Message
}

// ErrorResponse representa la respuesta de error HTTP
type ErrorResponse struct {
	Success bool     `json:"success"`
	Error   AppError `json:"error"`
}

// Errores predefinidos
var (
	ErrBadRequest         = &AppError{Code: http.StatusBadRequest, Message: "Solicitud inválida"}
	ErrUnauthorized       = &AppError{Code: http.StatusUnauthorized, Message: "No autorizado"}
	ErrForbidden          = &AppError{Code: http.StatusForbidden, Message: "Prohibido"}
	ErrNotFound           = &AppError{Code: http.StatusNotFound, Message: "No encontrado"}
	ErrConflict           = &AppError{Code: http.StatusConflict, Message: "Conflicto"}
	ErrInternalServer     = &AppError{Code: http.StatusInternalServerError, Message: "Error interno del servidor"}
	ErrServiceUnavailable = &AppError{Code: http.StatusServiceUnavailable, Message: "Servicio no disponible"}
	ErrInstanceNotFound   = &AppError{Code: http.StatusNotFound, Message: "Instancia no encontrada"}
	ErrInstanceExists     = &AppError{Code: http.StatusConflict, Message: "La instancia ya existe"}
	ErrNotAuthenticated   = &AppError{Code: http.StatusUnauthorized, Message: "Instancia no autenticada"}
	ErrDatabaseLocked     = &AppError{Code: http.StatusServiceUnavailable, Message: "Base de datos ocupada, intente nuevamente"}
	ErrRateLimitReached   = &AppError{Code: http.StatusTooManyRequests, Message: "Límite de mensajes alcanzado, reintentando asíncronamente"}
)

// New crea un nuevo error personalizado
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// WithDetails añade detalles al error
func (e *AppError) WithDetails(details string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

// WriteJSON escribe el error como JSON en la respuesta HTTP
func WriteJSON(w http.ResponseWriter, err *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)

	response := ErrorResponse{
		Success: false,
		Error:   *err,
	}

	json.NewEncoder(w).Encode(response)
}

// FromError convierte un error genérico en AppError
func FromError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return ErrInternalServer.WithDetails(err.Error())
}
