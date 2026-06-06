package admin

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/po"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CreateUserLogic) CreateUser(req *types.CreateAdminUserRequest) (*types.CreateAdminUserResponse, error) {
	tenantID := auth.TenantIDFromContext(l.ctx)

	if err := auth.ValidateEmail(req.Email); err != nil {
		return nil, errorx.NewBusinessError(errorx.CodeBadRequest, err.Error())
	}
	if err := auth.ValidatePassword(req.Password); err != nil {
		return nil, errorx.NewBusinessError(errorx.CodeBadRequest, err.Error())
	}

	// 检查邮箱在租户内是否已存在
	existing, err := l.svcCtx.UserRepo.GetByEmail(l.ctx, tenantID, req.Email)
	if err != nil {
		return nil, errorx.NewInternalError("查询用户失败")
	}
	if existing != nil {
		return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "该邮箱已注册")
	}

	hashed, err := auth.EncryptPassword(req.Password)
	if err != nil {
		return nil, errorx.NewInternalError("密码加密失败")
	}

	role := req.Role
	if role == "" {
		role = "member"
	}

	user := &po.User{
		UUID:     uuid.New().String(),
		TenantID: tenantID,
		Email:    req.Email,
		NickName: req.NickName,
		Password: hashed,
		Role:     role,
		Status:   "active",
	}

	if err = l.svcCtx.UserRepo.Create(l.ctx, user); err != nil {
		l.Logger.Errorf("CreateUser err: %v", err)
		return nil, errorx.NewInternalError("创建用户失败")
	}

	return &types.CreateAdminUserResponse{
		Code:      0,
		Msg:       "创建成功",
		Timestamp: time.Now().Unix(),
		Data: types.ShowAdminUser{
			ID:        user.ID,
			UUID:      user.UUID,
			Email:     user.Email,
			NickName:  user.NickName,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt.Unix(),
		},
	}, nil
}
