package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"

	"kero-kero/internal/config"
)

// Database representa la conexión a la base de datos
type Database struct {
	DB     *sql.DB
	Driver string
}

// NewDatabase crea una nueva conexión a la base de datos
func NewDatabase(cfg *config.Config) (*Database, error) {
	driver := cfg.Database.Driver
	dsn := cfg.GetDSN()

	// Mapear "sqlite" a "sqlite3" para el driver
	sqlDriver := driver
	if driver == "sqlite" {
		sqlDriver = "sqlite3"
	}

	log.Info().
		Str("driver", driver).
		Msg("Conectando a base de datos")

	db, err := sql.Open(sqlDriver, dsn)
	if err != nil {
		return nil, fmt.Errorf("error abriendo base de datos: %w", err)
	}

	// Configurar pool de conexiones
	if driver == "postgres" {
		db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
		db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
		db.SetConnMaxLifetime(time.Hour)
	} else {
		// SQLite: una sola conexión de escritura
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
	}

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error conectando a base de datos: %w", err)
	}

	database := &Database{
		DB:     db,
		Driver: driver,
	}

	// Configurar SQLite si es necesario
	if driver == "sqlite" {
		if err := database.configureSQLite(); err != nil {
			return nil, err
		}
	}

	// Ejecutar migraciones
	if err := database.Migrate(); err != nil {
		return nil, fmt.Errorf("error ejecutando migraciones: %w", err)
	}

	log.Info().Msg("Base de datos conectada exitosamente")
	return database, nil
}

// configureSQLite configura SQLite para mejor concurrencia
func (d *Database) configureSQLite() error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=1000000000",
		"PRAGMA foreign_keys=ON",
		"PRAGMA temp_store=MEMORY",
	}

	for _, pragma := range pragmas {
		if _, err := d.DB.Exec(pragma); err != nil {
			return fmt.Errorf("error configurando SQLite (%s): %w", pragma, err)
		}
	}

	log.Debug().Msg("SQLite configurado con WAL mode y optimizaciones")
	return nil
}

// Migrate ejecuta las migraciones de base de datos
func (d *Database) Migrate() error {
	log.Info().Msg("Ejecutando migraciones de base de datos")

	migrations := d.getMigrations()

	for _, migration := range migrations {
		log.Debug().Str("migration", migration.name).Msg("Ejecutando migración")
		if _, err := d.DB.Exec(migration.sql); err != nil {
			// Ignorar error si la columna ya existe (migración idempotente simple)
			if d.Driver == "sqlite" && (err.Error() == "duplicate column name: sync_history" || err.Error() == "table instance_metadata already has a column named sync_history") {
				continue
			}
			if d.Driver == "postgres" && (err.Error() == `pq: column "sync_history" of relation "instance_metadata" already exists`) {
				continue
			}
			// Si la tabla ya existe, también ignorar (para create table)
			// Pero create table usa IF NOT EXISTS, así que no debería fallar por eso.

			// Hack: Si el error contiene "already exists" o "duplicate column", lo logueamos y continuamos
			// Esto es para permitir agregar columnas de forma segura sin un sistema de migración complejo
			errMsg := err.Error()
			if contains(errMsg, "already exists") || contains(errMsg, "duplicate column") {
				log.Warn().Str("migration", migration.name).Err(err).Msg("Migración ya aplicada o conflicto ignorado")
				continue
			}

			return fmt.Errorf("error en migración %s: %w", migration.name, err)
		}
	}

	log.Info().Msg("Migraciones completadas exitosamente")
	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}

type migration struct {
	name string
	sql  string
}

// getMigrations retorna las migraciones según el driver
func (d *Database) getMigrations() []migration {
	if d.Driver == "postgres" {
		return d.getPostgresMigrations()
	}
	return d.getSQLiteMigrations()
}

// getSQLiteMigrations retorna migraciones para SQLite
func (d *Database) getSQLiteMigrations() []migration {
	return []migration{
		{
			name: "create_instance_mapping",
			sql: `CREATE TABLE IF NOT EXISTS instance_mapping (
				instance_id TEXT PRIMARY KEY,
				jid TEXT NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
		},
		{
			name: "create_instance_metadata",
			sql: `CREATE TABLE IF NOT EXISTS instance_metadata (
				instance_id TEXT PRIMARY KEY,
				name TEXT,
				webhook_url TEXT,
				status TEXT DEFAULT 'disconnected',
				sync_history BOOLEAN DEFAULT FALSE,
				last_connected_at DATETIME,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (instance_id) REFERENCES instance_mapping(instance_id) ON DELETE CASCADE
			)`,
		},
		{
			name: "add_sync_history_column",
			sql:  `ALTER TABLE instance_metadata ADD COLUMN sync_history BOOLEAN DEFAULT FALSE`,
		},
		{
			name: "create_message_queue",
			sql: `CREATE TABLE IF NOT EXISTS message_queue (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				instance_id TEXT NOT NULL,
				recipient TEXT NOT NULL,
				message_type TEXT NOT NULL,
				content TEXT NOT NULL,
				status TEXT DEFAULT 'pending',
				attempts INTEGER DEFAULT 0,
				error TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				processed_at DATETIME
			)`,
		},
		{
			name: "create_messages",
			sql: `CREATE TABLE IF NOT EXISTS messages (
				id TEXT PRIMARY KEY,
				instance_id TEXT NOT NULL,
				jid TEXT NOT NULL,
				from_me BOOLEAN NOT NULL,
				content TEXT,
				push_name TEXT,
				timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
				status TEXT,
				type TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
		},
		{
			name: "add_push_name_to_messages",
			sql:  `ALTER TABLE messages ADD COLUMN push_name TEXT`,
		},
	}
}

// getPostgresMigrations retorna migraciones para PostgreSQL
func (d *Database) getPostgresMigrations() []migration {
	return []migration{
		{
			name: "create_instance_mapping",
			sql: `CREATE TABLE IF NOT EXISTS instance_mapping (
				instance_id TEXT PRIMARY KEY,
				jid TEXT NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
		},
		{
			name: "create_instance_metadata",
			sql: `CREATE TABLE IF NOT EXISTS instance_metadata (
				instance_id TEXT PRIMARY KEY,
				name TEXT,
				webhook_url TEXT,
				status TEXT DEFAULT 'disconnected',
				sync_history BOOLEAN DEFAULT FALSE,
				last_connected_at TIMESTAMP,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (instance_id) REFERENCES instance_mapping(instance_id) ON DELETE CASCADE
			)`,
		},
		{
			name: "add_sync_history_column",
			sql:  `ALTER TABLE instance_metadata ADD COLUMN IF NOT EXISTS sync_history BOOLEAN DEFAULT FALSE`,
		},
		{
			name: "create_message_queue",
			sql: `CREATE TABLE IF NOT EXISTS message_queue (
				id SERIAL PRIMARY KEY,
				instance_id TEXT NOT NULL,
				recipient TEXT NOT NULL,
				message_type TEXT NOT NULL,
				content TEXT NOT NULL,
				status TEXT DEFAULT 'pending',
				attempts INTEGER DEFAULT 0,
				error TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				processed_at TIMESTAMP
			)`,
		},
		{
			name: "create_messages",
			sql: `CREATE TABLE IF NOT EXISTS messages (
				id TEXT PRIMARY KEY,
				instance_id TEXT NOT NULL,
				jid TEXT NOT NULL,
				from_me BOOLEAN NOT NULL,
				content TEXT,
				push_name TEXT,
				timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				status TEXT,
				type TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
		},
		{
			name: "add_push_name_to_messages",
			sql:  `ALTER TABLE messages ADD COLUMN IF NOT EXISTS push_name TEXT`,
		},
	}
}

// Close cierra la conexión a la base de datos
func (d *Database) Close() error {
	log.Info().Msg("Cerrando conexión a base de datos")
	return d.DB.Close()
}

// Health verifica el estado de la base de datos
func (d *Database) Health(ctx context.Context) error {
	return d.DB.PingContext(ctx)
}
