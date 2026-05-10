// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package open

import (
	"net/http"

	"aipivot/internal/logic/open"
	"aipivot/internal/svc"
	"aipivot/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Webhook 入站消息 — 接收外部平台推送的用户消息并返回 AI 回复
func WebhookInboundHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WebhookInboundRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := open.NewWebhookInboundLogic(r.Context(), svcCtx)
		resp, err := l.WebhookInbound(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
