package repo

import (
	"context"

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
