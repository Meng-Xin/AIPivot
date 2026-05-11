package chat

import (
	"encoding/json"

	"aipivot/internal/shared/po"
	"aipivot/internal/types"
)

// ToShowConversation 将会话 PO 转换为展示类型
func ToShowConversation(c *po.Conversation) types.ShowConversation {
	var kbID int64
	if c.KnowledgeBaseID != nil {
		kbID = *c.KnowledgeBaseID
	}
	return types.ShowConversation{
		ID:              c.ID,
		UUID:            c.UUID,
		KnowledgeBaseID: kbID,
		Title:           c.Title,
		Model:           c.Model,
		Status:          c.Status,
		Channel:         c.Channel,
		MessageCount:    c.MessageCount,
		CreatedAt:       c.CreatedAt.Unix(),
		UpdatedAt:       c.UpdatedAt.Unix(),
	}
}

// ToShowConversationList 批量将会话 PO 转换为展示类型
func ToShowConversationList(list []*po.Conversation) []types.ShowConversation {
	result := make([]types.ShowConversation, 0, len(list))
	for _, c := range list {
		result = append(result, ToShowConversation(c))
	}
	return result
}

// ToShowMessage 将消息 PO 转换为展示类型
func ToShowMessage(m *po.Message) types.ShowMessage {
	var sources []string
	// sources 存储为 JSON 数组字符串，解析失败则忽略
	_ = json.Unmarshal([]byte(m.Sources), &sources)

	return types.ShowMessage{
		ID:          m.ID,
		UUID:        m.UUID,
		Role:        m.Role,
		Content:     m.Content,
		ContentType: m.ContentType,
		TokenCount:  m.TokenCount,
		Model:       m.Model,
		LatencyMs:   m.LatencyMs,
		Sources:     sources,
		CreatedAt:   m.CreatedAt.Unix(),
	}
}

// ToShowMessageList 批量将消息 PO 转换为展示类型
func ToShowMessageList(list []*po.Message) []types.ShowMessage {
	result := make([]types.ShowMessage, 0, len(list))
	for _, m := range list {
		result = append(result, ToShowMessage(m))
	}
	return result
}
