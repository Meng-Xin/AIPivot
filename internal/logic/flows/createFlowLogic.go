// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package flows

import (
	"context"
	"strings"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/po"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateFlowLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建 Flow
func NewCreateFlowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFlowLogic {
	return &CreateFlowLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateFlowLogic) CreateFlow(req *types.CreateFlowRequest) (resp *types.FlowResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("unauthenticated")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "name 不能为空")
	}

	definition, err := parseDefinition(req.Definition)
	if err != nil {
		return nil, err
	}
	status, err := normalizeStatus(req.Status)
	if err != nil {
		return nil, err
	}

	flow := &po.Flow{
		UUID:        uuid.New().String(),
		TenantID:    tenantID,
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		Definition:  definition,
		Status:      status,
		Version:     1,
	}

	if err := l.svcCtx.FlowRepo.Create(l.ctx, flow); err != nil {
		l.Logger.Errorf("CreateFlow repo err: %v", err)
		return nil, errorx.NewInternalError("创建 Flow 失败")
	}

	return &types.FlowResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data:      toShowFlow(flow),
	}, nil
}
