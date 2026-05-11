package auth

import "context"

type contextKey string

const (
	claimsKey          contextKey = "auth_claims"
	apiKeyTenantIDKey  contextKey = "api_key_tenant_id"
	apiKeyIDKey        contextKey = "api_key_id"
)

// ClaimsFromContext 从 JWT 认证的 context 中提取 Claims
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*Claims)
	return claims, ok
}

// TenantIDFromContext 从 JWT 认证的 context 中提取 tenantID
func TenantIDFromContext(ctx context.Context) int64 {
	if claims, ok := ClaimsFromContext(ctx); ok {
		return claims.TenantID
	}
	return 0
}

// UserIDFromContext 从 JWT 认证的 context 中提取 userID
func UserIDFromContext(ctx context.Context) int64 {
	if claims, ok := ClaimsFromContext(ctx); ok {
		return claims.UserID
	}
	return 0
}

// TenantIDFromAPIKeyContext 从 API Key 认证的 context 中提取 tenantID
func TenantIDFromAPIKeyContext(ctx context.Context) int64 {
	if v, ok := ctx.Value(apiKeyTenantIDKey).(int64); ok {
		return v
	}
	return 0
}

// APIKeyIDFromContext 从 context 中提取 API Key ID
func APIKeyIDFromContext(ctx context.Context) int64 {
	if v, ok := ctx.Value(apiKeyIDKey).(int64); ok {
		return v
	}
	return 0
}
