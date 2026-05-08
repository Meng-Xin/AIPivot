package assembler

import (
	"time"

	"aipivot/internal/modules/auth/domain/model"
	authtypes "aipivot/internal/modules/auth/types"
	"aipivot/internal/shared/po"
)

// ① Request → Domain Model
func LoginRequestToModelUser(req *authtypes.LoginRequest) *model.UserAuth {
	return &model.UserAuth{
		Email:    req.Email,
		Password: req.Password,
	}
}

// ① Request → Domain Model
func RegisterRequestToModelUser(req *authtypes.RegisterRequest) *model.UserAuth {
	return &model.UserAuth{
		NickName: req.NickName,
		Email:    req.Email,
		Password: req.Password,
	}
}

// ② Domain Model → PO
func ModelUserToUserPo(m *model.UserAuth, tenantID int64, encryptedPwd string) *po.User {
	return &po.User{
		TenantID: tenantID,
		Email:    m.Email,
		NickName: m.NickName,
		Password: encryptedPwd,
		Role:     "member",
		Status:   "active",
	}
}

// ③ PO → ShowType
func UserPoToShowUser(u *po.User) authtypes.ShowUser {
	var lastLogin int64
	if u.LastLogin != nil {
		lastLogin = u.LastLogin.Unix()
	}
	return authtypes.ShowUser{
		ID:        u.ID,
		UUID:      u.UUID,
		TenantID:  u.TenantID,
		Email:     u.Email,
		NickName:  u.NickName,
		Role:      u.Role,
		Status:    u.Status,
		LastLogin: lastLogin,
		CreatedAt: u.CreatedAt.Unix(),
	}
}

// ③ PO → LoginResponse token data
func UserPoToLoginData(u *po.User, token string) authtypes.LoginData {
	return authtypes.LoginData{
		Token:    token,
		ExpireAt: time.Now().Add(24 * time.Hour).Unix(),
		User:     UserPoToShowUser(u),
	}
}
