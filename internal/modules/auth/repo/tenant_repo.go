package repo

import (
	"context"

	"aipivot/internal/modules/auth/repo/dao"
	"aipivot/internal/shared/po"
)

type TenantRepo struct {
	tenantDao *dao.TenantDao
}

func NewTenantRepo(tenantDao *dao.TenantDao) *TenantRepo {
	return &TenantRepo{tenantDao: tenantDao}
}

func (r *TenantRepo) GetBySlug(ctx context.Context, slug string) (*po.Tenant, error) {
	return r.tenantDao.GetBySlug(ctx, slug)
}

func (r *TenantRepo) GetByID(ctx context.Context, id int64) (*po.Tenant, error) {
	return r.tenantDao.GetByID(ctx, id)
}
