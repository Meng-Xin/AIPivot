package auth

import (
	"context"
	"net/http"
	"strings"

	"aipivot/internal/config"
	"aipivot/internal/shared/errorx"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type contextKey string

const claimsKey contextKey = "auth_claims"

const (
	authHeader = "Authorization"
	bearerPfx  = "Bearer "
)

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

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*Claims)
	return claims, ok
}

func TenantIDFromContext(ctx context.Context) int64 {
	if claims, ok := ClaimsFromContext(ctx); ok {
		return claims.TenantID
	}
	return 0
}

func UserIDFromContext(ctx context.Context) int64 {
	if claims, ok := ClaimsFromContext(ctx); ok {
		return claims.UserID
	}
	return 0
}
