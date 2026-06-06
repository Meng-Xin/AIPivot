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

type GetTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantLogic {
	return &GetTenantLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GetTenantLogic) GetTenant() (*types.GetTenantResponse, error) {
	tenantID := auth.TenantIDFromContext(l.ctx)

	tenant, err := l.svcCtx.TenantRepo.GetByID(l.ctx, tenantID)
	if err != nil {
		return nil, errorx.NewInternalError("查询租户失败")
	}
	if tenant == nil {
		return nil, errorx.NewNotFoundError("租户不存在")
	}

	return &types.GetTenantResponse{
		Code:      0,
		Msg:       "OK",
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
