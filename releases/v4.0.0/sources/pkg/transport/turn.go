package transport

import (
	"context"
	"net"
	"sort"
	"sync"
	"time"
)

// TURNTransport represents the transport protocol for TURN
type TURNTransport string

const (
	// TURNTransportUDP uses UDP transport
	TURNTransportUDP TURNTransport = "udp"
	// TURNTransportTCP uses TCP transport
	TURNTransportTCP TURNTransport = "tcp"
	// TURNTransportTLS uses TLS transport
	TURNTransportTLS TURNTransport = "tls"
)

// TURNServer represents a TURN server configuration
type TURNServer struct {
	Address   string        `json:"address"`
	Port      int           `json:"port"`
	Username  string        `json:"username,omitempty"`
	Password  string        `json:"password,omitempty"`
	Transport TURNTransport `json:"transport"`
	Priority  int           `json:"priority"` // Higher is better
	Realm     string        `json:"realm,omitempty"`
}

// DefaultTURNservers returns a list of default TURN servers
func DefaultTURNservers() []TURNServer {
	return []TURNServer{
		{
			Address:   "turn.l.google.com",
			Port:      19302,
			Transport: TURNTransportUDP,
			Priority:  100,
		},
		{
			Address:   "turn1.l.google.com",
			Port:      19302,
			Transport: TURNTransportUDP,
			Priority:  90,
		},
		{
			Address:   "turn2.l.google.com",
			Port:      19302,
			Transport: TURNTransportUDP,
			Priority:  90,
		},
		{
			Address:   "turn.services.mozilla.com",
			Port:      3478,
			Transport: TURNTransportUDP,
			Priority:  80,
		},
	}
}

// TURNSelector manages TURN server selection
type TURNSelector struct {
	servers      []TURNServer
	mu           sync.RWMutex
	timeout      time.Duration
	maxRetries   int
	healthChecks map[string]*ServerHealth
	mu2          sync.RWMutex
}

// ServerHealth represents server health information
type ServerHealth struct {
	Address     string    `json:"address"`
	IsHealthy   bool      `json:"is_healthy"`
	LastChecked time.Time `json:"last_checked"`
	Latency     int64     `json:"latency_ms"` // in milliseconds
	Failures    int       `json:"failures"`
	LastSuccess time.Time `json:"last_success"`
	LastFailure time.Time `json:"last_failure"`
}

// TURNSelectorConfig represents TURN selector configuration
type TURNSelectorConfig struct {
	Timeout    time.Duration
	MaxRetries int
}

// NewTURNSelector creates a new TURN selector
func NewTURNSelector(servers []TURNServer, config *TURNSelectorConfig) *TURNSelector {
	if config == nil {
		config = &TURNSelectorConfig{
			Timeout:    10 * time.Second,
			MaxRetries: 3,
		}
	}

	selector := &TURNSelector{
		servers:      servers,
		timeout:      config.Timeout,
		maxRetries:   config.MaxRetries,
		healthChecks: make(map[string]*ServerHealth),
	}

	// Initialize health checks
	for _, server := range servers {
		key := selector.getServerKey(&server)
		selector.healthChecks[key] = &ServerHealth{
			Address:     server.Address,
			IsHealthy:   true,
			LastChecked: time.Now(),
		}
	}

	return selector
}

// SelectBestServer selects the best available TURN server
func (ts *TURNSelector) SelectBestServer() *TURNServer {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	if len(ts.servers) == 0 {
		return nil
	}

	// Filter healthy servers
	healthyServers := make([]TURNServer, 0)
	for _, server := range ts.servers {
		key := ts.getServerKey(&server)
		ts.mu2.RLock()
		health := ts.healthChecks[key]
		ts.mu2.RUnlock()

		if health != nil && health.IsHealthy && health.Failures < ts.maxRetries {
			healthyServers = append(healthyServers, server)
		}
	}

	if len(healthyServers) == 0 {
		// Fallback to all servers if no healthy ones
		healthyServers = ts.servers
	}

	// Sort by priority (descending)
	sort.Slice(healthyServers, func(i, j int) bool {
		return healthyServers[i].Priority > healthyServers[j].Priority
	})

	return &healthyServers[0]
}

// TestConnection tests connection to a TURN server
func (ts *TURNSelector) TestConnection(ctx context.Context, server *TURNServer) (int64, error) {
	start := time.Now()

	// Create connection based on transport
	var conn net.Conn
	var err error

	addr := net.JoinHostPort(server.Address, string(rune(server.Port)))

	switch server.Transport {
	case TURNTransportTCP:
		conn, err = net.DialTimeout("tcp", addr, ts.timeout)
	case TURNTransportTLS:
		// For TLS, you would use tls.Dial in a real implementation
		conn, err = net.DialTimeout("tcp", addr, ts.timeout)
	default: // UDP
		conn, err = net.DialTimeout("udp", addr, ts.timeout)
	}

	if err != nil {
		ts.reportServerFailure(server, err)
		return 0, err
	}
	defer conn.Close()

	latency := time.Since(start).Milliseconds()

	// Update health on success
	ts.reportServerSuccess(server, latency)

	return latency, nil
}

// GetAllHealth returns health status of all servers
func (ts *TURNSelector) GetAllHealth() map[string]*ServerHealth {
	ts.mu2.RLock()
	defer ts.mu2.RUnlock()

	result := make(map[string]*ServerHealth)
	for key, health := range ts.healthChecks {
		result[key] = health
	}
	return result
}

// GetHealthyServers returns a list of healthy TURN servers
func (ts *TURNSelector) GetHealthyServers() []TURNServer {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	healthy := make([]TURNServer, 0)
	for _, server := range ts.servers {
		key := ts.getServerKey(&server)
		ts.mu2.RLock()
		health := ts.healthChecks[key]
		ts.mu2.RUnlock()

		if health != nil && health.IsHealthy {
			healthy = append(healthy, server)
		}
	}

	return healthy
}

// GetStats returns TURN selector statistics
func (ts *TURNSelector) GetStats() map[string]interface{} {
	ts.mu.RLock()
	ts.mu2.RLock()
	defer ts.mu.RUnlock()
	defer ts.mu2.RUnlock()

	totalServers := len(ts.servers)
	healthyServers := 0
	totalFailures := 0
	avgLatency := int64(0)
	latencyCount := 0

	for _, health := range ts.healthChecks {
		if health.IsHealthy {
			healthyServers++
		}
		totalFailures += health.Failures
		if health.Latency > 0 {
			avgLatency += health.Latency
			latencyCount++
		}
	}

	if latencyCount > 0 {
		avgLatency /= int64(latencyCount)
	}

	return map[string]interface{}{
		"total_servers":   totalServers,
		"healthy_servers": healthyServers,
		"total_failures":  totalFailures,
		"average_latency": avgLatency,
		"max_retries":     ts.maxRetries,
		"timeout_seconds": ts.timeout.Seconds(),
	}
}

// AddServer adds a new TURN server
func (ts *TURNSelector) AddServer(server TURNServer) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.servers = append(ts.servers, server)

	key := ts.getServerKey(&server)
	ts.mu2.Lock()
	ts.healthChecks[key] = &ServerHealth{
		Address:     server.Address,
		IsHealthy:   true,
		LastChecked: time.Now(),
	}
	ts.mu2.Unlock()
}

// RemoveServer removes a TURN server
func (ts *TURNSelector) RemoveServer(address string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	for i, server := range ts.servers {
		if server.Address == address {
			ts.servers = append(ts.servers[:i], ts.servers[i+1:]...)

			key := ts.getServerKey(&server)
			ts.mu2.Lock()
			delete(ts.healthChecks, key)
			ts.mu2.Unlock()

			break
		}
	}
}

// ResetHealth resets all health checks
func (ts *TURNSelector) ResetHealth() {
	ts.mu2.Lock()
	defer ts.mu2.Unlock()

	for key := range ts.healthChecks {
		ts.healthChecks[key] = &ServerHealth{
			Address:     ts.healthChecks[key].Address,
			IsHealthy:   true,
			LastChecked: time.Now(),
			Latency:     0,
			Failures:    0,
		}
	}
}

// reportServerSuccess reports a successful server connection
func (ts *TURNSelector) reportServerSuccess(server *TURNServer, latency int64) {
	key := ts.getServerKey(server)

	ts.mu2.Lock()
	defer ts.mu2.Unlock()

	health, exists := ts.healthChecks[key]
	if !exists {
		health = &ServerHealth{Address: server.Address}
		ts.healthChecks[key] = health
	}

	health.IsHealthy = true
	health.LastChecked = time.Now()
	health.LastSuccess = time.Now()
	health.Latency = latency
	health.Failures = 0
}

// reportServerFailure reports a server connection failure
func (ts *TURNSelector) reportServerFailure(server *TURNServer, err error) {
	key := ts.getServerKey(server)

	ts.mu2.Lock()
	defer ts.mu2.Unlock()

	health, exists := ts.healthChecks[key]
	if !exists {
		health = &ServerHealth{Address: server.Address}
		ts.healthChecks[key] = health
	}

	health.IsHealthy = false
	health.LastChecked = time.Now()
	health.LastFailure = time.Now()
	health.Failures++

	// Mark as unhealthy after max retries
	if health.Failures >= ts.maxRetries {
		health.IsHealthy = false
	}
}

// getServerKey generates a unique key for a server
func (ts *TURNSelector) getServerKey(server *TURNServer) string {
	return server.Address + ":" + string(rune(server.Port)) + ":" + string(server.Transport)
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Server  *TURNServer
	Healthy bool
	Latency int64
	Error   error
}

// CheckAllHealth checks health of all servers concurrently
func (ts *TURNSelector) CheckAllHealth(ctx context.Context) []HealthCheckResult {
	ts.mu.RLock()
	servers := make([]TURNServer, len(ts.servers))
	copy(servers, ts.servers)
	ts.mu.RUnlock()

	results := make([]HealthCheckResult, 0, len(servers))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, server := range servers {
		wg.Add(1)
		go func(s TURNServer) {
			defer wg.Done()

			latency, err := ts.TestConnection(ctx, &s)
			healthy := err == nil

			mu.Lock()
			results = append(results, HealthCheckResult{
				Server:  &s,
				Healthy: healthy,
				Latency: latency,
				Error:   err,
			})
			mu.Unlock()
		}(server)
	}

	wg.Wait()
	return results
}

// SelectByLatency selects the server with the lowest latency
func (ts *TURNSelector) SelectByLatency() *TURNServer {
	ts.mu2.RLock()
	defer ts.mu2.RUnlock()

	var best *TURNServer
	bestLatency := int64(-1)

	for i := range ts.servers {
		key := ts.getServerKey(&ts.servers[i])
		health := ts.healthChecks[key]

		if health != nil && health.IsHealthy && health.Latency > 0 {
			if best == nil || health.Latency < bestLatency {
				best = &ts.servers[i]
				bestLatency = health.Latency
			}
		}
	}

	// Fallback to priority-based selection if no latency data
	if best == nil {
		return ts.SelectBestServer()
	}

	return best
}
