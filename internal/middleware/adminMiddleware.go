package middleware

import (
	"encoding/json"
	"net/http"

	"aipivot/internal/config"
	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
)

// AdminMiddleware 在 AuthMiddleware 基础上追加 role=admin 校验。
// 非 admin 角色的合法 JWT 请求将被 403 拒绝，防止普通用户访问管理端点。
type AdminMiddleware struct {
	authConf config.AuthConf
}

func NewAdminMiddleware(authConf config.AuthConf) *AdminMiddleware {
	return &AdminMiddleware{authConf: authConf}
}

func (m *AdminMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	jwtHandler := auth.JWTMiddleware(m.authConf)
	return jwtHandler(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.ClaimsFromContext(r.Context())
		if !ok || claims.Role != "admin" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(errorx.NewForbidError("需要管理员权限"))
			return
		}
		next(w, r)
	})
}
