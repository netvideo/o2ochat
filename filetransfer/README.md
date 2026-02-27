# FileTransfer Module - 文件传输模块

## 功能概述
负责大文件的分块传输、多源下载、完整性验证和断点续传功能。

## 核心功能
1. **文件分块**：将大文件分割为固定大小的块
2. **Merkle树验证**：构建和验证文件完整性
3. **多源下载**：从多个Peer并发下载不同块
4. **调度算法**：智能选择下载源和块顺序
5. **断点续传**：支持传输中断后恢复
6. **进度监控**：实时传输进度和速度统计

## 接口定义

### 类型定义
```go
// 文件元数据
type FileMetadata struct {
    FileID      string    `json:"file_id"`      // 文件唯一标识（Merkle根哈希）
    FileName    string    `json:"file_name"`    // 文件名
    FileSize    int64     `json:"file_size"`    // 文件大小（字节）
    TotalChunks int       `json:"total_chunks"` // 总块数
    ChunkSize   int       `json:"chunk_size"`   // 块大小（字节）
    MerkleRoot  []byte    `json:"merkle_root"`  // Merkle树根哈希
    CreatedAt   time.Time `json:"created_at"`   // 创建时间
    ModifiedAt  time.Time `json:"modified_at"`  // 修改时间
}

// 块信息
type ChunkInfo struct {
    FileID    string `json:"file_id"`    // 文件ID
    Index     int    `json:"index"`      // 块索引（0-based）
    Offset    int64  `json:"offset"`     // 文件偏移量
    Size      int    `json:"size"`       // 块大小
    Hash      []byte `json:"hash"`       // 块哈希
    Completed bool   `json:"completed"`  // 是否已完成
    Verified  bool   `json:"verified"`   // 是否已验证
}

// 传输任务
type TransferTask struct {
    TaskID      string            `json:"task_id"`      // 任务ID
    FileID      string            `json:"file_id"`      // 文件ID
    Direction   TransferDirection `json:"direction"`    // 传输方向
    SourcePeers []string          `json:"source_peers"` // 源Peer列表
    DestPath    string            `json:"dest_path"`    // 目标路径
    StartedAt   time.Time         `json:"started_at"`   // 开始时间
    Status      TransferStatus    `json:"status"`       // 传输状态
    Progress    TransferProgress  `json:"progress"`     // 传输进度
}

// 传输状态
type TransferStatus string

const (
    StatusPending    TransferStatus = "pending"
    StatusDownloading TransferStatus = "downloading"
    StatusUploading  TransferStatus = "uploading"
    StatusPaused     TransferStatus = "paused"
    StatusCompleted  TransferStatus = "completed"
    StatusFailed     TransferStatus = "failed"
    StatusCancelled  TransferStatus = "cancelled"
)

// 传输方向
type TransferDirection string

const (
    DirectionDownload TransferDirection = "download"
    DirectionUpload   TransferDirection = "upload"
)

// 传输进度
type TransferProgress struct {
    TotalChunks   int     `json:"total_chunks"`   // 总块数
    Completed     int     `json:"completed"`      // 已完成块数
    Failed        int     `json:"failed"`         // 失败块数
    BytesTransferred int64 `json:"bytes_transferred"` // 已传输字节数
    Speed         float64 `json:"speed"`          // 传输速度（KB/s）
    EstimatedTime string  `json:"estimated_time"` // 预计剩余时间
}
```

### 主要接口
```go
// 文件传输管理器接口
type FileTransferManager interface {
    // 创建下载任务
    CreateDownloadTask(fileID, destPath string, sourcePeers []string) (string, error)
    
    // 创建上传任务
    CreateUploadTask(filePath string, targetPeers []string) (string, error)
    
    // 开始传输
    StartTransfer(taskID string) error
    
    // 暂停传输
    PauseTransfer(taskID string) error
    
    // 恢复传输
    ResumeTransfer(taskID string) error
    
    // 取消传输
    CancelTransfer(taskID string) error
    
    // 获取任务状态
    GetTaskStatus(taskID string) (*TransferTask, error)
    
    // 获取所有任务
    GetAllTasks() ([]*TransferTask, error)
    
    // 清理已完成任务
    CleanupCompletedTasks() error
}

// 块管理器接口
type ChunkManager interface {
    // 分块文件
    ChunkFile(filePath string, chunkSize int) (*FileMetadata, error)
    
    // 合并文件
    MergeFile(fileID, outputPath string) error
    
    // 获取块信息
    GetChunkInfo(fileID string, index int) (*ChunkInfo, error)
    
    // 获取所有块信息
    GetAllChunks(fileID string) ([]*ChunkInfo, error)
    
    // 验证块完整性
    VerifyChunk(fileID string, index int, data []byte) (bool, error)
    
    // 验证文件完整性
    VerifyFile(fileID string) (bool, error)
    
    // 保存块数据
    SaveChunk(fileID string, index int, data []byte) error
    
    // 读取块数据
    ReadChunk(fileID string, index int) ([]byte, error)
}

// 调度器接口
type Scheduler interface {
    // 选择下一个要下载的块
    SelectNextChunk(fileID string, availableChunks map[int][]string) (int, []string, error)
    
    // 分配下载任务
    AssignDownloadTasks(fileID string, maxConcurrent int) (map[int][]string, error)
    
    // 更新块状态
    UpdateChunkStatus(fileID string, index int, success bool, peerID string) error
    
    // 获取调度统计
    GetSchedulerStats(fileID string) (*SchedulerStats, error)
}

// Merkle树接口
type MerkleTree interface {
    // 构建Merkle树
    BuildTree(chunks [][]byte) ([][]byte, error)
    
    // 获取根哈希
    GetRootHash() []byte
    
    // 验证块
    VerifyChunk(index int, chunkData []byte, proof [][]byte) (bool, error)
    
    // 生成验证证明
    GenerateProof(index int) ([][]byte, error)
    
    // 序列化树
    Serialize() ([]byte, error)
    
    // 反序列化树
    Deserialize(data []byte) error
}
```

## 实现要求

### 1. 文件分块
- 固定块大小（默认1MB）
- 支持可变块大小（可选）
- 处理文件末尾不足块的情况

### 2. Merkle树实现
- 使用SHA256哈希算法
- 支持快速验证单个块
- 实现高效的证明生成

### 3. 调度算法
1. **稀缺优先**：优先下载稀有块
2. **随机选择**：避免热点块竞争
3. **带宽感知**：根据Peer带宽分配任务
4. **失败重试**：失败块重新调度

### 4. 多源下载
- 从多个Peer并发下载
- 动态调整并发连接数
- 实现负载均衡

### 5. 断点续传
- 保存传输状态
- 支持暂停和恢复
- 清理临时文件

## 测试要求

### 单元测试
```bash
# 运行文件传输模块测试
go test ./filetransfer -v

# 测试特定功能
go test ./filetransfer -run TestChunking
go test ./filetransfer -run TestMerkleTree
go test ./filetransfer -run TestScheduler
```

### 集成测试
```bash
# 需要实际文件系统
go test ./filetransfer -tags=integration

# 测试大文件传输
go test ./filetransfer -tags=largefile
```

### 测试用例
1. **分块测试**：测试各种文件大小的分块
2. **完整性测试**：测试Merkle树验证
3. **调度测试**：测试调度算法
4. **并发测试**：测试多源并发下载
5. **恢复测试**：测试断点续传

### 性能测试
```bash
# 基准测试
go test ./filetransfer -bench=.
go test ./filetransfer -bench=BenchmarkChunking
go test ./filetransfer -bench=BenchmarkMerkleTree
```

## 依赖关系
- transport模块：用于数据传输
- crypto模块：用于哈希计算
- storage模块：用于块存储

## 使用示例

```go
// 创建文件传输管理器
manager := NewFileTransferManager()

// 分块文件
metadata, err := manager.ChunkFile("/path/to/largefile.iso", 1024*1024) // 1MB块

// 创建下载任务
taskID, err := manager.CreateDownloadTask(
    metadata.FileID,
    "/path/to/download",
    []string{"QmPeer123", "QmPeer456"},
)

// 开始下载
err = manager.StartTransfer(taskID)

// 监控进度
for {
    task, err := manager.GetTaskStatus(taskID)
    if err != nil {
        break
    }
    
    progress := task.Progress
    fmt.Printf("进度: %.2f%%, 速度: %.2f KB/s\n",
        float64(progress.Completed)/float64(progress.TotalChunks)*100,
        progress.Speed,
    )
    
    if task.Status == StatusCompleted || task.Status == StatusFailed {
        break
    }
    
    time.Sleep(1 * time.Second)
}

// 验证文件完整性
valid, err := manager.VerifyFile(metadata.FileID)
if valid {
    fmt.Println("文件验证成功")
}
```

## 调度算法示例

```go
// 稀缺优先调度器
type RarestFirstScheduler struct {
    chunkAvailability map[int]int // 块索引 -> 可用Peer数
}

func (s *RarestFirstScheduler) SelectNextChunk(
    fileID string,
    availableChunks map[int][]string,
) (int, []string, error) {
    // 找到最稀有的块
    var rarestChunk int
    minAvailability := math.MaxInt32
    
    for chunkIndex, peers := range availableChunks {
        availability := len(peers)
        if availability < minAvailability {
            minAvailability = availability
            rarestChunk = chunkIndex
        }
    }
    
    return rarestChunk, availableChunks[rarestChunk], nil
}
```

## 错误处理
- 块下载失败必须重试（有限次数）
- 完整性验证失败必须重新下载
- 磁盘空间不足必须暂停传输

## 监控指标
1. 传输速度（实时/平均）
2. 完成百分比
3. 并发连接数
4. 块成功率
5. 磁盘使用情况

## 优化建议
1. **内存映射文件**：大文件处理
2. **零拷贝传输**：减少内存复制
3. **预取策略**：提前下载后续块
4. **压缩传输**：减少网络流量（可选）