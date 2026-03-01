package webrtc

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/prop"
)

// MediaStream represents a media stream (audio/video)
type MediaStream struct {
	ID          string
	Track       *webrtc.TrackLocalStaticRTP
	Kind        string // "audio" or "video"
	IsLocal     bool
	IsSending   bool
	IsReceiving bool
	StartedAt   time.Time
	mu          sync.RWMutex
}

// MediaManager manages media streams for calls
type MediaManager struct {
	api          *webrtc.API
	localStreams map[string]*MediaStream
	remoteStreams map[string]*MediaStream
	config       *MediaConfig
	mu           sync.RWMutex
	stats        MediaStats
}

// MediaConfig represents media configuration
type MediaConfig struct {
	EnableAudio   bool
	EnableVideo   bool
	VideoWidth    int
	VideoHeight   int
	VideoFrameRate float32
	AudioBitrate  int
	VideoBitrate  int
}

// MediaStats represents media statistics
type MediaStats struct {
	TotalStreams      int
	ActiveStreams     int
	TotalBytesSent    uint64
	TotalBytesReceived uint64
}

// DefaultMediaConfig returns default media configuration
func DefaultMediaConfig() *MediaConfig {
	return &MediaConfig{
		EnableAudio:   true,
		EnableVideo:   true,
		VideoWidth:    640,
		VideoHeight:   480,
		VideoFrameRate: 30.0,
		AudioBitrate:  64000,
		VideoBitrate:  500000,
	}
}

// NewMediaManager creates a new media manager
func NewMediaManager(config *MediaConfig) (*MediaManager, error) {
	if config == nil {
		config = DefaultMediaConfig()
	}

	// Create WebRTC API
	api := webrtc.NewAPI()

	manager := &MediaManager{
		api:           api,
		localStreams:  make(map[string]*MediaStream),
		remoteStreams: make(map[string]*MediaStream),
		config:        config,
	}

	return manager, nil
}

// StartLocalStream starts capturing local media stream
func (mm *MediaManager) StartLocalStream(ctx context.Context, callID string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Check if already started
	if _, exists := mm.localStreams[callID]; exists {
		return errors.New("local stream already started")
	}

	// Initialize media devices
	err := mediadevices.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize media devices: %w", err)
	}

	// Get user media
	stream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(constraint *mediadevices.MediaTrackConstraints) {
			if mm.config.EnableVideo {
				constraint.Width = prop.Int(mm.config.VideoWidth)
				constraint.Height = prop.Int(mm.config.VideoHeight)
				constraint.FrameRate = prop.Float32(mm.config.VideoFrameRate)
			}
		},
		Audio: func(constraint *mediadevices.MediaTrackConstraints) {
			if mm.config.EnableAudio {
				// Audio constraints
			}
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get user media: %w", err)
	}

	// Create tracks from stream
	tracks := stream.GetTracks()
	for _, track := range tracks {
		kind := "audio"
		if track.Kind() == webrtc.RTPCodecTypeVideo {
			kind = "video"
		}

		mediaStream := &MediaStream{
			ID:        callID + "-" + kind,
			Track:     track.(*webrtc.TrackLocalStaticRTP),
			Kind:      kind,
			IsLocal:   true,
			IsSending: true,
			StartedAt: time.Now(),
		}

		mm.localStreams[callID] = mediaStream
		mm.stats.TotalStreams++
		mm.stats.ActiveStreams++
	}

	return nil
}

// AddRemoteTrack adds a remote track to media manager
func (mm *MediaManager) AddRemoteTrack(callID string, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	kind := "audio"
	if track.Kind() == webrtc.RTPCodecTypeVideo {
		kind = "video"
	}

	mediaStream := &MediaStream{
		ID:          callID + "-" + kind + "-remote",
		Track:       nil, // Remote track
		Kind:        kind,
		IsLocal:     false,
		IsReceiving: true,
		StartedAt:   time.Now(),
	}

	mm.remoteStreams[callID] = mediaStream
	mm.stats.TotalStreams++
	mm.stats.ActiveStreams++

	return nil
}

// StopLocalStream stops local media stream
func (mm *MediaManager) StopLocalStream(callID string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	stream, exists := mm.localStreams[callID]
	if !exists {
		return errors.New("local stream not found")
	}

	stream.IsSending = false
	stream.mu.Lock()
	stream.mu.Unlock()

	delete(mm.localStreams, callID)
	mm.stats.ActiveStreams--

	return nil
}

// GetLocalStream gets local media stream
func (mm *MediaManager) GetLocalStream(callID string) (*MediaStream, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	stream, exists := mm.localStreams[callID]
	if !exists {
		return nil, errors.New("local stream not found")
	}

	return stream, nil
}

// GetRemoteStream gets remote media stream
func (mm *MediaManager) GetRemoteStream(callID string) (*MediaStream, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	stream, exists := mm.remoteStreams[callID]
	if !exists {
		return nil, errors.New("remote stream not found")
	}

	return stream, nil
}

// GetStats gets media statistics
func (mm *MediaManager) GetStats() MediaStats {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return mm.stats
}

// EnableAudio enables/disables audio
func (mm *MediaManager) EnableAudio(callID string, enable bool) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	stream, exists := mm.localStreams[callID]
	if !exists {
		return errors.New("stream not found")
	}

	if stream.Kind != "audio" {
		return errors.New("not an audio stream")
	}

	stream.IsSending = enable

	return nil
}

// EnableVideo enables/disables video
func (mm *MediaManager) EnableVideo(callID string, enable bool) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	stream, exists := mm.localStreams[callID]
	if !exists {
		return errors.New("stream not found")
	}

	if stream.Kind != "video" {
		return errors.New("not a video stream")
	}

	stream.IsSending = enable

	return nil
}

// SwitchCamera switches camera (front/back)
func (mm *MediaManager) SwitchCamera(callID string) error {
	// TODO: Implement camera switching
	// This requires platform-specific implementation
	return errors.New("camera switching not implemented")
}

// SetVideoQuality sets video quality
func (mm *MediaManager) SetVideoQuality(callID string, width, height int) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Update config
	mm.config.VideoWidth = width
	mm.config.VideoHeight = height

	// Restart stream with new quality
	if stream, exists := mm.localStreams[callID]; exists {
		if stream.Kind == "video" {
			// Stop and restart stream
			mm.StopLocalStream(callID)
			mm.mu.Unlock()
			err := mm.StartLocalStream(context.Background(), callID)
			mm.mu.Lock()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// MuteAudio mutes audio
func (mm *MediaManager) MuteAudio(callID string) error {
	return mm.EnableAudio(callID, false)
}

// UnmuteAudio unmutes audio
func (mm *MediaManager) UnmuteAudio(callID string) error {
	return mm.EnableAudio(callID, true)
}

// StartVideo starts video
func (mm *MediaManager) StartVideo(callID string) error {
	return mm.EnableVideo(callID, true)
}

// StopVideo stops video
func (mm *MediaManager) StopVideo(callID string) error {
	return mm.EnableVideo(callID, false)
}

// GetAllStreams gets all streams
func (mm *MediaManager) GetAllStreams() ([]*MediaStream, []*MediaStream) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	local := make([]*MediaStream, 0, len(mm.localStreams))
	for _, stream := range mm.localStreams {
		local = append(local, stream)
	}

	remote := make([]*MediaStream, 0, len(mm.remoteStreams))
	for _, stream := range mm.remoteStreams {
		remote = append(remote, stream)
	}

	return local, remote
}

// Cleanup cleans up media manager
func (mm *MediaManager) Cleanup() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Stop all local streams
	for callID := range mm.localStreams {
		mm.StopLocalStream(callID)
	}

	// Clear remote streams
	mm.remoteStreams = make(map[string]*MediaStream)
}

// MediaStreamToJSON converts media stream to JSON
func MediaStreamToJSON(stream *MediaStream) map[string]interface{} {
	stream.mu.RLock()
	defer stream.mu.RUnlock()

	return map[string]interface{}{
		"id":           stream.ID,
		"kind":         stream.Kind,
		"is_local":     stream.IsLocal,
		"is_sending":   stream.IsSending,
		"is_receiving": stream.IsReceiving,
		"started_at":   stream.StartedAt,
	}
}
