package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ChatBot represents an AI chat bot
type ChatBot struct {
	model       string
	endpoint    string
	apiKey      string
	conversations map[string][]Message
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"` // "user", "assistant", "system"
	Content string `json:"content"`
	Time    time.Time `json:"time"`
}

// SmartReply represents a smart reply suggestion
type SmartReply struct {
	Text     string
	Confidence float64
	Context  string
}

// NewChatBot creates a new AI chat bot
func NewChatBot(config map[string]string) *ChatBot {
	return &ChatBot{
		model:         config["model"],
		endpoint:      config["endpoint"],
		apiKey:        config["api_key"],
		conversations: make(map[string][]Message),
	}
}

// Chat sends a message and gets response
func (cb *ChatBot) Chat(ctx context.Context, conversationID, message string) (string, error) {
	// Initialize conversation if new
	if _, exists := cb.conversations[conversationID]; !exists {
		cb.conversations[conversationID] = []Message{
			{Role: "system", Content: "You are a helpful assistant for O2OChat P2P messaging app.", Time: time.Now()},
		}
	}

	// Add user message
	userMsg := Message{
		Role:    "user",
		Content: message,
		Time:    time.Now(),
	}
	cb.conversations[conversationID] = append(cb.conversations[conversationID], userMsg)

	// Generate response (simplified - would call AI API in production)
	response := cb.generateResponse(message)

	// Add assistant response
	assistantMsg := Message{
		Role:    "assistant",
		Content: response,
		Time:    time.Now(),
	}
	cb.conversations[conversationID] = append(cb.conversations[conversationID], assistantMsg)

	return response, nil
}

// generateResponse generates a response (simplified)
func (cb *ChatBot) generateResponse(message string) string {
	message = strings.ToLower(message)

	// Simple rule-based responses
	if strings.Contains(message, "hello") || strings.Contains(message, "hi") {
		return "Hello! How can I help you today?"
	}
	if strings.Contains(message, "how are you") {
		return "I'm doing great! How about you?"
	}
	if strings.Contains(message, "thank") {
		return "You're welcome! Anything else I can help with?"
	}
	if strings.Contains(message, "bye") {
		return "Goodbye! Have a great day!"
	}
	if strings.Contains(message, "help") {
		return "I can help with:\n- Answering questions\n- Smart replies\n- Content moderation\n- Translation\n\nWhat do you need?"
	}

	// Default response
	return "I understand. Tell me more about that."
}

// GetSmartReplies generates smart reply suggestions
func (cb *ChatBot) GetSmartReplies(ctx context.Context, context string) ([]SmartReply, error) {
	// Generate smart replies based on context
	replies := []SmartReply{
		{Text: "Sounds good!", Confidence: 0.9, Context: "positive"},
		{Text: "I'll check and get back to you", Confidence: 0.85, Context: "neutral"},
		{Text: "Thanks for letting me know", Confidence: 0.8, Context: "gratitude"},
		{Text: "That's interesting!", Confidence: 0.75, Context: "interest"},
		{Text: "Let me think about it", Confidence: 0.7, Context: "consideration"},
	}

	return replies, nil
}

// TranslateMessage translates a message
func (cb *ChatBot) TranslateMessage(ctx context.Context, text, fromLang, toLang string) (string, error) {
	// Simplified translation (would use AI translation API in production)
	return fmt.Sprintf("[Translated from %s to %s] %s", fromLang, toLang, text), nil
}

// ModerateContent checks if content is appropriate
func (cb *ChatBot) ModerateContent(ctx context.Context, content string) (bool, string, error) {
	// Simple content moderation (would use AI content moderation in production)
	inappropriate := []string{"spam", "scam", "fake", "illegal"}

	for _, word := range inappropriate {
		if strings.Contains(strings.ToLower(content), word) {
			return false, "inappropriate_content", nil
		}
	}

	return true, "approved", nil
}

// SummarizeConversation summarizes a conversation
func (cb *ChatBot) SummarizeConversation(ctx context.Context, conversationID string) (string, error) {
	messages, exists := cb.conversations[conversationID]
	if !exists {
		return "", fmt.Errorf("conversation not found")
	}

	// Simple summarization
	if len(messages) == 0 {
		return "Empty conversation", nil
	}

	summary := fmt.Sprintf("Conversation with %d messages. Last message: %s",
		len(messages), messages[len(messages)-1].Content)

	return summary, nil
}

// ExportConversation exports conversation to JSON
func (cb *ChatBot) ExportConversation(conversationID string) ([]byte, error) {
	messages, exists := cb.conversations[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation not found")
	}

	return json.MarshalIndent(messages, "", "  ")
}

// ClearConversation clears a conversation
func (cb *ChatBot) ClearConversation(conversationID) {
	delete(cb.conversations, conversationID)
}

// GetAllConversations gets all conversation IDs
func (cb *ChatBot) GetAllConversations() []string {
	ids := make([]string, 0, len(cb.conversations))
	for id := range cb.conversations {
		ids = append(ids, id)
	}
	return ids
}

// GetConversationStats gets conversation statistics
func (cb *ChatBot) GetConversationStats() map[string]interface{} {
	totalMessages := 0
	for _, messages := range cb.conversations {
		totalMessages += len(messages)
	}

	return map[string]interface{}{
		"total_conversations": len(cb.conversations),
		"total_messages":      totalMessages,
		"avg_messages_per_conv": float64(totalMessages) / float64(len(cb.conversations)),
	}
}
