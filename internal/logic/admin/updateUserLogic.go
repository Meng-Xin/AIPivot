package admin

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UpdateUserLogic) UpdateUser(req *types.UpdateAdminUserRequest) (*types.CreateAdminUserResponse, error) {
	tenantID := auth.TenantIDFromContext(l.ctx)
	callerID := auth.UserIDFromContext(l.ctx)

	user, err := l.svcCtx.UserRepo.GetByID(l.ctx, req.ID)
	if err != nil || user == nil {
		return nil, errorx.NewNotFoundError("用户不存在")
	}

	// 只能操作同租户用户
	if user.TenantID != tenantID {
		return nil, errorx.NewForbidError("无权操作该用户")
	}

	// 不允许管理员降权/禁用自己，避免租户无管理员的情况
	if user.ID == callerID && (req.Status == "disabled" || req.Role == "member") {
		return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "不能修改自己的角色或禁用自己")
	}

	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	if err = l.svcCtx.UserRepo.Update(l.ctx, user); err != nil {
		l.Logger.Errorf("UpdateUser err: %v", err)
		return nil, errorx.NewInternalError("更新用户失败")
	}

	var lastLogin int64
	if user.LastLogin != nil {
		lastLogin = user.LastLogin.Unix()
	}

	return &types.CreateAdminUserResponse{
		Code:      0,
		Msg:       "更新成功",
		Timestamp: time.Now().Unix(),
		Data: types.ShowAdminUser{
			ID:        user.ID,
			UUID:      user.UUID,
			Email:     user.Email,
			NickName:  user.NickName,
			Role:      user.Role,
			Status:    user.Status,
			LastLogin: lastLogin,
			CreatedAt: user.CreatedAt.Unix(),
		},
	}, nil
}
