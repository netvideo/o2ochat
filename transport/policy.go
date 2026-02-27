package transport

import (
	"context"
	"net"
	"sort"
	"sync"
	"time"
)

// ConnectionSelector 连接选择器 - 根据优先级选择最佳连接
type ConnectionSelector struct {
	mu         sync.RWMutex
	strategies []ConnectionStrategy
}

// ConnectionStrategy 连接策略接口
type ConnectionStrategy interface {
	// GetPriority 获取策略优先级，值越小优先级越高
	GetPriority() int
	// CanConnect 判断是否可以使用此策略连接
	CanConnect(ctx context.Context, config *ConnectionConfig) bool
	// Connect 建立连接
	Connect(ctx context.Context, config *ConnectionConfig) (Connection, error)
	// GetName 获取策略名称
	GetName() string
}

// ConnectionAttempt 连接尝试记录
type ConnectionAttempt struct {
	StrategyName string
	TargetAddr   string
	Success      bool
	Error        error
	Duration     time.Duration
	Timestamp    time.Time
}

// NewConnectionSelector 创建连接选择器
func NewConnectionSelector() *ConnectionSelector {
	return &ConnectionSelector{
		strategies: make([]ConnectionStrategy, 0),
	}
}

// AddStrategy 添加连接策略
func (cs *ConnectionSelector) AddStrategy(strategy ConnectionStrategy) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.strategies = append(cs.strategies, strategy)

	// 按优先级排序
	sort.Slice(cs.strategies, func(i, j int) bool {
		return cs.strategies[i].GetPriority() < cs.strategies[j].GetPriority()
	})
}

// RemoveStrategy 移除连接策略
func (cs *ConnectionSelector) RemoveStrategy(name string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for i, strategy := range cs.strategies {
		if strategy.GetName() == name {
			cs.strategies = append(cs.strategies[:i], cs.strategies[i+1:]...)
			break
		}
	}
}

// SelectAndConnect 选择并建立连接（核心降级逻辑）
func (cs *ConnectionSelector) SelectAndConnect(ctx context.Context, config *ConnectionConfig) (Connection, []ConnectionAttempt, error) {
	cs.mu.RLock()
	strategies := make([]ConnectionStrategy, len(cs.strategies))
	copy(strategies, cs.strategies)
	cs.mu.RUnlock()

	attempts := make([]ConnectionAttempt, 0)
	var lastErr error

	for _, strategy := range strategies {
		attempt := ConnectionAttempt{
			StrategyName: strategy.GetName(),
			Timestamp:    time.Now(),
		}

		// 检查是否可以使用此策略
		if !strategy.CanConnect(ctx, config) {
			attempt.Error = ErrConnectionFailed
			attempts = append(attempts, attempt)
			continue
		}

		// 尝试建立连接
		start := time.Now()
		conn, err := strategy.Connect(ctx, config)
		attempt.Duration = time.Since(start)

		if err != nil {
			attempt.Success = false
			attempt.Error = err
			lastErr = err
			attempts = append(attempts, attempt)
			continue
		}

		attempt.Success = true
		attempts = append(attempts, attempt)

		return conn, attempts, nil
	}

	if lastErr != nil {
		return nil, attempts, lastErr
	}
	return nil, attempts, ErrConnectionFailed
}

// FallbackStrategy 降级策略
type FallbackStrategy struct {
	mu              sync.RWMutex
	enabled         bool
	attemptInterval time.Duration
	maxAttempts     int
}

// NewFallbackStrategy 创建降级策略
func NewFallbackStrategy() *FallbackStrategy {
	return &FallbackStrategy{
		enabled:         true,
		attemptInterval: 1 * time.Second,
		maxAttempts:     3,
	}
}

// Enable 启用降级策略
func (fs *FallbackStrategy) Enable() {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.enabled = true
}

// Disable 禁用降级策略
func (fs *FallbackStrategy) Disable() {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.enabled = false
}

// IsEnabled 检查是否启用
func (fs *FallbackStrategy) IsEnabled() bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.enabled
}

// NetworkDetector 网络检测器
type NetworkDetector struct {
	mu            sync.RWMutex
	enabled       bool
	checkInterval time.Duration
	lastResult    *NetworkInfo
	listeners     []NetworkChangeListener
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NetworkChangeListener 网络变化监听器
type NetworkChangeListener interface {
	OnNetworkChange(oldInfo, newInfo *NetworkInfo)
}

// NetworkChangeListenerFunc 函数类型的网络变化监听器
type NetworkChangeListenerFunc func(oldInfo, newInfo *NetworkInfo)

// OnNetworkChange implements NetworkChangeListener
func (f NetworkChangeListenerFunc) OnNetworkChange(oldInfo, newInfo *NetworkInfo) {
	f(oldInfo, newInfo)
}

// NetworkInfo 网络信息
type NetworkInfo struct {
	Type          NetworkType   `json:"type"`
	LocalAddrs    []string      `json:"local_addrs"`
	HasIPv6       bool          `json:"has_ipv6"`
	HasIPv4       bool          `json:"has_ipv4"`
	PublicIP      string        `json:"public_ip,omitempty"`
	NATType       string        `json:"nat_type,omitempty"`
	Latency       time.Duration `json:"latency"`
	BandwidthUp   int64         `json:"bandwidth_up"`
	BandwidthDown int64         `json:"bandwidth_down"`
	CheckedAt     time.Time     `json:"checked_at"`
}

// NewNetworkDetector 创建网络检测器
func NewNetworkDetector() *NetworkDetector {
	return &NetworkDetector{
		enabled:       true,
		checkInterval: 30 * time.Second,
		listeners:     make([]NetworkChangeListener, 0),
		stopChan:      make(chan struct{}),
	}
}

// Start 启动网络检测
func (nd *NetworkDetector) Start() {
	nd.wg.Add(1)
	go func() {
		defer nd.wg.Done()
		ticker := time.NewTicker(nd.checkInterval)
		defer ticker.Stop()

		// 立即执行一次检测
		nd.detect()

		for {
			select {
			case <-nd.stopChan:
				return
			case <-ticker.C:
				nd.detect()
			}
		}
	}()
}

// Stop 停止网络检测
func (nd *NetworkDetector) Stop() {
	close(nd.stopChan)
	nd.wg.Wait()
}

// detect 执行网络检测
func (nd *NetworkDetector) detect() {
	info := &NetworkInfo{
		CheckedAt: time.Now(),
	}

	// 检测本地网络接口
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				ip := ipNet.IP
				if !ip.IsLoopback() {
					info.LocalAddrs = append(info.LocalAddrs, ip.String())
					if ip.To4() != nil {
						info.HasIPv4 = true
					} else {
						info.HasIPv6 = true
					}
				}
			}
		}
	}

	// 确定网络类型
	if info.HasIPv6 {
		info.Type = NetworkTypeIPv6
	} else if info.HasIPv4 {
		info.Type = NetworkTypeIPv4
	} else {
		info.Type = NetworkTypeUnknown
	}

	// 通知监听器
	nd.mu.RLock()
	oldInfo := nd.lastResult
	listeners := make([]NetworkChangeListener, len(nd.listeners))
	copy(listeners, nd.listeners)
	disabled := !nd.enabled
	nd.mu.RUnlock()

	if !disabled && oldInfo != nil && nd.hasNetworkChanged(oldInfo, info) {
		for _, listener := range listeners {
			go listener.OnNetworkChange(oldInfo, info)
		}
	}

	nd.mu.Lock()
	nd.lastResult = info
	nd.mu.Unlock()
}

// hasNetworkChanged 检查网络是否发生变化
func (nd *NetworkDetector) hasNetworkChanged(oldInfo, newInfo *NetworkInfo) bool {
	if oldInfo.Type != newInfo.Type {
		return true
	}
	if oldInfo.HasIPv6 != newInfo.HasIPv6 || oldInfo.HasIPv4 != newInfo.HasIPv4 {
		return true
	}
	if len(oldInfo.LocalAddrs) != len(newInfo.LocalAddrs) {
		return true
	}
	return false
}

// GetLastResult 获取最后一次检测结果
func (nd *NetworkDetector) GetLastResult() *NetworkInfo {
	nd.mu.RLock()
	defer nd.mu.RUnlock()
	return nd.lastResult
}

// AddListener 添加网络变化监听器
func (nd *NetworkDetector) AddListener(listener NetworkChangeListener) {
	nd.mu.Lock()
	defer nd.mu.Unlock()
	nd.listeners = append(nd.listeners, listener)
}

// RemoveListener 移除网络变化监听器
func (nd *NetworkDetector) RemoveListener(listener NetworkChangeListener) {
	nd.mu.Lock()
	defer nd.mu.Unlock()
	for i, l := range nd.listeners {
		if l == listener {
			nd.listeners = append(nd.listeners[:i], nd.listeners[i+1:]...)
			break
		}
	}
}

// SetCheckInterval 设置检测间隔
func (nd *NetworkDetector) SetCheckInterval(interval time.Duration) {
	nd.mu.Lock()
	defer nd.mu.Unlock()
	nd.checkInterval = interval
}

// Enable 启用检测器
func (nd *NetworkDetector) Enable() {
	nd.mu.Lock()
	defer nd.mu.Unlock()
	nd.enabled = true
}

// Disable 禁用检测器
func (nd *NetworkDetector) Disable() {
	nd.mu.Lock()
	defer nd.mu.Unlock()
	nd.enabled = false
}

// ConnectionOptimizer 连接优化器
type ConnectionOptimizer struct {
	mu                sync.RWMutex
	enabled           bool
	bufferSize        int
	maxStreams        int
	keepAliveInterval time.Duration
	congestionControl string
}

// NewConnectionOptimizer 创建连接优化器
func NewConnectionOptimizer() *ConnectionOptimizer {
	return &ConnectionOptimizer{
		enabled:           true,
		bufferSize:        64 * 1024,
		maxStreams:        100,
		keepAliveInterval: 30 * time.Second,
		congestionControl: "cubic",
	}
}

// OptimizeConfig 优化连接配置
func (co *ConnectionOptimizer) OptimizeConfig(config *ConnectionConfig) *ConnectionConfig {
	co.mu.RLock()
	defer co.mu.RUnlock()

	if !co.enabled || config == nil {
		return config
	}

	// 创建副本并优化
	optimized := &ConnectionConfig{
		PeerID:        config.PeerID,
		IPv6Addresses: make([]string, len(config.IPv6Addresses)),
		IPv4Addresses: make([]string, len(config.IPv4Addresses)),
		Priority:      make([]ConnectionType, len(config.Priority)),
		Timeout:       config.Timeout,
		RetryCount:    config.RetryCount,
	}

	copy(optimized.IPv6Addresses, config.IPv6Addresses)
	copy(optimized.IPv4Addresses, config.IPv4Addresses)
	copy(optimized.Priority, config.Priority)

	// 根据网络条件优化超时
	if optimized.Timeout == 0 {
		optimized.Timeout = 10 * time.Second
	}

	// 优化重试次数
	if optimized.RetryCount == 0 {
		optimized.RetryCount = 3
	}

	return optimized
}

// Enable 启用优化器
func (co *ConnectionOptimizer) Enable() {
	co.mu.Lock()
	defer co.mu.Unlock()
	co.enabled = true
}

// Disable 禁用优化器
func (co *ConnectionOptimizer) Disable() {
	co.mu.Lock()
	defer co.mu.Unlock()
	co.enabled = false
}

// IsEnabled 检查是否启用
func (co *ConnectionOptimizer) IsEnabled() bool {
	co.mu.RLock()
	defer co.mu.RUnlock()
	return co.enabled
}

// GetBufferSize 获取缓冲区大小
func (co *ConnectionOptimizer) GetBufferSize() int {
	co.mu.RLock()
	defer co.mu.RUnlock()
	return co.bufferSize
}

// SetBufferSize 设置缓冲区大小
func (co *ConnectionOptimizer) SetBufferSize(size int) {
	co.mu.Lock()
	defer co.mu.Unlock()
	co.bufferSize = size
}
