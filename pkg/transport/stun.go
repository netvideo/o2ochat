package transport

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

// STUNServer represents a STUN server configuration
type STUNServer struct {
	Address  string `json:"address"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"` // "udp" or "tcp"
}

// DefaultSTUNServers returns a list of default STUN servers
func DefaultSTUNServers() []STUNServer {
	return []STUNServer{
		{Address: "stun.l.google.com", Port: 19302, Protocol: "udp"},
		{Address: "stun1.l.google.com", Port: 19302, Protocol: "udp"},
		{Address: "stun2.l.google.com", Port: 19302, Protocol: "udp"},
		{Address: "stun3.l.google.com", Port: 19302, Protocol: "udp"},
		{Address: "stun4.l.google.com", Port: 19302, Protocol: "udp"},
		{Address: "stun.services.mozilla.com", Port: 3478, Protocol: "udp"},
		{Address: "stun.stunprotocol.org", Port: 3478, Protocol: "udp"},
		{Address: "stun.voip.blackberry.com", Port: 3478, Protocol: "udp"},
	}
}

// STUNPoller manages STUN server polling
type STUNPoller struct {
	servers       []STUNServer
	currentIndex  int
	mu            sync.RWMutex
	timeout       time.Duration
	maxRetries    int
	lastSuccess   time.Time
	failedServers map[string]int
	mu2           sync.RWMutex
}

// STUNPollerConfig represents STUN poller configuration
type STUNPollerConfig struct {
	Timeout    time.Duration
	MaxRetries int
}

// NewSTUNPoller creates a new STUN poller
func NewSTUNPoller(servers []STUNServer, config *STUNPollerConfig) *STUNPoller {
	if config == nil {
		config = &STUNPollerConfig{
			Timeout:    5 * time.Second,
			MaxRetries: 3,
		}
	}

	return &STUNPoller{
		servers:       servers,
		currentIndex:  0,
		timeout:       config.Timeout,
		maxRetries:    config.MaxRetries,
		failedServers: make(map[string]int),
	}
}

// GetNextServer gets the next available STUN server
func (sp *STUNPoller) GetNextServer() *STUNServer {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	if len(sp.servers) == 0 {
		return nil
	}

	return &sp.servers[sp.currentIndex%len(sp.servers)]
}

// RotateServer rotates to the next STUN server
func (sp *STUNPoller) RotateServer() {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	sp.currentIndex++
}

// ProbeServer probes a STUN server and returns the public address
func (sp *STUNPoller) ProbeServer(ctx context.Context, server *STUNServer) (*net.UDPAddr, error) {
	// Create UDP connection
	conn, err := net.DialTimeout(server.Protocol, server.Address, sp.timeout)
	if err != nil {
		sp.reportServerFailure(server)
		return nil, err
	}
	defer conn.Close()

	// Set deadline
	conn.SetDeadline(time.Now().Add(sp.timeout))

	// Send STUN binding request (simplified)
	// In a real implementation, you would send a proper STUN binding request
	stunRequest := []byte{0x00, 0x01, 0x00, 0x00, 0x21, 0x12, 0xA4, 0x42,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00}

	_, err = conn.Write(stunRequest)
	if err != nil {
		sp.reportServerFailure(server)
		return nil, err
	}

	// Read response
	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	if err != nil {
		sp.reportServerFailure(server)
		return nil, err
	}

	// Parse STUN response (simplified)
	// In a real implementation, you would parse the STUN response to extract XOR-MAPPED-ADDRESS
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	// Update last success
	sp.mu.Lock()
	sp.lastSuccess = time.Now()
	sp.mu.Unlock()

	return localAddr, nil
}

// GetPublicAddress gets the public address using STUN
func (sp *STUNPoller) GetPublicAddress(ctx context.Context) (*net.UDPAddr, error) {
	var lastErr error

	for retry := 0; retry < sp.maxRetries; retry++ {
		server := sp.GetNextServer()
		if server == nil {
			return nil, errors.New("no STUN servers available")
		}

		addr, err := sp.ProbeServer(ctx, server)
		if err == nil {
			sp.RotateServer()
			return addr, nil
		}

		lastErr = err
		sp.RotateServer()
	}

	return nil, lastErr
}

// TestConnectivity tests connectivity to STUN servers
func (sp *STUNPoller) TestConnectivity(ctx context.Context) map[string]bool {
	sp.mu2.RLock()
	servers := make([]STUNServer, len(sp.servers))
	copy(servers, sp.servers)
	sp.mu2.RUnlock()

	results := make(map[string]bool)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, server := range servers {
		wg.Add(1)
		go func(s STUNServer) {
			defer wg.Done()
			addr, err := sp.ProbeServer(ctx, &s)
			mu.Lock()
			defer mu.Unlock()
			key := s.Address + ":" + string(rune(s.Port))
			if err == nil && addr != nil {
				results[key] = true
			} else {
				results[key] = false
			}
		}(server)
	}

	wg.Wait()
	return results
}

// GetHealthyServers returns a list of healthy STUN servers
func (sp *STUNPoller) GetHealthyServers(ctx context.Context) []STUNServer {
	sp.mu2.RLock()
	servers := make([]STUNServer, 0, len(sp.servers))
	for _, server := range sp.servers {
		failCount := sp.failedServers[server.Address]
		if failCount < sp.maxRetries {
			servers = append(servers, server)
		}
	}
	sp.mu2.RUnlock()
	return servers
}

// GetLastSuccess returns the last successful probe time
func (sp *STUNPoller) GetLastSuccess() time.Time {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.lastSuccess
}

// GetStats returns STUN poller statistics
func (sp *STUNPoller) GetStats() map[string]interface{} {
	sp.mu.RLock()
	sp.mu2.RLock()
	defer sp.mu.RUnlock()
	defer sp.mu2.RUnlock()

	return map[string]interface{}{
		"total_servers":   len(sp.servers),
		"current_index":   sp.currentIndex,
		"last_success":    sp.lastSuccess,
		"failed_servers":  sp.failedServers,
		"healthy_servers": len(sp.GetHealthyServers(context.Background())),
	}
}

// reportServerFailure reports a server failure
func (sp *STUNPoller) reportServerFailure(server *STUNServer) {
	sp.mu2.Lock()
	defer sp.mu2.Unlock()
	sp.failedServers[server.Address]++
}

// ResetFailures resets failure counts
func (sp *STUNPoller) ResetFailures() {
	sp.mu2.Lock()
	defer sp.mu2.Unlock()
	sp.failedServers = make(map[string]int)
}

// AddServer adds a new STUN server
func (sp *STUNPoller) AddServer(server STUNServer) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.servers = append(sp.servers, server)
}

// RemoveServer removes a STUN server
func (sp *STUNPoller) RemoveServer(address string) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	for i, server := range sp.servers {
		if server.Address == address {
			sp.servers = append(sp.servers[:i], sp.servers[i+1:]...)
			break
		}
	}
}
