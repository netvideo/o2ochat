package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/netvideo/o2ochat/pkg/identity"
	"github.com/netvideo/o2ochat/pkg/transport/p2p"
)

// Config 应用配置
type Config struct {
	Identity IdentityConfig `json:"identity"`
	P2P      P2PConfig      `json:"p2p"`
	DataDir  string         `json:"data_dir"`
	LogLevel string         `json:"log_level"`
	Debug    bool           `json:"debug"`
}

// IdentityConfig 身份配置
type IdentityConfig struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// P2PConfig P2P 配置
type P2PConfig struct {
	SignalingServer string   `json:"signaling_server"`
	ICE             []string `json:"ice"`
	STUNServers     []string `json:"stun_servers"`
}

// App 应用程序
type App struct {
	config     *Config
	identity   *identity.Identity
	p2pManager *p2p.Manager
	dataDir    string
}

// NewApp 创建新应用
func NewApp(config *Config) (*App, error) {
	app := &App{
		config:  config,
		dataDir: config.DataDir,
	}

	// 创建数据目录
	if err := os.MkdirAll(app.dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	return app, nil
}

// Init 初始化应用
func (app *App) Init() error {
	log.Println("正在初始化 O2OChat...")

	// 1. 加载或创建身份
	if err := app.initIdentity(); err != nil {
		return fmt.Errorf("初始化身份失败: %w", err)
	}

	// 2. 初始化 P2P 管理器
	if err := app.initP2P(); err != nil {
		return fmt.Errorf("初始化 P2P 失败: %w", err)
	}

	log.Println("O2OChat 初始化完成!")
	log.Printf("身份: %s (%s)", app.identity.PeerID, app.config.Identity.Nickname)

	return nil
}

// initIdentity 初始化身份
func (app *App) initIdentity() error {
	identityFile := filepath.Join(app.dataDir, "identity.json")

	// 尝试加载已有身份
	if _, err := os.Stat(identityFile); err == nil {
		data, err := os.ReadFile(identityFile)
		if err != nil {
			return fmt.Errorf("读取身份文件失败: %w", err)
		}

		identity := &identity.Identity{}
		if err := json.Unmarshal(data, identity); err != nil {
			return fmt.Errorf("解析身份文件失败: %w", err)
		}

		app.identity = identity
		log.Println("已加载现有身份")
		return nil
	}

	// 创建新身份
	log.Println("创建新身份...")

	config := &identity.Config{
		KeyType:   identity.KeyTypeEd25519,
		KeyLength: 256,
	}

	id, err := identity.CreateIdentity(config)
	if err != nil {
		return fmt.Errorf("创建身份失败: %w", err)
	}

	// 设置昵称和头像
	id.Nickname = app.config.Identity.Nickname
	id.Avatar = app.config.Identity.Avatar

	app.identity = id

	// 保存身份
	data, err := json.MarshalIndent(id, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化身份失败: %w", err)
	}

	if err := os.WriteFile(identityFile, data, 0600); err != nil {
		return fmt.Errorf("保存身份文件失败: %w", err)
	}

	log.Printf("新身份已创建: %s", id.PeerID)

	return nil
}

// initP2P 初始化 P2P
func (app *App) initP2P() error {
	log.Println("初始化 P2P...")

	config := &p2p.Config{
		SignalingServer: app.config.P2P.SignalingServer,
		ICE:             app.config.P2P.ICE,
		STUNServers:     app.config.P2P.STUNServers,
		Identity:        app.identity,
	}

	manager, err := p2p.NewManager(config)
	if err != nil {
		return fmt.Errorf("创建 P2P 管理器失败: %w", err)
	}

	app.p2pManager = manager

	// 设置事件处理器
	app.setupP2PEventHandlers()

	log.Println("P2P 初始化完成")

	return nil
}

// setupP2PEventHandlers 设置 P2P 事件处理器
func (app *App) setupP2PEventHandlers() {
	app.p2pManager.SetOnPeerConnected(func(peerID string) {
		log.Printf("Peer 已连接: %s", peerID)
	})

	app.p2pManager.SetOnPeerDisconnected(func(peerID string) {
		log.Printf("Peer 已断开: %s", peerID)
	})

	app.p2pManager.SetOnMessageReceived(func(from string, message []byte) {
		log.Printf("收到来自 %s 的消息: %s", from, string(message))
	})

	app.p2pManager.SetOnFileReceived(func(from string, fileName string, fileSize int64, fileData []byte) {
		log.Printf("收到来自 %s 的文件: %s (%d bytes)", from, fileName, fileSize)
	})
}

// Run 运行应用
func (app *App) Run() error {
	log.Println("O2OChat 正在运行...")
	log.Println("按 Ctrl+C 退出")

	// 设置信号处理器
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-sigChan

	log.Println("正在关闭 O2OChat...")

	// 关闭 P2P 管理器
	if app.p2pManager != nil {
		if err := app.p2pManager.Close(); err != nil {
			log.Printf("关闭 P2P 管理器失败: %v", err)
		}
	}

	log.Println("O2OChat 已关闭")

	return nil
}

// Shutdown 关闭应用
func (app *App) Shutdown() error {
	log.Println("正在关闭 O2OChat...")

	// 关闭 P2P 管理器
	if app.p2pManager != nil {
		if err := app.p2pManager.Close(); err != nil {
			return fmt.Errorf("关闭 P2P 管理器失败: %w", err)
		}
	}

	log.Println("O2OChat 已关闭")

	return nil
}

// GetIdentity 获取身份
func (app *App) GetIdentity() *identity.Identity {
	return app.identity
}

// GetP2PManager 获取 P2P 管理器
func (app *App) GetP2PManager() *p2p.Manager {
	return app.p2pManager
}

// GetConfig 获取配置
func (app *App) GetConfig() *Config {
	return app.config
}

// GetDataDir 获取数据目录
func (app *App) GetDataDir() string {
	return app.dataDir
}

// GetIdentityFile 获取身份文件路径
func (app *App) GetIdentityFile() string {
	return filepath.Join(app.dataDir, "identity.json")
}

// GetConfigFile 获取配置文件路径
func (app *App) GetConfigFile() string {
	return filepath.Join(app.dataDir, "config.json")
}

// SaveConfig 保存配置
func (app *App) SaveConfig() error {
	configFile := app.GetConfigFile()

	data, err := json.MarshalIndent(app.config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	return nil
}

// LoadConfig 加载配置
func (app *App) LoadConfig() error {
	configFile := app.GetConfigFile()

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 配置文件不存在，使用默认配置
			return nil
		}
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	app.config = config

	return nil
}

// InitConfig 初始化配置
func (app *App) InitConfig() error {
	// 加载配置
	if err := app.LoadConfig(); err != nil {
		return err
	}

	// 设置默认配置
	if app.config.Identity.Nickname == "" {
		app.config.Identity.Nickname = "User"
	}

	if app.config.DataDir == "" {
		app.config.DataDir = app.dataDir
	}

	// 保存配置
	if err := app.SaveConfig(); err != nil {
		return err
	}

	return nil
}
