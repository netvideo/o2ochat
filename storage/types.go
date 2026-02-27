package storage

import (
	"time"
)

type StorageType string

const (
	StorageTypeSQLite  StorageType = "sqlite"
	StorageTypeBoltDB  StorageType = "boltdb"
	StorageTypeLevelDB StorageType = "leveldb"
	StorageTypeBadger  StorageType = "badger"
	StorageTypeMemory  StorageType = "memory"
)

type StorageConfig struct {
	Type        StorageType `json:"type"`
	Path        string      `json:"path"`
	MaxSize     int64       `json:"max_size"`
	Compression bool        `json:"compression"`
	Encryption  bool        `json:"encryption"`
	CacheSize   int         `json:"cache_size"`
}

type StorageStats struct {
	TotalSize    int64     `json:"total_size"`
	UsedSize     int64     `json:"used_size"`
	FreeSize     int64     `json:"free_size"`
	FileCount    int64     `json:"file_count"`
	ChunkCount   int64     `json:"chunk_count"`
	MessageCount int64     `json:"message_count"`
	LastBackup   time.Time `json:"last_backup"`
}

type StorageOptions struct {
	TTL         time.Duration `json:"ttl"`
	Compression bool         `json:"compression"`
	Encryption  bool         `json:"encryption"`
	Priority    int          `json:"priority"`
}

type StorageEntry struct {
	Key        string            `json:"key"`
	Value      []byte            `json:"value"`
	Metadata   map[string]string `json:"metadata"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	ExpiresAt  time.Time         `json:"expires_at"`
	Size       int               `json:"size"`
}

type MessageType string

const (
	MessageTypeText    MessageType = "text"
	MessageTypeImage  MessageType = "image"
	MessageTypeFile   MessageType = "file"
	MessageTypeVoice  MessageType = "voice"
	MessageTypeVideo  MessageType = "video"
	MessageTypeSystem MessageType = "system"
)

type ChatMessage struct {
	ID        string       `json:"id"`
	From      string       `json:"from"`
	To        string       `json:"to"`
	Content   []byte       `json:"content"`
	Type      MessageType `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Delivered bool        `json:"delivered"`
	Read      bool        `json:"read"`
	Encrypted bool        `json:"encrypted"`
}

type MessageStats struct {
	TotalCount     int64   `json:"total_count"`
	UnreadCount    int64   `json:"unread_count"`
	TotalSize      int64   `json:"total_size"`
	LastMessageAt  *time.Time `json:"last_message_at"`
}

type ChunkStats struct {
	TotalCount    int64 `json:"total_count"`
	TotalSize     int64 `json:"total_size"`
	ChunkCount    int   `json:"chunk_count"`
}

type CacheStats struct {
	Size          int   `json:"size"`
	Items         int   `json:"items"`
	Hits          int64 `json:"hits"`
	Misses        int64 `json:"misses"`
	Evictions     int64 `json:"evictions"`
}

func DefaultStorageConfig() *StorageConfig {
	return &StorageConfig{
		Type:        StorageTypeSQLite,
		Path:        "./data",
		MaxSize:     10 * 1024 * 1024 * 1024,
		Compression: true,
		Encryption:  true,
		CacheSize:   256,
	}
}

func DefaultStorageOptions() *StorageOptions {
	return &StorageOptions{
		TTL:         24 * time.Hour,
		Compression: true,
		Encryption:  true,
		Priority:    0,
	}
}
