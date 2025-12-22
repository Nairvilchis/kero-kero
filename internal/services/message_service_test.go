package services

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	"kero-kero/internal/models"
	"kero-kero/internal/repository"
	"kero-kero/internal/testutil"
	"kero-kero/internal/whatsapp"
)

func setupMessageService(t *testing.T) (*MessageService, *whatsapp.Manager, *repository.MessageRepository, func()) {
	db := testutil.NewMockDatabase(t)
	mr, redisClient := testutil.NewMockRedis(t)

	redisRepo := &repository.RedisClient{Client: redisClient}
	instanceRepo := repository.NewInstanceRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	container, err := sqlstore.New(context.Background(), "sqlite3", "file::memory:?cache=shared&_foreign_keys=on", waLog.Noop)
	require.NoError(t, err)

	waManager := whatsapp.NewManager(container, instanceRepo, messageRepo, redisRepo)
	service := NewMessageService(waManager, messageRepo)

	cleanup := func() {
		testutil.CleanupRedis(t, mr, redisClient)
		testutil.CleanupDatabase(t, db)
	}

	return service, waManager, messageRepo, cleanup
}

func TestMessageService_Validation(t *testing.T) {
	service, _, _, cleanup := setupMessageService(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("error si instancia no existe", func(t *testing.T) {
		req := &models.SendTextRequest{
			Phone:   "5215512345678",
			Message: "Hola",
		}
		_, err := service.SendText(ctx, "non-existent", req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no encontrada")
	})

	t.Run("error si el numero es invalido", func(t *testing.T) {
		// Primero crear la instancia para pasar el primer check
		service.waManager.CreateInstance(ctx, "test-inst", false)

		req := &models.SendTextRequest{
			Phone:   "abc",
			Message: "Hola",
		}
		// Nota: SendText fallará por "no autenticado" antes que por validación si no simulamos estar logueados
		// Pero como ValidatePhoneNumber se llama al inicio, debería fallar por validación primero.
		_, err := service.SendText(ctx, "test-inst", req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dígitos")
	})
}
