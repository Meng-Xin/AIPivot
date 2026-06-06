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

// 创建 Flow
func CreateFlowHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateFlowRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := flows.NewCreateFlowLogic(r.Context(), svcCtx)
		resp, err := l.CreateFlow(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
