package admin

import (
	"net/http"

	"aipivot/internal/logic/admin"
	"aipivot/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetTenantHandler 获取当前租户信息
func GetTenantHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := admin.NewGetTenantLogic(r.Context(), svcCtx)
		resp, err := l.GetTenant()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
