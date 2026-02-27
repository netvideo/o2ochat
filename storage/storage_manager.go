package storage

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorageManager is the main storage manager that coordinates all storage components.
// It provides a unified interface for message, chunk, config, and cache storage.
type SQLiteStorageManager struct {
	config         *StorageConfig
	db             *sql.DB
	messageStorage *SQLiteMessageStorage
	chunkStorage   *SQLiteChunkStorage
	configStorage  *SQLiteConfigStorage
	cacheManager   *LRUCacheManager
	mu             sync.RWMutex
	initialized    bool
	closed         bool
}

// NewSQLiteStorageManager creates a new SQLiteStorageManager instance.
func NewSQLiteStorageManager() *SQLiteStorageManager {
	return &SQLiteStorageManager{}
}

// Initialize initializes the storage manager with the given configuration.
// It creates the database, initializes all storage components, and sets up the schema.
func (s *SQLiteStorageManager) Initialize(config *StorageConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.initialized {
		return nil
	}

	s.config = config

	if err := os.MkdirAll(config.Path, 0755); err != nil {
		return NewStorageError("INITIALIZE", "failed to create data directory", err)
	}

	dbPath := filepath.Join(config.Path, "o2ochat.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return NewStorageError("INITIALIZE", "failed to open database", err)
	}

	if err := s.initializeDatabase(db); err != nil {
		db.Close()
		return err
	}

	s.db = db

	s.messageStorage = NewSQLiteMessageStorage(db)

	s.configStorage = NewSQLiteConfigStorage(db)

	chunkDir := filepath.Join(config.Path, "chunks")
	chunkStorage, err := NewSQLiteChunkStorage(db, chunkDir)
	if err != nil {
		db.Close()
		return err
	}
	s.chunkStorage = chunkStorage

	s.cacheManager = NewLRUCacheManager(config.CacheSize * 1024 * 1024)

	s.initialized = true

	return nil
}

func (s *SQLiteStorageManager) initializeDatabase(db *sql.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			from_peer TEXT NOT NULL,
			to_peer TEXT NOT NULL,
			content BLOB NOT NULL,
			type TEXT NOT NULL,
			timestamp INTEGER NOT NULL,
			delivered INTEGER DEFAULT 0,
			read INTEGER DEFAULT 0,
			encrypted INTEGER DEFAULT 1
		);

		CREATE TABLE IF NOT EXISTS chunks (
			file_id TEXT NOT NULL,
			chunk_index INTEGER NOT NULL,
			data BLOB NOT NULL,
			hash TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			expires_at INTEGER,
			PRIMARY KEY (file_id, chunk_index)
		);

		CREATE TABLE IF NOT EXISTS config (
			section TEXT NOT NULL,
			key TEXT NOT NULL,
			value BLOB NOT NULL,
			updated_at INTEGER NOT NULL,
			PRIMARY KEY (section, key)
		);

		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			timestamp INTEGER NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(from_peer, to_peer, timestamp);
		CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
		CREATE INDEX IF NOT EXISTS idx_chunks_file ON chunks(file_id);
		CREATE INDEX IF NOT EXISTS idx_chunks_expires ON chunks(expires_at) WHERE expires_at IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_config_section ON config(section);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return NewStorageError("INITIALIZE", "failed to create schema", err)
	}

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM schema_version`).Scan(&count)
	if err != nil {
		_, err = db.Exec(`INSERT INTO schema_version (version, timestamp) VALUES (?, ?)`, 1, time.Now().Unix())
		if err != nil {
			return NewStorageError("INITIALIZE", "failed to insert schema version", err)
		}
	}

	return nil
}

func (s *SQLiteStorageManager) Put(key string, value []byte, options *StorageOptions) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return ErrStorageNotInitialized
	}

	if s.closed {
		return ErrStorageNotInitialized
	}

	if options != nil && options.TTL > 0 {
		if err := s.cacheManager.Set(key, value, options.TTL); err != nil {
			return err
		}
	}

	section, k := s.parseKey(key)
	return s.configStorage.StoreConfig(section, k, value)
}

func (s *SQLiteStorageManager) Get(key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrStorageNotInitialized
	}

	if s.closed {
		return nil, ErrStorageNotInitialized
	}

	if value, err := s.cacheManager.Get(key); err == nil {
		return value, nil
	}

	section, k := s.parseKey(key)
	value, err := s.configStorage.GetConfig(section, k, nil)
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, ErrKeyNotFound
	}

	if str, ok := value.(string); ok {
		if decoded, err := base64.StdEncoding.DecodeString(str); err == nil {
			return decoded, nil
		}
		return []byte(str), nil
	}

	valueBytes, ok := value.([]byte)
	if !ok {
		if bytes, err := interfaceToBytes(value); err == nil {
			return bytes, nil
		}
		return nil, ErrInvalidValue
	}

	return valueBytes, nil
}

func (s *SQLiteStorageManager) Delete(key string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return ErrStorageNotInitialized
	}

	if s.closed {
		return ErrStorageNotInitialized
	}

	s.cacheManager.Delete(key)

	section, k := s.parseKey(key)
	return s.configStorage.DeleteConfig(section, k)
}

func (s *SQLiteStorageManager) Exists(key string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return false, ErrStorageNotInitialized
	}

	if s.closed {
		return false, ErrStorageNotInitialized
	}

	if _, err := s.cacheManager.Get(key); err == nil {
		return true, nil
	}

	section, k := s.parseKey(key)
	_, err := s.configStorage.GetConfig(section, k, nil)
	if err == ErrConfigNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *SQLiteStorageManager) List(prefix string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrStorageNotInitialized
	}

	if s.closed {
		return nil, ErrStorageNotInitialized
	}

	section, keyPrefix := s.parseKey(prefix)
	if section == "" {
		section = "default"
	}

	config, err := s.configStorage.GetAllConfig(section)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(config))
	for k := range config {
		if keyPrefix == "" || strings.HasPrefix(k, keyPrefix) {
			keys = append(keys, section+":"+k)
		}
	}

	return keys, nil
}

func (s *SQLiteStorageManager) BatchPut(entries map[string][]byte, options *StorageOptions) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return ErrStorageNotInitialized
	}

	if s.closed {
		return ErrStorageNotInitialized
	}

	for key, value := range entries {
		if options != nil && options.TTL > 0 {
			if err := s.cacheManager.Set(key, value, options.TTL); err != nil {
				return err
			}
		}

		section, k := s.parseKey(key)
		if err := s.configStorage.StoreConfig(section, k, value); err != nil {
			return err
		}
	}

	return nil
}

func (s *SQLiteStorageManager) BatchGet(keys []string) (map[string][]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrStorageNotInitialized
	}

	if s.closed {
		return nil, ErrStorageNotInitialized
	}

	result := make(map[string][]byte)
	for _, key := range keys {
		value, err := s.Get(key)
		if err == nil {
			result[key] = value
		}
	}

	return result, nil
}

func (s *SQLiteStorageManager) BatchDelete(keys []string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return ErrStorageNotInitialized
	}

	if s.closed {
		return ErrStorageNotInitialized
	}

	for _, key := range keys {
		s.Delete(key)
	}

	return nil
}

func (s *SQLiteStorageManager) GetStats() (*StorageStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return nil, ErrStorageNotInitialized
	}

	if s.closed {
		return nil, ErrStorageNotInitialized
	}

	stats := &StorageStats{}

	dbPath := filepath.Join(s.config.Path, "o2ochat.db")
	fileInfo, err := os.Stat(dbPath)
	if err == nil {
		stats.TotalSize = fileInfo.Size()
		stats.UsedSize = fileInfo.Size()
	}

	var messageCount, chunkCount int64
	err = s.db.QueryRow(`SELECT COUNT(*) FROM messages`).Scan(&messageCount)
	if err == nil {
		stats.MessageCount = messageCount
	}

	err = s.db.QueryRow(`SELECT COUNT(*) FROM chunks`).Scan(&chunkCount)
	if err == nil {
		stats.ChunkCount = chunkCount
	}

	cacheStats, err := s.cacheManager.GetCacheStats()
	if err == nil {
		stats.FileCount = int64(cacheStats.Items)
	}

	return stats, nil
}

func (s *SQLiteStorageManager) CleanupExpired() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return ErrStorageNotInitialized
	}

	if s.closed {
		return ErrStorageNotInitialized
	}

	s.cacheManager.CleanupExpired()

	if err := s.chunkStorage.CleanupTemporaryChunks(); err != nil {
		return err
	}

	cleanupTime := time.Now().AddDate(0, -1, 0)
	if err := s.messageStorage.CleanupOldMessages(cleanupTime); err != nil {
		return err
	}

	return nil
}

func (s *SQLiteStorageManager) Backup(backupPath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return ErrStorageNotInitialized
	}

	if s.closed {
		return ErrStorageNotInitialized
	}

	dbPath := filepath.Join(s.config.Path, "o2ochat.db")
	migration := NewDataMigration(s.db, dbPath, backupPath)
	return migration.BackupDatabase(backupPath)
}

func (s *SQLiteStorageManager) Restore(backupPath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return ErrStorageNotInitialized
	}

	if s.closed {
		return ErrStorageNotInitialized
	}

	dbPath := filepath.Join(s.config.Path, "o2ochat.db")
	migration := NewDataMigration(s.db, backupPath, dbPath)
	return migration.RestoreDatabase(backupPath)
}

func (s *SQLiteStorageManager) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized || s.closed {
		return nil
	}

	s.closed = true

	if s.cacheManager != nil {
		s.cacheManager.Close()
	}

	if s.configStorage != nil {
		s.configStorage.Close()
	}

	if s.chunkStorage != nil {
		s.chunkStorage.Close()
	}

	if s.messageStorage != nil {
		s.messageStorage.Close()
	}

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return NewStorageError("CLOSE", "failed to close database", err)
		}
	}

	s.initialized = false

	return nil
}

func (s *SQLiteStorageManager) GetMessageStorage() MessageStorage {
	return s.messageStorage
}

func (s *SQLiteStorageManager) GetChunkStorage() ChunkStorage {
	return s.chunkStorage
}

func (s *SQLiteStorageManager) GetConfigStorage() ConfigStorage {
	return s.configStorage
}

func (s *SQLiteStorageManager) GetCacheManager() CacheManager {
	return s.cacheManager
}

func (s *SQLiteStorageManager) parseKey(key string) (section, k string) {
	// key format: "section:key" or "section:sub:key"
	parts := splitKey(key, 2)
	if len(parts) == 1 {
		return "default", parts[0]
	}
	return parts[0], parts[1]
}

func splitKey(s string, n int) []string {
	parts := make([]string, 0)
	start := 0
	for i := 0; i < len(s) && len(parts) < n-1; i++ {
		if s[i] == ':' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func interfaceToBytes(v interface{}) ([]byte, error) {
	switch val := v.(type) {
	case []byte:
		return val, nil
	case string:
		return []byte(val), nil
	default:
		return json.Marshal(v)
	}
}
