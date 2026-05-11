package auth

import (
	"context"
	"net/http"
	"strings"

	"aipivot/internal/config"
	"aipivot/internal/shared/errorx"

	"github.com/zeromicro/go-zero/rest/httpx"
)

const (
	authHeader = "Authorization"
	bearerPfx  = "Bearer "
)

// JWTMiddleware 返回 JWT 认证中间件，解析 Bearer token 并注入 Claims 到 context。
func JWTMiddleware(conf config.AuthConf) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authStr := r.Header.Get(authHeader)
			if authStr == "" || !strings.HasPrefix(authStr, bearerPfx) {
				httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized,
					errorx.NewUnauthError("missing or invalid authorization header"))
				return
			}

			tokenStr := strings.TrimPrefix(authStr, bearerPfx)
			claims, err := ParseToken(conf, tokenStr)
			if err != nil {
				httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized,
					errorx.NewUnauthError("invalid or expired token"))
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next(w, r.WithContext(ctx))
		}
	}
}
