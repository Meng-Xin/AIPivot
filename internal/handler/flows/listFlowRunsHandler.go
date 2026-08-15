// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package flows

import (
	"net/http"

	"aipivot/internal/logic/flows"
	"aipivot/internal/svc"
	"aipivot/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取 Flow 执行历史
func ListFlowRunsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListFlowRunsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := flows.NewListFlowRunsLogic(r.Context(), svcCtx)
		resp, err := l.ListFlowRuns(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
