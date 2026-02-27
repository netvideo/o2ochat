package media

import (
	"encoding/binary"
	"errors"
	"sync"
	"time"
)

var (
	ErrEncodeFailed = errors.New("encode failed")
	ErrDecodeFailed = errors.New("decode failed")
	ErrInvalidInput = errors.New("invalid input")
)

type OpusCodec struct {
	mu          sync.RWMutex
	bitrate     int
	sampleRate  int
	channels    int
	application string
	frameSize   int
	initialized bool
	encodeCount int64
	decodeCount int64
}

func NewOpusCodec(sampleRate, channels, bitrate int) (*OpusCodec, error) {
	if sampleRate != 8000 && sampleRate != 16000 && sampleRate != 24000 && sampleRate != 48000 {
		return nil, errors.New("unsupported sample rate")
	}
	if channels != 1 && channels != 2 {
		return nil, errors.New("unsupported channel count")
	}

	codec := &OpusCodec{
		sampleRate:  sampleRate,
		channels:    channels,
		bitrate:     bitrate,
		application: "voip",
		frameSize:   sampleRate * 20 / 1000,
		initialized: true,
	}

	return codec, nil
}

func (c *OpusCodec) EncodeFrame(input []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return nil, ErrCodecNotSupported
	}

	if len(input) == 0 {
		return nil, ErrInvalidInput
	}

	frameSize := c.frameSize * c.channels * 2
	if len(input) < frameSize {
		padding := make([]byte, frameSize-len(input))
		input = append(input, padding...)
	}

	encoded := c.encodeOPUS(input)
	c.encodeCount++

	return encoded, nil
}

func (c *OpusCodec) encodeOPUS(input []byte) []byte {
	encoded := make([]byte, len(input)/10+2)
	encoded[0] = 0x80
	encoded[1] = byte(c.sampleRate / 1000)

	hash := uint16(0)
	for i, b := range input {
		hash ^= uint16(b) << (i % 8)
	}
	binary.BigEndian.PutUint16(encoded[2:], hash)

	copy(encoded[4:], input)

	return encoded
}

func (c *OpusCodec) DecodeFrame(input []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return nil, ErrCodecNotSupported
	}

	if len(input) < 4 {
		return nil, ErrInvalidInput
	}

	decoded := c.decodeOPUS(input)
	c.decodeCount++

	return decoded, nil
}

func (c *OpusCodec) decodeOPUS(input []byte) []byte {
	if len(input) < 4 {
		return nil
	}

	frameSize := c.frameSize * c.channels * 2
	decoded := make([]byte, frameSize)

	if len(input) > 4 {
		copy(decoded, input[4:])
	}

	return decoded
}

func (c *OpusCodec) GetCodecInfo() *CodecInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &CodecInfo{
		Name:       "opus",
		MediaType:  MediaTypeAudio,
		Bitrate:    c.bitrate,
		SampleRate: c.sampleRate,
		Channels:   c.channels,
	}
}

func (c *OpusCodec) SetEncoderParams(params map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := params["bitrate"].(int); ok {
		c.bitrate = v
	}
	if v, ok := params["application"].(string); ok {
		c.application = v
	}
	if v, ok := params["sample_rate"].(int); ok {
		c.sampleRate = v
		c.frameSize = v * 20 / 1000
	}
	if v, ok := params["channels"].(int); ok {
		c.channels = v
	}

	return nil
}

func (c *OpusCodec) Reset() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.encodeCount = 0
	c.decodeCount = 0

	return nil
}

func (c *OpusCodec) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.initialized = false
	return nil
}

func (c *OpusCodec) GetEncodeCount() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.encodeCount
}

func (c *OpusCodec) GetDecodeCount() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.decodeCount
}

type AudioProcessorImpl struct {
	mu           sync.RWMutex
	aecEnabled   bool
	nsEnabled    bool
	agcEnabled   bool
	sampleRate   int
	channels     int
	frameSize    int
	initialized  bool
	inputBuffer  []byte
	outputBuffer []byte
}

func NewAudioProcessor(sampleRate, channels int) (*AudioProcessorImpl, error) {
	if sampleRate <= 0 || channels <= 0 {
		return nil, errors.New("invalid sample rate or channel count")
	}

	processor := &AudioProcessorImpl{
		sampleRate:  sampleRate,
		channels:    channels,
		frameSize:   sampleRate * 20 / 1000 * channels * 2,
		aecEnabled:  true,
		nsEnabled:   true,
		agcEnabled:  true,
		initialized: true,
	}

	return processor, nil
}

func (p *AudioProcessorImpl) ProcessInput(data []byte) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return nil, errors.New("processor not initialized")
	}

	if len(data) == 0 {
		return nil, ErrInvalidInput
	}

	processed := make([]byte, len(data))
	copy(processed, data)

	if p.nsEnabled {
		processed = p.applyNS(processed)
	}

	if p.agcEnabled {
		processed = p.applyAGC(processed)
	}

	p.inputBuffer = append(p.inputBuffer, processed...)

	return processed, nil
}

func (p *AudioProcessorImpl) ProcessOutput(data []byte) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return nil, errors.New("processor not initialized")
	}

	if len(data) == 0 {
		return nil, ErrInvalidInput
	}

	processed := make([]byte, len(data))
	copy(processed, data)

	if p.aecEnabled {
		processed = p.applyAEC(processed)
	}

	p.outputBuffer = append(p.outputBuffer, processed...)

	return processed, nil
}

func (p *AudioProcessorImpl) applyNS(data []byte) []byte {
	if len(data) < 2 {
		return data
	}

	threshold := float64(500)
	for i := 0; i < len(data)-1; i += 2 {
		sample := int16(binary.LittleEndian.Uint16(data[i:]))
		absSample := float64(abs(int(sample)))

		if absSample < threshold {
			attenuated := float64(sample) * 0.5
			binary.LittleEndian.PutUint16(data[i:i+2], uint16(int16(attenuated)))
		}
	}

	return data
}

func (p *AudioProcessorImpl) applyAGC(data []byte) []byte {
	if len(data) < 2 {
		return data
	}

	targetLevel := float64(16000)
	var maxSample float64

	for i := 0; i < len(data)-1; i += 2 {
		sample := int16(binary.LittleEndian.Uint16(data[i:]))
		absSample := float64(abs(int(sample)))
		if absSample > maxSample {
			maxSample = absSample
		}
	}

	if maxSample > 0 && maxSample < targetLevel {
		gain := targetLevel / maxSample
		if gain > 4.0 {
			gain = 4.0
		}

		for i := 0; i < len(data)-1; i += 2 {
			sample := int16(binary.LittleEndian.Uint16(data[i:]))
			gained := float64(sample) * gain
			if gained > 32767 {
				gained = 32767
			} else if gained < -32768 {
				gained = -32768
			}
			binary.LittleEndian.PutUint16(data[i:i+2], uint16(int16(gained)))
		}
	}

	return data
}

func (p *AudioProcessorImpl) applyAEC(data []byte) []byte {
	if len(p.inputBuffer) < len(data) || len(data) < 2 {
		return data
	}

	delay := p.frameSize
	if len(p.inputBuffer) > delay {
		refSignal := p.inputBuffer[len(p.inputBuffer)-delay-len(data) : len(p.inputBuffer)-delay]

		for i := 0; i < len(data)-1 && i < len(refSignal)-1; i += 2 {
			outputSample := int16(binary.LittleEndian.Uint16(data[i:]))
			refSample := int16(binary.LittleEndian.Uint16(refSignal[i:]))

			correlation := int32(outputSample) * int32(refSample)
			if correlation > 0 {
				adapted := int32(refSample) * 80 / 100
				outputSample -= int16(adapted)
			}

			binary.LittleEndian.PutUint16(data[i:i+2], uint16(outputSample))
		}
	}

	return data
}

func (p *AudioProcessorImpl) SetAEC(enabled bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return errors.New("processor not initialized")
	}

	p.aecEnabled = enabled
	return nil
}

func (p *AudioProcessorImpl) SetNS(enabled bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return errors.New("processor not initialized")
	}

	p.nsEnabled = enabled
	return nil
}

func (p *AudioProcessorImpl) SetAGC(enabled bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return errors.New("processor not initialized")
	}

	p.agcEnabled = enabled
	return nil
}

func (p *AudioProcessorImpl) GetAECEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.aecEnabled
}

func (p *AudioProcessorImpl) GetNSEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.nsEnabled
}

func (p *AudioProcessorImpl) GetAGCEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.agcEnabled
}

func (p *AudioProcessorImpl) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.initialized = false
	p.inputBuffer = nil
	p.outputBuffer = nil

	return nil
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

type VP8Codec struct {
	mu               sync.RWMutex
	bitrate          int
	width            int
	height           int
	frameRate        int
	keyFrameInterval int
	initialized      bool
	encodeCount      int64
	decodeCount      int64
	keyFrameCount    int64
}

func NewVP8Codec(width, height, bitrate, frameRate int) (*VP8Codec, error) {
	if width <= 0 || height <= 0 {
		return nil, errors.New("invalid width or height")
	}

	codec := &VP8Codec{
		width:            width,
		height:           height,
		bitrate:          bitrate,
		frameRate:        frameRate,
		keyFrameInterval: 3000,
		initialized:      true,
	}

	return codec, nil
}

func (c *VP8Codec) EncodeFrame(input []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return nil, ErrCodecNotSupported
	}

	if len(input) == 0 {
		return nil, ErrInvalidInput
	}

	isKeyFrame := c.encodeCount%int64(c.frameRate*int(c.keyFrameInterval)/1000) == 0
	if isKeyFrame {
		c.keyFrameCount++
	}

	encoded := c.encodeVP8(input, isKeyFrame)
	c.encodeCount++

	return encoded, nil
}

func (c *VP8Codec) encodeVP8(input []byte, isKeyFrame bool) []byte {
	encoded := make([]byte, len(input)/4+10)
	encoded[0] = 0x00

	if isKeyFrame {
		encoded[0] |= 0x01
	}

	encoded[1] = byte(c.width >> 8)
	encoded[2] = byte(c.width & 0xFF)
	encoded[3] = byte(c.height >> 8)
	encoded[4] = byte(c.height & 0xFF)

	hash := uint16(0)
	for i, b := range input {
		hash ^= uint16(b) << (i % 8)
	}
	binary.BigEndian.PutUint16(encoded[5:], hash)

	copy(encoded[7:], input)

	return encoded
}

func (c *VP8Codec) DecodeFrame(input []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return nil, ErrCodecNotSupported
	}

	if len(input) < 7 {
		return nil, ErrInvalidInput
	}

	decoded := c.decodeVP8(input)
	c.decodeCount++

	return decoded, nil
}

func (c *VP8Codec) decodeVP8(input []byte) []byte {
	frameSize := c.width * c.height * 3 / 2
	decoded := make([]byte, frameSize)

	if len(input) > 7 {
		copy(decoded, input[7:])
	}

	return decoded
}

func (c *VP8Codec) GetCodecInfo() *CodecInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &CodecInfo{
		Name:      "vp8",
		MediaType: MediaTypeVideo,
		Bitrate:   c.bitrate,
		Width:     c.width,
		Height:    c.height,
		FrameRate: c.frameRate,
	}
}

func (c *VP8Codec) SetEncoderParams(params map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := params["target_bitrate"].(int); ok {
		c.bitrate = v
	}
	if v, ok := params["keyframe_interval"].(int); ok {
		c.keyFrameInterval = v
	}
	if v, ok := params["width"].(int); ok {
		c.width = v
	}
	if v, ok := params["height"].(int); ok {
		c.height = v
	}

	return nil
}

func (c *VP8Codec) Reset() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.encodeCount = 0
	c.decodeCount = 0

	return nil
}

func (c *VP8Codec) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.initialized = false
	return nil
}

func (c *VP8Codec) IsKeyFrame(input []byte) bool {
	if len(input) < 1 {
		return false
	}
	return (input[0] & 0x01) != 0
}

func NewCodec(mediaType MediaType, config *MediaConfig) (Codec, error) {
	switch mediaType {
	case MediaTypeAudio:
		if config.Codec == "" {
			config.Codec = "opus"
		}
		if config.Codec == "opus" {
			return NewOpusCodec(config.SampleRate, config.Channels, config.Bitrate)
		}
	case MediaTypeVideo:
		if config.Codec == "" {
			config.Codec = "vp8"
		}
		if config.Codec == "vp8" {
			return NewVP8Codec(config.Width, config.Height, config.Bitrate, config.FrameRate)
		}
	}

	return nil, ErrCodecNotSupported
}

type AudioFrame struct {
	Samples    []int16
	SampleRate int
	Channels   int
	Timestamp  time.Duration
}

func NewAudioFrame(sampleRate, channels int, samples []int16) *AudioFrame {
	return &AudioFrame{
		Samples:    samples,
		SampleRate: sampleRate,
		Channels:   channels,
	}
}

func (f *AudioFrame) ToBytes() []byte {
	data := make([]byte, len(f.Samples)*2)
	for i, sample := range f.Samples {
		binary.LittleEndian.PutUint16(data[i*2:], uint16(sample))
	}
	return data
}

func AudioFrameFromBytes(data []byte, sampleRate, channels int) *AudioFrame {
	samples := make([]int16, len(data)/2)
	for i := range samples {
		samples[i] = int16(binary.LittleEndian.Uint16(data[i*2:]))
	}
	return &AudioFrame{
		Samples:    samples,
		SampleRate: sampleRate,
		Channels:   channels,
	}
}
