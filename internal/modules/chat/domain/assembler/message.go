package assembler

import (
	"encoding/json"

	"aipivot/internal/shared/po"
	"aipivot/internal/types"
)

func MessagePoToShow(m *po.Message) types.ShowMessage {
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

func MessagePoListToShowList(list []*po.Message) []types.ShowMessage {
	result := make([]types.ShowMessage, 0, len(list))
	for _, m := range list {
		result = append(result, MessagePoToShow(m))
	}
	return result
}
