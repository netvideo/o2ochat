package filetransfer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type StateManager struct {
	storageDir string
	states     map[string]*TransferState
	mu         sync.RWMutex
}

func NewStateManager(storageDir string) (*StateManager, error) {
	if storageDir == "" {
		storageDir = "./filetransfer_state"
	}

	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, err
	}

	sm := &StateManager{
		storageDir: storageDir,
		states:     make(map[string]*TransferState),
	}

	if err := sm.loadAllStates(); err != nil {
		return nil, err
	}

	return sm, nil
}

func (sm *StateManager) loadAllStates() error {
	entries, err := os.ReadDir(sm.storageDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		taskID := entry.Name()[:len(entry.Name())-5]
		state, err := sm.loadState(taskID)
		if err != nil {
			continue
		}

		sm.states[taskID] = state
	}

	return nil
}

func (sm *StateManager) loadState(taskID string) (*TransferState, error) {
	statePath := filepath.Join(sm.storageDir, taskID+".json")

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, err
	}

	var state TransferState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

func (sm *StateManager) SaveState(state *TransferState) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.states[state.TaskID] = state

	statePath := filepath.Join(sm.storageDir, state.TaskID+".json")

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0644)
}

func (sm *StateManager) LoadState(taskID string) (*TransferState, error) {
	sm.mu.RLock()
	state, ok := sm.states[taskID]
	sm.mu.RUnlock()

	if ok {
		return state, nil
	}

	return sm.loadState(taskID)
}

func (sm *StateManager) DeleteState(taskID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.states, taskID)

	statePath := filepath.Join(sm.storageDir, taskID+".json")
	return os.Remove(statePath)
}

func (sm *StateManager) GetAllStates() []*TransferState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	states := make([]*TransferState, 0, len(sm.states))
	for _, state := range sm.states {
		states = append(states, state)
	}

	return states
}

func (sm *StateManager) CleanupOldStates(maxAge time.Duration) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)

	for taskID, state := range sm.states {
		isOld := false
		switch state.Status {
		case StatusCompleted, StatusFailed, StatusCancelled:
			isOld = state.PausedAt.Before(cutoff)
		}

		if isOld {
			delete(sm.states, taskID)
			statePath := filepath.Join(sm.storageDir, taskID+".json")
			os.Remove(statePath)
		}
	}

	return nil
}

func (sm *StateManager) CreateStateFromTask(task *TransferTask) *TransferState {
	return &TransferState{
		TaskID:          task.TaskID,
		FileID:          task.FileID,
		Direction:       task.Direction,
		SourcePeers:     task.SourcePeers,
		DestPath:        task.DestPath,
		StartedAt:       task.StartedAt,
		CompletedChunks: make([]int, 0),
		FailedChunks:    make([]int, 0),
		BytesTransferred: 0,
		TotalBytes:      int64(task.Progress.TotalChunks) * 1024 * 1024,
		Status:          task.Status,
	}
}

func (sm *StateManager) UpdateStateFromTask(task *TransferTask) error {
	state := sm.CreateStateFromTask(task)
	state.BytesTransferred = task.Progress.BytesTransferred

	return sm.SaveState(state)
}
