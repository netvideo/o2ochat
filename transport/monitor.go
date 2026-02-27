package transport

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// TransportMonitor 传输监控器
type TransportMonitor struct {
	mu         sync.RWMutex
	enabled    bool
	manager    *transportManager
	metrics    *TransportMetrics
	handlers   []MonitorEventHandler
	eventChan  chan MonitorEvent
	stopChan   chan struct{}
	wg         sync.WaitGroup
	alertRules []AlertRule
	logger     *log.Logger
}

// MonitorEventHandler 监控事件处理器
type MonitorEventHandler interface {
	OnEvent(event MonitorEvent)
}

// MonitorEventHandlerFunc 函数类型的监控事件处理器
type MonitorEventHandlerFunc func(event MonitorEvent)

// OnEvent implements MonitorEventHandler
func (f MonitorEventHandlerFunc) OnEvent(event MonitorEvent) {
	f(event)
}

// MonitorEvent 监控事件
type MonitorEvent struct {
	Type      MonitorEventType `json:"type"`
	Timestamp time.Time        `json:"timestamp"`
	Data      interface{}      `json:"data"`
	Source    string           `json:"source"`
	Severity  EventSeverity    `json:"severity"`
}

// MonitorEventType 监控事件类型
type MonitorEventType string

const (
	EventTypeConnection    MonitorEventType = "connection"
	EventTypeDisconnection MonitorEventType = "disconnection"
	EventTypeError         MonitorEventType = "error"
	EventTypeMetric        MonitorEventType = "metric"
	EventTypeAlert         MonitorEventType = "alert"
	EventTypeStateChange   MonitorEventType = "state_change"
)

// EventSeverity 事件严重级别
type EventSeverity string

const (
	SeverityInfo     EventSeverity = "info"
	SeverityWarning  EventSeverity = "warning"
	SeverityError    EventSeverity = "error"
	SeverityCritical EventSeverity = "critical"
)

// TransportMetrics 传输指标
type TransportMetrics struct {
	mu                   sync.RWMutex
	TotalConnections     int64            `json:"total_connections"`
	ActiveConnections    int64            `json:"active_connections"`
	TotalBytesSent       uint64           `json:"total_bytes_sent"`
	TotalBytesReceived   uint64           `json:"total_bytes_received"`
	TotalPacketsSent     uint64           `json:"total_packets_sent"`
	TotalPacketsReceived uint64           `json:"total_packets_received"`
	FailedConnections    int64            `json:"failed_connections"`
	Retransmissions      uint64           `json:"retransmissions"`
	AverageLatency       time.Duration    `json:"average_latency"`
	AverageThroughput    int64            `json:"average_throughput"`
	StartTime            time.Time        `json:"start_time"`
	LastUpdateTime       time.Time        `json:"last_update_time"`
	ConnectionTypeCounts map[string]int64 `json:"connection_type_counts"`
}

// AlertRule 告警规则
type AlertRule struct {
	Name       string        `json:"name"`
	Condition  string        `json:"condition"`
	Threshold  float64       `json:"threshold"`
	Duration   time.Duration `json:"duration"`
	Severity   EventSeverity `json:"severity"`
	Enabled    bool          `json:"enabled"`
	Action     AlertAction   `json:"action"`
	LastAlert  time.Time     `json:"last_alert"`
	AlertCount int64         `json:"alert_count"`
}

// AlertAction 告警动作
type AlertAction struct {
	Type       string   `json:"type"`
	Recipients []string `json:"recipients"`
	WebhookURL string   `json:"webhook_url"`
	Command    string   `json:"command"`
}

// NewTransportMonitor 创建新的传输监控器
func NewTransportMonitor(manager *transportManager) *TransportMonitor {
	return &TransportMonitor{
		enabled:    true,
		manager:    manager,
		metrics:    NewTransportMetrics(),
		handlers:   make([]MonitorEventHandler, 0),
		eventChan:  make(chan MonitorEvent, 1000),
		stopChan:   make(chan struct{}),
		alertRules: make([]AlertRule, 0),
		logger:     log.New(log.Writer(), "[TransportMonitor] ", log.LstdFlags),
	}
}

// NewTransportMetrics 创建新的传输指标
func NewTransportMetrics() *TransportMetrics {
	return &TransportMetrics{
		StartTime:            time.Now(),
		LastUpdateTime:       time.Now(),
		ConnectionTypeCounts: make(map[string]int64),
	}
}

// Enable 启用监控器
func (tm *TransportMonitor) Enable() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.enabled = true
	tm.logger.Println("Monitor enabled")
}

// Disable 禁用监控器
func (tm *TransportMonitor) Disable() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.enabled = false
	tm.logger.Println("Monitor disabled")
}

// IsEnabled 检查监控器是否启用
func (tm *TransportMonitor) IsEnabled() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.enabled
}

// Start 启动监控器
func (tm *TransportMonitor) Start(ctx context.Context) {
	tm.wg.Add(2)

	// 启动事件处理goroutine
	go func() {
		defer tm.wg.Done()
		tm.processEvents(ctx)
	}()

	// 启动定期指标收集goroutine
	go func() {
		defer tm.wg.Done()
		tm.collectMetrics(ctx)
	}()

	tm.logger.Println("Monitor started")
}

// Stop 停止监控器
func (tm *TransportMonitor) Stop() {
	close(tm.stopChan)
	tm.wg.Wait()
	tm.logger.Println("Monitor stopped")
}

// processEvents 处理监控事件
func (tm *TransportMonitor) processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-tm.stopChan:
			return
		case event := <-tm.eventChan:
			tm.handleEvent(event)
		}
	}
}

// handleEvent 处理单个事件
func (tm *TransportMonitor) handleEvent(event MonitorEvent) {
	// 调用注册的事件处理器
	tm.mu.RLock()
	handlers := make([]MonitorEventHandler, len(tm.handlers))
	copy(handlers, tm.handlers)
	tm.mu.RUnlock()

	for _, handler := range handlers {
		go handler.OnEvent(event)
	}

	// 根据事件类型更新指标
	switch event.Type {
	case EventTypeConnection:
		tm.recordConnection(event)
	case EventTypeDisconnection:
		tm.recordDisconnection(event)
	case EventTypeError:
		tm.recordError(event)
	}

	// 检查告警规则
	tm.checkAlertRules(event)
}

// recordConnection 记录连接事件
func (tm *TransportMonitor) recordConnection(event MonitorEvent) {
	tm.metrics.mu.Lock()
	defer tm.metrics.mu.Unlock()

	tm.metrics.TotalConnections++
	tm.metrics.ActiveConnections++
	tm.metrics.LastUpdateTime = time.Now()
}

// recordDisconnection 记录断开连接事件
func (tm *TransportMonitor) recordDisconnection(event MonitorEvent) {
	tm.metrics.mu.Lock()
	defer tm.metrics.mu.Unlock()

	tm.metrics.ActiveConnections--
	tm.metrics.LastUpdateTime = time.Now()
}

// recordError 记录错误事件
func (tm *TransportMonitor) recordError(event MonitorEvent) {
	tm.metrics.mu.Lock()
	defer tm.metrics.mu.Unlock()

	tm.metrics.FailedConnections++
	tm.metrics.LastUpdateTime = time.Now()
}

// collectMetrics 定期收集指标
func (tm *TransportMonitor) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tm.stopChan:
			return
		case <-ticker.C:
			tm.updateMetrics()
		}
	}
}

// updateMetrics 更新指标
func (tm *TransportMonitor) updateMetrics() {
	tm.metrics.mu.Lock()
	defer tm.metrics.mu.Unlock()

	// 计算平均吞吐量
	uptime := time.Since(tm.metrics.StartTime)
	if uptime > 0 {
		totalBytes := tm.metrics.TotalBytesSent + tm.metrics.TotalBytesReceived
		tm.metrics.AverageThroughput = int64(float64(totalBytes) / uptime.Seconds())
	}

	tm.metrics.LastUpdateTime = time.Now()
}

// checkAlertRules 检查告警规则
func (tm *TransportMonitor) checkAlertRules(event MonitorEvent) {
	tm.mu.RLock()
	rules := tm.alertRules
	tm.mu.RUnlock()

	for i := range rules {
		if !rules[i].Enabled {
			continue
		}

		if tm.evaluateAlertRule(&rules[i], event) {
			tm.triggerAlert(&rules[i], event)
		}
	}
}

// evaluateAlertRule 评估告警规则
func (tm *TransportMonitor) evaluateAlertRule(rule *AlertRule, event MonitorEvent) bool {
	// 简化的告警规则评估
	switch rule.Condition {
	case "error_rate":
		tm.metrics.mu.RLock()
		total := tm.metrics.TotalConnections
		failed := tm.metrics.FailedConnections
		tm.metrics.mu.RUnlock()
		if total > 0 && float64(failed)/float64(total) > rule.Threshold {
			return true
		}
	case "latency":
		tm.metrics.mu.RLock()
		latency := tm.metrics.AverageLatency
		tm.metrics.mu.RUnlock()
		if latency > time.Duration(rule.Threshold) {
			return true
		}
	}
	return false
}

// triggerAlert 触发告警
func (tm *TransportMonitor) triggerAlert(rule *AlertRule, event MonitorEvent) {
	rule.LastAlert = time.Now()
	rule.AlertCount++

	alertEvent := MonitorEvent{
		Type:      EventTypeAlert,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"rule_name":     rule.Name,
			"severity":      rule.Severity,
			"trigger_event": event,
		},
		Source:   "alert_manager",
		Severity: rule.Severity,
	}

	tm.handleEvent(alertEvent)
}

// AddEventHandler 添加事件处理器
func (tm *TransportMonitor) AddEventHandler(handler MonitorEventHandler) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.handlers = append(tm.handlers, handler)
}

// RemoveEventHandler 移除事件处理器
func (tm *TransportMonitor) RemoveEventHandler(handler MonitorEventHandler) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i, h := range tm.handlers {
		if h == handler {
			tm.handlers = append(tm.handlers[:i], tm.handlers[i+1:]...)
			break
		}
	}
}

// AddAlertRule 添加告警规则
func (tm *TransportMonitor) AddAlertRule(rule AlertRule) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.alertRules = append(tm.alertRules, rule)
}

// RemoveAlertRule 移除告警规则
func (tm *TransportMonitor) RemoveAlertRule(name string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i, rule := range tm.alertRules {
		if rule.Name == name {
			tm.alertRules = append(tm.alertRules[:i], tm.alertRules[i+1:]...)
			break
		}
	}
}

// EmitEvent 发送事件
func (tm *TransportMonitor) EmitEvent(event MonitorEvent) {
	if !tm.IsEnabled() {
		return
	}

	select {
	case tm.eventChan <- event:
	default:
		// 事件队列已满，丢弃事件
		tm.logger.Println("Event queue is full, dropping event")
	}
}

// GetMetrics 获取指标
func (tm *TransportMonitor) GetMetrics() *TransportMetrics {
	tm.metrics.mu.RLock()
	defer tm.metrics.mu.RUnlock()

	// 返回副本
	metrics := &TransportMetrics{
		TotalConnections:     tm.metrics.TotalConnections,
		ActiveConnections:    tm.metrics.ActiveConnections,
		FailedConnections:    tm.metrics.FailedConnections,
		TotalBytesSent:       tm.metrics.TotalBytesSent,
		TotalBytesReceived:   tm.metrics.TotalBytesReceived,
		TotalPacketsSent:     tm.metrics.TotalPacketsSent,
		TotalPacketsReceived: tm.metrics.TotalPacketsReceived,
		Retransmissions:      tm.metrics.Retransmissions,
		AverageLatency:       tm.metrics.AverageLatency,
		AverageThroughput:    tm.metrics.AverageThroughput,
		StartTime:            tm.metrics.StartTime,
		LastUpdateTime:       tm.metrics.LastUpdateTime,
		ConnectionTypeCounts: make(map[string]int64),
	}

	for k, v := range tm.metrics.ConnectionTypeCounts {
		metrics.ConnectionTypeCounts[k] = v
	}

	return metrics
}

// MetricsToJSON 将指标转换为JSON
func (tm *TransportMonitor) MetricsToJSON() ([]byte, error) {
	metrics := tm.GetMetrics()
	return json.Marshal(metrics)
}

// GetConnectionCount 获取连接数
func (tm *TransportMonitor) GetConnectionCount() int64 {
	tm.metrics.mu.RLock()
	defer tm.metrics.mu.RUnlock()
	return tm.metrics.ActiveConnections
}

// IsHealthy 检查传输层是否健康
func (tm *TransportMonitor) IsHealthy() bool {
	tm.metrics.mu.RLock()
	defer tm.metrics.mu.RUnlock()

	// 健康检查逻辑：活跃连接数不为负，错误率不超标
	if tm.metrics.ActiveConnections < 0 {
		return false
	}

	if tm.metrics.TotalConnections > 0 {
		errorRate := float64(tm.metrics.FailedConnections) / float64(tm.metrics.TotalConnections)
		if errorRate > 0.5 { // 错误率超过50%认为不健康
			return false
		}
	}

	return true
}
