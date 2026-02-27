package media

import (
	"testing"
)

func TestSimpleAudioCapturer_Initialize(t *testing.T) {
	capturer, err := NewSimpleAudioCapturer(48000, 2)
	if err != nil {
		t.Fatalf("Failed to create capturer: %v", err)
	}

	caps := capturer.GetCapabilities()
	if caps == nil {
		t.Error("Expected capabilities")
	}

	if len(caps.SampleRates) == 0 {
		t.Error("Expected at least one sample rate")
	}
}

func TestSimpleAudioCapturer_StartStop(t *testing.T) {
	capturer, _ := NewSimpleAudioCapturer(48000, 2)

	err := capturer.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	err = capturer.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestSimpleAudioCapturer_SetSampleRate(t *testing.T) {
	capturer, _ := NewSimpleAudioCapturer(48000, 2)

	err := capturer.SetSampleRate(16000)
	if err != nil {
		t.Errorf("SetSampleRate failed: %v", err)
	}

	err = capturer.SetSampleRate(12345)
	if err == nil {
		t.Error("Expected error for invalid sample rate")
	}
}

func TestSimpleAudioCapturer_SetChannels(t *testing.T) {
	capturer, _ := NewSimpleAudioCapturer(48000, 2)

	err := capturer.SetChannels(1)
	if err != nil {
		t.Errorf("SetChannels failed: %v", err)
	}

	err = capturer.SetChannels(5)
	if err == nil {
		t.Error("Expected error for invalid channel count")
	}
}

func TestSimpleAudioCapturer_InvalidSampleRate(t *testing.T) {
	_, err := NewSimpleAudioCapturer(12345, 2)
	if err == nil {
		t.Error("Expected error for invalid sample rate")
	}
}

func TestSimpleAudioCapturer_InvalidChannels(t *testing.T) {
	_, err := NewSimpleAudioCapturer(48000, 5)
	if err == nil {
		t.Error("Expected error for invalid channel count")
	}
}

func TestSimpleAudioPlayer_Initialize(t *testing.T) {
	player := NewSimpleAudioPlayer()

	err := player.Initialize()
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}
}

func TestSimpleAudioPlayer_Play(t *testing.T) {
	player := NewSimpleAudioPlayer()
	player.Initialize()

	data := make([]byte, 480)
	err := player.Play(data)
	if err != nil {
		t.Errorf("Play failed: %v", err)
	}
}

func TestSimpleAudioPlayer_SetVolume(t *testing.T) {
	player := NewSimpleAudioPlayer()

	err := player.SetVolume(50)
	if err != nil {
		t.Errorf("SetVolume failed: %v", err)
	}

	err = player.SetVolume(150)
	if err == nil {
		t.Error("Expected error for volume > 100")
	}

	err = player.SetVolume(-10)
	if err == nil {
		t.Error("Expected error for volume < 0")
	}
}

func TestSimpleAudioPlayer_GetCapabilities(t *testing.T) {
	player := NewSimpleAudioPlayer()

	caps := player.GetCapabilities()
	if caps == nil {
		t.Error("Expected capabilities")
	}

	if len(caps.Channels) == 0 {
		t.Error("Expected at least one channel option")
	}
}

func TestSimpleAudioPlayer_Close(t *testing.T) {
	player := NewSimpleAudioPlayer()
	player.Initialize()

	err := player.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestAudioDeviceManager_GetInputDevices(t *testing.T) {
	manager := NewAudioDeviceManager()

	devices := manager.GetInputDevices()
	if len(devices) == 0 {
		t.Error("Expected at least one input device")
	}
}

func TestAudioDeviceManager_GetOutputDevices(t *testing.T) {
	manager := NewAudioDeviceManager()

	devices := manager.GetOutputDevices()
	if len(devices) == 0 {
		t.Error("Expected at least one output device")
	}
}

func TestAudioDeviceManager_SelectInputDevice(t *testing.T) {
	manager := NewAudioDeviceManager()

	err := manager.SelectInputDevice("default-mic")
	if err != nil {
		t.Errorf("SelectInputDevice failed: %v", err)
	}

	device := manager.GetSelectedInputDevice()
	if device == nil {
		t.Error("Expected selected input device")
	}
}

func TestAudioDeviceManager_SelectOutputDevice(t *testing.T) {
	manager := NewAudioDeviceManager()

	err := manager.SelectOutputDevice("default-speaker")
	if err != nil {
		t.Errorf("SelectOutputDevice failed: %v", err)
	}

	device := manager.GetSelectedOutputDevice()
	if device == nil {
		t.Error("Expected selected output device")
	}
}

func TestAudioMixer_AddStream(t *testing.T) {
	mixer := NewAudioMixer(5)

	err := mixer.AddStream("stream-1")
	if err != nil {
		t.Errorf("AddStream failed: %v", err)
	}
}

func TestAudioMixer_AddStream_MaxReached(t *testing.T) {
	mixer := NewAudioMixer(2)

	mixer.AddStream("stream-1")
	mixer.AddStream("stream-2")

	err := mixer.AddStream("stream-3")
	if err == nil {
		t.Error("Expected error when max streams reached")
	}
}

func TestAudioMixer_RemoveStream(t *testing.T) {
	mixer := NewAudioMixer(5)

	mixer.AddStream("stream-1")

	err := mixer.RemoveStream("stream-1")
	if err != nil {
		t.Errorf("RemoveStream failed: %v", err)
	}
}

func TestAudioMixer_SetStreamVolume(t *testing.T) {
	mixer := NewAudioMixer(5)

	mixer.AddStream("stream-1")

	err := mixer.SetStreamVolume("stream-1", 50)
	if err != nil {
		t.Errorf("SetStreamVolume failed: %v", err)
	}
}

func TestAudioMixer_SetStreamMuted(t *testing.T) {
	mixer := NewAudioMixer(5)

	mixer.AddStream("stream-1")

	err := mixer.SetStreamMuted("stream-1", true)
	if err != nil {
		t.Errorf("SetStreamMuted failed: %v", err)
	}
}

func TestAudioMixer_SetMasterVolume(t *testing.T) {
	mixer := NewAudioMixer(5)

	err := mixer.SetMasterVolume(80)
	if err != nil {
		t.Errorf("SetMasterVolume failed: %v", err)
	}
}

func TestAudioMixer_GetStreamCount(t *testing.T) {
	mixer := NewAudioMixer(5)

	mixer.AddStream("stream-1")
	mixer.AddStream("stream-2")

	count := mixer.GetStreamCount()
	if count != 2 {
		t.Errorf("Expected 2 streams, got %d", count)
	}
}

func TestAudioEncoder_Encode(t *testing.T) {
	codec, _ := NewOpusCodec(48000, 2, 64000)
	encoder, _ := NewAudioEncoder(codec, 48000, 2)

	data := make([]byte, 1920)
	for i := range data {
		data[i] = byte(i % 256)
	}

	encoded, err := encoder.Encode(data)
	if err != nil {
		t.Errorf("Encode failed: %v", err)
	}

	if len(encoded) == 0 {
		t.Error("Expected encoded data")
	}
}

func TestAudioEncoder_GetEncodedCount(t *testing.T) {
	codec, _ := NewOpusCodec(48000, 2, 64000)
	encoder, _ := NewAudioEncoder(codec, 48000, 2)

	data := make([]byte, 1920)
	encoder.Encode(data)

	count := encoder.GetEncodedCount()
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestAudioEncoder_Close(t *testing.T) {
	codec, _ := NewOpusCodec(48000, 2, 64000)
	encoder, _ := NewAudioEncoder(codec, 48000, 2)

	err := encoder.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestAudioDecoder_Decode(t *testing.T) {
	codec, _ := NewOpusCodec(48000, 2, 64000)
	decoder, _ := NewAudioDecoder(codec)

	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i % 256)
	}

	decoded, err := decoder.Decode(data)
	if err != nil {
		t.Logf("Decode failed (expected for random data): %v", err)
	}

	_ = decoded
}

func TestAudioDecoder_GetDecodedCount(t *testing.T) {
	codec, _ := NewOpusCodec(48000, 2, 64000)
	decoder, _ := NewAudioDecoder(codec)

	count := decoder.GetDecodedCount()
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestAudioDecoder_Close(t *testing.T) {
	codec, _ := NewOpusCodec(48000, 2, 64000)
	decoder, _ := NewAudioDecoder(codec)

	err := decoder.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestAudioLevelDetector_Process(t *testing.T) {
	detector := NewAudioLevelDetector(1000)

	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i % 256)
	}

	detected := detector.Process(data)
	t.Logf("Audio level detected: %v", detected)

	level := detector.GetLevel()
	t.Logf("Audio level: %f", level)
}

func TestAudioLevelDetector_SetThreshold(t *testing.T) {
	detector := NewAudioLevelDetector(1000)

	detector.SetThreshold(2000)

	level := detector.GetLevel()
	if level != 0 {
		t.Error("Expected level to be reset")
	}
}

func TestAudioSession_NewSession(t *testing.T) {
	session, err := NewAudioSession("session-1", "peer-1", 48000, 2)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.sessionID != "session-1" {
		t.Errorf("Expected session ID session-1, got %s", session.sessionID)
	}
}

func TestAudioSession_StartStop(t *testing.T) {
	session, _ := NewAudioSession("session-1", "peer-1", 48000, 2)

	err := session.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	err = session.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestAudioSession_Close(t *testing.T) {
	session, _ := NewAudioSession("session-1", "peer-1", 48000, 2)

	err := session.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestAudioSession_GetStats(t *testing.T) {
	session, _ := NewAudioSession("session-1", "peer-1", 48000, 2)

	stats := session.GetStats()
	if stats == nil {
		t.Error("Expected stats")
	}
}

func TestNewAudioSessionWithConfig(t *testing.T) {
	config := &MediaConfig{
		MediaType:  MediaTypeAudio,
		Codec:      "opus",
		SampleRate: 16000,
		Channels:   1,
		Bitrate:    64000,
	}

	session, err := NewAudioSessionWithConfig("session-1", "peer-1", config)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session == nil {
		t.Error("Expected session")
	}
}

func TestAudioDeviceFactory(t *testing.T) {
	factory := NewAudioDeviceFactory()

	capturer, err := factory.CreateCapturer(48000, 2)
	if err != nil {
		t.Fatalf("CreateCapturer failed: %v", err)
	}

	if capturer == nil {
		t.Error("Expected capturer")
	}

	player := factory.CreatePlayer()
	if player == nil {
		t.Error("Expected player")
	}

	manager := factory.CreateDeviceManager()
	if manager == nil {
		t.Error("Expected device manager")
	}

	mixer := factory.CreateMixer(5)
	if mixer == nil {
		t.Error("Expected mixer")
	}
}
