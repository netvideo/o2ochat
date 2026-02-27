package media

import (
	"time"
)

type MediaType string

const (
	MediaTypeAudio MediaType = "audio"
	MediaTypeVideo MediaType = "video"
)

type MediaConfig struct {
	MediaType        MediaType `json:"media_type"`
	Enabled          bool      `json:"enabled"`
	Codec            string    `json:"codec"`
	Bitrate          int       `json:"bitrate"`
	SampleRate       int       `json:"sample_rate"`
	Channels         int       `json:"channels"`
	Width            int       `json:"width"`
	Height           int       `json:"height"`
	FrameRate        int       `json:"frame_rate"`
	KeyFrameInterval int       `json:"key_frame_interval"`
}

type DeviceInfo struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Type     MediaType `json:"type"`
	Default  bool     `json:"default"`
}

type MediaFrame struct {
	Type      MediaType     `json:"type"`
	Timestamp uint32        `json:"timestamp"`
	Sequence  uint16        `json:"sequence"`
	Payload   []byte        `json:"payload"`
	Size      int           `json:"size"`
	KeyFrame  bool          `json:"key_frame"`
	Duration  time.Duration `json:"duration"`
}

type CallConfig struct {
	AudioConfig  *MediaConfig `json:"audio_config"`
	VideoConfig  *MediaConfig `json:"video_config"`
	MaxBitrate   int          `json:"max_bitrate"`
	MinBitrate   int          `json:"min_bitrate"`
	StartBitrate int          `json:"start_bitrate"`
	UseFEC       bool         `json:"use_fec"`
	UseNACK      bool         `json:"use_nack"`
	UsePLI       bool         `json:"use_pli"`
}

type CallStats struct {
	AudioStats  *StreamStats  `json:"audio_stats"`
	VideoStats  *StreamStats  `json:"video_stats"`
	NetworkStats *NetworkStats `json:"network_stats"`
	Quality     float64       `json:"quality"`
}

type StreamStats struct {
	Bitrate          int           `json:"bitrate"`
	PacketLoss       float64       `json:"packet_loss"`
	Jitter           time.Duration `json:"jitter"`
	Latency          time.Duration `json:"latency"`
	FramesSent       int64         `json:"frames_sent"`
	FramesReceived   int64         `json:"frames_received"`
	FramesDropped    int64         `json:"frames_dropped"`
}

type NetworkStats struct {
	Bandwidth      int64         `json:"bandwidth"`
	PacketLoss     float64       `json:"packet_loss"`
	Jitter         time.Duration `json:"jitter"`
	Latency        time.Duration `json:"latency"`
	Retransmits    int           `json:"retransmits"`
}

type PeerInfo struct {
	PeerID    string   `json:"peer_id"`
	PublicKey []byte   `json:"public_key"`
	AudioMuted bool   `json:"audio_muted"`
	VideoMuted bool   `json:"video_muted"`
}

type CodecInfo struct {
	Name        string `json:"name"`
	MediaType   MediaType `json:"media_type"`
	Bitrate     int    `json:"bitrate"`
	SampleRate  int    `json:"sample_rate"`
	Channels    int    `json:"channels"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	FrameRate   int    `json:"frame_rate"`
}

type RTPPacket struct {
	Version    uint8
	Padding    bool
	Extension  bool
	CSRCCount  uint8
	Marker     bool
	PayloadType uint8
	Sequence   uint16
	Timestamp  uint32
	SSRC       uint32
	CSRC       []uint32
	ExtensionHeader []byte
	Payload    []byte
}

type RTPStats struct {
	PacketsSent      uint64 `json:"packets_sent"`
	PacketsReceived  uint64 `json:"packets_received"`
	PacketsLost     uint64 `json:"packets_lost"`
	Jitter          uint32 `json:"jitter"`
	RoundTripTime   time.Duration `json:"round_trip_time"`
}

type BufferStatus struct {
	Size       int           `json:"size"`
	MaxSize    int           `json:"max_size"`
	PacketCount int          `json:"packet_count"`
	FramesReady int          `json:"frames_ready"`
}

func DefaultAudioConfig() *MediaConfig {
	return &MediaConfig{
		MediaType:  MediaTypeAudio,
		Enabled:    true,
		Codec:      "opus",
		Bitrate:    64000,
		SampleRate: 48000,
		Channels:   2,
	}
}

func DefaultVideoConfig() *MediaConfig {
	return &MediaConfig{
		MediaType:  MediaTypeVideo,
		Enabled:    true,
		Codec:      "vp8",
		Bitrate:    500000,
		Width:      640,
		Height:     480,
		FrameRate:  30,
		KeyFrameInterval: 3000,
	}
}

func DefaultCallConfig() *CallConfig {
	return &CallConfig{
		AudioConfig:  DefaultAudioConfig(),
		VideoConfig:  DefaultVideoConfig(),
		MaxBitrate:   1000000,
		MinBitrate:   100000,
		StartBitrate: 500000,
		UseFEC:       true,
		UseNACK:      true,
		UsePLI:       true,
	}
}
