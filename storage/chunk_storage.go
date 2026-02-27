package storage

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteChunkStorage implements ChunkStorage using SQLite for metadata and filesystem for chunk data.
type SQLiteChunkStorage struct {
	db           *sql.DB
	chunkDir     string
	mu           sync.RWMutex
	closed       bool
	maxChunkSize int64
}

type ChunkStorageConfig struct {
	ChunkDir     string        `json:"chunk_dir"`
	MaxChunkSize int64         `json:"max_chunk_size"`
	TTL          time.Duration `json:"ttl"`
}

func NewSQLiteChunkStorage(db *sql.DB, chunkDir string) (*SQLiteChunkStorage, error) {
	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		return nil, NewStorageError("CREATE_CHUNK_DIR", "failed to create chunk directory", err)
	}

	return &SQLiteChunkStorage{
		db:           db,
		chunkDir:     chunkDir,
		closed:       false,
		maxChunkSize: 1024 * 1024,
	}, nil
}

func (c *SQLiteChunkStorage) StoreChunk(fileID string, index int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrStorageNotInitialized
	}

	if fileID == "" {
		return ErrInvalidKey
	}

	if index < 0 {
		return ErrInvalidValue
	}

	hash := sha256.Sum256(data)
	hashStr := string(hash[:])

	chunkPath := c.getChunkPath(fileID, index)
	if err := os.WriteFile(chunkPath, data, 0644); err != nil {
		return NewStorageError("STORE_CHUNK", "failed to write chunk file", err)
	}

	query := `
		INSERT OR REPLACE INTO chunks 
		(file_id, chunk_index, data, hash, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	var expiresAt *int64
	_, err := c.db.Exec(query, fileID, index, hashStr, hashStr, time.Now().Unix(), expiresAt)
	if err != nil {
		os.Remove(chunkPath)
		return NewStorageError("STORE_CHUNK", "failed to store chunk metadata", err)
	}

	return nil
}

func (c *SQLiteChunkStorage) GetChunk(fileID string, index int) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrStorageNotInitialized
	}

	chunkPath := c.getChunkPath(fileID, index)

	data, err := os.ReadFile(chunkPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrChunkNotFound
		}
		return nil, NewStorageError("GET_CHUNK", "failed to read chunk file", err)
	}

	query := `SELECT hash FROM chunks WHERE file_id = ? AND chunk_index = ?`
	var storedHash string
	err = c.db.QueryRow(query, fileID, index).Scan(&storedHash)
	if err == sql.ErrNoRows {
		return nil, ErrChunkNotFound
	}
	if err != nil {
		return nil, NewStorageError("GET_CHUNK", "failed to get chunk metadata", err)
	}

	computedHash := sha256.Sum256(data)
	if string(computedHash[:]) != storedHash {
		return nil, NewStorageError("GET_CHUNK", "chunk hash mismatch", nil)
	}

	return data, nil
}

func (c *SQLiteChunkStorage) ChunkExists(fileID string, index int) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return false, ErrStorageNotInitialized
	}

	query := `SELECT COUNT(*) FROM chunks WHERE file_id = ? AND chunk_index = ?`
	var count int
	err := c.db.QueryRow(query, fileID, index).Scan(&count)
	if err != nil {
		return false, NewStorageError("CHUNK_EXISTS", "failed to check chunk existence", err)
	}

	return count > 0, nil
}

func (c *SQLiteChunkStorage) GetChunkIndices(fileID string) ([]int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrStorageNotInitialized
	}

	query := `SELECT chunk_index FROM chunks WHERE file_id = ? ORDER BY chunk_index`
	rows, err := c.db.Query(query, fileID)
	if err != nil {
		return nil, NewStorageError("GET_CHUNK_INDICES", "failed to get chunk indices", err)
	}
	defer rows.Close()

	indices := make([]int, 0)
	for rows.Next() {
		var index int
		if err := rows.Scan(&index); err != nil {
			return nil, NewStorageError("GET_CHUNK_INDICES", "failed to scan index", err)
		}
		indices = append(indices, index)
	}

	if err := rows.Err(); err != nil {
		return nil, NewStorageError("GET_CHUNK_INDICES", "rows iteration error", err)
	}

	return indices, nil
}

func (c *SQLiteChunkStorage) DeleteChunk(fileID string, index int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrStorageNotInitialized
	}

	query := `DELETE FROM chunks WHERE file_id = ? AND chunk_index = ?`
	result, err := c.db.Exec(query, fileID, index)
	if err != nil {
		return NewStorageError("DELETE_CHUNK", "failed to delete chunk metadata", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewStorageError("DELETE_CHUNK", "failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return ErrChunkNotFound
	}

	chunkPath := c.getChunkPath(fileID, index)
	if err := os.Remove(chunkPath); err != nil && !os.IsNotExist(err) {
		return NewStorageError("DELETE_CHUNK", "failed to delete chunk file", err)
	}

	return nil
}

func (c *SQLiteChunkStorage) DeleteAllChunks(fileID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrStorageNotInitialized
	}

	indices, err := c.getChunkIndicesLocked(fileID)
	if err != nil {
		return err
	}

	query := `DELETE FROM chunks WHERE file_id = ?`
	_, err = c.db.Exec(query, fileID)
	if err != nil {
		return NewStorageError("DELETE_ALL_CHUNKS", "failed to delete chunk metadata", err)
	}

	for _, index := range indices {
		chunkPath := c.getChunkPath(fileID, index)
		os.Remove(chunkPath)
	}

	return nil
}

func (c *SQLiteChunkStorage) GetChunkStats(fileID string) (*ChunkStats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrStorageNotInitialized
	}

	stats := &ChunkStats{}

	query := `SELECT COUNT(*), SUM(length(data)) FROM chunks WHERE file_id = ?`
	err := c.db.QueryRow(query, fileID).Scan(&stats.ChunkCount, &stats.TotalSize)
	if err != nil && err != sql.ErrNoRows {
		return nil, NewStorageError("GET_CHUNK_STATS", "failed to get chunk stats", err)
	}

	stats.TotalCount = int64(stats.ChunkCount)

	return stats, nil
}

func (c *SQLiteChunkStorage) CleanupTemporaryChunks() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrStorageNotInitialized
	}

	query := `SELECT file_id, chunk_index FROM chunks WHERE expires_at IS NOT NULL AND expires_at < ?`
	rows, err := c.db.Query(query, time.Now().Unix())
	if err != nil {
		return NewStorageError("CLEANUP_CHUNKS", "failed to query expired chunks", err)
	}
	defer rows.Close()

	deleted := 0
	for rows.Next() {
		var fileID string
		var index int
		if err := rows.Scan(&fileID, &index); err != nil {
			continue
		}

		chunkPath := c.getChunkPath(fileID, index)
		if err := os.Remove(chunkPath); err == nil {
			deleted++
		}
	}

	deleteQuery := `DELETE FROM chunks WHERE expires_at IS NOT NULL AND expires_at < ?`
	_, err = c.db.Exec(deleteQuery, time.Now().Unix())
	if err != nil {
		return NewStorageError("CLEANUP_CHUNKS", "failed to delete expired chunks", err)
	}

	return nil
}

func (c *SQLiteChunkStorage) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return nil
}

func (c *SQLiteChunkStorage) getChunkPath(fileID string, index int) string {
	safeFileID := filepath.Base(fileID)
	dirPath := filepath.Join(c.chunkDir, safeFileID)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return ""
	}
	return filepath.Join(dirPath, fmt.Sprintf("chunk_%d", index))
}

func (c *SQLiteChunkStorage) getChunkIndicesLocked(fileID string) ([]int, error) {
	query := `SELECT chunk_index FROM chunks WHERE file_id = ? ORDER BY chunk_index`
	rows, err := c.db.Query(query, fileID)
	if err != nil {
		return nil, NewStorageError("GET_CHUNK_INDICES", "failed to get chunk indices", err)
	}
	defer rows.Close()

	indices := make([]int, 0)
	for rows.Next() {
		var index int
		if err := rows.Scan(&index); err != nil {
			return nil, NewStorageError("GET_CHUNK_INDICES", "failed to scan index", err)
		}
		indices = append(indices, index)
	}

	if err := rows.Err(); err != nil {
		return nil, NewStorageError("GET_CHUNK_INDICES", "rows iteration error", err)
	}

	return indices, nil
}
