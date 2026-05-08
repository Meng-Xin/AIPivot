package handler

import (
	"net/http"

	"aipivot/internal/svc"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func MetricsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if svcCtx.Metrics == nil {
			http.NotFound(w, r)
			return
		}

		promhttp.HandlerFor(svcCtx.Metrics.Registry(), promhttp.HandlerOpts{}).ServeHTTP(w, r)
	}
}
