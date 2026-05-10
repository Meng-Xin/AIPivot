# AIPivot 项目说明

## 项目概述

AIPivot 是一个基于 Go 语言的 **多租户 AI 能力中台**，使用 go-zero 框架构建 REST API 服务。项目采用多租户隔离架构，所有业务数据通过 `tenant_id` 关联。当前已完成：**运行时基础设施 + 可观测性 + 数据库迁移 + 认证登录模块（MVP）**。

## 技术栈

| 技术 | 用途 |
|---|---|
| Go 1.25 | 编程语言 |
| go-zero v1.7.6 | REST Web 框架 + goctl 代码生成 |
| GORM + GORM Gen | PostgreSQL ORM + 类型安全查询代码生成 |
| golang-migrate | 数据库迁移（SQL 文件驱动） |
| golang-jwt/v4 | JWT 认证 |
| go-redis/v9 | Redis 客户端 |
| OpenTelemetry + Jaeger | 链路追踪（OTLP gRPC） |
| Prometheus + Grafana | 指标采集 + 可视化 |
| PostgreSQL 16 | 数据库 |
| Redis 7 | 缓存 |
| Docker Compose | 本地开发环境（PG/Redis/Jaeger/Prometheus/Grafana） |

## 项目结构

```
├── aipivot.go                          # 服务入口
├── api/                                # API 定义层（API-First）
│   ├── entry.api                       #   入口文件，import 所有子模块
│   ├── comm.api                        #   通用类型（CommResponse）
│   ├── infra.api                       #   基础设施接口（health/ready/metrics/ping）
│   └── auth.api                        #   认证接口（login）
├── etc/
│   └── aipivot-api.yaml                # 运行时配置
├── migrations/
│   ├── 000001_init_tenants.up.sql      # 多租户基础表（tenants/users/api_keys）
│   └── 000001_init_tenants.down.sql
├── cmd/gen/main.go                     # GORM Gen 代码生成入口
├── internal/
│   ├── config/config.go                # 配置结构体
│   ├── svc/servicecontext.go           # 服务上下文：依赖注入容器
│   ├── handler/                        # HTTP 处理器层（goctl 生成）
│   │   ├── routes.go                   #   路由注册（goctl 生成，禁止编辑）
│   │   ├── auth/                       #   认证模块 handler
│   │   └── infra/                      #   基础设施 handler
│   ├── logic/                          # 业务逻辑层（goctl 脚手架 + 手动实现）
│   │   ├── auth/loginLogic.go          #   登录逻辑
│   │   └── infra/                      #   健康检查/指标/ping 逻辑
│   ├── types/types.go                  # 请求/响应类型（goctl 生成，禁止编辑）
│   ├── modules/                        # ★ 业务模块层（按领域划分）
│   │   └── auth/                       #   认证模块
│   │       ├── domain/model/           #     领域模型 + 校验（CheckEmail/CheckPassword）
│   │       ├── domain/assembler/       #     DTO 转换器（Request↔Model↔PO↔Show）
│   │       ├── repo/                   #     仓储层（接口 + 实现）
│   │       │   ├── interface.go        #       UserRepository / TenantRepository 接口
│   │       │   ├── user_repo.go        #       用户仓储实现
│   │       │   ├── tenant_repo.go      #       租户仓储实现
│   │       │   └── dao/               #       数据访问对象（最底层 GORM 操作）
│   │       ├── jwt.go                  #     JWT 令牌生成
│   │       └── middleware.go           #     认证中间件
│   ├── shared/                         # ★ 跨模块共享层
│   │   ├── po/                         #   持久化对象（手动定义，GORM 模型）
│   │   │   ├── tenant.go
│   │   │   ├── user.go
│   │   │   └── api_key.go
│   │   ├── query/                      #   GORM Gen 生成的类型安全查询（禁止手动编辑）
│   │   ├── errorx/                     #   统一错误处理（BusinessError + 全局拦截）
│   │   └── response/                   #   统一响应工具
│   ├── infra/                          # 基础设施层
│   │   ├── health.go                   #   依赖健康检查
│   │   ├── postgres.go                 #   PostgreSQL 初始化 + 连接池
│   │   ├── redis.go                    #   Redis 初始化
│   │   └── migrate.go                  #   数据库迁移（golang-migrate）
│   └── observability/                  # 可观测性层
│       ├── tracing.go                  #   OpenTelemetry 初始化
│       ├── metrics.go                  #   Prometheus 指标定义
│       ├── middleware.go               #   统一中间件（RequestID/Tracing/Metrics/Log）
│       ├── context.go                  #   RequestID 上下文传递
│       └── responsewriter.go           #   ResponseWriter 包装器
├── deploy/
│   ├── docker-compose.yml              # 本地依赖编排
│   ├── prometheus/prometheus.yml
│   └── grafana/                        # Grafana 自动 provisioning
├── docs/
│   ├── project-design-spec.md          # ★ 工程设计规范（六层架构 + 命名规范 + SOP）
│   ├── swagger/aipivot.json            # Swagger 文档
│   └── 产品需求.md
└── Makefile                            # 自动化命令
```

## 架构设计

### 分层模式（六层）

```
API 定义层    api/*.api                      ← 接口契约，API-First
Handler 层    internal/handler/{group}/      ← HTTP 入口，参数解析（goctl 生成）
Logic 层      internal/logic/{group}/        ← 业务编排（调用 modules 层）
Module 层     internal/modules/{module}/     ← 领域模型 + 仓储 + 模块专属逻辑
Shared 层     internal/shared/              ← PO / Query / errorx / response
Infra 层      internal/infra/              ← DB/Redis/迁移/健康检查
```

### 依赖注入链路

```
DB → Query(GORM Gen) → DAO → Repo(接口) → ServiceContext → Logic
```

- ServiceContext 中 Repo 字段使用**接口类型**（便于 Mock 测试）
- Logic 通过 Repo 接口操作数据，**禁止**穿透 Repo 直接访问 DAO

### 启动流程

1. 加载 YAML 配置
2. 初始化 OpenTelemetry 追踪
3. 运行数据库迁移（golang-migrate）
4. 初始化 PostgreSQL 连接池
5. 初始化 Redis
6. 初始化 Prometheus 指标
7. 组装 DAO → Repo → ServiceContext
8. 注册全局错误处理 + 可观测性中间件
9. 注册路由 → 启动服务

### API 端点

| 端点 | 方法 | 说明 |
|---|---|---|
| `/healthz` | GET | 存活检查（K8s livenessProbe） |
| `/readyz` | GET | 就绪检查（PG + Redis 依赖探测） |
| `/metrics` | GET | Prometheus 指标暴露 |
| `/v1/ping` | GET | 连通性测试（pong + traceId + requestId） |
| `/api/v1/auth/login` | POST | 用户登录（邮箱 + 密码 → JWT） |

### 数据库表

| 表 | 说明 |
|---|---|
| `tenants` | 租户表（多租户隔离的最小单元） |
| `users` | 用户表（归属租户，租户内 email 唯一） |
| `api_keys` | API 密钥表（程序化访问凭证） |

## 代码生成

### goctl（API 层）

```bash
make api    # goctl api go -api api/entry.api -dir . -style goZero
```

**入口文件**：`api/entry.api`（import 所有子模块 .api 文件）

| 生成产物 | 可否编辑 |
|---|---|
| `internal/types/types.go` | ❌ 禁止编辑 |
| `internal/handler/routes.go` | ❌ 禁止编辑 |
| `internal/handler/{group}/*.go` | ✅ 通常无需改动 |
| `internal/logic/{group}/*.go` | ✅ **主要业务实现区** |

### GORM Gen（Query 层）

```bash
make gen    # go run cmd/gen/main.go
```

- PO 模型定义在 `internal/shared/po/`（手动维护）
- 生成产物在 `internal/shared/query/`（**禁止手动编辑**）

### Swagger 文档

```bash
make swagger
```

## 新增业务模块 SOP

1. **建表**：在 `migrations/` 添加迁移 SQL（遵循 PostgreSQL DDL 规范：COMMENT ON 必须有）
2. **定义 PO**：在 `internal/shared/po/` 添加 GORM 模型
3. **生成 Query**：`make gen`
4. **定义 API**：在 `api/` 添加 `.api` 文件，在 `entry.api` 中 import
5. **生成代码**：`make api`
6. **编写 Module**：在 `internal/modules/{module}/` 下创建：
   - `domain/model/` — 领域模型 + Check 校验方法
   - `domain/assembler/` — 三方向转换（Request→Model, Model→PO, PO→Show）
   - `repo/interface.go` — 仓储接口定义
   - `repo/dao/` — DAO 实现
   - `repo/{entity}_repo.go` — Repo 实现
7. **注册依赖**：在 `internal/svc/servicecontext.go` 中组装并注入
8. **实现 Logic**：在 `internal/logic/{group}/` 中编写业务编排

## 本地开发

```bash
# 启动依赖服务
docker compose -f deploy/docker-compose.yml up -d

# 运行服务（数据库迁移自动执行）
go run aipivot.go -f etc/aipivot-api.yaml

# 测试 & 编译
go test ./...
go build ./...

# 一键重新生成所有代码
make regen
```

| 服务 | 地址 |
|---|---|
| AIPivot API | `http://127.0.0.1:8888` |
| Jaeger UI | `http://127.0.0.1:16686` |
| Prometheus | `http://127.0.0.1:9090` |
| Grafana | `http://127.0.0.1:3000`（admin/admin） |

## 关键约定

1. **API-First**：任何接口变更必须先修改 `.api` 文件，再通过 goctl 生成代码
2. **分层严格**：Handler 只做参数绑定 + 响应输出；Logic 只做业务编排；Domain 承载校验规则；Repo 封装数据访问
3. **禁止跨层**：Logic 禁止直接访问 DAO / SQL / GORM；Handler 禁止编写业务逻辑
4. **Assembler 三方向覆盖**：Request→Model、Model→PO、PO→Show，禁止在 Logic 中手动构造 PO
5. **错误处理**：业务错误用 `errorx.NewBusinessError`（HTTP 200 + code），系统错误用 `errorx.NewInternalError`
6. **可观测性贯穿**：每个请求自动包含 RequestID、TraceID、Metrics、结构化日志
7. **优先本地验证**：`go test ./...` + `go build ./...` 先通过再推进
8. **注释规范**：复杂逻辑关键决策点注释 WHY，不注释 WHAT；禁止逐行注释
9. **DDL 规范**：PostgreSQL 建表必须有 `COMMENT ON TABLE` + `COMMENT ON COLUMN`
10. **设计规范文档**：详细架构与命名规范参见 `docs/project-design-spec.md`
