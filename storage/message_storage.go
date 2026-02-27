package storage

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteMessageStorage implements MessageStorage using SQLite as the backend.
type SQLiteMessageStorage struct {
	db     *sql.DB
	mu     sync.RWMutex
	closed bool
}

type MessageStorageConfig struct {
	MaxMessagesPerConversation int           `json:"max_messages_per_conversation"`
	CleanupInterval            time.Duration `json:"cleanup_interval"`
	MessageTTL                 time.Duration `json:"message_ttl"`
}

func NewSQLiteMessageStorage(db *sql.DB) *SQLiteMessageStorage {
	return &SQLiteMessageStorage{
		db:     db,
		closed: false,
	}
}

func (m *SQLiteMessageStorage) StoreMessage(message *ChatMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrStorageNotInitialized
	}

	if message == nil {
		return ErrInvalidValue
	}

	contentStr := string(message.Content)

	query := `
		INSERT OR REPLACE INTO messages 
		(id, from_peer, to_peer, content, type, timestamp, delivered, read, encrypted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := m.db.Exec(query,
		message.ID,
		message.From,
		message.To,
		contentStr,
		string(message.Type),
		message.Timestamp.Unix(),
		boolToInt(message.Delivered),
		boolToInt(message.Read),
		boolToInt(message.Encrypted),
	)

	if err != nil {
		return NewStorageError("STORE_MESSAGE", "failed to store message", err)
	}

	return nil
}

func (m *SQLiteMessageStorage) GetMessage(messageID string) (*ChatMessage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, ErrStorageNotInitialized
	}

	query := `
		SELECT id, from_peer, to_peer, content, type, timestamp, delivered, read, encrypted
		FROM messages
		WHERE id = ?
	`

	var msg ChatMessage
	var contentStr string
	var timestamp int64
	var delivered, read, encrypted int

	err := m.db.QueryRow(query, messageID).Scan(
		&msg.ID,
		&msg.From,
		&msg.To,
		&contentStr,
		&msg.Type,
		&timestamp,
		&delivered,
		&read,
		&encrypted,
	)

	if err == sql.ErrNoRows {
		return nil, ErrMessageNotFound
	}
	if err != nil {
		return nil, NewStorageError("GET_MESSAGE", "failed to get message", err)
	}

	msg.Content = []byte(contentStr)

	msg.Timestamp = time.Unix(timestamp, 0)
	msg.Delivered = intToBool(delivered)
	msg.Read = intToBool(read)
	msg.Encrypted = intToBool(encrypted)

	return &msg, nil
}

func (m *SQLiteMessageStorage) GetConversationMessages(peerID string, limit int, offset int) ([]*ChatMessage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, ErrStorageNotInitialized
	}

	query := `
		SELECT id, from_peer, to_peer, content, type, timestamp, delivered, read, encrypted
		FROM messages
		WHERE from_peer = ? OR to_peer = ?
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`

	rows, err := m.db.Query(query, peerID, peerID, limit, offset)
	if err != nil {
		return nil, NewStorageError("GET_CONVERSATION", "failed to get messages", err)
	}
	defer rows.Close()

	messages, err := scanMessages(rows)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (m *SQLiteMessageStorage) SearchMessages(queryStr string, peerID string, limit int) ([]*ChatMessage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, ErrStorageNotInitialized
	}

	query := `
		SELECT id, from_peer, to_peer, content, type, timestamp, delivered, read, encrypted
		FROM messages
		WHERE (from_peer = ? OR to_peer = ?) AND content LIKE ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	searchPattern := "%" + queryStr + "%"
	rows, err := m.db.Query(query, peerID, peerID, searchPattern, limit)
	if err != nil {
		return nil, NewStorageError("SEARCH_MESSAGES", "failed to search messages", err)
	}
	defer rows.Close()

	messages, err := scanMessages(rows)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (m *SQLiteMessageStorage) DeleteMessage(messageID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrStorageNotInitialized
	}

	query := `DELETE FROM messages WHERE id = ?`
	result, err := m.db.Exec(query, messageID)
	if err != nil {
		return NewStorageError("DELETE_MESSAGE", "failed to delete message", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewStorageError("DELETE_MESSAGE", "failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return ErrMessageNotFound
	}

	return nil
}

func (m *SQLiteMessageStorage) CleanupOldMessages(before time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrStorageNotInitialized
	}

	query := `DELETE FROM messages WHERE timestamp < ?`
	_, err := m.db.Exec(query, before.Unix())
	if err != nil {
		return NewStorageError("CLEANUP_MESSAGES", "failed to cleanup old messages", err)
	}

	return nil
}

func (m *SQLiteMessageStorage) GetMessageStats(peerID string) (*MessageStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, ErrStorageNotInitialized
	}

	stats := &MessageStats{}

	totalQuery := `SELECT COUNT(*), SUM(length(content)), COALESCE(MAX(timestamp), 0) FROM messages WHERE from_peer = ? OR to_peer = ?`
	var maxTimestamp int64
	err := m.db.QueryRow(totalQuery, peerID, peerID).Scan(&stats.TotalCount, &stats.TotalSize, &maxTimestamp)
	if err != nil && err != sql.ErrNoRows {
		return nil, NewStorageError("GET_STATS", "failed to get total count", err)
	}

	if maxTimestamp > 0 {
		t := time.Unix(maxTimestamp, 0)
		stats.LastMessageAt = &t
	}

	unreadQuery := `SELECT COUNT(*) FROM messages WHERE (from_peer = ? OR to_peer = ?) AND read = 0`
	err = m.db.QueryRow(unreadQuery, peerID, peerID).Scan(&stats.UnreadCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, NewStorageError("GET_STATS", "failed to get unread count", err)
	}

	return stats, nil
}

func (m *SQLiteMessageStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}

	m.closed = true
	if err := m.db.Close(); err != nil {
		return NewStorageError("CLOSE", "failed to close message storage", err)
	}

	return nil
}

func scanMessages(rows *sql.Rows) ([]*ChatMessage, error) {
	messages := make([]*ChatMessage, 0)

	for rows.Next() {
		var msg ChatMessage
		var contentStr string
		var timestamp int64
		var delivered, read, encrypted int

		err := rows.Scan(
			&msg.ID,
			&msg.From,
			&msg.To,
			&contentStr,
			&msg.Type,
			&timestamp,
			&delivered,
			&read,
			&encrypted,
		)

		if err != nil {
			return nil, NewStorageError("SCAN_MESSAGE", "failed to scan message", err)
		}

		msg.Content = []byte(contentStr)

		msg.Timestamp = time.Unix(timestamp, 0)
		msg.Delivered = intToBool(delivered)
		msg.Read = intToBool(read)
		msg.Encrypted = intToBool(encrypted)

		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, NewStorageError("SCAN_MESSAGE", "rows iteration error", err)
	}

	return messages, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i != 0
}
