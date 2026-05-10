package assembler

import (
	"aipivot/internal/shared/po"
	"aipivot/internal/types"
)

func ConversationPoToShow(c *po.Conversation) types.ShowConversation {
	var kbID int64
	if c.KnowledgeBaseID != nil {
		kbID = *c.KnowledgeBaseID
	}
	return types.ShowConversation{
		ID:              c.ID,
		UUID:            c.UUID,
		KnowledgeBaseID: kbID,
		Title:           c.Title,
		Status:          c.Status,
		Channel:         c.Channel,
		MessageCount:    c.MessageCount,
		CreatedAt:       c.CreatedAt.Unix(),
		UpdatedAt:       c.UpdatedAt.Unix(),
	}
}

func ConversationPoListToShowList(list []*po.Conversation) []types.ShowConversation {
	result := make([]types.ShowConversation, 0, len(list))
	for _, c := range list {
		result = append(result, ConversationPoToShow(c))
	}
	return result
}
