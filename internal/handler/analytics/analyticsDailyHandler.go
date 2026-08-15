package analytics

import (
	"net/http"

	"aipivot/internal/logic/analytics"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// AnalyticsDailyHandler 获取日粒度对话统计趋势
func AnalyticsDailyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AnalyticsDailyRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := analytics.NewAnalyticsDailyLogic(r.Context(), svcCtx)
		resp, err := l.AnalyticsDaily(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
