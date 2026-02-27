package filetransfer

import (
	"testing"
)

func TestFileTransferManager_New(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	scheduler := NewScheduler()

	manager := NewFileTransferManager(chunkMgr, scheduler, 4)
	if manager == nil {
		t.Fatal("NewFileTransferManager returned nil")
	}
}

func TestFileTransferManager_New_DefaultConcurrency(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	scheduler := NewScheduler()

	manager := NewFileTransferManager(chunkMgr, scheduler, 0)
	if manager == nil {
		t.Fatal("NewFileTransferManager returned nil")
	}
}

func TestFileTransferManager_CreateDownloadTask(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	scheduler := NewScheduler()
	manager := NewFileTransferManager(chunkMgr, scheduler, 4)

	taskID, err := manager.CreateDownloadTask("file123", "/tmp/download", []string{"peer1", "peer2"})
	if err != nil {
		t.Fatalf("CreateDownloadTask failed: %v", err)
	}

	if taskID == "" {
		t.Error("Empty task ID returned")
	}
}

func TestFileTransferManager_CreateDownloadTask_EmptyFileID(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	scheduler := NewScheduler()
	manager := NewFileTransferManager(chunkMgr, scheduler, 4)

	_, err := manager.CreateDownloadTask("", "/tmp/download", []string{"peer1"})
	if err == nil {
		t.Fatal("Expected error for empty file ID")
	}
}

func TestFileTransferManager_CreateDownloadTask_NoPeers(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	scheduler := NewScheduler()
	manager := NewFileTransferManager(chunkMgr, scheduler, 4)

	_, err := manager.CreateDownloadTask("file123", "/tmp/download", []string{})
	if err == nil {
		t.Fatal("Expected error for no peers")
	}
}

func TestFileTransferManager_CreateUploadTask(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	scheduler := NewScheduler()
	manager := NewFileTransferManager(chunkMgr, scheduler, 4)

	taskID, err := manager.CreateUploadTask("/tmp/nonexistent.txt", []string{"peer1"})
	if err == nil {
		t.Logf("CreateUploadTask returned task ID: %s", taskID)
	}
}

func TestFileTransferManager_CreateUploadTask_NoPeers(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	scheduler := NewScheduler()
	manager := NewFileTransferManager(chunkMgr, scheduler, 4)

	_, err := manager.CreateUploadTask("/tmp/nonexistent.txt", []string{})
	if err == nil {
		t.Fatal("Expected error for no peers")
	}
}

func TestFileTransferManager_GetTaskStatus_NotFound(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	scheduler := NewScheduler()
	manager := NewFileTransferManager(chunkMgr, scheduler, 4)

	_, err := manager.GetTaskStatus("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent task")
	}
}

func TestFileTransferManager_GetAllTasks_Empty(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	scheduler := NewScheduler()
	manager := NewFileTransferManager(chunkMgr, scheduler, 4)

	tasks, err := manager.GetAllTasks()
	if err != nil {
		t.Fatalf("GetAllTasks failed: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}
}

func TestFileTransferManager_GetAllTasks(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	scheduler := NewScheduler()
	manager := NewFileTransferManager(chunkMgr, scheduler, 4)

	manager.CreateDownloadTask("file1", "/tmp/dest", []string{"peer1"})
	manager.CreateDownloadTask("file2", "/tmp/dest", []string{"peer1"})

	tasks, err := manager.GetAllTasks()
	if err != nil {
		t.Fatalf("GetAllTasks failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}
}

func TestFileTransferManager_PauseTransfer_NotFound(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	scheduler := NewScheduler()
	manager := NewFileTransferManager(chunkMgr, scheduler, 4)

	err := manager.PauseTransfer("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent task")
	}
}

func TestFileTransferManager_ResumeTransfer_NotFound(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	scheduler := NewScheduler()
	manager := NewFileTransferManager(chunkMgr, scheduler, 4)

	err := manager.ResumeTransfer("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent task")
	}
}

func TestFileTransferManager_CancelTransfer_NotFound(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	scheduler := NewScheduler()
	manager := NewFileTransferManager(chunkMgr, scheduler, 4)

	err := manager.CancelTransfer("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent task")
	}
}

func TestFileTransferManager_CleanupCompletedTasks(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	scheduler := NewScheduler()
	manager := NewFileTransferManager(chunkMgr, scheduler, 4)

	taskID, _ := manager.CreateDownloadTask("file1", "/tmp/dest", []string{"peer1"})

	task, _ := manager.GetTaskStatus(taskID)
	task.Status = StatusCompleted

	err := manager.CleanupCompletedTasks()
	if err != nil {
		t.Fatalf("CleanupCompletedTasks failed: %v", err)
	}
}
