package assembler

import (
	"time"

	"aipivot/internal/modules/auth/domain/model"
	"aipivot/internal/shared/po"
	"aipivot/internal/types"

	"github.com/google/uuid"
)

// ① Request → Domain Model
func LoginRequestToModelUser(req *types.LoginRequest) *model.UserAuth {
	return &model.UserAuth{
		Email:    req.Email,
		Password: req.Password,
	}
}

// ① Request → Domain Model
func RegisterRequestToModelUser(req *types.RegisterRequest) *model.UserAuth {
	return &model.UserAuth{
		NickName: req.NickName,
		Email:    req.Email,
		Password: req.Password,
	}
}

// ② Domain Model → PO
func ModelUserToUserPo(m *model.UserAuth, tenantID int64, encryptedPwd string) *po.User {
	return &po.User{
		UUID:     uuid.New().String(),
		TenantID: tenantID,
		Email:    m.Email,
		NickName: m.NickName,
		Password: encryptedPwd,
		Role:     "member",
		Status:   "active",
	}
}

// ③ PO → ShowType
func UserPoToShowUser(u *po.User) types.ShowUser {
	var lastLogin int64
	if u.LastLogin != nil {
		lastLogin = u.LastLogin.Unix()
	}
	return types.ShowUser{
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
func UserPoToLoginData(u *po.User, token string) types.LoginData {
	return types.LoginData{
		Token:    token,
		ExpireAt: time.Now().Add(24 * time.Hour).Unix(),
		User:     UserPoToShowUser(u),
	}
}
