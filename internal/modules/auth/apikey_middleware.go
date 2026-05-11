package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"aipivot/internal/shared/errorx"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

const apiKeyHeader = "X-API-Key"

// APIKeyMiddleware 返回 API Key 认证中间件：SHA-256 哈希 → 查库 → 检查过期 → 注入 tenantID 到 context。
func APIKeyMiddleware(apiKeyRepo ApiKeyRepository) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			rawKey := r.Header.Get(apiKeyHeader)
			if rawKey == "" {
				httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized,
					errorx.NewUnauthError("missing X-API-Key header"))
				return
			}

			keyHash := hashAPIKey(rawKey)
			apiKey, err := apiKeyRepo.GetByKeyHash(r.Context(), keyHash)
			if err != nil {
				logx.WithContext(r.Context()).Errorf("APIKeyMiddleware lookup err: %v", err)
				httpx.WriteJsonCtx(r.Context(), w, http.StatusInternalServerError,
					errorx.NewInternalError("authentication failed"))
				return
			}
			if apiKey == nil {
				httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized,
					errorx.NewUnauthError("invalid API key"))
				return
			}

			// 密钥过期检查
			if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
				httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized,
					errorx.NewUnauthError("API key has expired"))
				return
			}

			// 异步更新最近使用时间（不阻塞请求）
			go func() {
				_ = apiKeyRepo.UpdateLastUsed(context.Background(), apiKey.ID)
			}()

			ctx := context.WithValue(r.Context(), apiKeyTenantIDKey, apiKey.TenantID)
			ctx = context.WithValue(ctx, apiKeyIDKey, apiKey.ID)
			next(w, r.WithContext(ctx))
		}
	}
}

func hashAPIKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}
