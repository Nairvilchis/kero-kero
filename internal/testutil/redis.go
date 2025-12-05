package testutil

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// NewMockRedis crea un servidor Redis en memoria para testing
func NewMockRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()

	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Verificar conexión
	err = client.Ping(context.Background()).Err()
	require.NoError(t, err)

	return mr, client
}

// CleanupRedis limpia el servidor Redis mock
func CleanupRedis(t *testing.T, mr *miniredis.Miniredis, client *redis.Client) {
	t.Helper()
	if client != nil {
		client.Close()
	}
	if mr != nil {
		mr.Close()
	}
}
