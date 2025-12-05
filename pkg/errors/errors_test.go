package errors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError_Error(t *testing.T) {
	err := New(404, "Not found")
	assert.Equal(t, "Not found", err.Error())
}

func TestAppError_WithDetails(t *testing.T) {
	err := New(400, "Bad request")
	errWithDetails := err.WithDetails("Invalid JSON")

	assert.Equal(t, 400, errWithDetails.Code)
	assert.Equal(t, "Bad request", errWithDetails.Message)
	assert.Equal(t, "Invalid JSON", errWithDetails.Details)
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		wantCode int
		wantMsg  string
	}{
		{
			name:     "ErrBadRequest",
			err:      ErrBadRequest,
			wantCode: http.StatusBadRequest,
			wantMsg:  "Solicitud inválida",
		},
		{
			name:     "ErrUnauthorized",
			err:      ErrUnauthorized,
			wantCode: http.StatusUnauthorized,
			wantMsg:  "No autorizado",
		},
		{
			name:     "ErrNotFound",
			err:      ErrNotFound,
			wantCode: http.StatusNotFound,
			wantMsg:  "No encontrado",
		},
		{
			name:     "ErrInternalServer",
			err:      ErrInternalServer,
			wantCode: http.StatusInternalServerError,
			wantMsg:  "Error interno del servidor",
		},
		{
			name:     "ErrInstanceNotFound",
			err:      ErrInstanceNotFound,
			wantCode: http.StatusNotFound,
			wantMsg:  "Instancia no encontrada",
		},
		{
			name:     "ErrNotAuthenticated",
			err:      ErrNotAuthenticated,
			wantCode: http.StatusUnauthorized,
			wantMsg:  "Instancia no autenticada",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCode, tt.err.Code)
			assert.Equal(t, tt.wantMsg, tt.err.Message)
		})
	}
}

func TestNew(t *testing.T) {
	err := New(418, "I'm a teapot")
	assert.Equal(t, 418, err.Code)
	assert.Equal(t, "I'm a teapot", err.Message)
	assert.Empty(t, err.Details)
}
