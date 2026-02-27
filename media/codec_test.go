package media

import (
	"testing"
)

func TestOpusCodec_Initialize(t *testing.T) {
	codec, err := NewOpusCodec(48000, 2, 64000)
	if err != nil {
		t.Fatalf("Failed to create Opus codec: %v", err)
	}

	info := codec.GetCodecInfo()
	if info.Name != "opus" {
		t.Errorf("Expected codec name 'opus', got '%s'", info.Name)
	}
	if info.SampleRate != 48000 {
		t.Errorf("Expected sample rate 48000, got %d", info.SampleRate)
	}
	if info.Channels != 2 {
		t.Errorf("Expected channels 2, got %d", info.Channels)
	}
}

func TestOpusCodec_EncodeDecode(t *testing.T) {
	codec, _ := NewOpusCodec(48000, 2, 64000)

	input := make([]byte, 1920)
	for i := range input {
		input[i] = byte(i % 256)
	}

	encoded, err := codec.EncodeFrame(input)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if len(encoded) == 0 {
		t.Fatal("Encoded data is empty")
	}

	decoded, err := codec.DecodeFrame(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(decoded) == 0 {
		t.Fatal("Decoded data is empty")
	}
}

func TestOpusCodec_InvalidSampleRate(t *testing.T) {
	_, err := NewOpusCodec(12345, 2, 64000)
	if err == nil {
		t.Error("Expected error for invalid sample rate")
	}
}

func TestOpusCodec_InvalidChannels(t *testing.T) {
	_, err := NewOpusCodec(48000, 5, 64000)
	if err == nil {
		t.Error("Expected error for invalid channel count")
	}
}

func TestOpusCodec_SetEncoderParams(t *testing.T) {
	codec, _ := NewOpusCodec(48000, 2, 64000)

	params := map[string]interface{}{
		"bitrate":     128000,
		"application": "audio",
	}

	err := codec.SetEncoderParams(params)
	if err != nil {
		t.Errorf("SetEncoderParams failed: %v", err)
	}
}

func TestOpusCodec_Reset(t *testing.T) {
	codec, _ := NewOpusCodec(48000, 2, 64000)

	codec.EncodeFrame([]byte{1, 2, 3, 4})
	codec.DecodeFrame([]byte{1, 2, 3, 4})

	err := codec.Reset()
	if err != nil {
		t.Errorf("Reset failed: %v", err)
	}
}

func TestAudioProcessor_Initialize(t *testing.T) {
	processor, err := NewAudioProcessor(48000, 2)
	if err != nil {
		t.Fatalf("Failed to create audio processor: %v", err)
	}

	if !processor.GetAECEnabled() {
		t.Error("AEC should be enabled by default")
	}
	if !processor.GetNSEnabled() {
		t.Error("NS should be enabled by default")
	}
	if !processor.GetAGCEnabled() {
		t.Error("AGC should be enabled by default")
	}
}

func TestAudioProcessor_ProcessInput(t *testing.T) {
	processor, _ := NewAudioProcessor(48000, 2)

	input := make([]byte, 1920)
	for i := range input {
		input[i] = byte(i % 256)
	}

	output, err := processor.ProcessInput(input)
	if err != nil {
		t.Fatalf("ProcessInput failed: %v", err)
	}

	if len(output) != len(input) {
		t.Errorf("Output length mismatch: expected %d, got %d", len(input), len(output))
	}
}

func TestAudioProcessor_ProcessOutput(t *testing.T) {
	processor, _ := NewAudioProcessor(48000, 2)

	input := make([]byte, 1920)
	for i := range input {
		input[i] = byte(i % 256)
	}

	output, err := processor.ProcessOutput(input)
	if err != nil {
		t.Fatalf("ProcessOutput failed: %v", err)
	}

	if len(output) != len(input) {
		t.Errorf("Output length mismatch: expected %d, got %d", len(input), len(output))
	}
}

func TestAudioProcessor_SetAEC(t *testing.T) {
	processor, _ := NewAudioProcessor(48000, 2)

	err := processor.SetAEC(false)
	if err != nil {
		t.Errorf("SetAEC failed: %v", err)
	}

	if processor.GetAECEnabled() {
		t.Error("AEC should be disabled")
	}
}

func TestAudioProcessor_SetNS(t *testing.T) {
	processor, _ := NewAudioProcessor(48000, 2)

	err := processor.SetNS(false)
	if err != nil {
		t.Errorf("SetNS failed: %v", err)
	}

	if processor.GetNSEnabled() {
		t.Error("NS should be disabled")
	}
}

func TestAudioProcessor_SetAGC(t *testing.T) {
	processor, _ := NewAudioProcessor(48000, 2)

	err := processor.SetAGC(false)
	if err != nil {
		t.Errorf("SetAGC failed: %v", err)
	}

	if processor.GetAGCEnabled() {
		t.Error("AGC should be disabled")
	}
}

func TestAudioProcessor_Close(t *testing.T) {
	processor, _ := NewAudioProcessor(48000, 2)

	err := processor.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestVP8Codec_Initialize(t *testing.T) {
	codec, err := NewVP8Codec(640, 480, 500000, 30)
	if err != nil {
		t.Fatalf("Failed to create VP8 codec: %v", err)
	}

	info := codec.GetCodecInfo()
	if info.Name != "vp8" {
		t.Errorf("Expected codec name 'vp8', got '%s'", info.Name)
	}
	if info.Width != 640 {
		t.Errorf("Expected width 640, got %d", info.Width)
	}
	if info.Height != 480 {
		t.Errorf("Expected height 480, got %d", info.Height)
	}
}

func TestVP8Codec_EncodeDecode(t *testing.T) {
	codec, _ := NewVP8Codec(640, 480, 500000, 30)

	input := make([]byte, 460800)
	for i := range input {
		input[i] = byte(i % 256)
	}

	encoded, err := codec.EncodeFrame(input)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if len(encoded) == 0 {
		t.Fatal("Encoded data is empty")
	}

	decoded, err := codec.DecodeFrame(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(decoded) == 0 {
		t.Fatal("Decoded data is empty")
	}
}

func TestVP8Codec_InvalidDimensions(t *testing.T) {
	_, err := NewVP8Codec(0, 480, 500000, 30)
	if err == nil {
		t.Error("Expected error for invalid width")
	}

	_, err = NewVP8Codec(640, 0, 500000, 30)
	if err == nil {
		t.Error("Expected error for invalid height")
	}
}

func TestCodec_NewCodec(t *testing.T) {
	audioConfig := &MediaConfig{
		MediaType:  MediaTypeAudio,
		Codec:      "opus",
		SampleRate: 48000,
		Channels:   2,
		Bitrate:    64000,
	}

	codec, err := NewCodec(MediaTypeAudio, audioConfig)
	if err != nil {
		t.Fatalf("Failed to create audio codec: %v", err)
	}

	if codec.GetCodecInfo().Name != "opus" {
		t.Errorf("Expected opus codec")
	}

	videoConfig := &MediaConfig{
		MediaType: MediaTypeVideo,
		Codec:     "vp8",
		Width:     640,
		Height:    480,
		Bitrate:   500000,
		FrameRate: 30,
	}

	codec, err = NewCodec(MediaTypeVideo, videoConfig)
	if err != nil {
		t.Fatalf("Failed to create video codec: %v", err)
	}

	if codec.GetCodecInfo().Name != "vp8" {
		t.Errorf("Expected vp8 codec")
	}
}

func TestCodec_Unsupported(t *testing.T) {
	config := &MediaConfig{
		MediaType: MediaTypeAudio,
		Codec:     "unsupported",
	}

	_, err := NewCodec(MediaTypeAudio, config)
	if err == nil {
		t.Error("Expected error for unsupported codec")
	}
}
