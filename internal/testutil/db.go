package testutil

import (
	"testing"

	"kero-kero/internal/config"
	"kero-kero/internal/repository"

	"github.com/stretchr/testify/require"
)

// NewMockDatabase crea una base de datos SQLite en memoria para tests
func NewMockDatabase(t *testing.T) *repository.Database {
	t.Helper()

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Driver:     "sqlite",
			SQLitePath: ":memory:",
		},
	}

	db, err := repository.NewDatabase(cfg)
	require.NoError(t, err)

	return db
}

// CleanupDatabase cierra la conexión a la base de datos
func CleanupDatabase(t *testing.T, db *repository.Database) {
	t.Helper()
	if db != nil {
		db.Close()
	}
}
