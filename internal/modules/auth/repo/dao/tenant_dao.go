package dao

import (
	"context"
	"errors"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type TenantDao struct {
	q *query.Query
}

func NewTenantDao(q *query.Query) *TenantDao {
	return &TenantDao{q: q}
}

func (d *TenantDao) WithTx(tx *query.Query) *TenantDao {
	return &TenantDao{q: tx}
}

func (d *TenantDao) GetBySlug(ctx context.Context, slug string) (*po.Tenant, error) {
	t := d.q.Tenant
	tenant, err := t.WithContext(ctx).Where(t.Slug.Eq(slug)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("GetBySlug err: %v", err)
		return nil, err
	}
	return tenant, nil
}

func (d *TenantDao) GetByID(ctx context.Context, id int64) (*po.Tenant, error) {
	t := d.q.Tenant
	tenant, err := t.WithContext(ctx).Where(t.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("GetByID err: %v", err)
		return nil, err
	}
	return tenant, nil
}
