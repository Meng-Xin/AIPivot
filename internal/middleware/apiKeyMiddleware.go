package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	authRepo "aipivot/internal/modules/auth/repo"
	"aipivot/internal/shared/errorx"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// apiKeyContextKey 用于在 context 中传递 API Key 认证信息，与 JWT 的 claimsKey 隔离。
type apiKeyContextKey string

const (
	apiKeyHeader                       = "X-API-Key"
	apiKeyTenantIDKey apiKeyContextKey = "api_key_tenant_id"
	apiKeyIDKey       apiKeyContextKey = "api_key_id"
)

type ApiKeyMiddleware struct {
	apiKeyRepo authRepo.ApiKeyRepository
}

func NewApiKeyMiddleware(apiKeyRepo authRepo.ApiKeyRepository) *ApiKeyMiddleware {
	return &ApiKeyMiddleware{apiKeyRepo: apiKeyRepo}
}

// Handle 校验 X-API-Key 请求头：SHA-256 哈希 → 查库 → 检查过期 → 注入 tenantID 到 context。
func (m *ApiKeyMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawKey := r.Header.Get(apiKeyHeader)
		if rawKey == "" {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized,
				errorx.NewUnauthError("missing X-API-Key header"))
			return
		}

		keyHash := hashAPIKey(rawKey)
		apiKey, err := m.apiKeyRepo.GetByKeyHash(r.Context(), keyHash)
		if err != nil {
			logx.WithContext(r.Context()).Errorf("ApiKeyMiddleware lookup err: %v", err)
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
			_ = m.apiKeyRepo.UpdateLastUsed(context.Background(), apiKey.ID)
		}()

		ctx := context.WithValue(r.Context(), apiKeyTenantIDKey, apiKey.TenantID)
		ctx = context.WithValue(ctx, apiKeyIDKey, apiKey.ID)
		next(w, r.WithContext(ctx))
	}
}

func hashAPIKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}

// TenantIDFromAPIKeyContext 从 API Key 认证的 context 中提取 tenantID。
func TenantIDFromAPIKeyContext(ctx context.Context) int64 {
	if v, ok := ctx.Value(apiKeyTenantIDKey).(int64); ok {
		return v
	}
	return 0
}

// APIKeyIDFromContext 从 context 中提取 API Key ID。
func APIKeyIDFromContext(ctx context.Context) int64 {
	if v, ok := ctx.Value(apiKeyIDKey).(int64); ok {
		return v
	}
	return 0
}
