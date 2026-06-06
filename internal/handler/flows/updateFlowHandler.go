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

// 更新 Flow
func UpdateFlowHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateFlowRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := flows.NewUpdateFlowLogic(r.Context(), svcCtx)
		resp, err := l.UpdateFlow(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
