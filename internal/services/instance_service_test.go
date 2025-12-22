package services

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mau.fi/whatsmeow/store/sqlstore"

	"kero-kero/internal/models"
	"kero-kero/internal/repository"
	"kero-kero/internal/testutil"
	"kero-kero/internal/whatsapp"

	waLog "go.mau.fi/whatsmeow/util/log"
)

func setupInstanceService(t *testing.T) (*InstanceService, *repository.Database, *repository.RedisClient, func()) {
	db := testutil.NewMockDatabase(t)
	mr, redisClient := testutil.NewMockRedis(t)

	redisRepo := &repository.RedisClient{Client: redisClient}
	instanceRepo := repository.NewInstanceRepository(db)
	webhookRepo := repository.NewWebhookRepository(redisRepo)
	webhookSvc := NewWebhookService(webhookRepo)

	// Setup whatsmeow container in memory
	container, err := sqlstore.New(context.Background(), "sqlite3", "file::memory:?cache=shared&_foreign_keys=on", waLog.Noop)
	require.NoError(t, err)

	waManager := whatsapp.NewManager(container, instanceRepo, nil, redisRepo)

	service := NewInstanceService(waManager, instanceRepo, redisRepo, webhookSvc)

	cleanup := func() {
		testutil.CleanupRedis(t, mr, redisClient)
		testutil.CleanupDatabase(t, db)
	}

	return service, db, redisRepo, cleanup
}

func TestInstanceService_CreateInstance(t *testing.T) {
	service, _, _, cleanup := setupInstanceService(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("crear instancia exitosamente", func(t *testing.T) {
		req := &models.CreateInstanceRequest{
			InstanceID:  "test-instance",
			SyncHistory: false,
		}

		instance, err := service.CreateInstance(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "test-instance", instance.InstanceID)
		assert.Equal(t, string(models.StatusDisconnected), instance.Status)
	})

	t.Run("error al crear instancia duplicada", func(t *testing.T) {
		req := &models.CreateInstanceRequest{
			InstanceID: "test-instance", // Ya existe del test anterior (si no limpiamos base de datos, pero setupInstanceService usa :memory: por test?)
		}
		// En este caso, setupInstanceService se llama una vez para todos los t.Run si lo pongo arriba.
		// Volvamos a crear para asegurar aislamiento si es necesario, o usemos IDs diferentes.

		_, err := service.CreateInstance(ctx, req)
		assert.Error(t, err)
	})
}

func TestInstanceService_ListInstances(t *testing.T) {
	service, _, _, cleanup := setupInstanceService(t)
	defer cleanup()

	ctx := context.Background()

	service.CreateInstance(ctx, &models.CreateInstanceRequest{InstanceID: "inst-1"})
	service.CreateInstance(ctx, &models.CreateInstanceRequest{InstanceID: "inst-2"})

	instances, err := service.ListInstances(ctx)
	require.NoError(t, err)
	assert.Len(t, instances, 2)
}

func TestInstanceService_DeleteInstance(t *testing.T) {
	service, _, _, cleanup := setupInstanceService(t)
	defer cleanup()

	ctx := context.Background()
	service.CreateInstance(ctx, &models.CreateInstanceRequest{InstanceID: "to-delete"})

	err := service.DeleteInstance(ctx, "to-delete")
	require.NoError(t, err)

	exists, _ := service.instanceRepo.Exists(ctx, "to-delete")
	assert.False(t, exists)
}
