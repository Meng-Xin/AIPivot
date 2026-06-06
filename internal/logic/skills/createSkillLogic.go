// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package skills

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/po"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSkillLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建自定义工具
func NewCreateSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSkillLogic {
	return &CreateSkillLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSkillLogic) CreateSkill(req *types.CreateSkillRequest) (resp *types.SkillResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("unauthenticated")
	}

	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Endpoint) == "" {
		return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "name 和 endpoint 不能为空")
	}

	// 解析 Parameters JSON string 转 JSONMap
	var params po.JSONMap
	if req.Parameters != "" {
		if err := json.Unmarshal([]byte(req.Parameters), &params); err != nil {
			return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "parameters 不是有效的 JSON")
		}
	}

	var headers po.JSONMap
	if req.Headers != "" {
		if err := json.Unmarshal([]byte(req.Headers), &headers); err != nil {
			return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "headers 不是有效的 JSON")
		}
	}

	method := strings.ToUpper(req.Method)
	if method == "" {
		method = "POST"
	}
	timeoutMs := req.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 5000
	}

	s := &po.Skill{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Parameters:  params,
		Endpoint:    req.Endpoint,
		Method:      method,
		Headers:     headers,
		TimeoutMs:   timeoutMs,
		Enabled:     req.Enabled,
	}

	if err := l.svcCtx.SkillRepo.Create(l.ctx, s); err != nil {
		l.Logger.Errorf("CreateSkill repo err: %v", err)
		return nil, errorx.NewInternalError("创建工具失败")
	}

	return &types.SkillResponse{
		Code:      0,
		Msg:       "ok",
		Timestamp: time.Now().Unix(),
		Data:      toShowSkill(s),
	}, nil
}
