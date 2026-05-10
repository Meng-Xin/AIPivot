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

type UserDao struct {
	q *query.Query
}

func NewUserDao(q *query.Query) *UserDao {
	return &UserDao{q: q}
}

func (d *UserDao) WithTx(tx *query.Query) *UserDao {
	return &UserDao{q: tx}
}

func (d *UserDao) CreateUser(ctx context.Context, user *po.User) error {
	err := d.q.User.WithContext(ctx).Create(user)
	if err != nil {
		logx.WithContext(ctx).Errorf("CreateUser err: %v", err)
		return err
	}
	return nil
}

func (d *UserDao) GetByEmail(ctx context.Context, tenantID int64, email string) (*po.User, error) {
	u := d.q.User
	user, err := u.WithContext(ctx).
		Where(u.TenantID.Eq(tenantID), u.Email.Eq(email)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("GetByEmail err: %v", err)
		return nil, err
	}
	return user, nil
}

func (d *UserDao) GetByID(ctx context.Context, id int64) (*po.User, error) {
	u := d.q.User
	user, err := u.WithContext(ctx).Where(u.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("GetByID err: %v", err)
		return nil, err
	}
	return user, nil
}

func (d *UserDao) UpdateLastLogin(ctx context.Context, id int64) error {
	u := d.q.User
	now := time.Now()
	_, err := u.WithContext(ctx).Where(u.ID.Eq(id)).Update(u.LastLogin, &now)
	if err != nil {
		logx.WithContext(ctx).Errorf("UpdateLastLogin err: %v", err)
		return err
	}
	return nil
}
