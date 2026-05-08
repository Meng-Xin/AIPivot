package repo

import (
	"context"

	"aipivot/internal/modules/auth/repo/dao"
	"aipivot/internal/shared/po"
)

type UserRepo struct {
	userDao *dao.UserDao
}

func NewUserRepo(userDao *dao.UserDao) *UserRepo {
	return &UserRepo{userDao: userDao}
}

func (r *UserRepo) Create(ctx context.Context, user *po.User) error {
	return r.userDao.CreateUser(ctx, user)
}

func (r *UserRepo) GetByEmail(ctx context.Context, tenantID int64, email string) (*po.User, error) {
	return r.userDao.GetByEmail(ctx, tenantID, email)
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*po.User, error) {
	return r.userDao.GetByID(ctx, id)
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, id int64) error {
	return r.userDao.UpdateLastLogin(ctx, id)
}
