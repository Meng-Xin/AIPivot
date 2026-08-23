package knowledge

import (
	"encoding/json"

	"aipivot/internal/shared/po"
	"aipivot/internal/types"

	"github.com/google/uuid"
)

// maxSuggestedQuestions 引导问答条数硬上限（与 .api 注释保持一致）。
const maxSuggestedQuestions = 6

// SerializeSuggestedQuestions 将 string 切片序列化为 JSONB 兼容的字符串。
// nil / 空 都存 "[]"，避免 NOT NULL 列写入空字符串。
func SerializeSuggestedQuestions(qs []string) string {
	if len(qs) == 0 {
		return "[]"
	}
	b, err := json.Marshal(qs)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// DeserializeSuggestedQuestions 反序列化 JSONB 字符串为 string 切片，解析失败兜底空切片。
func DeserializeSuggestedQuestions(raw string) []string {
	if raw == "" {
		return []string{}
	}
	var qs []string
	if err := json.Unmarshal([]byte(raw), &qs); err != nil {
		return []string{}
	}
	return qs
}

// NormalizeSuggestedQuestions 校验+裁剪：最多 6 条，每条 ≤ 100 字，去空白。
func NormalizeSuggestedQuestions(qs []string) []string {
	if len(qs) == 0 {
		return nil
	}
	out := make([]string, 0, len(qs))
	for _, q := range qs {
		// 截断超长问题（避免 JSONB 列被滥用），不做硬拒绝以便后台兼容历史脏数据
		if len(q) > 100 {
			q = q[:100]
		}
		out = append(out, q)
	}
	if len(out) > maxSuggestedQuestions {
		out = out[:maxSuggestedQuestions]
	}
	return out
}

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
		// settings 是 NOT NULL JSONB：GORM 会把零值 "" 显式写入并触发 22P02，
		// 必须显式给空对象而不是依赖列的 DEFAULT '{}'。
		Settings:           "{}",
		SuggestedQuestions: SerializeSuggestedQuestions(NormalizeSuggestedQuestions(req.SuggestedQuestions)),
	}
}

// ToShowKB 将 PO 转换为展示类型
func ToShowKB(kb *po.KnowledgeBase) types.ShowKnowledgeBase {
	return types.ShowKnowledgeBase{
		ID:                 kb.ID,
		UUID:               kb.UUID,
		Name:               kb.Name,
		Description:        kb.Description,
		Model:              kb.Model,
		Dimension:          kb.Dimension,
		Status:             kb.Status,
		DocCount:           kb.DocCount,
		ChunkCount:         kb.ChunkCount,
		SuggestedQuestions: DeserializeSuggestedQuestions(kb.SuggestedQuestions),
		CreatedAt:          kb.CreatedAt.Unix(),
		UpdatedAt:          kb.UpdatedAt.Unix(),
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
