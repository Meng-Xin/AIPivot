package chat

import (
	"net/http"

	"aipivot/internal/logic/chat"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// SendMessageStreamHandler SSE 流式发送消息。
// 手写 Handler（非 goctl 生成），因为 SSE 不走标准 JSON 响应模式。
func SendMessageStreamHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SendMessageRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := chat.NewSendMessageStreamLogic(r.Context(), svcCtx)
		l.SendMessageStream(w, &req)
	}
}
