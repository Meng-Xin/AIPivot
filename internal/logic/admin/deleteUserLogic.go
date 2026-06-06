package admin

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/response"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DeleteUserLogic) DeleteUser(req *types.DeleteAdminUserRequest) (*response.CommResponse, error) {
	tenantID := auth.TenantIDFromContext(l.ctx)
	callerID := auth.UserIDFromContext(l.ctx)

	if req.ID == callerID {
		return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "不能删除自己")
	}

	user, err := l.svcCtx.UserRepo.GetByID(l.ctx, req.ID)
	if err != nil || user == nil {
		return nil, errorx.NewNotFoundError("用户不存在")
	}

	// 只能删除同租户用户
	if user.TenantID != tenantID {
		return nil, errorx.NewForbidError("无权操作该用户")
	}

	if err = l.svcCtx.UserRepo.Delete(l.ctx, req.ID); err != nil {
		l.Logger.Errorf("DeleteUser err: %v", err)
		return nil, errorx.NewInternalError("删除用户失败")
	}

	return &response.CommResponse{
		Code:      0,
		Msg:       "删除成功",
		Timestamp: time.Now().Unix(),
	}, nil
}
