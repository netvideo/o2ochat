package transport

import (
	"sync"
	"time"
)

// ConnectionQuality represents connection quality metrics
type ConnectionQuality struct {
	ConnectionID string    `json:"connection_id"`
	PeerID       string    `json:"peer_id"`
	Latency      int64     `json:"latency_ms"`     // Round-trip time in ms
	Bandwidth    float64   `json:"bandwidth_mbps"` // Available bandwidth in Mbps
	PacketLoss   float64   `json:"packet_loss"`    // Packet loss rate (0.0-1.0)
	Jitter       int64     `json:"jitter_ms"`      // Jitter in ms
	LastChecked  time.Time `json:"last_checked"`
	QualityScore float64   `json:"quality_score"` // 0-100
	mu           sync.RWMutex
}

// QualityMonitor monitors connection quality
type QualityMonitor struct {
	qualities       map[string]*ConnectionQuality
	metrics         map[string][]QualityMetric
	maxMetricsCount int
	mu              sync.RWMutex
	thresholds      QualityThresholds
	alerts          []QualityAlert
	mu2             sync.RWMutex
}

// QualityMetric represents a quality metric sample
type QualityMetric struct {
	Timestamp  time.Time `json:"timestamp"`
	Latency    int64     `json:"latency_ms"`
	Bandwidth  float64   `json:"bandwidth_mbps"`
	PacketLoss float64   `json:"packet_loss"`
}

// QualityThresholds defines quality thresholds
type QualityThresholds struct {
	ExcellentLatency     int64   `json:"excellent_latency"`      // < 50ms
	GoodLatency          int64   `json:"good_latency"`           // < 100ms
	AcceptableLatency    int64   `json:"acceptable_latency"`     // < 200ms
	ExcellentBandwidth   float64 `json:"excellent_bandwidth"`    // > 10 Mbps
	GoodBandwidth        float64 `json:"good_bandwidth"`         // > 5 Mbps
	AcceptableBandwidth  float64 `json:"acceptable_bandwidth"`   // > 1 Mbps
	AcceptablePacketLoss float64 `json:"acceptable_packet_loss"` // < 0.01 (1%)
}

// QualityAlert represents a quality alert
type QualityAlert struct {
	ConnectionID string    `json:"connection_id"`
	PeerID       string    `json:"peer_id"`
	AlertType    string    `json:"alert_type"`
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
	Severity     string    `json:"severity"` // "info", "warning", "critical"
}

// DefaultQualityThresholds returns default quality thresholds
func DefaultQualityThresholds() QualityThresholds {
	return QualityThresholds{
		ExcellentLatency:     50,
		GoodLatency:          100,
		AcceptableLatency:    200,
		ExcellentBandwidth:   10.0,
		GoodBandwidth:        5.0,
		AcceptableBandwidth:  1.0,
		AcceptablePacketLoss: 0.01,
	}
}

// NewQualityMonitor creates a new quality monitor
func NewQualityMonitor(thresholds *QualityThresholds) *QualityMonitor {
	if thresholds == nil {
		thresholds = &QualityThresholds{}
		*thresholds = DefaultQualityThresholds()
	}

	return &QualityMonitor{
		qualities:       make(map[string]*ConnectionQuality),
		metrics:         make(map[string][]QualityMetric),
		maxMetricsCount: 100,
		thresholds:      *thresholds,
		alerts:          make([]QualityAlert, 0),
	}
}

// UpdateQuality updates connection quality metrics
func (qm *QualityMonitor) UpdateQuality(connID, peerID string, latency int64, bandwidth float64, packetLoss float64) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	// Get or create quality object
	quality, exists := qm.qualities[connID]
	if !exists {
		quality = &ConnectionQuality{
			ConnectionID: connID,
			PeerID:       peerID,
		}
		qm.qualities[connID] = quality
	}

	quality.mu.Lock()
	quality.Latency = latency
	quality.Bandwidth = bandwidth
	quality.PacketLoss = packetLoss
	quality.LastChecked = time.Now()
	quality.QualityScore = qm.calculateQualityScore(latency, bandwidth, packetLoss)
	quality.mu.Unlock()

	// Record metric
	qm.recordMetric(connID, latency, bandwidth, packetLoss)

	// Check for alerts
	qm.checkAlerts(connID, peerID, latency, bandwidth, packetLoss)
}

// GetQuality gets the quality for a connection
func (qm *QualityMonitor) GetQuality(connID string) *ConnectionQuality {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	return qm.qualities[connID]
}

// GetQualityScore gets the quality score for a connection
func (qm *QualityMonitor) GetQualityScore(connID string) float64 {
	qm.mu.RLock()
	quality, exists := qm.qualities[connID]
	qm.mu.RUnlock()

	if !exists {
		return 0.0
	}

	quality.mu.RLock()
	defer quality.mu.RUnlock()
	return quality.QualityScore
}

// IsHealthy checks if a connection is healthy
func (qm *QualityMonitor) IsHealthy(connID string) bool {
	return qm.GetQualityScore(connID) >= 60.0
}

// GetMetrics gets quality metrics for a connection
func (qm *QualityMonitor) GetMetrics(connID string) []QualityMetric {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	metrics, exists := qm.metrics[connID]
	if !exists {
		return nil
	}

	// Return copy
	result := make([]QualityMetric, len(metrics))
	copy(result, metrics)
	return result
}

// GetAlerts gets quality alerts
func (qm *QualityMonitor) GetAlerts() []QualityAlert {
	qm.mu2.RLock()
	defer qm.mu2.RUnlock()

	result := make([]QualityAlert, len(qm.alerts))
	copy(result, qm.alerts)
	return result
}

// GetStats returns quality monitoring statistics
func (qm *QualityMonitor) GetStats() map[string]interface{} {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	totalConnections := len(qm.qualities)
	healthyConnections := 0
	excellentConnections := 0
	goodConnections := 0
	poorConnections := 0

	for _, quality := range qm.qualities {
		quality.mu.RLock()
		score := quality.QualityScore
		if score >= 80.0 {
			excellentConnections++
			healthyConnections++
		} else if score >= 60.0 {
			goodConnections++
			healthyConnections++
		} else {
			poorConnections++
		}
		quality.mu.RUnlock()
	}

	return map[string]interface{}{
		"total_connections":     totalConnections,
		"healthy_connections":   healthyConnections,
		"excellent_connections": excellentConnections,
		"good_connections":      goodConnections,
		"poor_connections":      poorConnections,
		"health_rate":           float64(healthyConnections) / float64(totalConnections),
	}
}

// ClearAlerts clears all alerts
func (qm *QualityMonitor) ClearAlerts() {
	qm.mu2.Lock()
	defer qm.mu2.Unlock()
	qm.alerts = make([]QualityAlert, 0)
}

// calculateQualityScore calculates overall quality score (0-100)
func (qm *QualityMonitor) calculateQualityScore(latency int64, bandwidth float64, packetLoss float64) float64 {
	// Latency score (0-40 points)
	latencyScore := 40.0
	if latency > qm.thresholds.AcceptableLatency {
		latencyScore = 10.0
	} else if latency > qm.thresholds.GoodLatency {
		latencyScore = 20.0
	} else if latency > qm.thresholds.ExcellentLatency {
		latencyScore = 30.0
	}

	// Bandwidth score (0-40 points)
	bandwidthScore := 40.0
	if bandwidth < qm.thresholds.AcceptableBandwidth {
		bandwidthScore = 10.0
	} else if bandwidth < qm.thresholds.GoodBandwidth {
		bandwidthScore = 20.0
	} else if bandwidth < qm.thresholds.ExcellentBandwidth {
		bandwidthScore = 30.0
	}

	// Packet loss score (0-20 points)
	packetLossScore := 20.0
	if packetLoss > qm.thresholds.AcceptablePacketLoss*10 {
		packetLossScore = 5.0
	} else if packetLoss > qm.thresholds.AcceptablePacketLoss {
		packetLossScore = 10.0
	} else if packetLoss > 0.001 {
		packetLossScore = 15.0
	}

	return latencyScore + bandwidthScore + packetLossScore
}

// recordMetric records a quality metric sample
func (qm *QualityMonitor) recordMetric(connID string, latency int64, bandwidth float64, packetLoss float64) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	metric := QualityMetric{
		Timestamp:  time.Now(),
		Latency:    latency,
		Bandwidth:  bandwidth,
		PacketLoss: packetLoss,
	}

	metrics := qm.metrics[connID]
	metrics = append(metrics, metric)

	// Keep only last N metrics
	if len(metrics) > qm.maxMetricsCount {
		metrics = metrics[len(metrics)-qm.maxMetricsCount:]
	}

	qm.metrics[connID] = metrics
}

// checkAlerts checks for quality alerts
func (qm *QualityMonitor) checkAlerts(connID, peerID string, latency int64, bandwidth float64, packetLoss float64) {
	alerts := make([]QualityAlert, 0)

	// Check latency
	if latency > qm.thresholds.AcceptableLatency*2 {
		alerts = append(alerts, QualityAlert{
			ConnectionID: connID,
			PeerID:       peerID,
			AlertType:    "high_latency",
			Message:      "Very high latency detected",
			Timestamp:    time.Now(),
			Severity:     "critical",
		})
	} else if latency > qm.thresholds.AcceptableLatency {
		alerts = append(alerts, QualityAlert{
			ConnectionID: connID,
			PeerID:       peerID,
			AlertType:    "high_latency",
			Message:      "High latency detected",
			Timestamp:    time.Now(),
			Severity:     "warning",
		})
	}

	// Check bandwidth
	if bandwidth < qm.thresholds.AcceptableBandwidth/2 {
		alerts = append(alerts, QualityAlert{
			ConnectionID: connID,
			PeerID:       peerID,
			AlertType:    "low_bandwidth",
			Message:      "Very low bandwidth detected",
			Timestamp:    time.Now(),
			Severity:     "critical",
		})
	} else if bandwidth < qm.thresholds.AcceptableBandwidth {
		alerts = append(alerts, QualityAlert{
			ConnectionID: connID,
			PeerID:       peerID,
			AlertType:    "low_bandwidth",
			Message:      "Low bandwidth detected",
			Timestamp:    time.Now(),
			Severity:     "warning",
		})
	}

	// Check packet loss
	if packetLoss > qm.thresholds.AcceptablePacketLoss*10 {
		alerts = append(alerts, QualityAlert{
			ConnectionID: connID,
			PeerID:       peerID,
			AlertType:    "high_packet_loss",
			Message:      "Very high packet loss detected",
			Timestamp:    time.Now(),
			Severity:     "critical",
		})
	} else if packetLoss > qm.thresholds.AcceptablePacketLoss {
		alerts = append(alerts, QualityAlert{
			ConnectionID: connID,
			PeerID:       peerID,
			AlertType:    "high_packet_loss",
			Message:      "High packet loss detected",
			Timestamp:    time.Now(),
			Severity:     "warning",
		})
	}

	// Add alerts
	if len(alerts) > 0 {
		qm.mu2.Lock()
		qm.alerts = append(qm.alerts, alerts...)
		qm.mu2.Unlock()
	}
}

// GetAverageLatency gets average latency for a connection
func (qm *QualityMonitor) GetAverageLatency(connID string) int64 {
	metrics := qm.GetMetrics(connID)
	if len(metrics) == 0 {
		return 0
	}

	var total int64
	for _, m := range metrics {
		total += m.Latency
	}
	return total / int64(len(metrics))
}

// GetAverageBandwidth gets average bandwidth for a connection
func (qm *QualityMonitor) GetAverageBandwidth(connID string) float64 {
	metrics := qm.GetMetrics(connID)
	if len(metrics) == 0 {
		return 0.0
	}

	var total float64
	for _, m := range metrics {
		total += m.Bandwidth
	}
	return total / float64(len(metrics))
}

// RemoveConnection removes a connection from monitoring
func (qm *QualityMonitor) RemoveConnection(connID string) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	delete(qm.qualities, connID)
	delete(qm.metrics, connID)
}
