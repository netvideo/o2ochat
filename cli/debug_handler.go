package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"sync"
	"time"
)

type DebugHandler struct {
	logPath        string
	enableDebug    bool
	mu             sync.RWMutex
	monitoring     bool
	stopMonitor    chan struct{}
	profileDataDir string
}

func NewDebugHandler(logPath, profileDataDir string) *DebugHandler {
	if logPath == "" {
		logPath = os.Getenv("HOME") + "/.o2ochat/logs"
	}
	if profileDataDir == "" {
		profileDataDir = os.Getenv("HOME") + "/.o2ochat/debug"
	}
	return &DebugHandler{
		logPath:        logPath,
		enableDebug:    os.Getenv("O2OCHAT_DEBUG") == "1",
		stopMonitor:    make(chan struct{}),
		profileDataDir: profileDataDir,
	}
}

func (h *DebugHandler) ShowLogs(level string, lines int) (*CommandResult, error) {
	if lines <= 0 {
		lines = 100
	}

	logFiles, err := filepath.Glob(filepath.Join(h.logPath, "*.log"))
	if err != nil || len(logFiles) == 0 {
		return &CommandResult{
			Success:  true,
			Message:  "no log files found",
			Data:     []string{},
			ExitCode: 0,
		}, nil
	}

	sort.Sort(sort.Reverse(sort.StringSlice(logFiles)))
	latestLog := logFiles[0]

	file, err := os.Open(latestLog)
	if err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to open log file: %v", err),
			ExitCode: 1,
		}, nil
	}
	defer file.Close()

	var filteredLogs []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if h.matchesLevel(line, level) {
			filteredLogs = append(filteredLogs, line)
		}
	}

	if len(filteredLogs) > lines {
		filteredLogs = filteredLogs[len(filteredLogs)-lines:]
	}

	return &CommandResult{
		Success: true,
		Message: fmt.Sprintf("showing %d log entries from %s", len(filteredLogs), filepath.Base(latestLog)),
		Data: map[string]interface{}{
			"level":     level,
			"count":     len(filteredLogs),
			"log_file":  filepath.Base(latestLog),
			"log_lines": filteredLogs,
		},
		ExitCode: 0,
	}, nil
}

func (h *DebugHandler) matchesLevel(line, level string) bool {
	if level == "" || level == "all" {
		return true
	}
	levelPatterns := map[string][]string{
		"debug": {"DEBUG", "DEBU"},
		"info":  {"INFO", "INF"},
		"warn":  {"WARN", "WAR", "WARNING"},
		"error": {"ERROR", "ERR", "FATAL"},
	}

	patterns, ok := levelPatterns[level]
	if !ok {
		return true
	}

	lineUpper := strings.ToUpper(line)
	for _, p := range patterns {
		if strings.Contains(lineUpper, p) {
			return true
		}
	}
	return false
}

func (h *DebugHandler) MonitorPerformance(interval int) (*CommandResult, error) {
	if interval <= 0 {
		interval = 5
	}

	if h.monitoring {
		return &CommandResult{
			Success:  false,
			Message:  "performance monitoring is already running",
			ExitCode: 1,
		}, nil
	}

	h.monitoring = true
	h.stopMonitor = make(chan struct{})

	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				h.collectMetrics()
			case <-h.stopMonitor:
				return
			}
		}
	}()

	return &CommandResult{
		Success: true,
		Message: fmt.Sprintf("performance monitoring started (interval: %ds)", interval),
		Data: map[string]interface{}{
			"interval_seconds": interval,
			"status":           "running",
		},
		ExitCode: 0,
	}, nil
}

func (h *DebugHandler) collectMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	_ = map[string]interface{}{
		"timestamp":    time.Now().Format(time.RFC3339),
		"goroutines":   runtime.NumGoroutine(),
		"memory_alloc": m.Alloc,
		"memory_total": m.TotalAlloc,
		"memory_sys":   m.Sys,
		"gc_runs":      m.NumGC,
		"heap_objects": m.HeapObjects,
		"heap_alloc":   m.HeapAlloc,
		"heap_sys":     m.HeapSys,
	}
}

func (h *DebugHandler) StopMonitoring() (*CommandResult, error) {
	if !h.monitoring {
		return &CommandResult{
			Success:  false,
			Message:  "performance monitoring is not running",
			ExitCode: 1,
		}, nil
	}

	close(h.stopMonitor)
	h.monitoring = false

	return &CommandResult{
		Success:  true,
		Message:  "performance monitoring stopped",
		ExitCode: 0,
	}, nil
}

func (h *DebugHandler) ProfileMemory(duration int) (*CommandResult, error) {
	if duration <= 0 {
		duration = 30
	}

	if err := os.MkdirAll(h.profileDataDir, 0755); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to create profile directory: %v", err),
			ExitCode: 1,
		}, nil
	}

	profilePath := filepath.Join(h.profileDataDir, fmt.Sprintf("mem_%s.pprof", time.Now().Format("20060102_150405")))

	f, err := os.Create(profilePath)
	if err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to create profile file: %v", err),
			ExitCode: 1,
		}, nil
	}
	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to write heap profile: %v", err),
			ExitCode: 1,
		}, nil
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &CommandResult{
		Success: true,
		Message: fmt.Sprintf("memory profile saved to %s", profilePath),
		Data: map[string]interface{}{
			"profile_path": profilePath,
			"duration":     duration,
			"goroutines":   runtime.NumGoroutine(),
			"memory_alloc": m.Alloc,
			"memory_total": m.TotalAlloc,
			"memory_sys":   m.Sys,
			"heap_objects": m.HeapObjects,
			"gc_count":     m.NumGC,
		},
		ExitCode: 0,
	}, nil
}

func (h *DebugHandler) ProfileCPU(duration int) (*CommandResult, error) {
	if duration <= 0 {
		duration = 30
	}

	if err := os.MkdirAll(h.profileDataDir, 0755); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to create profile directory: %v", err),
			ExitCode: 1,
		}, nil
	}

	profilePath := filepath.Join(h.profileDataDir, fmt.Sprintf("cpu_%s.pprof", time.Now().Format("20060102_150405")))

	f, err := os.Create(profilePath)
	if err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to create profile file: %v", err),
			ExitCode: 1,
		}, nil
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to start CPU profile: %v", err),
			ExitCode: 1,
		}, nil
	}

	time.Sleep(time.Duration(duration) * time.Second)
	pprof.StopCPUProfile()

	return &CommandResult{
		Success: true,
		Message: fmt.Sprintf("CPU profile saved to %s", profilePath),
		Data: map[string]interface{}{
			"profile_path": profilePath,
			"duration":     duration,
		},
		ExitCode: 0,
	}, nil
}

func (h *DebugHandler) TraceCall(function string) (*CommandResult, error) {
	if err := os.MkdirAll(h.profileDataDir, 0755); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to create trace directory: %v", err),
			ExitCode: 1,
		}, nil
	}

	tracePath := filepath.Join(h.profileDataDir, fmt.Sprintf("trace_%s.pprof", time.Now().Format("20060102_150405")))

	f, err := os.Create(tracePath)
	if err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to create trace file: %v", err),
			ExitCode: 1,
		}, nil
	}
	defer f.Close()

	if err := pprof.Lookup("goroutine").WriteTo(f, 1); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to write goroutine profile: %v", err),
			ExitCode: 1,
		}, nil
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &CommandResult{
		Success: true,
		Message: fmt.Sprintf("goroutine trace saved to %s", tracePath),
		Data: map[string]interface{}{
			"trace_path":   tracePath,
			"function":     function,
			"goroutines":   runtime.NumGoroutine(),
			"memory_alloc": m.Alloc,
		},
		ExitCode: 0,
	}, nil
}

func (h *DebugHandler) StressTest(target string, duration int) (*CommandResult, error) {
	if duration <= 0 {
		duration = 60
	}

	if target == "" {
		target = "cpu"
	}

	validTargets := map[string]bool{"cpu": true, "memory": true, "io": true}
	if !validTargets[target] {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("invalid target: %s (valid: cpu, memory, io)", target),
			ExitCode: 1,
		}, nil
	}

	results := map[string]interface{}{
		"target":   target,
		"duration": duration,
		"status":   "running",
		"metrics":  map[string]interface{}{},
	}

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		opsCount := 0
		for {
			select {
			case <-ticker.C:
				switch target {
				case "cpu":
					for i := 0; i < 1000000; i++ {
						_ = i * i
					}
				case "memory":
					data := make([]byte, 1024*1024)
					_ = len(data)
				case "io":
					buf := new(bytes.Buffer)
					buf.WriteString("stress test data\n")
				}
				opsCount++
				if opsCount >= duration {
					close(doneCh)
					return
				}
			case <-stopCh:
				close(doneCh)
				return
			}
		}
	}()

	select {
	case <-doneCh:
		results["status"] = "completed"
		results["ops_count"] = duration
	case <-time.After(time.Duration(duration+1) * time.Second):
		results["status"] = "timeout"
		close(stopCh)
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	results["metrics"] = map[string]interface{}{
		"goroutines":   runtime.NumGoroutine(),
		"memory_alloc": m.Alloc,
		"gc_runs":      m.NumGC,
	}

	return &CommandResult{
		Success:  true,
		Message:  fmt.Sprintf("stress test completed (target: %s, duration: %ds)", target, duration),
		Data:     results,
		ExitCode: 0,
	}, nil
}

type ShowLogsHandler struct {
	debugHandler *DebugHandler
}

func NewShowLogsHandler(logPath string) *ShowLogsHandler {
	return &ShowLogsHandler{
		debugHandler: NewDebugHandler(logPath, ""),
	}
}

func (h *ShowLogsHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	level := "all"
	lines := 100

	if v, ok := ctx.Flags["level"].(string); ok && v != "" {
		level = v
	}
	if v, ok := ctx.Flags["lines"].(int); ok && v > 0 {
		lines = v
	}
	if v, ok := ctx.Flags["n"].(int); ok && v > 0 {
		lines = v
	}

	return h.debugHandler.ShowLogs(level, lines)
}

func (h *ShowLogsHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *ShowLogsHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	if len(word) > 1 && word[0] == '-' {
		return []string{"--level", "--lines", "--n", "--help"}, nil
	}
	return nil, nil
}

func (h *ShowLogsHandler) Help() string {
	return `查看O2OChat日志。

用法:
  o2ochat debug logs [选项]

选项:
  -l, --level     日志级别 (debug, info, warn, error, all)
  -n, --lines     显示行数 (默认: 100)
  -h, --help      显示帮助信息

示例:
  o2ochat debug logs
  o2ochat debug logs --level error --lines 50`
}

type MonitorPerformanceHandler struct {
	debugHandler *DebugHandler
}

func NewMonitorPerformanceHandler(logPath string) *MonitorPerformanceHandler {
	return &MonitorPerformanceHandler{
		debugHandler: NewDebugHandler(logPath, ""),
	}
}

func (h *MonitorPerformanceHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	interval := 5

	if v, ok := ctx.Flags["interval"].(int); ok && v > 0 {
		interval = v
	}
	if v, ok := ctx.Flags["i"].(int); ok && v > 0 {
		interval = v
	}

	return h.debugHandler.MonitorPerformance(interval)
}

func (h *MonitorPerformanceHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *MonitorPerformanceHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	if len(word) > 1 && word[0] == '-' {
		return []string{"--interval", "--i", "--help"}, nil
	}
	return nil, nil
}

func (h *MonitorPerformanceHandler) Help() string {
	return `监控性能指标。

用法:
  o2ochat debug monitor [选项]

选项:
  -i, --interval    监控间隔(秒，默认: 5)
  -h, --help       显示帮助信息

示例:
  o2ochat debug monitor
  o2ochat debug monitor --interval 10`
}

type ProfileMemoryHandler struct {
	debugHandler *DebugHandler
}

func NewProfileMemoryHandler(logPath string) *ProfileMemoryHandler {
	return &ProfileMemoryHandler{
		debugHandler: NewDebugHandler(logPath, ""),
	}
}

func (h *ProfileMemoryHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	duration := 30

	if v, ok := ctx.Flags["duration"].(int); ok && v > 0 {
		duration = v
	}
	if v, ok := ctx.Flags["d"].(int); ok && v > 0 {
		duration = v
	}

	return h.debugHandler.ProfileMemory(duration)
}

func (h *ProfileMemoryHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *ProfileMemoryHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	if len(word) > 1 && word[0] == '-' {
		return []string{"--duration", "--d", "--help"}, nil
	}
	return nil, nil
}

func (h *ProfileMemoryHandler) Help() string {
	return `分析内存使用情况。

用法:
  o2ochat debug profile memory [选项]

选项:
  -d, --duration    采样时长(秒，默认: 30)
  -h, --help       显示帮助信息

示例:
  o2ochat debug profile memory
  o2ochat debug profile memory --duration 60`
}

type ProfileCPUHandler struct {
	debugHandler *DebugHandler
}

func NewProfileCPUHandler(logPath string) *ProfileCPUHandler {
	return &ProfileCPUHandler{
		debugHandler: NewDebugHandler(logPath, ""),
	}
}

func (h *ProfileCPUHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	duration := 30

	if v, ok := ctx.Flags["duration"].(int); ok && v > 0 {
		duration = v
	}
	if v, ok := ctx.Flags["d"].(int); ok && v > 0 {
		duration = v
	}

	return h.debugHandler.ProfileCPU(duration)
}

func (h *ProfileCPUHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *ProfileCPUHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	if len(word) > 1 && word[0] == '-' {
		return []string{"--duration", "--d", "--help"}, nil
	}
	return nil, nil
}

func (h *ProfileCPUHandler) Help() string {
	return `分析CPU使用情况。

用法:
  o2ochat debug profile cpu [选项]

选项:
  -d, --duration    采样时长(秒，默认: 30)
  -h, --help       显示帮助信息

示例:
  o2ochat debug profile cpu
  o2ochat debug profile cpu --duration 60`
}

type TraceCallHandler struct {
	debugHandler *DebugHandler
}

func NewTraceCallHandler(logPath string) *TraceCallHandler {
	return &TraceCallHandler{
		debugHandler: NewDebugHandler(logPath, ""),
	}
}

func (h *TraceCallHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	function := ""

	if v, ok := ctx.Flags["function"].(string); ok {
		function = v
	}
	if v, ok := ctx.Flags["f"].(string); ok {
		function = v
	}

	return h.debugHandler.TraceCall(function)
}

func (h *TraceCallHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *TraceCallHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	if len(word) > 1 && word[0] == '-' {
		return []string{"--function", "--f", "--help"}, nil
	}
	return nil, nil
}

func (h *TraceCallHandler) Help() string {
	return `跟踪函数调用。

用法:
  o2ochat debug trace [选项]

选项:
  -f, --function    函数名
  -h, --help       显示帮助信息

示例:
  o2ochat debug trace
  o2ochat debug trace --function SendMessage`
}

type StressTestHandler struct {
	debugHandler *DebugHandler
}

func NewStressTestHandler(logPath string) *StressTestHandler {
	return &StressTestHandler{
		debugHandler: NewDebugHandler(logPath, ""),
	}
}

func (h *StressTestHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	target := "cpu"
	duration := 60

	if v, ok := ctx.Flags["target"].(string); ok && v != "" {
		target = v
	}
	if v, ok := ctx.Flags["t"].(string); ok && v != "" {
		target = v
	}
	if v, ok := ctx.Flags["duration"].(int); ok && v > 0 {
		duration = v
	}
	if v, ok := ctx.Flags["d"].(int); ok && v > 0 {
		duration = v
	}

	return h.debugHandler.StressTest(target, duration)
}

func (h *StressTestHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *StressTestHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	if len(word) > 1 && word[0] == '-' {
		return []string{"--target", "--t", "--duration", "--d", "--help"}, nil
	}
	return nil, nil
}

func (h *StressTestHandler) Help() string {
	return `压力测试。

用法:
  o2ochat debug stress [选项]

选项:
  -t, --target      测试目标 (cpu, memory, io)
  -d, --duration   测试时长(秒，默认: 60)
  -h, --help       显示帮助信息

示例:
  o2ochat debug stress
  o2ochat debug stress --target cpu --duration 120`
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Source    string `json:"source"`
}

func parseLogLine(line string) *LogEntry {
	re := regexp.MustCompile(`^\[(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})\]\s+\[(\w+)\]\s+(.+)$`)
	matches := re.FindStringSubmatch(line)

	if matches == nil {
		return &LogEntry{
			Message: line,
		}
	}

	return &LogEntry{
		Timestamp: matches[1],
		Level:     matches[2],
		Message:   matches[3],
	}
}

func formatLogsJSON(logs []string) (string, error) {
	var entries []LogEntry
	for _, line := range logs {
		entries = append(entries, *parseLogLine(line))
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func formatLogsText(logs []string) string {
	var buf bytes.Buffer
	for _, line := range logs {
		buf.WriteString(line + "\n")
	}
	return buf.String()
}
