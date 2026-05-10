package repo

import (
	"context"

	"aipivot/internal/shared/po"
)

type UserRepository interface {
	Create(ctx context.Context, user *po.User) error
	GetByEmail(ctx context.Context, tenantID int64, email string) (*po.User, error)
	GetByID(ctx context.Context, id int64) (*po.User, error)
	UpdateLastLogin(ctx context.Context, id int64) error
}

type TenantRepository interface {
	GetBySlug(ctx context.Context, slug string) (*po.Tenant, error)
	GetByID(ctx context.Context, id int64) (*po.Tenant, error)
}

type ApiKeyRepository interface {
	Create(ctx context.Context, key *po.ApiKey) error
	GetByKeyHash(ctx context.Context, keyHash string) (*po.ApiKey, error)
	GetByID(ctx context.Context, id int64) (*po.ApiKey, error)
	GetListByTenant(ctx context.Context, tenantID int64) ([]*po.ApiKey, error)
	UpdateLastUsed(ctx context.Context, id int64) error
	Revoke(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
}
