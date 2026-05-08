package logic

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/modules/auth/domain/assembler"
	"aipivot/internal/modules/auth/repo"
	authtypes "aipivot/internal/modules/auth/types"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *authtypes.LoginRequest) (resp *authtypes.LoginResponse, err error) {
	// 1. Request → Domain Model
	userModel := assembler.LoginRequestToModelUser(req)

	// 2. 领域校验
	if err = userModel.CheckEmail(); err != nil {
		return nil, errorx.NewInternalError(err.Error())
	}
	if err = userModel.CheckPassword(); err != nil {
		return nil, errorx.NewInternalError(err.Error())
	}

	// 3. 通过 Repo 查询用户（MVP: 使用 default 租户）
	user, err := l.getUserRepo().GetByEmail(l.ctx, l.getDefaultTenantID(), userModel.Email)
	if err != nil {
		l.Logger.Errorf("Login GetByEmail err: %v", err)
		return nil, errorx.NewInternalError("登录失败")
	}
	if user == nil {
		return nil, errorx.NewBusinessError(errorx.CodeUnauth, "用户不存在或密码错误")
	}

	// 4. 密码验证（Domain Model 行为方法）
	if !userModel.CheckPasswordMatch(user.Password) {
		return nil, errorx.NewBusinessError(errorx.CodeUnauth, "用户不存在或密码错误")
	}

	// 5. 生成 JWT Token
	token, err := auth.GenerateToken(l.svcCtx.Config.Auth, user.ID, user.TenantID, user.Email, user.Role)
	if err != nil {
		l.Logger.Errorf("Login GenerateToken err: %v", err)
		return nil, errorx.NewInternalError("登录失败")
	}

	// 6. 更新最后登录时间（非关键操作，失败仅记录日志）
	if updateErr := l.getUserRepo().UpdateLastLogin(l.ctx, user.ID); updateErr != nil {
		l.Logger.Errorf("Login UpdateLastLogin err: %v", updateErr)
	}

	// 7. assembler 转换 PO → 响应
	loginData := assembler.UserPoToLoginData(user, token)
	return &authtypes.LoginResponse{
		Code:      0,
		Msg:       "登录成功",
		Timestamp: time.Now().Unix(),
		Data:      loginData,
	}, nil
}

func (l *LoginLogic) getUserRepo() repo.UserRepository {
	return l.svcCtx.UserRepo
}

func (l *LoginLogic) getDefaultTenantID() int64 {
	// MVP: 使用默认租户 ID=1
	return 1
}
