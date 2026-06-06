package flows

import (
	"encoding/json"

	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/po"
	"aipivot/internal/types"
)

const (
	flowStatusDraft     = "draft"
	flowStatusPublished = "published"
	flowStatusArchived  = "archived"
)

func defaultDefinition() po.JSONMap {
	return po.JSONMap{
		"nodes": []map[string]interface{}{
			{
				"id":    "start",
				"type":  "trigger",
				"label": "开始",
				"x":     120,
				"y":     140,
				"config": map[string]interface{}{
					"event": "conversation.message",
				},
			},
			{
				"id":    "reply",
				"type":  "llm",
				"label": "AI 回复",
				"x":     420,
				"y":     140,
				"config": map[string]interface{}{
					"mode": "rag",
				},
			},
		},
		"edges": []map[string]interface{}{
			{
				"id":     "edge-start-reply",
				"source": "start",
				"target": "reply",
			},
		},
		"viewport": map[string]interface{}{
			"x":    0,
			"y":    0,
			"zoom": 1,
		},
	}
}

func parseDefinition(input string) (po.JSONMap, error) {
	if input == "" {
		return defaultDefinition(), nil
	}

	var definition po.JSONMap
	if err := json.Unmarshal([]byte(input), &definition); err != nil {
		return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "definition 不是有效的 JSON")
	}
	if definition == nil {
		definition = po.JSONMap{}
	}
	return definition, nil
}

func normalizeStatus(status string) (string, error) {
	if status == "" {
		return flowStatusDraft, nil
	}
	switch status {
	case flowStatusDraft, flowStatusPublished, flowStatusArchived:
		return status, nil
	default:
		return "", errorx.NewBusinessError(errorx.CodeBadRequest, "status 只能是 draft / published / archived")
	}
}

func toShowFlow(f *po.Flow) types.ShowFlow {
	definitionJSON, _ := json.Marshal(f.Definition)
	return types.ShowFlow{
		ID:          f.ID,
		UUID:        f.UUID,
		Name:        f.Name,
		Description: f.Description,
		Definition:  string(definitionJSON),
		Status:      f.Status,
		Version:     f.Version,
		CreatedAt:   f.CreatedAt.Unix(),
		UpdatedAt:   f.UpdatedAt.Unix(),
	}
}

func toShowFlows(list []*po.Flow) []types.ShowFlow {
	result := make([]types.ShowFlow, 0, len(list))
	for _, f := range list {
		result = append(result, toShowFlow(f))
	}
	return result
}
