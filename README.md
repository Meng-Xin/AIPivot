# AIPivot

AIPivot is an AI chat robot project built with Go. The first phase focuses on runtime infrastructure and observability.

## Observability Foundation

The current service is `aipivot-api`, a go-zero REST API with:

- PostgreSQL readiness check
- Redis readiness check
- OpenTelemetry traces exported to Jaeger through OTLP HTTP
- Prometheus metrics at `/metrics`
- go-zero request logging
- infrastructure endpoints at `/healthz`, `/readyz`, and `/v1/ping`

## Local Dependencies

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

Jaeger UI: `http://127.0.0.1:16686`

Prometheus UI: `http://127.0.0.1:9090`

---

# 项目功能总结

AIPivot 项目实现了一个基于 go-zero 的 AI 聊天机器人服务，具备如下基础设施能力，适合作为后续项目的脚手架：

## 已具备的基础设施能力

- **配置管理**：YAML 配置文件统一管理服务、数据库、Redis、链路追踪、指标等参数
- **依赖初始化**：ServiceContext 统一初始化和管理依赖（PostgreSQL、Redis、OpenTelemetry、Prometheus）
- **健康检查**：/healthz（存活）、/readyz（依赖就绪）接口，自动检测依赖状态
- **链路追踪**：OpenTelemetry 集成 Jaeger，所有 HTTP 请求自动注入 TraceID
- **指标采集**：Prometheus 集成，自动采集 HTTP 请求数、耗时、依赖状态等
- **结构化日志**：访问日志自动带 RequestID、TraceID，便于排查
- **分层架构**：严格 Handler/Logic/ServiceContext 分层，易于扩展
- **本地开发环境**：docker-compose 一键启动 PG/Redis/Jaeger/Prometheus
- **goctl 工作流**：API 设计、代码生成、类型定义全流程规范

## 适用场景

适合所有需要 REST API、数据库、缓存、可观测性、健康检查等基础设施的 Go 服务，尤其是 go-zero 体系下的微服务。

---

# 快速搭建基础设施指导文档

## 1. 目录结构建议

```
├── aipivot.go                 # 服务入口
├── aipivot.api                # go-zero API 定义
├── go.mod / go.sum            # Go 依赖
├── etc/
│   └── aipivot-api.yaml       # 配置文件
├── internal/
│   ├── config/                # 配置结构体
│   ├── svc/                   # 服务上下文
│   ├── handler/               # HTTP 处理器
│   ├── logic/                 # 业务逻辑
│   ├── types/                 # 类型定义
│   ├── infra/                 # 基础设施（PG/Redis/健康检查）
│   └── observability/         # 可观测性（tracing/metrics/middleware）
├── deploy/
│   └── docker-compose.yml     # 本地依赖编排
└── docs/                      # 文档
```

## 2. 基础设施搭建步骤

1. **初始化 go-zero 项目**
	- 建议直接复制本项目结构和 etc/aipivot-api.yaml 配置模板
2. **配置 PostgreSQL/Redis**
	- infra/postgres.go、infra/redis.go 参考本项目实现
3. **接入 OpenTelemetry/Prometheus**
	- observability/tracing.go、metrics.go 参考本项目
	- docker-compose.yml 启动 Jaeger/Prometheus
4. **健康检查接口**
	- handler/healthhandler.go、readyhandler.go、infra/health.go 参考实现
5. **goctl 规范 API 设计**
	- 统一在 aipivot.api 维护 API，goctl 生成 handler/logic/types
6. **ServiceContext 依赖注入**
	- internal/svc/servicecontext.go 统一初始化依赖
7. **本地开发环境**
	- `docker compose -f deploy/docker-compose.yml up -d` 启动依赖
	- `go run aipivot.go -f etc/aipivot-api.yaml` 启动服务

## 3. 复用建议

- 直接复制 internal/config、svc、infra、observability 目录到新项目
- 复制 aipivot.go、etc/aipivot-api.yaml、deploy/docker-compose.yml
- 修改 go.mod、aipivot.api、业务相关 handler/logic/types
- 通过 goctl 生成新 API 代码
- 保持健康检查、可观测性、依赖初始化等能力不变

---

如需扩展业务，只需在 aipivot.api 增加接口，goctl 生成骨架后实现 logic 层即可。

---

**本模板适合所有 go-zero 微服务项目的基础设施搭建，建议作为新项目脚手架直接复用。**
