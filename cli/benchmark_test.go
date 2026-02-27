package cli

import (
	"testing"
	"time"
)

func BenchmarkCLIManagerInitialize(b *testing.B) {
	manager := NewCLIManager()
	config := DefaultCLIConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.Initialize(config)
	}
}

func BenchmarkCLIManagerExecute(b *testing.B) {
	manager := NewCLIManager()
	config := &CLIConfig{
		Timeout: 30,
	}
	manager.Initialize(config)

	startCmd := &CommandConfig{
		Name:        "start",
		Description: "Start application",
		Usage:       "o2ochat start",
		Flags: []Flag{
			{
				Name:        "daemon",
				Short:       "d",
				Description: "Run as daemon",
				Type:        "bool",
				Default:     false,
			},
		},
	}

	handler := &MockCommandHandler{}
	manager.RegisterCommand(startCmd, handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.Execute([]string{"start", "--daemon"})
	}
}

func BenchmarkCommandParsing(b *testing.B) {
	manager := NewCLIManager()
	manager.Initialize(&CLIConfig{Timeout: 30})

	args := []string{"start", "--daemon", "--config", "/path/to/config.yaml", "--timeout", "60"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.Execute(args)
	}
}

func BenchmarkFlagParsing(b *testing.B) {
	flags := []Flag{
		{Name: "daemon", Short: "d", Type: "bool", Default: false},
		{Name: "config", Short: "c", Type: "string", Default: ""},
		{Name: "timeout", Short: "t", Type: "int", Default: 30},
		{Name: "verbose", Short: "v", Type: "bool", Default: false},
		{Name: "output", Short: "o", Type: "string", Default: ""},
	}

	args := []string{"--daemon", "--config", "/path/to/config", "--timeout", "60", "--verbose", "--output", "file.txt"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseFlags(args, flags)
	}
}

func parseFlags(args []string, flagDefs []Flag) map[string]interface{} {
	flags := make(map[string]interface{})

	for _, fd := range flagDefs {
		flags[fd.Name] = fd.Default
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if len(arg) == 0 || arg[0] != '-' {
			continue
		}

		name := arg[1:]
		var flagDef *Flag

		for j := range flagDefs {
			if flagDefs[j].Name == name || flagDefs[j].Short == name {
				flagDef = &flagDefs[j]
				break
			}
		}

		if flagDef == nil {
			continue
		}

		if flagDef.Type == "bool" {
			flags[flagDef.Name] = true
		} else if i+1 < len(args) {
			i++
			flags[flagDef.Name] = args[i]
		}
	}

	return flags
}

func BenchmarkConfigHandlerShowConfig(b *testing.B) {
	handler := NewConfigHandler("/tmp/test_config.json")
	handler.configData = getDefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ShowConfig("network")
	}
}

func BenchmarkDataHandlerBackup(b *testing.B) {
	handler := NewDataHandler("/tmp/test_data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.BackupData("")
	}
}

func BenchmarkNetworkHandlerDiagnose(b *testing.B) {
	handler := NewNetworkHandler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.DiagnoseNetwork()
	}
}

func BenchmarkDebugHandlerShowLogs(b *testing.B) {
	handler := NewDebugHandler("/tmp/test_logs", "/tmp/test_debug")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ShowLogs("info", 100)
	}
}

func BenchmarkPerformanceOptimizerRecord(b *testing.B) {
	optimizer := NewPerformanceOptimizer(1000)
	optimizer.Start()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		optimizer.RecordCommand("test_command", time.Millisecond*50, nil)
	}
}

func BenchmarkCommandCache(b *testing.B) {
	cache := NewCommandCache(100, 5*time.Minute)
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key1")
		cache.Get("key2")
		cache.Get("key3")
	}
}

func BenchmarkRateLimiter(b *testing.B) {
	limiter := NewRateLimiter(100, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow()
	}
}

func BenchmarkOutputBuffer(b *testing.B) {
	buf := NewOutputBuffer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.WriteFormat("test message %d: %s\n", i, "hello world")
	}
}

func BenchmarkTableWriter(b *testing.B) {
	tw := NewTableWriter([]string{"Name", "Status", "Duration", "Size"})
	tw.AddRow([]string{"test1", "running", "1s", "100KB"})
	tw.AddRow([]string{"test2", "stopped", "5s", "500KB"})
	tw.AddRow([]string{"test3", "running", "10s", "1MB"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tw.Render()
	}
}

func BenchmarkScriptEngineExecute(b *testing.B) {
	engine := NewScriptEngine("/tmp/test_scripts")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.variables = make(map[string]string)
		engine.executeLine("set name=test")
		engine.executeLine("echo hello $name")
	}
}

func BenchmarkTaskSchedulerRunTask(b *testing.B) {
	scheduler := NewTaskScheduler()
	scheduler.AddTask("test_task", "/tmp/test.o2o", "@daily")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheduler.RunTask("test_task")
	}
}

type MockCommandHandler struct{}

func (h *MockCommandHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	return &CommandResult{
		Success:  true,
		Message:  "Command executed",
		ExitCode: 0,
	}, nil
}

func (h *MockCommandHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *MockCommandHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	return nil, nil
}

func (h *MockCommandHandler) Help() string {
	return "Mock command help"
}

func TestPerformanceOptimizer(t *testing.T) {
	optimizer := NewPerformanceOptimizer(100)
	optimizer.Start()

	optimizer.RecordCommand("test", time.Millisecond*100, nil)
	optimizer.RecordCommand("test", time.Millisecond*200, nil)
	optimizer.RecordCommand("test", time.Millisecond*150, nil)

	stats := optimizer.GetStats()
	if len(stats) == 0 {
		t.Error("expected stats to be populated")
	}

	top := optimizer.GetTopCommands(10)
	if len(top) == 0 {
		t.Error("expected top commands to be populated")
	}

	optimizer.ResetStats()
	stats = optimizer.GetStats()
	if len(stats) != 0 {
		t.Error("expected stats to be reset")
	}
}

func TestCommandCache(t *testing.T) {
	cache := NewCommandCache(10, time.Minute)

	cache.Set("key1", "value1")

	val, ok := cache.Get("key1")
	if !ok {
		t.Error("expected to get value1")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}

	if _, ok := cache.Get("nonexistent"); ok {
		t.Error("expected not to find nonexistent key")
	}

	cache.Delete("key1")
	if _, ok := cache.Get("key1"); ok {
		t.Error("expected key1 to be deleted")
	}

	stats := cache.GetStats()
	if stats["size"].(int) != 0 {
		t.Error("expected cache to be empty")
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(10, 100)

	allowed := 0
	for i := 0; i < 20; i++ {
		if limiter.Allow() {
			allowed++
		}
	}

	if allowed > 10 {
		t.Errorf("expected at most 10 allowed, got %d", allowed)
	}

	limiter.Wait()

	stats := limiter.GetStats()
	if stats["tokens"] == nil {
		t.Error("expected tokens in stats")
	}
}

func TestOutputBuffer(t *testing.T) {
	buf := NewOutputBuffer()

	n, err := buf.WriteString("test")
	if err != nil || n != 4 {
		t.Error("expected to write 4 bytes")
	}

	n, err = buf.WriteFormat(" number %d", 42)
	if err != nil {
		t.Error("expected no error")
	}

	if buf.Len() != 14 {
		t.Errorf("expected length 14, got %d", buf.Len())
	}

	if buf.String() != "test number 42" {
		t.Error("unexpected buffer content")
	}

	buf.Reset()
	if buf.Len() != 0 {
		t.Error("expected buffer to be reset")
	}
}

func TestTableWriter(t *testing.T) {
	tw := NewTableWriter([]string{"Name", "Age"})
	tw.AddRow([]string{"Alice", "30"})
	tw.AddRow([]string{"Bob", "25"})

	if len(tw.rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(tw.rows))
	}

	if tw.colWidths[0] != 5 {
		t.Errorf("expected colWidths[0]=5, got %d", tw.colWidths[0])
	}
	if tw.colWidths[1] != 2 {
		t.Errorf("expected colWidths[1]=2, got %d", tw.colWidths[1])
	}
}

func TestProgressBar(t *testing.T) {
	pb := NewProgressBar(nil, 100)

	pb.Add(25)
	if pb.current != 25 {
		t.Errorf("expected current=25, got %d", pb.current)
	}

	pb.Set(50)
	if pb.current != 50 {
		t.Errorf("expected current=50, got %d", pb.current)
	}

	pb.Finish()
	if pb.current != 100 {
		t.Errorf("expected current=100 after Finish, got %d", pb.current)
	}
}

func TestSystemInfo(t *testing.T) {
	info := GetSystemInfo()

	if info.GoVersion == "" {
		t.Error("expected GoVersion to be set")
	}

	if info.NumCPU == 0 {
		t.Error("expected NumCPU to be set")
	}

	if info.NumGoroutine == 0 {
		t.Error("expected NumGoroutine to be set")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, test := range tests {
		result := FormatBytes(test.input)
		if result != test.expected {
			t.Errorf("FormatBytes(%d) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{time.Millisecond * 500, "500ms"},
		{time.Second * 5, "5.0s"},
		{time.Minute * 2, "2.0m"},
		{time.Hour * 1, "1.0h"},
	}

	for _, test := range tests {
		result := FormatDuration(test.input)
		if result != test.expected {
			t.Errorf("FormatDuration(%v) = %s; want %s", test.input, result, test.expected)
		}
	}
}
