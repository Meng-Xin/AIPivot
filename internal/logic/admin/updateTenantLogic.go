package admin

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantLogic {
	return &UpdateTenantLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UpdateTenantLogic) UpdateTenant(req *types.UpdateTenantRequest) (*types.UpdateTenantResponse, error) {
	tenantID := auth.TenantIDFromContext(l.ctx)

	tenant, err := l.svcCtx.TenantRepo.GetByID(l.ctx, tenantID)
	if err != nil || tenant == nil {
		return nil, errorx.NewNotFoundError("租户不存在")
	}

	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.Plan != "" {
		tenant.Plan = req.Plan
	}

	if err = l.svcCtx.TenantRepo.Update(l.ctx, tenant); err != nil {
		l.Logger.Errorf("UpdateTenant err: %v", err)
		return nil, errorx.NewInternalError("更新租户失败")
	}

	// 重新查询以取得最新 UpdatedAt
	tenant, _ = l.svcCtx.TenantRepo.GetByID(l.ctx, tenantID)

	return &types.UpdateTenantResponse{
		Code:      0,
		Msg:       "更新成功",
		Timestamp: time.Now().Unix(),
		Data: types.ShowTenant{
			ID:        tenant.ID,
			UUID:      tenant.UUID,
			Name:      tenant.Name,
			Slug:      tenant.Slug,
			Plan:      tenant.Plan,
			Status:    tenant.Status,
			CreatedAt: tenant.CreatedAt.Unix(),
			UpdatedAt: tenant.UpdatedAt.Unix(),
		},
	}, nil
}
