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

type ListSkillsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取工具列表
func NewListSkillsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSkillsLogic {
	return &ListSkillsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListSkillsLogic) ListSkills() (resp *types.SkillListResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("unauthenticated")
	}

	list, err := l.svcCtx.SkillRepo.GetListByTenant(l.ctx, tenantID)
	if err != nil {
		l.Logger.Errorf("ListSkills repo err: %v", err)
		return nil, errorx.NewInternalError("获取工具列表失败")
	}

	return &types.SkillListResponse{
		Code:      0,
		Msg:       "ok",
		Timestamp: time.Now().Unix(),
		Data:      toShowSkills(list),
	}, nil
}
