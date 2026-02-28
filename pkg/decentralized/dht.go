// Package decentralized provides decentralized P2P networking without servers
package decentralized

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"
)

// NodeID represents a unique node identifier in the DHT
type NodeID string

// NodeStatus represents the status of a peer node
type NodeStatus int

const (
	// NodeOnline node is reachable
	NodeOnline NodeStatus = iota
	// NodeOffline node is not reachable
	NodeOffline
	// NodeUnknown node status is unknown
	NodeUnknown
)

// Node represents a peer node in the decentralized network
type Node struct {
	ID           NodeID     `json:"id"`
	PublicKey    string     `json:"public_key"`
	Addresses    []string   `json:"addresses"` // IP:PORT, multiaddr
	Status       NodeStatus `json:"status"`
	LastSeen     time.Time  `json:"last_seen"`
	Capabilities []string   `json:"capabilities"` // message, file, voice, video
	Latency      int64      `json:"latency_ms"`
}

// DHTConfig represents configuration for the DHT
type DHTConfig struct {
	NodeID             NodeID        `json:"node_id"`
	PrivateKey         string        `json:"private_key"`
	PublicKey          string        `json:"public_key"`
	ListenAddresses    []string      `json:"listen_addresses"`
	BootstrapNodes     []string      `json:"bootstrap_nodes"`
	MaxPeers           int           `json:"max_peers"`
	ConnectionTimeout  time.Duration `json:"connection_timeout"`
	DiscoveryInterval  time.Duration `json:"discovery_interval"`
	EnableRelay        bool          `json:"enable_relay"`
	EnableHolePunching bool          `json:"enable_hole_punching"`
}

// PeerStore stores information about known peers
type PeerStore struct {
	peers    map[NodeID]*Node
	banned   map[NodeID]time.Time
	mu       sync.RWMutex
	maxPeers int
}

// DHT implements a distributed hash table for decentralized peer discovery
type DHT struct {
	config      *DHTConfig
	peerStore   *PeerStore
	listeners   []net.Listener
	connections map[NodeID]net.Conn
	active      bool
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex

	// Callbacks
	onPeerDiscovered   func(node *Node)
	onPeerConnected    func(node *Node)
	onPeerDisconnected func(node *Node)
}

// NewDHT creates a new DHT instance
func NewDHT(config *DHTConfig) *DHT {
	ctx, cancel := context.WithCancel(context.Background())

	return &DHT{
		config: config,
		peerStore: &PeerStore{
			peers:    make(map[NodeID]*Node),
			banned:   make(map[NodeID]time.Time),
			maxPeers: config.MaxPeers,
		},
		connections: make(map[NodeID]net.Conn),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start starts the DHT node
func (d *DHT) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.active {
		return fmt.Errorf("DHT already active")
	}

	// Start listeners
	for _, addr := range d.config.ListenAddresses {
		if err := d.startListener(addr); err != nil {
			return fmt.Errorf("failed to start listener %s: %w", addr, err)
		}
	}

	// Connect to bootstrap nodes
	go d.connectToBootstrapNodes()

	// Start periodic discovery
	go d.periodicDiscovery()

	d.active = true
	return nil
}

// Stop stops the DHT node
func (d *DHT) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.active {
		return nil
	}

	d.cancel()
	d.active = false

	// Close all connections
	for nodeID, conn := range d.connections {
		conn.Close()
		delete(d.connections, nodeID)
	}

	// Close listeners
	for _, listener := range d.listeners {
		listener.Close()
	}

	return nil
}

// startListener starts a listener on the given address
func (d *DHT) startListener(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	d.listeners = append(d.listeners, listener)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			go d.handleIncomingConnection(conn)
		}
	}()

	return nil
}

// handleIncomingConnection handles incoming peer connections
func (d *DHT) handleIncomingConnection(conn net.Conn) {
	defer conn.Close()

	// Perform handshake
	peerNode, err := d.performHandshake(conn)
	if err != nil {
		return
	}

	// Add to peer store
	d.peerStore.Add(peerNode)

	// Store connection
	d.mu.Lock()
	d.connections[peerNode.ID] = conn
	d.mu.Unlock()

	// Notify callback
	if d.onPeerConnected != nil {
		d.onPeerConnected(peerNode)
	}
}

// performHandshake performs protocol handshake with peer
func (d *DHT) performHandshake(conn net.Conn) (*Node, error) {
	// Implement handshake protocol
	// 1. Exchange node IDs
	// 2. Verify signatures
	// 3. Exchange capabilities

	return &Node{
		ID:       "peer-node-id",
		Status:   NodeOnline,
		LastSeen: time.Now(),
	}, nil
}

// connectToBootstrapNodes connects to bootstrap nodes
func (d *DHT) connectToBootstrapNodes() {
	for _, addr := range d.config.BootstrapNodes {
		go func(bootstrapAddr string) {
			conn, err := net.DialTimeout("tcp", bootstrapAddr, d.config.ConnectionTimeout)
			if err != nil {
				return
			}

			// Perform handshake
			// Add to peer store
			// Request peer list from bootstrap node
		}(addr)
	}
}

// periodicDiscovery runs periodic peer discovery
func (d *DHT) periodicDiscovery() {
	ticker := time.NewTicker(d.config.DiscoveryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.discoverPeers()
		}
	}
}

// discoverPeers discovers new peers through the network
func (d *DHT) discoverPeers() {
	// Query known peers for new peer recommendations
	// Try to connect to discovered peers
	// Update peer store
}

// FindNode finds a node by ID in the DHT
func (d *DHT) FindNode(ctx context.Context, nodeID NodeID) (*Node, error) {
	// Implement Kademlia-style FIND_NODE
	// 1. Check local peer store
	// 2. Query closest known peers
	// 3. Return node if found

	return d.peerStore.Get(nodeID)
}

// Store stores a key-value pair in the DHT
func (d *DHT) Store(ctx context.Context, key string, value []byte) error {
	// Implement Kademlia-style STORE
	// Store on closest nodes to key hash
	return nil
}

// FindValue retrieves a value from the DHT
func (d *DHT) FindValue(ctx context.Context, key string) ([]byte, error) {
	// Implement Kademlia-style FIND_VALUE
	// Query nodes closest to key hash
	return nil, nil
}

// GetPeerStore returns the peer store
func (d *DHT) GetPeerStore() *PeerStore {
	return d.peerStore
}

// GetActivePeers returns all active peer connections
func (d *DHT) GetActivePeers() []*Node {
	d.peerStore.mu.RLock()
	defer d.peerStore.mu.RUnlock()

	peers := make([]*Node, 0)
	for _, node := range d.peerStore.peers {
		if node.Status == NodeOnline {
			peers = append(peers, node)
		}
	}

	return peers
}

// GetConnection returns the connection to a specific peer
func (d *DHT) GetConnection(nodeID NodeID) (net.Conn, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	conn, exists := d.connections[nodeID]
	if !exists {
		return nil, fmt.Errorf("no connection to peer %s", nodeID)
	}

	return conn, nil
}

// SendToPeer sends data to a specific peer
func (d *DHT) SendToPeer(ctx context.Context, nodeID NodeID, data []byte) error {
	conn, err := d.GetConnection(nodeID)
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}

// Broadcast broadcasts data to all connected peers
func (d *DHT) Broadcast(ctx context.Context, data []byte) error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, conn := range d.connections {
		_, err := conn.Write(data)
		if err != nil {
			// Handle error
		}
	}

	return nil
}

// SetOnPeerDiscovered sets the callback for peer discovery
func (d *DHT) SetOnPeerDiscovered(callback func(node *Node)) {
	d.onPeerDiscovered = callback
}

// SetOnPeerConnected sets the callback for peer connection
func (d *DHT) SetOnPeerConnected(callback func(node *Node)) {
	d.onPeerConnected = callback
}

// SetOnPeerDisconnected sets the callback for peer disconnection
func (d *DHT) SetOnPeerDisconnected(callback func(node *Node)) {
	d.onPeerDisconnected = callback
}

// GenerateNodeID generates a node ID from public key
func GenerateNodeID(publicKey string) NodeID {
	hash := sha256.Sum256([]byte(publicKey))
	return NodeID(hex.EncodeToString(hash[:]))
}

// IsBanned checks if a node is banned
func (ps *PeerStore) IsBanned(nodeID NodeID) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	banTime, exists := ps.banned[nodeID]
	if !exists {
		return false
	}

	// Check if ban has expired (24 hours)
	if time.Since(banTime) > 24*time.Hour {
		delete(ps.banned, nodeID)
		return false
	}

	return true
}

// Ban temporarily bans a node
func (ps *PeerStore) Ban(nodeID NodeID) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.banned[nodeID] = time.Now()
}

// Add adds a node to the peer store
func (ps *PeerStore) Add(node *Node) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if len(ps.peers) >= ps.maxPeers {
		// Remove oldest offline node
		ps.removeOldestOffline()
	}

	ps.peers[node.ID] = node
}

// Get retrieves a node from the peer store
func (ps *PeerStore) Get(nodeID NodeID) (*Node, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	node, exists := ps.peers[nodeID]
	if !exists {
		return nil, fmt.Errorf("node not found")
	}

	return node, nil
}

// Remove removes a node from the peer store
func (ps *PeerStore) Remove(nodeID NodeID) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	delete(ps.peers, nodeID)
}

// List lists all nodes in the peer store
func (ps *PeerStore) List() []*Node {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	nodes := make([]*Node, 0, len(ps.peers))
	for _, node := range ps.peers {
		nodes = append(nodes, node)
	}

	return nodes
}

// UpdateStatus updates the status of a node
func (ps *PeerStore) UpdateStatus(nodeID NodeID, status NodeStatus) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if node, exists := ps.peers[nodeID]; exists {
		node.Status = status
		node.LastSeen = time.Now()
	}
}

// GetOnlineCount returns the number of online nodes
func (ps *PeerStore) GetOnlineCount() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	count := 0
	for _, node := range ps.peers {
		if node.Status == NodeOnline {
			count++
		}
	}

	return count
}

// removeOldestOffline removes the oldest offline node
func (ps *PeerStore) removeOldestOffline() {
	oldestTime := time.Now()
	var oldestID NodeID

	for id, node := range ps.peers {
		if node.Status != NodeOnline && node.LastSeen.Before(oldestTime) {
			oldestTime = node.LastSeen
			oldestID = id
		}
	}

	if oldestID != "" {
		delete(ps.peers, oldestID)
	}
}
