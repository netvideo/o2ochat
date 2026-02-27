package media

import (
	"bytes"
	"encoding/binary"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrAudioDeviceNotAvailable = errors.New("audio device not available")
	ErrAudioNotRunning         = errors.New("audio not running")
	ErrBufferOverflow          = errors.New("buffer overflow")
	ErrInvalidSampleRate       = errors.New("invalid sample rate")
)

type AudioCapturer interface {
	Start() error
	Stop() error
	GetFrame() ([]byte, error)
	SetSampleRate(rate int) error
	SetChannels(channels int) error
	GetCapabilities() *AudioCapabilities
}

type AudioCapabilities struct {
	SampleRates   []int
	Channels      []int
	BitsPerSample int
	BufferSize    int
}

type AudioPlayer interface {
	Initialize() error
	Play(data []byte) error
	Stop() error
	SetVolume(volume int) error
	GetCapabilities() *AudioCapabilities
	Close() error
}

type SimpleAudioCapturer struct {
	mu            sync.RWMutex
	sampleRate    int
	channels      int
	bitsPerSample int
	running       bool
	frameCount    uint64
	sampleCount   uint64
	buffer        *bytes.Buffer
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

func NewSimpleAudioCapturer(sampleRate, channels int) (*SimpleAudioCapturer, error) {
	if sampleRate != 8000 && sampleRate != 16000 && sampleRate != 32000 && sampleRate != 48000 {
		return nil, ErrInvalidSampleRate
	}
	if channels != 1 && channels != 2 {
		return nil, errors.New("invalid channel count")
	}

	capturer := &SimpleAudioCapturer{
		sampleRate:    sampleRate,
		channels:      channels,
		bitsPerSample: 16,
		running:       false,
		buffer:        new(bytes.Buffer),
		stopChan:      make(chan struct{}),
	}

	return capturer, nil
}

func (c *SimpleAudioCapturer) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return nil
	}

	c.running = true
	c.buffer.Reset()
	c.stopChan = make(chan struct{})

	c.wg.Add(1)
	go c.captureLoop()

	return nil
}

func (c *SimpleAudioCapturer) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return nil
	}

	close(c.stopChan)
	c.wg.Wait()
	c.running = false

	return nil
}

func (c *SimpleAudioCapturer) captureLoop() {
	defer c.wg.Done()

	frameSize := c.sampleRate / 100 * c.channels * 2
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			samples := make([]int16, frameSize/2)
			for i := range samples {
				samples[i] = int16((i * 12345) % 32768)
			}

			c.mu.Lock()
			for _, sample := range samples {
				var buf [2]byte
				binary.LittleEndian.PutUint16(buf[:], uint16(sample))
				c.buffer.Write(buf[:])
			}
			c.sampleCount += uint64(len(samples))
			c.mu.Unlock()
		}
	}
}

func (c *SimpleAudioCapturer) GetFrame() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.running {
		return nil, ErrAudioNotRunning
	}

	frameSize := c.sampleRate / 100 * c.channels * 2
	if c.buffer.Len() < frameSize {
		return nil, errors.New("insufficient data")
	}

	frame := make([]byte, frameSize)
	n, err := c.buffer.Read(frame)
	if err != nil {
		return nil, err
	}

	if n < frameSize {
		frame = frame[:n]
	}

	c.frameCount++

	return frame, nil
}

func (c *SimpleAudioCapturer) SetSampleRate(rate int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return errors.New("cannot change sample rate while running")
	}

	if rate != 8000 && rate != 16000 && rate != 32000 && rate != 48000 {
		return ErrInvalidSampleRate
	}

	c.sampleRate = rate
	return nil
}

func (c *SimpleAudioCapturer) SetChannels(channels int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return errors.New("cannot change channels while running")
	}

	if channels != 1 && channels != 2 {
		return errors.New("invalid channel count")
	}

	c.channels = channels
	return nil
}

func (c *SimpleAudioCapturer) GetCapabilities() *AudioCapabilities {
	return &AudioCapabilities{
		SampleRates:   []int{8000, 16000, 32000, 48000},
		Channels:      []int{1, 2},
		BitsPerSample: 16,
		BufferSize:    4096,
	}
}

type SimpleAudioPlayer struct {
	mu            sync.RWMutex
	sampleRate    int
	channels      int
	bitsPerSample int
	volume        int
	initialized   bool
	playing       bool
	bytesPlayed   uint64
	framesPlayed  uint64
	device        *AudioDeviceInfo
}

type AudioDeviceInfo struct {
	ID         string
	Name       string
	SampleRate int
	Channels   int
}

func NewSimpleAudioPlayer() *SimpleAudioPlayer {
	return &SimpleAudioPlayer{
		sampleRate:    48000,
		channels:      2,
		bitsPerSample: 16,
		volume:        100,
		initialized:   false,
		playing:       false,
		device: &AudioDeviceInfo{
			ID:   "default-audio-output",
			Name: "Default Speaker",
		},
	}
}

func (p *SimpleAudioPlayer) Initialize() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return nil
	}

	p.initialized = true
	return nil
}

func (p *SimpleAudioPlayer) Play(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return errors.New("player not initialized")
	}

	if len(data) == 0 {
		return nil
	}

	if p.volume > 0 {
		data = p.applyVolume(data)
	}

	p.bytesPlayed += uint64(len(data))
	p.framesPlayed++

	return nil
}

func (p *SimpleAudioPlayer) applyVolume(data []byte) []byte {
	if p.volume == 100 || len(data) < 2 {
		return data
	}

	volumeFactor := float64(p.volume) / 100.0
	result := make([]byte, len(data))

	for i := 0; i < len(data)-1; i += 2 {
		sample := int16(binary.LittleEndian.Uint16(data[i:]))
		sampled := float64(sample) * volumeFactor
		if sampled > 32767 {
			sampled = 32767
		} else if sampled < -32768 {
			sampled = -32768
		}
		binary.LittleEndian.PutUint16(result[i:], uint16(int16(sampled)))
	}

	copy(result[len(data):], data[len(data):])

	return result
}

func (p *SimpleAudioPlayer) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.playing = false
	return nil
}

func (p *SimpleAudioPlayer) SetVolume(volume int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if volume < 0 || volume > 100 {
		return errors.New("volume must be between 0 and 100")
	}

	p.volume = volume
	return nil
}

func (p *SimpleAudioPlayer) GetCapabilities() *AudioCapabilities {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return &AudioCapabilities{
		SampleRates:   []int{8000, 16000, 32000, 48000},
		Channels:      []int{1, 2},
		BitsPerSample: 16,
		BufferSize:    4096,
	}
}

func (p *SimpleAudioPlayer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.initialized = false
	p.playing = false

	return nil
}

type AudioDeviceManager struct {
	mu             sync.RWMutex
	inputDevices   []*AudioDeviceInfo
	outputDevices  []*AudioDeviceInfo
	selectedInput  *AudioDeviceInfo
	selectedOutput *AudioDeviceInfo
}

func NewAudioDeviceManager() *AudioDeviceManager {
	return &AudioDeviceManager{
		inputDevices: []*AudioDeviceInfo{
			{
				ID:         "default-mic",
				Name:       "Default Microphone",
				SampleRate: 48000,
				Channels:   2,
			},
		},
		outputDevices: []*AudioDeviceInfo{
			{
				ID:         "default-speaker",
				Name:       "Default Speaker",
				SampleRate: 48000,
				Channels:   2,
			},
		},
		selectedInput:  nil,
		selectedOutput: nil,
	}
}

func (m *AudioDeviceManager) GetInputDevices() []*AudioDeviceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]*AudioDeviceInfo, len(m.inputDevices))
	copy(devices, m.inputDevices)

	return devices
}

func (m *AudioDeviceManager) GetOutputDevices() []*AudioDeviceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]*AudioDeviceInfo, len(m.outputDevices))
	copy(devices, m.outputDevices)

	return devices
}

func (m *AudioDeviceManager) SelectInputDevice(deviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, device := range m.inputDevices {
		if device.ID == deviceID {
			m.selectedInput = device
			return nil
		}
	}

	return errors.New("input device not found")
}

func (m *AudioDeviceManager) SelectOutputDevice(deviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, device := range m.outputDevices {
		if device.ID == deviceID {
			m.selectedOutput = device
			return nil
		}
	}

	return errors.New("output device not found")
}

func (m *AudioDeviceManager) GetSelectedInputDevice() *AudioDeviceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.selectedInput != nil {
		return m.selectedInput
	}

	if len(m.inputDevices) > 0 {
		return m.inputDevices[0]
	}

	return nil
}

func (m *AudioDeviceManager) GetSelectedOutputDevice() *AudioDeviceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.selectedOutput != nil {
		return m.selectedOutput
	}

	if len(m.outputDevices) > 0 {
		return m.outputDevices[0]
	}

	return nil
}

type AudioMixer struct {
	mu           sync.RWMutex
	streams      map[string]*AudioStream
	maxStreams   int
	masterVolume int
	initialized  bool
}

type AudioStream struct {
	ID       string
	volume   int
	muted    bool
	active   bool
	dataChan chan []byte
}

func NewAudioMixer(maxStreams int) *AudioMixer {
	return &AudioMixer{
		streams:      make(map[string]*AudioStream),
		maxStreams:   maxStreams,
		masterVolume: 100,
		initialized:  true,
	}
}

func (m *AudioMixer) AddStream(streamID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.streams) >= m.maxStreams {
		return errors.New("max streams reached")
	}

	if _, exists := m.streams[streamID]; exists {
		return errors.New("stream already exists")
	}

	stream := &AudioStream{
		ID:       streamID,
		volume:   100,
		muted:    false,
		active:   true,
		dataChan: make(chan []byte, 100),
	}

	m.streams[streamID] = stream

	return nil
}

func (m *AudioMixer) RemoveStream(streamID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	stream, exists := m.streams[streamID]
	if !exists {
		return errors.New("stream not found")
	}

	close(stream.dataChan)
	delete(m.streams, streamID)

	return nil
}

func (m *AudioMixer) SetStreamVolume(streamID string, volume int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	stream, exists := m.streams[streamID]
	if !exists {
		return errors.New("stream not found")
	}

	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}

	stream.volume = volume

	return nil
}

func (m *AudioMixer) SetStreamMuted(streamID string, muted bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	stream, exists := m.streams[streamID]
	if !exists {
		return errors.New("stream not found")
	}

	stream.muted = muted

	return nil
}

func (m *AudioMixer) Mix() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.streams) == 0 {
		return nil, nil
	}

	samples := make([]int32, 1024)
	activeCount := 0

	for _, stream := range m.streams {
		if !stream.active || stream.muted {
			continue
		}

		activeCount++

		select {
		case data := <-stream.dataChan:
			if len(data) >= 2 {
				for i := 0; i < len(data)/2 && i < len(samples); i++ {
					sample := int16(binary.LittleEndian.Uint16(data[i*2:]))
					volumeFactor := float64(stream.volume) * float64(m.masterVolume) / 10000.0
					samples[i] += int32(float64(sample) * volumeFactor)
				}
			}
		default:
		}
	}

	if activeCount == 0 {
		return nil, nil
	}

	result := make([]byte, len(samples)*2)
	for i, sample := range samples {
		if sample > 32767 {
			sample = 32767
		} else if sample < -32768 {
			sample = -32768
		}
		binary.LittleEndian.PutUint16(result[i*2:], uint16(int16(sample)))
	}

	return result, nil
}

func (m *AudioMixer) SetMasterVolume(volume int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}

	m.masterVolume = volume

	return nil
}

func (m *AudioMixer) GetStreamCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.streams)
}

type AudioEncoder struct {
	mu           sync.RWMutex
	codec        Codec
	sampleRate   int
	channels     int
	frameSize    int
	initialized  bool
	encodedCount uint64
}

func NewAudioEncoder(codec Codec, sampleRate, channels int) (*AudioEncoder, error) {
	if sampleRate <= 0 || channels <= 0 {
		return nil, errors.New("invalid parameters")
	}

	frameSize := sampleRate / 1000 * 20 * channels * 2

	encoder := &AudioEncoder{
		codec:       codec,
		sampleRate:  sampleRate,
		channels:    channels,
		frameSize:   frameSize,
		initialized: true,
	}

	return encoder, nil
}

func (e *AudioEncoder) Encode(data []byte) ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.initialized {
		return nil, errors.New("encoder not initialized")
	}

	if len(data) < e.frameSize {
		padding := make([]byte, e.frameSize-len(data))
		data = append(data, padding...)
	}

	encoded, err := e.codec.EncodeFrame(data)
	if err != nil {
		return nil, err
	}

	e.encodedCount++

	return encoded, nil
}

func (e *AudioEncoder) GetEncodedCount() uint64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.encodedCount
}

func (e *AudioEncoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.initialized = false

	if e.codec != nil {
		return e.codec.Close()
	}

	return nil
}

type AudioDecoder struct {
	mu           sync.RWMutex
	codec        Codec
	initialized  bool
	decodedCount uint64
}

func NewAudioDecoder(codec Codec) (*AudioDecoder, error) {
	if codec == nil {
		return nil, errors.New("codec cannot be nil")
	}

	decoder := &AudioDecoder{
		codec:       codec,
		initialized: true,
	}

	return decoder, nil
}

func (d *AudioDecoder) Decode(data []byte) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.initialized {
		return nil, errors.New("decoder not initialized")
	}

	decoded, err := d.codec.DecodeFrame(data)
	if err != nil {
		return nil, err
	}

	d.decodedCount++

	return decoded, nil
}

func (d *AudioDecoder) GetDecodedCount() uint64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.decodedCount
}

func (d *AudioDecoder) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.initialized = false

	if d.codec != nil {
		return d.codec.Close()
	}

	return nil
}

type AudioLevelDetector struct {
	mu          sync.RWMutex
	threshold   float64
	detecting   bool
	level       float64
	sampleCount uint64
}

func NewAudioLevelDetector(threshold float64) *AudioLevelDetector {
	return &AudioLevelDetector{
		threshold: threshold,
		detecting: true,
		level:     0,
	}
}

func (d *AudioLevelDetector) Process(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	var sum float64
	sampleCount := 0

	for i := 0; i < len(data)-1; i += 2 {
		sample := int16(binary.LittleEndian.Uint16(data[i:]))
		sum += float64(sample * sample)
		sampleCount++
	}

	if sampleCount > 0 {
		rms := sum / float64(sampleCount)
		d.level = rms

		d.mu.RLock()
		threshold := d.threshold
		d.mu.RUnlock()

		return rms > threshold*threshold
	}

	return false
}

func (d *AudioLevelDetector) GetLevel() float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.level
}

func (d *AudioLevelDetector) SetThreshold(threshold float64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.threshold = threshold
}

type AudioSession struct {
	mu        sync.RWMutex
	capturer  *SimpleAudioCapturer
	player    *SimpleAudioPlayer
	encoder   *AudioEncoder
	decoder   *AudioDecoder
	processor AudioProcessor
	mixer     *AudioMixer
	running   bool
	sessionID string
	remoteID  string
	stats     *AudioSessionStats
}

type AudioSessionStats struct {
	FramesCaptured uint64
	FramesPlayed   uint64
	BytesSent      uint64
	BytesReceived  uint64
	DroppedFrames  uint64
	CaptureLatency time.Duration
	PlayLatency    time.Duration
}

func NewAudioSession(sessionID, remoteID string, sampleRate, channels int) (*AudioSession, error) {
	capturer, err := NewSimpleAudioCapturer(sampleRate, channels)
	if err != nil {
		return nil, err
	}

	player := NewSimpleAudioPlayer()

	opusCodec, err := NewOpusCodec(sampleRate, channels, 64000)
	if err != nil {
		return nil, err
	}

	encoder, err := NewAudioEncoder(opusCodec, sampleRate, channels)
	if err != nil {
		return nil, err
	}

	decoderOpus, _ := NewOpusCodec(sampleRate, channels, 64000)
	decoder, err := NewAudioDecoder(decoderOpus)
	if err != nil {
		return nil, err
	}

	processor, err := NewAudioProcessor(sampleRate, channels)
	if err != nil {
		return nil, err
	}

	mixer := NewAudioMixer(10)

	session := &AudioSession{
		capturer:  capturer,
		player:    player,
		encoder:   encoder,
		decoder:   decoder,
		processor: processor,
		mixer:     mixer,
		running:   false,
		sessionID: sessionID,
		remoteID:  remoteID,
		stats:     &AudioSessionStats{},
	}

	return session, nil
}

func (s *AudioSession) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	if err := s.capturer.Start(); err != nil {
		return err
	}

	if err := s.player.Initialize(); err != nil {
		return err
	}

	s.running = true

	return nil
}

func (s *AudioSession) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.capturer.Stop()
	s.player.Stop()

	s.running = false

	return nil
}

func (s *AudioSession) CaptureAndEncode() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running {
		return nil, ErrAudioNotRunning
	}

	frame, err := s.capturer.GetFrame()
	if err != nil {
		return nil, err
	}

	processed, err := s.processor.ProcessInput(frame)
	if err != nil {
		return nil, err
	}

	encoded, err := s.encoder.Encode(processed)
	if err != nil {
		return nil, err
	}

	atomic.AddUint64(&s.stats.FramesCaptured, 1)
	atomic.AddUint64(&s.stats.BytesSent, uint64(len(encoded)))

	return encoded, nil
}

func (s *AudioSession) DecodeAndPlay(data []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running {
		return ErrAudioNotRunning
	}

	decoded, err := s.decoder.Decode(data)
	if err != nil {
		return err
	}

	processed, err := s.processor.ProcessOutput(decoded)
	if err != nil {
		return err
	}

	if err := s.player.Play(processed); err != nil {
		return err
	}

	atomic.AddUint64(&s.stats.FramesPlayed, 1)
	atomic.AddUint64(&s.stats.BytesReceived, uint64(len(data)))

	return nil
}

func (s *AudioSession) GetStats() *AudioSessionStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &AudioSessionStats{
		FramesCaptured: atomic.LoadUint64(&s.stats.FramesCaptured),
		FramesPlayed:   atomic.LoadUint64(&s.stats.FramesPlayed),
		BytesSent:      atomic.LoadUint64(&s.stats.BytesSent),
		BytesReceived:  atomic.LoadUint64(&s.stats.BytesReceived),
		DroppedFrames:  atomic.LoadUint64(&s.stats.DroppedFrames),
	}
}

func (s *AudioSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.running = false

	s.capturer.Stop()
	s.player.Close()
	s.encoder.Close()
	s.decoder.Close()
	s.processor.Close()

	return nil
}

func NewAudioSessionWithConfig(sessionID, remoteID string, config *MediaConfig) (*AudioSession, error) {
	if config == nil {
		config = DefaultAudioConfig()
	}

	return NewAudioSession(sessionID, remoteID, config.SampleRate, config.Channels)
}

type AudioDeviceFactory struct{}

func NewAudioDeviceFactory() *AudioDeviceFactory {
	return &AudioDeviceFactory{}
}

func (f *AudioDeviceFactory) CreateCapturer(sampleRate, channels int) (AudioCapturer, error) {
	return NewSimpleAudioCapturer(sampleRate, channels)
}

func (f *AudioDeviceFactory) CreatePlayer() AudioPlayer {
	return NewSimpleAudioPlayer()
}

func (f *AudioDeviceFactory) CreateDeviceManager() *AudioDeviceManager {
	return NewAudioDeviceManager()
}

func (f *AudioDeviceFactory) CreateMixer(maxStreams int) *AudioMixer {
	return NewAudioMixer(maxStreams)
}
