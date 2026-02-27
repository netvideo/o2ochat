package storage

import "time"

type StorageManager interface {
	Initialize(config *StorageConfig) error
	Put(key string, value []byte, options *StorageOptions) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	Exists(key string) (bool, error)
	List(prefix string) ([]string, error)
	BatchPut(entries map[string][]byte, options *StorageOptions) error
	BatchGet(keys []string) (map[string][]byte, error)
	BatchDelete(keys []string) error
	GetStats() (*StorageStats, error)
	CleanupExpired() error
	Backup(backupPath string) error
	Restore(backupPath string) error
	Close() error
}

type MessageStorage interface {
	StoreMessage(message *ChatMessage) error
	GetMessage(messageID string) (*ChatMessage, error)
	GetConversationMessages(peerID string, limit int, offset int) ([]*ChatMessage, error)
	SearchMessages(query string, peerID string, limit int) ([]*ChatMessage, error)
	DeleteMessage(messageID string) error
	CleanupOldMessages(before time.Time) error
	GetMessageStats(peerID string) (*MessageStats, error)
}

type ChunkStorage interface {
	StoreChunk(fileID string, index int, data []byte) error
	GetChunk(fileID string, index int) ([]byte, error)
	ChunkExists(fileID string, index int) (bool, error)
	GetChunkIndices(fileID string) ([]int, error)
	DeleteChunk(fileID string, index int) error
	DeleteAllChunks(fileID string) error
	GetChunkStats(fileID string) (*ChunkStats, error)
	CleanupTemporaryChunks() error
	Close() error
}

type ConfigStorage interface {
	StoreConfig(section string, key string, value interface{}) error
	GetConfig(section string, key string, defaultValue interface{}) (interface{}, error)
	DeleteConfig(section string, key string) error
	GetAllConfig(section string) (map[string]interface{}, error)
	ImportConfig(configPath string) error
	ExportConfig(configPath string) error
	ResetConfig() error
}

type CacheManager interface {
	Set(key string, value []byte, ttl time.Duration) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	Clear() error
	GetCacheStats() (*CacheStats, error)
	Resize(newSize int) error
}
