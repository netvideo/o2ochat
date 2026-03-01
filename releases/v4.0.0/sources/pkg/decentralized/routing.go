package decentralized

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"sync"
	"time"
)

// RoutingTable represents an optimized DHT routing table
type RoutingTable struct {
	buckets    []*KBucket
	nodeID     NodeID
	bucketSize int
	mu         sync.RWMutex
}

// KBucket represents a Kademlia K-bucket
type KBucket struct {
	nodes       []*Node
	lastUpdated time.Time
	mu          sync.RWMutex
}

// RoutingConfig represents routing table configuration
type RoutingConfig struct {
	BucketSize int // K value (typically 20)
	Alpha      int // Parallelism factor (typically 3)
}

// NewRoutingTable creates a new optimized routing table
func NewRoutingTable(nodeID NodeID, config *RoutingConfig) *RoutingTable {
	if config == nil {
		config = &RoutingConfig{
			BucketSize: 20,
			Alpha:      3,
		}
	}

	return &RoutingTable{
		buckets:    make([]*KBucket, 256),
		nodeID:     nodeID,
		bucketSize: config.BucketSize,
	}
}

// Initialize creates all K-buckets
func (rt *RoutingTable) Initialize() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	for i := 0; i < 256; i++ {
		rt.buckets[i] = &KBucket{
			nodes: make([]*Node, 0),
		}
	}
}

// AddNode adds a node to the routing table
func (rt *RoutingTable) AddNode(node *Node) bool {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	bucketIndex := rt.getBucketIndex(node.ID)
	bucket := rt.buckets[bucketIndex]

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Check if node already exists
	for i, n := range bucket.nodes {
		if n.ID == node.ID {
			// Update existing node
			bucket.nodes[i] = node
			bucket.lastUpdated = time.Now()
			return true
		}
	}

	// Add new node if bucket not full
	if len(bucket.nodes) < rt.bucketSize {
		bucket.nodes = append(bucket.nodes, node)
		bucket.lastUpdated = time.Now()
		return true
	}

	// Bucket is full, check if we should replace
	// (implement eviction policy based on node age/activity)
	return false
}

// RemoveNode removes a node from the routing table
func (rt *RoutingTable) RemoveNode(nodeID NodeID) bool {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	bucketIndex := rt.getBucketIndex(nodeID)
	bucket := rt.buckets[bucketIndex]

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	for i, n := range bucket.nodes {
		if n.ID == nodeID {
			bucket.nodes = append(bucket.nodes[:i], bucket.nodes[i+1:]...)
			return true
		}
	}

	return false
}

// FindClosestNodes finds the k closest nodes to target
func (rt *RoutingTable) FindClosestNodes(target NodeID, k int) []*Node {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	allNodes := make([]*Node, 0)
	for _, bucket := range rt.buckets {
		bucket.mu.RLock()
		allNodes = append(allNodes, bucket.nodes...)
		bucket.mu.RUnlock()
	}

	// Sort by distance to target
	sortByDistance(allNodes, target)

	if len(allNodes) > k {
		return allNodes[:k]
	}
	return allNodes
}

// getBucketIndex returns the bucket index for a node ID
func (rt *RoutingTable) getBucketIndex(nodeID NodeID) int {
	// XOR distance calculation
	dist := calculateXORDistance(string(rt.nodeID), string(nodeID))

	// Find the first differing bit
	for i := 0; i < 256; i++ {
		if dist.Bit(i) == 1 {
			return i
		}
	}
	return 0
}

// GetNode retrieves a node from the routing table
func (rt *RoutingTable) GetNode(nodeID NodeID) (*Node, bool) {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	bucketIndex := rt.getBucketIndex(nodeID)
	bucket := rt.buckets[bucketIndex]

	bucket.mu.RLock()
	defer bucket.mu.RUnlock()

	for _, n := range bucket.nodes {
		if n.ID == nodeID {
			return n, true
		}
	}

	return nil, false
}

// GetBucketCount returns the number of nodes in routing table
func (rt *RoutingTable) GetBucketCount() int {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	count := 0
	for _, bucket := range rt.buckets {
		bucket.mu.RLock()
		count += len(bucket.nodes)
		bucket.mu.RUnlock()
	}
	return count
}

// CalculateXORDistance calculates XOR distance between two node IDs
func calculateXORDistance(id1, id2 string) *big.Int {
	hash1 := sha256.Sum256([]byte(id1))
	hash2 := sha256.Sum256([]byte(id2))

	int1 := new(big.Int).SetBytes(hash1[:])
	int2 := new(big.Int).SetBytes(hash2[:])

	return new(big.Int).Xor(int1, int2)
}

// SortByDistance sorts nodes by distance to target
func sortByDistance(nodes []*Node, target NodeID) {
	targetHash := sha256.Sum256([]byte(target))
	targetInt := new(big.Int).SetBytes(targetHash[:])

	// Simple bubble sort (replace with better algorithm for production)
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			nodeHash := sha256.Sum256([]byte(nodes[i].ID))
			nodeInt := new(big.Int).SetBytes(nodeHash[:])

			distI := new(big.Int).Xor(nodeInt, targetInt)

			nodeHashJ := sha256.Sum256([]byte(nodes[j].ID))
			nodeIntJ := new(big.Int).SetBytes(nodeHashJ[:])

			distJ := new(big.Int).Xor(nodeIntJ, targetInt)

			if distI.Cmp(distJ) > 0 {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}

// NodeLookup represents a node lookup operation
type NodeLookup struct {
	target    NodeID
	lookupID  string
	started   time.Time
	completed time.Time
	visited   map[NodeID]bool
	results   []*Node
	mu        sync.RWMutex
}

// NewNodeLookup creates a new node lookup
func NewNodeLookup(target NodeID) *NodeLookup {
	hash := sha256.Sum256([]byte(target))
	lookupID := hex.EncodeToString(hash[:])

	return &NodeLookup{
		target:   target,
		lookupID: lookupID,
		started:  time.Now(),
		visited:  make(map[NodeID]bool),
		results:  make([]*Node, 0),
	}
}

// MarkVisited marks a node as visited
func (nl *NodeLookup) MarkVisited(nodeID NodeID) {
	nl.mu.Lock()
	defer nl.mu.Unlock()
	nl.visited[nodeID] = true
}

// IsVisited checks if a node has been visited
func (nl *NodeLookup) IsVisited(nodeID NodeID) bool {
	nl.mu.RLock()
	defer nl.mu.RUnlock()
	return nl.visited[nodeID]
}

// AddResult adds a result node
func (nl *NodeLookup) AddResult(node *Node) {
	nl.mu.Lock()
	defer nl.mu.Unlock()
	nl.results = append(nl.results, node)
}

// GetResults returns lookup results
func (nl *NodeLookup) GetResults() []*Node {
	nl.mu.RLock()
	defer nl.mu.RUnlock()
	return nl.results
}

// Complete marks lookup as complete
func (nl *NodeLookup) Complete() {
	nl.mu.Lock()
	defer nl.mu.Unlock()
	nl.completed = time.Now()
}

// GetDuration returns lookup duration
func (nl *NodeLookup) GetDuration() time.Duration {
	nl.mu.RLock()
	defer nl.mu.RUnlock()

	if !nl.completed.IsZero() {
		return nl.completed.Sub(nl.started)
	}
	return time.Since(nl.started)
}
