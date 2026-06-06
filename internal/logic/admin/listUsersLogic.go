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

type ListUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListUsersLogic) ListUsers(req *types.ListAdminUsersRequest) (*types.ListAdminUsersResponse, error) {
	tenantID := auth.TenantIDFromContext(l.ctx)

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := l.svcCtx.UserRepo.GetListByTenant(l.ctx, tenantID, page, pageSize)
	if err != nil {
		l.Logger.Errorf("ListUsers err: %v", err)
		return nil, errorx.NewInternalError("获取用户列表失败")
	}

	list := make([]types.ShowAdminUser, 0, len(users))
	for _, u := range users {
		var lastLogin int64
		if u.LastLogin != nil {
			lastLogin = u.LastLogin.Unix()
		}
		list = append(list, types.ShowAdminUser{
			ID:        u.ID,
			UUID:      u.UUID,
			Email:     u.Email,
			NickName:  u.NickName,
			Role:      u.Role,
			Status:    u.Status,
			LastLogin: lastLogin,
			CreatedAt: u.CreatedAt.Unix(),
		})
	}

	return &types.ListAdminUsersResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data: types.AdminUsersData{
			Total: total,
			List:  list,
		},
	}, nil
}
