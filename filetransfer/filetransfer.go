package filetransfer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type FileTransferManagerImpl struct {
	chunkManager      ChunkManager
	scheduler         Scheduler
	tasks             map[string]*TransferTask
	mu                sync.RWMutex
	maxConcurrent    int
	progressCallbacks map[string]TransferProgressCallback
	eventChan        chan TransferEvent
}

func NewFileTransferManager(chunkManager ChunkManager, scheduler Scheduler, maxConcurrent int) FileTransferManager {
	if maxConcurrent <= 0 {
		maxConcurrent = 4
	}

	return &FileTransferManagerImpl{
		chunkManager:      chunkManager,
		scheduler:         scheduler,
		tasks:             make(map[string]*TransferTask),
		maxConcurrent:     maxConcurrent,
		progressCallbacks: make(map[string]TransferProgressCallback),
		eventChan:        make(chan TransferEvent, 100),
	}
}

func (m *FileTransferManagerImpl) RegisterProgressCallback(taskID string, callback TransferProgressCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.progressCallbacks[taskID] = callback
}

func (m *FileTransferManagerImpl) UnregisterProgressCallback(taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.progressCallbacks, taskID)
}

func (m *FileTransferManagerImpl) GetEventChan() <-chan TransferEvent {
	return m.eventChan
}

func (m *FileTransferManagerImpl) emitEvent(eventType TransferEventType, taskID, fileID string, err error) {
	event := TransferEvent{
		EventType: eventType,
		TaskID:    taskID,
		FileID:    fileID,
		Timestamp: time.Now(),
		Error:     err,
	}

	select {
	case m.eventChan <- event:
	default:
	}
}

func (m *FileTransferManagerImpl) notifyProgress(taskID string, progress TransferProgress) {
	m.mu.RLock()
	callback, ok := m.progressCallbacks[taskID]
	m.mu.RUnlock()

	if ok && callback != nil {
		callback(taskID, progress)
	}
}

func (m *FileTransferManagerImpl) CreateDownloadTask(fileID, destPath string, sourcePeers []string) (string, error) {
	if fileID == "" {
		return "", ErrInvalidFileID
	}

	if len(sourcePeers) == 0 {
		return "", ErrInsufficientPeers
	}

	taskID := generateTaskID()

	task := &TransferTask{
		TaskID:      taskID,
		FileID:      fileID,
		Direction:   DirectionDownload,
		SourcePeers: sourcePeers,
		DestPath:    destPath,
		StartedAt:   time.Now(),
		Status:      StatusPending,
		Progress: TransferProgress{
			TotalChunks:      0,
			Completed:        0,
			Failed:           0,
			BytesTransferred: 0,
			Speed:            0,
			EstimatedTime:    "",
		},
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[taskID]; exists {
		return "", ErrTaskAlreadyExists
	}

	m.tasks[taskID] = task

	return taskID, nil
}

func (m *FileTransferManagerImpl) CreateUploadTask(filePath string, targetPeers []string) (string, error) {
	if filePath == "" {
		return "", ErrInvalidFilePath
	}

	if len(targetPeers) == 0 {
		return "", ErrInsufficientPeers
	}

	metadata, err := m.chunkManager.ChunkFile(filePath, 1024*1024)
	if err != nil {
		return "", err
	}

	taskID := generateTaskID()

	task := &TransferTask{
		TaskID:      taskID,
		FileID:      metadata.FileID,
		Direction:   DirectionUpload,
		SourcePeers: targetPeers,
		DestPath:    filePath,
		StartedAt:   time.Now(),
		Status:      StatusPending,
		Progress: TransferProgress{
			TotalChunks:      metadata.TotalChunks,
			Completed:        0,
			Failed:           0,
			BytesTransferred: 0,
			Speed:            0,
			EstimatedTime:    "",
		},
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[taskID]; exists {
		return "", ErrTaskAlreadyExists
	}

	m.tasks[taskID] = task

	return taskID, nil
}

func (m *FileTransferManagerImpl) StartTransfer(taskID string) error {
	m.mu.Lock()
	task, ok := m.tasks[taskID]
	m.mu.Unlock()

	if !ok {
		return ErrTaskNotFound
	}

	if task.Status != StatusPending && task.Status != StatusPaused {
		return ErrInvalidTaskState
	}

	if task.Direction == DirectionDownload {
		task.Status = StatusDownloading
	} else {
		task.Status = StatusUploading
	}

	m.mu.Lock()
	m.tasks[taskID] = task
	m.mu.Unlock()

	go m.runTransfer(taskID)

	return nil
}

func (m *FileTransferManagerImpl) runTransfer(taskID string) {
	m.mu.RLock()
	task, ok := m.tasks[taskID]
	m.mu.RUnlock()

	if !ok {
		return
	}

	m.emitEvent(EventStarted, taskID, task.FileID, nil)

	availableChunks := make(map[int][]string)
	for i := 0; i < task.Progress.TotalChunks; i++ {
		availableChunks[i] = task.SourcePeers
	}

	startTime := time.Now()
	var totalBytes int64
	var lastUpdateTime time.Time
	speedSamples := make([]float64, 0, 10)

	for {
		m.mu.RLock()
		currentTask, ok := m.tasks[taskID]
		m.mu.RUnlock()

		if !ok || currentTask.Status == StatusPaused || currentTask.Status == StatusCancelled {
			break
		}

		chunkIdx, _, err := m.scheduler.SelectNextChunk(task.FileID, availableChunks)
		if err != nil || chunkIdx < 0 {
			break
		}

		chunkSize := int64(1024 * 1024)
		transferStart := time.Now()
		success := m.simulateTransfer(task, chunkIdx)
		transferDuration := time.Since(transferStart).Seconds()

		if success {
			totalBytes += chunkSize
			currentTask.Progress.Completed++
			currentTask.Progress.BytesTransferred = totalBytes

			elapsed := time.Since(startTime).Seconds()
			if elapsed > 0 {
				currentTask.Progress.Speed = float64(totalBytes) / elapsed / 1024
			}

			if transferDuration > 0 {
				instantSpeed := float64(chunkSize) / transferDuration / 1024
				speedSamples = append(speedSamples, instantSpeed)
				if len(speedSamples) > 10 {
					speedSamples = speedSamples[1:]
				}

				var avgSpeed float64
				for _, s := range speedSamples {
					avgSpeed += s
				}
				avgSpeed /= float64(len(speedSamples))

				remainingChunks := currentTask.Progress.TotalChunks - currentTask.Progress.Completed
				if avgSpeed > 0 && remainingChunks > 0 {
					etaSeconds := float64(remainingChunks) * float64(chunkSize) / 1024 / avgSpeed
					if etaSeconds < 60 {
						currentTask.Progress.EstimatedTime = fmt.Sprintf("%.0fs", etaSeconds)
					} else if etaSeconds < 3600 {
						currentTask.Progress.EstimatedTime = fmt.Sprintf("%.0fm", etaSeconds/60)
					} else {
						currentTask.Progress.EstimatedTime = fmt.Sprintf("%.1fh", etaSeconds/3600)
					}
				}
			}
		} else {
			currentTask.Progress.Failed++
		}

		m.scheduler.UpdateChunkStatus(task.FileID, chunkIdx, success, "")

		m.mu.Lock()
		if currentTask.Progress.Completed+currentTask.Progress.Failed >= currentTask.Progress.TotalChunks {
			if currentTask.Progress.Failed > 0 {
				currentTask.Status = StatusFailed
				m.emitEvent(EventFailed, taskID, task.FileID, ErrDownloadFailed)
			} else {
				currentTask.Status = StatusCompleted
				m.emitEvent(EventCompleted, taskID, task.FileID, nil)
			}
		}
		m.tasks[taskID] = currentTask
		m.mu.Unlock()

		if time.Since(lastUpdateTime) > 500*time.Millisecond {
			m.notifyProgress(taskID, currentTask.Progress)
			m.emitEvent(EventProgress, taskID, task.FileID, nil)
			lastUpdateTime = time.Now()
		}

		if currentTask.Status == StatusCompleted || currentTask.Status == StatusFailed {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func (m *FileTransferManagerImpl) simulateTransfer(task *TransferTask, chunkIdx int) bool {
	time.Sleep(50 * time.Millisecond)
	return true
}

func (m *FileTransferManagerImpl) PauseTransfer(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return ErrTaskNotFound
	}

	if task.Status != StatusDownloading && task.Status != StatusUploading {
		return ErrInvalidTaskState
	}

	task.Status = StatusPaused
	m.tasks[taskID] = task

	m.emitEvent(EventPaused, taskID, task.FileID, nil)

	return nil
}

func (m *FileTransferManagerImpl) ResumeTransfer(taskID string) error {
	m.mu.Lock()
	task, ok := m.tasks[taskID]
	if !ok {
		m.mu.Unlock()
		return ErrTaskNotFound
	}

	if task.Status != StatusPaused {
		m.mu.Unlock()
		return ErrInvalidTaskState
	}

	if task.Direction == DirectionDownload {
		task.Status = StatusDownloading
	} else {
		task.Status = StatusUploading
	}

	m.tasks[taskID] = task
	m.mu.Unlock()

	m.emitEvent(EventResumed, taskID, task.FileID, nil)

	go m.runTransfer(taskID)

	return nil
}

func (m *FileTransferManagerImpl) CancelTransfer(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return ErrTaskNotFound
	}

	task.Status = StatusCancelled
	m.tasks[taskID] = task

	m.emitEvent(EventCancelled, taskID, task.FileID, ErrCancelled)

	return nil
}

func (m *FileTransferManagerImpl) GetTaskStatus(taskID string) (*TransferTask, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

func (m *FileTransferManagerImpl) GetAllTasks() ([]*TransferTask, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*TransferTask, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (m *FileTransferManagerImpl) CleanupCompletedTasks() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	toDelete := make([]string, 0)
	for taskID, task := range m.tasks {
		if task.Status == StatusCompleted || task.Status == StatusFailed || task.Status == StatusCancelled {
			toDelete = append(toDelete, taskID)
		}
	}

	for _, taskID := range toDelete {
		delete(m.tasks, taskID)
	}

	return nil
}

func generateTaskID() string {
	hash := sha256.Sum256([]byte(time.Now().String() + fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(hash[:])[:16]
}
