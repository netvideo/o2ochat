package decentralized

import (
	"sync"
	"time"
)

// ReputationScore represents a node's reputation score
type ReputationScore float64

const (
	// MinReputation is the minimum reputation score
	MinReputation ReputationScore = 0.0

	// MaxReputation is the maximum reputation score
	MaxReputation ReputationScore = 100.0

	// DefaultReputation is the default reputation for new nodes
	DefaultReputation ReputationScore = 50.0

	// TrustedReputation is the threshold for trusted nodes
	TrustedReputation ReputationScore = 80.0

	// UntrustedReputation is the threshold for untrusted nodes
	UntrustedReputation ReputationScore = 20.0
)

// NodeReputation represents a node's reputation information
type NodeReputation struct {
	NodeID                 NodeID           `json:"node_id"`
	Score                  ReputationScore  `json:"score"`
	TotalInteractions      int              `json:"total_interactions"`
	SuccessfulInteractions int              `json:"successful_interactions"`
	FailedInteractions     int              `json:"failed_interactions"`
	FirstSeen              time.Time        `json:"first_seen"`
	LastUpdated            time.Time        `json:"last_updated"`
	Flags                  []ReputationFlag `json:"flags,omitempty"`
	mu                     sync.RWMutex
}

// ReputationFlag represents a reputation flag
type ReputationFlag string

const (
	// FlagVerified indicates the node has been verified
	FlagVerified ReputationFlag = "verified"

	// FlagMalicious indicates the node has been flagged as malicious
	FlagMalicious ReputationFlag = "malicious"

	// FlagTrusted indicates the node is trusted
	FlagTrusted ReputationFlag = "trusted"

	// FlagBanned indicates the node is banned
	FlagBanned ReputationFlag = "banned"
)

// ReputationManager manages node reputations
type ReputationManager struct {
	reputations map[NodeID]*NodeReputation
	mu          sync.RWMutex
}

// NewReputationManager creates a new reputation manager
func NewReputationManager() *ReputationManager {
	return &ReputationManager{
		reputations: make(map[NodeID]*NodeReputation),
	}
}

// GetReputation gets the reputation for a node
func (rm *ReputationManager) GetReputation(nodeID NodeID) *NodeReputation {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return rm.reputations[nodeID]
}

// GetOrCreateReputation gets or creates a reputation for a node
func (rm *ReputationManager) GetOrCreateReputation(nodeID NodeID) *NodeReputation {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rep, exists := rm.reputations[nodeID]; exists {
		return rep
	}

	// Create new reputation with default score
	rep := &NodeReputation{
		NodeID:      nodeID,
		Score:       DefaultReputation,
		FirstSeen:   time.Now(),
		LastUpdated: time.Now(),
		Flags:       make([]ReputationFlag, 0),
	}

	rm.reputations[nodeID] = rep
	return rep
}

// UpdateReputation updates a node's reputation based on interaction
func (rm *ReputationManager) UpdateReputation(nodeID NodeID, success bool) {
	rep := rm.GetOrCreateReputation(nodeID)

	rep.mu.Lock()
	defer rep.mu.Unlock()

	rep.TotalInteractions++
	if success {
		rep.SuccessfulInteractions++
		// Increase reputation for successful interaction
		rep.Score = Min(rep.Score+1.0, MaxReputation)
	} else {
		rep.FailedInteractions++
		// Decrease reputation for failed interaction
		rep.Score = Max(rep.Score-5.0, MinReputation)
	}

	rep.LastUpdated = time.Now()

	// Add flags based on score
	if rep.Score >= TrustedReputation && !rep.HasFlag(FlagTrusted) {
		rep.Flags = append(rep.Flags, FlagTrusted)
	}

	if rep.Score < UntrustedReputation && !rep.HasFlag(FlagMalicious) {
		rep.Flags = append(rep.Flags, FlagMalicious)
	}
}

// ReportMalicious reports a node as malicious
func (rm *ReputationManager) ReportMalicious(nodeID NodeID) {
	rep := rm.GetOrCreateReputation(nodeID)

	rep.mu.Lock()
	defer rep.mu.Unlock()

	// Severe penalty for malicious behavior
	rep.Score = Max(rep.Score-20.0, MinReputation)
	rep.LastUpdated = time.Now()

	// Add malicious flag if not already present
	if !rep.HasFlag(FlagMalicious) {
		rep.Flags = append(rep.Flags, FlagMalicious)
	}
}

// VerifyNode marks a node as verified
func (rm *ReputationManager) VerifyNode(nodeID NodeID) {
	rep := rm.GetOrCreateReputation(nodeID)

	rep.mu.Lock()
	defer rep.mu.Unlock()

	// Bonus for verified nodes
	rep.Score = Min(rep.Score+10.0, MaxReputation)
	rep.LastUpdated = time.Now()

	// Add verified flag if not already present
	if !rep.HasFlag(FlagVerified) {
		rep.Flags = append(rep.Flags, FlagVerified)
	}
}

// BanNode bans a node
func (rm *ReputationManager) BanNode(nodeID NodeID) {
	rep := rm.GetOrCreateReputation(nodeID)

	rep.mu.Lock()
	defer rep.mu.Unlock()

	// Set minimum reputation
	rep.Score = MinReputation
	rep.LastUpdated = time.Now()

	// Add banned flag if not already present
	if !rep.HasFlag(FlagBanned) {
		rep.Flags = append(rep.Flags, FlagBanned)
	}
}

// IsTrusted checks if a node is trusted
func (rm *ReputationManager) IsTrusted(nodeID NodeID) bool {
	rep := rm.GetReputation(nodeID)
	if rep == nil {
		return false
	}

	rep.mu.RLock()
	defer rep.mu.RUnlock()

	return rep.Score >= TrustedReputation || rep.HasFlag(FlagTrusted)
}

// IsMalicious checks if a node is malicious
func (rm *ReputationManager) IsMalicious(nodeID NodeID) bool {
	rep := rm.GetReputation(nodeID)
	if rep == nil {
		return false
	}

	rep.mu.RLock()
	defer rep.mu.RUnlock()

	return rep.Score < UntrustedReputation || rep.HasFlag(FlagMalicious)
}

// IsBanned checks if a node is banned
func (rm *ReputationManager) IsBanned(nodeID NodeID) bool {
	rep := rm.GetReputation(nodeID)
	if rep == nil {
		return false
	}

	rep.mu.RLock()
	defer rep.mu.RUnlock()

	return rep.HasFlag(FlagBanned)
}

// HasFlag checks if a node has a specific flag
func (nr *NodeReputation) HasFlag(flag ReputationFlag) bool {
	for _, f := range nr.Flags {
		if f == flag {
			return true
		}
	}
	return false
}

// GetStats returns reputation statistics
func (rm *ReputationManager) GetStats() map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	totalNodes := len(rm.reputations)
	trustedNodes := 0
	maliciousNodes := 0
	bannedNodes := 0
	totalScore := ReputationScore(0)

	for _, rep := range rm.reputations {
		rep.mu.RLock()
		if rep.Score >= TrustedReputation {
			trustedNodes++
		}
		if rep.Score < UntrustedReputation {
			maliciousNodes++
		}
		if rep.HasFlag(FlagBanned) {
			bannedNodes++
		}
		totalScore += rep.Score
		rep.mu.RUnlock()
	}

	avgScore := ReputationScore(0)
	if totalNodes > 0 {
		avgScore = totalScore / ReputationScore(totalNodes)
	}

	return map[string]interface{}{
		"total_nodes":     totalNodes,
		"trusted_nodes":   trustedNodes,
		"malicious_nodes": maliciousNodes,
		"banned_nodes":    bannedNodes,
		"average_score":   avgScore,
	}
}

// Min returns the minimum of two reputation scores
func Min(a, b ReputationScore) ReputationScore {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two reputation scores
func Max(a, b ReputationScore) ReputationScore {
	if a > b {
		return a
	}
	return b
}
