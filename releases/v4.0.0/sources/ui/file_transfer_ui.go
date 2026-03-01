package ui

import (
	"sync"
)

type DefaultFileTransferUI struct {
	mu                  sync.RWMutex
	tasks               map[string]*TransferTaskUI
	fileSelectCallback  func(filePaths []string)
	folderSelectCallback func(folderPath string)
}

func NewFileTransferUI() FileTransferUI {
	return &DefaultFileTransferUI{
		tasks: make(map[string]*TransferTaskUI),
	}
}

func (f *DefaultFileTransferUI) ShowFileTransfer() error {
	return nil
}

func (f *DefaultFileTransferUI) AddTransferTask(task *TransferTaskUI) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if task == nil || task.TaskID == "" {
		return ErrInvalidParameter
	}

	f.tasks[task.TaskID] = task
	return nil
}

func (f *DefaultFileTransferUI) UpdateTransferProgress(taskID string, progress float64, speed float64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	task, ok := f.tasks[taskID]
	if !ok {
		return ErrTaskNotFound
	}

	task.Progress = progress
	task.Speed = speed
	return nil
}

func (f *DefaultFileTransferUI) CompleteTransferTask(taskID string, success bool, errorMsg string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	task, ok := f.tasks[taskID]
	if !ok {
		return ErrTaskNotFound
	}

	if success {
		task.Status = "completed"
		task.Progress = 100.0
	} else {
		task.Status = "failed"
	}

	delete(f.tasks, taskID)
	return nil
}

func (f *DefaultFileTransferUI) CancelTransferTask(taskID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.tasks[taskID]; !ok {
		return ErrTaskNotFound
	}

	delete(f.tasks, taskID)
	return nil
}

func (f *DefaultFileTransferUI) OpenFileLocation(filePath string) error {
	if filePath == "" {
		return ErrInvalidParameter
	}

	return nil
}

func (f *DefaultFileTransferUI) SetFileSelectCallback(callback func(filePaths []string)) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.fileSelectCallback = callback
	return nil
}

func (f *DefaultFileTransferUI) SetFolderSelectCallback(callback func(folderPath string)) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.folderSelectCallback = callback
	return nil
}
