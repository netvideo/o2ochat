# CLI Module - 命令行接口模块

## 功能概述
提供命令行界面，用于应用程序管理、调试、配置和自动化操作。

## 核心功能
1. **应用程序管理**：启动、停止、重启应用
2. **配置管理**：查看和修改配置
3. **调试工具**：网络诊断、性能监控
4. **数据管理**：导入、导出、清理数据
5. **自动化脚本**：批量操作和任务调度
6. **系统集成**：与其他工具集成

## 接口定义

### 类型定义
```go
// CLI命令配置
type CommandConfig struct {
    Name        string   `json:"name"`        // 命令名称
    Description string   `json:"description"` // 命令描述
    Usage       string   `json:"usage"`       // 使用说明
    Aliases     []string `json:"aliases"`     // 命令别名
    Flags       []Flag   `json:"flags"`       // 命令标志
    Subcommands []string `json:"subcommands"` // 子命令列表
    Hidden      bool     `json:"hidden"`      // 是否隐藏
}

// 命令标志
type Flag struct {
    Name        string      `json:"name"`        // 标志名称
    Short       string      `json:"short"`       // 短标志
    Description string      `json:"description"` // 标志描述
    Type        FlagType    `json:"type"`        // 标志类型
    Default     interface{} `json:"default"`     // 默认值
    Required    bool        `json:"required"`    // 是否必需
    Hidden      bool        `json:"hidden"`      // 是否隐藏
}

// 标志类型
type FlagType string

const (
    FlagTypeString  FlagType = "string"
    FlagTypeInt     FlagType = "int"
    FlagTypeBool    FlagType = "bool"
    FlagTypeFloat   FlagType = "float"
    FlagTypeDuration FlagType = "duration"
    FlagTypePath    FlagType = "path"
)

// 命令上下文
type CommandContext struct {
    Args    []string          // 命令行参数
    Flags   map[string]interface{} // 解析的标志
    Config  *CLIConfig        // CLI配置
    Output  io.Writer         // 输出流
    Error   io.Writer         // 错误流
    Input   io.Reader         // 输入流
}

// CLI配置
type CLIConfig struct {
    LogLevel    string `json:"log_level"`    // 日志级别
    LogFormat   string `json:"log_format"`   // 日志格式
    ColorOutput bool   `json:"color_output"` // 彩色输出
    Interactive bool   `json:"interactive"`  // 交互模式
    Timeout     int    `json:"timeout"`      // 命令超时（秒）
    ConfigPath  string `json:"config_path"`  // 配置文件路径
}

// 命令结果
type CommandResult struct {
    Success   bool        `json:"success"`   // 是否成功
    Message   string      `json:"message"`   // 结果消息
    Data      interface{} `json:"data"`      // 返回数据
    ExitCode  int         `json:"exit_code"` // 退出码
    Duration  time.Duration `json:"duration"` // 执行时间
}
```

### 主要接口
```go
// CLI管理器接口
type CLIManager interface {
    // 初始化CLI
    Initialize(config *CLIConfig) error
    
    // 注册命令
    RegisterCommand(cmd *CommandConfig, handler CommandHandler) error
    
    // 执行命令
    Execute(args []string) (*CommandResult, error)
    
    // 获取命令帮助
    GetHelp(command string) (string, error)
    
    // 获取命令列表
    GetCommands() ([]*CommandConfig, error)
    
    // 设置输出格式
    SetOutputFormat(format OutputFormat) error
    
    // 运行交互式Shell
    RunInteractive() error
    
    // 关闭CLI
    Close() error
}

// 命令处理器接口
type CommandHandler interface {
    // 执行命令
    Execute(ctx *CommandContext) (*CommandResult, error)
    
    // 验证参数
    Validate(ctx *CommandContext) error
    
    // 自动补全
    Autocomplete(ctx *CommandContext, word string) ([]string, error)
    
    // 命令帮助
    Help() string
}

// 配置管理器接口
type ConfigCLI interface {
    // 显示配置
    ShowConfig(section string) (*CommandResult, error)
    
    // 设置配置
    SetConfig(section, key, value string) (*CommandResult, error)
    
    // 重置配置
    ResetConfig() (*CommandResult, error)
    
    // 导入配置
    ImportConfig(path string) (*CommandResult, error)
    
    // 导出配置
    ExportConfig(path string) (*CommandResult, error)
    
    // 验证配置
    ValidateConfig() (*CommandResult, error)
}

// 网络诊断接口
type NetworkCLI interface {
    // 测试连接
    TestConnection(peerID string) (*CommandResult, error)
    
    // 诊断网络
    DiagnoseNetwork() (*CommandResult, error)
    
    // 查看连接状态
    ShowConnections() (*CommandResult, error)
    
    // 查看网络统计
    ShowNetworkStats() (*CommandResult, error)
    
    // 测试NAT穿透
    TestNATTraversal() (*CommandResult, error)
    
    // 查看路由表
    ShowRoutingTable() (*CommandResult, error)
}

// 数据管理接口
type DataCLI interface {
    // 备份数据
    BackupData(path string) (*CommandResult, error)
    
    // 恢复数据
    RestoreData(path string) (*CommandResult, error)
    
    // 清理数据
    CleanupData(olderThan string) (*CommandResult, error)
    
    // 导出消息
    ExportMessages(peerID, path string) (*CommandResult, error)
    
    // 导入消息
    ImportMessages(path string) (*CommandResult, error)
    
    // 查看存储统计
    ShowStorageStats() (*CommandResult, error)
}

// 调试工具接口
type DebugCLI interface {
    // 查看日志
    ShowLogs(level string, lines int) (*CommandResult, error)
    
    // 性能监控
    MonitorPerformance(interval int) (*CommandResult, error)
    
    // 内存分析
    ProfileMemory(duration int) (*CommandResult, error)
    
    // CPU分析
    ProfileCPU(duration int) (*CommandResult, error)
    
    // 跟踪调用
    TraceCall(function string) (*CommandResult, error)
    
    // 压力测试
    StressTest(target string, duration int) (*CommandResult, error)
}
```

## 实现要求

### 1. 命令解析
- 支持子命令和嵌套命令
- 支持标志解析和验证
- 支持环境变量和配置文件
- 支持命令自动补全

### 2. 输出格式
- 文本格式（默认）
- JSON格式（机器可读）
- YAML格式（可读性更好）
- 表格格式（数据展示）

### 3. 错误处理
- 友好的错误消息
- 详细的错误堆栈（调试模式）
- 建议的解决方案
- 一致的退出码

### 4. 交互模式
- 支持交互式Shell
- 命令历史记录
- 语法高亮
- 实时验证

## 测试要求

### 单元测试
```bash
# 运行CLI模块测试
go test ./cli -v

# 测试特定功能
go test ./cli -run TestCommandParser
go test ./cli -run TestCommandExecution
go test ./cli -run TestOutputFormat
```

### 集成测试
```bash
# 测试完整CLI流程
go test ./cli -tags=integration

# 测试交互模式
go test ./cli -tags=interactive
```

### 测试用例
1. **命令解析测试**：测试参数和标志解析
2. **命令执行测试**：测试命令执行结果
3. **错误处理测试**：测试各种错误场景
4. **输出格式测试**：测试不同输出格式
5. **交互模式测试**：测试交互式功能

## 依赖关系
- 所有功能模块：用于命令实现
- storage模块：用于配置管理

## 使用示例

```go
// 创建CLI管理器
config := &CLIConfig{
    LogLevel:    "info",
    LogFormat:   "text",
    ColorOutput: true,
    Interactive: false,
    Timeout:     30,
    ConfigPath:  "~/.o2ochat/cli.yaml",
}

cliManager, err := NewCLIManager()
err = cliManager.Initialize(config)

// 注册命令
startCmd := &CommandConfig{
    Name:        "start",
    Description: "启动应用程序",
    Usage:       "o2ochat start [options]",
    Flags: []Flag{
        {
            Name:        "daemon",
            Short:       "d",
            Description: "以守护进程方式运行",
            Type:        FlagTypeBool,
            Default:     false,
        },
        {
            Name:        "config",
            Short:       "c",
            Description: "配置文件路径",
            Type:        FlagTypePath,
            Default:     "",
        },
    },
}

err = cliManager.RegisterCommand(startCmd, &StartCommandHandler{})

// 执行命令
result, err := cliManager.Execute([]string{"start", "--daemon", "--config", "/path/to/config.yaml"})

if result.Success {
    fmt.Println("命令执行成功:", result.Message)
} else {
    fmt.Println("命令执行失败:", result.Message)
    os.Exit(result.ExitCode)
}

// 获取命令帮助
helpText, err := cliManager.GetHelp("start")
fmt.Println(helpText)

// 运行交互式Shell
if config.Interactive {
    err = cliManager.RunInteractive()
}

// 关闭CLI
err = cliManager.Close()
```

## 命令示例

```bash
# 启动应用程序
o2ochat start --daemon

# 查看配置
o2ochat config show
o2ochat config show --section network

# 设置配置
o2ochat config set network.max_connections 100
o2ochat config set ui.theme dark

# 网络诊断
o2ochat network test --peer QmPeer123
o2ochat network diagnose

# 查看连接状态
o2ochat connections list
o2ochat connections show QmPeer123

# 数据管理
o2ochat data backup --path ./backup.zip
o2ochat data restore --path ./backup.zip
o2ochat data cleanup --older-than 30d

# 消息管理
o2ochat messages export --peer QmPeer123 --path ./messages.json
o2ochat messages import --path ./messages.json

# 调试工具
o2ochat debug logs --level debug --lines 100
o2ochat debug profile --cpu --duration 30
o2ochat debug monitor --interval 5

# 文件传输
o2ochat file send --peer QmPeer123 --path ./document.pdf
o2ochat file receive --task-id task123 --path ./downloads/

# 联系人管理
o2ochat contacts list
o2ochat contacts add --peer QmPeer123 --name "张三"
o2ochat contacts remove --peer QmPeer123

# 系统信息
o2ochat system info
o2ochat system stats
o2ochat system version

# 帮助信息
o2ochat help
o2ochat help start
o2ochat help config
```

## 命令处理器示例

```go
// 启动命令处理器
type StartCommandHandler struct{}

func (h *StartCommandHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
    startTime := time.Now()
    
    // 解析标志
    daemonMode := ctx.Flags["daemon"].(bool)
    configPath := ctx.Flags["config"].(string)
    
    // 执行启动逻辑
    err := startApplication(daemonMode, configPath)
    if err != nil {
        return &CommandResult{
            Success:  false,
            Message:  fmt.Sprintf("启动失败: %v", err),
            ExitCode: 1,
            Duration: time.Since(startTime),
        }, nil
    }
    
    return &CommandResult{
        Success:  true,
        Message:  "应用程序启动成功",
        ExitCode: 0,
        Duration: time.Since(startTime),
    }, nil
}

func (h *StartCommandHandler) Validate(ctx *CommandContext) error {
    // 验证配置文件是否存在
    if configPath, ok := ctx.Flags["config"].(string); ok && configPath != "" {
        if _, err := os.Stat(configPath); os.IsNotExist(err) {
            return fmt.Errorf("配置文件不存在: %s", configPath)
        }
    }
    return nil
}

func (h *StartCommandHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
    // 提供自动补全建议
    if strings.HasPrefix(word, "--") {
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
  -c, --config string  配置文件路径
  -h, --help       显示帮助信息

示例:
  o2ochat start --daemon
  o2ochat start --config /etc/o2ochat/config.yaml`
}
```

## 输出格式示例

```json
// JSON格式输出
{
  "success": true,
  "message": "命令执行成功",
  "data": {
    "connections": [
      {
        "peer_id": "QmPeer123",
        "type": "quic",
        "state": "connected",
        "duration": "5m30s",
        "bytes_sent": 1024000,
        "bytes_received": 512000
      }
    ],
    "total_connections": 1
  },
  "exit_code": 0,
  "duration": "123.456ms"
}

// 表格格式输出
+------------+--------+-----------+----------+------------+---------------+
| PEER ID    | TYPE   | STATE     | DURATION | BYTES SENT | BYTES RECEIVED|
+------------+--------+-----------+----------+------------+---------------+
| QmPeer123  | quic   | connected | 5m30s    | 1.00 MB    | 512.00 KB     |
| QmPeer456  | webrtc | connecting| 10s      | 0 B        | 0 B           |
+------------+--------+-----------+----------+------------+---------------+
Total: 2 connections
```

## 错误处理
- 无效命令必须显示帮助信息
- 参数错误必须提供具体建议
- 执行失败必须返回适当的退出码
- 资源不足必须清理临时文件

## 安全注意事项
1. 敏感信息（密码、密钥）不在命令行显示
2. 配置文件权限限制
3. 命令执行权限验证
4. 防止命令注入攻击

## 扩展性
- 支持插件系统添加自定义命令
- 支持脚本自动化
- 支持远程命令执行（可选）
- 支持API接口调用