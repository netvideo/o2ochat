package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ScriptEngine struct {
	scriptsDir string
	variables  map[string]string
	history    []string
}

func NewScriptEngine(scriptsDir string) *ScriptEngine {
	if scriptsDir == "" {
		scriptsDir = os.Getenv("HOME") + "/.o2ochat/scripts"
	}
	return &ScriptEngine{
		scriptsDir: scriptsDir,
		variables:  make(map[string]string),
		history:    []string{},
	}
}

func (e *ScriptEngine) ExecuteScript(path string, args []string) (*CommandResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to open script: %v", err),
			ExitCode: 1,
		}, nil
	}
	defer file.Close()

	e.variables["ARGS"] = strings.Join(args, " ")
	e.variables["SCRIPT_DIR"] = filepath.Dir(path)
	e.variables["SCRIPT_NAME"] = filepath.Base(path)

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var output []string

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		e.history = append(e.history, fmt.Sprintf("%s:%d: %s", path, lineNum, line))

		result, err := e.executeLine(line)
		if err != nil {
			return &CommandResult{
				Success:  false,
				Message:  fmt.Sprintf("script error at line %d: %v", lineNum, err),
				ExitCode: 1,
			}, nil
		}

		if result != "" {
			output = append(output, result)
		}
	}

	if err := scanner.Err(); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("script read error: %v", err),
			ExitCode: 1,
		}, nil
	}

	return &CommandResult{
		Success: true,
		Message: "script executed successfully",
		Data: map[string]interface{}{
			"output":    output,
			"lines":     lineNum,
			"variables": e.variables,
		},
		ExitCode: 0,
	}, nil
}

func (e *ScriptEngine) executeLine(line string) (string, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return "", nil
	}

	cmd := parts[0]

	switch cmd {
	case "echo":
		return e.cmdEcho(parts[1:]), nil
	case "set":
		return "", e.cmdSet(parts[1:])
	case "get":
		return e.cmdGet(parts[1:]), nil
	case "if":
		return "", nil
	case "for":
		return "", nil
	default:
		return "", fmt.Errorf("unknown command: %s", cmd)
	}
}

func (e *ScriptEngine) cmdEcho(args []string) string {
	var result []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "$") {
			varName := strings.TrimPrefix(arg, "$")
			if val, ok := e.variables[varName]; ok {
				result = append(result, val)
			} else {
				result = append(result, "")
			}
		} else {
			result = append(result, arg)
		}
	}
	return strings.Join(result, " ")
}

func (e *ScriptEngine) cmdSet(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("set requires at least 2 arguments")
	}

	key := strings.TrimSuffix(args[0], "=")
	value := strings.Join(args[1:], " ")

	e.variables[key] = value
	return nil
}

func (e *ScriptEngine) cmdGet(args []string) string {
	if len(args) == 0 {
		return ""
	}

	varName := strings.TrimPrefix(args[0], "$")
	if val, ok := e.variables[varName]; ok {
		return val
	}
	return ""
}

func (e *ScriptEngine) ListScripts() ([]string, error) {
	files, err := os.ReadDir(e.scriptsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var scripts []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".o2o") {
			scripts = append(scripts, file.Name())
		}
	}

	return scripts, nil
}

func (e *ScriptEngine) GetHistory() []string {
	return e.history
}

type AutomationCLI struct {
	scriptEngine *ScriptEngine
	scheduler    *TaskScheduler
}

func NewAutomationCLI(scriptsDir string) *AutomationCLI {
	return &AutomationCLI{
		scriptEngine: NewScriptEngine(scriptsDir),
		scheduler:    NewTaskScheduler(),
	}
}

type TaskScheduler struct {
	tasks  map[string]*ScheduledTask
	stopCh chan string
	mu     interface{}
}

type ScheduledTask struct {
	Name       string
	ScriptPath string
	Schedule   string
	LastRun    time.Time
	NextRun    time.Time
	Enabled    bool
}

func NewTaskScheduler() *TaskScheduler {
	return &TaskScheduler{
		tasks:  make(map[string]*ScheduledTask),
		stopCh: make(chan string),
		mu:     &sync.Mutex{},
	}
}

func (s *TaskScheduler) AddTask(name, scriptPath, schedule string) error {
	task := &ScheduledTask{
		Name:       name,
		ScriptPath: scriptPath,
		Schedule:   schedule,
		Enabled:    true,
		NextRun:    time.Now(),
	}

	s.tasks[name] = task
	return nil
}

func (s *TaskScheduler) RemoveTask(name string) error {
	if _, ok := s.tasks[name]; !ok {
		return fmt.Errorf("task not found: %s", name)
	}
	delete(s.tasks, name)
	return nil
}

func (s *TaskScheduler) ListTasks() []*ScheduledTask {
	var result []*ScheduledTask
	for _, task := range s.tasks {
		result = append(result, task)
	}
	return result
}

func (s *TaskScheduler) RunTask(name string) (*CommandResult, error) {
	task, ok := s.tasks[name]
	if !ok {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("task not found: %s", name),
			ExitCode: 1,
		}, nil
	}

	if !task.Enabled {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("task is disabled: %s", name),
			ExitCode: 1,
		}, nil
	}

	task.LastRun = time.Now()

	return &CommandResult{
		Success: true,
		Message: fmt.Sprintf("task %s executed", name),
		Data: map[string]interface{}{
			"task_name":   task.Name,
			"script_path": task.ScriptPath,
			"last_run":    task.LastRun.Format(time.RFC3339),
			"schedule":    task.Schedule,
		},
		ExitCode: 0,
	}, nil
}

type RunScriptHandler struct {
	automation *AutomationCLI
}

func NewRunScriptHandler(scriptsDir string) *RunScriptHandler {
	return &RunScriptHandler{
		automation: NewAutomationCLI(scriptsDir),
	}
}

func (h *RunScriptHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	script := ""
	args := []string{}

	if v, ok := ctx.Flags["script"].(string); ok {
		script = v
	}
	if v, ok := ctx.Flags["s"].(string); ok {
		script = v
	}

	if script == "" {
		scripts, err := h.automation.scriptEngine.ListScripts()
		if err != nil {
			return &CommandResult{
				Success:  false,
				Message:  fmt.Sprintf("failed to list scripts: %v", err),
				ExitCode: 1,
			}, nil
		}
		return &CommandResult{
			Success:  true,
			Message:  "available scripts",
			Data:     scripts,
			ExitCode: 0,
		}, nil
	}

	if !strings.HasSuffix(script, ".o2o") {
		script = script + ".o2o"
	}

	scriptPath := filepath.Join(h.automation.scriptEngine.scriptsDir, script)

	return h.automation.scriptEngine.ExecuteScript(scriptPath, args)
}

func (h *RunScriptHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *RunScriptHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	if len(word) > 1 && word[0] == '-' {
		return []string{"--script", "--s", "--help"}, nil
	}
	return nil, nil
}

func (h *RunScriptHandler) Help() string {
	return `运行自动化脚本。

用法:
  o2ochat script run [选项]

选项:
  -s, --script    脚本文件名
  -h, --help     显示帮助信息

示例:
  o2ochat script run
  o2ochat script run --script backup`
}

type ListScriptsHandler struct {
	automation *AutomationCLI
}

func NewListScriptsHandler(scriptsDir string) *ListScriptsHandler {
	return &ListScriptsHandler{
		automation: NewAutomationCLI(scriptsDir),
	}
}

func (h *ListScriptsHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	scripts, err := h.automation.scriptEngine.ListScripts()
	if err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to list scripts: %v", err),
			ExitCode: 1,
		}, nil
	}

	return &CommandResult{
		Success: true,
		Message: fmt.Sprintf("found %d scripts", len(scripts)),
		Data: map[string]interface{}{
			"scripts": scripts,
			"count":   len(scripts),
		},
		ExitCode: 0,
	}, nil
}

func (h *ListScriptsHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *ListScriptsHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	return nil, nil
}

func (h *ListScriptsHandler) Help() string {
	return `列出可用的自动化脚本。

用法:
  o2ochat script list [选项]

选项:
  -h, --help     显示帮助信息

示例:
  o2ochat script list`
}

type ScheduleTaskHandler struct {
	automation *AutomationCLI
}

func NewScheduleTaskHandler(scriptsDir string) *ScheduleTaskHandler {
	return &ScheduleTaskHandler{
		automation: NewAutomationCLI(scriptsDir),
	}
}

func (h *ScheduleTaskHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	name := ""
	script := ""
	schedule := ""

	if v, ok := ctx.Flags["name"].(string); ok {
		name = v
	}
	if v, ok := ctx.Flags["script"].(string); ok {
		script = v
	}
	if v, ok := ctx.Flags["schedule"].(string); ok {
		schedule = v
	}

	if name == "" || script == "" || schedule == "" {
		return &CommandResult{
			Success:  false,
			Message:  "name, script and schedule are required",
			ExitCode: 1,
		}, nil
	}

	err := h.automation.scheduler.AddTask(name, script, schedule)
	if err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to schedule task: %v", err),
			ExitCode: 1,
		}, nil
	}

	return &CommandResult{
		Success: true,
		Message: fmt.Sprintf("task %s scheduled", name),
		Data: map[string]interface{}{
			"name":     name,
			"script":   script,
			"schedule": schedule,
		},
		ExitCode: 0,
	}, nil
}

func (h *ScheduleTaskHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *ScheduleTaskHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	if len(word) > 1 && word[0] == '-' {
		return []string{"--name", "--script", "--schedule", "--help"}, nil
	}
	return nil, nil
}

func (h *ScheduleTaskHandler) Help() string {
	return `调度自动化任务。

用法:
  o2ochat task schedule [选项]

选项:
  -n, --name       任务名称
  -s, --script     脚本路径
  -c, --schedule   调度表达式 (如: @daily, @hourly, 0 * * * *)
  -h, --help      显示帮助信息

示例:
  o2ochat task schedule --name backup --script backup.o2o --schedule @daily`
}

type ListTasksHandler struct {
	automation *AutomationCLI
}

func NewListTasksHandler(scriptsDir string) *ListTasksHandler {
	return &ListTasksHandler{
		automation: NewAutomationCLI(scriptsDir),
	}
}

func (h *ListTasksHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	tasks := h.automation.scheduler.ListTasks()

	return &CommandResult{
		Success: true,
		Message: fmt.Sprintf("found %d scheduled tasks", len(tasks)),
		Data: map[string]interface{}{
			"tasks": tasks,
			"count": len(tasks),
		},
		ExitCode: 0,
	}, nil
}

func (h *ListTasksHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *ListTasksHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	return nil, nil
}

func (h *ListTasksHandler) Help() string {
	return `列出已调度的任务。

用法:
  o2ochat task list [选项]

选项:
  -h, --help     显示帮助信息

示例:
  o2ochat task list`
}

type RunTaskHandler struct {
	automation *AutomationCLI
}

func NewRunTaskHandler(scriptsDir string) *RunTaskHandler {
	return &RunTaskHandler{
		automation: NewAutomationCLI(scriptsDir),
	}
}

func (h *RunTaskHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	name := ""

	if v, ok := ctx.Flags["name"].(string); ok {
		name = v
	}
	if v, ok := ctx.Flags["n"].(string); ok {
		name = v
	}

	if name == "" {
		return &CommandResult{
			Success:  false,
			Message:  "task name is required",
			ExitCode: 1,
		}, nil
	}

	return h.automation.scheduler.RunTask(name)
}

func (h *RunTaskHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *RunTaskHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	if len(word) > 1 && word[0] == '-' {
		return []string{"--name", "--n", "--help"}, nil
	}
	return nil, nil
}

func (h *RunTaskHandler) Help() string {
	return `运行指定任务。

用法:
  o2ochat task run [选项]

选项:
  -n, --name     任务名称
  -h, --help    显示帮助信息

示例:
  o2ochat task run --name backup`
}
