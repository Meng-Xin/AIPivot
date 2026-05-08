package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestRequestIDFromHeaderIsStoredAndReturned(t *testing.T) {
	metrics := NewMetrics(prometheus.NewRegistry())
	middleware := Middleware(metrics, "aipivot-api")
	req := httptest.NewRequest(http.MethodGet, "/v1/ping", nil)
	req.Header.Set(RequestIDHeader, "request-123")
	rec := httptest.NewRecorder()

	handler := middleware(func(w http.ResponseWriter, r *http.Request) {
		if got := RequestIDFromContext(r.Context()); got != "request-123" {
			t.Fatalf("expected request id request-123, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	})

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(RequestIDHeader); got != "request-123" {
		t.Fatalf("expected response request id request-123, got %q", got)
	}
}

func TestRequestIDIsGeneratedWhenMissing(t *testing.T) {
	metrics := NewMetrics(prometheus.NewRegistry())
	middleware := Middleware(metrics, "aipivot-api")
	req := httptest.NewRequest(http.MethodGet, "/v1/ping", nil)
	rec := httptest.NewRecorder()

	handler := middleware(func(w http.ResponseWriter, r *http.Request) {
		if got := RequestIDFromContext(r.Context()); got == "" {
			t.Fatalf("expected generated request id")
		}
		w.WriteHeader(http.StatusOK)
	})

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(RequestIDHeader); got == "" {
		t.Fatalf("expected response request id")
	}
}

func TestMiddlewareRecordsHTTPMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewMetrics(registry)
	middleware := Middleware(metrics, "aipivot-api")
	req := httptest.NewRequest(http.MethodGet, "/v1/ping", nil)
	rec := httptest.NewRecorder()

	handler := middleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	handler.ServeHTTP(rec, req)

	metricsRec := httptest.NewRecorder()
	promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(metricsRec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := metricsRec.Body.String()

	if !strings.Contains(body, `aipivot_http_requests_total{method="GET",path="/v1/ping",status="201"} 1`) {
		t.Fatalf("expected request counter in metrics body, got:\n%s", body)
	}
}

func TestMiddlewareCreatesTraceID(t *testing.T) {
	provider := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
	})

	metrics := NewMetrics(prometheus.NewRegistry())
	middleware := Middleware(metrics, "aipivot-api")
	req := httptest.NewRequest(http.MethodGet, "/v1/ping", nil)
	rec := httptest.NewRecorder()

	handler := middleware(func(w http.ResponseWriter, r *http.Request) {
		if got := TraceIDFromContext(r.Context()); got == "" {
			t.Fatalf("expected trace id in request context")
		}
		w.WriteHeader(http.StatusOK)
	})

	handler.ServeHTTP(rec, req)
}
