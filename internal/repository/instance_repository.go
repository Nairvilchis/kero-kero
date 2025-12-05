package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"kero-kero/internal/models"
)

// InstanceRepository maneja las operaciones de base de datos para instancias
type InstanceRepository struct {
	db *Database
}

// NewInstanceRepository crea un nuevo repositorio de instancias
func NewInstanceRepository(db *Database) *InstanceRepository {
	return &InstanceRepository{db: db}
}

// Create crea una nueva instancia en la base de datos
func (r *InstanceRepository) Create(ctx context.Context, instance *models.Instance) error {
	query := `
		INSERT INTO instance_mapping (instance_id, jid, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`

	now := time.Now()
	instance.CreatedAt = now
	instance.UpdatedAt = now

	_, err := r.db.DB.ExecContext(ctx, query, instance.InstanceID, instance.JID, now, now)
	if err != nil {
		return fmt.Errorf("error creando instancia: %w", err)
	}

	// Crear metadata
	metaQuery := `
		INSERT INTO instance_metadata (instance_id, name, webhook_url, status, sync_history, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.DB.ExecContext(ctx, metaQuery,
		instance.InstanceID,
		instance.Name,
		instance.WebhookURL,
		instance.Status,
		instance.SyncHistory,
		now,
		now,
	)

	return err
}

// GetByID obtiene una instancia por su ID
func (r *InstanceRepository) GetByID(ctx context.Context, instanceID string) (*models.Instance, error) {
	query := `
		SELECT 
			im.instance_id, 
			im.jid, 
			COALESCE(meta.name, '') as name,
			COALESCE(meta.webhook_url, '') as webhook_url,
			COALESCE(meta.status, 'disconnected') as status,
			COALESCE(meta.sync_history, FALSE) as sync_history,
			meta.last_connected_at,
			im.created_at, 
			im.updated_at
		FROM instance_mapping im
		LEFT JOIN instance_metadata meta ON im.instance_id = meta.instance_id
		WHERE im.instance_id = $1
	`

	instance := &models.Instance{}
	var lastConnected sql.NullTime

	err := r.db.DB.QueryRowContext(ctx, query, instanceID).Scan(
		&instance.InstanceID,
		&instance.JID,
		&instance.Name,
		&instance.WebhookURL,
		&instance.Status,
		&instance.SyncHistory,
		&lastConnected,
		&instance.CreatedAt,
		&instance.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error obteniendo instancia: %w", err)
	}

	if lastConnected.Valid {
		instance.LastConnectedAt = &lastConnected.Time
	}

	return instance, nil
}

// GetAll obtiene todas las instancias
func (r *InstanceRepository) GetAll(ctx context.Context) ([]*models.Instance, error) {
	query := `
		SELECT 
			im.instance_id, 
			im.jid, 
			COALESCE(meta.name, '') as name,
			COALESCE(meta.webhook_url, '') as webhook_url,
			COALESCE(meta.status, 'disconnected') as status,
			COALESCE(meta.sync_history, FALSE) as sync_history,
			meta.last_connected_at,
			im.created_at, 
			im.updated_at
		FROM instance_mapping im
		LEFT JOIN instance_metadata meta ON im.instance_id = meta.instance_id
		ORDER BY im.created_at DESC
	`

	rows, err := r.db.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listando instancias: %w", err)
	}
	defer rows.Close()

	var instances []*models.Instance
	for rows.Next() {
		instance := &models.Instance{}
		var lastConnected sql.NullTime

		err := rows.Scan(
			&instance.InstanceID,
			&instance.JID,
			&instance.Name,
			&instance.WebhookURL,
			&instance.Status,
			&instance.SyncHistory,
			&lastConnected,
			&instance.CreatedAt,
			&instance.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando instancia: %w", err)
		}

		if lastConnected.Valid {
			instance.LastConnectedAt = &lastConnected.Time
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

// UpdateStatus actualiza el estado de una instancia
func (r *InstanceRepository) UpdateStatus(ctx context.Context, instanceID string, status models.InstanceStatus) error {
	query := `
		UPDATE instance_metadata 
		SET status = $1, updated_at = $2
		WHERE instance_id = $3
	`

	_, err := r.db.DB.ExecContext(ctx, query, status, time.Now(), instanceID)
	return err
}

// UpdateLastConnected actualiza la última vez que se conectó la instancia
func (r *InstanceRepository) UpdateLastConnected(ctx context.Context, instanceID string) error {
	query := `
		UPDATE instance_metadata 
		SET last_connected_at = $1, updated_at = $2
		WHERE instance_id = $3
	`

	now := time.Now()
	_, err := r.db.DB.ExecContext(ctx, query, now, now, instanceID)
	return err
}

// UpdateJID actualiza el JID de una instancia
func (r *InstanceRepository) UpdateJID(ctx context.Context, instanceID, jid string) error {
	query := `
		UPDATE instance_mapping 
		SET jid = $1, updated_at = $2
		WHERE instance_id = $3
	`

	_, err := r.db.DB.ExecContext(ctx, query, jid, time.Now(), instanceID)
	return err
}

// Delete elimina una instancia
func (r *InstanceRepository) Delete(ctx context.Context, instanceID string) error {
	// Primero eliminar metadata (por si no hay CASCADE)
	_, err := r.db.DB.ExecContext(ctx, "DELETE FROM instance_metadata WHERE instance_id = $1", instanceID)
	if err != nil {
		return fmt.Errorf("error eliminando metadata: %w", err)
	}

	// Luego eliminar el mapeo
	_, err = r.db.DB.ExecContext(ctx, "DELETE FROM instance_mapping WHERE instance_id = $1", instanceID)
	if err != nil {
		return fmt.Errorf("error eliminando instancia: %w", err)
	}

	return nil
}

// Exists verifica si una instancia existe
func (r *InstanceRepository) Exists(ctx context.Context, instanceID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM instance_mapping WHERE instance_id = $1)`
	err := r.db.DB.QueryRowContext(ctx, query, instanceID).Scan(&exists)
	return exists, err
}

// Update actualiza la configuración de una instancia
func (r *InstanceRepository) Update(ctx context.Context, instanceID string, req *models.UpdateInstanceRequest) error {
	query := `
		UPDATE instance_metadata 
		SET name = COALESCE(NULLIF($1, ''), name),
			webhook_url = COALESCE(NULLIF($2, ''), webhook_url),
			updated_at = $3
		WHERE instance_id = $4
	`

	_, err := r.db.DB.ExecContext(ctx, query, req.Name, req.WebhookURL, time.Now(), instanceID)
	return err
}
