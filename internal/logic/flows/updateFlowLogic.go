// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package flows

import (
	"context"
	"strings"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFlowLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新 Flow
func NewUpdateFlowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFlowLogic {
	return &UpdateFlowLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFlowLogic) UpdateFlow(req *types.UpdateFlowRequest) (resp *types.FlowResponse, err error) {
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

	if strings.TrimSpace(req.Name) != "" {
		flow.Name = strings.TrimSpace(req.Name)
	}
	if req.Description != "" {
		flow.Description = strings.TrimSpace(req.Description)
	}
	if req.Definition != "" {
		definition, err := parseDefinition(req.Definition)
		if err != nil {
			return nil, err
		}
		flow.Definition = definition
	}
	if req.Status != "" {
		status, err := normalizeStatus(req.Status)
		if err != nil {
			return nil, err
		}
		flow.Status = status
	}
	flow.Version++

	if err := l.svcCtx.FlowRepo.Update(l.ctx, flow); err != nil {
		l.Logger.Errorf("UpdateFlow repo err: %v", err)
		return nil, errorx.NewInternalError("更新 Flow 失败")
	}

	return &types.FlowResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data:      toShowFlow(flow),
	}, nil
}
