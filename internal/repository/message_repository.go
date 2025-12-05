package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"kero-kero/internal/models"
)

type MessageRepository struct {
	db *Database
}

func NewMessageRepository(db *Database) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create guarda un nuevo mensaje
func (r *MessageRepository) Create(ctx context.Context, msg *models.Message) error {
	query := `
		INSERT INTO messages (id, instance_id, jid, from_me, content, push_name, timestamp, status, type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO NOTHING
	`

	// Ajustar query para SQLite si es necesario (aunque $N funciona en sqlite moderno)
	if r.db.Driver == "sqlite" || r.db.Driver == "sqlite3" {
		query = `
			INSERT OR IGNORE INTO messages (id, instance_id, jid, from_me, content, push_name, timestamp, status, type)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
	}

	_, err := r.db.DB.ExecContext(ctx, query,
		msg.ID,
		msg.InstanceID,
		msg.To, // Usaremos 'To' como JID del chat por ahora, o 'From' si es entrante. Mejor unificar en 'JID'.
		msg.IsFromMe,
		msg.Content,
		msg.PushName,
		time.Unix(msg.Timestamp, 0),
		msg.Status,
		msg.Type,
	)

	if err != nil {
		return fmt.Errorf("error creating message: %w", err)
	}

	return nil
}

// GetByJID obtiene los mensajes de un chat
func (r *MessageRepository) GetByJID(ctx context.Context, instanceID, jid string, limit int) ([]models.Message, error) {
	query := `
		SELECT id, instance_id, jid, from_me, content, push_name, timestamp, status, type
		FROM messages
		WHERE instance_id = $1 AND jid = $2
		ORDER BY timestamp DESC
		LIMIT $3
	`

	if r.db.Driver == "sqlite" || r.db.Driver == "sqlite3" {
		query = `
			SELECT id, instance_id, jid, from_me, content, push_name, timestamp, status, type
			FROM messages
			WHERE instance_id = ? AND jid = ?
			ORDER BY timestamp DESC
			LIMIT ?
		`
	}

	rows, err := r.db.DB.QueryContext(ctx, query, instanceID, jid, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var ts interface{} // Usar interface{} para robustez
		var instanceIDDB string
		var jidDB string
		var pushName sql.NullString // Usar NullString por si es NULL

		if err := rows.Scan(
			&msg.ID,
			&instanceIDDB,
			&jidDB,
			&msg.IsFromMe,
			&msg.Content,
			&pushName,
			&ts,
			&msg.Status,
			&msg.Type,
		); err != nil {
			return nil, fmt.Errorf("error scanning message: %w", err)
		}

		if pushName.Valid {
			msg.PushName = pushName.String
		}

		if ts != nil {
			switch v := ts.(type) {
			case time.Time:
				msg.Timestamp = v.Unix()
			case string:
				if t, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", v); err == nil {
					msg.Timestamp = t.Unix()
				} else if t, err := time.Parse(time.RFC3339, v); err == nil {
					msg.Timestamp = t.Unix()
				} else if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
					msg.Timestamp = t.Unix()
				}
			case []byte:
				s := string(v)
				if t, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", s); err == nil {
					msg.Timestamp = t.Unix()
				} else if t, err := time.Parse(time.RFC3339, s); err == nil {
					msg.Timestamp = t.Unix()
				} else if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
					msg.Timestamp = t.Unix()
				}
			case int64:
				msg.Timestamp = v
			}
		}

		// Reconstruir From/To basado en IsFromMe y JID
		if msg.IsFromMe {
			msg.From = "me"
			msg.To = jidDB
		} else {
			msg.From = jidDB
			msg.To = "me"
		}

		messages = append(messages, msg)
	}

	// Invertir orden para mostrar cronológicamente (antiguos primero)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetChatsWithMessages obtiene la lista de chats que tienen mensajes, ordenados por último mensaje
func (r *MessageRepository) GetChatsWithMessages(ctx context.Context, instanceID string) ([]models.Chat, error) {
	query := `
		SELECT 
			jid,
			COUNT(*) as message_count,
			MAX(timestamp) as last_message_time,
			(SELECT push_name FROM messages m2 WHERE m2.instance_id = messages.instance_id AND m2.jid = messages.jid AND push_name IS NOT NULL AND push_name != '' ORDER BY m2.timestamp DESC LIMIT 1) as push_name,
			(SELECT content FROM messages m3 WHERE m3.instance_id = messages.instance_id AND m3.jid = messages.jid ORDER BY m3.timestamp DESC LIMIT 1) as last_message
		FROM messages
		WHERE instance_id = $1
		GROUP BY jid
		ORDER BY last_message_time DESC
	`

	if r.db.Driver == "sqlite" || r.db.Driver == "sqlite3" {
		query = `
			SELECT 
				jid,
				COUNT(*) as message_count,
				MAX(timestamp) as last_message_time,
				(SELECT push_name FROM messages m2 WHERE m2.instance_id = messages.instance_id AND m2.jid = messages.jid AND push_name IS NOT NULL AND push_name != '' ORDER BY m2.timestamp DESC LIMIT 1) as push_name,
				(SELECT content FROM messages m3 WHERE m3.instance_id = messages.instance_id AND m3.jid = messages.jid ORDER BY m3.timestamp DESC LIMIT 1) as last_message
			FROM messages
			WHERE instance_id = ?
			GROUP BY jid
			ORDER BY last_message_time DESC
		`
	}

	rows, err := r.db.DB.QueryContext(ctx, query, instanceID)
	if err != nil {
		return nil, fmt.Errorf("error querying chats: %w", err)
	}
	defer rows.Close()

	var chats []models.Chat
	for rows.Next() {
		var chat models.Chat
		var messageCount int
		var lastMessageTime interface{}
		var pushName sql.NullString
		var lastMessageContent sql.NullString

		if err := rows.Scan(&chat.JID, &messageCount, &lastMessageTime, &pushName, &lastMessageContent); err != nil {
			// Loguear el error pero continuar o retornar
			return nil, fmt.Errorf("error scanning chat (jid=%s): %w", chat.JID, err)
		}

		if pushName.Valid {
			chat.Name = pushName.String
		}

		if lastMessageContent.Valid {
			chat.LastMessage = lastMessageContent.String
		}

		if lastMessageTime != nil {
			switch v := lastMessageTime.(type) {
			case time.Time:
				chat.LastMessageTime = v.Unix()
			case string:
				// Intentar parsear formatos comunes de SQLite
				if t, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", v); err == nil {
					chat.LastMessageTime = t.Unix()
				} else if t, err := time.Parse(time.RFC3339, v); err == nil {
					chat.LastMessageTime = t.Unix()
				} else if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
					chat.LastMessageTime = t.Unix()
				}
			case []byte:
				s := string(v)
				if t, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", s); err == nil {
					chat.LastMessageTime = t.Unix()
				} else if t, err := time.Parse(time.RFC3339, s); err == nil {
					chat.LastMessageTime = t.Unix()
				} else if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
					chat.LastMessageTime = t.Unix()
				}
			case int64:
				chat.LastMessageTime = v
			}
		}
		chats = append(chats, chat)
	}

	return chats, nil
}

// DeleteChatMessages elimina todos los mensajes de un chat
func (r *MessageRepository) DeleteChatMessages(ctx context.Context, instanceID, jid string) error {
	query := `DELETE FROM messages WHERE instance_id = $1 AND jid = $2`

	if r.db.Driver == "sqlite" || r.db.Driver == "sqlite3" {
		query = `DELETE FROM messages WHERE instance_id = ? AND jid = ?`
	}

	_, err := r.db.DB.ExecContext(ctx, query, instanceID, jid)
	if err != nil {
		return fmt.Errorf("error deleting chat messages: %w", err)
	}

	return nil
}
