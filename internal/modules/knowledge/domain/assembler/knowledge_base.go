package assembler

import (
	"aipivot/internal/modules/knowledge/domain/model"
	"aipivot/internal/shared/po"
	"aipivot/internal/types"
)

// ① Request → Domain Model
func CreateKBRequestToModel(req *types.CreateKnowledgeBaseRequest) *model.KnowledgeBase {
	m := req.Model
	if m == "" {
		m = "text-embedding-3-small"
	}
	return &model.KnowledgeBase{
		Name:        req.Name,
		Description: req.Description,
		Model:       m,
	}
}

// ② Domain Model → PO
func ModelKBToKnowledgeBasePo(m *model.KnowledgeBase, tenantID int64) *po.KnowledgeBase {
	return &po.KnowledgeBase{
		TenantID:    tenantID,
		Name:        m.Name,
		Description: m.Description,
		Model:       m.Model,
		Dimension:   m.ResolveDimension(),
		Status:      "active",
	}
}

// ③ PO → ShowType
func KnowledgeBasePoToShow(kb *po.KnowledgeBase) types.ShowKnowledgeBase {
	return types.ShowKnowledgeBase{
		ID:          kb.ID,
		UUID:        kb.UUID,
		Name:        kb.Name,
		Description: kb.Description,
		Model:       kb.Model,
		Dimension:   kb.Dimension,
		Status:      kb.Status,
		DocCount:    kb.DocCount,
		ChunkCount:  kb.ChunkCount,
		CreatedAt:   kb.CreatedAt.Unix(),
		UpdatedAt:   kb.UpdatedAt.Unix(),
	}
}

// ③ PO List → ShowType List
func KnowledgeBasePoListToShowList(list []*po.KnowledgeBase) []types.ShowKnowledgeBase {
	result := make([]types.ShowKnowledgeBase, 0, len(list))
	for _, kb := range list {
		result = append(result, KnowledgeBasePoToShow(kb))
	}
	return result
}
