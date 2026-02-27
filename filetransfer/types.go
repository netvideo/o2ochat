package filetransfer

import "time"

type TransferStatus string

const (
	StatusPending     TransferStatus = "pending"
	StatusDownloading TransferStatus = "downloading"
	StatusUploading   TransferStatus = "uploading"
	StatusPaused      TransferStatus = "paused"
	StatusCompleted   TransferStatus = "completed"
	StatusFailed      TransferStatus = "failed"
	StatusCancelled   TransferStatus = "cancelled"
)

type TransferDirection string

const (
	DirectionDownload TransferDirection = "download"
	DirectionUpload   TransferDirection = "upload"
)

type FileMetadata struct {
	FileID      string    `json:"file_id"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	TotalChunks int       `json:"total_chunks"`
	ChunkSize   int       `json:"chunk_size"`
	MerkleRoot  []byte    `json:"merkle_root"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
}

type ChunkInfo struct {
	FileID    string `json:"file_id"`
	Index     int    `json:"index"`
	Offset    int64  `json:"offset"`
	Size      int    `json:"size"`
	Hash      []byte `json:"hash"`
	Completed bool   `json:"completed"`
	Verified  bool   `json:"verified"`
}

type TransferTask struct {
	TaskID      string            `json:"task_id"`
	FileID      string            `json:"file_id"`
	Direction   TransferDirection `json:"direction"`
	SourcePeers []string          `json:"source_peers"`
	DestPath    string            `json:"dest_path"`
	StartedAt   time.Time         `json:"started_at"`
	Status      TransferStatus    `json:"status"`
	Progress    TransferProgress  `json:"progress"`
}

type TransferProgress struct {
	TotalChunks      int     `json:"total_chunks"`
	Completed        int     `json:"completed"`
	Failed           int     `json:"failed"`
	BytesTransferred int64   `json:"bytes_transferred"`
	Speed            float64 `json:"speed"`
	EstimatedTime    string  `json:"estimated_time"`
}

type SchedulerStats struct {
	TotalScheduled    int         `json:"total_scheduled"`
	Successful        int         `json:"successful"`
	Failed            int         `json:"failed"`
	Pending           int         `json:"pending"`
	ChunkAvailability map[int]int `json:"chunk_availability"`
}

type ChunkData struct {
	Index int
	Data  []byte
}

type DownloadRequest struct {
	TaskID   string
	FileID   string
	ChunkIdx int
	PeerID   string
}

type UploadRequest struct {
	TaskID      string
	FileID      string
	ChunkIdx    int
	TargetPeers []string
}

type TransferState struct {
	TaskID          string            `json:"task_id"`
	FileID          string            `json:"file_id"`
	Direction       TransferDirection `json:"direction"`
	SourcePeers     []string          `json:"source_peers"`
	DestPath        string            `json:"dest_path"`
	StartedAt       time.Time         `json:"started_at"`
	PausedAt        time.Time         `json:"paused_at"`
	CompletedChunks []int             `json:"completed_chunks"`
	FailedChunks    []int             `json:"failed_chunks"`
	BytesTransferred int64            `json:"bytes_transferred"`
	TotalBytes      int64             `json:"total_bytes"`
	Status          TransferStatus    `json:"status"`
}

type TransferProgressCallback func(taskID string, progress TransferProgress)

type TransferEventType string

const (
	EventStarted   TransferEventType = "started"
	EventProgress  TransferEventType = "progress"
	EventCompleted TransferEventType = "completed"
	EventFailed    TransferEventType = "failed"
	EventPaused    TransferEventType = "paused"
	EventResumed   TransferEventType = "resumed"
	EventCancelled TransferEventType = "cancelled"
)

type TransferEvent struct {
	EventType TransferEventType
	TaskID    string
	FileID    string
	Timestamp time.Time
	Error     error
}
