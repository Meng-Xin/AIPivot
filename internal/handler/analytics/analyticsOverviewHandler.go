package analytics

import (
	"net/http"

	"aipivot/internal/logic/analytics"
	"aipivot/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// AnalyticsOverviewHandler 获取对话分析概览
func AnalyticsOverviewHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := analytics.NewAnalyticsOverviewLogic(r.Context(), svcCtx)
		resp, err := l.AnalyticsOverview()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
