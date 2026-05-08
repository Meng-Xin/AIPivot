package dao

import (
	"context"
	"time"

	"aipivot/internal/shared/po"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{db: db}
}

func (d *UserDao) WithTx(tx *gorm.DB) *UserDao {
	return &UserDao{db: tx}
}

func (d *UserDao) CreateUser(ctx context.Context, user *po.User) error {
	err := d.db.WithContext(ctx).Create(user).Error
	if err != nil {
		logx.WithContext(ctx).Errorf("CreateUser err: %v", err)
		return err
	}
	return nil
}

func (d *UserDao) GetByEmail(ctx context.Context, tenantID int64, email string) (*po.User, error) {
	var user po.User
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND email = ?", tenantID, email).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("GetByEmail err: %v", err)
		return nil, err
	}
	return &user, nil
}

func (d *UserDao) GetByID(ctx context.Context, id int64) (*po.User, error) {
	var user po.User
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("GetByID err: %v", err)
		return nil, err
	}
	return &user, nil
}

func (d *UserDao) UpdateLastLogin(ctx context.Context, id int64) error {
	err := d.db.WithContext(ctx).
		Model(&po.User{}).
		Where("id = ?", id).
		Update("last_login", time.Now()).Error
	if err != nil {
		logx.WithContext(ctx).Errorf("UpdateLastLogin err: %v", err)
		return err
	}
	return nil
}
