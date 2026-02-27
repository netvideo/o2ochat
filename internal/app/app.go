package app

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/netvideo/cli"
	"github.com/netvideo/crypto"
	"github.com/netvideo/filetransfer"
	"github.com/netvideo/identity"
	"github.com/netvideo/media"
	"github.com/netvideo/signaling"
	"github.com/netvideo/storage"
	"github.com/netvideo/transport"
	"github.com/netvideo/ui"
)

type AppConfig struct {
	ConfigPath string
	DataDir    string
	Debug      bool
}

type Application struct {
	config *AppConfig

	IdentityManager  identity.IdentityManager
	CryptoManager    crypto.CryptoManager
	StorageManager   storage.StorageManager
	TransportManager transport.TransportManager
	SignalingClient  signaling.SignalingClient
	FileTransferMgr  filetransfer.FileTransferManager
	MediaManager     media.MediaManager
	UIManager        ui.UIManager
	CLIManager       cli.CLIManager
}

func NewApplication(config *AppConfig) (*Application, error) {
	app := &Application{
		config: config,
	}

	// 初始化所有模块
	if err := app.initializeModules(); err != nil {
		return nil, fmt.Errorf("failed to initialize modules: %w", err)
	}

	return app, nil
}

func (a *Application) Start() error {
	log.Println("Starting O2OChat application...")
	log.Println("All modules initialized successfully")
	return nil
}

func (a *Application) Stop() error {
	log.Println("Stopping O2OChat application...")

	// 清理资源
	if a.UIManager != nil {
		a.UIManager.Destroy()
	}
	if a.MediaManager != nil {
		a.MediaManager.Destroy()
	}
	if a.SignalingClient != nil {
		a.SignalingClient.Close()
	}
	if a.TransportManager != nil {
		a.TransportManager.Close()
	}

	log.Println("Application stopped")
	return nil
}

func (a *Application) initializeModules() error {
	dataDir := a.config.DataDir

	// 1. 初始化存储模块
	log.Println("Initializing storage module...")
	storageManager := storage.NewSQLiteStorageManager()
	storageConfig := &storage.StorageConfig{
		Path: filepath.Join(dataDir, "storage"),
	}
	if err := storageManager.Initialize(storageConfig); err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	a.StorageManager = storageManager

	// 2. 初始化加密模块
	log.Println("Initializing crypto module...")
	cryptoConfig := crypto.DefaultSecurityConfig()
	cryptoManager := crypto.NewCryptoManager(cryptoConfig)
	a.CryptoManager = cryptoManager

	// 3. 初始化身份模块
	log.Println("Initializing identity module...")
	identityStore, err := identity.NewFileIdentityStore(filepath.Join(dataDir, "identity"))
	if err != nil {
		return fmt.Errorf("failed to create identity store: %w", err)
	}
	keyStore, err := identity.NewFileKeyStorage(filepath.Join(dataDir, "keys"))
	if err != nil {
		return fmt.Errorf("failed to create key store: %w", err)
	}
	identityManager, err := identity.NewIdentityManager(identityStore, keyStore)
	if err != nil {
		return fmt.Errorf("failed to create identity manager: %w", err)
	}
	a.IdentityManager = identityManager

	// 4. 初始化传输模块
	log.Println("Initializing transport module...")
	transportConfig := &transport.TransportConfig{
		MaxConnections: 100,
		KeepAlive:      true,
	}
	transportManager := transport.NewTransportManager(transportConfig)
	a.TransportManager = transportManager

	// 5. 初始化信令模块
	log.Println("Initializing signaling module...")
	signalingConfig := &signaling.ClientConfig{
		ServerURL: "ws://localhost:8080",
	}
	signalingClient := signaling.NewWebSocketClient(signalingConfig)
	a.SignalingClient = signalingClient

	// 6. 初始化文件传输模块
	log.Println("Initializing file transfer module...")
	chunkManager, err := filetransfer.NewChunkManager(1024*1024, filepath.Join(dataDir, "chunks"))
	if err != nil {
		return fmt.Errorf("failed to create chunk manager: %w", err)
	}
	fileTransferManager := filetransfer.NewFileTransferManager(chunkManager, nil, 10)
	a.FileTransferMgr = fileTransferManager

	// 7. 初始化媒体模块
	log.Println("Initializing media module...")
	mediaManager, err := media.NewMediaManager()
	if err != nil {
		return fmt.Errorf("failed to create media manager: %w", err)
	}
	a.MediaManager = mediaManager

	// 8. 初始化 UI 模块
	log.Println("Initializing UI module...")
	uiManager := ui.NewUIManager()
	a.UIManager = uiManager

	// 9. 初始化 CLI 模块
	log.Println("Initializing CLI module...")
	cliManager := cli.NewCLIManager()
	a.CLIManager = cliManager

	log.Println("All modules initialized successfully")
	return nil
}

func (a *Application) GetIdentityManager() identity.IdentityManager {
	return a.IdentityManager
}

func (a *Application) GetTransportManager() transport.TransportManager {
	return a.TransportManager
}

func (a *Application) GetSignalingClient() signaling.SignalingClient {
	return a.SignalingClient
}

func (a *Application) GetFileTransferManager() filetransfer.FileTransferManager {
	return a.FileTransferMgr
}

func (a *Application) GetMediaManager() media.MediaManager {
	return a.MediaManager
}

func (a *Application) GetUIManager() ui.UIManager {
	return a.UIManager
}

func (a *Application) GetStorageManager() storage.StorageManager {
	return a.StorageManager
}

func (a *Application) GetCryptoManager() crypto.CryptoManager {
	return a.CryptoManager
}
