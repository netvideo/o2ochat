package filetransfer

import (
	"sync"
	"time"
)

// ParallelTransferManager manages parallel file transfers
type ParallelTransferManager struct {
	maxParallel int
	transfers   map[string]*ParallelTransfer
	mu          sync.RWMutex
	stats       ParallelStats
}

// ParallelTransfer represents a parallel transfer
type ParallelTransfer struct {
	ID            string
	FilePath      string
	RemotePath    string
	ChunkCount    int
	ParallelCount int
	Status        TransferStatus
	StartTime     time.Time
	EndTime       time.Time
	Error         error
	mu            sync.RWMutex
}

// ParallelStats represents parallel transfer statistics
type ParallelStats struct {
	TotalTransfers    int
	CompletedTransfers int
	FailedTransfers   int
	AverageSpeed      float64
	MaxSpeed          float64
}

// NewParallelTransferManager creates a new parallel transfer manager
func NewParallelTransferManager(maxParallel int) *ParallelTransferManager {
	if maxParallel <= 0 {
		maxParallel = 4 // Default 4 parallel transfers
	}

	return &ParallelTransferManager{
		maxParallel: maxParallel,
		transfers:   make(map[string]*ParallelTransfer),
	}
}

// StartParallelTransfer starts a parallel file transfer
func (ptm *ParallelTransferManager) StartParallelTransfer(fileID, filePath, remotePath string, chunkCount int) (*ParallelTransfer, error) {
	ptm.mu.Lock()
	defer ptm.mu.Unlock()

	transfer := &ParallelTransfer{
		ID:            fileID,
		FilePath:      filePath,
		RemotePath:    remotePath,
		ChunkCount:    chunkCount,
		ParallelCount: ptm.maxParallel,
		Status:        TransferStatusInProgress,
		StartTime:     time.Now(),
	}

	ptm.transfers[fileID] = transfer
	ptm.stats.TotalTransfers++

	return transfer, nil
}

// TransferChunks transfers chunks in parallel
func (ptm *ParallelTransferManager) TransferChunks(fileID string, chunks []*FileChunk, uploadFunc func(chunk *FileChunk) error) error {
	ptm.mu.RLock()
	transfer, exists := ptm.transfers[fileID]
	ptm.mu.RUnlock()

	if !exists {
		return ErrTransferNotFound
	}

	// Create worker pool
	semaphore := make(chan struct{}, ptm.maxParallel)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var hasError error

	for _, chunk := range chunks {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(c *FileChunk) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			// Check if transfer is still active
			transfer.mu.RLock()
			status := transfer.Status
			transfer.mu.RUnlock()

			if status != TransferStatusInProgress {
				return
			}

			// Upload chunk
			err := uploadFunc(c)
			if err != nil {
				mu.Lock()
				hasError = err
				mu.Unlock()
				return
			}

			c.Uploaded = true
		}(chunk)
	}

	wg.Wait()
	return hasError
}

// GetParallelStats gets parallel transfer statistics
func (ptm *ParallelTransferManager) GetParallelStats() ParallelStats {
	ptm.mu.RLock()
	defer ptm.mu.RUnlock()
	return ptm.stats
}

// SetMaxParallel sets maximum parallel transfers
func (ptm *ParallelTransferManager) SetMaxParallel(max int) {
	ptm.mu.Lock()
	defer ptm.mu.Unlock()
	ptm.maxParallel = max
}

// GetMaxParallel gets maximum parallel transfers
func (ptm *ParallelTransferManager) GetMaxParallel() int {
	ptm.mu.RLock()
	defer ptm.mu.RUnlock()
	return ptm.maxParallel
}
