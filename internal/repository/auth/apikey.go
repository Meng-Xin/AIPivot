package auth

import (
	"context"
	"errors"
	"time"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"gorm.io/gorm"
)

// ApiKeyRepo API Key 数据仓储实现。
type ApiKeyRepo struct {
	q *query.Query
}

func NewApiKeyRepo(q *query.Query) *ApiKeyRepo {
	return &ApiKeyRepo{q: q}
}

func (r *ApiKeyRepo) Create(ctx context.Context, key *po.ApiKey) error {
	return r.q.ApiKey.WithContext(ctx).Create(key)
}

// GetByKeyHash 根据密钥哈希查找 API Key（用于认证时的快速查找）。
func (r *ApiKeyRepo) GetByKeyHash(ctx context.Context, keyHash string) (*po.ApiKey, error) {
	k := r.q.ApiKey
	result, err := k.WithContext(ctx).
		Where(k.KeyHash.Eq(keyHash), k.Status.Eq("active")).
		First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return result, err
}

func (r *ApiKeyRepo) GetByID(ctx context.Context, id int64) (*po.ApiKey, error) {
	k := r.q.ApiKey
	result, err := k.WithContext(ctx).Where(k.ID.Eq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return result, err
}

func (r *ApiKeyRepo) GetListByTenant(ctx context.Context, tenantID int64) ([]*po.ApiKey, error) {
	k := r.q.ApiKey
	return k.WithContext(ctx).
		Where(k.TenantID.Eq(tenantID)).
		Order(k.CreatedAt.Desc()).
		Find()
}

// UpdateLastUsed 更新密钥的最近使用时间（认证成功时异步调用）。
func (r *ApiKeyRepo) UpdateLastUsed(ctx context.Context, id int64) error {
	k := r.q.ApiKey
	now := time.Now()
	_, err := k.WithContext(ctx).Where(k.ID.Eq(id)).Update(k.LastUsed, &now)
	return err
}

func (r *ApiKeyRepo) Revoke(ctx context.Context, id int64) error {
	k := r.q.ApiKey
	_, err := k.WithContext(ctx).Where(k.ID.Eq(id)).Update(k.Status, "revoked")
	return err
}

func (r *ApiKeyRepo) Delete(ctx context.Context, id int64) error {
	k := r.q.ApiKey
	_, err := k.WithContext(ctx).Where(k.ID.Eq(id)).Delete()
	return err
}
