package chat

import (
	"net/http"

	"aipivot/internal/logic/chat"
	"aipivot/internal/svc"
	"aipivot/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 转接人工客服
func EscalateConversationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.EscalateConversationRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := chat.NewEscalateConversationLogic(r.Context(), svcCtx)
		resp, err := l.EscalateConversation(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
