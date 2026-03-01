// Package filetransfer provides file transfer capabilities for P2P communications
package filetransfer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"os"
	"sync"
	"time"
)

// TransferState represents the state of a file transfer
type TransferState int

const (
	StatePending TransferState = iota
	StateInProgress
	StatePaused
	StateCompleted
	StateFailed
	StateCancelled
)

func (s TransferState) String() string {
	switch s {
	case StatePending:
		return "pending"
	case StateInProgress:
		return "in_progress"
	case StatePaused:
		return "paused"
	case StateCompleted:
		return "completed"
	case StateFailed:
		return "failed"
	case StateCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// TransferDirection represents the direction of a transfer
type TransferDirection int

const (
	DirectionUpload TransferDirection = iota
	DirectionDownload
)

// Chunk represents a file chunk
type Chunk struct {
	Index      int    `json:"index"`
	Data       []byte `json:"data"`
	Checksum   string `json:"checksum"`
	TransferID string `json:"transfer_id"`
}

// TransferInfo holds information about a file transfer
type TransferInfo struct {
	ID             string            `json:"id"`
	FileName       string            `json:"file_name"`
	FilePath       string            `json:"file_path,omitempty"`
	FileSize       int64             `json:"file_size"`
	FileType       string            `json:"file_type"`
	Checksum       string            `json:"checksum"`
	ChunkSize      int               `json:"chunk_size"`
	TotalChunks    int               `json:"total_chunks"`
	ReceivedChunks []int             `json:"received_chunks,omitempty"`
	State          TransferState     `json:"state"`
	Direction      TransferDirection `json:"direction"`
	PeerID         string            `json:"peer_id"`
	Progress       float64           `json:"progress"`
	Speed          float64           `json:"speed"` // bytes per second
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	CompletedAt    *time.Time        `json:"completed_at,omitempty"`
	Error          string            `json:"error,omitempty"`
}

// ProgressCallback is called when transfer progress updates
type ProgressCallback func(info *TransferInfo)

// EventHandler is called when transfer state changes
type EventHandler func(event TransferEvent)

// TransferEvent represents a transfer state change event
type TransferEvent struct {
	Type       EventType
	TransferID string
	Timestamp  time.Time
	Data       interface{}
}

// EventType represents the type of transfer event
type EventType int

const (
	EventStarted EventType = iota
	EventProgress
	EventCompleted
	EventFailed
	EventCancelled
	EventPaused
	EventResumed
)

// Manager defines the interface for file transfer management
type Manager interface {
	// Upload initiates a file upload
	Upload(ctx context.Context, filePath string, peerID string, callback ProgressCallback) (*TransferInfo, error)

	// Download initiates a file download
	Download(ctx context.Context, transferID string, savePath string, callback ProgressCallback) (*TransferInfo, error)

	// Pause pauses an active transfer
	Pause(transferID string) error

	// Resume resumes a paused transfer
	Resume(transferID string) error

	// Cancel cancels an active transfer
	Cancel(transferID string) error

	// GetTransfer returns transfer information
	GetTransfer(transferID string) (*TransferInfo, error)

	// ListTransfers lists all transfers (optionally filtered by state)
	ListTransfers(state *TransferState) ([]*TransferInfo, error)

	// OnProgress sets a callback for progress updates
	OnProgress(callback ProgressCallback)

	// OnEvent sets a callback for transfer events
	OnEvent(handler EventHandler)

	// Close shuts down the manager
	Close() error
}

// manager implements Manager
type manager struct {
	transfers    map[string]*transfer
	transfersMu  sync.RWMutex
	chunkSize    int
	downloadPath string

	onProgress ProgressCallback
	onEvent    EventHandler
	mu         sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// transfer represents an active file transfer
type transfer struct {
	info       *TransferInfo
	file       *os.File
	hash       hash.Hash
	ctx        context.Context
	cancel     context.CancelFunc
	pauseChan  chan struct{}
	resumeChan chan struct{}
	mu         sync.RWMutex
}

// Ensure manager implements Manager
var _ Manager = (*manager)(nil)

// Config holds file transfer configuration
type Config struct {
	ChunkSize      int
	ParallelChunks int
	DownloadPath   string
	MaxRetries     int
	RetryInterval  time.Duration
	ResumeEnabled  bool
}

// DefaultTransferConfig returns default configuration
func DefaultTransferConfig() Config {
	return Config{
		ChunkSize:      256 * 1024, // 256KB
		ParallelChunks: 5,
		DownloadPath:   "./downloads",
		MaxRetries:     3,
		RetryInterval:  5 * time.Second,
		ResumeEnabled:  true,
	}
}

// NewManager creates a new file transfer manager
func NewManager(config Config) (Manager, error) {
	// Ensure download path exists
	if err := os.MkdirAll(config.DownloadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create download directory: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &manager{
		transfers:    make(map[string]*transfer),
		chunkSize:    config.ChunkSize,
		downloadPath: config.DownloadPath,
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

// Upload initiates a file upload
func (m *manager) Upload(ctx context.Context, filePath string, peerID string, callback ProgressCallback) (*TransferInfo, error) {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Calculate checksum
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}
	checksum := hex.EncodeToString(hash.Sum(nil))

	// Reset file pointer
	file.Seek(0, 0)

	// Create transfer info
	transferID := generateTransferID()
	totalChunks := int((stat.Size() + int64(m.chunkSize) - 1) / int64(m.chunkSize))

	info := &TransferInfo{
		ID:          transferID,
		FileName:    stat.Name(),
		FilePath:    filePath,
		FileSize:    stat.Size(),
		FileType:    "",
		Checksum:    checksum,
		ChunkSize:   m.chunkSize,
		TotalChunks: totalChunks,
		State:       StateInProgress,
		Direction:   DirectionUpload,
		PeerID:      peerID,
		Progress:    0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create transfer context
	tctx, cancel := context.WithCancel(ctx)

	// Store transfer
	t := &transfer{
		info:       info,
		file:       file,
		hash:       sha256.New(),
		ctx:        tctx,
		cancel:     cancel,
		pauseChan:  make(chan struct{}),
		resumeChan: make(chan struct{}),
	}

	m.transfersMu.Lock()
	m.transfers[transferID] = t
	m.transfersMu.Unlock()

	// Emit event
	m.emitEvent(TransferEvent{
		Type:       EventStarted,
		TransferID: transferID,
		Timestamp:  time.Now(),
	})

	return info, nil
}

// Download initiates a file download
func (m *manager) Download(ctx context.Context, transferID string, savePath string, callback ProgressCallback) (*TransferInfo, error) {
	// This would initiate a download request to the peer
	// For now, return an error
	return nil, fmt.Errorf("not implemented")
}

// Pause pauses an active transfer
func (m *manager) Pause(transferID string) error {
	m.transfersMu.RLock()
	t, exists := m.transfers[transferID]
	m.transfersMu.RUnlock()

	if !exists {
		return fmt.Errorf("transfer not found: %s", transferID)
	}

	t.mu.Lock()
	if t.info.State != StateInProgress {
		t.mu.Unlock()
		return fmt.Errorf("transfer is not in progress")
	}
	t.info.State = StatePaused
	t.info.UpdatedAt = time.Now()
	t.mu.Unlock()

	// Signal pause
	select {
	case t.pauseChan <- struct{}{}:
	default:
	}

	// Emit event
	m.emitEvent(TransferEvent{
		Type:       EventPaused,
		TransferID: transferID,
		Timestamp:  time.Now(),
	})

	return nil
}

// Resume resumes a paused transfer
func (m *manager) Resume(transferID string) error {
	m.transfersMu.RLock()
	t, exists := m.transfers[transferID]
	m.transfersMu.RUnlock()

	if !exists {
		return fmt.Errorf("transfer not found: %s", transferID)
	}

	t.mu.Lock()
	if t.info.State != StatePaused {
		t.mu.Unlock()
		return fmt.Errorf("transfer is not paused")
	}
	t.info.State = StateInProgress
	t.info.UpdatedAt = time.Now()
	t.mu.Unlock()

	// Signal resume
	select {
	case t.resumeChan <- struct{}{}:
	default:
	}

	// Emit event
	m.emitEvent(TransferEvent{
		Type:       EventResumed,
		TransferID: transferID,
		Timestamp:  time.Now(),
	})

	return nil
}

// Cancel cancels an active transfer
func (m *manager) Cancel(transferID string) error {
	m.transfersMu.Lock()
	t, exists := m.transfers[transferID]
	if !exists {
		m.transfersMu.Unlock()
		return fmt.Errorf("transfer not found: %s", transferID)
	}

	delete(m.transfers, transferID)
	m.transfersMu.Unlock()

	// Cancel transfer context
	t.cancel()

	// Close file if open
	if t.file != nil {
		t.file.Close()
	}

	// Emit event
	m.emitEvent(TransferEvent{
		Type:       EventCancelled,
		TransferID: transferID,
		Timestamp:  time.Now(),
	})

	return nil
}

// GetTransfer returns transfer information
func (m *manager) GetTransfer(transferID string) (*TransferInfo, error) {
	m.transfersMu.RLock()
	t, exists := m.transfers[transferID]
	m.transfersMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("transfer not found: %s", transferID)
	}

	t.mu.RLock()
	info := *t.info // Copy
	t.mu.RUnlock()

	return &info, nil
}

// ListTransfers lists all transfers (optionally filtered by state)
func (m *manager) ListTransfers(state *TransferState) ([]*TransferInfo, error) {
	m.transfersMu.RLock()
	defer m.transfersMu.RUnlock()

	var transfers []*TransferInfo
	for _, t := range m.transfers {
		t.mu.RLock()
		if state == nil || t.info.State == *state {
			info := *t.info // Copy
			transfers = append(transfers, &info)
		}
		t.mu.RUnlock()
	}

	return transfers, nil
}

// OnProgress sets a callback for progress updates
func (m *manager) OnProgress(callback ProgressCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onProgress = callback
}

// OnEvent sets a callback for transfer events
func (m *manager) OnEvent(handler EventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onEvent = handler
}

// Close shuts down the manager
func (m *manager) Close() error {
	m.cancel()

	// Cancel all active transfers
	m.transfersMu.Lock()
	transfers := make([]*transfer, 0, len(m.transfers))
	for _, t := range m.transfers {
		transfers = append(transfers, t)
	}
	m.transfersMu.Unlock()

	for _, t := range transfers {
		m.Cancel(t.info.ID)
	}

	m.wg.Wait()
	return nil
}

// emitEvent emits a transfer event
func (m *manager) emitEvent(event TransferEvent) {
	m.mu.RLock()
	handler := m.onEvent
	m.mu.RUnlock()

	if handler != nil {
		handler(event)
	}
}

// emitProgress emits a progress update
func (m *manager) emitProgress(info *TransferInfo) {
	m.mu.RLock()
	callback := m.onProgress
	m.mu.RUnlock()

	if callback != nil {
		callback(info)
	}
}

// generateTransferID generates a unique transfer ID
func generateTransferID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().UnixMilli())
}

// MerkleTree represents a Merkle tree for file integrity verification
type MerkleTree struct {
	Root   string   `json:"root"`
	Leaves []string `json:"leaves"`
}

// BuildMerkleTree builds a Merkle tree from file chunks
func BuildMerkleTree(chunks [][]byte) *MerkleTree {
	if len(chunks) == 0 {
		return nil
	}

	// Calculate leaf hashes
	leaves := make([]string, len(chunks))
	for i, chunk := range chunks {
		hash := sha256.Sum256(chunk)
		leaves[i] = hex.EncodeToString(hash[:])
	}

	// Build tree bottom-up
	level := leaves
	for len(level) > 1 {
		nextLevel := make([]string, 0, (len(level)+1)/2)
		for i := 0; i < len(level); i += 2 {
			if i+1 < len(level) {
				combined := level[i] + level[i+1]
				hash := sha256.Sum256([]byte(combined))
				nextLevel = append(nextLevel, hex.EncodeToString(hash[:]))
			} else {
				nextLevel = append(nextLevel, level[i])
			}
		}
		level = nextLevel
	}

	return &MerkleTree{
		Root:   level[0],
		Leaves: leaves,
	}
}

// VerifyChunk verifies a chunk against the Merkle tree
func (m *MerkleTree) VerifyChunk(index int, chunk []byte) bool {
	if index < 0 || index >= len(m.Leaves) {
		return false
	}

	hash := sha256.Sum256(chunk)
	chunkHash := hex.EncodeToString(hash[:])

	return chunkHash == m.Leaves[index]
}

// ResumeInfo holds information for resuming a transfer
type ResumeInfo struct {
	TransferID     string   `json:"transfer_id"`
	FilePath       string   `json:"file_path"`
	Checksum       string   `json:"checksum"`
	TotalChunks    int      `json:"total_chunks"`
	ReceivedChunks []int    `json:"received_chunks"`
	ChunkChecksums []string `json:"chunk_checksums,omitempty"`
}

// SaveResumeInfo saves resume information to a file
func SaveResumeInfo(info *ResumeInfo, path string) error {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal resume info: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write resume info: %w", err)
	}

	return nil
}

// LoadResumeInfo loads resume information from a file
func LoadResumeInfo(path string) (*ResumeInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read resume info: %w", err)
	}

	var info ResumeInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resume info: %w", err)
	}

	return &info, nil
}

// CalculateFileChecksum calculates the SHA256 checksum of a file
func CalculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
