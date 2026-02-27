package mocks

import (
	"github.com/netvideo/signaling"
	"github.com/stretchr/testify/mock"
)

// MockSignalingClient 模拟信令客户端
type MockSignalingClient struct {
	mock.Mock
}

func (m *MockSignalingClient) Connect(serverURL string) error {
	args := m.Called(serverURL)
	return args.Error(0)
}

func (m *MockSignalingClient) SendMessage(msg *signaling.SignalingMessage) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockSignalingClient) ReceiveMessage() (*signaling.SignalingMessage, error) {
	args := m.Called()

	if msg, ok := args.Get(0).(*signaling.SignalingMessage); ok {
		return msg, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockSignalingClient) Register(peerInfo *signaling.PeerInfo) error {
	args := m.Called(peerInfo)
	return args.Error(0)
}

func (m *MockSignalingClient) Unregister() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSignalingClient) LookupPeer(peerID string) (*signaling.PeerInfo, error) {
	args := m.Called(peerID)

	if info, ok := args.Get(0).(*signaling.PeerInfo); ok {
		return info, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockSignalingClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockSignalingServer 模拟信令服务器
type MockSignalingServer struct {
	mock.Mock
}

func (m *MockSignalingServer) Start(addr string) error {
	args := m.Called(addr)
	return args.Error(0)
}

func (m *MockSignalingServer) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSignalingServer) Broadcast(msg *signaling.SignalingMessage) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockSignalingServer) GetOnlinePeers() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockSignalingServer) GetPeerInfo(peerID string) (*signaling.PeerInfo, error) {
	args := m.Called(peerID)

	if info, ok := args.Get(0).(*signaling.PeerInfo); ok {
		return info, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockSignalingServer) HealthCheck() *signaling.ServerHealth {
	args := m.Called()
	return args.Get(0).(*signaling.ServerHealth)
}
