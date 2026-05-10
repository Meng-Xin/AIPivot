package repo

import (
	"context"

	"aipivot/internal/modules/auth/repo/dao"
	"aipivot/internal/shared/po"
)

type ApiKeyRepo struct {
	apiKeyDao *dao.ApiKeyDao
}

func NewApiKeyRepo(apiKeyDao *dao.ApiKeyDao) *ApiKeyRepo {
	return &ApiKeyRepo{apiKeyDao: apiKeyDao}
}

func (r *ApiKeyRepo) Create(ctx context.Context, key *po.ApiKey) error {
	return r.apiKeyDao.Create(ctx, key)
}

func (r *ApiKeyRepo) GetByKeyHash(ctx context.Context, keyHash string) (*po.ApiKey, error) {
	return r.apiKeyDao.GetByKeyHash(ctx, keyHash)
}

func (r *ApiKeyRepo) GetByID(ctx context.Context, id int64) (*po.ApiKey, error) {
	return r.apiKeyDao.GetByID(ctx, id)
}

func (r *ApiKeyRepo) GetListByTenant(ctx context.Context, tenantID int64) ([]*po.ApiKey, error) {
	return r.apiKeyDao.GetListByTenant(ctx, tenantID)
}

func (r *ApiKeyRepo) UpdateLastUsed(ctx context.Context, id int64) error {
	return r.apiKeyDao.UpdateLastUsed(ctx, id)
}

func (r *ApiKeyRepo) Revoke(ctx context.Context, id int64) error {
	return r.apiKeyDao.Revoke(ctx, id)
}

func (r *ApiKeyRepo) Delete(ctx context.Context, id int64) error {
	return r.apiKeyDao.Delete(ctx, id)
}
