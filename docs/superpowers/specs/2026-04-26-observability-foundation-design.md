# Observability Foundation Design

## Goal

Build the first infrastructure phase for AIPivot: a single go-zero REST API service with PostgreSQL, Redis, OpenTelemetry tracing to Jaeger, Prometheus metrics, go-zero logging, and health/readiness endpoints.

This phase intentionally does not implement AI chat behavior. It creates the runtime foundation that future chat, session, model-provider, and streaming features will rely on.

## Confirmed Decisions

- Service shape: one go-zero REST API service.
- Go module: `aipivot`.
- Service name: `aipivot-api`.
- Database: PostgreSQL.
- Cache: Redis.
- Local environment: Docker Compose for PostgreSQL, Redis, Jaeger, and Prometheus.
- Tracing: OpenTelemetry SDK exporting traces to Jaeger.
- Metrics: Prometheus scrape endpoint at `/metrics`.
- Logging: go-zero `logx`, with request logs enriched by request and trace identifiers.
- Initial API scope: infrastructure endpoints only.

## Architecture Boundary

The first phase builds one runnable service: `aipivot-api`. It is a REST API service, not a microservice set. It does not include users, chat sessions, message storage, AI provider integrations, authentication, streaming, or business tables.

On startup, the service will:

1. Load YAML configuration.
2. Initialize go-zero REST server configuration.
3. Initialize go-zero logging.
4. Initialize PostgreSQL through GORM.
5. Initialize Redis through `go-redis`.
6. Initialize OpenTelemetry tracing with a Jaeger exporter.
7. Initialize Prometheus collectors.
8. Register HTTP middleware for request IDs, tracing, logging, and metrics.
9. Register infrastructure endpoints.

Request flow:

```text
client
  -> go-zero REST server
  -> observability middleware
  -> handler
  -> logic
  -> service context
  -> PostgreSQL / Redis checks where needed
```

Completion for this phase means a local developer can start dependencies with Docker Compose, run `aipivot-api`, call the infrastructure endpoints, see request logs, inspect traces in Jaeger, and query HTTP metrics in Prometheus.

## Directory Structure

The project should follow go-zero's standard REST layout, with small packages for observability and infrastructure glue.

```text
AIPivot/
  go.mod
  go.sum
  aipivot.api
  aipivot.go

  etc/
    aipivot-api.yaml

  internal/
    config/
      config.go

    svc/
      servicecontext.go

    handler/
      routes.go
      healthhandler.go
      readyhandler.go
      pinghandler.go
      metricshandler.go

    logic/
      healthlogic.go
      readylogic.go
      pinglogic.go

    types/
      types.go

    observability/
      logger.go
      tracing.go
      metrics.go
      middleware.go

    infra/
      postgres.go
      redis.go
      health.go

  deploy/
    docker-compose.yml
    prometheus/
      prometheus.yml

  docs/
    superpowers/
      specs/
        2026-04-26-observability-foundation-design.md
```

## Module Responsibilities

`aipivot.go` owns startup orchestration. It loads configuration, creates the service context, registers middleware and routes, starts the go-zero server, and shuts down tracing cleanly.

`internal/config` defines all runtime configuration. It should include service settings, go-zero logging, PostgreSQL, Redis, telemetry, and metrics fields.

`internal/svc` owns application dependencies. It should hold the parsed config, GORM database handle, Redis client, metrics collector, and shutdown hooks needed by the service.

`internal/infra` owns external dependency setup and readiness checks. It initializes PostgreSQL and Redis and exposes dependency check functions that do not depend on HTTP handlers.

`internal/observability` owns request observability. It initializes tracing, defines Prometheus collectors, provides HTTP middleware, and contains helpers for request ID and trace fields.

`internal/handler` and `internal/logic` follow go-zero layering. Handlers parse and write HTTP responses. Logic builds endpoint results. Logic accesses dependencies through `svc.ServiceContext`.

`deploy` contains local development orchestration only. Production deployment manifests are outside this phase.

## Configuration

The initial config file should live at `etc/aipivot-api.yaml`.

```yaml
Name: aipivot-api
Host: 0.0.0.0
Port: 8888
Mode: dev

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

The service runs on the host machine by default. Docker Compose runs PostgreSQL, Redis, Jaeger, and Prometheus. Prometheus should scrape the host service with `host.docker.internal:8888` on Windows.

## Local Runtime

The local development flow should be:

```powershell
docker compose -f deploy/docker-compose.yml up -d
go run aipivot.go -f etc/aipivot-api.yaml
```

Validation endpoints:

- `GET http://127.0.0.1:8888/healthz`
- `GET http://127.0.0.1:8888/readyz`
- `GET http://127.0.0.1:8888/v1/ping`
- `GET http://127.0.0.1:8888/metrics`

Local observability UIs:

- Jaeger UI: `http://127.0.0.1:16686`
- Prometheus UI: `http://127.0.0.1:9090`

Docker Compose should include:

- `postgres:16`
- `redis:7`
- `jaegertracing/all-in-one`
- `prom/prometheus`

Prometheus configuration:

```yaml
scrape_configs:
  - job_name: aipivot-api
    static_configs:
      - targets: ["host.docker.internal:8888"]
```

## API Endpoints

`GET /healthz`

Returns `200 OK` when the process is alive. It does not check PostgreSQL or Redis.

Example response:

```json
{
  "status": "ok"
}
```

`GET /readyz`

Checks whether the service can reach PostgreSQL and Redis. Returns `200 OK` when all dependencies are ready. Returns `503 Service Unavailable` when any required dependency fails.

Example healthy response:

```json
{
  "status": "ready",
  "dependencies": [
    {
      "name": "postgres",
      "ready": true
    },
    {
      "name": "redis",
      "ready": true
    }
  ]
}
```

Example unhealthy response:

```json
{
  "status": "not_ready",
  "dependencies": [
    {
      "name": "postgres",
      "ready": true
    },
    {
      "name": "redis",
      "ready": false,
      "error": "connection refused"
    }
  ]
}
```

`GET /metrics`

Exposes Prometheus metrics.

`GET /v1/ping`

Exercises the normal request path and returns request diagnostics.

Example response:

```json
{
  "message": "pong",
  "traceId": "4bf92f3577b34da6a3ce929d0e0e4736",
  "requestId": "d6f2d843-04dc-4fd5-8c49-b3f268cdb92b"
}
```

## Request Observability

All HTTP requests pass through `observability.Middleware`.

### Request ID

The middleware reads `X-Request-Id`. If the header is absent, it generates a UUID. The request ID is stored in request context and written back to the response as `X-Request-Id`.

### Tracing

The middleware extracts incoming W3C trace context from headers. If no context is present, it creates a new root span.

Span naming:

```text
HTTP <method> <path>
```

Example:

```text
HTTP GET /v1/ping
```

Span attributes should include:

- `http.method`
- `http.route` or path
- `http.status_code`
- `http.user_agent`
- `net.peer.ip`

The span status is normal for successful responses and records error state for 4xx/5xx responses.

### Logging

The middleware logs one access record after each request using go-zero `logx`. The record should include:

- HTTP method
- path
- status code
- duration in milliseconds
- remote address
- request ID
- trace ID
- span ID

For the first phase, all endpoints are logged. A configurable skip list for noisy paths is out of scope for this design.

### Metrics

The service exposes these Prometheus metrics:

- `aipivot_http_requests_total{method,path,status}`
- `aipivot_http_request_duration_seconds_bucket{method,path,status}`
- `aipivot_dependency_ready{dependency}`

The readiness endpoint updates `aipivot_dependency_ready` with `1` for ready and `0` for not ready.

## Testing Strategy

Implementation should follow test-first development for handwritten behavior. Generated go-zero code and static configuration may be verified after generation.

Unit tests:

- `observability` middleware generates or propagates `X-Request-Id`.
- `observability` middleware records HTTP request metrics.
- Ping logic returns `pong`, trace ID, and request ID.
- Dependency check aggregation reports ready when all checks pass.
- Dependency check aggregation reports not ready when any dependency fails.

HTTP handler tests:

- `/healthz` returns `200`.
- `/readyz` returns `200` when all dependency checks pass.
- `/readyz` returns `503` when any dependency check fails.
- `/v1/ping` returns `message`, `traceId`, and `requestId`.

Manual integration checks:

- Docker Compose starts PostgreSQL, Redis, Jaeger, and Prometheus.
- `go run aipivot.go -f etc/aipivot-api.yaml` starts the service.
- Calling `/v1/ping` creates a trace visible in Jaeger.
- Prometheus target for `aipivot-api` is up.
- Prometheus can query `aipivot_http_requests_total`.
- Stopping Redis or PostgreSQL makes `/readyz` return `503`.
- `go test ./...` passes.

## Out Of Scope

This phase does not include:

- OpenAI or other model provider calls.
- User accounts.
- Authentication or authorization.
- Chat sessions.
- Message tables.
- GORM Gen model generation.
- Streaming responses through SSE or WebSocket.
- Grafana dashboards.
- Kubernetes or Helm deployment.
- Production-grade alerting.
- A full business error code system.

## Acceptance Criteria

The phase is complete when these commands and checks work in a local Windows development environment:

```powershell
docker compose -f deploy/docker-compose.yml up -d
go run aipivot.go -f etc/aipivot-api.yaml
```

Endpoint checks:

```powershell
curl http://127.0.0.1:8888/healthz
curl http://127.0.0.1:8888/readyz
curl http://127.0.0.1:8888/v1/ping
curl http://127.0.0.1:8888/metrics
```

Expected outcomes:

- `/healthz` returns `200`.
- `/readyz` returns `200` when PostgreSQL and Redis are reachable.
- `/readyz` returns `503` when PostgreSQL or Redis is unavailable.
- `/v1/ping` returns `pong`, a trace ID, and a request ID.
- `/metrics` returns Prometheus text format.
- go-zero logs include method, path, status, duration, request ID, trace ID, and span ID.
- Jaeger UI shows traces for `aipivot-api`.
- Prometheus UI shows an up target for `aipivot-api`.
- `go test ./...` passes.
