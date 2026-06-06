package flow

import (
	"context"
	"errors"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// Repository Flow 数据仓储接口。
type Repository interface {
	Create(ctx context.Context, f *po.Flow) error
	GetByID(ctx context.Context, id, tenantID int64) (*po.Flow, error)
	GetListByTenant(ctx context.Context, tenantID int64) ([]*po.Flow, error)
	Update(ctx context.Context, f *po.Flow) error
	Delete(ctx context.Context, id, tenantID int64) error
}

type FlowRepo struct {
	q *query.Query
}

func NewFlowRepo(q *query.Query) *FlowRepo {
	return &FlowRepo{q: q}
}

func (r *FlowRepo) Create(ctx context.Context, f *po.Flow) error {
	if err := r.q.Flow.WithContext(ctx).Create(f); err != nil {
		logx.WithContext(ctx).Errorf("FlowRepo.Create err: %v", err)
		return err
	}
	return nil
}

func (r *FlowRepo) GetByID(ctx context.Context, id, tenantID int64) (*po.Flow, error) {
	fl := r.q.Flow
	result, err := fl.WithContext(ctx).
		Where(fl.ID.Eq(id), fl.TenantID.Eq(tenantID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("FlowRepo.GetByID err: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *FlowRepo) GetListByTenant(ctx context.Context, tenantID int64) ([]*po.Flow, error) {
	fl := r.q.Flow
	list, err := fl.WithContext(ctx).
		Where(fl.TenantID.Eq(tenantID)).
		Order(fl.UpdatedAt.Desc()).
		Find()
	if err != nil {
		logx.WithContext(ctx).Errorf("FlowRepo.GetListByTenant err: %v", err)
		return nil, err
	}
	return list, nil
}

func (r *FlowRepo) Update(ctx context.Context, f *po.Flow) error {
	_, err := r.q.Flow.WithContext(ctx).
		Where(r.q.Flow.ID.Eq(f.ID), r.q.Flow.TenantID.Eq(f.TenantID)).
		Updates(f)
	if err != nil {
		logx.WithContext(ctx).Errorf("FlowRepo.Update err: %v", err)
		return err
	}
	return nil
}

func (r *FlowRepo) Delete(ctx context.Context, id, tenantID int64) error {
	fl := r.q.Flow
	_, err := fl.WithContext(ctx).
		Where(fl.ID.Eq(id), fl.TenantID.Eq(tenantID)).
		Delete()
	if err != nil {
		logx.WithContext(ctx).Errorf("FlowRepo.Delete err: %v", err)
		return err
	}
	return nil
}
