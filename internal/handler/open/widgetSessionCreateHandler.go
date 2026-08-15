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

// Widget — 创建访客会话（返回 sessionToken 作为后续凭证）
func WidgetSessionCreateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WidgetSessionCreateRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := open.NewWidgetSessionCreateLogic(r.Context(), svcCtx)
		resp, err := l.WidgetSessionCreate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
