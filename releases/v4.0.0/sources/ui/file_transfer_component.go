package ui

import (
	"sort"
	"sync"
	"time"
)

type TransferTask struct {
	TaskID       string
	FileName     string
	FileSize     int64
	Direction    string
	PeerID       string
	PeerName     string
	Progress     float64
	Speed        float64
	Status       string
	StartTime    time.Time
	EndTime      time.Time
	ErrorMsg     string
}

type FileTransferComponent struct {
	mu              sync.RWMutex
	tasks           map[string]*TransferTask
	onTaskComplete  func(taskID string, success bool, errorMsg string)
	onTaskCancel    func(taskID string)
	onTaskClick     func(taskID string)
	onFileSelect    func()
	onFolderSelect  func()
}

func NewFileTransferComponent() *FileTransferComponent {
	return &FileTransferComponent{
		tasks: make(map[string]*TransferTask),
	}
}

func (ft *FileTransferComponent) AddTask(task *TransferTask) {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	task.StartTime = time.Now()
	task.Status = "pending"
	ft.tasks[task.TaskID] = task
}

func (ft *FileTransferComponent) RemoveTask(taskID string) {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	delete(ft.tasks, taskID)
}

func (ft *FileTransferComponent) GetTask(taskID string) (*TransferTask, bool) {
	ft.mu.RLock()
	defer ft.mu.RUnlock()
	task, ok := ft.tasks[taskID]
	return task, ok
}

func (ft *FileTransferComponent) GetAllTasks() []*TransferTask {
	ft.mu.RLock()
	defer ft.mu.RUnlock()

	tasks := make([]*TransferTask, 0, len(ft.tasks))
	for _, t := range ft.tasks {
		tasks = append(tasks, t)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].StartTime.After(tasks[j].StartTime)
	})

	return tasks
}

func (ft *FileTransferComponent) GetActiveTasks() []*TransferTask {
	ft.mu.RLock()
	defer ft.mu.RUnlock()

	var active []*TransferTask
	for _, t := range ft.tasks {
		if t.Status == "pending" || t.Status == "downloading" || t.Status == "uploading" {
			active = append(active, t)
		}
	}
	return active
}

func (ft *FileTransferComponent) GetCompletedTasks() []*TransferTask {
	ft.mu.RLock()
	defer ft.mu.RUnlock()

	var completed []*TransferTask
	for _, t := range ft.tasks {
		if t.Status == "completed" || t.Status == "failed" || t.Status == "cancelled" {
			completed = append(completed, t)
		}
	}
	return completed
}

func (ft *FileTransferComponent) UpdateProgress(taskID string, progress float64, speed float64) {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	if task, ok := ft.tasks[taskID]; ok {
		task.Progress = progress
		task.Speed = speed

		if progress >= 100 {
			task.Status = "completed"
			task.EndTime = time.Now()
		} else if progress > 0 {
			task.Status = "downloading"
		}
	}
}

func (ft *FileTransferComponent) CompleteTask(taskID string, success bool, errorMsg string) {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	if task, ok := ft.tasks[taskID]; ok {
		task.EndTime = time.Now()
		if success {
			task.Status = "completed"
			task.Progress = 100
		} else {
			task.Status = "failed"
			task.ErrorMsg = errorMsg
		}

		if ft.onTaskComplete != nil {
			ft.onTaskComplete(taskID, success, errorMsg)
		}
	}
}

func (ft *FileTransferComponent) CancelTask(taskID string) {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	if task, ok := ft.tasks[taskID]; ok {
		task.Status = "cancelled"
		task.EndTime = time.Now()

		if ft.onTaskCancel != nil {
			ft.onTaskCancel(taskID)
		}
	}
}

func (ft *FileTransferComponent) PauseTask(taskID string) {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	if task, ok := ft.tasks[taskID]; ok {
		if task.Status == "downloading" || task.Status == "uploading" {
			task.Status = "paused"
		}
	}
}

func (ft *FileTransferComponent) ResumeTask(taskID string) {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	if task, ok := ft.tasks[taskID]; ok {
		if task.Status == "paused" {
			task.Status = "downloading"
		}
	}
}

func (ft *FileTransferComponent) ClearCompletedTasks() {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	for id, task := range ft.tasks {
		if task.Status == "completed" || task.Status == "failed" || task.Status == "cancelled" {
			delete(ft.tasks, id)
		}
	}
}

func (ft *FileTransferComponent) GetTotalProgress() (totalSize int64, completedSize int64) {
	ft.mu.RLock()
	defer ft.mu.RUnlock()

	for _, t := range ft.tasks {
		totalSize += t.FileSize
		completedSize += int64(float64(t.FileSize) * t.Progress / 100)
	}
	return totalSize, completedSize
}

func (ft *FileTransferComponent) GetTotalSpeed() float64 {
	ft.mu.RLock()
	defer ft.mu.RUnlock()

	var totalSpeed float64
	for _, t := range ft.tasks {
		if t.Status == "downloading" || t.Status == "uploading" {
			totalSpeed += t.Speed
		}
	}
	return totalSpeed
}

func (ft *FileTransferComponent) SetOnTaskComplete(callback func(taskID string, success bool, errorMsg string)) {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	ft.onTaskComplete = callback
}

func (ft *FileTransferComponent) SetOnTaskCancel(callback func(taskID string)) {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	ft.onTaskCancel = callback
}

func (ft *FileTransferComponent) SetOnTaskClick(callback func(taskID string)) {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	ft.onTaskClick = callback
}

func (ft *FileTransferComponent) SetOnFileSelect(callback func()) {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	ft.onFileSelect = callback
}

func (ft *FileTransferComponent) SetOnFolderSelect(callback func()) {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	ft.onFolderSelect = callback
}

func (ft *FileTransferComponent) HandleFileSelect() {
	if ft.onFileSelect != nil {
		ft.onFileSelect()
	}
}

func (ft *FileTransferComponent) HandleFolderSelect() {
	if ft.onFolderSelect != nil {
		ft.onFolderSelect()
	}
}

func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return "< 1 KB"
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	sizeFloat := float64(size) / float64(div)
	return formatFloat(sizeFloat) + " " + []string{"KB", "MB", "GB", "TB"}[exp]
}

func formatFloat(f float64) string {
	if f == float64(int64(f)) {
		return string(rune('0' + int(f)))
	}
	return "0.0"
}

func formatSpeed(speed float64) string {
	if speed < 1024 {
		return "< 1 KB/s"
	}
	if speed < 1024*1024 {
		return formatFloat(speed/1024) + " KB/s"
	}
	return formatFloat(speed/1024/1024) + " MB/s"
}

func formatDuration(start, end time.Time) string {
	duration := end.Sub(start)
	if duration < time.Minute {
		return "< 1 min"
	}
	if duration < time.Hour {
		return formatFloat(duration.Minutes()) + " min"
	}
	return formatFloat(duration.Hours()) + " h"
}
