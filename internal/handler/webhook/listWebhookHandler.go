// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package webhook

import (
	"net/http"

	"aipivot/internal/logic/webhook"
	"aipivot/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取 Webhook 列表
func ListWebhookHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := webhook.NewListWebhookLogic(r.Context(), svcCtx)
		resp, err := l.ListWebhook()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
