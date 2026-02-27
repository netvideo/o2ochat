package cli

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DataHandler struct {
	dataPath string
}

func NewDataHandler(dataPath string) *DataHandler {
	if dataPath == "" {
		dataPath = os.Getenv("HOME") + "/.o2ochat/data"
	}
	return &DataHandler{dataPath: dataPath}
}

func (h *DataHandler) BackupData(path string) (*CommandResult, error) {
	if path == "" {
		timestamp := time.Now().Format("20060102_150405")
		path = filepath.Join(".", fmt.Sprintf("o2ochat_backup_%s.zip", timestamp))
	}

	if err := h.createBackup(path); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("backup failed: %v", err),
			ExitCode: 1,
		}, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("backup failed: %v", err),
			ExitCode: 1,
		}, nil
	}

	return &CommandResult{
		Success: true,
		Message: fmt.Sprintf("backup created: %s", path),
		Data: map[string]interface{}{
			"path":      path,
			"size":      info.Size(),
			"timestamp": time.Now().Format(time.RFC3339),
		},
		ExitCode: 0,
	}, nil
}

func (h *DataHandler) createBackup(backupPath string) error {
	file, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	return filepath.Walk(h.dataPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(h.dataPath, path)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

func (h *DataHandler) RestoreData(path string) (*CommandResult, error) {
	if path == "" {
		return &CommandResult{
			Success:  false,
			Message:  "backup path is required",
			ExitCode: 1,
		}, nil
	}

	if err := h.extractBackup(path); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("restore failed: %v", err),
			ExitCode: 1,
		}, nil
	}

	return &CommandResult{
		Success:  true,
		Message:  fmt.Sprintf("data restored from %s", path),
		ExitCode: 0,
	}, nil
}

func (h *DataHandler) extractBackup(backupPath string) error {
	reader, err := zip.OpenReader(backupPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := os.MkdirAll(h.dataPath, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(h.dataPath, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		destFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer destFile.Close()

		srcFile, err := file.Open()
		if err != nil {
			return err
		}
		defer srcFile.Close()

		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *DataHandler) CleanupData(olderThan string) (*CommandResult, error) {
	cutoffTime := time.Now()

	if olderThan != "" {
		duration, err := parseDuration(olderThan)
		if err != nil {
			return &CommandResult{
				Success:  false,
				Message:  fmt.Sprintf("invalid duration format: %v", err),
				ExitCode: 1,
			}, nil
		}
		cutoffTime = time.Now().Add(-duration)
	}

	deletedCount := 0
	deletedSize := int64(0)

	err := filepath.Walk(h.dataPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".log" && ext != ".tmp" && ext != ".cache" {
			return nil
		}

		if info.ModTime().Before(cutoffTime) {
			size := info.Size()
			if err := os.Remove(path); err == nil {
				deletedCount++
				deletedSize += size
			}
		}

		return nil
	})

	if err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("cleanup failed: %v", err),
			ExitCode: 1,
		}, nil
	}

	return &CommandResult{
		Success: true,
		Message: fmt.Sprintf("cleanup completed"),
		Data: map[string]interface{}{
			"deleted_files": deletedCount,
			"deleted_size":  deletedSize,
			"cutoff_time":   cutoffTime.Format(time.RFC3339),
			"data_path":     h.dataPath,
		},
		ExitCode: 0,
	}, nil
}

func (h *DataHandler) ExportMessages(peerID, path string) (*CommandResult, error) {
	if path == "" {
		timestamp := time.Now().Format("20060102_150405")
		path = fmt.Sprintf("messages_%s.json", timestamp)
	}

	messages := h.getMessages(peerID)

	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("export failed: %v", err),
			ExitCode: 1,
		}, nil
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("export failed: %v", err),
			ExitCode: 1,
		}, nil
	}

	return &CommandResult{
		Success: true,
		Message: fmt.Sprintf("messages exported to %s", path),
		Data: map[string]interface{}{
			"path":          path,
			"message_count": len(messages),
			"peer_id":       peerID,
		},
		ExitCode: 0,
	}, nil
}

func (h *DataHandler) getMessages(peerID string) []map[string]interface{} {
	messagesDir := filepath.Join(h.dataPath, "messages")
	if _, err := os.Stat(messagesDir); os.IsNotExist(err) {
		return []map[string]interface{}{}
	}

	var messages []map[string]interface{}

	filepath.Walk(messagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if strings.HasPrefix(info.Name(), "msg_") && strings.HasSuffix(info.Name(), ".json") {
			data, err := os.ReadFile(path)
			if err == nil {
				var msg map[string]interface{}
				if err := json.Unmarshal(data, &msg); err == nil {
					if peerID == "" || msg["peer_id"] == peerID {
						messages = append(messages, msg)
					}
				}
			}
		}
		return nil
	})

	return messages
}

func (h *DataHandler) ImportMessages(path string) (*CommandResult, error) {
	if path == "" {
		return &CommandResult{
			Success:  false,
			Message:  "import path is required",
			ExitCode: 1,
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("import failed: %v", err),
			ExitCode: 1,
		}, nil
	}

	var messages []map[string]interface{}
	if err := json.Unmarshal(data, &messages); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("import failed: %v", err),
			ExitCode: 1,
		}, nil
	}

	messagesDir := filepath.Join(h.dataPath, "messages")
	if err := os.MkdirAll(messagesDir, 0755); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("import failed: %v", err),
			ExitCode: 1,
		}, nil
	}

	importedCount := 0
	for _, msg := range messages {
		timestamp := time.Now().UnixNano()
		filename := filepath.Join(messagesDir, fmt.Sprintf("msg_%d.json", timestamp))

		msgData, err := json.MarshalIndent(msg, "", "  ")
		if err != nil {
			continue
		}

		if err := os.WriteFile(filename, msgData, 0644); err == nil {
			importedCount++
		}
	}

	return &CommandResult{
		Success: true,
		Message: fmt.Sprintf("imported %d messages", importedCount),
		Data: map[string]interface{}{
			"imported_count": importedCount,
			"total_count":    len(messages),
		},
		ExitCode: 0,
	}, nil
}

func (h *DataHandler) ShowStorageStats() (*CommandResult, error) {
	stats := StorageStats{}

	filepath.Walk(h.dataPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		stats.TotalSize += info.Size()

		if strings.HasPrefix(filepath.Base(path), "msg_") {
			stats.MessageCount++
		} else if strings.HasPrefix(filepath.Base(path), "chunk_") {
			stats.FileCount++
		}

		return nil
	})

	stats.UsedSize = stats.TotalSize

	return &CommandResult{
		Success:  true,
		Message:  "storage statistics",
		Data:     stats,
		ExitCode: 0,
	}, nil
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)

	multipliers := map[string]time.Duration{
		"s": time.Second,
		"m": time.Minute,
		"h": time.Hour,
		"d": 24 * time.Hour,
		"w": 7 * 24 * time.Hour,
	}

	for suffix, mult := range multipliers {
		if strings.HasSuffix(s, suffix) {
			numStr := strings.TrimSuffix(s, suffix)
			var num float64
			if _, err := fmt.Sscanf(numStr, "%f", &num); err != nil {
				return 0, err
			}
			return time.Duration(num) * mult, nil
		}
	}

	return time.ParseDuration(s)
}

type BackupCommandHandler struct {
	dataHandler *DataHandler
}

func NewBackupCommandHandler(dataPath string) *BackupCommandHandler {
	return &BackupCommandHandler{
		dataHandler: NewDataHandler(dataPath),
	}
}

func (h *BackupCommandHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	path := ""
	if v, ok := ctx.Flags["path"].(string); ok {
		path = v
	}
	return h.dataHandler.BackupData(path)
}

func (h *BackupCommandHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *BackupCommandHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	if len(word) > 1 && word[0] == '-' {
		return []string{"--path", "--help"}, nil
	}
	return nil, nil
}

func (h *BackupCommandHandler) Help() string {
	return `备份O2OChat数据。

用法:
  o2ochat data backup [选项]

选项:
  -p, --path    备份文件路径
  -h, --help    显示帮助信息

示例:
  o2ochat data backup
  o2ochat data backup --path ./backup.zip`
}

type RestoreCommandHandler struct {
	dataHandler *DataHandler
}

func NewRestoreCommandHandler(dataPath string) *RestoreCommandHandler {
	return &RestoreCommandHandler{
		dataHandler: NewDataHandler(dataPath),
	}
}

func (h *RestoreCommandHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	path := ""
	if v, ok := ctx.Flags["path"].(string); ok {
		path = v
	}
	return h.dataHandler.RestoreData(path)
}

func (h *RestoreCommandHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *RestoreCommandHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	if len(word) > 1 && word[0] == '-' {
		return []string{"--path", "--help"}, nil
	}
	return nil, nil
}

func (h *RestoreCommandHandler) Help() string {
	return `恢复O2OChat数据。

用法:
  o2ochat data restore [选项]

选项:
  -p, --path    备份文件路径
  -h, --help    显示帮助信息

示例:
  o2ochat data restore --path ./backup.zip`
}

type CleanupCommandHandler struct {
	dataHandler *DataHandler
}

func NewCleanupCommandHandler(dataPath string) *CleanupCommandHandler {
	return &CleanupCommandHandler{
		dataHandler: NewDataHandler(dataPath),
	}
}

func (h *CleanupCommandHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	olderThan := ""
	if v, ok := ctx.Flags["older-than"].(string); ok {
		olderThan = v
	}
	return h.dataHandler.CleanupData(olderThan)
}

func (h *CleanupCommandHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *CleanupCommandHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	if len(word) > 1 && word[0] == '-' {
		return []string{"--older-than", "--help"}, nil
	}
	return nil, nil
}

func (h *CleanupCommandHandler) Help() string {
	return `清理O2OChat临时文件。

用法:
  o2ochat data cleanup [选项]

选项:
  -o, --older-than    清理此时间之前的文件 (如: 7d, 24h, 30m)
  -h, --help          显示帮助信息

示例:
  o2ochat data cleanup
  o2ochat data cleanup --older-than 7d`
}

type StorageStatsCommandHandler struct {
	dataHandler *DataHandler
}

func NewStorageStatsCommandHandler(dataPath string) *StorageStatsCommandHandler {
	return &StorageStatsCommandHandler{
		dataHandler: NewDataHandler(dataPath),
	}
}

func (h *StorageStatsCommandHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	return h.dataHandler.ShowStorageStats()
}

func (h *StorageStatsCommandHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *StorageStatsCommandHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	return nil, nil
}

func (h *StorageStatsCommandHandler) Help() string {
	return `显示存储统计信息。

用法:
  o2ochat data stats [选项]

选项:
  -h, --help    显示帮助信息

示例:
  o2ochat data stats`
}
