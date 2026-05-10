package dao

import (
	"context"
	"errors"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type KnowledgeBaseDao struct {
	q *query.Query
}

func NewKnowledgeBaseDao(q *query.Query) *KnowledgeBaseDao {
	return &KnowledgeBaseDao{q: q}
}

func (d *KnowledgeBaseDao) WithTx(tx *query.Query) *KnowledgeBaseDao {
	return &KnowledgeBaseDao{q: tx}
}

func (d *KnowledgeBaseDao) Create(ctx context.Context, kb *po.KnowledgeBase) error {
	err := d.q.KnowledgeBase.WithContext(ctx).Create(kb)
	if err != nil {
		logx.WithContext(ctx).Errorf("KnowledgeBaseDao.Create err: %v", err)
		return err
	}
	return nil
}

func (d *KnowledgeBaseDao) GetByID(ctx context.Context, id int64) (*po.KnowledgeBase, error) {
	kb := d.q.KnowledgeBase
	result, err := kb.WithContext(ctx).Where(kb.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("KnowledgeBaseDao.GetByID err: %v", err)
		return nil, err
	}
	return result, nil
}

func (d *KnowledgeBaseDao) GetList(ctx context.Context, tenantID int64, page, pageSize int, name string) ([]*po.KnowledgeBase, int64, error) {
	kb := d.q.KnowledgeBase
	q := kb.WithContext(ctx).Where(kb.TenantID.Eq(tenantID))

	if name != "" {
		q = q.Where(kb.Name.Like("%" + name + "%"))
	}

	total, err := q.Count()
	if err != nil {
		logx.WithContext(ctx).Errorf("KnowledgeBaseDao.GetList count err: %v", err)
		return nil, 0, err
	}

	list, err := q.Order(kb.CreatedAt.Desc()).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find()
	if err != nil {
		logx.WithContext(ctx).Errorf("KnowledgeBaseDao.GetList find err: %v", err)
		return nil, 0, err
	}

	return list, total, nil
}

func (d *KnowledgeBaseDao) Update(ctx context.Context, id int64, updates map[string]any) error {
	kb := d.q.KnowledgeBase
	_, err := kb.WithContext(ctx).Where(kb.ID.Eq(id)).Updates(updates)
	if err != nil {
		logx.WithContext(ctx).Errorf("KnowledgeBaseDao.Update err: %v", err)
		return err
	}
	return nil
}

func (d *KnowledgeBaseDao) Delete(ctx context.Context, id int64) error {
	kb := d.q.KnowledgeBase
	_, err := kb.WithContext(ctx).Where(kb.ID.Eq(id)).Delete()
	if err != nil {
		logx.WithContext(ctx).Errorf("KnowledgeBaseDao.Delete err: %v", err)
		return err
	}
	return nil
}
