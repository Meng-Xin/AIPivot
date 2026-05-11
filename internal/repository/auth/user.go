package auth

import (
	"context"
	"errors"
	"time"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"gorm.io/gorm"
)

// UserRepo 用户数据仓储实现。
type UserRepo struct {
	q *query.Query
}

func NewUserRepo(q *query.Query) *UserRepo {
	return &UserRepo{q: q}
}

func (r *UserRepo) Create(ctx context.Context, user *po.User) error {
	return r.q.User.WithContext(ctx).Create(user)
}

func (r *UserRepo) GetByEmail(ctx context.Context, tenantID int64, email string) (*po.User, error) {
	u := r.q.User
	user, err := u.WithContext(ctx).
		Where(u.TenantID.Eq(tenantID), u.Email.Eq(email)).
		First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return user, err
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*po.User, error) {
	u := r.q.User
	user, err := u.WithContext(ctx).Where(u.ID.Eq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return user, err
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, id int64) error {
	u := r.q.User
	now := time.Now()
	_, err := u.WithContext(ctx).Where(u.ID.Eq(id)).Update(u.LastLogin, &now)
	return err
}
