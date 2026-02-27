package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type PerformanceOptimizer struct {
	commandStats map[string]*CommandStats
	statsMu      sync.RWMutex
	enabled      atomic.Bool
	historySize  int
}

type CommandStats struct {
	Name          string
	CallCount     atomic.Int64
	TotalDuration atomic.Int64
	MaxDuration   atomic.Int64
	ErrorCount    atomic.Int64
	LastCall      atomic.Value
}

func NewPerformanceOptimizer(historySize int) *PerformanceOptimizer {
	if historySize <= 0 {
		historySize = 1000
	}
	return &PerformanceOptimizer{
		commandStats: make(map[string]*CommandStats),
		historySize:  historySize,
	}
}

func (p *PerformanceOptimizer) Start() {
	p.enabled.Store(true)
}

func (p *PerformanceOptimizer) Stop() {
	p.enabled.Store(false)
}

func (p *PerformanceOptimizer) IsEnabled() bool {
	return p.enabled.Load()
}

func (p *PerformanceOptimizer) RecordCommand(name string, duration time.Duration, err error) {
	if !p.enabled.Load() {
		return
	}

	p.statsMu.RLock()
	stats, exists := p.commandStats[name]
	p.statsMu.RUnlock()

	if !exists {
		p.statsMu.Lock()
		if stats, exists = p.commandStats[name]; !exists {
			stats = &CommandStats{Name: name}
			p.commandStats[name] = stats
		}
		p.statsMu.Unlock()
	}

	stats.CallCount.Add(1)
	stats.TotalDuration.Add(int64(duration))

	currentMax := int64(stats.MaxDuration.Load())
	if int64(duration) > currentMax {
		stats.MaxDuration.Store(int64(duration))
	}

	if err != nil {
		stats.ErrorCount.Add(1)
	}

	stats.LastCall.Store(time.Now())
}

func (p *PerformanceOptimizer) GetStats() map[string]interface{} {
	p.statsMu.RLock()
	defer p.statsMu.RUnlock()

	result := make(map[string]interface{})
	for name, stats := range p.commandStats {
		callCount := stats.CallCount.Load()
		if callCount == 0 {
			continue
		}

		totalDuration := stats.TotalDuration.Load()
		avgDuration := time.Duration(totalDuration / callCount)

		var lastCall time.Time
		if v := stats.LastCall.Load(); v != nil {
			lastCall = v.(time.Time)
		}

		result[name] = map[string]interface{}{
			"call_count":     callCount,
			"total_duration": totalDuration,
			"avg_duration":   avgDuration.String(),
			"max_duration":   stats.MaxDuration.Load(),
			"error_count":    stats.ErrorCount.Load(),
			"last_call":      lastCall.Format(time.RFC3339),
			"success_rate":   float64(callCount-stats.ErrorCount.Load()) / float64(callCount) * 100,
		}
	}

	return result
}

func (p *PerformanceOptimizer) ResetStats() {
	p.statsMu.Lock()
	defer p.statsMu.Unlock()
	p.commandStats = make(map[string]*CommandStats)
}

func (p *PerformanceOptimizer) GetTopCommands(limit int) []map[string]interface{} {
	if limit <= 0 {
		limit = 10
	}

	stats := p.GetStats()

	type commandStat struct {
		name  string
		stats map[string]interface{}
	}

	var commands []commandStat
	for name, s := range stats {
		cmdStats, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		commands = append(commands, commandStat{name: name, stats: cmdStats})
	}

	for i := 0; i < len(commands); i++ {
		for j := i + 1; j < len(commands); j++ {
			c1 := commands[i].stats["call_count"].(int64)
			c2 := commands[j].stats["call_count"].(int64)
			if c2 > c1 {
				commands[i], commands[j] = commands[j], commands[i]
			}
		}
	}

	if len(commands) > limit {
		commands = commands[:limit]
	}

	result := make([]map[string]interface{}, len(commands))
	for i, cmd := range commands {
		result[i] = cmd.stats
		result[i]["name"] = cmd.name
	}

	return result
}

type CommandCache struct {
	cache     map[string]*CacheEntry
	cacheMu   sync.RWMutex
	maxSize   int
	ttl       time.Duration
	hitCount  atomic.Int64
	missCount atomic.Int64
}

type CacheEntry struct {
	Value      interface{}
	ExpireTime time.Time
}

func NewCommandCache(maxSize int, ttl time.Duration) *CommandCache {
	if maxSize <= 0 {
		maxSize = 100
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	cache := &CommandCache{
		cache:   make(map[string]*CacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}

	go cache.cleanup()

	return cache
}

func (c *CommandCache) Get(key string) (interface{}, bool) {
	c.cacheMu.RLock()
	entry, exists := c.cache[key]
	c.cacheMu.RUnlock()

	if !exists {
		c.missCount.Add(1)
		return nil, false
	}

	if time.Now().After(entry.ExpireTime) {
		c.cacheMu.Lock()
		delete(c.cache, key)
		c.cacheMu.Unlock()
		c.missCount.Add(1)
		return nil, false
	}

	c.hitCount.Add(1)
	return entry.Value, true
}

func (c *CommandCache) Set(key string, value interface{}) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	c.cache[key] = &CacheEntry{
		Value:      value,
		ExpireTime: time.Now().Add(c.ttl),
	}
}

func (c *CommandCache) Delete(key string) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	delete(c.cache, key)
}

func (c *CommandCache) Clear() {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cache = make(map[string]*CacheEntry)
}

func (c *CommandCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.cache {
		if oldestTime.IsZero() || entry.ExpireTime.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.ExpireTime
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
	}
}

func (c *CommandCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cacheMu.Lock()
		now := time.Now()
		for key, entry := range c.cache {
			if now.After(entry.ExpireTime) {
				delete(c.cache, key)
			}
		}
		c.cacheMu.Unlock()
	}
}

func (c *CommandCache) GetStats() map[string]interface{} {
	hitCount := c.hitCount.Load()
	missCount := c.missCount.Load()
	total := hitCount + missCount

	var hitRate float64
	if total > 0 {
		hitRate = float64(hitCount) / float64(total) * 100
	}

	return map[string]interface{}{
		"size":       len(c.cache),
		"max_size":   c.maxSize,
		"hit_count":  hitCount,
		"miss_count": missCount,
		"hit_rate":   hitRate,
		"ttl":        c.ttl.String(),
	}
}

type OutputBuffer struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

func NewOutputBuffer() *OutputBuffer {
	return &OutputBuffer{}
}

func (ob *OutputBuffer) Write(p []byte) (n int, err error) {
	ob.mu.Lock()
	defer ob.mu.Unlock()
	return ob.buffer.Write(p)
}

func (ob *OutputBuffer) WriteString(s string) (n int, err error) {
	ob.mu.Lock()
	defer ob.mu.Unlock()
	return ob.buffer.WriteString(s)
}

func (ob *OutputBuffer) WriteFormat(format string, args ...interface{}) (n int, err error) {
	ob.mu.Lock()
	defer ob.mu.Unlock()
	return fmt.Fprintf(&ob.buffer, format, args...)
}

func (ob *OutputBuffer) String() string {
	ob.mu.Lock()
	defer ob.mu.Unlock()
	return ob.buffer.String()
}

func (ob *OutputBuffer) Reset() {
	ob.mu.Lock()
	defer ob.mu.Unlock()
	ob.buffer.Reset()
}

func (ob *OutputBuffer) Len() int {
	ob.mu.Lock()
	defer ob.mu.Unlock()
	return ob.buffer.Len()
}

type ProgressBar struct {
	writer      io.Writer
	total       int
	current     int
	startTime   time.Time
	mu          sync.Mutex
	width       int
	showPercent bool
}

func NewProgressBar(writer io.Writer, total int) *ProgressBar {
	if writer == nil {
		writer = os.Stdout
	}
	return &ProgressBar{
		writer:      writer,
		total:       total,
		width:       50,
		showPercent: true,
		startTime:   time.Now(),
	}
}

func (pb *ProgressBar) SetWidth(width int) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	if width > 0 {
		pb.width = width
	}
}

func (pb *ProgressBar) ShowPercent(show bool) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.showPercent = show
}

func (pb *ProgressBar) Add(n int) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.current += n
	if pb.current > pb.total {
		pb.current = pb.total
	}
}

func (pb *ProgressBar) Set(n int) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.current = n
	if pb.current > pb.total {
		pb.current = pb.total
	}
}

func (pb *ProgressBar) Render() {
	pb.mu.Lock()
	current := pb.current
	total := pb.total
	width := pb.width
	showPercent := pb.showPercent
	pb.mu.Unlock()

	if total == 0 {
		return
	}

	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))

	var buf bytes.Buffer
	buf.WriteString("\r[")
	for i := 0; i < width; i++ {
		if i < filled {
			buf.WriteString("=")
		} else if i == filled {
			buf.WriteString(">")
		} else {
			buf.WriteString(" ")
		}
	}
	buf.WriteString("]")

	if showPercent {
		buf.WriteString(fmt.Sprintf(" %.1f%%", percent*100))
	}

	if current > 0 {
		elapsed := time.Since(pb.startTime)
		rate := float64(current) / elapsed.Seconds()
		remain := time.Duration(float64(total-current) / rate * float64(time.Second))
		buf.WriteString(fmt.Sprintf(" %d/%d %s", current, total, remain))
	}

	pb.writer.Write(buf.Bytes())
}

func (pb *ProgressBar) Finish() {
	pb.mu.Lock()
	pb.current = pb.total
	pb.mu.Unlock()
	pb.Render()
	fmt.Fprintln(pb.writer)
}

type RateLimiter struct {
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

func NewRateLimiter(maxTokens int, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:     float64(maxTokens),
		maxTokens:  float64(maxTokens),
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens += elapsed * rl.refillRate
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}
	rl.lastRefill = now

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) Wait() {
	for !rl.Allow() {
		time.Sleep(10 * time.Millisecond)
	}
}

func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	return map[string]interface{}{
		"tokens":      rl.tokens,
		"max_tokens":  rl.maxTokens,
		"refill_rate": rl.refillRate,
	}
}

type SystemInfo struct {
	GoVersion    string
	NumCPU       int
	NumGoroutine int
	MemStats     runtime.MemStats
}

func GetSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		MemStats:     m,
	}
}

func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

type TableWriter struct {
	writer    io.Writer
	headers   []string
	rows      [][]string
	colWidths []int
}

func NewTableWriter(headers []string) *TableWriter {
	return &TableWriter{
		writer:    os.Stdout,
		headers:   headers,
		colWidths: make([]int, len(headers)),
	}
}

func (tw *TableWriter) SetWriter(w io.Writer) {
	tw.writer = w
}

func (tw *TableWriter) AddRow(row []string) {
	if len(row) != len(tw.headers) {
		return
	}

	for i, cell := range row {
		if len(cell) > tw.colWidths[i] {
			tw.colWidths[i] = len(cell)
		}
	}

	tw.rows = append(tw.rows, row)
}

func (tw *TableWriter) Render() {
	tw.renderSeparator('+', '-')

	headerLine := tw.formatRow(tw.headers)
	tw.writer.Write([]byte(headerLine + "\n"))

	tw.renderSeparator('+', '-')

	for _, row := range tw.rows {
		line := tw.formatRow(row)
		tw.writer.Write([]byte(line + "\n"))
	}

	tw.renderSeparator('+', '-')
}

func (tw *TableWriter) formatRow(row []string) string {
	var buf bytes.Buffer
	buf.WriteString("|")

	for i, cell := range row {
		buf.WriteString(" ")
		buf.WriteString(fmt.Sprintf("%-*s", tw.colWidths[i], cell))
		buf.WriteString(" |")
	}

	return buf.String()
}

func (tw *TableWriter) renderSeparator(c byte, fill byte) {
	var buf bytes.Buffer
	buf.WriteByte(c)

	for _, width := range tw.colWidths {
		buf.WriteByte(fill)
		for i := 0; i < width; i++ {
			buf.WriteByte(fill)
		}
		buf.WriteByte(fill)
		buf.WriteByte(c)
	}

	buf.WriteByte('\n')
	tw.writer.Write(buf.Bytes())
}

func ReadLine(reader *bufio.Reader, prompt string) (string, error) {
	if prompt != "" {
		fmt.Print(prompt)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return line[:len(line)-1], nil
}
