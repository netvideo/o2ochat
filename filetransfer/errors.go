package filetransfer

import "errors"

var (
	ErrFileNotFound      = errors.New("file not found")
	ErrInvalidChunkIndex = errors.New("invalid chunk index")
	ErrChunkVerification = errors.New("chunk verification failed")
	ErrFileVerification  = errors.New("file verification failed")
	ErrTaskNotFound      = errors.New("task not found")
	ErrTaskAlreadyExists = errors.New("task already exists")
	ErrInvalidTaskState  = errors.New("invalid task state for operation")
	ErrInsufficientPeers = errors.New("insufficient peers for transfer")
	ErrDownloadFailed    = errors.New("download failed")
	ErrUploadFailed      = errors.New("upload failed")
	ErrStorageError      = errors.New("storage error")
	ErrDiskSpaceExceeded = errors.New("disk space exceeded")
	ErrCancelled         = errors.New("transfer cancelled")
	ErrTimeout           = errors.New("transfer timeout")
	ErrMerkleTreeBuild   = errors.New("merkle tree build failed")
	ErrMerkleTreeVerify  = errors.New("merkle tree verify failed")
	ErrInvalidFileID     = errors.New("invalid file ID")
	ErrInvalidFilePath   = errors.New("invalid file path")
)

type FileTransferError struct {
	Code    string
	Message string
	Err     error
}

func (e *FileTransferError) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + " - " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *FileTransferError) Unwrap() error {
	return e.Err
}

func NewFileTransferError(code, message string, err error) *FileTransferError {
	return &FileTransferError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
