package skills

import (
	"encoding/json"

	"aipivot/internal/shared/po"
	"aipivot/internal/types"
)

// toShowSkill 将 Skill PO 转换为 API 展示对象。
// JSONMap 字段序列化为 JSON 字符串供前端直接使用。
func toShowSkill(s *po.Skill) types.ShowSkill {
	paramsJSON, _ := json.Marshal(s.Parameters)
	headersJSON, _ := json.Marshal(s.Headers)
	return types.ShowSkill{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Parameters:  string(paramsJSON),
		Endpoint:    s.Endpoint,
		Method:      s.Method,
		Headers:     string(headersJSON),
		TimeoutMs:   s.TimeoutMs,
		Enabled:     s.Enabled,
		CreatedAt:   s.CreatedAt.Unix(),
		UpdatedAt:   s.UpdatedAt.Unix(),
	}
}

// toShowSkills 批量转换。
func toShowSkills(list []*po.Skill) []types.ShowSkill {
	result := make([]types.ShowSkill, 0, len(list))
	for _, s := range list {
		result = append(result, toShowSkill(s))
	}
	return result
}
