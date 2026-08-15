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

// Widget — 提交消息满意度评分
func WidgetMessageFeedbackHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WidgetMessageFeedbackRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := open.NewWidgetMessageFeedbackLogic(r.Context(), svcCtx)
		resp, err := l.WidgetMessageFeedback(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
