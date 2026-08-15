package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"time"

	"aipivot/internal/shared/errorx"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

const apiKeyHeader = "X-API-Key"

// 公共密钥相关常量
const (
	KeyTypeMaster = "master"
	KeyTypePublic = "public"
)

// APIKeyMiddleware 返回 API Key 认证中间件：SHA-256 哈希 → 查库 → 检查过期 → 注入 tenantID 到 context。
// 对于 public key（pk_ 前缀），额外校验 Origin/Referer 是否命中 allowed_origins 白名单。
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

			// public key 必须通过域名白名单校验（fail-closed：空白名单直接拒绝）
			if apiKey.KeyType == KeyTypePublic {
				origin := resolveRequestOrigin(r)
				if !originAllowed(apiKey.AllowedOrigins, origin) {
					logx.WithContext(r.Context()).Infof("APIKeyMiddleware origin rejected: key=%d origin=%q", apiKey.ID, origin)
					httpx.WriteJsonCtx(r.Context(), w, http.StatusForbidden,
						errorx.NewForbidError("origin not allowed"))
					return
				}
			}

			// 异步更新最近使用时间（不阻塞请求）
			go func() {
				_ = apiKeyRepo.UpdateLastUsed(context.Background(), apiKey.ID)
			}()

			ctx := context.WithValue(r.Context(), apiKeyTenantIDKey, apiKey.TenantID)
			ctx = context.WithValue(ctx, apiKeyIDKey, apiKey.ID)
			if apiKey.KeyType == KeyTypePublic {
				ctx = context.WithValue(ctx, apiKeyIsPublicKeyKey, true)
				if apiKey.KnowledgeBaseID != nil && *apiKey.KnowledgeBaseID > 0 {
					ctx = context.WithValue(ctx, apiKeyKBIDKey, *apiKey.KnowledgeBaseID)
				}
			}
			next(w, r.WithContext(ctx))
		}
	}
}

// originAllowed 严格匹配 origin 是否在白名单中（不支持通配符，空列表 fail-closed）。
func originAllowed(allowed []string, origin string) bool {
	if len(allowed) == 0 || origin == "" {
		return false
	}
	// 大小写不敏感比较 scheme+host（端口区分），如 https://Example.com:443
	normalized := normalizeOrigin(origin)
	for _, a := range allowed {
		if normalizeOrigin(strings.TrimSpace(a)) == normalized {
			return true
		}
	}
	return false
}

// normalizeOrigin 统一 origin 格式：去尾斜杠、转小写。
func normalizeOrigin(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// 解析为 URL 取 scheme://host，避免 path/query 干扰
	if u, err := url.Parse(s); err == nil && u.Scheme != "" && u.Host != "" {
		s = u.Scheme + "://" + u.Host
	}
	s = strings.TrimRight(s, "/")
	return strings.ToLower(s)
}

// resolveRequestOrigin 优先取 Origin 头，缺失时降级从 Referer 提取 scheme://host。
func resolveRequestOrigin(r *http.Request) string {
	if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
		return origin
	}
	if referer := strings.TrimSpace(r.Header.Get("Referer")); referer != "" {
		if u, err := url.Parse(referer); err == nil && u.Scheme != "" && u.Host != "" {
			return u.Scheme + "://" + u.Host
		}
	}
	return ""
}

func hashAPIKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}
