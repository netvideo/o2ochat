package media

import (
	"errors"
	"image"
	"sync"
)

var (
	ErrVideoProcessorImplNotInitialized = errors.New("video processor not initialized")
	ErrInvalidDimensions                = errors.New("invalid dimensions")
	ErrUnsupportedFilter                = errors.New("unsupported filter")
	ErrInvalidAngle                     = errors.New("invalid angle")
)

type VideoProcessorImpl struct {
	mu           sync.RWMutex
	width        int
	height       int
	targetWidth  int
	targetHeight int
	initialized  bool
	filters      []string
	rotation     int
	flipH        bool
	flipV        bool
	brightness   float64
	contrast     float64
	saturation   float64
}

func NewVideoProcessorImpl(width, height int) (*VideoProcessorImpl, error) {
	if width <= 0 || height <= 0 {
		return nil, ErrInvalidDimensions
	}

	processor := &VideoProcessorImpl{
		width:        width,
		height:       height,
		targetWidth:  width,
		targetHeight: height,
		initialized:  true,
		brightness:   0,
		contrast:     1,
		saturation:   1,
	}

	return processor, nil
}

func (p *VideoProcessorImpl) ProcessFrame(frame []byte) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return nil, ErrVideoProcessorImplNotInitialized
	}

	if len(frame) == 0 {
		return nil, ErrInvalidInput
	}

	processed := make([]byte, len(frame))
	copy(processed, frame)

	if p.brightness != 0 || p.contrast != 1 || p.saturation != 1 {
		processed = p.applyColorAdjustment(processed)
	}

	return processed, nil
}

func (p *VideoProcessorImpl) Scale(width, height int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if width <= 0 || height <= 0 {
		return ErrInvalidDimensions
	}

	if width == p.width && height == p.height {
		return nil
	}

	p.targetWidth = width
	p.targetHeight = height

	return nil
}

func (p *VideoProcessorImpl) Crop(x, y, width, height int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if x < 0 || y < 0 || width <= 0 || height <= 0 {
		return ErrInvalidDimensions
	}

	if x+width > p.width || y+height > p.height {
		return ErrInvalidDimensions
	}

	p.targetWidth = width
	p.targetHeight = height

	return nil
}

func (p *VideoProcessorImpl) Rotate(angle int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch angle {
	case 0, 90, 180, 270:
		p.rotation = angle
	case -90, -180, -270:
		p.rotation = (360 + angle) % 360
	default:
		return ErrInvalidAngle
	}

	if p.rotation == 90 || p.rotation == 270 {
		p.targetWidth = p.height
		p.targetHeight = p.width
	} else {
		p.targetWidth = p.width
		p.targetHeight = p.height
	}

	return nil
}

func (p *VideoProcessorImpl) ApplyFilter(filter string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return ErrVideoProcessorImplNotInitialized
	}

	supportedFilters := map[string]bool{
		"grayscale":  true,
		"blur":       true,
		"sharpen":    true,
		"edge":       true,
		"sepia":      true,
		"invert":     true,
		"brightness": true,
		"contrast":   true,
	}

	if !supportedFilters[filter] {
		return ErrUnsupportedFilter
	}

	p.filters = append(p.filters, filter)

	return nil
}

func (p *VideoProcessorImpl) GetWidth() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.targetWidth
}

func (p *VideoProcessorImpl) GetHeight() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.targetHeight
}

func (p *VideoProcessorImpl) SetBrightness(value float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.brightness = value
}

func (p *VideoProcessorImpl) SetContrast(value float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.contrast = value
}

func (p *VideoProcessorImpl) SetSaturation(value float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.saturation = value
}

func (p *VideoProcessorImpl) applyColorAdjustment(data []byte) []byte {
	if len(data) < 3 {
		return data
	}

	frameSize := p.width * p.height
	if len(data) < frameSize*3 {
		return data
	}

	for i := 0; i < frameSize; i++ {
		offset := i * 3

		r := float64(data[offset])
		g := float64(data[offset+1])
		b := float64(data[offset+2])

		r += p.brightness * 255
		g += p.brightness * 255
		b += p.brightness * 255

		r = (r-128)*p.contrast + 128
		g = (g-128)*p.contrast + 128
		b = (b-128)*p.contrast + 128

		gray := 0.299*r + 0.587*g + 0.114*b
		r = gray + p.saturation*(r-gray)
		g = gray + p.saturation*(g-gray)
		b = gray + p.saturation*(b-gray)

		if r < 0 {
			r = 0
		} else if r > 255 {
			r = 255
		}
		if g < 0 {
			g = 0
		} else if g > 255 {
			g = 255
		}
		if b < 0 {
			b = 0
		} else if b > 255 {
			b = 255
		}

		data[offset] = uint8(r)
		data[offset+1] = uint8(g)
		data[offset+2] = uint8(b)
	}

	return data
}

func (p *VideoProcessorImpl) Reset() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.targetWidth = p.width
	p.targetHeight = p.height
	p.filters = nil
	p.rotation = 0
	p.flipH = false
	p.flipV = false
	p.brightness = 0
	p.contrast = 1
	p.saturation = 1

	return nil
}

func (p *VideoProcessorImpl) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.initialized = false
	p.filters = nil

	return nil
}

type VideoCapturer interface {
	Start() error
	Stop() error
	GetFrame() ([]byte, error)
	SetResolution(width, height int) error
	SetFrameRate(fps int) error
	GetSupportedResolutions() []image.Point
}

type VideoRenderer interface {
	Initialize() error
	RenderFrame(data []byte, width, height int) error
	SetVolume(volume int) error
	GetCapabilities() *RendererCapabilities
	Close() error
}

type RendererCapabilities struct {
	MaxWidth        int
	MaxHeight       int
	MaxFrameRate    int
	SupportedCodecs []string
	HardwareAccel   bool
}

type SimpleVideoCapturer struct {
	mu                   sync.RWMutex
	width                int
	height               int
	frameRate            int
	running              bool
	frameCount           uint64
	supportedResolutions []image.Point
}

func NewSimpleVideoCapturer(width, height, frameRate int) (*SimpleVideoCapturer, error) {
	if width <= 0 || height <= 0 || frameRate <= 0 {
		return nil, ErrInvalidDimensions
	}

	capturer := &SimpleVideoCapturer{
		width:     width,
		height:    height,
		frameRate: frameRate,
		running:   false,
		supportedResolutions: []image.Point{
			{640, 480},
			{1280, 720},
			{1920, 1080},
		},
	}

	return capturer, nil
}

func (c *SimpleVideoCapturer) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return nil
	}

	c.running = true
	return nil
}

func (c *SimpleVideoCapturer) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.running = false
	return nil
}

func (c *SimpleVideoCapturer) GetFrame() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.running {
		return nil, errors.New("capturer not running")
	}

	frameSize := c.width * c.height * 3 / 2
	frame := make([]byte, frameSize)

	for i := range frame {
		frame[i] = byte(i % 256)
	}

	c.frameCount++

	return frame, nil
}

func (c *SimpleVideoCapturer) SetResolution(width, height int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if width <= 0 || height <= 0 {
		return ErrInvalidDimensions
	}

	c.width = width
	c.height = height

	return nil
}

func (c *SimpleVideoCapturer) SetFrameRate(fps int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if fps <= 0 {
		return errors.New("invalid frame rate")
	}

	c.frameRate = fps
	return nil
}

func (c *SimpleVideoCapturer) GetSupportedResolutions() []image.Point {
	c.mu.RLock()
	defer c.mu.RUnlock()

	resolutions := make([]image.Point, len(c.supportedResolutions))
	copy(resolutions, c.supportedResolutions)

	return resolutions
}

type SimpleVideoRenderer struct {
	mu           sync.RWMutex
	width        int
	height       int
	volume       int
	initialized  bool
	frameCount   uint64
	capabilities *RendererCapabilities
}

func NewSimpleVideoRenderer() *SimpleVideoRenderer {
	return &SimpleVideoRenderer{
		width:       640,
		height:      480,
		volume:      100,
		initialized: false,
		capabilities: &RendererCapabilities{
			MaxWidth:        3840,
			MaxHeight:       2160,
			MaxFrameRate:    60,
			SupportedCodecs: []string{"vp8", "vp9", "h264"},
			HardwareAccel:   false,
		},
	}
}

func (r *SimpleVideoRenderer) Initialize() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.initialized = true
	return nil
}

func (r *SimpleVideoRenderer) RenderFrame(data []byte, width, height int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.initialized {
		return errors.New("renderer not initialized")
	}

	if len(data) == 0 {
		return ErrInvalidInput
	}

	r.width = width
	r.height = height
	r.frameCount++

	return nil
}

func (r *SimpleVideoRenderer) SetVolume(volume int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if volume < 0 || volume > 100 {
		return errors.New("volume must be between 0 and 100")
	}

	r.volume = volume
	return nil
}

func (r *SimpleVideoRenderer) GetCapabilities() *RendererCapabilities {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.capabilities
}

func (r *SimpleVideoRenderer) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.initialized = false
	return nil
}

type VideoDevice struct {
	ID        string
	Name      string
	Index     int
	Width     int
	Height    int
	FrameRate int
	IsDefault bool
}

type VideoDeviceManager struct {
	mu       sync.RWMutex
	devices  []*VideoDevice
	selected *VideoDevice
}

func NewVideoDeviceManager() *VideoDeviceManager {
	return &VideoDeviceManager{
		devices: []*VideoDevice{
			{
				ID:        "default-video",
				Name:      "Default Camera",
				Index:     0,
				Width:     640,
				Height:    480,
				FrameRate: 30,
				IsDefault: true,
			},
		},
	}
}

func (m *VideoDeviceManager) GetDevices() []*VideoDevice {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]*VideoDevice, len(m.devices))
	copy(devices, m.devices)

	return devices
}

func (m *VideoDeviceManager) SelectDevice(deviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, device := range m.devices {
		if device.ID == deviceID {
			m.selected = device
			return nil
		}
	}

	return errors.New("device not found")
}

func (m *VideoDeviceManager) GetSelectedDevice() *VideoDevice {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.selected
}

func (m *VideoDeviceManager) GetDefaultDevice() *VideoDevice {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, device := range m.devices {
		if device.IsDefault {
			return device
		}
	}

	if len(m.devices) > 0 {
		return m.devices[0]
	}

	return nil
}
