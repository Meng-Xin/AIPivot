// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package flows

import (
	"net/http"

	"aipivot/internal/logic/flows"
	"aipivot/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取 Flow 列表
func ListFlowsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := flows.NewListFlowsLogic(r.Context(), svcCtx)
		resp, err := l.ListFlows()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
