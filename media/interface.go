package media

import "time"

type MediaManager interface {
	Initialize() error
	GetDevices(mediaType MediaType) ([]*DeviceInfo, error)
	CreateCallSession(config *CallConfig) (CallSession, error)
	JoinCall(sessionID string) (CallSession, error)
	LeaveCall(sessionID string) error
	GetCallStats(sessionID string) (*CallStats, error)
	Destroy() error
}

type CallSession interface {
	Start() error
	Stop() error
	Pause(mediaType MediaType) error
	Resume(mediaType MediaType) error
	SwitchDevice(mediaType MediaType, deviceID string) error
	AdjustBitrate(targetBitrate int) error
	SendFrame(frame *MediaFrame) error
	ReceiveFrame() (*MediaFrame, error)
	GetSessionID() string
	GetRemoteInfo() *PeerInfo
	GetStats() *CallStats
	Close() error
}

type Codec interface {
	EncodeFrame(input []byte) ([]byte, error)
	DecodeFrame(input []byte) ([]byte, error)
	GetCodecInfo() *CodecInfo
	SetEncoderParams(params map[string]interface{}) error
	Reset() error
	Close() error
}

type RTPProcessor interface {
	Packetize(frame *MediaFrame) ([]*RTPPacket, error)
	Depacketize(packet *RTPPacket) (*MediaFrame, error)
	HandleNACK(seqNums []uint16) ([]*RTPPacket, error)
	HandlePLI() error
	HandleFIR() error
	GetRTPStats() *RTPStats
	SetMaxPacketSize(size int) error
	Close() error
}

type JitterBuffer interface {
	AddPacket(packet *RTPPacket) error
	GetNextFrame() (*MediaFrame, error)
	SetBufferSize(size time.Duration) error
	GetBufferStatus() *BufferStatus
	Reset() error
	Close() error
}

type AudioProcessor interface {
	ProcessInput(data []byte) ([]byte, error)
	ProcessOutput(data []byte) ([]byte, error)
	SetAEC(enabled bool) error
	SetNS(enabled bool) error
	SetAGC(enabled bool) error
	Close() error
}

type VideoProcessor interface {
	ProcessFrame(frame []byte) ([]byte, error)
	Scale(width, height int) error
	Crop(x, y, width, height int) error
	Rotate(angle int) error
	ApplyFilter(filter string) error
}
