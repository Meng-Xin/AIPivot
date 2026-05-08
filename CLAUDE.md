# AIPivot 项目说明

## 项目概述

AIPivot 是一个基于 Go 语言的 **AI 聊天机器人项目**，使用 go-zero 框架构建 REST API 服务。项目当前处于**第一阶段（基础设施阶段）**，已完成运行时基础设施和可观测性能力建设。

## 技术栈

| 技术 | 用途 |
|---|---|
| Go 1.22 | 编程语言 |
| go-zero v1.7.6 | REST Web 框架 |
| GORM | PostgreSQL ORM |
| go-redis | Redis 客户端 |
| OpenTelemetry + Jaeger | 链路追踪 |
| Prometheus | 指标收集 |
| PostgreSQL 16 | 数据库 |
| Redis 7 | 缓存 |
| Docker Compose | 本地开发环境 |

## 项目结构

```
├── aipivot.go                 # 服务入口，编排启动流程
├── aipivot.api                # go-zero API 定义文件
├── go.mod / go.sum            # Go 模块依赖
├── etc/
│   └── aipivot-api.yaml       # 服务运行时配置
├── internal/
│   ├── config/config.go       # 配置结构体（嵌入 rest.RestConf，扩展 PG/Redis/Telemetry/Metrics 配置）
│   ├── svc/servicecontext.go  # 服务上下文：依赖初始化与生命周期管理
│   ├── handler/               # HTTP 处理器层（health/ready/ping/metrics）
│   ├── logic/                  # 业务逻辑层（health/ready/ping）
│   ├── types/types.go         # 响应类型定义
│   ├── infra/                 # 基础设施层（PostgreSQL/Redis 初始化与健康检查）
│   └── observability/         # 可观测性层（tracing/metrics/middleware/context）
├── deploy/
│   ├── docker-compose.yml     # 本地依赖编排（PG/Redis/Jaeger/Prometheus）
│   └── prometheus/prometheus.yml
└── docs/
    ├── 需求.md                 # 需求文档
    └── memory.md               # 项目经验记录
```

## 架构要点

### 分层模式
遵循 go-zero 标准分层：**Handler（HTTP 解析） -> Logic（业务逻辑） -> ServiceContext（依赖注入）**

### 启动流程
1. 加载 YAML 配置 -> 2. 初始化 go-zero REST Server -> 3. 初始化 PostgreSQL -> 4. 初始化 Redis -> 5. 初始化 OpenTelemetry 追踪 -> 6. 初始化 Prometheus 指标 -> 7. 注册中间件（RequestID/追踪/日志/指标） -> 8. 注册路由 -> 9. 启动服务

### API 端点

| 端点 | 说明 |
|---|---|
| GET `/healthz` | 存活检查 |
| GET `/readyz` | 就绪检查（检查 PG/Redis 依赖） |
| GET `/metrics` | Prometheus 指标暴露 |
| GET `/v1/ping` | 请求通道验证（返回 pong + traceId + requestId） |

### 可观测性
- **链路追踪**：OpenTelemetry OTLP gRPC 导出到 Jaeger
- **指标**：HTTP 请求计数、请求耗时直方图、依赖就绪状态
- **日志**：结构化访问日志，含 RequestID 和 TraceID

## goctl 工具使用

goctl 是 go-zero 的命令行代码生成工具。本项目 `aipivot.api` 是唯一的 API 定义文件，所有代码生成操作围绕它进行。

### 安装 goctl

```bash
go install github.com/zeromicro/go-zero/tools/goctl@latest
```

### 常用命令

```bash
# 从 .api 文件生成全部 Go 代码（handler/logic/types/routes）
goctl api go -api aipivot.api -dir .

# 校验 API 定义文件语法
goctl api validate -api aipivot.api

# 格式化 API 定义文件
goctl api format -api aipivot.api
```

### 生成产物说明

| 生成目录/文件 | 说明 |
|---|---|
| `internal/handler/*handler.go` | HTTP 处理器（goctl 生成） |
| `internal/logic/*logic.go` | 业务逻辑骨架（goctl 生成） |
| `internal/types/types.go` | 请求/响应类型（goctl 生成） |
| `internal/handler/routes.go` | 路由注册（goctl 生成） |

### 手动维护的文件（不会被 goctl 覆盖）

以下文件由开发人员手动编写和维护，`goctl` 不会生成或覆盖：

- `internal/config/` — 配置结构体
- `internal/svc/` — 服务上下文与依赖初始化
- `internal/infra/` — 基础设施层（PostgreSQL/Redis/健康检查）
- `internal/observability/` — 可观测性层（追踪/指标/中间件）
- `aipivot.go` — 服务入口
- `etc/aipivot-api.yaml` — 运行时配置

### 新增 API 端点的工作流

1. 编辑 `aipivot.api`，添加新的 service 方法、Request/Response 类型
2. 运行 `goctl api go -api aipivot.api -dir .` 重新生成代码
3. 在生成的 logic 文件中实现业务逻辑
4. 运行 `go build ./...` 和 `go test ./...` 验证

## 本地开发

```bash
# 启动依赖服务
docker compose -f deploy/docker-compose.yml up -d

# 运行服务
go run aipivot.go -f etc/aipivot-api.yaml

# 运行测试
go test ./...

# 编译检查
go build ./...
```

- 服务地址：`http://127.0.0.1:8888`
- Jaeger UI：`http://127.0.0.1:16686`
- Prometheus UI：`http://127.0.0.1:9090`

## 关键约定

1. **小改动直接在主工作区修改**，避免创建 worktree
2. **优先本地验证**：`go test ./...`、`go build ./...` 先通过再推进
3. **区分代码问题与环境问题**，不要混淆处理
4. **遵循 go-zero 框架约定**：Handler/Logic/ServiceContext 标准分层
5. **可观测性贯穿所有请求**：每个 HTTP 请求自动包含追踪、指标和日志
