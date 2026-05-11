// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package auth

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户注册 - 使用邮箱注册新账户（MVP 阶段自动绑定默认租户）
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.CommResponse, err error) {
	// 1. 校验
	if err = auth.ValidateEmail(req.Email); err != nil {
		return nil, errorx.NewInternalError(err.Error())
	}
	if err = auth.ValidatePassword(req.Password); err != nil {
		return nil, errorx.NewInternalError(err.Error())
	}

	// 3. 唯一性检查（MVP: 使用 default 租户）
	existing, err := l.svcCtx.UserRepo.GetByEmail(l.ctx, l.getDefaultTenantID(), req.Email)
	if err != nil {
		l.Logger.Errorf("Register GetByEmail err: %v", err)
		return nil, errorx.NewInternalError("注册失败")
	}
	if existing != nil {
		return nil, errorx.NewBusinessError(errorx.CodeFailed, "该邮箱已注册")
	}

	// 4. 加密密码 + 构造 PO
	encryptedPwd, err := auth.EncryptPassword(req.Password)
	if err != nil {
		l.Logger.Errorf("Register EncryptPassword err: %v", err)
		return nil, errorx.NewInternalError("注册失败")
	}
	userPo := auth.NewUserPo(req.NickName, req.Email, encryptedPwd, l.getDefaultTenantID())

	// 5. 持久化
	if err = l.svcCtx.UserRepo.Create(l.ctx, userPo); err != nil {
		l.Logger.Errorf("Register Create err: %v", err)
		return nil, errorx.NewInternalError("注册失败")
	}

	return &types.CommResponse{
		Code:      0,
		Msg:       "注册成功",
		Timestamp: time.Now().Unix(),
	}, nil
}

func (l *RegisterLogic) getDefaultTenantID() int64 {
	// MVP: 使用默认租户 ID=1
	return 1
}
