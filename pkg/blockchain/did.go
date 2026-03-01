package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// DID represents a Decentralized Identity
type DID struct {
	ID           string            `json:"id"`
	PublicKey    string            `json:"publicKey"`
	Controller   string            `json:"controller"`
	ServiceEndpoints []ServiceEndpoint `json:"serviceEndpoints"`
	Created      time.Time         `json:"created"`
	Updated      time.Time         `json:"updated"`
	Status       string            `json:"status"` // "active", "revoked", "deactivated"
}

// ServiceEndpoint represents a DID service endpoint
type ServiceEndpoint struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

// DIDDocument represents a DID document
type DIDDocument struct {
	Context          []string   `json:"@context"`
	ID               string     `json:"id"`
	Controller       string     `json:"controller,omitempty"`
	VerificationMethod []VerificationMethod `json:"verificationMethod"`
	Authentication   []string   `json:"authentication"`
	Service          []ServiceEndpoint `json:"service"`
	Created          time.Time  `json:"created"`
	Updated          time.Time  `json:"updated"`
}

// VerificationMethod represents a verification method
type VerificationMethod struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	Controller      string `json:"controller"`
	PublicKeyBase58 string `json:"publicKeyBase58"`
}

// DIDManager manages decentralized identities
type DIDManager struct {
	dids map[string]*DID
}

// NewDIDManager creates a new DID manager
func NewDIDManager() *DIDManager {
	return &DIDManager{
		dids: make(map[string]*DID),
	}
}

// CreateDID creates a new decentralized identity
func (dm *DIDManager) CreateDID(publicKey string, controller string) (*DID, error) {
	// Generate DID
	didID := generateDID(publicKey)

	did := &DID{
		ID:           didID,
		PublicKey:    publicKey,
		Controller:   controller,
		ServiceEndpoints: make([]ServiceEndpoint, 0),
		Created:      time.Now(),
		Updated:      time.Now(),
		Status:       "active",
	}

	dm.dids[didID] = did
	return did, nil
}

// GetDID gets a DID by ID
func (dm *DIDManager) GetDID(didID string) (*DID, error) {
	did, exists := dm.dids[didID]
	if !exists {
		return nil, fmt.Errorf("DID not found")
	}
	return did, nil
}

// UpdateDID updates a DID
func (dm *DIDManager) UpdateDID(didID string, updates map[string]interface{}) error {
	did, exists := dm.dids[didID]
	if !exists {
		return fmt.Errorf("DID not found")
	}

	if did.Status != "active" {
		return fmt.Errorf("DID is not active")
	}

	// Apply updates
	if publicKey, ok := updates["publicKey"].(string); ok {
		did.PublicKey = publicKey
	}

	if controller, ok := updates["controller"].(string); ok {
		did.Controller = controller
	}

	if status, ok := updates["status"].(string); ok {
		did.Status = status
	}

	did.Updated = time.Now()
	return nil
}

// AddServiceEndpoint adds a service endpoint to DID
func (dm *DIDManager) AddServiceEndpoint(didID string, endpoint ServiceEndpoint) error {
	did, exists := dm.dids[didID]
	if !exists {
		return fmt.Errorf("DID not found")
	}

	did.ServiceEndpoints = append(did.ServiceEndpoints, endpoint)
	did.Updated = time.Now()
	return nil
}

// RemoveServiceEndpoint removes a service endpoint from DID
func (dm *DIDManager) RemoveServiceEndpoint(didID, endpointID string) error {
	did, exists := dm.dids[didID]
	if !exists {
		return fmt.Errorf("DID not found")
	}

	for i, endpoint := range did.ServiceEndpoints {
		if endpoint.ID == endpointID {
			did.ServiceEndpoints = append(did.ServiceEndpoints[:i], did.ServiceEndpoints[i+1:]...)
			did.Updated = time.Now()
			return nil
		}
	}

	return fmt.Errorf("endpoint not found")
}

// RevokeDID revokes a DID
func (dm *DIDManager) RevokeDID(didID string) error {
	did, exists := dm.dids[didID]
	if !exists {
		return fmt.Errorf("DID not found")
	}

	did.Status = "revoked"
	did.Updated = time.Now()
	return nil
}

// DeactivateDID deactivates a DID
func (dm *DIDManager) DeactivateDID(didID string) error {
	did, exists := dm.dids[didID]
	if !exists {
		return fmt.Errorf("DID not found")
	}

	did.Status = "deactivated"
	did.Updated = time.Now()
	return nil
}

// GetDIDDocument gets DID document
func (dm *DIDManager) GetDIDDocument(didID string) (*DIDDocument, error) {
	did, err := dm.GetDID(didID)
	if err != nil {
		return nil, err
	}

	doc := &DIDDocument{
		Context: []string{
			"https://www.w3.org/ns/did/v1",
			"https://w3id.org/security/suites/ed25519-2018/v1",
		},
		ID:          did.ID,
		Controller:  did.Controller,
		Created:     did.Created,
		Updated:     did.Updated,
		Service:     did.ServiceEndpoints,
	}

	// Add verification method
	doc.VerificationMethod = []VerificationMethod{
		{
			ID:              did.ID + "#keys-1",
			Type:            "Ed25519VerificationKey2018",
			Controller:      did.Controller,
			PublicKeyBase58: did.PublicKey,
		},
	}

	// Add authentication
	doc.Authentication = []string{did.ID + "#keys-1"}

	return doc, nil
}

// VerifyDID verifies a DID
func (dm *DIDManager) VerifyDID(didID string) (bool, error) {
	did, err := dm.GetDID(didID)
	if err != nil {
		return false, err
	}

	return did.Status == "active", nil
}

// ListDIDs lists all DIDs
func (dm *DIDManager) ListDIDs() []*DID {
	dids := make([]*DID, 0, len(dm.dids))
	for _, did := range dm.dids {
		dids = append(dids, did)
	}
	return dids
}

// GetDIDStats gets DID statistics
func (dm *DIDManager) GetDIDStats() map[string]interface{} {
	total := len(dm.dids)
	active := 0
	revoked := 0
	deactivated := 0

	for _, did := range dm.dids {
		switch did.Status {
		case "active":
			active++
		case "revoked":
			revoked++
		case "deactivated":
			deactivated++
		}
	}

	return map[string]interface{}{
		"total_dids":     total,
		"active_dids":    active,
		"revoked_dids":   revoked,
		"deactivated_dids": deactivated,
	}
}

// generateDID generates a DID from public key
func generateDID(publicKey string) string {
	hash := sha256.Sum256([]byte(publicKey))
	return fmt.Sprintf("did:o2o:%s", hex.EncodeToString(hash[:8]))
}
