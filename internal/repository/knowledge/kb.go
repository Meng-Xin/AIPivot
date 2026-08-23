package knowledge

import (
	"context"
	"errors"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"gorm.io/gorm"
)

// KBRepo 知识库数据仓储实现。
type KBRepo struct {
	q *query.Query
}

func NewKBRepo(q *query.Query) *KBRepo {
	return &KBRepo{q: q}
}

func (r *KBRepo) Create(ctx context.Context, kb *po.KnowledgeBase) error {
	// JSONB 列不接受空字符串，兜底成空对象 / 空数组（与 MessageRepo.Create 一致）
	if kb.Settings == "" {
		kb.Settings = "{}"
	}
	if kb.SuggestedQuestions == "" {
		kb.SuggestedQuestions = "[]"
	}
	return r.q.KnowledgeBase.WithContext(ctx).Create(kb)
}

func (r *KBRepo) GetByID(ctx context.Context, id int64) (*po.KnowledgeBase, error) {
	kb := r.q.KnowledgeBase
	result, err := kb.WithContext(ctx).Where(kb.ID.Eq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return result, err
}

func (r *KBRepo) GetList(ctx context.Context, tenantID int64, page, pageSize int, name string) ([]*po.KnowledgeBase, int64, error) {
	kb := r.q.KnowledgeBase
	q := kb.WithContext(ctx).Where(kb.TenantID.Eq(tenantID))

	if name != "" {
		q = q.Where(kb.Name.Like("%" + name + "%"))
	}

	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	list, err := q.Order(kb.CreatedAt.Desc()).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find()
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *KBRepo) Update(ctx context.Context, id int64, updates map[string]any) error {
	kb := r.q.KnowledgeBase
	_, err := kb.WithContext(ctx).Where(kb.ID.Eq(id)).Updates(updates)
	return err
}

func (r *KBRepo) Delete(ctx context.Context, id int64) error {
	kb := r.q.KnowledgeBase
	_, err := kb.WithContext(ctx).Where(kb.ID.Eq(id)).Delete()
	return err
}
