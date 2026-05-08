package observability

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

func Middleware(metrics *Metrics, serviceName string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := r.Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = uuid.NewString()
			}

			propagator := otel.GetTextMapPropagator()
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			ctx = ContextWithRequestID(ctx, requestID)

			tracer := otel.Tracer(serviceName)
			ctx, span := tracer.Start(ctx, "HTTP "+r.Method+" "+r.URL.Path)
			defer span.End()

			recorder := newStatusRecorder(w)
			recorder.Header().Set(RequestIDHeader, requestID)

			next(recorder, r.WithContext(ctx))

			duration := time.Since(start)
			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
				attribute.Int("http.status_code", recorder.status),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("net.peer.ip", r.RemoteAddr),
			)
			if recorder.status >= http.StatusBadRequest {
				span.SetStatus(codes.Error, http.StatusText(recorder.status))
			}

			if metrics != nil {
				metrics.RecordHTTPRequest(r.Method, r.URL.Path, recorder.status, duration)
			}

			logx.WithContext(ctx).Infow("http_request",
				logx.Field("service", serviceName),
				logx.Field("method", r.Method),
				logx.Field("path", r.URL.Path),
				logx.Field("status", recorder.status),
				logx.Field("duration_ms", duration.Milliseconds()),
				logx.Field("remote_addr", r.RemoteAddr),
				logx.Field("request_id", requestID),
			)
		}
	}
}
