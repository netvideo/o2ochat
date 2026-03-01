package filetransfer

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileChunk represents a file chunk
type FileChunk struct {
	Index    int
	Data     []byte
	Hash     string
	Size     int64
	Uploaded bool
}

// FileInfo represents file information
type FileInfo struct {
	Name       string
	Size       int64
	Hash       string
	ChunkCount int
	ChunkSize  int64
}

// FileTransferManager manages file transfers
type FileTransferManager struct {
	chunks       map[string][]*FileChunk
	transfers    map[string]*FileTransfer
	mu           sync.RWMutex
	stats        TransferStats
}

// FileTransfer represents a file transfer
type FileTransfer struct {
	ID           string
	FilePath     string
	RemotePath   string
	Size         int64
	Transferred  int64
	Status       TransferStatus
	StartTime    time.Time
	EndTime      time.Time
	Error        error
	mu           sync.RWMutex
}

// TransferStatus represents transfer status
type TransferStatus string

const (
	TransferStatusPending    TransferStatus = "pending"
	TransferStatusInProgress TransferStatus = "in_progress"
	TransferStatusCompleted  TransferStatus = "completed"
	TransferStatusFailed     TransferStatus = "failed"
	TransferStatusPaused     TransferStatus = "paused"
)

// TransferStats represents transfer statistics
type TransferStats struct {
	TotalTransfers   int
	CompletedTransfers int
	FailedTransfers  int
	TotalBytes       int64
	TransferredBytes int64
}

// DefaultChunkSize is the default chunk size (1MB)
const DefaultChunkSize = 1024 * 1024

// NewFileTransferManager creates a new file transfer manager
func NewFileTransferManager() *FileTransferManager {
	return &FileTransferManager{
		chunks:    make(map[string][]*FileChunk),
		transfers: make(map[string]*FileTransfer),
	}
}

// SplitFile splits a file into chunks
func (ftm *FileTransferManager) SplitFile(filePath string, chunkSize int64) (*FileInfo, error) {
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Calculate chunk count
	chunkCount := int((fileInfo.Size() + chunkSize - 1) / chunkSize)

	// Create chunks
	chunks := make([]*FileChunk, chunkCount)
	buffer := make([]byte, chunkSize)

	for i := 0; i < chunkCount; i++ {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}

		chunkData := make([]byte, n)
		copy(chunkData, buffer[:n])

		chunk := &FileChunk{
			Index: i,
			Data:  chunkData,
			Hash:  calculateHash(chunkData),
			Size:  int64(n),
		}

		chunks[i] = chunk
	}

	// Store chunks
	fileID := filepath.Base(filePath)
	ftm.mu.Lock()
	ftm.chunks[fileID] = chunks
	ftm.mu.Unlock()

	// Calculate file hash
	fileHash, err := calculateFileHash(filePath)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		Name:       filepath.Base(filePath),
		Size:       fileInfo.Size(),
		Hash:       fileHash,
		ChunkCount: chunkCount,
		ChunkSize:  chunkSize,
	}, nil
}

// MergeChunks merges chunks into a file
func (ftm *FileTransferManager) MergeChunks(fileID, outputPath string) error {
	ftm.mu.RLock()
	chunks, exists := ftm.chunks[fileID]
	ftm.mu.RUnlock()

	if !exists {
		return errors.New("chunks not found")
	}

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write chunks
	for _, chunk := range chunks {
		if !chunk.Uploaded {
			continue
		}

		_, err := file.Write(chunk.Data)
		if err != nil {
			return err
		}
	}

	// Verify file hash
	mergedHash, err := calculateFileHash(outputPath)
	if err != nil {
		return err
	}

	// Calculate expected hash from chunks
	expectedHash := ""
	for _, chunk := range chunks {
		expectedHash += chunk.Hash
	}
	expectedHash = calculateHash([]byte(expectedHash))

	if mergedHash != expectedHash {
		return errors.New("file hash mismatch")
	}

	return nil
}

// StartTransfer starts a file transfer
func (ftm *FileTransferManager) StartTransfer(fileID, filePath, remotePath string) (*FileTransfer, error) {
	ftm.mu.Lock()
	defer ftm.mu.Unlock()

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	// Create transfer
	transfer := &FileTransfer{
		ID:          fileID,
		FilePath:    filePath,
		RemotePath:  remotePath,
		Size:        fileInfo.Size(),
		Status:      TransferStatusInProgress,
		StartTime:   time.Now(),
	}

	ftm.transfers[fileID] = transfer
	ftm.stats.TotalTransfers++

	return transfer, nil
}

// UpdateTransferProgress updates transfer progress
func (ftm *FileTransferManager) UpdateTransferProgress(fileID string, transferred int64) error {
	ftm.mu.Lock()
	defer ftm.mu.Unlock()

	transfer, exists := ftm.transfers[fileID]
	if !exists {
		return errors.New("transfer not found")
	}

	transfer.mu.Lock()
	transfer.Transferred = transferred
	if transferred >= transfer.Size {
		transfer.Status = TransferStatusCompleted
		transfer.EndTime = time.Now()
		ftm.stats.CompletedTransfers++
	}
	ftm.stats.TransferredBytes += transferred
	transfer.mu.Unlock()

	return nil
}

// PauseTransfer pauses a transfer
func (ftm *FileTransferManager) PauseTransfer(fileID string) error {
	ftm.mu.Lock()
	defer ftm.mu.Unlock()

	transfer, exists := ftm.transfers[fileID]
	if !exists {
		return errors.New("transfer not found")
	}

	transfer.mu.Lock()
	transfer.Status = TransferStatusPaused
	transfer.mu.Unlock()

	return nil
}

// ResumeTransfer resumes a paused transfer
func (ftm *FileTransferManager) ResumeTransfer(fileID string) error {
	ftm.mu.Lock()
	defer ftm.mu.Unlock()

	transfer, exists := ftm.transfers[fileID]
	if !exists {
		return errors.New("transfer not found")
	}

	transfer.mu.Lock()
	if transfer.Status == TransferStatusPaused {
		transfer.Status = TransferStatusInProgress
	}
	transfer.mu.Unlock()

	return nil
}

// CancelTransfer cancels a transfer
func (ftm *FileTransferManager) CancelTransfer(fileID string) error {
	ftm.mu.Lock()
	defer ftm.mu.Unlock()

	transfer, exists := ftm.transfers[fileID]
	if !exists {
		return errors.New("transfer not found")
	}

	transfer.mu.Lock()
	transfer.Status = TransferStatusFailed
	transfer.EndTime = time.Now()
	transfer.Error = errors.New("transfer cancelled")
	ftm.stats.FailedTransfers++
	transfer.mu.Unlock()

	return nil
}

// GetTransfer gets a transfer by ID
func (ftm *FileTransferManager) GetTransfer(fileID string) (*FileTransfer, error) {
	ftm.mu.RLock()
	defer ftm.mu.RUnlock()

	transfer, exists := ftm.transfers[fileID]
	if !exists {
		return nil, errors.New("transfer not found")
	}

	return transfer, nil
}

// GetTransferProgress gets transfer progress
func (ftm *FileTransferManager) GetTransferProgress(fileID string) (float64, error) {
	ftm.mu.RLock()
	defer ftm.mu.RUnlock()

	transfer, exists := ftm.transfers[fileID]
	if !exists {
		return 0, errors.New("transfer not found")
	}

	transfer.mu.RLock()
	defer transfer.mu.RUnlock()

	if transfer.Size == 0 {
		return 0, nil
	}

	return float64(transfer.Transferred) / float64(transfer.Size) * 100, nil
}

// GetStats gets transfer statistics
func (ftm *FileTransferManager) GetStats() TransferStats {
	ftm.mu.RLock()
	defer ftm.mu.RUnlock()
	return ftm.stats
}

// GetTransferSpeed gets transfer speed in bytes per second
func (ftm *FileTransferManager) GetTransferSpeed(fileID string) (float64, error) {
	ftm.mu.RLock()
	defer ftm.mu.RUnlock()

	transfer, exists := ftm.transfers[fileID]
	if !exists {
		return 0, errors.New("transfer not found")
	}

	transfer.mu.RLock()
	defer transfer.mu.RUnlock()

	if transfer.Status != TransferStatusInProgress {
		return 0, nil
	}

	elapsed := time.Since(transfer.StartTime).Seconds()
	if elapsed == 0 {
		return 0, nil
	}

	return float64(transfer.Transferred) / elapsed, nil
}

// GetETAGetResumeInfo gets resume information for a transfer
func (ftm *FileTransferManager) GetResumeInfo(fileID string) (int64, error) {
	ftm.mu.RLock()
	defer ftm.mu.RUnlock()

	transfer, exists := ftm.transfers[fileID]
	if !exists {
		return 0, errors.New("transfer not found")
	}

	transfer.mu.RLock()
	defer transfer.mu.RUnlock()

	return transfer.Transferred, nil
}

// calculateHash calculates SHA256 hash of data
func calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// calculateFileHash calculates SHA256 hash of a file
func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
