package media

import (
	"testing"
	"time"
)

func TestMediaTypes(t *testing.T) {
	tests := []struct {
		name     string
		expected MediaType
	}{
		{"Audio", MediaTypeAudio},
		{"Video", MediaTypeVideo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.expected) == "" {
				t.Error("media type should not be empty")
			}
		})
	}
}

func TestMediaConfig(t *testing.T) {
	audioConfig := DefaultAudioConfig()

	if audioConfig.MediaType != MediaTypeAudio {
		t.Error("media type should be audio")
	}
	if audioConfig.Codec != "opus" {
		t.Error("codec should be opus")
	}
	if audioConfig.Bitrate <= 0 {
		t.Error("bitrate should be positive")
	}
	if audioConfig.SampleRate <= 0 {
		t.Error("sample rate should be positive")
	}
	if audioConfig.Channels <= 0 {
		t.Error("channels should be positive")
	}
}

func TestVideoConfig(t *testing.T) {
	videoConfig := DefaultVideoConfig()

	if videoConfig.MediaType != MediaTypeVideo {
		t.Error("media type should be video")
	}
	if videoConfig.Codec != "vp8" {
		t.Error("codec should be vp8")
	}
	if videoConfig.Width <= 0 {
		t.Error("width should be positive")
	}
	if videoConfig.Height <= 0 {
		t.Error("height should be positive")
	}
	if videoConfig.FrameRate <= 0 {
		t.Error("frame rate should be positive")
	}
}

func TestCallConfig(t *testing.T) {
	config := DefaultCallConfig()

	if config.AudioConfig == nil {
		t.Error("audio config should not be nil")
	}
	if config.VideoConfig == nil {
		t.Error("video config should not be nil")
	}
	if config.MaxBitrate <= 0 {
		t.Error("max bitrate should be positive")
	}
	if config.MinBitrate <= 0 {
		t.Error("min bitrate should be positive")
	}
	if config.StartBitrate < config.MinBitrate || config.StartBitrate > config.MaxBitrate {
		t.Error("start bitrate should be between min and max")
	}
}

func TestDeviceInfo(t *testing.T) {
	device := &DeviceInfo{
		ID:      "device-1",
		Name:    "Test Microphone",
		Type:    MediaTypeAudio,
		Default: true,
	}

	if device.ID == "" {
		t.Error("device ID should not be empty")
	}
	if device.Name == "" {
		t.Error("device name should not be empty")
	}
}

func TestMediaFrame(t *testing.T) {
	frame := &MediaFrame{
		Type:      MediaTypeAudio,
		Timestamp: 12345,
		Sequence:  1,
		Payload:   []byte("audio data"),
		Size:      10,
		KeyFrame:  false,
		Duration:  20 * time.Millisecond,
	}

	if frame.Type == "" {
		t.Error("frame type should not be empty")
	}
	if frame.Payload == nil {
		t.Error("payload should not be nil")
	}
	if frame.Duration <= 0 {
		t.Error("duration should be positive")
	}
}

func TestCallStats(t *testing.T) {
	stats := &CallStats{
		AudioStats: &StreamStats{
			Bitrate:    64000,
			PacketLoss: 0.01,
		},
		VideoStats: &StreamStats{
			Bitrate:    500000,
			PacketLoss: 0.02,
		},
		NetworkStats: &NetworkStats{
			Bandwidth:  1000000,
			PacketLoss: 0.015,
		},
		Quality: 0.9,
	}

	if stats.AudioStats == nil {
		t.Error("audio stats should not be nil")
	}
	if stats.VideoStats == nil {
		t.Error("video stats should not be nil")
	}
	if stats.Quality < 0 || stats.Quality > 1 {
		t.Error("quality should be between 0 and 1")
	}
}

func TestStreamStats(t *testing.T) {
	stats := &StreamStats{
		Bitrate:        500000,
		PacketLoss:     0.01,
		Jitter:         20 * time.Millisecond,
		Latency:        100 * time.Millisecond,
		FramesSent:     1000,
		FramesReceived: 990,
		FramesDropped:  10,
	}

	if stats.Bitrate <= 0 {
		t.Error("bitrate should be positive")
	}
	if stats.PacketLoss < 0 || stats.PacketLoss > 1 {
		t.Error("packet loss should be between 0 and 1")
	}
}

func TestCodecInfo(t *testing.T) {
	info := &CodecInfo{
		Name:       "opus",
		MediaType:  MediaTypeAudio,
		Bitrate:    64000,
		SampleRate: 48000,
		Channels:   2,
	}

	if info.Name == "" {
		t.Error("codec name should not be empty")
	}
}

func TestRTPPacket(t *testing.T) {
	packet := &RTPPacket{
		Version:     2,
		Marker:      true,
		PayloadType: 96,
		Sequence:    1234,
		Timestamp:   160,
		SSRC:        0x12345678,
		Payload:     []byte(" RTP payload"),
	}

	if packet.Version != 2 {
		t.Error("version should be 2")
	}
	if packet.Payload == nil {
		t.Error("payload should not be nil")
	}
}

func TestBufferStatus(t *testing.T) {
	status := &BufferStatus{
		Size:        100,
		MaxSize:     500,
		PacketCount: 50,
		FramesReady: 10,
	}

	if status.Size < 0 {
		t.Error("size should not be negative")
	}
	if status.MaxSize <= 0 {
		t.Error("max size should be positive")
	}
	if status.Size > status.MaxSize {
		t.Error("size should not exceed max size")
	}
}

func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		err  error
		name string
	}{
		{ErrNotInitialized, "ErrNotInitialized"},
		{ErrDeviceNotFound, "ErrDeviceNotFound"},
		{ErrDeviceNotAvailable, "ErrDeviceNotAvailable"},
		{ErrSessionNotFound, "ErrSessionNotFound"},
		{ErrSessionAlreadyExists, "ErrSessionAlreadyExists"},
		{ErrCodecNotSupported, "ErrCodecNotSupported"},
		{ErrEncodingFailed, "ErrEncodingFailed"},
		{ErrDecodingFailed, "ErrDecodingFailed"},
		{ErrFrameTooLarge, "ErrFrameTooLarge"},
		{ErrInvalidFrame, "ErrInvalidFrame"},
		{ErrRTPPacketizeFailed, "ErrRTPPacketizeFailed"},
		{ErrRTPDepacketizeFailed, "ErrRTPDepacketizeFailed"},
		{ErrBufferFull, "ErrBufferFull"},
		{ErrBufferEmpty, "ErrBufferEmpty"},
		{ErrCallNotActive, "ErrCallNotActive"},
		{ErrCallAlreadyActive, "ErrCallAlreadyActive"},
		{ErrBitrateAdjustFailed, "ErrBitrateAdjustFailed"},
		{ErrDeviceSwitchFailed, "ErrDeviceSwitchFailed"},
		{ErrPermissionDenied, "ErrPermissionDenied"},
		{ErrNoAudioDevice, "ErrNoAudioDevice"},
		{ErrNoVideoDevice, "ErrNoVideoDevice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("error should not be nil")
			}
			if tt.err.Error() == "" {
				t.Error("error message should not be empty")
			}
		})
	}
}

func TestMediaError(t *testing.T) {
	innerErr := ErrNotInitialized
	mediaErr := NewMediaError("NOT_INIT", "not initialized", innerErr)

	if mediaErr.Code != "NOT_INIT" {
		t.Errorf("expected code NOT_INIT, got %s", mediaErr.Code)
	}
	if mediaErr.Message != "not initialized" {
		t.Errorf("expected message 'not initialized', got %s", mediaErr.Message)
	}
	if mediaErr.Unwrap() != innerErr {
		t.Error("unwrap should return inner error")
	}
	if mediaErr.Error() == "" {
		t.Error("error should not be empty")
	}
}

func TestInterfaceCompatibility(t *testing.T) {
	var _ MediaManager = nil
	var _ CallSession = nil
	var _ Codec = nil
	var _ RTPProcessor = nil
	var _ JitterBuffer = nil
	var _ AudioProcessor = nil
	var _ VideoProcessor = nil
}
