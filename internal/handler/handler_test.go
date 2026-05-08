package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aipivot/internal/infra"
	"aipivot/internal/observability"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/prometheus/client_golang/prometheus"
)

func TestHealthHandlerReturnsOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	HealthHandler(&svc.ServiceContext{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp types.HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != "ok" {
		t.Fatalf("expected ok, got %q", resp.Status)
	}
}

func TestReadyHandlerReturnsOKWhenDependenciesPass(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	svcCtx := &svc.ServiceContext{
		Metrics: observability.NewMetrics(prometheus.NewRegistry()),
		HealthChecks: []infra.DependencyCheck{
			{Name: "postgres", Check: func(context.Context) error { return nil }},
			{Name: "redis", Check: func(context.Context) error { return nil }},
		},
	}

	ReadyHandler(svcCtx).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReadyHandlerReturnsUnavailableWhenDependencyFails(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	svcCtx := &svc.ServiceContext{
		Metrics: observability.NewMetrics(prometheus.NewRegistry()),
		HealthChecks: []infra.DependencyCheck{
			{Name: "postgres", Check: func(context.Context) error { return nil }},
			{Name: "redis", Check: func(context.Context) error { return errors.New("connection refused") }},
		},
	}

	ReadyHandler(svcCtx).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"not_ready"`) {
		t.Fatalf("expected not_ready response, got %s", rec.Body.String())
	}
}

func TestPingHandlerReturnsPong(t *testing.T) {
	ctx := observability.ContextWithRequestID(context.Background(), "request-123")
	req := httptest.NewRequest(http.MethodGet, "/v1/ping", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	PingHandler(&svc.ServiceContext{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp types.PingResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Message != "pong" {
		t.Fatalf("expected pong, got %q", resp.Message)
	}
	if resp.RequestID != "request-123" {
		t.Fatalf("expected request-123, got %q", resp.RequestID)
	}
}
