package storage

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteConfigStorage implements ConfigStorage using SQLite as the backend.
type SQLiteConfigStorage struct {
	db       *sql.DB
	configMu sync.RWMutex
	closed   bool
}

func NewSQLiteConfigStorage(db *sql.DB) *SQLiteConfigStorage {
	return &SQLiteConfigStorage{
		db:     db,
		closed: false,
	}
}

func (c *SQLiteConfigStorage) StoreConfig(section string, key string, value interface{}) error {
	c.configMu.Lock()
	defer c.configMu.Unlock()

	if c.closed {
		return ErrStorageNotInitialized
	}

	if section == "" || key == "" {
		return ErrInvalidKey
	}

	valueBytes, err := json.Marshal(value)
	if err != nil {
		return NewStorageError("STORE_CONFIG", "failed to marshal value", err)
	}

	query := `
		INSERT OR REPLACE INTO config 
		(section, key, value, updated_at)
		VALUES (?, ?, ?, ?)
	`

	_, err = c.db.Exec(query, section, key, valueBytes, time.Now().Unix())
	if err != nil {
		return NewStorageError("STORE_CONFIG", "failed to store config", err)
	}

	return nil
}

func (c *SQLiteConfigStorage) GetConfig(section string, key string, defaultValue interface{}) (interface{}, error) {
	c.configMu.RLock()
	defer c.configMu.RUnlock()

	if c.closed {
		return nil, ErrStorageNotInitialized
	}

	query := `SELECT value FROM config WHERE section = ? AND key = ?`
	var valueBytes []byte
	err := c.db.QueryRow(query, section, key).Scan(&valueBytes)
	if err == sql.ErrNoRows {
		return defaultValue, ErrConfigNotFound
	}
	if err != nil {
		return nil, NewStorageError("GET_CONFIG", "failed to get config", err)
	}

	var value interface{}
	if err := json.Unmarshal(valueBytes, &value); err != nil {
		return nil, NewStorageError("GET_CONFIG", "failed to unmarshal value", err)
	}

	return value, nil
}

func (c *SQLiteConfigStorage) DeleteConfig(section string, key string) error {
	c.configMu.Lock()
	defer c.configMu.Unlock()

	if c.closed {
		return ErrStorageNotInitialized
	}

	query := `DELETE FROM config WHERE section = ? AND key = ?`
	result, err := c.db.Exec(query, section, key)
	if err != nil {
		return NewStorageError("DELETE_CONFIG", "failed to delete config", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewStorageError("DELETE_CONFIG", "failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return ErrConfigNotFound
	}

	return nil
}

func (c *SQLiteConfigStorage) GetAllConfig(section string) (map[string]interface{}, error) {
	c.configMu.RLock()
	defer c.configMu.RUnlock()

	if c.closed {
		return nil, ErrStorageNotInitialized
	}

	query := `SELECT key, value FROM config WHERE section = ?`
	rows, err := c.db.Query(query, section)
	if err != nil {
		return nil, NewStorageError("GET_ALL_CONFIG", "failed to get config", err)
	}
	defer rows.Close()

	config := make(map[string]interface{})
	for rows.Next() {
		var key string
		var valueBytes []byte
		if err := rows.Scan(&key, &valueBytes); err != nil {
			return nil, NewStorageError("GET_ALL_CONFIG", "failed to scan config", err)
		}

		var value interface{}
		if err := json.Unmarshal(valueBytes, &value); err != nil {
			return nil, NewStorageError("GET_ALL_CONFIG", "failed to unmarshal value", err)
		}

		config[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, NewStorageError("GET_ALL_CONFIG", "rows iteration error", err)
	}

	return config, nil
}

func (c *SQLiteConfigStorage) ImportConfig(configPath string) error {
	c.configMu.Lock()
	defer c.configMu.Unlock()

	if c.closed {
		return ErrStorageNotInitialized
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return NewStorageError("IMPORT_CONFIG", "failed to read config file", err)
	}

	var configData map[string]map[string]interface{}
	if err := json.Unmarshal(data, &configData); err != nil {
		return NewStorageError("IMPORT_CONFIG", "failed to parse config file", err)
	}

	tx, err := c.db.Begin()
	if err != nil {
		return NewStorageError("IMPORT_CONFIG", "failed to begin transaction", err)
	}
	defer tx.Rollback()

	query := `
		INSERT OR REPLACE INTO config 
		(section, key, value, updated_at)
		VALUES (?, ?, ?, ?)
	`

	for section, keys := range configData {
		for key, value := range keys {
			valueBytes, err := json.Marshal(value)
			if err != nil {
				return NewStorageError("IMPORT_CONFIG", "failed to marshal value", err)
			}

			_, err = tx.Exec(query, section, key, valueBytes, time.Now().Unix())
			if err != nil {
				return NewStorageError("IMPORT_CONFIG", "failed to store config", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return NewStorageError("IMPORT_CONFIG", "failed to commit transaction", err)
	}

	return nil
}

func (c *SQLiteConfigStorage) ExportConfig(configPath string) error {
	c.configMu.RLock()
	defer c.configMu.RUnlock()

	if c.closed {
		return ErrStorageNotInitialized
	}

	query := `SELECT section, key, value FROM config`
	rows, err := c.db.Query(query)
	if err != nil {
		return NewStorageError("EXPORT_CONFIG", "failed to get config", err)
	}
	defer rows.Close()

	configData := make(map[string]map[string]interface{})
	for rows.Next() {
		var section, key string
		var valueBytes []byte
		if err := rows.Scan(&section, &key, &valueBytes); err != nil {
			return NewStorageError("EXPORT_CONFIG", "failed to scan config", err)
		}

		if _, exists := configData[section]; !exists {
			configData[section] = make(map[string]interface{})
		}

		var value interface{}
		if err := json.Unmarshal(valueBytes, &value); err != nil {
			return NewStorageError("EXPORT_CONFIG", "failed to unmarshal value", err)
		}

		configData[section][key] = value
	}

	if err := rows.Err(); err != nil {
		return NewStorageError("EXPORT_CONFIG", "rows iteration error", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return NewStorageError("EXPORT_CONFIG", "failed to create directory", err)
	}

	data, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return NewStorageError("EXPORT_CONFIG", "failed to marshal config", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return NewStorageError("EXPORT_CONFIG", "failed to write config file", err)
	}

	return nil
}

func (c *SQLiteConfigStorage) ResetConfig() error {
	c.configMu.Lock()
	defer c.configMu.Unlock()

	if c.closed {
		return ErrStorageNotInitialized
	}

	query := `DELETE FROM config`
	_, err := c.db.Exec(query)
	if err != nil {
		return NewStorageError("RESET_CONFIG", "failed to reset config", err)
	}

	return nil
}

func (c *SQLiteConfigStorage) Close() error {
	c.configMu.Lock()
	defer c.configMu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return nil
}
