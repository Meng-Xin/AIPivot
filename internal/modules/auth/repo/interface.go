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
