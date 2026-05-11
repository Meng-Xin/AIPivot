package auth

import (
	"time"

	"aipivot/internal/shared/po"
	"aipivot/internal/types"

	"github.com/google/uuid"
)

// NewUserPo 从注册请求直接构造 User PO
func NewUserPo(nickName, email, encryptedPwd string, tenantID int64) *po.User {
	return &po.User{
		UUID:     uuid.New().String(),
		TenantID: tenantID,
		Email:    email,
		NickName: nickName,
		Password: encryptedPwd,
		Role:     "member",
		Status:   "active",
	}
}

// ToShowUser 将 User PO 转换为展示类型
func ToShowUser(u *po.User) types.ShowUser {
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

// ToLoginData 将 User PO + token 转换为登录响应数据
func ToLoginData(u *po.User, token string) types.LoginData {
	return types.LoginData{
		Token:    token,
		ExpireAt: time.Now().Add(24 * time.Hour).Unix(),
		User:     ToShowUser(u),
	}
}
