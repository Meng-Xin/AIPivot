package dao

import (
	"context"
	"errors"
	"time"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ApiKeyDao struct {
	q *query.Query
}

func NewApiKeyDao(q *query.Query) *ApiKeyDao {
	return &ApiKeyDao{q: q}
}

func (d *ApiKeyDao) WithTx(tx *query.Query) *ApiKeyDao {
	return &ApiKeyDao{q: tx}
}

func (d *ApiKeyDao) Create(ctx context.Context, key *po.ApiKey) error {
	err := d.q.ApiKey.WithContext(ctx).Create(key)
	if err != nil {
		logx.WithContext(ctx).Errorf("ApiKeyDao.Create err: %v", err)
		return err
	}
	return nil
}

// GetByKeyHash 根据密钥哈希查找 API Key（用于认证时的快速查找）。
func (d *ApiKeyDao) GetByKeyHash(ctx context.Context, keyHash string) (*po.ApiKey, error) {
	k := d.q.ApiKey
	result, err := k.WithContext(ctx).
		Where(k.KeyHash.Eq(keyHash), k.Status.Eq("active")).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("ApiKeyDao.GetByKeyHash err: %v", err)
		return nil, err
	}
	return result, nil
}

func (d *ApiKeyDao) GetByID(ctx context.Context, id int64) (*po.ApiKey, error) {
	k := d.q.ApiKey
	result, err := k.WithContext(ctx).Where(k.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("ApiKeyDao.GetByID err: %v", err)
		return nil, err
	}
	return result, nil
}

func (d *ApiKeyDao) GetListByTenant(ctx context.Context, tenantID int64) ([]*po.ApiKey, error) {
	k := d.q.ApiKey
	list, err := k.WithContext(ctx).
		Where(k.TenantID.Eq(tenantID)).
		Order(k.CreatedAt.Desc()).
		Find()
	if err != nil {
		logx.WithContext(ctx).Errorf("ApiKeyDao.GetListByTenant err: %v", err)
		return nil, err
	}
	return list, nil
}

// UpdateLastUsed 更新密钥的最近使用时间（认证成功时异步调用）。
func (d *ApiKeyDao) UpdateLastUsed(ctx context.Context, id int64) error {
	k := d.q.ApiKey
	now := time.Now()
	_, err := k.WithContext(ctx).Where(k.ID.Eq(id)).Update(k.LastUsed, &now)
	if err != nil {
		logx.WithContext(ctx).Errorf("ApiKeyDao.UpdateLastUsed err: %v", err)
		return err
	}
	return nil
}

func (d *ApiKeyDao) Revoke(ctx context.Context, id int64) error {
	k := d.q.ApiKey
	_, err := k.WithContext(ctx).Where(k.ID.Eq(id)).Update(k.Status, "revoked")
	if err != nil {
		logx.WithContext(ctx).Errorf("ApiKeyDao.Revoke err: %v", err)
		return err
	}
	return nil
}

func (d *ApiKeyDao) Delete(ctx context.Context, id int64) error {
	k := d.q.ApiKey
	_, err := k.WithContext(ctx).Where(k.ID.Eq(id)).Delete()
	if err != nil {
		logx.WithContext(ctx).Errorf("ApiKeyDao.Delete err: %v", err)
		return err
	}
	return nil
}
