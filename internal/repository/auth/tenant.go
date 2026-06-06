package auth

import (
	"context"
	"errors"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"gorm.io/gorm"
)

// TenantRepo 租户数据仓储实现。
type TenantRepo struct {
	q *query.Query
}

func NewTenantRepo(q *query.Query) *TenantRepo {
	return &TenantRepo{q: q}
}

func (r *TenantRepo) GetBySlug(ctx context.Context, slug string) (*po.Tenant, error) {
	t := r.q.Tenant
	tenant, err := t.WithContext(ctx).Where(t.Slug.Eq(slug)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return tenant, err
}

func (r *TenantRepo) GetByID(ctx context.Context, id int64) (*po.Tenant, error) {
	t := r.q.Tenant
	tenant, err := t.WithContext(ctx).Where(t.ID.Eq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return tenant, err
}

func (r *TenantRepo) Update(ctx context.Context, tenant *po.Tenant) error {
	t := r.q.Tenant
	_, err := t.WithContext(ctx).Where(t.ID.Eq(tenant.ID)).Updates(tenant)
	return err
}
