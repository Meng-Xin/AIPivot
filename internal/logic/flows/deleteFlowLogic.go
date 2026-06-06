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

type DeleteFlowLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除 Flow
func NewDeleteFlowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFlowLogic {
	return &DeleteFlowLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFlowLogic) DeleteFlow(req *types.DeleteFlowRequest) (resp *types.CommResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("unauthenticated")
	}

	flow, err := l.svcCtx.FlowRepo.GetByID(l.ctx, req.ID, tenantID)
	if err != nil {
		return nil, errorx.NewInternalError("查询 Flow 失败")
	}
	if flow == nil {
		return nil, errorx.NewNotFoundError("Flow 不存在")
	}

	if err := l.svcCtx.FlowRepo.Delete(l.ctx, req.ID, tenantID); err != nil {
		l.Logger.Errorf("DeleteFlow repo err: %v", err)
		return nil, errorx.NewInternalError("删除 Flow 失败")
	}

	return &types.CommResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
	}, nil
}
