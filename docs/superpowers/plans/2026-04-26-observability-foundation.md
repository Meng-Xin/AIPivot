# Observability Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first AIPivot infrastructure phase: a runnable go-zero REST API with PostgreSQL, Redis, OpenTelemetry tracing to Jaeger, Prometheus metrics, go-zero logging, and health/readiness endpoints.

**Architecture:** Keep the go-zero Handler -> Logic -> ServiceContext pattern. Put external dependency setup in `internal/infra` and request observability in `internal/observability`, then wire both through `internal/svc` and `aipivot.go`.

**Tech Stack:** Go, go-zero REST, GORM PostgreSQL driver, go-redis, OpenTelemetry, Jaeger exporter, Prometheus client, Docker Compose.

---

## File Structure

Create these files:

- `go.mod`: Go module and direct dependencies.
- `aipivot.api`: go-zero API definition for infrastructure endpoints.
- `aipivot.go`: service entrypoint.
- `etc/aipivot-api.yaml`: local service configuration.
- `deploy/docker-compose.yml`: local PostgreSQL, Redis, Jaeger, Prometheus.
- `deploy/prometheus/prometheus.yml`: Prometheus scrape config.
- `internal/config/config.go`: runtime config structs.
- `internal/svc/servicecontext.go`: dependency initialization and shutdown holder.
- `internal/infra/postgres.go`: GORM PostgreSQL initialization.
- `internal/infra/redis.go`: go-redis initialization.
- `internal/infra/health.go`: dependency checker abstraction and aggregation.
- `internal/observability/context.go`: request ID and trace ID context helpers.
- `internal/observability/metrics.go`: Prometheus collectors.
- `internal/observability/middleware.go`: HTTP middleware for request ID, trace, logs, metrics.
- `internal/observability/responsewriter.go`: status-capturing response writer.
- `internal/observability/tracing.go`: OpenTelemetry tracer setup and shutdown.
- `internal/types/types.go`: endpoint response types.
- `internal/logic/healthlogic.go`: health response logic.
- `internal/logic/readylogic.go`: readiness response logic.
- `internal/logic/pinglogic.go`: ping response logic.
- `internal/handler/routes.go`: route registration.
- `internal/handler/healthhandler.go`: `/healthz` handler.
- `internal/handler/readyhandler.go`: `/readyz` handler.
- `internal/handler/pinghandler.go`: `/v1/ping` handler.
- `internal/handler/metricshandler.go`: `/metrics` handler.

Test files:

- `internal/infra/health_test.go`
- `internal/observability/middleware_test.go`
- `internal/logic/pinglogic_test.go`
- `internal/handler/handler_test.go`

Do not modify:

- `docs/需求.md`
- `.idea/`

## Task 1: Project Skeleton And Static Configuration

**Files:**

- Create: `go.mod`
- Create: `aipivot.api`
- Create: `etc/aipivot-api.yaml`
- Create: `deploy/docker-compose.yml`
- Create: `deploy/prometheus/prometheus.yml`

- [ ] **Step 1: Create the Go module file**

Create `go.mod`:

```go
module aipivot

go 1.22

require (
	github.com/google/uuid v1.6.0
	github.com/prometheus/client_golang v1.20.5
	github.com/redis/go-redis/v9 v9.7.0
	github.com/zeromicro/go-zero v1.7.6
	go.opentelemetry.io/otel v1.31.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/sdk v1.31.0
	go.opentelemetry.io/otel/trace v1.31.0
	gorm.io/driver/postgres v1.5.9
	gorm.io/gorm v1.25.12
)
```

- [ ] **Step 2: Download dependencies**

Run:

```powershell
go mod tidy
```

Expected: `go.sum` is created and command exits successfully.

- [ ] **Step 3: Create the go-zero API definition**

Create `aipivot.api`:

```go
syntax = "v1"

info(
	title: "AIPivot API"
	desc: "Infrastructure endpoints for AIPivot"
	author: "AIPivot"
	version: "v1"
)

type (
	HealthResponse {
		Status string `json:"status"`
	}

	DependencyStatus {
		Name  string `json:"name"`
		Ready bool   `json:"ready"`
		Error string `json:"error,optional"`
	}

	ReadyResponse {
		Status       string             `json:"status"`
		Dependencies []DependencyStatus `json:"dependencies"`
	}

	PingResponse {
		Message   string `json:"message"`
		TraceId   string `json:"traceId"`
		RequestId string `json:"requestId"`
	}
)

service aipivot-api {
	@handler HealthHandler
	get /healthz returns (HealthResponse)

	@handler ReadyHandler
	get /readyz returns (ReadyResponse)

	@handler MetricsHandler
	get /metrics

	@handler PingHandler
	get /v1/ping returns (PingResponse)
}
```

- [ ] **Step 4: Create the local service config**

Create `etc/aipivot-api.yaml`:

```yaml
Name: aipivot-api
Host: 0.0.0.0
Port: 8888
Mode: dev
Timeout: 30000

Log:
  ServiceName: aipivot-api
  Mode: console
  Level: info
  Encoding: plain

Postgres:
  Host: 127.0.0.1
  Port: 5432
  User: aipivot
  Password: aipivot
  Database: aipivot
  SSLMode: disable
  TimeZone: Asia/Shanghai
  MaxOpenConns: 20
  MaxIdleConns: 10

Redis:
  Addr: 127.0.0.1:6379
  Password: ""
  DB: 0

Telemetry:
  ServiceName: aipivot-api
  Environment: dev
  JaegerEndpoint: http://127.0.0.1:14268/api/traces
  SampleRatio: 1.0

Metrics:
  Enabled: true
  Path: /metrics
```

- [ ] **Step 5: Create Docker Compose dependencies**

Create `deploy/docker-compose.yml`:

```yaml
services:
  postgres:
    image: postgres:16
    container_name: aipivot-postgres
    environment:
      POSTGRES_USER: aipivot
      POSTGRES_PASSWORD: aipivot
      POSTGRES_DB: aipivot
      TZ: Asia/Shanghai
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U aipivot -d aipivot"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7
    container_name: aipivot-redis
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  jaeger:
    image: jaegertracing/all-in-one:1.57
    container_name: aipivot-jaeger
    environment:
      COLLECTOR_ZIPKIN_HOST_PORT: :9411
    ports:
      - "16686:16686"
      - "14268:14268"

  prometheus:
    image: prom/prometheus:v2.54.1
    container_name: aipivot-prometheus
    command:
      - "--config.file=/etc/prometheus/prometheus.yml"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
    ports:
      - "9090:9090"
```

- [ ] **Step 6: Create Prometheus scrape config**

Create `deploy/prometheus/prometheus.yml`:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: aipivot-api
    static_configs:
      - targets: ["host.docker.internal:8888"]
```

- [ ] **Step 7: Verify static files are present**

Run:

```powershell
Get-ChildItem -Recurse go.mod,aipivot.api,etc,deploy | Select-Object FullName
```

Expected: the command lists all files created in this task.

- [ ] **Step 8: Commit**

```powershell
git add go.mod go.sum aipivot.api etc/aipivot-api.yaml deploy/docker-compose.yml deploy/prometheus/prometheus.yml
git commit -m "chore: add observability foundation skeleton"
```

## Task 2: Configuration Structs

**Files:**

- Create: `internal/config/config.go`
- Test: no separate tests; this is declarative config loaded by go-zero and exercised by service tests in Task 9.

- [ ] **Step 1: Create config structs**

Create `internal/config/config.go`:

```go
package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	Postgres  PostgresConf
	Redis     RedisConf
	Telemetry TelemetryConf
	Metrics   MetricsConf
}

type PostgresConf struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	SSLMode      string
	TimeZone     string
	MaxOpenConns int
	MaxIdleConns int
}

type RedisConf struct {
	Addr     string
	Password string `json:",optional"`
	DB       int
}

type TelemetryConf struct {
	ServiceName    string
	Environment    string
	JaegerEndpoint string
	SampleRatio     float64 `json:",default=1"`
}

type MetricsConf struct {
	Enabled bool   `json:",default=true"`
	Path    string `json:",default=/metrics"`
}
```

- [ ] **Step 2: Verify config package compiles**

Run:

```powershell
go test ./internal/config
```

Expected: output contains `?   	aipivot/internal/config	[no test files]`.

- [ ] **Step 3: Commit**

```powershell
git add internal/config/config.go
git commit -m "chore: add service configuration structs"
```

## Task 3: Dependency Health Aggregation

**Files:**

- Create: `internal/infra/health_test.go`
- Create: `internal/infra/health.go`

- [ ] **Step 1: Write failing tests for dependency aggregation**

Create `internal/infra/health_test.go`:

```go
package infra

import (
	"context"
	"errors"
	"testing"
)

func TestCheckDependenciesReturnsReadyWhenAllChecksPass(t *testing.T) {
	checks := []DependencyCheck{
		{Name: "postgres", Check: func(context.Context) error { return nil }},
		{Name: "redis", Check: func(context.Context) error { return nil }},
	}

	result := CheckDependencies(context.Background(), checks)

	if !result.Ready {
		t.Fatalf("expected result to be ready")
	}
	if result.Status != "ready" {
		t.Fatalf("expected status ready, got %q", result.Status)
	}
	if len(result.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(result.Dependencies))
	}
	for _, dep := range result.Dependencies {
		if !dep.Ready {
			t.Fatalf("expected dependency %s to be ready", dep.Name)
		}
		if dep.Error != "" {
			t.Fatalf("expected dependency %s error to be empty, got %q", dep.Name, dep.Error)
		}
	}
}

func TestCheckDependenciesReturnsNotReadyWhenAnyCheckFails(t *testing.T) {
	checks := []DependencyCheck{
		{Name: "postgres", Check: func(context.Context) error { return nil }},
		{Name: "redis", Check: func(context.Context) error { return errors.New("connection refused") }},
	}

	result := CheckDependencies(context.Background(), checks)

	if result.Ready {
		t.Fatalf("expected result to be not ready")
	}
	if result.Status != "not_ready" {
		t.Fatalf("expected status not_ready, got %q", result.Status)
	}
	if result.Dependencies[1].Ready {
		t.Fatalf("expected redis to be not ready")
	}
	if result.Dependencies[1].Error != "connection refused" {
		t.Fatalf("expected redis error, got %q", result.Dependencies[1].Error)
	}
}
```

- [ ] **Step 2: Run tests and verify red**

Run:

```powershell
go test ./internal/infra -run TestCheckDependencies
```

Expected: FAIL because `DependencyCheck` and `CheckDependencies` are undefined.

- [ ] **Step 3: Implement dependency aggregation**

Create `internal/infra/health.go`:

```go
package infra

import "context"

type DependencyCheck struct {
	Name  string
	Check func(ctx context.Context) error
}

type DependencyStatus struct {
	Name  string
	Ready bool
	Error string
}

type ReadinessResult struct {
	Status       string
	Ready        bool
	Dependencies []DependencyStatus
}

func CheckDependencies(ctx context.Context, checks []DependencyCheck) ReadinessResult {
	result := ReadinessResult{
		Status:       "ready",
		Ready:        true,
		Dependencies: make([]DependencyStatus, 0, len(checks)),
	}

	for _, check := range checks {
		status := DependencyStatus{Name: check.Name, Ready: true}
		if err := check.Check(ctx); err != nil {
			status.Ready = false
			status.Error = err.Error()
			result.Ready = false
			result.Status = "not_ready"
		}

		result.Dependencies = append(result.Dependencies, status)
	}

	return result
}
```

- [ ] **Step 4: Run tests and verify green**

Run:

```powershell
go test ./internal/infra -run TestCheckDependencies
```

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add internal/infra/health.go internal/infra/health_test.go
git commit -m "feat: add dependency readiness aggregation"
```

## Task 4: PostgreSQL And Redis Initialization

**Files:**

- Create: `internal/infra/postgres.go`
- Create: `internal/infra/redis.go`
- Modify: `internal/infra/health.go`

- [ ] **Step 1: Add PostgreSQL initializer**

Create `internal/infra/postgres.go`:

```go
package infra

import (
	"context"
	"fmt"
	"time"

	"aipivot/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres(conf config.PostgresConf) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(postgresDSN(conf)), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get postgres raw db: %w", err)
	}

	if conf.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(conf.MaxOpenConns)
	}
	if conf.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
	}

	return db, nil
}

func CheckPostgres(db *gorm.DB) DependencyCheck {
	return DependencyCheck{
		Name: "postgres",
		Check: func(ctx context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return fmt.Errorf("get postgres raw db: %w", err)
			}

			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			if err := sqlDB.PingContext(pingCtx); err != nil {
				return fmt.Errorf("ping postgres: %w", err)
			}

			return nil
		},
	}
}

func ClosePostgres(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get postgres raw db: %w", err)
	}

	return sqlDB.Close()
}

func postgresDSN(conf config.PostgresConf) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		conf.Host,
		conf.User,
		conf.Password,
		conf.Database,
		conf.Port,
		conf.SSLMode,
		conf.TimeZone,
	)
}
```

- [ ] **Step 2: Add Redis initializer**

Create `internal/infra/redis.go`:

```go
package infra

import (
	"context"
	"fmt"
	"time"

	"aipivot/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedis(conf config.RedisConf) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
	})
}

func CheckRedis(client *redis.Client) DependencyCheck {
	return DependencyCheck{
		Name: "redis",
		Check: func(ctx context.Context) error {
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			if err := client.Ping(pingCtx).Err(); err != nil {
				return fmt.Errorf("ping redis: %w", err)
			}

			return nil
		},
	}
}
```

- [ ] **Step 3: Run package tests**

Run:

```powershell
go test ./internal/infra
```

Expected: PASS.

- [ ] **Step 4: Commit**

```powershell
git add internal/infra/postgres.go internal/infra/redis.go internal/infra/health.go
git commit -m "feat: add postgres and redis infrastructure"
```

## Task 5: Observability Context And Metrics

**Files:**

- Create: `internal/observability/middleware_test.go`
- Create: `internal/observability/context.go`
- Create: `internal/observability/metrics.go`
- Create: `internal/observability/responsewriter.go`

- [ ] **Step 1: Write failing tests for request ID and metrics**

Create `internal/observability/middleware_test.go`:

```go
package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
```

- [ ] **Step 2: Run tests and verify red**

Run:

```powershell
go test ./internal/observability -run "TestRequestID|TestMiddlewareRecords"
```

Expected: FAIL because `NewMetrics`, `Middleware`, `RequestIDHeader`, and context helpers are undefined.

- [ ] **Step 3: Implement request context helpers**

Create `internal/observability/context.go`:

```go
package observability

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

const RequestIDHeader = "X-Request-Id"

type contextKey string

const requestIDKey contextKey = "request_id"

func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey).(string)
	return value
}

func TraceIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()
	if !spanContext.HasTraceID() {
		return ""
	}

	return spanContext.TraceID().String()
}
```

- [ ] **Step 4: Implement Prometheus metrics**

Create `internal/observability/metrics.go`:

```go
package observability

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	registry          *prometheus.Registry
	httpRequests     *prometheus.CounterVec
	httpDuration     *prometheus.HistogramVec
	dependencyReady  *prometheus.GaugeVec
}

func NewMetrics(registry *prometheus.Registry) *Metrics {
	if registry == nil {
		registry = prometheus.NewRegistry()
	}

	metrics := &Metrics{
		registry: registry,
		httpRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "aipivot_http_requests_total",
			Help: "Total number of HTTP requests.",
		}, []string{"method", "path", "status"}),
		httpDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "aipivot_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path", "status"}),
		dependencyReady: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "aipivot_dependency_ready",
			Help: "Dependency readiness state, 1 for ready and 0 for not ready.",
		}, []string{"dependency"}),
	}

	registry.MustRegister(metrics.httpRequests, metrics.httpDuration, metrics.dependencyReady)
	return metrics
}

func (m *Metrics) Registry() *prometheus.Registry {
	return m.registry
}

func (m *Metrics) RecordHTTPRequest(method, path string, status int, duration time.Duration) {
	statusValue := strconv.Itoa(status)
	m.httpRequests.WithLabelValues(method, path, statusValue).Inc()
	m.httpDuration.WithLabelValues(method, path, statusValue).Observe(duration.Seconds())
}

func (m *Metrics) SetDependencyReady(name string, ready bool) {
	value := 0.0
	if ready {
		value = 1.0
	}
	m.dependencyReady.WithLabelValues(name).Set(value)
}
```

- [ ] **Step 5: Implement status-capturing response writer**

Create `internal/observability/responsewriter.go`:

```go
package observability

import "net/http"

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
```

- [ ] **Step 6: Implement middleware without tracing**

Create `internal/observability/middleware.go`:

```go
package observability

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

func Middleware(metrics *Metrics, serviceName string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := r.Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = uuid.NewString()
			}

			ctx := ContextWithRequestID(r.Context(), requestID)
			recorder := newStatusRecorder(w)
			recorder.Header().Set(RequestIDHeader, requestID)

			next(recorder, r.WithContext(ctx))

			duration := time.Since(start)
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
				logx.Field("trace_id", TraceIDFromContext(ctx)),
			)
		}
	}
}
```

- [ ] **Step 7: Run tests and verify green**

Run:

```powershell
go test ./internal/observability -run "TestRequestID|TestMiddlewareRecords"
```

Expected: PASS.

- [ ] **Step 8: Commit**

```powershell
git add internal/observability/context.go internal/observability/metrics.go internal/observability/responsewriter.go internal/observability/middleware.go internal/observability/middleware_test.go
git commit -m "feat: add request observability middleware"
```

## Task 6: OpenTelemetry Tracing

**Files:**

- Modify: `internal/observability/middleware_test.go`
- Modify: `internal/observability/middleware.go`
- Create: `internal/observability/tracing.go`

- [ ] **Step 1: Add failing test for trace context**

Add these imports to `internal/observability/middleware_test.go`:

```go
	"context"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
```

Append this test to `internal/observability/middleware_test.go`:

```go
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
```

- [ ] **Step 2: Run test and verify red**

Run:

```powershell
go test ./internal/observability -run TestMiddlewareCreatesTraceID
```

Expected: FAIL because middleware has not started a span.

- [ ] **Step 3: Implement tracing setup**

Create `internal/observability/tracing.go`:

```go
package observability

import (
	"context"
	"fmt"

	"aipivot/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracing(ctx context.Context, conf config.TelemetryConf) (func(context.Context) error, error) {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(conf.JaegerEndpoint)))
	if err != nil {
		return nil, fmt.Errorf("create jaeger exporter: %w", err)
	}

	sampler := sdktrace.AlwaysSample()
	if conf.SampleRatio >= 0 && conf.SampleRatio < 1 {
		sampler = sdktrace.TraceIDRatioBased(conf.SampleRatio)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(resource.NewWithAttributes(
			"",
			attribute.String("service.name", conf.ServiceName),
			attribute.String("deployment.environment", conf.Environment),
		)),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return provider.Shutdown, nil
}
```

- [ ] **Step 4: Add tracing to middleware**

Replace `internal/observability/middleware.go` with:

```go
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
				logx.Field("trace_id", TraceIDFromContext(ctx)),
				logx.Field("span_id", span.SpanContext().SpanID().String()),
			)
		}
	}
}
```

- [ ] **Step 5: Run observability tests**

Run:

```powershell
go test ./internal/observability
```

Expected: PASS.

- [ ] **Step 6: Commit**

```powershell
git add internal/observability/middleware.go internal/observability/middleware_test.go internal/observability/tracing.go
git commit -m "feat: add opentelemetry tracing"
```

## Task 7: Types And Logic Layer

**Files:**

- Create: `internal/types/types.go`
- Create: `internal/logic/pinglogic_test.go`
- Create: `internal/logic/healthlogic.go`
- Create: `internal/logic/readylogic.go`
- Create: `internal/logic/pinglogic.go`
- Modify: `internal/infra/health.go`

- [ ] **Step 1: Create response types**

Create `internal/types/types.go`:

```go
package types

type HealthResponse struct {
	Status string `json:"status"`
}

type DependencyStatus struct {
	Name  string `json:"name"`
	Ready bool   `json:"ready"`
	Error string `json:"error,omitempty"`
}

type ReadyResponse struct {
	Status       string             `json:"status"`
	Dependencies []DependencyStatus `json:"dependencies"`
}

type PingResponse struct {
	Message   string `json:"message"`
	TraceID   string `json:"traceId"`
	RequestID string `json:"requestId"`
}
```

- [ ] **Step 2: Add logic tests for ping response**

Create `internal/logic/pinglogic_test.go`:

```go
package logic

import (
	"context"
	"testing"

	"aipivot/internal/observability"
	"aipivot/internal/svc"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestPingReturnsMessageAndRequestID(t *testing.T) {
	ctx := observability.ContextWithRequestID(context.Background(), "request-123")
	logic := NewPingLogic(ctx, &svc.ServiceContext{})

	resp, err := logic.Ping()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Message != "pong" {
		t.Fatalf("expected pong, got %q", resp.Message)
	}
	if resp.RequestID != "request-123" {
		t.Fatalf("expected request-123, got %q", resp.RequestID)
	}
}

func TestPingReturnsTraceID(t *testing.T) {
	provider := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
	})

	tracer := provider.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()
	ctx = observability.ContextWithRequestID(ctx, "request-123")
	logic := NewPingLogic(ctx, &svc.ServiceContext{})

	resp, err := logic.Ping()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.TraceID == "" {
		t.Fatalf("expected trace id")
	}
}
```

- [ ] **Step 3: Run tests and verify red**

Run:

```powershell
go test ./internal/logic -run TestPing
```

Expected: FAIL because `NewPingLogic` is undefined.

- [ ] **Step 4: Add minimal service context type**

Create `internal/svc/servicecontext.go`:

```go
package svc

import (
	"context"

	"aipivot/internal/config"
	"aipivot/internal/infra"
	"aipivot/internal/observability"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config      config.Config
	DB          *gorm.DB
	Redis       *redis.Client
	Metrics     *observability.Metrics
	HealthChecks []infra.DependencyCheck
	Shutdown     func(context.Context) error
}
```

- [ ] **Step 5: Implement logic layer**

Create `internal/logic/healthlogic.go`:

```go
package logic

import (
	"context"

	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HealthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthLogic {
	return &HealthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HealthLogic) Health() (*types.HealthResponse, error) {
	return &types.HealthResponse{Status: "ok"}, nil
}
```

Create `internal/logic/readylogic.go`:

```go
package logic

import (
	"context"

	"aipivot/internal/infra"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReadyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReadyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReadyLogic {
	return &ReadyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReadyLogic) Ready() (*types.ReadyResponse, bool, error) {
	result := infra.CheckDependencies(l.ctx, l.svcCtx.HealthChecks)
	deps := make([]types.DependencyStatus, 0, len(result.Dependencies))
	for _, dep := range result.Dependencies {
		if l.svcCtx.Metrics != nil {
			l.svcCtx.Metrics.SetDependencyReady(dep.Name, dep.Ready)
		}
		deps = append(deps, types.DependencyStatus{
			Name:  dep.Name,
			Ready: dep.Ready,
			Error: dep.Error,
		})
	}

	return &types.ReadyResponse{
		Status:       result.Status,
		Dependencies: deps,
	}, result.Ready, nil
}
```

Create `internal/logic/pinglogic.go`:

```go
package logic

import (
	"context"

	"aipivot/internal/observability"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PingLogic) Ping() (*types.PingResponse, error) {
	return &types.PingResponse{
		Message:   "pong",
		TraceID:   observability.TraceIDFromContext(l.ctx),
		RequestID: observability.RequestIDFromContext(l.ctx),
	}, nil
}
```

- [ ] **Step 6: Run logic tests**

Run:

```powershell
go test ./internal/logic -run TestPing
```

Expected: PASS.

- [ ] **Step 7: Commit**

```powershell
git add internal/types/types.go internal/logic/*.go internal/logic/pinglogic_test.go internal/svc/servicecontext.go
git commit -m "feat: add infrastructure endpoint logic"
```

## Task 8: HTTP Handlers And Routes

**Files:**

- Create: `internal/handler/handler_test.go`
- Create: `internal/handler/routes.go`
- Create: `internal/handler/healthhandler.go`
- Create: `internal/handler/readyhandler.go`
- Create: `internal/handler/pinghandler.go`
- Create: `internal/handler/metricshandler.go`

- [ ] **Step 1: Write failing handler tests**

Create `internal/handler/handler_test.go`:

```go
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
```

- [ ] **Step 2: Run tests and verify red**

Run:

```powershell
go test ./internal/handler -run "TestHealth|TestReady|TestPing"
```

Expected: FAIL because handlers are undefined.

- [ ] **Step 3: Implement handlers**

Create `internal/handler/healthhandler.go`:

```go
package handler

import (
	"net/http"

	"aipivot/internal/logic"
	"aipivot/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func HealthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewHealthLogic(r.Context(), svcCtx)
		resp, err := l.Health()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
```

Create `internal/handler/readyhandler.go`:

```go
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
```

Create `internal/handler/pinghandler.go`:

```go
package handler

import (
	"net/http"

	"aipivot/internal/logic"
	"aipivot/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func PingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewPingLogic(r.Context(), svcCtx)
		resp, err := l.Ping()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
```

Create `internal/handler/metricshandler.go`:

```go
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
```

- [ ] **Step 4: Implement route registration**

Create `internal/handler/routes.go`:

```go
package handler

import (
	"net/http"

	"aipivot/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, svcCtx *svc.ServiceContext) {
	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/healthz",
			Handler: HealthHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/readyz",
			Handler: ReadyHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/metrics",
			Handler: MetricsHandler(svcCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/v1/ping",
			Handler: PingHandler(svcCtx),
		},
	})
}
```

- [ ] **Step 5: Run handler tests**

Run:

```powershell
go test ./internal/handler -run "TestHealth|TestReady|TestPing"
```

Expected: PASS.

- [ ] **Step 6: Commit**

```powershell
git add internal/handler/*.go internal/handler/handler_test.go
git commit -m "feat: add infrastructure HTTP handlers"
```

## Task 9: Service Context And Entrypoint Wiring

**Files:**

- Modify: `internal/svc/servicecontext.go`
- Create: `aipivot.go`

- [ ] **Step 1: Replace service context with real dependency wiring**

Replace `internal/svc/servicecontext.go` with:

```go
package svc

import (
	"context"
	"fmt"

	"aipivot/internal/config"
	"aipivot/internal/infra"
	"aipivot/internal/observability"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config       config.Config
	DB           *gorm.DB
	Redis        *redis.Client
	Metrics      *observability.Metrics
	HealthChecks []infra.DependencyCheck
	Shutdown     func(context.Context) error
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	shutdown, err := observability.InitTracing(context.Background(), c.Telemetry)
	if err != nil {
		return nil, fmt.Errorf("init tracing: %w", err)
	}

	db, err := infra.NewPostgres(c.Postgres)
	if err != nil {
		return nil, fmt.Errorf("init postgres: %w", err)
	}

	redisClient := infra.NewRedis(c.Redis)
	metrics := observability.NewMetrics(prometheus.NewRegistry())

	return &ServiceContext{
		Config:  c,
		DB:      db,
		Redis:   redisClient,
		Metrics: metrics,
		HealthChecks: []infra.DependencyCheck{
			infra.CheckPostgres(db),
			infra.CheckRedis(redisClient),
		},
		Shutdown: func(ctx context.Context) error {
			var shutdownErr error
			if err := shutdown(ctx); err != nil {
				shutdownErr = fmt.Errorf("shutdown tracing: %w", err)
			}
			if err := redisClient.Close(); err != nil && shutdownErr == nil {
				shutdownErr = fmt.Errorf("close redis: %w", err)
			}
			if err := infra.ClosePostgres(db); err != nil && shutdownErr == nil {
				shutdownErr = fmt.Errorf("close postgres: %w", err)
			}
			return shutdownErr
		},
	}, nil
}
```

- [ ] **Step 2: Create entrypoint**

Create `aipivot.go`:

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"aipivot/internal/config"
	"aipivot/internal/handler"
	"aipivot/internal/observability"
	"aipivot/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/aipivot-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	svcCtx, err := svc.NewServiceContext(c)
	if err != nil {
		logx.Errorf("failed to initialize service context: %v", err)
		panic(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := svcCtx.Shutdown(ctx); err != nil {
			logx.Errorf("failed to shutdown service context: %v", err)
		}
	}()

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	server.Use(observability.Middleware(svcCtx.Metrics, c.Name))
	handler.RegisterHandlers(server, svcCtx)

	fmt.Printf("Starting %s at %s:%d...\n", c.Name, c.Host, c.Port)
	server.Start()
}
```

- [ ] **Step 3: Run full tests**

Run:

```powershell
go test ./...
```

Expected: PASS.

- [ ] **Step 4: Build service**

Run:

```powershell
go build ./...
```

Expected: command exits successfully.

- [ ] **Step 5: Commit**

```powershell
git add aipivot.go internal/svc/servicecontext.go
git commit -m "feat: wire service context and api entrypoint"
```

## Task 10: Full Local Integration Verification

**Files:**

- Modify: `README.md`

- [ ] **Step 1: Update README with local run commands**

Replace `README.md` with:

````markdown
# AIPivot

AIPivot is an AI chat robot project built with Go. The first phase focuses on runtime infrastructure and observability.

## Observability Foundation

The current service is `aipivot-api`, a go-zero REST API with:

- PostgreSQL readiness check
- Redis readiness check
- OpenTelemetry traces exported to Jaeger
- Prometheus metrics at `/metrics`
- go-zero request logging
- infrastructure endpoints at `/healthz`, `/readyz`, and `/v1/ping`

## Local Dependencies

Start PostgreSQL, Redis, Jaeger, and Prometheus:

```powershell
docker compose -f deploy/docker-compose.yml up -d
```

## Run The Service

```powershell
go run aipivot.go -f etc/aipivot-api.yaml
```

## Verify

```powershell
curl http://127.0.0.1:8888/healthz
curl http://127.0.0.1:8888/readyz
curl http://127.0.0.1:8888/v1/ping
curl http://127.0.0.1:8888/metrics
```

Jaeger UI:

```text
http://127.0.0.1:16686
```

Prometheus UI:

```text
http://127.0.0.1:9090
```
````

- [ ] **Step 2: Run unit tests**

Run:

```powershell
go test ./...
```

Expected: PASS.

- [ ] **Step 3: Start local dependencies**

Run:

```powershell
docker compose -f deploy/docker-compose.yml up -d
```

Expected: Docker starts `aipivot-postgres`, `aipivot-redis`, `aipivot-jaeger`, and `aipivot-prometheus`.

- [ ] **Step 4: Start service**

Run:

```powershell
go run aipivot.go -f etc/aipivot-api.yaml
```

Expected: output includes `Starting aipivot-api at 0.0.0.0:8888...`.

- [ ] **Step 5: Verify endpoints from a second terminal**

Run:

```powershell
curl http://127.0.0.1:8888/healthz
curl http://127.0.0.1:8888/readyz
curl http://127.0.0.1:8888/v1/ping
curl http://127.0.0.1:8888/metrics
```

Expected:

- `/healthz` response contains `"status":"ok"`.
- `/readyz` response contains `"status":"ready"`.
- `/v1/ping` response contains `"message":"pong"`, `"traceId"`, and `"requestId"`.
- `/metrics` response contains `aipivot_http_requests_total`.

- [ ] **Step 6: Verify Jaeger**

Open `http://127.0.0.1:16686` and search for service `aipivot-api`.

Expected: at least one trace appears after calling `/v1/ping`.

- [ ] **Step 7: Verify Prometheus**

Open `http://127.0.0.1:9090/targets`.

Expected: `aipivot-api` target is `UP`.

Query:

```text
aipivot_http_requests_total
```

Expected: Prometheus returns time series for the endpoints called in Step 5.

- [ ] **Step 8: Verify readiness failure**

Run:

```powershell
docker stop aipivot-redis
curl http://127.0.0.1:8888/readyz
docker start aipivot-redis
```

Expected: the curl response uses HTTP status `503` and body contains `"status":"not_ready"` while Redis is stopped.

- [ ] **Step 9: Commit**

```powershell
git add README.md
git commit -m "docs: add observability foundation runbook"
```

## Final Verification

Run:

```powershell
gofmt -w aipivot.go internal/config/config.go internal/svc/servicecontext.go internal/infra/*.go internal/observability/*.go internal/types/types.go internal/logic/*.go internal/handler/*.go
go test ./...
go build ./...
git status --short
```

Expected:

- `gofmt` completes without output.
- `go test ./...` passes.
- `go build ./...` passes.
- `git status --short` shows only pre-existing unrelated user files, if any.
