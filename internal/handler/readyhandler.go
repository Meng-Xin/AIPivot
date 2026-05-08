package handler

import (
	"net/http"

	"aipivot/internal/logic"
	"aipivot/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ReadyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewReadyLogic(r.Context(), svcCtx)
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
