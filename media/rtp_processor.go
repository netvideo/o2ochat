package media

import (
	"encoding/binary"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

const (
	RTPVersion          = 2
	MaxRTPPacketSize    = 1500
	MTUSize             = 1500
	RTPHeaderSize       = 12
	MaxCSRC             = 15
	MaxExtensionHeaders = 32
)

var (
	ErrInvalidPacket      = errors.New("invalid RTP packet")
	ErrPacketTooLarge     = errors.New("packet too large")
	ErrInvalidPayloadType = errors.New("invalid payload type")
	ErrBufferTooSmall     = errors.New("buffer too small")
)

func boolToUint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

func uint8ToBool(v uint8) bool {
	return v != 0
}

type RTPProcessorImpl struct {
	mu             sync.RWMutex
	ssrc           uint32
	sequenceNumber uint16
	timestamp      uint32
	timestampDelta uint32
	marker         bool
	payloadType    uint8
	clockRate      int
	maxPacketSize  int
	initialized    bool

	packetsSent     uint64
	packetsReceived uint64
	packetsLost     uint64
	retransmits     uint64

	nackHistory     map[uint16]time.Time
	retransmitQueue chan *RTPPacket

	statsLock       sync.RWMutex
	lastReceiveTime time.Time
	lastSendTime    time.Time
	jitter          uint32
	prevTimestamp   uint32
	prevReceiveTime time.Time
}

func NewRTPProcessor(ssrc uint32, payloadType uint8, clockRate int) (*RTPProcessorImpl, error) {
	if payloadType > 127 {
		return nil, ErrInvalidPayloadType
	}

	processor := &RTPProcessorImpl{
		ssrc:            ssrc,
		sequenceNumber:  uint16(time.Now().UnixNano() & 0xFFFF),
		timestamp:       uint32(time.Now().UnixNano() / 1000),
		timestampDelta:  uint32(clockRate / 1000 * 20),
		payloadType:     payloadType,
		clockRate:       clockRate,
		maxPacketSize:   MaxRTPPacketSize,
		initialized:     true,
		nackHistory:     make(map[uint16]time.Time),
		retransmitQueue: make(chan *RTPPacket, 100),
	}

	return processor, nil
}

func (p *RTPProcessorImpl) Packetize(frame *MediaFrame) ([]*RTPPacket, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return nil, errors.New("processor not initialized")
	}

	if frame == nil || len(frame.Payload) == 0 {
		return nil, ErrInvalidPacket
	}

	payloadSize := p.maxPacketSize - RTPHeaderSize
	if payloadSize <= 0 {
		return nil, ErrPacketTooLarge
	}

	var packets []*RTPPacket
	offset := 0

	for offset < len(frame.Payload) {
		remaining := len(frame.Payload) - offset
		chunkSize := payloadSize
		if remaining < chunkSize {
			chunkSize = remaining
		}

		marker := false
		if offset+chunkSize >= len(frame.Payload) {
			marker = true
		}

		packet := &RTPPacket{
			Version:     RTPVersion,
			Marker:      marker,
			PayloadType: p.payloadType,
			Sequence:    p.sequenceNumber,
			Timestamp:   frame.Timestamp,
			SSRC:        p.ssrc,
			Payload:     frame.Payload[offset : offset+chunkSize],
		}

		packets = append(packets, packet)

		offset += chunkSize
		p.sequenceNumber++

		if frame.Type == MediaTypeVideo && frame.KeyFrame && len(packets) == 1 {
			packet.Marker = true
		}
	}

	p.statsLock.Lock()
	p.packetsSent += uint64(len(packets))
	p.lastSendTime = time.Now()
	p.statsLock.Unlock()

	return packets, nil
}

func (p *RTPProcessorImpl) Depacketize(packet *RTPPacket) (*MediaFrame, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return nil, errors.New("processor not initialized")
	}

	if packet == nil {
		return nil, ErrInvalidPacket
	}

	if packet.Version != RTPVersion {
		return nil, ErrInvalidPacket
	}

	p.statsLock.Lock()
	p.packetsReceived++

	now := time.Now()
	if p.prevReceiveTime.IsZero() == false {
		delta := now.Sub(p.prevReceiveTime)
		if delta > 0 && delta < time.Second {
			tsDelta := int32(packet.Timestamp - p.prevTimestamp)
			if tsDelta < 0 {
				tsDelta = -tsDelta
			}
			jitter := uint32(delta.Seconds() * float64(p.clockRate))
			if jitter > p.jitter {
				p.jitter = (p.jitter*7 + jitter) / 8
			}
		}
	}
	p.prevTimestamp = packet.Timestamp
	p.prevReceiveTime = now
	p.lastReceiveTime = now
	p.statsLock.Unlock()

	mediaType := MediaTypeAudio
	if packet.PayloadType >= 96 && packet.PayloadType <= 127 {
		mediaType = MediaTypeVideo
	}

	frame := &MediaFrame{
		Type:      mediaType,
		Timestamp: packet.Timestamp,
		Sequence:  packet.Sequence,
		Payload:   packet.Payload,
		Size:      len(packet.Payload),
		KeyFrame:  p.isKeyFramePacket(packet),
		Duration:  time.Duration(p.timestampDelta) * time.Microsecond,
	}

	return frame, nil
}

func (p *RTPProcessorImpl) isKeyFramePacket(packet *RTPPacket) bool {
	if packet.PayloadType < 96 {
		return false
	}

	if len(packet.Payload) < 1 {
		return false
	}

	return (packet.Payload[0] & 0x01) != 0
}

func (p *RTPProcessorImpl) HandleNACK(seqNums []uint16) ([]*RTPPacket, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return nil, errors.New("processor not initialized")
	}

	var packets []*RTPPacket

	p.statsLock.Lock()
	now := time.Now()

	for _, seq := range seqNums {
		if nackTime, exists := p.nackHistory[seq]; exists {
			if now.Sub(nackTime) < time.Second {
				p.retransmits++
			}
		}
		p.nackHistory[seq] = now
	}

	p.statsLock.Unlock()

	return packets, nil
}

func (p *RTPProcessorImpl) HandlePLI() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return errors.New("processor not initialized")
	}

	return nil
}

func (p *RTPProcessorImpl) HandleFIR() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return errors.New("processor not initialized")
	}

	return nil
}

func (p *RTPProcessorImpl) GetRTPStats() *RTPStats {
	p.statsLock.RLock()
	defer p.statsLock.RUnlock()

	stats := &RTPStats{
		PacketsSent:     atomic.LoadUint64(&p.packetsSent),
		PacketsReceived: atomic.LoadUint64(&p.packetsReceived),
		PacketsLost:     atomic.LoadUint64(&p.packetsLost),
		Jitter:          p.jitter,
	}

	if !p.lastReceiveTime.IsZero() && !p.lastSendTime.IsZero() {
		stats.RoundTripTime = p.lastReceiveTime.Sub(p.lastSendTime)
		if stats.RoundTripTime < 0 {
			stats.RoundTripTime = -stats.RoundTripTime
		}
	}

	return stats
}

func (p *RTPProcessorImpl) SetMaxPacketSize(size int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if size < RTPHeaderSize {
		return errors.New("packet size too small")
	}

	p.maxPacketSize = size
	return nil
}

func (p *RTPProcessorImpl) SetPayloadType(payloadType uint8) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if payloadType > 127 {
		return ErrInvalidPayloadType
	}

	p.payloadType = payloadType
	return nil
}

func (p *RTPProcessorImpl) SetClockRate(clockRate int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if clockRate <= 0 {
		return errors.New("invalid clock rate")
	}

	p.clockRate = clockRate
	p.timestampDelta = uint32(clockRate / 1000 * 20)
	return nil
}

func (p *RTPProcessorImpl) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.initialized = false
	close(p.retransmitQueue)

	return nil
}

func (p *RTPProcessorImpl) SerializePacket(packet *RTPPacket) ([]byte, error) {
	if packet == nil {
		return nil, ErrInvalidPacket
	}

	size := RTPHeaderSize + len(packet.CSRC)*4
	if packet.Extension {
		size += 4 + len(packet.ExtensionHeader)
	}
	size += len(packet.Payload)

	buf := make([]byte, size)

	buf[0] = (packet.Version << 6) | (boolToUint8(packet.Padding) << 5) | (boolToUint8(packet.Extension) << 4) | packet.CSRCCount
	buf[1] = (boolToUint8(packet.Marker) << 7) | packet.PayloadType

	binary.BigEndian.PutUint16(buf[2:], packet.Sequence)
	binary.BigEndian.PutUint32(buf[4:], packet.Timestamp)
	binary.BigEndian.PutUint32(buf[8:], packet.SSRC)

	offset := RTPHeaderSize
	for _, csrc := range packet.CSRC {
		binary.BigEndian.PutUint32(buf[offset:], csrc)
		offset += 4
	}

	if packet.Extension {
		binary.BigEndian.PutUint16(buf[offset:], 1)
		binary.BigEndian.PutUint16(buf[offset+2:], uint16(len(packet.ExtensionHeader)))
		copy(buf[offset+4:], packet.ExtensionHeader)
		offset += 4 + len(packet.ExtensionHeader)
	}

	copy(buf[offset:], packet.Payload)

	return buf, nil
}

func (p *RTPProcessorImpl) DeserializePacket(data []byte) (*RTPPacket, error) {
	if len(data) < RTPHeaderSize {
		return nil, ErrInvalidPacket
	}

	packet := &RTPPacket{}

	packet.Version = (data[0] >> 6) & 0x03
	packet.Padding = uint8ToBool((data[0] >> 5) & 0x01)
	packet.Extension = uint8ToBool((data[0] >> 4) & 0x01)
	packet.CSRCCount = data[0] & 0x0F

	packet.Marker = uint8ToBool((data[1] >> 7) & 0x01)
	packet.PayloadType = data[1] & 0x7F

	packet.Sequence = binary.BigEndian.Uint16(data[2:])
	packet.Timestamp = binary.BigEndian.Uint32(data[4:])
	packet.SSRC = binary.BigEndian.Uint32(data[8:])

	offset := RTPHeaderSize

	if int(packet.CSRCCount)*4+offset > len(data) {
		return nil, ErrInvalidPacket
	}

	packet.CSRC = make([]uint32, packet.CSRCCount)
	for i := 0; i < int(packet.CSRCCount); i++ {
		packet.CSRC[i] = binary.BigEndian.Uint32(data[offset:])
		offset += 4
	}

	if packet.Extension {
		if offset+4 > len(data) {
			return nil, ErrInvalidPacket
		}
		_ = binary.BigEndian.Uint16(data[offset:])
		extensionLength := binary.BigEndian.Uint16(data[offset+2:])

		offset += 4
		if offset+int(extensionLength) > len(data) {
			return nil, ErrInvalidPacket
		}
		packet.ExtensionHeader = data[offset : offset+int(extensionLength)]
		offset += int(extensionLength)
	}

	packet.Payload = data[offset:]

	return packet, nil
}

type JitterBufferImpl struct {
	mu          sync.RWMutex
	maxSize     time.Duration
	packets     map[uint16]*RTPPacket
	frames      []*MediaFrame
	minSequence uint16
	maxSequence uint16
	initialized bool
	framesReady int
	closed      bool
	lock        sync.Mutex
}

func NewJitterBufferImpl(maxSize time.Duration) (*JitterBufferImpl, error) {
	if maxSize <= 0 {
		maxSize = 200 * time.Millisecond
	}

	buffer := &JitterBufferImpl{
		maxSize:     maxSize,
		packets:     make(map[uint16]*RTPPacket),
		initialized: true,
		framesReady: 0,
	}

	return buffer, nil
}

func (b *JitterBufferImpl) AddPacket(packet *RTPPacket) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if !b.initialized {
		return errors.New("buffer not initialized")
	}

	if b.closed {
		return errors.New("buffer closed")
	}

	if packet == nil {
		return ErrInvalidPacket
	}

	b.packets[packet.Sequence] = packet

	if b.minSequence == 0 || uint16Diff(packet.Sequence, b.minSequence) < uint16Diff(b.maxSequence, b.minSequence) {
		b.minSequence = packet.Sequence
	}

	if b.maxSequence == 0 || uint16Diff(packet.Sequence, b.maxSequence) > uint16Diff(b.maxSequence, b.minSequence) {
		b.maxSequence = packet.Sequence
	}

	b.updateFramesReady()

	return nil
}

func (b *JitterBufferImpl) GetNextFrame() (*MediaFrame, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if !b.initialized {
		return nil, errors.New("buffer not initialized")
	}

	if len(b.frames) > 0 {
		frame := b.frames[0]
		b.frames = b.frames[1:]
		b.framesReady--
		return frame, nil
	}

	if len(b.packets) == 0 {
		return nil, nil
	}

	if uint16Diff(b.maxSequence, b.minSequence) < 2 {
		return nil, nil
	}

	var frames []*MediaFrame
	seq := b.minSequence

	for {
		packet, exists := b.packets[seq]
		if !exists {
			break
		}

		frame := &MediaFrame{
			Timestamp: packet.Timestamp,
			Sequence:  packet.Sequence,
			Payload:   packet.Payload,
			Size:      len(packet.Payload),
		}
		frames = append(frames, frame)

		delete(b.packets, seq)
		seq++
	}

	b.minSequence = seq

	if len(frames) > 0 {
		firstFrame := frames[0]
		return firstFrame, nil
	}

	return nil, nil
}

func (b *JitterBufferImpl) SetBufferSize(size time.Duration) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if size <= 0 {
		return errors.New("invalid buffer size")
	}

	b.maxSize = size
	return nil
}

func (b *JitterBufferImpl) GetBufferStatus() *BufferStatus {
	b.lock.Lock()
	defer b.lock.Unlock()

	return &BufferStatus{
		Size:        len(b.packets),
		MaxSize:     int(b.maxSize.Milliseconds()),
		PacketCount: len(b.packets),
		FramesReady: b.framesReady,
	}
}

func (b *JitterBufferImpl) Reset() error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.packets = make(map[uint16]*RTPPacket)
	b.frames = nil
	b.minSequence = 0
	b.maxSequence = 0
	b.framesReady = 0

	return nil
}

func (b *JitterBufferImpl) Close() error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.initialized = false
	b.closed = true
	b.packets = nil
	b.frames = nil

	return nil
}

func (b *JitterBufferImpl) updateFramesReady() {
	packetsCount := len(b.packets)
	if packetsCount >= 3 {
		b.framesReady = packetsCount - 2
	} else {
		b.framesReady = 0
	}
}

func uint16Diff(a, b uint16) uint16 {
	if a >= b {
		return a - b
	}
	return b - a
}

type RTPProcessorFactory struct{}

func NewRTPProcessorFactory() *RTPProcessorFactory {
	return &RTPProcessorFactory{}
}

func (f *RTPProcessorFactory) CreateAudioProcessor(ssrc uint32) (*RTPProcessorImpl, error) {
	return NewRTPProcessor(ssrc, 96, 48000)
}

func (f *RTPProcessorFactory) CreateVideoProcessor(ssrc uint32) (*RTPProcessorImpl, error) {
	return NewRTPProcessor(ssrc, 100, 90000)
}

func (f *RTPProcessorFactory) CreateJitterBufferImpl() (*JitterBufferImpl, error) {
	return NewJitterBufferImpl(200 * time.Millisecond)
}
