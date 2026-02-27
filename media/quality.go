package media

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrBandwidthEstimateFailed = errors.New("bandwidth estimation failed")
	ErrQualityControlFailed    = errors.New("quality control failed")
)

type BandwidthEstimator struct {
	mu              sync.RWMutex
	samples         []BandwidthSample
	maxSamples      int
	windowSize      time.Duration
	lastUpdate      time.Time
	currentEstimate int64
	minBitrate      int64
	maxBitrate      int64
}

type BandwidthSample struct {
	Timestamp  time.Duration
	Bytes      int64
	Duration   time.Duration
	PacketLoss float64
	Jitter     time.Duration
}

func NewBandwidthEstimator(minBitrate, maxBitrate int) *BandwidthEstimator {
	return &BandwidthEstimator{
		samples:         make([]BandwidthSample, 0),
		maxSamples:      100,
		windowSize:      5 * time.Second,
		lastUpdate:      time.Now(),
		currentEstimate: int64(maxBitrate),
		minBitrate:      int64(minBitrate),
		maxBitrate:      int64(maxBitrate),
	}
}

func (e *BandwidthEstimator) AddSample(bytes int64, duration time.Duration, packetLoss float64, jitter time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	sample := BandwidthSample{
		Timestamp:  time.Since(e.lastUpdate),
		Bytes:      bytes,
		Duration:   duration,
		PacketLoss: packetLoss,
		Jitter:     jitter,
	}

	e.samples = append(e.samples, sample)

	if len(e.samples) > e.maxSamples {
		e.samples = e.samples[len(e.samples)-e.maxSamples:]
	}

	e.recalculateEstimate()
}

func (e *BandwidthEstimator) recalculateEstimate() {
	if len(e.samples) < 2 {
		return
	}

	var totalBytes int64
	var totalDuration time.Duration

	for _, sample := range e.samples {
		if time.Since(e.lastUpdate)-sample.Timestamp < e.windowSize {
			totalBytes += sample.Bytes
			totalDuration += sample.Duration
		}
	}

	if totalDuration > 0 {
		bitsPerSecond := float64(totalBytes*8) / totalDuration.Seconds()

		packetLossFactor := 1.0
		for _, sample := range e.samples {
			if sample.PacketLoss > 0.1 {
				packetLossFactor *= 0.5
			} else if sample.PacketLoss > 0.05 {
				packetLossFactor *= 0.8
			}
		}

		adjustedBitrate := int64(bitsPerSecond * packetLossFactor)

		if adjustedBitrate < e.minBitrate {
			adjustedBitrate = e.minBitrate
		} else if adjustedBitrate > e.maxBitrate {
			adjustedBitrate = e.maxBitrate
		}

		e.currentEstimate = (e.currentEstimate*3 + adjustedBitrate) / 4
	}
}

func (e *BandwidthEstimator) GetEstimate() int64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.currentEstimate
}

func (e *BandwidthEstimator) GetTargetBitrate(mediaType MediaType) int64 {
	e.mu.RLock()
	estimate := e.currentEstimate
	e.mu.RUnlock()

	base := estimate

	if mediaType == MediaTypeVideo {
		audioBitrate := int64(64000)
		base -= audioBitrate
		if base < 0 {
			base = 0
		}
	}

	lossThreshold := 0.03
	jitterThreshold := 50 * time.Millisecond

	e.mu.RLock()
	for _, sample := range e.samples {
		if sample.PacketLoss > lossThreshold {
			base = base * 80 / 100
		}
		if sample.Jitter > jitterThreshold {
			base = base * 90 / 100
		}
	}
	e.mu.RUnlock()

	return base
}

func (e *BandwidthEstimator) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.samples = make([]BandwidthSample, 0)
	e.currentEstimate = e.maxBitrate
}

type QualityController struct {
	mu                 sync.RWMutex
	currentQuality     QualityLevel
	targetQuality      QualityLevel
	minQuality         QualityLevel
	maxQuality         QualityLevel
	qualityLevels      []QualityLevel
	adaptationMode     AdaptationMode
	lastAdjustment     time.Time
	adjustmentInterval time.Duration
}

type QualityLevel int

const (
	QualityLow QualityLevel = iota
	QualityMedium
	QualityHigh
	QualityUltra
)

type AdaptationMode int

const (
	AdaptationModeManual AdaptationMode = iota
	AdaptationModeAuto
	AdaptationModeConservative
)

func (q QualityLevel) String() string {
	switch q {
	case QualityLow:
		return "low"
	case QualityMedium:
		return "medium"
	case QualityHigh:
		return "high"
	case QualityUltra:
		return "ultra"
	default:
		return "unknown"
	}
}

func NewQualityController() *QualityController {
	return &QualityController{
		currentQuality:     QualityHigh,
		targetQuality:      QualityHigh,
		minQuality:         QualityLow,
		maxQuality:         QualityUltra,
		qualityLevels:      []QualityLevel{QualityLow, QualityMedium, QualityHigh, QualityUltra},
		adaptationMode:     AdaptationModeAuto,
		lastAdjustment:     time.Now(),
		adjustmentInterval: 2 * time.Second,
	}
}

func (c *QualityController) GetQualityForBitrate(bitrate int64) QualityLevel {
	if bitrate < 200000 {
		return QualityLow
	} else if bitrate < 500000 {
		return QualityMedium
	} else if bitrate < 1500000 {
		return QualityHigh
	}
	return QualityUltra
}

func (c *QualityController) GetResolutionForQuality(quality QualityLevel) (width, height int) {
	switch quality {
	case QualityLow:
		return 320, 240
	case QualityMedium:
		return 640, 480
	case QualityHigh:
		return 1280, 720
	case QualityUltra:
		return 1920, 1080
	default:
		return 640, 480
	}
}

func (c *QualityController) GetBitrateForQuality(quality QualityLevel, mediaType MediaType) int {
	if mediaType == MediaTypeAudio {
		return 64000
	}

	switch quality {
	case QualityLow:
		return 200000
	case QualityMedium:
		return 500000
	case QualityHigh:
		return 1500000
	case QualityUltra:
		return 3000000
	default:
		return 500000
	}
}

func (c *QualityController) AdjustQuality(packetLoss, jitter float64, rtt time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Since(c.lastAdjustment) < c.adjustmentInterval {
		return false
	}

	if c.adaptationMode == AdaptationModeManual {
		return false
	}

	newQuality := c.currentQuality

	if packetLoss > 0.1 || rtt > 500*time.Millisecond {
		newQuality--
	} else if packetLoss < 0.01 && rtt < 100*time.Millisecond && jitter < 30 {
		newQuality++
	}

	if newQuality < c.minQuality {
		newQuality = c.minQuality
	}
	if newQuality > c.maxQuality {
		newQuality = c.maxQuality
	}

	if newQuality != c.currentQuality {
		c.currentQuality = newQuality
		c.lastAdjustment = time.Now()
		return true
	}

	return false
}

func (c *QualityController) GetCurrentQuality() QualityLevel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentQuality
}

func (c *QualityController) SetTargetQuality(quality QualityLevel) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if quality >= c.minQuality && quality <= c.maxQuality {
		c.targetQuality = quality
	}
}

func (c *QualityController) SetAdaptationMode(mode AdaptationMode) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.adaptationMode = mode
}

func (c *QualityController) GetCodecConfig() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	quality := c.currentQuality
	width, height := c.GetResolutionForQuality(quality)
	videoBitrate := c.GetBitrateForQuality(quality, MediaTypeVideo)

	return map[string]interface{}{
		"width":             width,
		"height":            height,
		"video_bitrate":     videoBitrate,
		"audio_bitrate":     64000,
		"frame_rate":        c.getFrameRateForQuality(quality),
		"keyframe_interval": c.getKeyFrameInterval(quality),
	}
}

func (c *QualityController) getFrameRateForQuality(quality QualityLevel) int {
	switch quality {
	case QualityLow:
		return 15
	case QualityMedium:
		return 24
	case QualityHigh:
		return 30
	case QualityUltra:
		return 60
	default:
		return 30
	}
}

func (c *QualityController) getKeyFrameInterval(quality QualityLevel) int {
	switch quality {
	case QualityLow:
		return 5000
	case QualityMedium:
		return 3000
	case QualityHigh:
		return 2000
	case QualityUltra:
		return 1000
	default:
		return 2000
	}
}

type NetworkAdaptor struct {
	mu                 sync.RWMutex
	bandwidthEstimator *BandwidthEstimator
	qualityController  *QualityController
	session            CallSession
	active             bool
	stopChan           chan struct{}
	wg                 sync.WaitGroup
	statsCollector     *StatsCollector
}

func NewNetworkAdaptor(session CallSession, minBitrate, maxBitrate int) *NetworkAdaptor {
	return &NetworkAdaptor{
		bandwidthEstimator: NewBandwidthEstimator(minBitrate, maxBitrate),
		qualityController:  NewQualityController(),
		session:            session,
		active:             false,
		stopChan:           make(chan struct{}),
		statsCollector:     NewStatsCollector(),
	}
}

func (a *NetworkAdaptor) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.active {
		return nil
	}

	a.active = true
	a.stopChan = make(chan struct{})

	a.wg.Add(1)
	go a.adaptLoop()

	return nil
}

func (a *NetworkAdaptor) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.active {
		return nil
	}

	close(a.stopChan)
	a.wg.Wait()
	a.active = false

	return nil
}

func (a *NetworkAdaptor) adaptLoop() {
	defer a.wg.Done()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopChan:
			return
		case <-ticker.C:
			a.performAdaptation()
		}
	}
}

func (a *NetworkAdaptor) performAdaptation() {
	stats := a.session.GetStats()
	if stats == nil {
		return
	}

	networkStats := stats.NetworkStats
	if networkStats == nil {
		return
	}

	a.bandwidthEstimator.AddSample(
		networkStats.Bandwidth/8,
		time.Second,
		networkStats.PacketLoss,
		networkStats.Jitter,
	)

	bitrate := a.bandwidthEstimator.GetEstimate()

	adjusted := a.qualityController.AdjustQuality(
		networkStats.PacketLoss,
		float64(networkStats.Jitter),
		networkStats.Latency,
	)

	if adjusted {
		codecConfig := a.qualityController.GetCodecConfig()

		if videoBitrate, ok := codecConfig["video_bitrate"].(int); ok {
			a.session.AdjustBitrate(videoBitrate)
		}
	}

	_ = bitrate
}

func (a *NetworkAdaptor) GetBandwidthEstimate() int64 {
	return a.bandwidthEstimator.GetEstimate()
}

func (a *NetworkAdaptor) GetCurrentQuality() QualityLevel {
	return a.qualityController.GetCurrentQuality()
}

type StatsCollector struct {
	mu           sync.RWMutex
	audioStats   StreamStatsData
	videoStats   StreamStatsData
	networkStats NetworkStatsData
	startTime    time.Time
	samples      []StatsSample
	maxSamples   int
}

type StreamStatsData struct {
	BytesSent       int64
	BytesReceived   int64
	PacketsSent     int64
	PacketsReceived int64
	PacketsLost     int64
	FrameCount      int64
	Bitrate         int
}

type NetworkStatsData struct {
	Bandwidth   int64
	PacketLoss  float64
	Jitter      time.Duration
	Latency     time.Duration
	Retransmits int
}

type StatsSample struct {
	Timestamp    time.Time
	AudioStats   StreamStatsData
	VideoStats   StreamStatsData
	NetworkStats NetworkStatsData
}

func NewStatsCollector() *StatsCollector {
	return &StatsCollector{
		startTime:  time.Now(),
		samples:    make([]StatsSample, 0),
		maxSamples: 1000,
	}
}

func (c *StatsCollector) RecordAudioStats(stats StreamStatsData) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.audioStats = stats
}

func (c *StatsCollector) RecordVideoStats(stats StreamStatsData) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.videoStats = stats
}

func (c *StatsCollector) RecordNetworkStats(stats NetworkStatsData) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.networkStats = stats

	sample := StatsSample{
		Timestamp:    time.Now(),
		AudioStats:   c.audioStats,
		VideoStats:   c.videoStats,
		NetworkStats: c.networkStats,
	}

	c.samples = append(c.samples, sample)
	if len(c.samples) > c.maxSamples {
		c.samples = c.samples[len(c.samples)-c.maxSamples:]
	}
}

func (c *StatsCollector) GetAudioStats() StreamStatsData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.audioStats
}

func (c *StatsCollector) GetVideoStats() StreamStatsData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.videoStats
}

func (c *StatsCollector) GetNetworkStats() NetworkStatsData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.networkStats
}

func (c *StatsCollector) GetAverageStats(duration time.Duration) (audio, video StreamStatsData, network NetworkStatsData) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.samples) == 0 {
		return c.audioStats, c.videoStats, c.networkStats
	}

	cutoff := time.Now().Add(-duration)
	var audioCount, videoCount, networkCount int

	for _, sample := range c.samples {
		if sample.Timestamp.After(cutoff) {
			audioCount++
			videoCount++
			networkCount++
		}
	}

	if audioCount > 0 {
		audio = c.audioStats
		video = c.videoStats
		network = c.networkStats
	}

	return
}

func (c *StatsCollector) GetQualityScore() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.samples) == 0 {
		return 0.5
	}

	var totalScore float64
	var count int

	for _, sample := range c.samples {
		lossPenalty := 1.0 - sample.NetworkStats.PacketLoss
		jitterPenalty := 1.0 - float64(sample.NetworkStats.Jitter.Milliseconds())/500.0
		latencyPenalty := 1.0 - float64(sample.NetworkStats.Latency.Milliseconds())/1000.0

		if jitterPenalty < 0 {
			jitterPenalty = 0
		}
		if latencyPenalty < 0 {
			latencyPenalty = 0
		}

		score := (lossPenalty + jitterPenalty + latencyPenalty) / 3.0
		totalScore += score
		count++
	}

	if count == 0 {
		return 0.5
	}

	return totalScore / float64(count)
}

func (c *StatsCollector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.samples = make([]StatsSample, 0)
	c.startTime = time.Now()
}

func CalculateMOS(packetLoss float64, latency, jitter time.Duration) float64 {
	if packetLoss > 0.5 {
		packetLoss = 0.5
	}

	R := 93.2 - (float64(latency.Milliseconds()) / 40.0)
	R -= (float64(jitter.Milliseconds()) / 10.0)
	R -= (packetLoss * 100 * 2.5)

	if R < 0 {
		R = 0
	}
	if R > 100 {
		R = 100
	}

	mos := 1.0 + 0.035*R + 0.000007*R*(R-60)*(100-R)

	if mos > 4.5 {
		mos = 4.5
	}
	if mos < 1.0 {
		mos = 1.0
	}

	return mos
}

type CallQualityMonitor struct {
	mu                sync.RWMutex
	sessionID         string
	statsCollector    *StatsCollector
	qualityController *QualityController
	callback          QualityCallback
	stopChan          chan struct{}
	wg                sync.WaitGroup
	active            bool
}

type QualityCallback func(quality QualityInfo)

type QualityInfo struct {
	SessionID  string
	Quality    QualityLevel
	MOS        float64
	PacketLoss float64
	Jitter     time.Duration
	Latency    time.Duration
	Bandwidth  int64
	Bitrate    int
	FrameRate  int
	Resolution string
	Timestamp  time.Time
}

func NewCallQualityMonitor(sessionID string) *CallQualityMonitor {
	return &CallQualityMonitor{
		sessionID:         sessionID,
		statsCollector:    NewStatsCollector(),
		qualityController: NewQualityController(),
		stopChan:          make(chan struct{}),
	}
}

func (m *CallQualityMonitor) SetCallback(callback QualityCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callback = callback
}

func (m *CallQualityMonitor) Start() {
	m.mu.Lock()
	if m.active {
		m.mu.Unlock()
		return
	}
	m.active = true
	m.stopChan = make(chan struct{})
	m.mu.Unlock()

	m.wg.Add(1)
	go m.monitorLoop()
}

func (m *CallQualityMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.active {
		return
	}

	close(m.stopChan)
	m.wg.Wait()
	m.active = false
}

func (m *CallQualityMonitor) monitorLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.checkQuality()
		}
	}
}

func (m *CallQualityMonitor) checkQuality() {
	networkStats := m.statsCollector.GetNetworkStats()

	mos := CalculateMOS(networkStats.PacketLoss, networkStats.Latency, networkStats.Jitter)

	quality := m.qualityController.GetQualityForBitrate(networkStats.Bandwidth)
	m.qualityController.AdjustQuality(networkStats.PacketLoss, float64(networkStats.Jitter), networkStats.Latency)

	width, height := m.qualityController.GetResolutionForQuality(quality)

	info := QualityInfo{
		SessionID:  m.sessionID,
		Quality:    quality,
		MOS:        mos,
		PacketLoss: networkStats.PacketLoss,
		Jitter:     networkStats.Jitter,
		Latency:    networkStats.Latency,
		Bandwidth:  networkStats.Bandwidth,
		Bitrate:    m.statsCollector.GetVideoStats().Bitrate,
		FrameRate:  m.qualityController.getFrameRateForQuality(quality),
		Resolution: fmt.Sprintf("%dx%d", width, height),
		Timestamp:  time.Now(),
	}

	m.mu.RLock()
	callback := m.callback
	m.mu.RUnlock()

	if callback != nil {
		callback(info)
	}
}

var mathify = &mathFormatter{}

type mathFormatter struct{}

func (m *mathFormatter) Printf(format string, args ...interface{}) string {
	if len(args) >= 2 {
		width, _ := args[0].(int)
		height, _ := args[1].(int)
		if width > 0 && height > 0 {
			return mathify.sprintf("%dx%d", width, height)
		}
	}
	return ""
}

func (m *mathFormatter) sprintf(format string, args ...interface{}) string {
	s := format
	for i, arg := range args {
		placeholder := "%" + string(rune('1'+i))
		s = replacePlaceholder(s, placeholder, arg)
	}
	return s
}

func replacePlaceholder(s string, placeholder string, arg interface{}) string {
	result := s
	switch v := arg.(type) {
	case int:
		result = replaceInt(result, placeholder, v)
	case string:
		result = replaceString(result, placeholder, v)
	}
	return result
}

func replaceInt(s, placeholder string, value int) string {
	for i := 0; i < len(s)-len(placeholder)+1; i++ {
		if s[i:i+len(placeholder)] == placeholder {
			return s[:i] + string(rune('0'+value/1000%10)) + string(rune('0'+value/100%10)) + string(rune('0'+value/10%10)) + string(rune('0'+value%10)) + s[i+len(placeholder):]
		}
	}
	return s
}

func replaceString(s, placeholder, value string) string {
	for i := 0; i < len(s)-len(placeholder)+1; i++ {
		if s[i:i+len(placeholder)] == placeholder {
			return s[:i] + value + s[i+len(placeholder):]
		}
	}
	return s
}

func (m *CallQualityMonitor) GetCurrentQuality() QualityInfo {
	networkStats := m.statsCollector.GetNetworkStats()
	mos := CalculateMOS(networkStats.PacketLoss, networkStats.Latency, networkStats.Jitter)
	quality := m.qualityController.GetCurrentQuality()
	width, height := m.qualityController.GetResolutionForQuality(quality)

	return QualityInfo{
		SessionID:  m.sessionID,
		Quality:    quality,
		MOS:        mos,
		PacketLoss: networkStats.PacketLoss,
		Jitter:     networkStats.Jitter,
		Latency:    networkStats.Latency,
		Bandwidth:  networkStats.Bandwidth,
		Resolution: fmt.Sprintf("%dx%d", width, height),
		Timestamp:  time.Now(),
	}
}

func NewCallQualityMonitorWithSession(session CallSession) *CallQualityMonitor {
	return NewCallQualityMonitor(session.GetSessionID())
}

type BitrateController struct {
	mu             sync.RWMutex
	currentBitrate int
	targetBitrate  int
	minBitrate     int
	maxBitrate     int
	increaseStep   int
	decreaseStep   int
	stableCount    int
	lastChange     time.Time
	changeInterval time.Duration
}

func NewBitrateController(minBitrate, maxBitrate, initialBitrate int) *BitrateController {
	return &BitrateController{
		currentBitrate: initialBitrate,
		targetBitrate:  initialBitrate,
		minBitrate:     minBitrate,
		maxBitrate:     maxBitrate,
		increaseStep:   maxBitrate / 20,
		decreaseStep:   maxBitrate / 10,
		stableCount:    0,
		lastChange:     time.Now(),
		changeInterval: 1 * time.Second,
	}
}

func (c *BitrateController) Adjust(packetLoss float64, rtt time.Duration) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Since(c.lastChange) < c.changeInterval {
		return c.currentBitrate
	}

	if packetLoss > 0.1 {
		c.currentBitrate -= c.decreaseStep
		if c.currentBitrate < c.minBitrate {
			c.currentBitrate = c.minBitrate
		}
		c.stableCount = 0
	} else if packetLoss < 0.01 && rtt < 100*time.Millisecond {
		c.stableCount++
		if c.stableCount >= 5 {
			c.currentBitrate += c.increaseStep
			if c.currentBitrate > c.maxBitrate {
				c.currentBitrate = c.maxBitrate
			}
			if c.currentBitrate > c.targetBitrate {
				c.currentBitrate = c.targetBitrate
			}
		}
	} else {
		c.stableCount = 0
	}

	c.lastChange = time.Now()
	return c.currentBitrate
}

func (c *BitrateController) SetTargetBitrate(bitrate int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if bitrate < c.minBitrate {
		bitrate = c.minBitrate
	}
	if bitrate > c.maxBitrate {
		bitrate = c.maxBitrate
	}

	c.targetBitrate = bitrate

	if c.currentBitrate > c.targetBitrate {
		c.currentBitrate = c.targetBitrate
	}
}

func (c *BitrateController) GetCurrentBitrate() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentBitrate
}

func (c *BitrateController) GetTargetBitrate() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.targetBitrate
}

type PacketLossDetector struct {
	mu            sync.RWMutex
	windowSize    int
	packets       []bool
	currentIndex  int
	totalPackets  int
	lostPackets   int
	threshold     float64
	recoveryCount int
}

func NewPacketLossDetector(threshold float64) *PacketLossDetector {
	return &PacketLossDetector{
		windowSize:    100,
		packets:       make([]bool, 100),
		currentIndex:  0,
		totalPackets:  0,
		lostPackets:   0,
		threshold:     threshold,
		recoveryCount: 0,
	}
}

func (d *PacketLossDetector) RecordPacket(seqNum uint16, received bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.packets[d.currentIndex] = received
	d.currentIndex = (d.currentIndex + 1) % d.windowSize
	d.totalPackets++

	if !received {
		d.lostPackets++
	}
}

func (d *PacketLossDetector) GetPacketLoss() float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.totalPackets == 0 {
		return 0
	}

	return float64(d.lostPackets) / float64(d.totalPackets)
}

func (d *PacketLossDetector) IsRecovering() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.recoveryCount > 3
}

func (d *PacketLossDetector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.packets = make([]bool, d.windowSize)
	d.currentIndex = 0
	d.totalPackets = 0
	d.lostPackets = 0
	d.recoveryCount = 0
}

func atomicLoadFloat64(v *float64) float64 {
	return *v
}

func atomicStoreFloat64(v *float64, val float64) {
	*v = val
}
