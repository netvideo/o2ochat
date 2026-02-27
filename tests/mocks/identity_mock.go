package mocks

import (
	"github.com/netvideo/identity"
	"github.com/stretchr/testify/mock"
)

// MockIdentityManager 模拟身份管理器
type MockIdentityManager struct {
	mock.Mock
}

func (m *MockIdentityManager) CreateIdentity(config *identity.IdentityConfig) (*identity.Identity, error) {
	args := m.Called(config)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*identity.Identity), args.Error(1)
}

func (m *MockIdentityManager) LoadIdentity(peerID string) (*identity.Identity, error) {
	args := m.Called(peerID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*identity.Identity), args.Error(1)
}

func (m *MockIdentityManager) VerifyIdentity(peerID string, proof *identity.IdentityProof) error {
	args := m.Called(peerID, proof)
	return args.Error(0)
}

func (m *MockIdentityManager) SignMessage(message []byte) ([]byte, error) {
	args := m.Called(message)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockIdentityManager) VerifySignature(peerID string, message, signature []byte) bool {
	args := m.Called(peerID, message, signature)
	return args.Bool(0)
}

func (m *MockIdentityManager) ExportIdentity(peerID, password string) ([]byte, error) {
	args := m.Called(peerID, password)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockIdentityManager) ImportIdentity(data []byte, password string) (*identity.Identity, error) {
	args := m.Called(data, password)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*identity.Identity), args.Error(1)
}

func (m *MockIdentityManager) DeleteIdentity(peerID string) error {
	args := m.Called(peerID)
	return args.Error(0)
}

func (m *MockIdentityManager) ListIdentities() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockIdentityManager) GetMetadata(peerID string) (map[string]interface{}, error) {
	args := m.Called(peerID)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockIdentityManager) UpdateMetadata(peerID string, metadata map[string]interface{}) error {
	args := m.Called(peerID, metadata)
	return args.Error(0)
}

func (m *MockIdentityManager) GenerateChallenge() (*identity.Challenge, error) {
	args := m.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*identity.Challenge), args.Error(1)
}

func (m *MockIdentityManager) VerifyChallenge(peerID string, challenge *identity.Challenge, signature []byte) error {
	args := m.Called(peerID, challenge, signature)
	return args.Error(0)
}

func (m *MockIdentityManager) Close() error {
	args := m.Called()
	return args.Error(0)
}
