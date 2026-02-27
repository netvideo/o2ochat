package transport

import (
	"context"
	"testing"
	"time"
)

// TestNewConnectionSelector 测试创建连接选择器
func TestNewConnectionSelector(t *testing.T) {
	selector := NewConnectionSelector()
	if selector == nil {
		t.Fatal("NewConnectionSelector returned nil")
	}

	if len(selector.strategies) != 0 {
		t.Errorf("Expected 0 strategies, got %d", len(selector.strategies))
	}
}

// TestConnectionSelectorAddStrategy 测试添加连接策略
func TestConnectionSelectorAddStrategy(t *testing.T) {
	selector := NewConnectionSelector()

	// 创建模拟策略
	strategy1 := &mockConnectionStrategy{
		name:       "Strategy1",
		priority:   1,
		canConnect: true,
	}
	strategy2 := &mockConnectionStrategy{
		name:       "Strategy2",
		priority:   2,
		canConnect: true,
	}

	selector.AddStrategy(strategy1)
	selector.AddStrategy(strategy2)

	if len(selector.strategies) != 2 {
		t.Errorf("Expected 2 strategies, got %d", len(selector.strategies))
	}

	// 验证按优先级排序
	if selector.strategies[0].GetPriority() != 1 {
		t.Errorf("Expected first strategy priority 1, got %d", selector.strategies[0].GetPriority())
	}
}

// TestConnectionSelectorRemoveStrategy 测试移除连接策略
func TestConnectionSelectorRemoveStrategy(t *testing.T) {
	selector := NewConnectionSelector()

	strategy := &mockConnectionStrategy{
		name:       "TestStrategy",
		priority:   1,
		canConnect: true,
	}

	selector.AddStrategy(strategy)
	if len(selector.strategies) != 1 {
		t.Fatal("Expected 1 strategy after add")
	}

	selector.RemoveStrategy("TestStrategy")
	if len(selector.strategies) != 0 {
		t.Errorf("Expected 0 strategies after remove, got %d", len(selector.strategies))
	}
}

// TestConnectionSelectorSelectAndConnect 测试选择和连接
func TestConnectionSelectorSelectAndConnect(t *testing.T) {
	selector := NewConnectionSelector()

	// 创建一个成功的策略
	successStrategy := &mockConnectionStrategy{
		name:       "SuccessStrategy",
		priority:   1,
		canConnect: true,
		connection: &mockConnection{},
	}

	selector.AddStrategy(successStrategy)

	config := &ConnectionConfig{
		PeerID:  "test-peer",
		Timeout: 5 * time.Second,
	}

	ctx := context.Background()
	conn, attempts, err := selector.SelectAndConnect(ctx, config)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if conn == nil {
		t.Error("Expected connection, got nil")
	}

	if len(attempts) != 1 {
		t.Errorf("Expected 1 attempt, got %d", len(attempts))
	}

	if attempts[0].Success != true {
		t.Error("Expected attempt to be successful")
	}
}

// TestConnectionSelectorFallback 测试降级机制
func TestConnectionSelectorFallback(t *testing.T) {
	selector := NewConnectionSelector()

	// 创建一个失败的策略
	failStrategy := &mockConnectionStrategy{
		name:       "FailStrategy",
		priority:   1,
		canConnect: true,
		shouldFail: true,
	}

	// 创建一个成功的策略
	successStrategy := &mockConnectionStrategy{
		name:       "SuccessStrategy",
		priority:   2,
		canConnect: true,
		connection: &mockConnection{},
	}

	selector.AddStrategy(failStrategy)
	selector.AddStrategy(successStrategy)

	config := &ConnectionConfig{
		PeerID:  "test-peer",
		Timeout: 5 * time.Second,
	}

	ctx := context.Background()
	conn, attempts, err := selector.SelectAndConnect(ctx, config)

	if err != nil {
		t.Errorf("Expected no error after fallback, got %v", err)
	}

	if conn == nil {
		t.Error("Expected connection after fallback, got nil")
	}

	if len(attempts) != 2 {
		t.Errorf("Expected 2 attempts (1 fail + 1 success), got %d", len(attempts))
	}

	// 第一个尝试应该失败
	if attempts[0].Success != false {
		t.Error("Expected first attempt to fail")
	}

	// 第二个尝试应该成功
	if attempts[1].Success != true {
		t.Error("Expected second attempt to succeed")
	}
}

// mockConnectionStrategy 用于测试的模拟连接策略
type mockConnectionStrategy struct {
	name       string
	priority   int
	canConnect bool
	connection Connection
	shouldFail bool
}

func (m *mockConnectionStrategy) GetPriority() int {
	return m.priority
}

func (m *mockConnectionStrategy) CanConnect(ctx context.Context, config *ConnectionConfig) bool {
	return m.canConnect
}

func (m *mockConnectionStrategy) Connect(ctx context.Context, config *ConnectionConfig) (Connection, error) {
	if m.shouldFail {
		return nil, ErrConnectionFailed
	}
	return m.connection, nil
}

func (m *mockConnectionStrategy) GetName() string {
	return m.name
}

// mockConnection 用于测试的模拟连接
type mockConnection struct {
	id     string
	closed bool
}

func (m *mockConnection) OpenStream(config *StreamConfig) (Stream, error) {
	return nil, nil
}

func (m *mockConnection) AcceptStream() (Stream, error) {
	return nil, nil
}

func (m *mockConnection) Close() error {
	m.closed = true
	return nil
}

func (m *mockConnection) GetInfo() ConnectionInfo {
	return ConnectionInfo{
		ID: m.id,
	}
}

func (m *mockConnection) SendControlMessage(msg []byte) error {
	return nil
}

func (m *mockConnection) ReceiveControlMessage() ([]byte, error) {
	return nil, nil
}

func (m *mockConnection) GetStats() *ConnectionStats {
	return &ConnectionStats{}
}
