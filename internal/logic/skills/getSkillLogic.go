// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package skills

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSkillLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取工具详情
func NewGetSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSkillLogic {
	return &GetSkillLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSkillLogic) GetSkill(req *types.GetSkillRequest) (resp *types.SkillResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("unauthenticated")
	}

	s, err := l.svcCtx.SkillRepo.GetByID(l.ctx, req.ID, tenantID)
	if err != nil {
		return nil, errorx.NewInternalError("查询工具失败")
	}
	if s == nil {
		return nil, errorx.NewBusinessError(errorx.CodeNotFound, "工具不存在")
	}

	return &types.SkillResponse{
		Code:      0,
		Msg:       "ok",
		Timestamp: time.Now().Unix(),
		Data:      toShowSkill(s),
	}, nil
}
