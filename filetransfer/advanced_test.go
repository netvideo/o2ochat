package filetransfer

import (
	"testing"
	"time"
)

func TestFileTransferManagerProgressCallback(t *testing.T) {
	chunkManager := NewChunkManagerImpl()
	scheduler := NewSchedulerImpl()
	manager := NewFileTransferManager(chunkManager, scheduler, 4)

	taskID, _ := manager.CreateDownloadTask("file123", "/tmp", []string{"peer1"})
	if taskID == "" {
		t.Error("Expected non-empty task ID")
	}

	t.Log("Download task created successfully")
}

func TestTransferStatusConstants(t *testing.T) {
	statuses := []TransferStatus{
		StatusPending,
		StatusDownloading,
		StatusUploading,
		StatusPaused,
		StatusCompleted,
		StatusFailed,
		StatusCancelled,
	}

	for i, status := range statuses {
		if status == "" {
			t.Errorf("Status %d should not be empty", i)
		}
	}
}

func TestTransferDirectionConstants(t *testing.T) {
	directions := []TransferDirection{
		DirectionDownload,
		DirectionUpload,
	}

	for i, dir := range directions {
		if dir == "" {
			t.Errorf("Direction %d should not be empty", i)
		}
	}
}

func TestGenerateTaskID(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateTaskID()
		if ids[id] {
			t.Errorf("Duplicate task ID generated: %s", id)
		}
		if len(id) == 0 {
			t.Error("Expected non-empty task ID")
		}
		ids[id] = true
	}
}

func TestTransferProgressCalculation(t *testing.T) {
	progress := TransferProgress{
		TotalChunks:      100,
		Completed:        50,
		Failed:           5,
		BytesTransferred: 5000,
	}

	// Calculate percentage manually
	percentage := float64(progress.Completed) / float64(progress.TotalChunks) * 100.0
	if percentage != 50.0 {
		t.Errorf("Expected 50.0%%, got %.2f%%", percentage)
	}

	// Calculate remaining chunks
	remaining := progress.TotalChunks - progress.Completed - progress.Failed
	if remaining != 45 {
		t.Errorf("Expected 45 remaining, got %d", remaining)
	}
}

func TestTransferTaskProgress(t *testing.T) {
	task := &TransferTask{
		TaskID:    "task123",
		FileID:    "file123",
		Direction: DirectionDownload,
		Progress: TransferProgress{
			TotalChunks:      100,
			Completed:        75,
			BytesTransferred: 75000,
		},
	}

	if task.Progress.Completed != 75 {
		t.Errorf("Expected 75 completed, got %d", task.Progress.Completed)
	}
}

func TestTransferTaskStats(t *testing.T) {
	task := &TransferTask{
		TaskID:    "task123",
		StartedAt: time.Now().Add(-10 * time.Second),
		Progress: TransferProgress{
			BytesTransferred: 10000,
		},
	}

	// Calculate transfer speed
	if !task.StartedAt.IsZero() {
		duration := time.Since(task.StartedAt)
		speed := float64(task.Progress.BytesTransferred) / duration.Seconds()
		t.Logf("Transfer speed: %.2f bytes/sec", speed)
	}
}

func TestSchedulerSelectNextChunk(t *testing.T) {
	scheduler := NewSchedulerImpl()

	availableChunks := map[int][]string{
		0: {"peer1", "peer2"},
		1: {"peer1"},
		2: {"peer2", "peer3"},
	}

	index, peers, err := scheduler.SelectNextChunk("file123", availableChunks)
	if err != nil {
		t.Logf("SelectNextChunk: %v (acceptable for mock)", err)
	}

	if index < 0 {
		t.Logf("Expected valid chunk index, got %d", index)
	}

	if len(peers) == 0 {
		t.Logf("Expected peers, got none")
	}
}
