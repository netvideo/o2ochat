package transport

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

// Connection represents a pooled connection
type Connection struct {
	ID         string
	PeerID     string
	Conn       net.Conn
	CreatedAt  time.Time
	LastUsedAt time.Time
	UsedCount  int
	Health     ConnectionHealth
	mu         sync.RWMutex
}

// ConnectionHealth represents connection health status
type ConnectionHealth int

const (
	HealthUnknown ConnectionHealth = iota
	HealthGood
	HealthDegraded
	HealthBad
)

// ConnectionPool manages a pool of connections
type ConnectionPool struct {
	connections           map[string]*Connection
	peerConnections       map[string][]string // peerID -> connection IDs
	maxConnectionsPerPeer int
	maxIdleTime           time.Duration
	maxLifetime           time.Duration
	maxTotalConnections   int
	mu                    sync.RWMutex
	cleanupTicker         *time.Ticker
	stopCleanup           chan struct{}
	stats                 PoolStats
	mu2                   sync.RWMutex
}

// PoolStats represents connection pool statistics
type PoolStats struct {
	TotalConnections  int `json:"total_connections"`
	ActiveConnections int `json:"active_connections"`
	IdleConnections   int `json:"idle_connections"`
	TotalCreated      int `json:"total_created"`
	TotalClosed       int `json:"total_closed"`
	TotalRequests     int `json:"total_requests"`
	HealthGood        int `json:"health_good"`
	HealthDegraded    int `json:"health_degraded"`
	HealthBad         int `json:"health_bad"`
}

// ConnectionPoolConfig represents connection pool configuration
type ConnectionPoolConfig struct {
	MaxConnectionsPerPeer int
	MaxIdleTime           time.Duration
	MaxLifetime           time.Duration
	MaxTotalConnections   int
	CleanupInterval       time.Duration
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config *ConnectionPoolConfig) *ConnectionPool {
	if config == nil {
		config = &ConnectionPoolConfig{
			MaxConnectionsPerPeer: 10,
			MaxIdleTime:           5 * time.Minute,
			MaxLifetime:           30 * time.Minute,
			MaxTotalConnections:   1000,
			CleanupInterval:       1 * time.Minute,
		}
	}

	pool := &ConnectionPool{
		connections:           make(map[string]*Connection),
		peerConnections:       make(map[string][]string),
		maxConnectionsPerPeer: config.MaxConnectionsPerPeer,
		maxIdleTime:           config.MaxIdleTime,
		maxLifetime:           config.MaxLifetime,
		maxTotalConnections:   config.MaxTotalConnections,
		cleanupTicker:         time.NewTicker(config.CleanupInterval),
		stopCleanup:           make(chan struct{}),
	}

	// Start cleanup goroutine
	go pool.cleanupLoop()

	return pool
}

// GetConnection gets or creates a connection for a peer
func (cp *ConnectionPool) GetConnection(ctx context.Context, peerID string, dialFunc func() (net.Conn, error)) (*Connection, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.mu2.Lock()
	cp.stats.TotalRequests++
	cp.mu2.Unlock()

	// Try to get existing connection
	conn := cp.getIdleConnection(peerID)
	if conn != nil {
		return conn, nil
	}

	// Check if we can create new connection
	if cp.canCreateConnection(peerID) {
		return cp.createConnection(ctx, peerID, dialFunc)
	}

	// Wait for available connection or timeout
	return cp.waitForConnection(ctx, peerID, dialFunc)
}

// ReturnConnection returns a connection to the pool
func (cp *ConnectionPool) ReturnConnection(connID string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	conn, exists := cp.connections[connID]
	if !exists {
		return
	}

	conn.mu.Lock()
	conn.LastUsedAt = time.Now()
	conn.mu.Unlock()
}

// CloseConnection closes and removes a connection from the pool
func (cp *ConnectionPool) CloseConnection(connID string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.removeConnection(connID)
}

// GetStats returns pool statistics
func (cp *ConnectionPool) GetStats() PoolStats {
	cp.mu.RLock()
	cp.mu2.RLock()
	defer cp.mu.RUnlock()
	defer cp.mu2.RUnlock()

	stats := cp.stats
	stats.TotalConnections = len(cp.connections)

	// Count health status
	for _, conn := range cp.connections {
		conn.mu.RLock()
		switch conn.Health {
		case HealthGood:
			stats.HealthGood++
		case HealthDegraded:
			stats.HealthDegraded++
		case HealthBad:
			stats.HealthBad++
		}
		conn.mu.RUnlock()
	}

	return stats
}

// Close closes the connection pool
func (cp *ConnectionPool) Close() {
	close(cp.stopCleanup)
	cp.cleanupTicker.Stop()

	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Close all connections
	for connID := range cp.connections {
		cp.removeConnection(connID)
	}
}

// getIdleConnection gets an idle connection for a peer
func (cp *ConnectionPool) getIdleConnection(peerID string) *Connection {
	connIDs, exists := cp.peerConnections[peerID]
	if !exists {
		return nil
	}

	for _, connID := range connIDs {
		conn, exists := cp.connections[connID]
		if !exists {
			continue
		}

		conn.mu.RLock()
		isIdle := conn.Conn != nil && time.Since(conn.LastUsedAt) < cp.maxIdleTime
		isHealthy := conn.Health == HealthGood || conn.Health == HealthDegraded
		conn.mu.RUnlock()

		if isIdle && isHealthy {
			return conn
		}
	}

	return nil
}

// canCreateConnection checks if we can create a new connection
func (cp *ConnectionPool) canCreateConnection(peerID string) bool {
	// Check per-peer limit
	connIDs, exists := cp.peerConnections[peerID]
	if exists && len(connIDs) >= cp.maxConnectionsPerPeer {
		return false
	}

	// Check total limit
	if len(cp.connections) >= cp.maxTotalConnections {
		return false
	}

	return true
}

// createConnection creates a new connection
func (cp *ConnectionPool) createConnection(ctx context.Context, peerID string, dialFunc func() (net.Conn, error)) (*Connection, error) {
	// Dial connection
	conn, err := dialFunc()
	if err != nil {
		return nil, err
	}

	// Create connection object
	connID := generateConnectionID()
	connection := &Connection{
		ID:         connID,
		PeerID:     peerID,
		Conn:       conn,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
		UsedCount:  1,
		Health:     HealthGood,
	}

	// Add to pool
	cp.connections[connID] = connection
	cp.peerConnections[peerID] = append(cp.peerConnections[peerID], connID)

	// Update stats
	cp.mu2.Lock()
	cp.stats.TotalCreated++
	cp.stats.ActiveConnections++
	cp.mu2.Unlock()

	return connection, nil
}

// waitForConnection waits for an available connection
func (cp *ConnectionPool) waitForConnection(ctx context.Context, peerID string, dialFunc func() (net.Conn, error)) (*Connection, error) {
	// Wait with timeout
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, errors.New("timeout waiting for available connection")
		case <-ticker.C:
			conn := cp.getIdleConnection(peerID)
			if conn != nil {
				return conn, nil
			}

			if cp.canCreateConnection(peerID) {
				return cp.createConnection(ctx, peerID, dialFunc)
			}
		}
	}
}

// removeConnection removes a connection from the pool
func (cp *ConnectionPool) removeConnection(connID string) {
	conn, exists := cp.connections[connID]
	if !exists {
		return
	}

	// Close underlying connection
	if conn.Conn != nil {
		conn.Conn.Close()
	}

	// Remove from peer connections
	peerID := conn.PeerID
	connIDs := cp.peerConnections[peerID]
	for i, id := range connIDs {
		if id == connID {
			cp.peerConnections[peerID] = append(connIDs[:i], connIDs[i+1:]...)
			break
		}
	}

	// Remove from connections
	delete(cp.connections, connID)

	// Update stats
	cp.mu2.Lock()
	cp.stats.TotalClosed++
	cp.stats.ActiveConnections--
	cp.mu2.Unlock()
}

// cleanupLoop periodically cleans up stale connections
func (cp *ConnectionPool) cleanupLoop() {
	for {
		select {
		case <-cp.stopCleanup:
			return
		case <-cp.cleanupTicker.C:
			cp.cleanup()
		}
	}
}

// cleanup removes stale connections
func (cp *ConnectionPool) cleanup() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	now := time.Now()
	toRemove := make([]string, 0)

	for connID, conn := range cp.connections {
		conn.mu.RLock()
		lifetime := now.Sub(conn.CreatedAt)
		idleTime := now.Sub(conn.LastUsedAt)
		health := conn.Health
		conn.mu.RUnlock()

		// Remove if expired or bad health
		if lifetime > cp.maxLifetime || idleTime > cp.maxIdleTime || health == HealthBad {
			toRemove = append(toRemove, connID)
		}
	}

	// Remove stale connections
	for _, connID := range toRemove {
		cp.removeConnection(connID)
	}

	// Update stats
	cp.mu2.Lock()
	cp.stats.IdleConnections = len(cp.connections)
	cp.mu2.Unlock()
}

// UpdateConnectionHealth updates the health status of a connection
func (cp *ConnectionPool) UpdateConnectionHealth(connID string, health ConnectionHealth) {
	cp.mu.RLock()
	conn, exists := cp.connections[connID]
	cp.mu.RUnlock()

	if !exists {
		return
	}

	conn.mu.Lock()
	conn.Health = health
	conn.mu.Unlock()
}

// GetConnectionByID gets a connection by ID
func (cp *ConnectionPool) GetConnectionByID(connID string) *Connection {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.connections[connID]
}

// GetConnectionsByPeer gets all connections for a peer
func (cp *ConnectionPool) GetConnectionsByPeer(peerID string) []*Connection {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	connIDs, exists := cp.peerConnections[peerID]
	if !exists {
		return nil
	}

	connections := make([]*Connection, 0, len(connIDs))
	for _, connID := range connIDs {
		if conn, exists := cp.connections[connID]; exists {
			connections = append(connections, conn)
		}
	}

	return connections
}

// Helper function to generate connection ID
func generateConnectionID() string {
	return time.Now().Format("20060102150405.000000")
}
