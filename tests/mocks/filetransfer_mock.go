package mocks

import (
	"github.com/netvideo/filetransfer"
	"github.com/stretchr/testify/mock"
)

// MockFileTransferManager 模拟文件传输管理器
type MockFileTransferManager struct {
	mock.Mock
}

func (m *MockFileTransferManager) ChunkFile(filePath string, chunkSize int64) (*filetransfer.FileMetadata, error) {
	args := m.Called(filePath, chunkSize)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*filetransfer.FileMetadata), args.Error(1)
}

func (m *MockFileTransferManager) MergeFile(outputPath string, metadata *filetransfer.FileMetadata) error {
	args := m.Called(outputPath, metadata)
	return args.Error(0)
}

func (m *MockFileTransferManager) CreateDownloadTask(fileID, destDir string, sources []string) (string, error) {
	args := m.Called(fileID, destDir, sources)
	return args.String(0), args.Error(1)
}

func (m *MockFileTransferManager) StartTransfer(taskID string) error {
	args := m.Called(taskID)
	return args.Error(0)
}

func (m *MockFileTransferManager) PauseTransfer(taskID string) error {
	args := m.Called(taskID)
	return args.Error(0)
}

func (m *MockFileTransferManager) ResumeTransfer(taskID string) error {
	args := m.Called(taskID)
	return args.Error(0)
}

func (m *MockFileTransferManager) CancelTransfer(taskID string) error {
	args := m.Called(taskID)
	return args.Error(0)
}

func (m *MockFileTransferManager) GetTaskStatus(taskID string) (*filetransfer.TransferTask, error) {
	args := m.Called(taskID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*filetransfer.TransferTask), args.Error(1)
}

func (m *MockFileTransferManager) ListTasks() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockFileTransferManager) SetProgressCallback(taskID string, callback filetransfer.ProgressCallback) {
	m.Called(taskID, callback)
}

func (m *MockFileTransferManager) SetEventHandler(handler filetransfer.EventHandler) {
	m.Called(handler)
}

func (m *MockFileTransferManager) Close() error {
	args := m.Called()
	return args.Error(0)
}
