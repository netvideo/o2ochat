package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type AppHandler struct {
	appPath   string
	dataPath  string
	pidFile   string
	isRunning bool
	version   string
	buildTime string
}

func NewAppHandler(appPath, dataPath string) *AppHandler {
	return &AppHandler{
		appPath:   appPath,
		dataPath:  dataPath,
		pidFile:   dataPath + "/app.pid",
		isRunning: false,
		version:   "1.0.0",
		buildTime: time.Now().Format("2006-01-02"),
	}
}

func (h *AppHandler) SetVersion(version, buildTime string) {
	h.version = version
	h.buildTime = buildTime
}

func (h *AppHandler) Start(daemon bool, configPath string) (*CommandResult, error) {
	if h.isRunning {
		return &CommandResult{
			Success:  false,
			Message:  "application is already running",
			ExitCode: 1,
		}, nil
	}

	if h.appPath == "" {
		h.appPath = os.Args[0]
	}

	if configPath == "" {
		configPath = h.dataPath + "/config.yaml"
	}

	args := []string{"start"}
	if configPath != "" {
		args = append(args, "--config", configPath)
	}
	if daemon {
		args = append(args, "--daemon")
	}

	cmd := exec.Command(h.appPath, args...)
	cmd.Env = append(os.Environ(),
		"O2OCHAT_DATA_PATH="+h.dataPath,
		"O2OCHAT_CONFIG_PATH="+configPath,
	)

	if daemon {
		cmd.Dir = "/"

		if err := cmd.Start(); err != nil {
			return &CommandResult{
				Success:  false,
				Message:  fmt.Sprintf("failed to start application: %v", err),
				ExitCode: 1,
			}, nil
		}

		pid := cmd.Process.Pid
		if err := os.WriteFile(h.pidFile, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
			return &CommandResult{
				Success:  false,
				Message:  fmt.Sprintf("failed to write pid file: %v", err),
				ExitCode: 1,
			}, nil
		}

		h.isRunning = true

		return &CommandResult{
			Success:  true,
			Message:  fmt.Sprintf("application started in daemon mode (PID: %d)", pid),
			ExitCode: 0,
		}, nil
	}

	if err := cmd.Run(); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to start application: %v", err),
			ExitCode: 1,
		}, nil
	}

	return &CommandResult{
		Success:  true,
		Message:  "application started successfully",
		ExitCode: 0,
	}, nil
}

func (h *AppHandler) Stop() (*CommandResult, error) {
	if !h.isRunning {
		pid, err := h.readPID()
		if err != nil {
			return &CommandResult{
				Success:  false,
				Message:  "application is not running",
				ExitCode: 1,
			}, nil
		}

		proc, err := os.FindProcess(pid)
		if err != nil {
			return &CommandResult{
				Success:  false,
				Message:  "application is not running",
				ExitCode: 1,
			}, nil
		}

		if err := proc.Kill(); err != nil {
			return &CommandResult{
				Success:  false,
				Message:  fmt.Sprintf("failed to stop application: %v", err),
				ExitCode: 1,
			}, nil
		}

		os.Remove(h.pidFile)
		h.isRunning = false

		return &CommandResult{
			Success:  true,
			Message:  "application stopped successfully",
			ExitCode: 0,
		}, nil
	}

	h.isRunning = false

	return &CommandResult{
		Success:  true,
		Message:  "application stopped",
		ExitCode: 0,
	}, nil
}

func (h *AppHandler) Restart() (*CommandResult, error) {
	result, err := h.Stop()
	if err != nil || !result.Success {
		return result, err
	}

	time.Sleep(500 * time.Millisecond)

	result, err = h.Start(false, "")
	if err != nil || !result.Success {
		return result, err
	}

	return &CommandResult{
		Success:  true,
		Message:  "application restarted successfully",
		ExitCode: 0,
	}, nil
}

func (h *AppHandler) Status() (*CommandResult, error) {
	status := map[string]interface{}{
		"running":    h.isRunning,
		"pid_file":   h.pidFile,
		"data_path":  h.dataPath,
		"app_path":   h.appPath,
		"platform":   runtime.GOOS + "/" + runtime.GOARCH,
		"version":    h.version,
		"build_time": h.buildTime,
	}

	if !h.isRunning {
		pid, err := h.readPID()
		if err == nil {
			status["pid"] = pid
			status["running"] = true
			h.isRunning = true
		}
	}

	if h.isRunning {
		return &CommandResult{
			Success:  true,
			Message:  "application is running",
			Data:     status,
			ExitCode: 0,
		}, nil
	}

	return &CommandResult{
		Success:  true,
		Message:  "application is not running",
		Data:     status,
		ExitCode: 3,
	}, nil
}

func (h *AppHandler) Version() (*CommandResult, error) {
	return &CommandResult{
		Success: true,
		Message: fmt.Sprintf("O2OChat v%s (%s)", h.version, h.buildTime),
		Data: map[string]string{
			"version":    h.version,
			"build_time": h.buildTime,
			"go_version": runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
		},
		ExitCode: 0,
	}, nil
}

func (h *AppHandler) readPID() (int, error) {
	data, err := os.ReadFile(h.pidFile)
	if err != nil {
		return 0, err
	}

	var pid int
	_, err = fmt.Sscanf(string(data), "%d", &pid)
	return pid, err
}

type StartCommandHandler struct {
	appHandler *AppHandler
}

func NewStartCommandHandler(appPath, dataPath string) *StartCommandHandler {
	return &StartCommandHandler{
		appHandler: NewAppHandler(appPath, dataPath),
	}
}

func (h *StartCommandHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	daemon := false
	configPath := ""

	if v, ok := ctx.Flags["daemon"].(bool); ok {
		daemon = v
	}
	if v, ok := ctx.Flags["config"].(string); ok {
		configPath = v
	}

	return h.appHandler.Start(daemon, configPath)
}

func (h *StartCommandHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *StartCommandHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	if len(word) > 1 && word[0] == '-' {
		return []string{"--daemon", "--config", "--help"}, nil
	}
	return nil, nil
}

func (h *StartCommandHandler) Help() string {
	return `启动O2OChat应用程序。

用法:
  o2ochat start [选项]

选项:
  -d, --daemon     以守护进程方式运行
  -c, --config     配置文件路径
  -h, --help       显示帮助信息

示例:
  o2ochat start
  o2ochat start --daemon
  o2ochat start --config /etc/o2ochat/config.yaml`
}

type StopCommandHandler struct {
	appHandler *AppHandler
}

func NewStopCommandHandler(appPath, dataPath string) *StopCommandHandler {
	return &StopCommandHandler{
		appHandler: NewAppHandler(appPath, dataPath),
	}
}

func (h *StopCommandHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	return h.appHandler.Stop()
}

func (h *StopCommandHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *StopCommandHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	return nil, nil
}

func (h *StopCommandHandler) Help() string {
	return `停止O2OChat应用程序。

用法:
  o2ochat stop [选项]

选项:
  -h, --help       显示帮助信息

示例:
  o2ochat stop`
}

type StatusCommandHandler struct {
	appHandler *AppHandler
}

func NewStatusCommandHandler(appPath, dataPath string) *StatusCommandHandler {
	return &StatusCommandHandler{
		appHandler: NewAppHandler(appPath, dataPath),
	}
}

func (h *StatusCommandHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	return h.appHandler.Status()
}

func (h *StatusCommandHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *StatusCommandHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	return nil, nil
}

func (h *StatusCommandHandler) Help() string {
	return `查看O2OChat应用程序状态。

用法:
  o2ochat status [选项]

选项:
  -h, --help       显示帮助信息

示例:
  o2ochat status`
}

type VersionCommandHandler struct {
	appHandler *AppHandler
}

func NewVersionCommandHandler(appPath, dataPath string) *VersionCommandHandler {
	h := NewAppHandler(appPath, dataPath)
	return &VersionCommandHandler{appHandler: h}
}

func (h *VersionCommandHandler) SetVersion(version, buildTime string) {
	h.appHandler.SetVersion(version, buildTime)
}

func (h *VersionCommandHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	return h.appHandler.Version()
}

func (h *VersionCommandHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *VersionCommandHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	return nil, nil
}

func (h *VersionCommandHandler) Help() string {
	return `显示O2OChat应用程序版本信息。

用法:
  o2ochat version [选项]

选项:
  -h, --help       显示帮助信息

示例:
  o2ochat version`
}
