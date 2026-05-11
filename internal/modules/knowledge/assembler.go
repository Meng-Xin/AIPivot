package knowledge

import (
	"aipivot/internal/shared/po"
	"aipivot/internal/types"

	"github.com/google/uuid"
)

// NewKnowledgeBasePo 从创建请求直接构造 PO，跳过中间领域模型。
func NewKnowledgeBasePo(req *types.CreateKnowledgeBaseRequest, tenantID int64) *po.KnowledgeBase {
	m := req.Model
	if m == "" {
		m = "text-embedding-3-small"
	}
	return &po.KnowledgeBase{
		UUID:        uuid.New().String(),
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Model:       m,
		Dimension:   ResolveDimension(m),
		Status:      "active",
	}
}

// ToShowKB 将 PO 转换为展示类型
func ToShowKB(kb *po.KnowledgeBase) types.ShowKnowledgeBase {
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

// ToShowKBList 批量将 PO 转换为展示类型
func ToShowKBList(list []*po.KnowledgeBase) []types.ShowKnowledgeBase {
	result := make([]types.ShowKnowledgeBase, 0, len(list))
	for _, kb := range list {
		result = append(result, ToShowKB(kb))
	}
	return result
}

// ToShowDocument 将文档 PO 转换为展示类型
func ToShowDocument(doc *po.Document) types.ShowDocument {
	return types.ShowDocument{
		ID:          doc.ID,
		UUID:        doc.UUID,
		Name:        doc.Name,
		ContentType: doc.ContentType,
		FileSize:    doc.FileSize,
		ChunkCount:  doc.ChunkCount,
		Status:      doc.Status,
		ErrorMsg:    doc.ErrorMsg,
		CreatedAt:   doc.CreatedAt.Unix(),
		UpdatedAt:   doc.UpdatedAt.Unix(),
	}
}

// ToShowDocumentList 批量将文档 PO 转换为展示类型
func ToShowDocumentList(list []*po.Document) []types.ShowDocument {
	result := make([]types.ShowDocument, 0, len(list))
	for _, doc := range list {
		result = append(result, ToShowDocument(doc))
	}
	return result
}
