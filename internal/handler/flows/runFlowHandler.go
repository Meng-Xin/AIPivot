package flows

import (
	"net/http"

	"aipivot/internal/logic/flows"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// RunFlowHandler 试运行 Flow，SSE 流式返回执行过程。
// 手写 Handler（非 goctl 模板）：SSE 不走标准 JSON 响应模式，参考 sendMessageStreamHandler。
func RunFlowHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RunFlowRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := flows.NewRunFlowLogic(r.Context(), svcCtx)
		l.RunFlow(w, &req)
	}
}
