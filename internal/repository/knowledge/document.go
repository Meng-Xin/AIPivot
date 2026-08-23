package knowledge

import (
	"context"
	"errors"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"gorm.io/gorm"
)

// DocumentRepo 文档数据仓储实现。
type DocumentRepo struct {
	q *query.Query
}

func NewDocumentRepo(q *query.Query) *DocumentRepo {
	return &DocumentRepo{q: q}
}

func (r *DocumentRepo) Create(ctx context.Context, doc *po.Document) error {
	// metadata 是 NOT NULL JSONB，空字符串会被 PG 拒绝（与 ConversationRepo.Create 一致）
	if doc.Metadata == "" {
		doc.Metadata = "{}"
	}
	return r.q.Document.WithContext(ctx).Create(doc)
}

func (r *DocumentRepo) GetByID(ctx context.Context, id int64) (*po.Document, error) {
	doc := r.q.Document
	result, err := doc.WithContext(ctx).Where(doc.ID.Eq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return result, err
}

func (r *DocumentRepo) GetList(ctx context.Context, kbID int64, page, pageSize int, status string) ([]*po.Document, int64, error) {
	doc := r.q.Document
	q := doc.WithContext(ctx).Where(doc.KnowledgeBaseID.Eq(kbID))

	if status != "" {
		q = q.Where(doc.Status.Eq(status))
	}

	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	list, err := q.Order(doc.CreatedAt.Desc()).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find()
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *DocumentRepo) Delete(ctx context.Context, id int64) error {
	doc := r.q.Document
	_, err := doc.WithContext(ctx).Where(doc.ID.Eq(id)).Delete()
	return err
}

func (r *DocumentRepo) UpdateStatus(ctx context.Context, id int64, status, errorMsg string) error {
	doc := r.q.Document
	_, err := doc.WithContext(ctx).Where(doc.ID.Eq(id)).Updates(map[string]any{
		"status":    status,
		"error_msg": errorMsg,
	})
	return err
}
