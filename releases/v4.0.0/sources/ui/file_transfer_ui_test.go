package ui

import (
	"testing"
)

func TestNewFileTransferUI(t *testing.T) {
	ft := NewFileTransferUI()
	if ft == nil {
		t.Error("expected non-nil FileTransferUI")
	}

	defaultFT, ok := ft.(*DefaultFileTransferUI)
	if !ok {
		t.Error("expected DefaultFileTransferUI type")
	}

	if defaultFT.tasks == nil {
		t.Error("expected tasks map to be initialized")
	}
}

func TestFileTransferUIAddTask(t *testing.T) {
	ft := NewFileTransferUI()

	task := &TransferTaskUI{
		TaskID:    "task1",
		FileName:  "test.pdf",
		FileSize:  1024,
		Direction: "upload",
	}

	err := ft.AddTransferTask(task)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = ft.AddTransferTask(nil)
	if err != ErrInvalidParameter {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}

	err = ft.AddTransferTask(&TransferTaskUI{TaskID: ""})
	if err != ErrInvalidParameter {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestFileTransferUIUpdateProgress(t *testing.T) {
	ft := NewFileTransferUI()

	task := &TransferTaskUI{
		TaskID:    "task1",
		FileName:  "test.pdf",
		FileSize:  1024,
		Direction: "upload",
		Progress:  0,
	}
	ft.AddTransferTask(task)

	err := ft.UpdateTransferProgress("task1", 50.0, 1024.0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = ft.UpdateTransferProgress("nonexistent", 50.0, 1024.0)
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestFileTransferUICompleteTask(t *testing.T) {
	ft := NewFileTransferUI()

	task := &TransferTaskUI{
		TaskID:    "task1",
		FileName:  "test.pdf",
		FileSize:  1024,
		Direction: "upload",
	}
	ft.AddTransferTask(task)

	err := ft.CompleteTransferTask("task1", true, "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = ft.CompleteTransferTask("nonexistent", true, "")
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestFileTransferUICancelTask(t *testing.T) {
	ft := NewFileTransferUI()

	task := &TransferTaskUI{
		TaskID:    "task1",
		FileName:  "test.pdf",
		FileSize:  1024,
	}
	ft.AddTransferTask(task)

	err := ft.CancelTransferTask("task1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = ft.CancelTransferTask("nonexistent")
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestFileTransferUIOpenFileLocation(t *testing.T) {
	ft := NewFileTransferUI()

	err := ft.OpenFileLocation("/path/to/file")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = ft.OpenFileLocation("")
	if err != ErrInvalidParameter {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestFileTransferUISetFileSelectCallback(t *testing.T) {
	ft := NewFileTransferUI()

	callbackCalled := false
	callback := func(filePaths []string) {
		callbackCalled = true
	}

	err := ft.SetFileSelectCallback(callback)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	defaultFT := ft.(*DefaultFileTransferUI)
	defaultFT.fileSelectCallback([]string{"file1.pdf", "file2.pdf"})
	if !callbackCalled {
		t.Error("expected callback to be called")
	}
}

func TestFileTransferUISetFolderSelectCallback(t *testing.T) {
	ft := NewFileTransferUI()

	callbackCalled := false
	callback := func(folderPath string) {
		callbackCalled = true
	}

	err := ft.SetFolderSelectCallback(callback)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	defaultFT := ft.(*DefaultFileTransferUI)
	defaultFT.folderSelectCallback("/downloads")
	if !callbackCalled {
		t.Error("expected callback to be called")
	}
}
