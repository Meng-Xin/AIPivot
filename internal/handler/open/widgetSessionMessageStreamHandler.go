package open

import (
	"net/http"

	"aipivot/internal/logic/open"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// WidgetSessionMessageStreamHandler SSE 流式 Widget 发消息 — 手写 Handler，绕过标准 JSON 响应。
func WidgetSessionMessageStreamHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WidgetMessageSendRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := open.NewWidgetSessionMessageStreamLogic(r.Context(), svcCtx)
		l.WidgetSessionMessageStream(w, &req)
	}
}
