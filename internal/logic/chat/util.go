package chat

import (
	"aipivot/internal/shared/po"
	"aipivot/pkg/llm"
)

// buildChatHistory 将最近消息列表转为 LLM ChatMessage 格式（排除系统消息）。
func buildChatHistory(msgs []*po.Message) []llm.ChatMessage {
	history := make([]llm.ChatMessage, 0, len(msgs))
	for _, m := range msgs {
		if m.Role == "user" || m.Role == "assistant" {
			history = append(history, llm.ChatMessage{Role: m.Role, Content: m.Content})
		}
	}
	return history
}
