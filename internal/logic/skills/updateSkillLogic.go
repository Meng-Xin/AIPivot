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

type UpdateSkillLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新工具
func NewUpdateSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSkillLogic {
	return &UpdateSkillLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateSkillLogic) UpdateSkill(req *types.UpdateSkillRequest) (resp *types.SkillResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("unauthenticated")
	}

	// 先查询确保存在且属于当前租户
	s, err := l.svcCtx.SkillRepo.GetByID(l.ctx, req.ID, tenantID)
	if err != nil {
		return nil, errorx.NewInternalError("查询工具失败")
	}
	if s == nil {
		return nil, errorx.NewBusinessError(errorx.CodeNotFound, "工具不存在")
	}

	// 只更新请求中非零値字段
	if req.Name != "" {
		s.Name = req.Name
	}
	if req.Description != "" {
		s.Description = req.Description
	}
	if req.Endpoint != "" {
		s.Endpoint = req.Endpoint
	}
	if req.Method != "" {
		s.Method = strings.ToUpper(req.Method)
	}
	if req.TimeoutMs > 0 {
		s.TimeoutMs = req.TimeoutMs
	}
	// Enabled 是 bool 无法区分“未设置” vs “false”，如果 payload 中包含 enabled 字段则拧覆
	s.Enabled = req.Enabled

	if req.Parameters != "" {
		var params po.JSONMap
		if err := json.Unmarshal([]byte(req.Parameters), &params); err != nil {
			return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "parameters 不是有效的 JSON")
		}
		s.Parameters = params
	}
	if req.Headers != "" {
		var headers po.JSONMap
		if err := json.Unmarshal([]byte(req.Headers), &headers); err != nil {
			return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "headers 不是有效的 JSON")
		}
		s.Headers = headers
	}

	if err := l.svcCtx.SkillRepo.Update(l.ctx, s); err != nil {
		l.Logger.Errorf("UpdateSkill repo err: %v", err)
		return nil, errorx.NewInternalError("更新工具失败")
	}

	return &types.SkillResponse{
		Code:      0,
		Msg:       "ok",
		Timestamp: time.Now().Unix(),
		Data:      toShowSkill(s),
	}, nil
}
