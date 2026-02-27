package filetransfer

type FileTransferManager interface {
	CreateDownloadTask(fileID, destPath string, sourcePeers []string) (string, error)
	CreateUploadTask(filePath string, targetPeers []string) (string, error)
	StartTransfer(taskID string) error
	PauseTransfer(taskID string) error
	ResumeTransfer(taskID string) error
	CancelTransfer(taskID string) error
	GetTaskStatus(taskID string) (*TransferTask, error)
	GetAllTasks() ([]*TransferTask, error)
	CleanupCompletedTasks() error
}

type ChunkManager interface {
	ChunkFile(filePath string, chunkSize int) (*FileMetadata, error)
	MergeFile(fileID, outputPath string) error
	GetChunkInfo(fileID string, index int) (*ChunkInfo, error)
	GetAllChunks(fileID string) ([]*ChunkInfo, error)
	VerifyChunk(fileID string, index int, data []byte) (bool, error)
	VerifyFile(fileID string) (bool, error)
	SaveChunk(fileID string, index int, data []byte) error
	ReadChunk(fileID string, index int) ([]byte, error)
}

type Scheduler interface {
	SelectNextChunk(fileID string, availableChunks map[int][]string) (int, []string, error)
	AssignDownloadTasks(fileID string, maxConcurrent int) (map[int][]string, error)
	UpdateChunkStatus(fileID string, index int, success bool, peerID string) error
	GetSchedulerStats(fileID string) (*SchedulerStats, error)
}

type MerkleTree interface {
	BuildTree(chunks [][]byte) ([][]byte, error)
	GetRootHash() []byte
	VerifyChunk(index int, chunkData []byte, proof [][]byte) (bool, error)
	GenerateProof(index int) ([][]byte, error)
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}
