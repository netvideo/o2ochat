package mocks

import (
	"github.com/netvideo/transport"
	"github.com/stretchr/testify/mock"
)

// MockTransportManager 模拟传输管理器
type MockTransportManager struct {
	mock.Mock
}

func (m *MockTransportManager) Listen(addr string) error {
	args := m.Called(addr)
	return args.Error(0)
}

func (m *MockTransportManager) Connect(config *transport.ConnectionConfig) (transport.Connection, error) {
	args := m.Called(config)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(transport.Connection), args.Error(1)
}

func (m *MockTransportManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransportManager) GetConnection(peerID string) (transport.Connection, error) {
	args := m.Called(peerID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(transport.Connection), args.Error(1)
}

func (m *MockTransportManager) ListConnections() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockTransportManager) GetStats() *transport.TransportStats {
	args := m.Called()
	return args.Get(0).(*transport.TransportStats)
}

// MockConnection 模拟连接
type MockConnection struct {
	mock.Mock
}

func (m *MockConnection) OpenStream(config *transport.StreamConfig) (transport.Stream, error) {
	args := m.Called(config)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(transport.Stream), args.Error(1)
}

func (m *MockConnection) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConnection) GetInfo() *transport.ConnectionInfo {
	args := m.Called()
	return args.Get(0).(*transport.ConnectionInfo)
}

func (m *MockConnection) GetStats() *transport.ConnectionStats {
	args := m.Called()
	return args.Get(0).(*transport.ConnectionStats)
}

func (m *MockConnection) SetHandler(handler transport.ConnectionHandler) {
	m.Called(handler)
}

// MockStream 模拟流
type MockStream struct {
	mock.Mock
}

func (m *MockStream) Read(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockStream) Write(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockStream) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStream) SetDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockStream) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockStream) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockStream) GetInfo() *transport.StreamInfo {
	args := m.Called()
	return args.Get(0).(*transport.StreamInfo)
}
