// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package infra

import (
	"net/http"

	"aipivot/internal/logic/infra"
	"aipivot/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ReadyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := infra.NewReadyLogic(r.Context(), svcCtx)
		resp, ready, err := l.Ready()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		statusCode := http.StatusOK
		if !ready {
			statusCode = http.StatusServiceUnavailable
		}

		httpx.WriteJsonCtx(r.Context(), w, statusCode, resp)
	}
}
