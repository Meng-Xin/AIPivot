package middleware

import (
	"net/http"

	"aipivot/internal/modules/auth"
)

type ApiKeyMiddleware struct {
	apiKeyRepo auth.ApiKeyRepository
}

func NewApiKeyMiddleware(apiKeyRepo auth.ApiKeyRepository) *ApiKeyMiddleware {
	return &ApiKeyMiddleware{apiKeyRepo: apiKeyRepo}
}

// Handle 委托给 auth 模块的 API Key 校验逻辑
func (m *ApiKeyMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return auth.APIKeyMiddleware(m.apiKeyRepo)(next)
}
