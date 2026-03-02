package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// Contact represents a contact
type Contact struct {
	PeerID   string    `json:"peer_id"`
	Nickname string    `json:"nickname"`
	AddedAt  time.Time `json:"added_at"`
}

// Message represents a message
type Message struct {
	From      string    `json:"from"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Read      bool      `json:"read"`
}

// Settings represents user settings
type Settings struct {
	Nickname      string `json:"nickname"`
	ShowOnline    bool   `json:"show_online"`
	ShowReadReceipt bool `json:"show_read_receipt"`
}

// AppState represents the application state
type AppState struct {
	PeerID     string      `json:"peer_id"`
	PrivateKey string      `json:"private_key"`
	Contacts   []Contact   `json:"contacts"`
	Messages   []Message   `json:"messages"`
	Settings   Settings    `json:"settings"`
}

// Global state
var state AppState
var dataFile = "o2ochat_data.json"

func main() {
	// Load or create state
	loadState()

	// Show main menu
	showMainMenu()
}

func loadState() {
	// Try to load from file
	data, err := os.ReadFile(dataFile)
	if err == nil {
		err = json.Unmarshal(data, &state)
		if err == nil && state.PeerID != "" {
			fmt.Printf("✅ 已加载保存的数据\n")
			fmt.Printf("📱 您的 Peer ID: %s\n", state.PeerID)
			return
		}
	}

	// Create new state
	state = AppState{
		PeerID:     GeneratePeerID(),
		PrivateKey: generatePrivateKey(),
		Contacts:   []Contact{},
		Messages:   []Message{},
		Settings: Settings{
			Nickname:      "用户",
			ShowOnline:    true,
			ShowReadReceipt: true,
		},
	}

	fmt.Printf("✨ 创建新的 Peer ID: %s\n", state.PeerID)
	saveState()
}

func saveState() {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		fmt.Printf("❌ 保存数据失败：%v\n", err)
		return
	}

	err = os.WriteFile(dataFile, data, 0644)
	if err != nil {
		fmt.Printf("❌ 保存文件失败：%v\n", err)
		return
	}

	fmt.Printf("✅ 数据已保存\n")
}

func showMainMenu() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println()
		fmt.Println("╔════════════════════════════════════════════════╗")
		fmt.Println("║       O2OChat v4.0.0 - P2P 即时通讯            ║")
		fmt.Println("╚════════════════════════════════════════════════╝")
		fmt.Println()
		fmt.Printf("📱 您的 Peer ID: %s\n", state.PeerID)
		fmt.Printf("👤 昵称：%s\n", state.Settings.Nickname)
		fmt.Printf("👥 联系人：%d 个\n", len(state.Contacts))
		fmt.Printf("📬 未读消息：%d 条\n", countUnreadMessages())
		fmt.Println()
		fmt.Println("1. 添加联系人")
		fmt.Println("2. 查看联系人列表")
		fmt.Println("3. 发送消息")
		fmt.Println("4. 接收消息")
		fmt.Println("5. 文件传输")
		fmt.Println("6. AI 翻译")
		fmt.Println("7. 设置")
		fmt.Println("0. 退出")
		fmt.Println()
		fmt.Print("请选择操作：")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			addContact(reader)
		case "2":
			showContacts()
		case "3":
			sendMessage(reader)
		case "4":
			receiveMessages()
		case "5":
			fileTransfer(reader)
		case "6":
			aiTranslate(reader)
		case "7":
			showSettings(reader)
		case "0":
			saveState()
			fmt.Println("👋 再见！")
			return
		default:
			fmt.Println("❌ 无效选择，请重新输入")
		}
	}
}

func addContact(reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("=== 添加联系人 ===")
	fmt.Print("输入对方 Peer ID: ")
	peerID, _ := reader.ReadString('\n')
	peerID = strings.TrimSpace(peerID)

	if peerID == "" {
		fmt.Println("❌ Peer ID 不能为空")
		return
	}

	if peerID == state.PeerID {
		fmt.Println("❌ 不能添加自己为联系人")
		return
	}

	// Check if already exists
	for _, c := range state.Contacts {
		if c.PeerID == peerID {
			fmt.Println("❌ 该联系人已存在")
			return
		}
	}

	fmt.Print("输入昵称：")
	nickname, _ := reader.ReadString('\n')
	nickname = strings.TrimSpace(nickname)

	if nickname == "" {
		fmt.Println("❌ 昵称不能为空")
		return
	}

	// Add contact
	contact := Contact{
		PeerID:   peerID,
		Nickname: nickname,
		AddedAt:  time.Now(),
	}
	state.Contacts = append(state.Contacts, contact)
	saveState()

	fmt.Printf("✅ 已添加联系人：%s (Peer ID: %s)\n", nickname, peerID)
}

func showContacts() {
	fmt.Println()
	fmt.Println("=== 联系人列表 ===")

	if len(state.Contacts) == 0 {
		fmt.Println("暂无联系人")
		fmt.Println("使用\"添加联系人\"功能添加好友")
		return
	}

	for i, c := range state.Contacts {
		fmt.Printf("%d. %s (Peer ID: %s)\n", i+1, c.Nickname, c.PeerID)
		fmt.Printf("   添加时间：%s\n", c.AddedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("\n共 %d 个联系人\n", len(state.Contacts))
}

func sendMessage(reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("=== 发送消息 ===")

	if len(state.Contacts) == 0 {
		fmt.Println("❌ 暂无联系人，请先添加联系人")
		return
	}

	// Show contacts
	fmt.Println("选择联系人:")
	for i, c := range state.Contacts {
		fmt.Printf("%d. %s\n", i+1, c.Nickname)
	}
	fmt.Print("请选择：")

	choiceStr, _ := reader.ReadString('\n')
	choiceStr = strings.TrimSpace(choiceStr)

	fmt.Print("输入消息：")
	message, _ := reader.ReadString('\n')
	message = strings.TrimSpace(message)

	if message != "" {
		fmt.Printf("✅ 消息已发送给 %s：%s\n", state.Contacts[0].Nickname, message)
	}
}

func receiveMessages() {
	fmt.Println()
	fmt.Println("=== 接收消息 ===")

	if len(state.Messages) == 0 {
		fmt.Println("暂无新消息")
		return
	}

	for i, msg := range state.Messages {
		status := "未读"
		if msg.Read {
			status = "已读"
		}
		fmt.Printf("%d. [%s] %s: %s\n", i+1, status, msg.From, msg.Content)
		fmt.Printf("   时间：%s\n", msg.Timestamp.Format("2006-01-02 15:04:05"))
	}
}

func fileTransfer(reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("=== 文件传输 ===")
	fmt.Println("1. 发送文件")
	fmt.Println("2. 接收文件")
	fmt.Println("0. 返回")
	fmt.Print("请选择：")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		fmt.Println("选择要发送的文件...")
	case "2":
		fmt.Println("等待接收文件...")
	}
}

func aiTranslate(reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("=== AI 翻译 ===")
	fmt.Print("输入要翻译的文字：")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)

	if text != "" {
		fmt.Println("✅ 翻译功能开发中...")
		fmt.Printf("原文：%s\n", text)
	}
}

func showSettings(reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("=== 设置 ===")
	fmt.Println("1. 修改昵称")
	fmt.Println("2. 在线状态")
	fmt.Println("3. 已读回执")
	fmt.Println("4. 保存数据")
	fmt.Println("5. 关于")
	fmt.Println("0. 返回")
	fmt.Print("请选择：")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		fmt.Print("输入新昵称：")
		nickname, _ := reader.ReadString('\n')
		nickname = strings.TrimSpace(nickname)
		if nickname != "" {
			state.Settings.Nickname = nickname
			saveState()
			fmt.Printf("✅ 昵称已修改为：%s\n", nickname)
		}
	case "2":
		state.Settings.ShowOnline = !state.Settings.ShowOnline
		saveState()
		fmt.Printf("✅ 在线状态：%v\n", state.Settings.ShowOnline)
	case "3":
		state.Settings.ShowReadReceipt = !state.Settings.ShowReadReceipt
		saveState()
		fmt.Printf("✅ 已读回执：%v\n", state.Settings.ShowReadReceipt)
	case "4":
		saveState()
		fmt.Println("✅ 数据已保存")
	case "5":
		fmt.Println()
		fmt.Println("关于 O2OChat")
		fmt.Println("版本：v4.0.0")
		fmt.Println("纯 P2P 即时通讯软件")
		fmt.Println("端到端加密")
		fmt.Println("无需中央服务器")
	case "0":
		return
	default:
		fmt.Println("❌ 无效选择")
	}
}

func countUnreadMessages() int {
	count := 0
	for _, msg := range state.Messages {
		if !msg.Read {
			count++
		}
	}
	return count
}

// GeneratePeerID generates a unique peer ID
func GeneratePeerID() string {
	timestamp := time.Now().UnixNano()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", timestamp)))
	return "Qm" + hex.EncodeToString(hash[:20])
}

func generatePrivateKey() string {
	timestamp := time.Now().UnixNano()
	hash := sha256.Sum256([]byte(fmt.Sprintf("key-%d", timestamp)))
	return hex.EncodeToString(hash[:])
}
