package knowledge

import (
	"context"
	"errors"

	"aipivot/internal/shared/po"
)

// ========== 校验 ==========

// EmbeddingModels 当前支持的 Embedding 模型白名单
var EmbeddingModels = map[string]int{
	"text-embedding-3-small": 1536,
	"text-embedding-3-large": 3072,
	"text-embedding-ada-002": 1536,
}

func ValidateName(name string) error {
	if name == "" {
		return errors.New("知识库名称不能为空")
	}
	if len(name) > 255 {
		return errors.New("知识库名称不能超过 255 个字符")
	}
	return nil
}

// ResolveDimension 根据模型名称解析向量维度，未知模型默认 1536
func ResolveDimension(model string) int {
	if dim, ok := EmbeddingModels[model]; ok {
		return dim
	}
	return 1536
}

// ========== Repository 接口 ==========

type KBRepository interface {
	Create(ctx context.Context, kb *po.KnowledgeBase) error
	GetByID(ctx context.Context, id int64) (*po.KnowledgeBase, error)
	GetList(ctx context.Context, tenantID int64, page, pageSize int, name string) ([]*po.KnowledgeBase, int64, error)
	Update(ctx context.Context, id int64, updates map[string]any) error
	Delete(ctx context.Context, id int64) error
}

type DocumentRepository interface {
	Create(ctx context.Context, doc *po.Document) error
	GetByID(ctx context.Context, id int64) (*po.Document, error)
	GetList(ctx context.Context, kbID int64, page, pageSize int, status string) ([]*po.Document, int64, error)
	Delete(ctx context.Context, id int64) error
	UpdateStatus(ctx context.Context, id int64, status, errorMsg string) error
}

type DocChunkRepository interface {
	BatchCreateWithEmbedding(ctx context.Context, chunks []ChunkWithEmbedding) error
	SimilaritySearch(ctx context.Context, kbID int64, queryEmbedding []float32, topK int) ([]ChunkSearchResult, error)
	DeleteByDocumentID(ctx context.Context, docID int64) error
	CountByKnowledgeBaseID(ctx context.Context, kbID int64) (int64, error)
}

// ========== 辅助类型 ==========

// ChunkWithEmbedding 带向量的切块数据，用于批量写入。
type ChunkWithEmbedding struct {
	DocumentID      int64
	KnowledgeBaseID int64
	TenantID        int64
	ChunkIndex      int
	Content         string
	TokenCount      int
	Embedding       []float32
	Metadata        string
}

// ChunkSearchResult 相似度搜索结果。
type ChunkSearchResult struct {
	ID              int64   `gorm:"column:id"`
	UUID            string  `gorm:"column:uuid"`
	DocumentID      int64   `gorm:"column:document_id"`
	KnowledgeBaseID int64   `gorm:"column:knowledge_base_id"`
	ChunkIndex      int     `gorm:"column:chunk_index"`
	Content         string  `gorm:"column:content"`
	TokenCount      int     `gorm:"column:token_count"`
	Metadata        string  `gorm:"column:metadata"`
	Score           float64 `gorm:"column:score"`
}
