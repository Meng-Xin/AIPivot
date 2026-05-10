package repo

import (
	"context"

	"aipivot/internal/modules/knowledge/repo/dao"
	"aipivot/internal/shared/po"
)

type KnowledgeBaseRepository interface {
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

type DocumentChunkRepository interface {
	BatchCreateWithEmbedding(ctx context.Context, chunks []dao.ChunkWithEmbedding) error
	SimilaritySearch(ctx context.Context, kbID int64, queryEmbedding []float32, topK int) ([]dao.ChunkSearchResult, error)
	DeleteByDocumentID(ctx context.Context, docID int64) error
	CountByKnowledgeBaseID(ctx context.Context, kbID int64) (int64, error)
}
