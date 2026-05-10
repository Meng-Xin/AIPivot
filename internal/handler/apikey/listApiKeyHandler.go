// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package apikey

import (
	"net/http"

	"aipivot/internal/logic/apikey"
	"aipivot/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取 API Key 列表
func ListApiKeyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := apikey.NewListApiKeyLogic(r.Context(), svcCtx)
		resp, err := l.ListApiKey()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
