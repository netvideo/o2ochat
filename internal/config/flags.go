package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// CommandLineFlags 命令行参数结构
type CommandLineFlags struct {
	ConfigPath string
	DataDir    string
	Debug      bool
	Version    bool
	Help       bool
	LogLevel   string
	LogOutput  string
	ServerAddr string
	ServerPort int
}

// ParseFlags 解析命令行参数
func ParseFlags() *CommandLineFlags {
	flags := &CommandLineFlags{}

	// 定义命令行参数
	flag.StringVar(&flags.ConfigPath, "config", "./config.json", "配置文件路径")
	flag.StringVar(&flags.ConfigPath, "c", "./config.json", "配置文件路径(简写)")

	flag.StringVar(&flags.DataDir, "data", "./data", "数据目录路径")
	flag.StringVar(&flags.DataDir, "d", "./data", "数据目录路径(简写)")

	flag.BoolVar(&flags.Debug, "debug", false, "启用调试模式")
	flag.BoolVar(&flags.Debug, "D", false, "启用调试模式(简写)")

	flag.StringVar(&flags.LogLevel, "log-level", "info", "日志级别: debug, info, warn, error")
	flag.StringVar(&flags.LogOutput, "log-output", "file", "日志输出: stdout, file, both")

	flag.StringVar(&flags.ServerAddr, "server", "", "信令服务器地址")
	flag.IntVar(&flags.ServerPort, "port", 0, "信令服务器端口")

	flag.BoolVar(&flags.Version, "version", false, "显示版本信息")
	flag.BoolVar(&flags.Version, "v", false, "显示版本信息(简写)")

	// 自定义帮助信息
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "O2OChat - P2P 即时通讯应用\n\n")
		fmt.Fprintf(os.Stderr, "用法: %s [选项]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  %s --config ./myconfig.json --data ./mydata\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --debug --log-level debug\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --server ws://example.com:8080\n", os.Args[0])
	}

	// 解析参数
	flag.Parse()

	// 如果指定了配置文件路径，确保是绝对路径
	if flags.ConfigPath != "" {
		if !filepath.IsAbs(flags.ConfigPath) {
			absPath, err := filepath.Abs(flags.ConfigPath)
			if err == nil {
				flags.ConfigPath = absPath
			}
		}
	}

	// 如果指定了数据目录，确保是绝对路径
	if flags.DataDir != "" {
		if !filepath.IsAbs(flags.DataDir) {
			absPath, err := filepath.Abs(flags.DataDir)
			if err == nil {
				flags.DataDir = absPath
			}
		}
	}

	return flags
}

// MergeWithFlags 将命令行参数合并到配置中
func (c *Config) MergeWithFlags(flags *CommandLineFlags) {
	// 覆盖数据目录
	if flags.DataDir != "" {
		c.App.DataDir = flags.DataDir
		c.Storage.Path = filepath.Join(flags.DataDir, "storage")
	}

	// 覆盖调试模式
	if flags.Debug {
		c.App.Debug = true
	}

	// 覆盖日志配置
	if flags.LogLevel != "" {
		c.Log.Level = flags.LogLevel
	}
	if flags.LogOutput != "" {
		c.Log.Output = flags.LogOutput
	}

	// 覆盖服务器配置
	if flags.ServerAddr != "" {
		c.Network.ListenAddr = flags.ServerAddr
	}
	if flags.ServerPort != 0 {
		c.Network.ListenPort = flags.ServerPort
	}
}
