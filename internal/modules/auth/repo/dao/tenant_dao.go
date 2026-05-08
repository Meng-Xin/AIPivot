package dao

import (
	"context"

	"aipivot/internal/shared/po"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type TenantDao struct {
	db *gorm.DB
}

func NewTenantDao(db *gorm.DB) *TenantDao {
	return &TenantDao{db: db}
}

func (d *TenantDao) WithTx(tx *gorm.DB) *TenantDao {
	return &TenantDao{db: tx}
}

func (d *TenantDao) GetBySlug(ctx context.Context, slug string) (*po.Tenant, error) {
	var tenant po.Tenant
	err := d.db.WithContext(ctx).Where("slug = ?", slug).First(&tenant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("GetBySlug err: %v", err)
		return nil, err
	}
	return &tenant, nil
}

func (d *TenantDao) GetByID(ctx context.Context, id int64) (*po.Tenant, error) {
	var tenant po.Tenant
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&tenant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("GetByID err: %v", err)
		return nil, err
	}
	return &tenant, nil
}
