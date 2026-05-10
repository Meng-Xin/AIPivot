package open

import (
	"net/http"

	"aipivot/internal/logic/open"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// ChatCompletionStreamHandler SSE 流式 Chat Completion — 手写 Handler，绕过标准 JSON 响应。
func ChatCompletionStreamHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatCompletionRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := open.NewChatCompletionStreamLogic(r.Context(), svcCtx)
		l.ChatCompletionStream(w, &req)
	}
}
