package middleware

import (
	"net/http"

	"aipivot/internal/shared/errorx"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// RateLimitMiddleware 基于 Redis 令牌桶的限流中间件，用于 Open API 端点。
// rate: 每秒允许请求数, burst: 令牌桶容量。
func RateLimitMiddleware(store *limit.TokenLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !store.Allow() {
				httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests,
					errorx.NewBusinessError(429, "rate limit exceeded, please try later"))
				return
			}
			next(w, r)
		}
	}
}
