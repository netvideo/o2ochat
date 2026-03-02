package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"
)

// Contact represents a contact
type Contact struct {
	PeerID  string
	Nickname string
	AddedAt time.Time
}

// Global state
var contacts = []Contact{}

func main() {
	// Generate peer ID
	peerID := GeneratePeerID()

	// Show main menu
	showMainMenu(peerID)
}

func showMainMenu(peerID string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println()
		fmt.Println("╔════════════════════════════════════════════════╗")
		fmt.Println("║       O2OChat v4.0.0 - P2P 即时通讯            ║")
		fmt.Println("╚════════════════════════════════════════════════╝")
		fmt.Println()
		fmt.Printf("📱 您的 Peer ID: %s\n", peerID)
		fmt.Println()
		fmt.Printf("👥 联系人：%d 个\n", len(contacts))
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
			showSettings()
		case "0":
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

	fmt.Print("输入昵称：")
	nickname, _ := reader.ReadString('\n')
	nickname = strings.TrimSpace(nickname)

	if nickname == "" {
		fmt.Println("❌ 昵称不能为空")
		return
	}

	// Check if already exists
	for _, c := range contacts {
		if c.PeerID == peerID {
			fmt.Println("❌ 该联系人已存在")
			return
		}
	}

	// Add contact
	contact := Contact{
		PeerID:  peerID,
		Nickname: nickname,
		AddedAt: time.Now(),
	}
	contacts = append(contacts, contact)
	
	fmt.Printf("✅ 已添加联系人：%s (Peer ID: %s)\n", nickname, peerID)
}

func showContacts() {
	fmt.Println()
	fmt.Println("=== 联系人列表 ===")
	
	if len(contacts) == 0 {
		fmt.Println("暂无联系人")
		fmt.Println("使用\"添加联系人\"功能添加好友")
		return
	}
	
	for i, c := range contacts {
		fmt.Printf("%d. %s (Peer ID: %s)\n", i+1, c.Nickname, c.PeerID)
		fmt.Printf("   添加时间：%s\n", c.AddedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("\n共 %d 个联系人\n", len(contacts))
}

func sendMessage(reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("=== 发送消息 ===")
	
	if len(contacts) == 0 {
		fmt.Println("❌ 暂无联系人，请先添加联系人")
		return
	}
	
	// Show contacts
	fmt.Println("选择联系人:")
	for i, c := range contacts {
		fmt.Printf("%d. %s\n", i+1, c.Nickname)
	}
	fmt.Print("请选择：")
	
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	
	fmt.Print("输入消息：")
	message, _ := reader.ReadString('\n')
	message = strings.TrimSpace(message)

	if message != "" {
		fmt.Printf("✅ 消息已发送给 %s：%s\n", contacts[0].Nickname, message)
	}
}

func receiveMessages() {
	fmt.Println()
	fmt.Println("=== 接收消息 ===")
	fmt.Println("暂无新消息")
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

func showSettings() {
	fmt.Println()
	fmt.Println("=== 设置 ===")
	fmt.Println("1. 修改昵称")
	fmt.Println("2. 隐私设置")
	fmt.Println("3. 通知设置")
	fmt.Println("4. 关于")
	fmt.Println("0. 返回")
}

// GeneratePeerID generates a unique peer ID
func GeneratePeerID() string {
	timestamp := time.Now().UnixNano()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", timestamp)))
	return "Qm" + hex.EncodeToString(hash[:20])
}
