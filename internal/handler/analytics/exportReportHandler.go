package analytics

import (
	"fmt"
	"net/http"
	"strconv"

	logicAnalytics "aipivot/internal/logic/analytics"
	"aipivot/internal/svc"
)

// ExportReportHandler 导出 SLA 报表（CSV 文件下载）
//
// GET /api/v1/analytics/export?days=30
// 响应：text/csv 附件，Content-Disposition: attachment; filename="sla-report-*.csv"
func ExportReportHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从 query 解析 days 参数，不合法时退化为默认值 30
		days, _ := strconv.Atoi(r.URL.Query().Get("days"))

		req := &logicAnalytics.ExportReportRequest{Days: days}
		l := logicAnalytics.NewExportLogic(r.Context(), svcCtx)
		data, filename, err := l.ExportCSV(req)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"code":500,"msg":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		// 允许前端 JS 通过 fetch 访问 Content-Disposition 头（跨域场景）
		w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}
