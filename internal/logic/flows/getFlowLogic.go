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

type GetFlowLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取 Flow 详情
func NewGetFlowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFlowLogic {
	return &GetFlowLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFlowLogic) GetFlow(req *types.GetFlowRequest) (resp *types.FlowResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("unauthenticated")
	}

	flow, err := l.svcCtx.FlowRepo.GetByID(l.ctx, req.ID, tenantID)
	if err != nil {
		l.Logger.Errorf("GetFlow repo err: %v", err)
		return nil, errorx.NewInternalError("获取 Flow 失败")
	}
	if flow == nil {
		return nil, errorx.NewNotFoundError("Flow 不存在")
	}

	return &types.FlowResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data:      toShowFlow(flow),
	}, nil
}
