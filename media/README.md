# Media Module - 音视频处理模块

## 功能概述
负责音视频采集、编码、传输、解码和渲染，支持实时语音和视频通话功能。

## 核心功能
1. **设备管理**：音视频设备枚举和选择
2. **媒体采集**：音频和视频帧采集
3. **编码解码**：Opus音频和VP8/H.264视频编解码
4. **传输处理**：RTP包封装和传输
5. **网络适应**：带宽估计和码率自适应
6. **质量优化**：丢包隐藏和抖动缓冲

## 接口定义

### 类型定义
```go
// 媒体类型
type MediaType string

const (
    MediaTypeAudio MediaType = "audio"
    MediaTypeVideo MediaType = "video"
)

// 媒体配置
type MediaConfig struct {
    MediaType     MediaType          // 媒体类型
    Enabled       bool               // 是否启用
    Codec         string             // 编解码器
    Bitrate       int                // 目标码率（bps）
    SampleRate    int                // 采样率（音频）
    Channels      int                // 声道数（音频）
    Width         int                // 宽度（视频）
    Height        int                // 高度（视频）
    FrameRate     int                // 帧率（视频）
    KeyFrameInterval int             // 关键帧间隔（视频）
}

// 设备信息
type DeviceInfo struct {
    ID          string    `json:"id"`          // 设备ID
    Name        string    `json:"name"`        // 设备名称
    Type        MediaType `json:"type"`        // 设备类型
    Default     bool      `json:"default"`     // 是否默认设备
}

// 媒体帧
type MediaFrame struct {
    Type        MediaType          // 帧类型
    Timestamp   uint32             // 时间戳（RTP）
    Sequence    uint16             // 序列号（RTP）
    Payload     []byte             // 帧数据
    Size        int                // 数据大小
    KeyFrame    bool               // 是否关键帧（视频）
    Duration    time.Duration      // 帧时长
}

// 通话配置
type CallConfig struct {
    AudioConfig  *MediaConfig      // 音频配置
    VideoConfig  *MediaConfig      // 视频配置
    MaxBitrate   int               // 最大总码率
    MinBitrate   int               // 最小总码率
    StartBitrate int               // 起始码率
    UseFEC       bool              // 是否使用前向纠错
    UseNACK      bool              // 是否使用NACK
    UsePLI       bool              // 是否使用PLI
}

// 通话统计
type CallStats struct {
    AudioStats  *StreamStats       // 音频统计
    VideoStats  *StreamStats       // 视频统计
    NetworkStats *NetworkStats     // 网络统计
    Quality     float64            // 通话质量评分（0-1）
}

// 流统计
type StreamStats struct {
    Bitrate      int               // 当前码率（bps）
    PacketLoss   float64           // 丢包率（0-1）
    Jitter       time.Duration     // 抖动
    Latency      time.Duration     // 延迟
    FramesSent   int64             // 发送帧数
    FramesReceived int64           // 接收帧数
    FramesDropped int64            // 丢弃帧数
}
```

### 主要接口
```go
// 媒体管理器接口
type MediaManager interface {
    // 初始化媒体引擎
    Initialize() error
    
    // 获取可用设备
    GetDevices(mediaType MediaType) ([]*DeviceInfo, error)
    
    // 创建通话会话
    CreateCallSession(config *CallConfig) (CallSession, error)
    
    // 加入通话
    JoinCall(sessionID string) (CallSession, error)
    
    // 离开通话
    LeaveCall(sessionID string) error
    
    // 获取通话统计
    GetCallStats(sessionID string) (*CallStats, error)
    
    // 销毁媒体引擎
    Destroy() error
}

// 通话会话接口
type CallSession interface {
    // 开始通话
    Start() error
    
    // 停止通话
    Stop() error
    
    // 暂停媒体流
    Pause(mediaType MediaType) error
    
    // 恢复媒体流
    Resume(mediaType MediaType) error
    
    // 切换设备
    SwitchDevice(mediaType MediaType, deviceID string) error
    
    // 调整码率
    AdjustBitrate(targetBitrate int) error
    
    // 发送媒体帧
    SendFrame(frame *MediaFrame) error
    
    // 接收媒体帧
    ReceiveFrame() (*MediaFrame, error)
    
    // 获取会话ID
    GetSessionID() string
    
    // 获取对端信息
    GetRemoteInfo() *PeerInfo
    
    // 关闭会话
    Close() error
}

// 编解码器接口
type Codec interface {
    // 编码帧
    EncodeFrame(input []byte) ([]byte, error)
    
    // 解码帧
    DecodeFrame(input []byte) ([]byte, error)
    
    // 获取编解码器信息
    GetCodecInfo() *CodecInfo
    
    // 设置编码参数
    SetEncoderParams(params map[string]interface{}) error
    
    // 重置编解码器
    Reset() error
    
    // 关闭编解码器
    Close() error
}

// RTP处理器接口
type RTPProcessor interface {
    // 封装RTP包
    Packetize(frame *MediaFrame) ([]*RTPPacket, error)
    
    // 解封装RTP包
    Depacketize(packet *RTPPacket) (*MediaFrame, error)
    
    // 处理NACK
    HandleNACK(seqNums []uint16) ([]*RTPPacket, error)
    
    // 处理PLI
    HandlePLI() error
    
    // 处理FIR
    HandleFIR() error
    
    // 获取RTP统计
    GetRTPStats() *RTPStats
}

// 抖动缓冲接口
type JitterBuffer interface {
    // 添加RTP包
    AddPacket(packet *RTPPacket) error
    
    // 获取下一个帧
    GetNextFrame() (*MediaFrame, error)
    
    // 设置缓冲区大小
    SetBufferSize(size time.Duration) error
    
    // 获取缓冲区状态
    GetBufferStatus() *BufferStatus
    
    // 重置缓冲区
    Reset() error
}
```

## 实现要求

### 1. 音频处理
- 使用Opus编解码器（推荐）
- 支持多种采样率（8k, 16k, 48k）
- 实现回声消除（AEC）
- 实现噪声抑制（NS）
- 实现自动增益控制（AGC）

### 2. 视频处理
- 使用VP8或H.264编解码器
- 支持硬件加速（可选）
- 实现视频缩放和裁剪
- 实现帧率控制
- 实现关键帧请求

### 3. RTP传输
- 实现RTP/RTCP协议
- 支持NACK重传
- 支持前向纠错（FEC）
- 实现带宽估计（REMB/TMMBR）

### 4. 网络适应
- 实现码率自适应
- 支持拥塞控制
- 实现丢包隐藏
- 支持网络切换

## 测试要求

### 单元测试
```bash
# 运行媒体模块测试
go test ./media -v

# 测试特定功能
go test ./media -run TestAudioCodec
go test ./media -run TestVideoCodec
go test ./media -run TestRTPProcessor
```

### 集成测试
```bash
# 需要音视频设备
go test ./media -tags=integration

# 测试通话功能
go test ./media -tags=calltest
```

### 测试用例
1. **编解码测试**：测试音视频编解码质量
2. **RTP测试**：测试RTP包封装和解封装
3. **抖动缓冲测试**：测试缓冲区管理
4. **网络适应测试**：测试码率自适应
5. **设备测试**：测试设备枚举和切换

### 性能测试
```bash
# 基准测试
go test ./media -bench=.
go test ./media -bench=BenchmarkAudioEncode
go test ./media -bench=BenchmarkVideoEncode
```

## 依赖关系
- transport模块：用于媒体数据传输
- crypto模块：用于SRTP加密（可选）
- ui模块：用于音视频渲染

## 使用示例

```go
// 创建媒体管理器
manager := NewMediaManager()
err := manager.Initialize()

// 获取音频设备
audioDevices, err := manager.GetDevices(MediaTypeAudio)

// 配置通话
config := &CallConfig{
    AudioConfig: &MediaConfig{
        MediaType:  MediaTypeAudio,
        Enabled:    true,
        Codec:      "opus",
        Bitrate:    64000, // 64kbps
        SampleRate: 48000,
        Channels:   2,
    },
    VideoConfig: &MediaConfig{
        MediaType:  MediaTypeVideo,
        Enabled:    true,
        Codec:      "vp8",
        Bitrate:    500000, // 500kbps
        Width:      640,
        Height:     480,
        FrameRate:  30,
    },
    MaxBitrate:   1000000, // 1Mbps
    MinBitrate:   100000,  // 100kbps
    StartBitrate: 500000,  // 500kbps
    UseFEC:       true,
    UseNACK:      true,
}

// 创建通话会话
session, err := manager.CreateCallSession(config)

// 开始通话
err = session.Start()

// 发送音频帧（示例）
go func() {
    for {
        // 从音频设备采集帧
        audioFrame := captureAudioFrame()
        
        frame := &MediaFrame{
            Type:      MediaTypeAudio,
            Timestamp: generateTimestamp(),
            Payload:   audioFrame,
            Size:      len(audioFrame),
        }
        
        err := session.SendFrame(frame)
        if err != nil {
            break
        }
        
        time.Sleep(20 * time.Millisecond) // 50fps
    }
}()

// 接收视频帧（示例）
go func() {
    for {
        frame, err := session.ReceiveFrame()
        if err != nil {
            break
        }
        
        if frame.Type == MediaTypeVideo {
            // 渲染视频帧
            renderVideoFrame(frame.Payload)
        }
    }
}()

// 监控通话质量
go func() {
    for {
        stats, err := manager.GetCallStats(session.GetSessionID())
        if err != nil {
            break
        }
        
        fmt.Printf("音频码率: %d bps, 视频码率: %d bps, 丢包率: %.2f%%\n",
            stats.AudioStats.Bitrate,
            stats.VideoStats.Bitrate,
            stats.NetworkStats.PacketLoss*100,
        )
        
        time.Sleep(5 * time.Second)
    }
}()
```

## 音频配置示例

```go
// Opus编码器配置
opusConfig := map[string]interface{}{
    "application":       "voip",      // voip, audio, lowdelay
    "bitrate":          64000,       // 目标码率
    "complexity":       5,           // 复杂度（0-10）
    "signal":           "voice",     // voice, music
    "inband_fec":       true,        // 带内FEC
    "packet_loss_perc": 5,           // 预期丢包率
}
```

## 视频配置示例

```go
// VP8编码器配置
vp8Config := map[string]interface{}{
    "target_bitrate":    500000,     // 目标码率
    "keyframe_interval": 3000,       // 关键帧间隔（ms）
    "deadline":          "realtime", // 编码截止时间
    "error_resilient":   true,       // 错误恢复
    "lag_in_frames":     0,          // 延迟帧数
}
```

## 错误处理
- 设备不可用必须优雅降级
- 编码失败必须跳过帧
- 网络中断必须尝试恢复
- 内存不足必须释放资源

## 质量优化
1. **自适应码率**：根据网络状况调整
2. **前向纠错**：提高抗丢包能力
3. **丢包隐藏**：掩盖丢失的音频/视频
4. **网络预测**：预测网络变化提前调整

## 平台兼容性
- Windows: DirectShow, Media Foundation
- macOS: AVFoundation
- Linux: V4L2, PulseAudio
- Web: WebRTC API