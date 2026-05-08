package handler

import (
	"net/http"

	authhandler "aipivot/internal/modules/auth/handler"
	"aipivot/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, svcCtx *svc.ServiceContext) {
	// 基础设施路由（无鉴权）
	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/healthz",
			Handler: HealthHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/readyz",
			Handler: ReadyHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/metrics",
			Handler: MetricsHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/v1/ping",
			Handler: PingHandler(svcCtx),
		},
	})

	// Auth 路由组（公开，无 JWT）
	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/v1/auth/login",
			Handler: authhandler.LoginHandler(svcCtx),
		},
	})
}
