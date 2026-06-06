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

type DeleteSkillLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除工具
func NewDeleteSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSkillLogic {
	return &DeleteSkillLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteSkillLogic) DeleteSkill(req *types.DeleteSkillRequest) (resp *types.CommResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("unauthenticated")
	}

	// 确认记录存在且属于当前租户
	s, err := l.svcCtx.SkillRepo.GetByID(l.ctx, req.ID, tenantID)
	if err != nil {
		return nil, errorx.NewInternalError("查询工具失败")
	}
	if s == nil {
		return nil, errorx.NewBusinessError(errorx.CodeNotFound, "工具不存在")
	}

	if err := l.svcCtx.SkillRepo.Delete(l.ctx, req.ID, tenantID); err != nil {
		l.Logger.Errorf("DeleteSkill repo err: %v", err)
		return nil, errorx.NewInternalError("删除工具失败")
	}

	return &types.CommResponse{
		Code:      0,
		Msg:       "删除成功",
		Timestamp: time.Now().Unix(),
	}, nil
}
