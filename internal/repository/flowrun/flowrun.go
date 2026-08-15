package flowrun

import (
	"context"
	"errors"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// ErrRunNotFound Finish 在 RowsAffected==0 时返回的哨兵错误，
// 由 logic 层翻译为业务错误（执行记录不存在 / 越权访问）。
var ErrRunNotFound = errors.New("flow run not found")

// Repository FlowRun 数据仓储接口。
type Repository interface {
	Create(ctx context.Context, r *po.FlowRun) error
	GetByID(ctx context.Context, id, tenantID int64) (*po.FlowRun, error)
	ListByFlow(ctx context.Context, flowID, tenantID int64, limit, offset int) ([]*po.FlowRun, error)
	// Finish 按 id + tenantID 更新执行收尾字段（status/output/node_results/error/total_ms/token_count）。
	// RowsAffected==0 时返回 ErrRunNotFound。
	Finish(ctx context.Context, id, tenantID int64, updates map[string]any) error
}

type FlowRunRepo struct {
	q *query.Query
}

func NewFlowRunRepo(q *query.Query) *FlowRunRepo {
	return &FlowRunRepo{q: q}
}

func (r *FlowRunRepo) Create(ctx context.Context, run *po.FlowRun) error {
	if run.NodeResults == "" {
		run.NodeResults = "[]"
	}
	if err := r.q.FlowRun.WithContext(ctx).Create(run); err != nil {
		logx.WithContext(ctx).Errorf("FlowRunRepo.Create err: %v", err)
		return err
	}
	return nil
}

func (r *FlowRunRepo) GetByID(ctx context.Context, id, tenantID int64) (*po.FlowRun, error) {
	fr := r.q.FlowRun
	result, err := fr.WithContext(ctx).
		Where(fr.ID.Eq(id), fr.TenantID.Eq(tenantID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("FlowRunRepo.GetByID err: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *FlowRunRepo) ListByFlow(ctx context.Context, flowID, tenantID int64, limit, offset int) ([]*po.FlowRun, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	fr := r.q.FlowRun
	list, err := fr.WithContext(ctx).
		Where(fr.FlowID.Eq(flowID), fr.TenantID.Eq(tenantID)).
		Order(fr.CreatedAt.Desc()).
		Offset(offset).
		Limit(limit).
		Find()
	if err != nil {
		logx.WithContext(ctx).Errorf("FlowRunRepo.ListByFlow err: %v", err)
		return nil, err
	}
	return list, nil
}

func (r *FlowRunRepo) Finish(ctx context.Context, id, tenantID int64, updates map[string]any) error {
	fr := r.q.FlowRun
	info, err := fr.WithContext(ctx).
		Where(fr.ID.Eq(id), fr.TenantID.Eq(tenantID)).
		UpdateColumns(updates)
	if err != nil {
		logx.WithContext(ctx).Errorf("FlowRunRepo.Finish err: %v", err)
		return err
	}
	if info.RowsAffected == 0 {
		return ErrRunNotFound
	}
	return nil
}
