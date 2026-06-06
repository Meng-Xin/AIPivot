// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package flows

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFlowsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取 Flow 列表
func NewListFlowsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFlowsLogic {
	return &ListFlowsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListFlowsLogic) ListFlows() (resp *types.FlowListResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("unauthenticated")
	}

	list, err := l.svcCtx.FlowRepo.GetListByTenant(l.ctx, tenantID)
	if err != nil {
		l.Logger.Errorf("ListFlows repo err: %v", err)
		return nil, errorx.NewInternalError("获取 Flow 列表失败")
	}

	return &types.FlowListResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data:      toShowFlows(list),
	}, nil
}
