package decentralized

import (
	"sync"
	"time"
)

// AnomalyType represents the type of anomaly
type AnomalyType string

const (
	AnomalyHighRequestRate   AnomalyType = "high_request_rate"
	AnomalyConnectionFlood   AnomalyType = "connection_flood"
	AnomalyMalformedData     AnomalyType = "malformed_data"
	AnomalySuspiciousPattern AnomalyType = "suspicious_pattern"
	AnomalyDDoSAttack        AnomalyType = "ddos_attack"
	AnomalySybilAttack       AnomalyType = "sybil_attack"
)

// Anomaly represents a detected anomaly
type Anomaly struct {
	ID        string                 `json:"id"`
	Type      AnomalyType            `json:"type"`
	Source    string                 `json:"source"`   // IP or NodeID
	Severity  string                 `json:"severity"` // "low", "medium", "high", "critical"
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Action    string                 `json:"action"` // "log", "warn", "block", "ban"
}

// AnomalyDetector detects anomalies in network behavior
type AnomalyDetector struct {
	requestRates    map[string][]time.Time // Source -> request timestamps
	connectionRates map[string][]time.Time
	dataPatterns    map[string][]byte
	baseline        *BaselineMetrics
	thresholds      *AnomalyThresholds
	anomalies       []*Anomaly
	mu              sync.RWMutex
	stats           AnomalyDetectorStats
	mu2             sync.RWMutex
}

// BaselineMetrics represents baseline network metrics
type BaselineMetrics struct {
	AvgRequestRate       float64 `json:"avg_request_rate"`
	AvgConnectionRate    float64 `json:"avg_connection_rate"`
	AvgDataSize          float64 `json:"avg_data_size"`
	StdDevRequestRate    float64 `json:"std_dev_request_rate"`
	StdDevConnectionRate float64 `json:"std_dev_connection_rate"`
	StdDevDataSize       float64 `json:"std_dev_data_size"`
}

// AnomalyThresholds defines anomaly detection thresholds
type AnomalyThresholds struct {
	HighRequestRate   int     `json:"high_request_rate"`   // requests per minute
	ConnectionFlood   int     `json:"connection_flood"`    // connections per second
	DataSizeDeviation float64 `json:"data_size_deviation"` // standard deviations
	SuspiciousPattern int     `json:"suspicious_pattern"`  // pattern match count
	DDoSThreshold     int     `json:"ddos_threshold"`      // requests from multiple sources
	SybilThreshold    int     `json:"sybil_threshold"`     // similar node IDs
}

// AnomalyDetectorStats represents anomaly detector statistics
type AnomalyDetectorStats struct {
	TotalAnomalies   int `json:"total_anomalies"`
	HighSeverity     int `json:"high_severity"`
	CriticalSeverity int `json:"critical_severity"`
	BlockedSources   int `json:"blocked_sources"`
	BannedSources    int `json:"banned_sources"`
}

// DefaultAnomalyThresholds returns default anomaly thresholds
func DefaultAnomalyThresholds() *AnomalyThresholds {
	return &AnomalyThresholds{
		HighRequestRate:   1000,  // 1000 requests per minute
		ConnectionFlood:   100,   // 100 connections per second
		DataSizeDeviation: 3.0,   // 3 standard deviations
		SuspiciousPattern: 10,    // 10 pattern matches
		DDoSThreshold:     10000, // 10000 requests from multiple sources
		SybilThreshold:    50,    // 50 similar node IDs
	}
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(thresholds *AnomalyThresholds) *AnomalyDetector {
	if thresholds == nil {
		thresholds = DefaultAnomalyThresholds()
	}

	return &AnomalyDetector{
		requestRates:    make(map[string][]time.Time),
		connectionRates: make(map[string][]time.Time),
		dataPatterns:    make(map[string][]byte),
		thresholds:      thresholds,
		anomalies:       make([]*Anomaly, 0),
		baseline: &BaselineMetrics{
			AvgRequestRate:    100.0,
			AvgConnectionRate: 10.0,
			AvgDataSize:       1000.0,
		},
	}
}

// RecordRequest records a request for anomaly detection
func (ad *AnomalyDetector) RecordRequest(source string) *Anomaly {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-time.Minute)

	// Get existing requests
	timestamps := ad.requestRates[source]

	// Filter old requests
	validTimestamps := make([]time.Time, 0)
	for _, ts := range timestamps {
		if ts.After(windowStart) {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Add new request
	validTimestamps = append(validTimestamps, now)
	ad.requestRates[source] = validTimestamps

	// Check for high request rate anomaly
	if len(validTimestamps) > ad.thresholds.HighRequestRate {
		anomaly := &Anomaly{
			ID:        generateAnomalyID(),
			Type:      AnomalyHighRequestRate,
			Source:    source,
			Severity:  "high",
			Timestamp: now,
			Data: map[string]interface{}{
				"request_count": len(validTimestamps),
				"threshold":     ad.thresholds.HighRequestRate,
			},
			Action: "block",
		}

		ad.anomalies = append(ad.anomalies, anomaly)
		ad.mu2.Lock()
		ad.stats.TotalAnomalies++
		ad.stats.HighSeverity++
		ad.mu2.Unlock()

		return anomaly
	}

	return nil
}

// RecordConnection records a connection for anomaly detection
func (ad *AnomalyDetector) RecordConnection(source string) *Anomaly {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-time.Second)

	// Get existing connections
	timestamps := ad.connectionRates[source]

	// Filter old connections
	validTimestamps := make([]time.Time, 0)
	for _, ts := range timestamps {
		if ts.After(windowStart) {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Add new connection
	validTimestamps = append(validTimestamps, now)
	ad.connectionRates[source] = validTimestamps

	// Check for connection flood anomaly
	if len(validTimestamps) > ad.thresholds.ConnectionFlood {
		anomaly := &Anomaly{
			ID:        generateAnomalyID(),
			Type:      AnomalyConnectionFlood,
			Source:    source,
			Severity:  "critical",
			Timestamp: now,
			Data: map[string]interface{}{
				"connection_count": len(validTimestamps),
				"threshold":        ad.thresholds.ConnectionFlood,
			},
			Action: "ban",
		}

		ad.anomalies = append(ad.anomalies, anomaly)
		ad.mu2.Lock()
		ad.stats.TotalAnomalies++
		ad.stats.CriticalSeverity++
		ad.mu2.Unlock()

		return anomaly
	}

	return nil
}

// RecordData records data for pattern analysis
func (ad *AnomalyDetector) RecordData(source string, data []byte) *Anomaly {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	// Check for malformed data
	if len(data) == 0 {
		anomaly := &Anomaly{
			ID:        generateAnomalyID(),
			Type:      AnomalyMalformedData,
			Source:    source,
			Severity:  "low",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"data_size": len(data),
			},
			Action: "log",
		}

		ad.anomalies = append(ad.anomalies, anomaly)
		ad.mu2.Lock()
		ad.stats.TotalAnomalies++
		ad.mu2.Unlock()

		return anomaly
	}

	// Store data pattern
	ad.dataPatterns[source] = data

	// Check for suspicious patterns (simplified)
	// In real implementation, use more sophisticated pattern matching

	return nil
}

// GetRecentAnomalies gets recent anomalies
func (ad *AnomalyDetector) GetRecentAnomalies(limit int) []*Anomaly {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	if len(ad.anomalies) <= limit {
		return ad.anomalies
	}

	return ad.anomalies[len(ad.anomalies)-limit:]
}

// GetStats returns anomaly detector statistics
func (ad *AnomalyDetector) GetStats() AnomalyDetectorStats {
	ad.mu2.RLock()
	defer ad.mu2.RUnlock()
	return ad.stats
}

// ClearOldData clears old data to prevent memory growth
func (ad *AnomalyDetector) ClearOldData() {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	now := time.Now()
	requestWindow := now.Add(-5 * time.Minute)
	connectionWindow := now.Add(-1 * time.Minute)

	// Clear old request rates
	for source, timestamps := range ad.requestRates {
		validTimestamps := make([]time.Time, 0)
		for _, ts := range timestamps {
			if ts.After(requestWindow) {
				validTimestamps = append(validTimestamps, ts)
			}
		}

		if len(validTimestamps) == 0 {
			delete(ad.requestRates, source)
		} else {
			ad.requestRates[source] = validTimestamps
		}
	}

	// Clear old connection rates
	for source, timestamps := range ad.connectionRates {
		validTimestamps := make([]time.Time, 0)
		for _, ts := range timestamps {
			if ts.After(connectionWindow) {
				validTimestamps = append(validTimestamps, ts)
			}
		}

		if len(validTimestamps) == 0 {
			delete(ad.connectionRates, source)
		} else {
			ad.connectionRates[source] = validTimestamps
		}
	}

	// Clear old anomalies (keep last 1000)
	if len(ad.anomalies) > 1000 {
		ad.anomalies = ad.anomalies[len(ad.anomalies)-1000:]
	}
}

// IsBlocked checks if a source is blocked
func (ad *AnomalyDetector) IsBlocked(source string) bool {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	// Check recent anomalies for this source
	for _, anomaly := range ad.anomalies {
		if anomaly.Source == source && (anomaly.Action == "block" || anomaly.Action == "ban") {
			if time.Since(anomaly.Timestamp) < 10*time.Minute {
				return true
			}
		}
	}

	return false
}

// Unblock removes block for a source
func (ad *AnomalyDetector) Unblock(source string) {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	// Clear data for this source
	delete(ad.requestRates, source)
	delete(ad.connectionRates, source)
	delete(ad.dataPatterns, source)
}

// Helper function to generate anomaly ID
func generateAnomalyID() string {
	return time.Now().Format("20060102150405.000000")
}
