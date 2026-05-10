// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package middleware

import (
	"net/http"

	"aipivot/internal/config"
	"aipivot/internal/modules/auth"
)

type AuthMiddleware struct {
	authConf config.AuthConf
}

func NewAuthMiddleware(authConf config.AuthConf) *AuthMiddleware {
	return &AuthMiddleware{authConf: authConf}
}

// Handle 委托给 auth 模块的 JWT 校验逻辑，解析 Bearer token 并注入 Claims 到 context
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return auth.JWTMiddleware(m.authConf)(next)
}
