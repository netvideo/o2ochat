# UI Module - 用户界面模块

## 功能概述
负责应用程序的用户界面，包括聊天界面、联系人管理、文件传输界面、音视频通话界面等。

## 核心功能
1. **主界面**：应用程序主窗口和导航
2. **聊天界面**：文本消息发送和接收
3. **联系人管理**：添加、删除、搜索联系人
4. **文件传输界面**：文件发送和接收进度显示
5. **音视频通话界面**：通话控制和状态显示
6. **设置界面**：应用程序配置
7. **通知系统**：桌面通知和声音提示

## 接口定义

### 类型定义
```go
// 界面主题
type UITheme string

const (
    ThemeLight UITheme = "light"
    ThemeDark  UITheme = "dark"
    ThemeAuto  UITheme = "auto"
)

// 界面配置
type UIConfig struct {
    Theme          UITheme `json:"theme"`           // 主题
    Language       string  `json:"language"`        // 语言
    FontSize       int     `json:"font_size"`       // 字体大小
    ShowAvatars    bool    `json:"show_avatars"`    // 显示头像
    ShowTimestamps bool    `json:"show_timestamps"` // 显示时间戳
    NotifySounds   bool    `json:"notify_sounds"`   // 通知声音
    NotifyDesktop  bool    `json:"notify_desktop"`  // 桌面通知
    AutoStart      bool    `json:"auto_start"`      // 开机自启
    MinimizeToTray bool    `json:"minimize_to_tray"` // 最小化到托盘
}

// 联系人信息
type ContactInfo struct {
    PeerID      string    `json:"peer_id"`      // Peer ID
    Name        string    `json:"name"`         // 显示名称
    Avatar      string    `json:"avatar"`       // 头像URL或路径
    LastSeen    time.Time `json:"last_seen"`    // 最后在线时间
    Online      bool      `json:"online"`       // 是否在线
    UnreadCount int       `json:"unread_count"` // 未读消息数
    IsFavorite  bool      `json:"is_favorite"`  // 是否收藏
    Groups      []string  `json:"groups"`       // 所属分组
}

// 消息显示项
type MessageItem struct {
    ID          string          `json:"id"`          // 消息ID
    From        string          `json:"from"`        // 发送方
    To          string          `json:"to"`          // 接收方
    Content     string          `json:"content"`     // 消息内容
    Type        MessageType     `json:"type"`        // 消息类型
    Timestamp   time.Time       `json:"timestamp"`   // 时间戳
    Status      MessageStatus   `json:"status"`      // 消息状态
    IsOwn       bool            `json:"is_own"`      // 是否自己发送
    Attachments []*Attachment   `json:"attachments"` // 附件
    Reactions   []*Reaction     `json:"reactions"`   // 反应
}

// 附件信息
type Attachment struct {
    ID        string    `json:"id"`        // 附件ID
    FileName  string    `json:"file_name"` // 文件名
    FileSize  int64     `json:"file_size"` // 文件大小
    MimeType  string    `json:"mime_type"` // MIME类型
    Progress  float64   `json:"progress"`  // 传输进度
    Status    string    `json:"status"`    // 传输状态
    LocalPath string    `json:"local_path"` // 本地路径
}

// 通话界面状态
type CallUIState struct {
    SessionID      string        `json:"session_id"`      // 会话ID
    PeerID         string        `json:"peer_id"`         // 对端ID
    PeerName       string        `json:"peer_name"`       // 对端名称
    IsIncoming     bool          `json:"is_incoming"`     // 是否来电
    HasVideo       bool          `json:"has_video"`       // 是否有视频
    IsMuted        bool          `json:"is_muted"`        // 是否静音
    IsVideoOff     bool          `json:"is_video_off"`    // 视频是否关闭
    IsScreenSharing bool         `json:"is_screen_sharing"` // 是否屏幕共享
    Duration       time.Duration `json:"duration"`        // 通话时长
    Quality        float64       `json:"quality"`         // 通话质量
    NetworkStats   *NetworkStats `json:"network_stats"`   // 网络统计
}
```

### 主要接口
```go
// 用户界面管理器接口
type UIManager interface {
    // 初始化界面
    Initialize(config *UIConfig) error
    
    // 显示主窗口
    ShowMainWindow() error
    
    // 隐藏主窗口
    HideMainWindow() error
    
    // 退出应用程序
    Quit() error
    
    // 更新界面配置
    UpdateConfig(config *UIConfig) error
    
    // 获取当前配置
    GetConfig() (*UIConfig, error)
    
    // 显示通知
    ShowNotification(title, message string) error
    
    // 播放声音
    PlaySound(soundType string) error
    
    // 设置托盘图标
    SetTrayIcon(iconData []byte, tooltip string) error
    
    // 更新未读计数
    UpdateUnreadCount(count int) error
    
    // 销毁界面
    Destroy() error
}

// 聊天界面接口
type ChatUI interface {
    // 打开聊天窗口
    OpenChat(peerID string) error
    
    // 关闭聊天窗口
    CloseChat(peerID string) error
    
    // 添加消息
    AddMessage(message *MessageItem) error
    
    // 更新消息状态
    UpdateMessageStatus(messageID string, status MessageStatus) error
    
    // 清除聊天记录
    ClearChat(peerID string) error
    
    // 搜索消息
    SearchMessages(peerID, query string) ([]*MessageItem, error)
    
    // 获取聊天历史
    GetChatHistory(peerID string, limit int) ([]*MessageItem, error)
    
    // 设置输入回调
    SetInputCallback(callback func(text string, attachments []string)) error
    
    // 设置反应回调
    SetReactionCallback(callback func(messageID string, reaction string)) error
}

// 联系人界面接口
type ContactUI interface {
    // 添加联系人
    AddContact(contact *ContactInfo) error
    
    // 删除联系人
    RemoveContact(peerID string) error
    
    // 更新联系人信息
    UpdateContact(contact *ContactInfo) error
    
    // 搜索联系人
    SearchContacts(query string) ([]*ContactInfo, error)
    
    // 获取所有联系人
    GetAllContacts() ([]*ContactInfo, error)
    
    // 获取在线联系人
    GetOnlineContacts() ([]*ContactInfo, error)
    
    // 设置联系人选择回调
    SetContactSelectCallback(callback func(peerID string)) error
    
    // 设置添加联系人回调
    SetAddContactCallback(callback func(peerID, name string)) error
}

// 文件传输界面接口
type FileTransferUI interface {
    // 显示文件传输窗口
    ShowFileTransfer() error
    
    // 添加传输任务
    AddTransferTask(task *TransferTaskUI) error
    
    // 更新传输进度
    UpdateTransferProgress(taskID string, progress float64, speed float64) error
    
    // 完成传输任务
    CompleteTransferTask(taskID string, success bool, errorMsg string) error
    
    // 取消传输任务
    CancelTransferTask(taskID string) error
    
    // 打开文件所在位置
    OpenFileLocation(filePath string) error
    
    // 设置文件选择回调
    SetFileSelectCallback(callback func(filePaths []string)) error
    
    // 设置文件夹选择回调
    SetFolderSelectCallback(callback func(folderPath string)) error
}

// 通话界面接口
type CallUI interface {
    // 显示来电界面
    ShowIncomingCall(callInfo *CallInfo) error
    
    // 显示去电界面
    ShowOutgoingCall(callInfo *CallInfo) error
    
    // 更新通话状态
    UpdateCallState(state *CallUIState) error
    
    // 结束通话界面
    EndCall(sessionID string) error
    
    // 设置视频帧回调
    SetVideoFrameCallback(callback func(frame []byte, width, height int)) error
    
    // 设置音频数据回调
    SetAudioDataCallback(callback func(data []byte, sampleRate int)) error
    
    // 设置通话控制回调
    SetCallControlCallback(callback func(action CallAction)) error
}

// 设置界面接口
type SettingsUI interface {
    // 显示设置窗口
    ShowSettings() error
    
    // 更新设置项
    UpdateSetting(section, key string, value interface{}) error
    
    // 获取设置项
    GetSetting(section, key string) (interface{}, error)
    
    // 重置设置
    ResetSettings() error
    
    // 设置保存回调
    SetSaveCallback(callback func(config *UIConfig)) error
    
    // 设置测试回调
    SetTestCallback(callback func(testType string)) error
}
```

## 实现要求

### 1. 跨平台支持
- **Windows**：WinUI 3 或传统Win32
- **macOS**：Cocoa 或 SwiftUI
- **Linux**：GTK 或 Qt
- **Web**：React/Vue（可选）

### 2. 响应式设计
- 支持窗口大小调整
- 适应不同屏幕分辨率
- 支持深色/浅色主题
- 国际化支持

### 3. 性能优化
- 虚拟列表显示大量消息
- 图片懒加载和缓存
- 减少UI重绘
- 异步UI更新

### 4. 用户体验
- 流畅的动画效果
- 直观的操作反馈
- 快捷键支持
- 无障碍访问

## 测试要求

### 单元测试
```bash
# 运行UI模块测试
go test ./ui -v

# 测试特定功能
go test ./ui -run TestUIManager
go test ./ui -run TestChatUI
go test ./ui -run TestContactUI
```

### UI测试
```bash
# 需要图形界面环境
go test ./ui -tags=uitest

# 测试界面交互
go test ./ui -tags=interaction
```

### 测试用例
1. **界面渲染测试**：测试UI组件渲染
2. **用户交互测试**：测试按钮点击等交互
3. **响应式测试**：测试窗口大小调整
4. **主题切换测试**：测试深色/浅色主题
5. **性能测试**：测试UI响应速度

## 依赖关系
- 所有功能模块：用于业务逻辑
- storage模块：用于配置存储

## 使用示例

```go
// 创建UI管理器
config := &UIConfig{
    Theme:          ThemeDark,
    Language:       "zh-CN",
    FontSize:       14,
    ShowAvatars:    true,
    ShowTimestamps: true,
    NotifySounds:   true,
    NotifyDesktop:  true,
    AutoStart:      false,
    MinimizeToTray: true,
}

uiManager, err := NewUIManager()
err = uiManager.Initialize(config)

// 显示主窗口
err = uiManager.ShowMainWindow()

// 设置托盘图标
iconData, _ := os.ReadFile("icon.png")
err = uiManager.SetTrayIcon(iconData, "O2OChat")

// 聊天界面
chatUI := NewChatUI()

// 打开聊天窗口
err = chatUI.OpenChat("QmPeer456")

// 添加消息
message := &MessageItem{
    ID:        "msg123",
    From:      "QmPeer456",
    To:        "QmPeer123",
    Content:   "你好！",
    Type:      MessageTypeText,
    Timestamp: time.Now(),
    Status:    MessageStatusDelivered,
    IsOwn:     false,
}
err = chatUI.AddMessage(message)

// 设置输入回调
err = chatUI.SetInputCallback(func(text string, attachments []string) {
    // 发送消息
    sendMessage("QmPeer456", text, attachments)
})

// 联系人界面
contactUI := NewContactUI()

// 添加联系人
contact := &ContactInfo{
    PeerID:      "QmPeer456",
    Name:        "张三",
    Avatar:      "avatar.png",
    LastSeen:    time.Now(),
    Online:      true,
    UnreadCount: 3,
    IsFavorite:  true,
    Groups:      []string{"朋友", "同事"},
}
err = contactUI.AddContact(contact)

// 设置联系人选择回调
err = contactUI.SetContactSelectCallback(func(peerID string) {
    // 打开聊天窗口
    chatUI.OpenChat(peerID)
})

// 文件传输界面
fileTransferUI := NewFileTransferUI()

// 添加传输任务
task := &TransferTaskUI{
    TaskID:     "task123",
    FileName:   "document.pdf",
    FileSize:   1024 * 1024 * 10, // 10MB
    Direction:  "download",
    PeerName:   "张三",
    Progress:   0.0,
    Speed:      0.0,
    Status:     "downloading",
}
err = fileTransferUI.AddTransferTask(task)

// 更新传输进度
err = fileTransferUI.UpdateTransferProgress("task123", 0.5, 1024.0) // 50%, 1MB/s

// 通话界面
callUI := NewCallUI()

// 显示来电界面
callInfo := &CallInfo{
    PeerID:     "QmPeer456",
    PeerName:   "张三",
    HasVideo:   true,
    IsIncoming: true,
}
err = callUI.ShowIncomingCall(callInfo)

// 设置通话控制回调
err = callUI.SetCallControlCallback(func(action CallAction) {
    switch action {
    case CallActionAccept:
        acceptCall(callInfo.PeerID)
    case CallActionReject:
        rejectCall(callInfo.PeerID)
    case CallActionMute:
        toggleMute()
    case CallActionVideoOff:
        toggleVideo()
    case CallActionEnd:
        endCall()
    }
})

// 设置界面
settingsUI := NewSettingsUI()

// 显示设置窗口
err = settingsUI.ShowSettings()

// 设置保存回调
err = settingsUI.SetSaveCallback(func(config *UIConfig) {
    // 保存配置
    saveConfig(config)
})

// 显示通知
err = uiManager.ShowNotification("新消息", "张三: 你好！")

// 播放声音
err = uiManager.PlaySound("message")

// 更新未读计数
err = uiManager.UpdateUnreadCount(5)

// 退出应用程序
err = uiManager.Quit()
```

## 界面布局示例

```html
<!-- 主界面布局 -->
<div class="app-container">
  <!-- 侧边栏 -->
  <div class="sidebar">
    <div class="user-info">
      <img class="avatar" src="user-avatar.png">
      <span class="name">我的名字</span>
      <span class="status online"></span>
    </div>
    
    <div class="search-box">
      <input type="text" placeholder="搜索联系人...">
    </div>
    
    <div class="contact-list">
      <!-- 联系人项 -->
      <div class="contact-item active">
        <img class="avatar" src="contact-avatar.png">
        <div class="contact-info">
          <span class="name">张三</span>
          <span class="last-message">你好！</span>
        </div>
        <span class="unread-count">3</span>
        <span class="status online"></span>
      </div>
    </div>
  </div>
  
  <!-- 主内容区 -->
  <div class="main-content">
    <!-- 聊天头部 -->
    <div class="chat-header">
      <img class="avatar" src="contact-avatar.png">
      <div class="contact-info">
        <span class="name">张三</span>
        <span class="status online">在线</span>
      </div>
      <div class="chat-actions">
        <button class="btn-video">视频通话</button>
        <button class="btn-voice">语音通话</button>
        <button class="btn-more">更多</button>
      </div>
    </div>
    
    <!-- 消息区域 -->
    <div class="message-area">
      <!-- 消息项 -->
      <div class="message-item received">
        <img class="avatar" src="contact-avatar.png">
        <div class="message-content">
          <div class="message-text">你好！</div>
          <div class="message-time">10:30</div>
        </div>
      </div>
      
      <div class="message-item sent">
        <div class="message-content">
          <div class="message-text">你好！最近怎么样？</div>
          <div class="message-time">10:31 ✓✓</div>
        </div>
      </div>
    </div>
    
    <!-- 输入区域 -->
    <div class="input-area">
      <button class="btn-attach">📎</button>
      <textarea class="message-input" placeholder="输入消息..."></textarea>
      <button class="btn-send">发送</button>
    </div>
  </div>
</div>
```

## 错误处理
- UI组件加载失败必须优雅降级
- 资源加载失败必须显示占位符
- 用户操作失败必须提供反馈
- 内存不足必须释放资源

## 优化建议
1. **虚拟滚动**：大量消息列表性能优化
2. **图片优化**：懒加载、压缩、缓存
3. **CSS优化**：减少重绘和回流
4. **事件委托**：减少事件监听器数量
5. **代码分割**：按需加载UI组件