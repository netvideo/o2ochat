# Storage Module - 数据存储模块

## 功能概述
负责应用程序数据的持久化存储，包括配置文件、消息历史、文件块、密钥等数据的存储和管理。

## 核心功能
1. **配置存储**：应用程序配置和用户设置
2. **消息存储**：聊天消息历史和元数据
3. **文件块存储**：文件传输的临时块数据
4. **密钥存储**：加密密钥的安全存储
5. **缓存管理**：内存缓存和磁盘缓存
6. **数据迁移**：版本升级时的数据迁移

## 接口定义

### 类型定义
```go
// 存储类型
type StorageType string

const (
    StorageTypeSQLite   StorageType = "sqlite"
    StorageTypeBoltDB   StorageType = "boltdb"
    StorageTypeLevelDB  StorageType = "leveldb"
    StorageTypeBadger   StorageType = "badger"
    StorageTypeMemory   StorageType = "memory"
)

// 存储配置
type StorageConfig struct {
    Type        StorageType `json:"type"`         // 存储类型
    Path        string      `json:"path"`         // 存储路径
    MaxSize     int64       `json:"max_size"`     // 最大存储大小（字节）
    Compression bool        `json:"compression"`  // 是否启用压缩
    Encryption  bool        `json:"encryption"`   // 是否启用加密
    CacheSize   int         `json:"cache_size"`   // 缓存大小（MB）
}

// 存储统计
type StorageStats struct {
    TotalSize     int64     `json:"total_size"`     // 总大小
    UsedSize      int64     `json:"used_size"`      // 已用大小
    FreeSize      int64     `json:"free_size"`      // 可用大小
    FileCount     int64     `json:"file_count"`     // 文件数量
    ChunkCount    int64     `json:"chunk_count"`    // 块数量
    MessageCount  int64     `json:"message_count"`  // 消息数量
    LastBackup    time.Time `json:"last_backup"`    // 最后备份时间
}

// 存储操作选项
type StorageOptions struct {
    TTL         time.Duration `json:"ttl"`          // 生存时间
    Compression bool          `json:"compression"`  // 是否压缩
    Encryption  bool          `json:"encryption"`   // 是否加密
    Priority    int           `json:"priority"`     // 存储优先级
}

// 存储条目
type StorageEntry struct {
    Key         string        `json:"key"`          // 键
    Value       []byte        `json:"value"`        // 值
    Metadata    map[string]string `json:"metadata"` // 元数据
    CreatedAt   time.Time     `json:"created_at"`   // 创建时间
    UpdatedAt   time.Time     `json:"updated_at"`   // 更新时间
    ExpiresAt   time.Time     `json:"expires_at"`   // 过期时间
    Size        int           `json:"size"`         // 大小
}
```

### 主要接口
```go
// 存储管理器接口
type StorageManager interface {
    // 初始化存储
    Initialize(config *StorageConfig) error
    
    // 存储数据
    Put(key string, value []byte, options *StorageOptions) error
    
    // 获取数据
    Get(key string) ([]byte, error)
    
    // 删除数据
    Delete(key string) error
    
    // 检查是否存在
    Exists(key string) (bool, error)
    
    // 列出所有键
    List(prefix string) ([]string, error)
    
    // 批量操作
    BatchPut(entries map[string][]byte, options *StorageOptions) error
    BatchGet(keys []string) (map[string][]byte, error)
    BatchDelete(keys []string) error
    
    // 获取存储统计
    GetStats() (*StorageStats, error)
    
    // 清理过期数据
    CleanupExpired() error
    
    // 备份数据
    Backup(backupPath string) error
    
    // 恢复数据
    Restore(backupPath string) error
    
    // 关闭存储
    Close() error
}

// 消息存储接口
type MessageStorage interface {
    // 存储消息
    StoreMessage(message *ChatMessage) error
    
    // 获取消息
    GetMessage(messageID string) (*ChatMessage, error)
    
    // 获取对话消息
    GetConversationMessages(peerID string, limit int, offset int) ([]*ChatMessage, error)
    
    // 搜索消息
    SearchMessages(query string, peerID string, limit int) ([]*ChatMessage, error)
    
    // 删除消息
    DeleteMessage(messageID string) error
    
    // 清理旧消息
    CleanupOldMessages(before time.Time) error
    
    // 获取消息统计
    GetMessageStats(peerID string) (*MessageStats, error)
}

// 文件块存储接口
type ChunkStorage interface {
    // 存储文件块
    StoreChunk(fileID string, index int, data []byte) error
    
    // 获取文件块
    GetChunk(fileID string, index int) ([]byte, error)
    
    // 检查块是否存在
    ChunkExists(fileID string, index int) (bool, error)
    
    // 获取所有块索引
    GetChunkIndices(fileID string) ([]int, error)
    
    // 删除文件块
    DeleteChunk(fileID string, index int) error
    
    // 删除所有文件块
    DeleteAllChunks(fileID string) error
    
    // 获取文件块统计
    GetChunkStats(fileID string) (*ChunkStats, error)
    
    // 清理临时块
    CleanupTemporaryChunks() error
}

// 配置存储接口
type ConfigStorage interface {
    // 存储配置
    StoreConfig(section string, key string, value interface{}) error
    
    // 获取配置
    GetConfig(section string, key string, defaultValue interface{}) (interface{}, error)
    
    // 删除配置
    DeleteConfig(section string, key string) error
    
    // 获取所有配置
    GetAllConfig(section string) (map[string]interface{}, error)
    
    // 导入配置
    ImportConfig(configPath string) error
    
    // 导出配置
    ExportConfig(configPath string) error
    
    // 重置配置
    ResetConfig() error
}

// 缓存管理器接口
type CacheManager interface {
    // 设置缓存
    Set(key string, value []byte, ttl time.Duration) error
    
    // 获取缓存
    Get(key string) ([]byte, error)
    
    // 删除缓存
    Delete(key string) error
    
    // 清空缓存
    Clear() error
    
    // 获取缓存统计
    GetCacheStats() (*CacheStats, error)
    
    // 调整缓存大小
    Resize(newSize int) error
}
```

## 实现要求

### 1. 存储引擎选择
- **SQLite**：关系数据（消息、配置）
- **BoltDB/LevelDB**：键值数据（块、缓存）
- **内存缓存**：热点数据加速
- **文件系统**：大文件块存储

### 2. 数据组织
```
data/
├── config/           # 配置文件
├── messages/         # 消息存储
├── chunks/           # 文件块存储
├── keys/             # 密钥存储（加密）
├── cache/            # 缓存数据
└── backups/          # 备份文件
```

### 3. 性能优化
- 实现LRU缓存策略
- 批量操作减少IO
- 压缩存储减少空间
- 异步写入提高响应速度

### 4. 数据安全
- 敏感数据加密存储
- 实现数据完整性验证
- 定期备份和恢复
- 防止数据损坏

## 测试要求

### 单元测试
```bash
# 运行存储模块测试
go test ./storage -v

# 测试特定功能
go test ./storage -run TestStorageManager
go test ./storage -run TestMessageStorage
go test ./storage -run TestChunkStorage
```

### 集成测试
```bash
# 需要实际文件系统
go test ./storage -tags=integration

# 测试大文件存储
go test ./storage -tags=largedata
```

### 测试用例
1. **基本操作测试**：测试CRUD操作
2. **并发测试**：测试多线程并发访问
3. **性能测试**：测试读写性能
4. **恢复测试**：测试备份和恢复
5. **错误测试**：测试磁盘满等错误场景

### 性能测试
```bash
# 基准测试
go test ./storage -bench=.
go test ./storage -bench=BenchmarkPut
go test ./storage -bench=BenchmarkGet
```

## 依赖关系
- crypto模块：用于数据加密
- filetransfer模块：使用块存储

## 使用示例

```go
// 创建存储管理器
config := &StorageConfig{
    Type:        StorageTypeSQLite,
    Path:        "./data",
    MaxSize:     10 * 1024 * 1024 * 1024, // 10GB
    Compression: true,
    Encryption:  true,
    CacheSize:   256, // 256MB
}

manager, err := NewStorageManager(config)
err = manager.Initialize()

// 存储数据
options := &StorageOptions{
    TTL:         24 * time.Hour,
    Compression: true,
    Encryption:  true,
}

data := []byte("需要存储的数据")
err = manager.Put("user:config:theme", data, options)

// 获取数据
value, err := manager.Get("user:config:theme")

// 批量操作
entries := map[string][]byte{
    "user:1:name": []byte("Alice"),
    "user:1:age":  []byte("30"),
    "user:2:name": []byte("Bob"),
}
err = manager.BatchPut(entries, options)

// 消息存储
messageStorage := NewMessageStorage(manager)

message := &ChatMessage{
    ID:        "msg123",
    From:      "QmPeer123",
    To:        "QmPeer456",
    Content:   "Hello!",
    Timestamp: time.Now(),
    Type:      MessageTypeText,
}
err = messageStorage.StoreMessage(message)

// 获取对话消息
messages, err := messageStorage.GetConversationMessages("QmPeer456", 50, 0)

// 文件块存储
chunkStorage := NewChunkStorage(manager)

fileID := "file123"
chunkData := make([]byte, 1024*1024) // 1MB
rand.Read(chunkData)

err = chunkStorage.StoreChunk(fileID, 0, chunkData)

retrievedChunk, err := chunkStorage.GetChunk(fileID, 0)

// 缓存管理
cache := NewCacheManager(256 * 1024 * 1024) // 256MB

err = cache.Set("hot:data", []byte("频繁访问的数据"), 5*time.Minute)

cachedData, err := cache.Get("hot:data")

// 获取存储统计
stats, err := manager.GetStats()
fmt.Printf("总大小: %d MB, 已用: %d MB, 文件数: %d\n",
    stats.TotalSize/1024/1024,
    stats.UsedSize/1024/1024,
    stats.FileCount,
)

// 备份数据
err = manager.Backup("./backups/data-backup-" + time.Now().Format("20060102"))

// 清理过期数据
err = manager.CleanupExpired()

// 关闭存储
err = manager.Close()
```

## SQLite表结构示例

```sql
-- 消息表
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    from_peer TEXT NOT NULL,
    to_peer TEXT NOT NULL,
    content BLOB NOT NULL,
    type TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    delivered INTEGER DEFAULT 0,
    read INTEGER DEFAULT 0,
    encrypted INTEGER DEFAULT 1
);

-- 文件块表
CREATE TABLE chunks (
    file_id TEXT NOT NULL,
    chunk_index INTEGER NOT NULL,
    data BLOB NOT NULL,
    hash TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    expires_at INTEGER,
    PRIMARY KEY (file_id, chunk_index)
);

-- 配置表
CREATE TABLE config (
    section TEXT NOT NULL,
    key TEXT NOT NULL,
    value BLOB NOT NULL,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY (section, key)
);

-- 创建索引
CREATE INDEX idx_messages_conversation ON messages(from_peer, to_peer, timestamp);
CREATE INDEX idx_messages_timestamp ON messages(timestamp);
CREATE INDEX idx_chunks_file ON chunks(file_id);
CREATE INDEX idx_chunks_expires ON chunks(expires_at) WHERE expires_at IS NOT NULL;
```

## 错误处理
- 磁盘空间不足必须优雅处理
- 数据损坏必须尝试恢复
- 并发冲突必须正确处理
- 加密失败必须保护敏感数据

## 优化建议
1. **分片存储**：大文件分片存储
2. **压缩算法**：根据数据类型选择压缩算法
3. **预取策略**：预测性数据加载
4. **写时复制**：减少锁竞争
5. **异步提交**：提高写入性能

## 数据迁移
- 支持版本升级数据迁移
- 实现数据格式转换
- 提供迁移回滚机制
- 验证迁移数据完整性