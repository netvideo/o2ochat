package storage

import "errors"

var (
	ErrStorageNotInitialized = errors.New("storage not initialized")
	ErrStorageClosed         = errors.New("storage is closed")
	ErrKeyNotFound           = errors.New("key not found")
	ErrKeyExists             = errors.New("key already exists")
	ErrInvalidKey            = errors.New("invalid key")
	ErrInvalidValue          = errors.New("invalid value")
	ErrStorageFull           = errors.New("storage is full")
	ErrCompressionFailed     = errors.New("compression failed")
	ErrDecompressionFailed   = errors.New("decompression failed")
	ErrEncryptionFailed      = errors.New("encryption failed")
	ErrDecryptionFailed      = errors.New("decryption failed")
	ErrBackupFailed          = errors.New("backup failed")
	ErrRestoreFailed         = errors.New("restore failed")
	ErrInvalidConfig         = errors.New("invalid config")
	ErrInvalidPath           = errors.New("invalid path")
	ErrMessageNotFound       = errors.New("message not found")
	ErrChunkNotFound         = errors.New("chunk not found")
	ErrConfigNotFound        = errors.New("config not found")
	ErrCacheNotFound         = errors.New("cache not found")
	ErrInvalidTTL            = errors.New("invalid TTL")
	ErrDatabaseCorrupted     = errors.New("database corrupted")
	ErrMigrationFailed       = errors.New("migration failed")
	ErrInvalidSchemaVersion  = errors.New("invalid schema version")
	ErrTransactionFailed     = errors.New("transaction failed")
)

type StorageError struct {
	Code    string
	Message string
	Err     error
}

func (e *StorageError) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *StorageError) Unwrap() error {
	return e.Err
}

func NewStorageError(code, message string, err error) *StorageError {
	return &StorageError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
