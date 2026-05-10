# AIPivot 项目说明

## 项目概述

AIPivot 是一个基于 Go 语言的 **多租户 AI 能力中台**，使用 go-zero 框架构建 REST API 服务。项目采用多租户隔离架构，所有业务数据通过 `tenant_id` 关联。当前已完成：**运行时基础设施 + 可观测性 + 数据库迁移 + 认证模块 + 知识库模块 + Chat 模块（SSE 流式）+ LLM/RAG 集成 + 前端 Chat Widget 原型**。

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
| pgvector/pgvector:pg16 | 数据库（含 pgvector 向量扩展） |
| Redis 7 | 缓存 + Asynq 任务队列 |
| Asynq | 异步任务（文档切块 + Embedding） |
| Docker Compose | 本地开发环境（PG/Redis/Jaeger/Prometheus/Grafana） |
| React 18 + Vite 6 | 前端框架 + 构建工具 |
| TailwindCSS 3 + Lucide React | 前端样式 + 图标 |
| Zustand 5 | 前端状态管理（auth + chat，localStorage 持久化） |

## 项目结构

```
├── aipivot.go                          # 服务入口
├── api/                                # API 定义层（API-First）
│   ├── entry.api                       #   入口文件，import 所有子模块
│   ├── comm.api                        #   通用类型（CommResponse）
│   ├── infra.api                       #   基础设施接口（health/ready/metrics/ping）
│   ├── auth.api                        #   认证接口（login/register）
│   ├── chat.api                        #   会话 + 消息接口（含 SSE 流式）
│   └── knowledge.api                   #   知识库 CRUD + 文档上传
├── etc/
│   └── aipivot-api.yaml                # 运行时配置
├── migrations/
│   ├── 000001_init_tenants.{up,down}.sql     # 多租户基础表（tenants/users/api_keys）
│   ├── 000002_knowledge_base.{up,down}.sql   # 知识库表（knowledge_bases/documents/document_chunks + pgvector）
│   └── 000003_conversations.{up,down}.sql    # 会话表（conversations/messages）
├── cmd/gen/main.go                     # GORM Gen 代码生成入口
├── pkg/                                # 可复用工具包
│   ├── chunker/chunker.go              #   文本切块器
│   └── llm/                            #   LLM 客户端（OpenAI 兼容 + SSE 流式）
├── internal/
│   ├── config/config.go                # 配置结构体
│   ├── svc/servicecontext.go           # 服务上下文：依赖注入容器
│   ├── handler/                        # HTTP 处理器层（goctl 生成）
│   │   ├── routes.go                   #   路由注册（goctl 生成，禁止编辑）
│   │   ├── auth/                       #   认证模块 handler
│   │   ├── chat/                       #   会话/消息 handler（含 SSE 手写 handler）
│   │   ├── knowledge/                  #   知识库 handler
│   │   └── infra/                      #   基础设施 handler
│   ├── logic/                          # 业务逻辑层（goctl 脚手架 + 手动实现）
│   │   ├── auth/                       #   登录/注册逻辑
│   │   ├── chat/                       #   会话 CRUD + 消息发送（同步/SSE 流式）
│   │   ├── knowledge/                  #   知识库 CRUD + 文档上传
│   │   └── infra/                      #   健康检查/指标/ping 逻辑
│   ├── types/types.go                  # 请求/响应类型（goctl 生成，禁止编辑）
│   ├── modules/                        # ★ 业务模块层（按领域划分）
│   │   ├── auth/                       #   认证模块
│   │   │   ├── domain/model/           #     领域模型 + 校验
│   │   │   ├── domain/assembler/       #     DTO 转换器
│   │   │   ├── repo/                   #     仓储层（接口 + 实现 + DAO）
│   │   │   ├── jwt.go                  #     JWT 令牌生成
│   │   │   └── middleware.go           #     认证中间件
│   │   ├── chat/                       #   会话模块
│   │   │   ├── domain/assembler/       #     Conversation/Message 转换器
│   │   │   └── repo/                   #     ConversationRepo / MessageRepo
│   │   └── knowledge/                  #   知识库模块
│   │       ├── domain/model/           #     知识库领域模型
│   │       ├── domain/assembler/       #     KnowledgeBase/Document 转换器
│   │       └── repo/                   #     KnowledgeBaseRepo / DocumentRepo / ChunkRepo
│   ├── shared/                         # ★ 跨模块共享层
│   │   ├── po/                         #   持久化对象（手动定义，GORM 模型）
│   │   ├── query/                      #   GORM Gen 生成的类型安全查询（禁止手动编辑）
│   │   ├── errorx/                     #   统一错误处理（BusinessError + 全局拦截）
│   │   ├── sse/                        #   SSE Writer（流式响应工具）
│   │   └── response/                   #   统一响应工具
│   ├── infra/                          # 基础设施层
│   │   ├── health.go                   #   依赖健康检查
│   │   ├── postgres.go                 #   PostgreSQL 初始化 + 连接池
│   │   ├── redis.go                    #   Redis 初始化
│   │   └── migrate.go                  #   数据库迁移（golang-migrate）
│   ├── worker/                         # 异步任务（Asynq）
│   ├── rag/                            # RAG 服务（检索增强生成）
│   └── observability/                  # 可观测性层
├── web/                                # ★ 前端（React + Vite）
│   ├── src/
│   │   ├── main.tsx                    #   入口
│   │   ├── App.tsx                     #   根组件（登录/聊天路由）
│   │   ├── pages/LoginPage.tsx         #   登录/注册页
│   │   ├── pages/ChatPage.tsx          #   聊天页（侧边栏 + 消息面板 + 输入区）
│   │   ├── store/auth.ts               #   Zustand auth store（JWT 持久化）
│   │   ├── store/chat.ts               #   Zustand chat store（会话/消息/流式状态）
│   │   └── lib/api.ts                  #   API 客户端 + SSE 流式解析
│   ├── vite.config.ts                  #   Vite 配置（proxy /api → 后端 8888）
│   └── package.json
├── deploy/
│   ├── docker-compose.yml              # 本地依赖编排（pgvector + Redis + Jaeger + Prometheus + Grafana）
│   ├── prometheus/prometheus.yml
│   └── grafana/                        # Grafana 自动 provisioning
├── docs/
│   ├── project-design-spec.md          # ★ 工程设计规范（六层架构 + 命名规范 + SOP）
│   ├── architecture-review.md          # ★ 架构评审 + 实施路线图 + 开发日志
│   ├── swagger/aipivot.json            # Swagger 文档
│   └── 产品需求.md
├── Makefile                            # 自动化命令（后端 + 前端）
└── .gitignore
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
| `/v1/ping` | GET | 连通性测试 |
| `/api/v1/auth/login` | POST | 用户登录（邮箱 + 密码 → JWT） |
| `/api/v1/auth/register` | POST | 用户注册 |
| `/api/v1/conversations` | GET/POST | 会话列表 / 创建会话 |
| `/api/v1/conversations/:convId` | GET/PUT/DELETE | 会话详情 / 更新 / 删除 |
| `/api/v1/conversations/:convId/messages` | GET | 消息列表 |
| `/api/v1/conversations/:convId/messages/send` | POST | 发送消息（同步） |
| `/api/v1/conversations/:convId/messages/stream` | POST | 发送消息（**SSE 流式**） |
| `/api/v1/knowledge-bases` | GET/POST | 知识库列表 / 创建 |
| `/api/v1/knowledge-bases/:kbId` | GET/PUT/DELETE | 知识库详情 / 更新 / 删除 |
| `/api/v1/knowledge-bases/:kbId/documents` | GET | 文档列表 |
| `/api/v1/knowledge-bases/:kbId/documents/upload` | POST | 上传文档 |

### 数据库表

| 表 | 说明 |
|---|---|
| `tenants` | 租户表（多租户隔离的最小单元） |
| `users` | 用户表（归属租户，租户内 email 唯一） |
| `api_keys` | API 密钥表（程序化访问凭证） |
| `knowledge_bases` | 知识库表 |
| `documents` | 文档表（归属知识库，含处理状态） |
| `document_chunks` | 文档切块表（含 pgvector 1536 维向量 + HNSW 索引） |
| `conversations` | 会话表（可关联知识库） |
| `messages` | 消息表（user/assistant 角色，含 token 统计） |

### SSE 流式协议

前端通过 `POST /api/v1/conversations/:convId/messages/stream` 发送消息，后端以 SSE 格式推送：

| 事件 | 数据 | 说明 |
|---|---|---|
| `message_start` | `{messageId, conversationId}` | 流开始 |
| `delta` | `{content}` | 增量 token |
| `message_end` | `{messageId, model, tokenCount, latencyMs, sources}` | 流结束元数据 |
| `error` | `{code, msg}` | 错误 |
| `done` | `[DONE]` | 流关闭信号 |

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

# 运行后端（数据库迁移自动执行）
go run aipivot.go -f etc/aipivot-api.yaml

# 安装前端依赖 + 启动开发服务器
make web-install
make web-dev

# 测试 & 编译
go test ./...
go build ./...

# 一键重新生成所有代码
make regen
```

### Makefile 命令

| 命令 | 说明 |
|---|---|
| `make gen` | GORM Gen 代码生成 |
| `make api` | goctl API 代码生成 |
| `make build` | 构建后端二进制 |
| `make test` | 运行测试 |
| `make tidy` | 依赖整理 |
| `make clean` | 清理构建产物（bin/ + web/dist/） |
| `make swagger` | 生成 Swagger 文档 |
| `make regen` | 一键重新生成所有代码 |
| `make web-install` | 安装前端依赖 |
| `make web-dev` | 启动前端开发服务器（5173，proxy → 8888） |
| `make web-build` | 构建前端生产包 |
| `make web-preview` | 预览前端生产构建 |

### 服务地址

| 服务 | 地址 |
|---|---|
| AIPivot API（后端） | `http://127.0.0.1:8888` |
| 前端开发服务器 | `http://127.0.0.1:5173` |
| Jaeger UI | `http://127.0.0.1:16686` |
| Prometheus | `http://127.0.0.1:9090` |
| Grafana | `http://127.0.0.1:3000`（admin/admin） |

## 关键约定

1. **API-First**：任何接口变更必须先修改 `.api` 文件，再通过 goctl 生成代码
2. **分层严格**：Handler 只做参数绑定 + 响应输出；Logic 只做业务编排；Domain 承载校验规则；Repo 封装数据访问
3. **禁止跨层**：Logic 禁止直接访问 DAO / SQL / GORM；Handler 禁止编写业务逻辑
4. **Assembler 三方向覆盖**：Request→Model、Model→PO、PO→Show；**创建 PO 时必须显式设置 UUID 字段**（`uuid.New().String()`），不能依赖数据库默认值（GORM 会传空字符串覆盖）
5. **错误处理**：业务错误用 `errorx.NewBusinessError`（HTTP 200 + code），系统错误用 `errorx.NewInternalError`
6. **可观测性贯穿**：每个请求自动包含 RequestID、TraceID、Metrics、结构化日志
7. **优先本地验证**：`go test ./...` + `go build ./...` 先通过再推进
8. **注释规范**：复杂逻辑关键决策点注释 WHY，不注释 WHAT；禁止逐行注释
9. **DDL 规范**：PostgreSQL 建表必须有 `COMMENT ON TABLE` + `COMMENT ON COLUMN`
10. **设计规范文档**：详细架构与命名规范参见 `docs/project-design-spec.md`
11. **npm 国内源**：已全局配置 `registry=https://registry.npmmirror.com`
12. **Docker 镜像**：PostgreSQL 必须使用 `pgvector/pgvector:pg16`（含 vector 扩展），不能用原版 `postgres:16`
