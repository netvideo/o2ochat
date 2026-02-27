package storage

import (
	"database/sql"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DataMigration provides functionality for migrating data between storage systems.
type DataMigration struct {
	db         *sql.DB
	sourcePath string
	targetPath string
}

// MigrationReport contains the results of a data migration operation.
type MigrationReport struct {
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	MessagesMigrated int64     `json:"messages_migrated"`
	ChunksMigrated   int64     `json:"chunks_migrated"`
	ConfigsMigrated  int64     `json:"configs_migrated"`
	Errors           []string  `json:"errors"`
	Success          bool      `json:"success"`
}

type SchemaVersion struct {
	Version   int       `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

func NewDataMigration(db *sql.DB, sourcePath, targetPath string) *DataMigration {
	return &DataMigration{
		db:         db,
		sourcePath: sourcePath,
		targetPath: targetPath,
	}
}

func (m *DataMigration) MigrateAll() (*MigrationReport, error) {
	report := &MigrationReport{
		StartTime: time.Now(),
		Errors:    make([]string, 0),
	}

	if err := m.migrateMessages(report); err != nil {
		report.Errors = append(report.Errors, "messages: "+err.Error())
	}

	if err := m.migrateChunks(report); err != nil {
		report.Errors = append(report.Errors, "chunks: "+err.Error())
	}

	if err := m.migrateConfig(report); err != nil {
		report.Errors = append(report.Errors, "config: "+err.Error())
	}

	report.EndTime = time.Now()
	report.Success = len(report.Errors) == 0

	return report, nil
}

func (m *DataMigration) migrateMessages(report *MigrationReport) error {
	query := `SELECT id, from_peer, to_peer, content, type, timestamp, delivered, read, encrypted FROM messages`
	rows, err := m.db.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	defer rows.Close()

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT OR REPLACE INTO messages 
		(id, from_peer, to_peer, content, type, timestamp, delivered, read, encrypted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	for rows.Next() {
		var msg ChatMessage
		var contentBytes []byte
		var timestamp int64
		var delivered, read, encrypted int

		if err := rows.Scan(
			&msg.ID, &msg.From, &msg.To, &contentBytes, &msg.Type,
			&timestamp, &delivered, &read, &encrypted,
		); err != nil {
			report.Errors = append(report.Errors, err.Error())
			continue
		}

		_, err := tx.Exec(insertQuery,
			msg.ID, msg.From, msg.To, contentBytes, msg.Type,
			timestamp, delivered, read, encrypted,
		)
		if err != nil {
			report.Errors = append(report.Errors, err.Error())
			continue
		}

		report.MessagesMigrated++
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (m *DataMigration) migrateChunks(report *MigrationReport) error {
	query := `SELECT file_id, chunk_index, data, hash, created_at, expires_at FROM chunks`
	rows, err := m.db.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	defer rows.Close()

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT OR REPLACE INTO chunks 
		(file_id, chunk_index, data, hash, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	for rows.Next() {
		var fileID string
		var index int
		var data, hash string
		var createdAt int64
		var expiresAt *int64

		if err := rows.Scan(&fileID, &index, &data, &hash, &createdAt, &expiresAt); err != nil {
			report.Errors = append(report.Errors, err.Error())
			continue
		}

		_, err := tx.Exec(insertQuery, fileID, index, data, hash, createdAt, expiresAt)
		if err != nil {
			report.Errors = append(report.Errors, err.Error())
			continue
		}

		report.ChunksMigrated++
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (m *DataMigration) migrateConfig(report *MigrationReport) error {
	query := `SELECT section, key, value, updated_at FROM config`
	rows, err := m.db.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	defer rows.Close()

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT OR REPLACE INTO config 
		(section, key, value, updated_at)
		VALUES (?, ?, ?, ?)
	`

	for rows.Next() {
		var section, key string
		var value []byte
		var updatedAt int64

		if err := rows.Scan(&section, &key, &value, &updatedAt); err != nil {
			report.Errors = append(report.Errors, err.Error())
			continue
		}

		_, err := tx.Exec(insertQuery, section, key, value, updatedAt)
		if err != nil {
			report.Errors = append(report.Errors, err.Error())
			continue
		}

		report.ConfigsMigrated++
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (m *DataMigration) BackupDatabase(backupPath string) error {
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		return NewStorageError("BACKUP", "failed to create backup directory", err)
	}

	srcFile, err := os.Open(m.sourcePath)
	if err != nil {
		return NewStorageError("BACKUP", "failed to open source database", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(backupPath)
	if err != nil {
		return NewStorageError("BACKUP", "failed to create backup file", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return NewStorageError("BACKUP", "failed to copy database", err)
	}

	return nil
}

func (m *DataMigration) RestoreDatabase(backupPath string) error {
	srcFile, err := os.Open(backupPath)
	if err != nil {
		return NewStorageError("RESTORE", "failed to open backup file", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(m.targetPath)
	if err != nil {
		return NewStorageError("RESTORE", "failed to create database file", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return NewStorageError("RESTORE", "failed to restore database", err)
	}

	return nil
}

func (m *DataMigration) GetSchemaVersion() (*SchemaVersion, error) {
	query := `SELECT version, timestamp FROM schema_version ORDER BY timestamp DESC LIMIT 1`
	var version SchemaVersion
	var timestamp int64

	err := m.db.QueryRow(query).Scan(&version.Version, &timestamp)
	if err == sql.ErrNoRows {
		return &SchemaVersion{Version: 0, Timestamp: time.Time{}}, nil
	}
	if err != nil {
		return nil, NewStorageError("GET_VERSION", "failed to get schema version", err)
	}

	version.Timestamp = time.Unix(timestamp, 0)
	return &version, nil
}

func (m *DataMigration) UpdateSchemaVersion(version int) error {
	query := `INSERT INTO schema_version (version, timestamp) VALUES (?, ?)`
	_, err := m.db.Exec(query, version, time.Now().Unix())
	if err != nil {
		return NewStorageError("UPDATE_VERSION", "failed to update schema version", err)
	}

	return nil
}

func ExportDataToJSON(db *sql.DB, outputPath string) error {
	data := make(map[string]interface{})

	messages, err := exportTableToJSON(db, "messages")
	if err != nil {
		return err
	}
	data["messages"] = messages

	chunks, err := exportTableToJSON(db, "chunks")
	if err != nil {
		return err
	}
	data["chunks"] = chunks

	configs, err := exportTableToJSON(db, "config")
	if err != nil {
		return err
	}
	data["config"] = configs

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return NewStorageError("EXPORT", "failed to marshal data", err)
	}

	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return NewStorageError("EXPORT", "failed to write export file", err)
	}

	return nil
}

func ImportDataFromJSON(db *sql.DB, inputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return NewStorageError("IMPORT", "failed to read import file", err)
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return NewStorageError("IMPORT", "failed to parse import file", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return NewStorageError("IMPORT", "failed to begin transaction", err)
	}
	defer tx.Rollback()

	if messages, ok := jsonData["messages"].([]interface{}); ok {
		for range messages {
			// import message
		}
	}

	if err := tx.Commit(); err != nil {
		return NewStorageError("IMPORT", "failed to commit transaction", err)
	}

	return nil
}

func exportTableToJSON(db *sql.DB, tableName string) ([]map[string]interface{}, error) {
	query := `SELECT * FROM ` + tableName
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results, nil
}
