# AIPivot 项目说明

## 项目概述

AIPivot 是一个基于 Go 语言的 **多租户 AI 能力中台**，使用 go-zero 框架构建 REST API 服务。项目采用多租户隔离架构，所有业务数据通过 `tenant_id` 关联。

当前已完成：**运行时基础设施 + 可观测性 + 数据库迁移（1→10）+ 认证/租户/用户管理 + 知识库（pgvector RAG）
+ Chat（同步 + SSE 流式）+ LLM 集成（Ark Responses API）+ Skills/Agent 编排 + 可视化 Flow（编辑 + 试运行 + 执行历史）
+ Webhook 出入站 + API Key（`sk_`/`pk_`）+ 数据分析看板 + React 管理台（8 个页面）+ Chat Widget SDK（Preact，gzip 17KB）**。

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
│   ├── knowledge.api                   #   知识库 CRUD + 文档上传
│   ├── admin.api                       #   管理台：租户 + 用户管理（AdminMiddleware）
│   ├── analytics.api                   #   数据分析：overview / daily / export
│   ├── flows.api                       #   可视化 Flow CRUD + 试运行 + 执行历史（AdminMiddleware）
│   ├── skills.api                      #   技能 CRUD（AdminMiddleware）
│   ├── models.api                      #   可用模型列表
│   └── open.api                        #   开放接入：chat/completions + webhook + Widget + API Key
├── etc/
│   └── aipivot-api.yaml                # 运行时配置
├── migrations/                         # golang-migrate SQL（启动时自动执行到最新版本）
│   ├── 000001_init_tenants                 # 多租户基础表（tenants/users/api_keys）
│   ├── 000002_knowledge_base               # 知识库表（knowledge_bases/documents/document_chunks + pgvector）
│   ├── 000003_conversations                # 会话表（conversations/messages）
│   ├── 000004_conversation_model           # 会话模型字段
│   ├── 000005_webhooks                     # Webhook 表
│   ├── 000006_skills                       # 技能表
│   ├── 000007_flows                        # 可视化 Flow 表
│   ├── 000008_widget_public_key            # Widget public key（pk_ 前缀 + allowed_origins）
│   ├── 000009_widget_feedback_suggested    # 满意度评分 + 引导问答
│   └── 000010_flow_runs                    # Flow 执行记录表（全量快照）
├── cmd/gen/main.go                     # GORM Gen 代码生成入口
├── pkg/                                # 可复用工具包
│   ├── chunker/chunker.go              #   文本切块器
│   └── llm/                            #   LLM 客户端（OpenAI 兼容 + SSE 流式）
├── internal/
│   ├── config/config.go                # 配置结构体
│   ├── svc/servicecontext.go           # 服务上下文：依赖注入容器
│   ├── handler/                        # HTTP 处理器层（goctl 生成）
│   │   ├── routes.go                   #   路由注册（goctl 生成，禁止编辑）
│   │   └── {group}/                    #   auth / chat / knowledge / admin / analytics / apikey
│   │                                   #   flows / infra / models / open / skills / webhook
│   │                                   #   （chat/flows/open 内含手写 SSE handler）
│   ├── logic/                          # 业务逻辑层（goctl 脚手架 + 手动实现）
│   │   └── {group}/                    #   与 handler 一一对应的 12 个分组，主要业务实现区
│   ├── types/types.go                  # 请求/响应类型（goctl 生成，禁止编辑）
│   ├── modules/                        # ★ 业务模块层（领域模型 + 校验 + Assembler，扁平文件）
│   │   ├── auth/                       #   认证：auth.go / assembler.go / jwt.go
│   │   │                               #        jwt_middleware.go / apikey_middleware.go / context.go
│   │   ├── chat/                       #   会话：chat.go / assembler.go
│   │   ├── knowledge/                  #   知识库：knowledge.go / assembler.go
│   │   ├── channel/                    #   渠道：channel.go + webhook/（投递服务）
│   │   ├── flow/                       #   Flow 引擎：engine.go / graph.go / executors.go
│   │   │                               #        expression.go / blackboard.go / definition.go
│   │   ├── rag/                        #   RAG：service.go / stream.go
│   │   └── agent/                      #   Agent：agent.go / orchestrator.go / registry.go / tools/
│   ├── repository/                     # ★ 仓储层（独立顶层目录，非 modules 子目录）
│   │   ├── auth/                       #   user.go / tenant.go / apikey.go
│   │   ├── chat/                       #   conversation.go / message.go
│   │   ├── knowledge/                  #   kb.go / document.go / chunk.go
│   │   ├── flow/  flowrun/             #   flow.go / flowrun.go
│   │   ├── skill/                      #   skill.go
│   │   └── webhook/                    #   webhook.go
│   ├── middleware/                     # HTTP 中间件（go-zero 注册）
│   │   ├── authMiddleware.go           #   JWT 认证
│   │   ├── adminMiddleware.go          #   JWT + role=admin（403 拦截）
│   │   ├── apiKeyMiddleware.go         #   API Key / public key 认证
│   │   ├── cors.go                     #   跨域
│   │   └── rate_limit.go               #   限流
│   ├── shared/                         # ★ 跨模块共享层
│   │   ├── po/                         #   持久化对象（手动定义，GORM 模型）
│   │   ├── query/                      #   GORM Gen 生成的类型安全查询（禁止手动编辑）
│   │   ├── errorx/                     #   统一错误处理（BusinessError + 全局拦截）
│   │   ├── sse/                        #   SSE Writer（流式响应工具）
│   │   ├── ratelimit/                  #   Token 日配额 + 滑动窗口限流
│   │   └── response/                   #   统一响应工具
│   ├── infra/                          # 基础设施层
│   │   ├── health.go                   #   依赖健康检查
│   │   ├── postgres.go                 #   PostgreSQL 初始化 + 连接池
│   │   ├── redis.go                    #   Redis 初始化
│   │   └── migrate.go                  #   数据库迁移（golang-migrate）
│   ├── worker/                         # 异步任务（Asynq）
│   └── observability/                  # 可观测性层
├── web/                                # ★ 前端（React + Vite，管理台）
│   ├── src/
│   │   ├── main.tsx                    #   入口
│   │   ├── App.tsx                     #   根组件（登录态 + 页面路由）
│   │   ├── pages/                      #   Login / Chat / Knowledge / Flow / Skill
│   │   │                               #   Webhook / Analytics / Admin（共 8 页）
│   │   ├── store/auth.ts               #   Zustand auth store（JWT 持久化）
│   │   ├── store/chat.ts               #   Zustand chat store（会话/消息/流式状态）
│   │   └── lib/api.ts                  #   API 客户端 + SSE 流式解析
│   ├── vite.config.ts                  #   Vite 配置（proxy /api → 后端 8888）
│   └── package.json
├── widget/                             # ★ Chat Widget SDK（Preact + Vite lib + Shadow DOM）
│   ├── src/
│   │   ├── index.tsx                   #   SDK 入口，导出 init() 挂载到 window.AIPivotWidget
│   │   ├── widget.tsx                  #   主组件（编排会话初始化/SSE 流/状态）
│   │   ├── client.ts                   #   WidgetClient（createSession/sendMessageStream/listMessages）
│   │   ├── store.ts                    #   Zustand 状态机（提炼自 web/src/store/chat.ts）
│   │   ├── storage.ts                  #   sessionToken + visitorId 持久化
│   │   ├── components/                 #   Launcher/ChatPanel/MessageList/MessageBubble/InputArea/TypingIndicator
│   │   ├── utils/                      #   sse 解析 / dom（Shadow DOM）/ escape（XSS）/ uuid / retry
│   │   └── styles/index.css            #   TailwindCSS + 打字光标动画
│   ├── examples/                       #   minimal.html / advanced.html 接入示例
│   ├── vite.config.ts                  #   lib 模式 IIFE 输出，CSS 内联
│   └── package.json
├── deploy/
│   ├── docker-compose.yml              # 本地依赖编排（pgvector + Redis + Jaeger + Prometheus + Grafana）
│   ├── prometheus/prometheus.yml
│   └── grafana/                        # Grafana 自动 provisioning
├── docs/
│   ├── usage-guide.md                  # ★ 本地启动 / 联调 / 排错手册（日常先看这个）
│   ├── project-design-spec.md          # ★ 工程设计规范（分层架构 + 命名规范 + SOP）
│   ├── architecture-review.md          # ★ 架构评审 + 实施路线图 + 开发日志
│   ├── swagger/aipivot.json            # Swagger 文档
│   └── 产品需求.md
├── Makefile                            # 自动化命令（后端 + 前端）
└── .gitignore
```

## 架构设计

### 分层模式（七层）

```
API 定义层    api/*.api                      ← 接口契约，API-First
Handler 层    internal/handler/{group}/      ← HTTP 入口，参数解析（goctl 生成）
Logic 层      internal/logic/{group}/        ← 业务编排（调用 modules + repository）
Module 层     internal/modules/{module}/     ← 领域模型 + 校验 + Assembler + 仓储接口（端口）
Repository 层 internal/repository/{module}/  ← 仓储接口实现（适配器，直接操作 query）
Shared 层     internal/shared/               ← PO / Query / errorx / response / sse / ratelimit
Infra 层      internal/infra/                ← DB/Redis/迁移/健康检查
```

**端口与适配器**：仓储接口（如 `chat.MessageRepository`）定义在 `internal/modules/{module}/{module}.go`，
实现（如 `chat.MessageRepo`）在 `internal/repository/{module}/`。ServiceContext 持有**接口类型**，便于 Mock。

### 依赖注入链路

```
DB → Query(GORM Gen) → Repo 实现(internal/repository) → Repo 接口(internal/modules) → ServiceContext → Logic
```

- ServiceContext 中 Repo 字段使用**接口类型**（便于 Mock 测试）
- Logic 通过 Repo 接口操作数据，**禁止**在 Logic 中直接写 GORM / SQL
- 项目中**没有独立 DAO 层**，Repo 实现直接调用 `internal/shared/query` 的类型安全查询

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

> **路径参数一律是数字自增 ID**（`.api` 中声明为 `int64`），不是 UUID。传 UUID 会得到
> `strconv.ParseInt: invalid syntax` 的 400 错误。列表端点必须显式带 `page` / `pageSize`。

**基础设施（无鉴权）**

| 端点 | 方法 | 说明 |
|---|---|---|
| `/healthz` | GET | 存活检查（K8s livenessProbe） |
| `/readyz` | GET | 就绪检查（PG + Redis 依赖探测） |
| `/metrics` | GET | Prometheus 指标暴露 |
| `/v1/ping` | GET | 连通性测试 |

**认证（无鉴权）**

| 端点 | 方法 | 说明 |
|---|---|---|
| `/api/v1/auth/login` | POST | 用户登录（邮箱 + 密码 → JWT） |
| `/api/v1/auth/register` | POST | 用户注册（自动绑定租户 1，角色固定 `member`） |

**会话与消息**（`AuthMiddleware`）

| 端点 | 方法 | 说明 |
|---|---|---|
| `/api/v1/conversations` | POST | 创建会话 |
| `/api/v1/conversations` | GET | 会话列表（必传 `page` / `pageSize`） |
| `/api/v1/conversations/:id` | GET | 会话详情 |
| `/api/v1/conversations/:id/close` | PUT | 关闭会话 |
| `/api/v1/conversations/:id/escalate` | PUT | 升级为人工 |
| `/api/v1/conversations/:convId/messages` | POST | 发送消息（同步，非 `/send`） |
| `/api/v1/conversations/:convId/messages/stream` | POST | 发送消息（**SSE 流式**） |
| `/api/v1/conversations/:convId/messages` | GET | 消息列表 |

**知识库**（`AuthMiddleware`）

| 端点 | 方法 | 说明 |
|---|---|---|
| `/api/v1/knowledge-bases` | GET/POST | 知识库列表 / 创建 |
| `/api/v1/knowledge-bases/:id` | GET/PUT/DELETE | 详情 / 更新 / 删除 |
| `/api/v1/knowledge-bases/:kbId/documents` | POST | 上传文档（非 `/documents/upload`） |
| `/api/v1/knowledge-bases/:kbId/documents` | GET | 文档列表 |
| `/api/v1/knowledge-bases/:kbId/documents/:id` | DELETE | 删除文档 |

**模型与分析**（`AuthMiddleware`）

| 端点 | 方法 | 说明 |
|---|---|---|
| `/api/v1/models` | GET | 可用模型列表（读 `LLM.ChatModels` / `EmbeddingModels` 配置） |
| `/api/v1/analytics/overview` | GET | 概览 KPI（含满意度） |
| `/api/v1/analytics/daily` | GET | 按日趋势 |
| `/api/v1/analytics/export` | GET | 导出 |

**Webhook 与 API Key**（`AuthMiddleware`）

| 端点 | 方法 | 说明 |
|---|---|---|
| `/api/v1/webhooks` | GET/POST | Webhook 列表 / 创建 |
| `/api/v1/webhooks/:id` | GET/PUT/DELETE | 详情 / 更新 / 删除 |
| `/api/v1/api-keys` | GET/POST | API Key 列表 / 创建（`sk_` master / `pk_` public） |
| `/api/v1/api-keys/:id/revoke` | PUT | 吊销 API Key |

**管理台**（`AdminMiddleware` — 非 `role=admin` 返回 403）

| 端点 | 方法 | 说明 |
|---|---|---|
| `/api/v1/admin/tenant` | GET/PUT | 租户信息 / 更新 |
| `/api/v1/admin/users` | GET/POST | 用户列表 / 创建 |
| `/api/v1/admin/users/:id` | PUT/DELETE | 更新 / 删除用户 |

**技能与 Flow**（`AdminMiddleware`）

| 端点 | 方法 | 说明 |
|---|---|---|
| `/api/v1/skills` | GET/POST | 技能列表 / 创建 |
| `/api/v1/skills/:id` | GET/PUT/DELETE | 详情 / 更新 / 删除 |
| `/api/v1/flows` | GET/POST | Flow 列表 / 创建 |
| `/api/v1/flows/:id` | GET/PUT/DELETE | 详情 / 更新 / 删除 |
| `/api/v1/flows/:id/run` | POST | **Flow 试运行**（SSE：run_start/node_start/delta/node_end/run_end） |
| `/api/v1/flows/:id/runs` | GET | Flow 执行历史列表（flow_runs 快照） |

**开放接入 `/api/v1/open`**

| 端点 | 方法 | 中间件 | 说明 |
|---|---|---|---|
| `/api/v1/open/chat/completions` | POST | ⚠️ 无 | OpenAI 兼容对话（**遗留技术债：未挂鉴权**） |
| `/api/v1/open/webhook/:webhookId/inbound` | POST | ⚠️ 无 | Webhook 入站（**遗留技术债：未挂鉴权**） |
| `/api/v1/open/widget/sessions` | POST | ApiKey | **Widget**：创建访客会话（`pk_` public key + Origin 白名单） |
| `/api/v1/open/widget/sessions/:sessionToken/messages` | GET | ApiKey | **Widget**：拉取历史消息 |
| `/api/v1/open/widget/sessions/:sessionToken/messages/stream` | POST | ApiKey | **Widget**：流式发送消息（SSE，持久化 user + assistant） |
| `/api/v1/open/widget/sessions/:sessionToken/messages/:messageId/feedback` | PUT | ApiKey | **Widget**：消息满意度评分（up/down） |




### 数据库表

| 表 | 说明 |
|---|---|
| `tenants` | 租户表（多租户隔离的最小单元） |
| `users` | 用户表（归属租户，租户内 email 唯一） |
| `api_keys` | API 密钥表（程序化访问凭证） |
| `knowledge_bases` | 知识库表 |
| `documents` | 文档表（归属知识库，含处理状态） |
| `document_chunks` | 文档切块表（含 pgvector 1536 维向量 + HNSW 索引） |
| `conversations` | 会话表（可关联知识库；`external_user_id` 复用于 Widget 访客关联） |
| `messages` | 消息表（user/assistant 角色，含 token 统计 + `rating` 满意度） |
| `flows` | 可视化 Flow 定义表（`definition` JSONB：nodes + edges） |
| `flow_runs` | Flow 执行记录表（全量快照：node_results + flow_version，历史回放不受后续编辑污染） |
| `skills` | 技能表（租户自定义工具） |
| `webhooks` | Webhook 表（出站投递配置） |

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
6. **编写 Module**：在 `internal/modules/{module}/` 下创建（**扁平文件，无 domain/ 子目录**）：
   - `{module}.go` — 领域模型 + Check 校验方法 + **仓储接口定义（端口）**
   - `assembler.go` — 三方向转换（Request→Model, Model→PO, PO→Show）
7. **实现 Repository**：在 `internal/repository/{module}/{entity}.go` 实现上一步的接口，
   直接调用 `internal/shared/query`（**没有 DAO 层**）
8. **注册依赖**：在 `internal/svc/servicecontext.go` 中组装并注入（字段用接口类型）
9. **实现 Logic**：在 `internal/logic/{group}/` 中编写业务编排

## 本地开发

> 完整的启动 / 联调 / 排错手册见 **`docs/usage-guide.md`**。

```bash
# 启动依赖服务
docker compose -f deploy/docker-compose.yml up -d

# 配置 LLM Key（.env 已在 .gitignore 中，内容：ARK_API_KEY=xxx）
set -a && . ./.env && set +a

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
| `make widget-install` | 安装 Widget SDK 依赖 |
| `make widget-dev` | 启动 Widget 开发服务器（5174） |
| `make widget-build` | 构建 Widget SDK 单文件（dist/aipivot-widget.js，gzip ≈ 16KB） |
| `make widget-preview` | 预览 Widget 生产构建 |

### 服务地址

| 服务 | 地址 |
|---|---|
| AIPivot API（后端） | `http://127.0.0.1:8888` |
| 前端开发服务器 | `http://127.0.0.1:5173` |
| Jaeger UI | `http://127.0.0.1:16686` |
| Prometheus | `http://127.0.0.1:9090` |
| Grafana | `http://127.0.0.1:3000`（admin/admin） |

## 关键约定

1. **API-First 单一真相源**：任何路由/请求/响应类型必须先在 `api/*.api` 中声明并由 goctl 生成。`routes.go` 和 `types.go` 已是纯 goctl 生成（DO NOT EDIT 标记），**严禁手写混入**——新增端点漏在 .api 中声明会导致 `make api` 后路由丢失
2. **分层严格**：Handler 只做参数绑定 + 响应输出；Logic 只做业务编排；Domain 承载校验规则；Repo 封装数据访问
3. **禁止跨层**：Logic 禁止直接访问 DAO / SQL / GORM；Handler 禁止编写业务逻辑；**Handler 只能依赖 `internal/types` 的契约类型**，禁止反向引用 logic 包的私有结构体
4. **手写 Handler 边界**：SSE 流式 / 二进制文件下载 / WebSocket 等 goctl 无法生成标准 JSON 响应的场景，允许手写 handler 实现，但必须满足两个约束 —— (a) 端点先在对应 `.api` 文件中声明（让 goctl 生成路由注册与参数解析）；(b) 请求/响应类型仍走 `internal/types`，不引入 logic 私有类型
5. **Assembler 三方向覆盖**：Request→Model、Model→PO、PO→Show；**创建 PO 时必须显式设置 UUID 字段**（`uuid.New().String()`），不能依赖数据库默认值（GORM 会传空字符串覆盖）
6. **错误处理**：业务错误用 `errorx.NewBusinessError`（HTTP 200 + code），系统错误用 `errorx.NewInternalError`
7. **可观测性贯穿**：每个请求自动包含 RequestID、TraceID、Metrics、结构化日志
8. **优先本地验证**：`go test ./...` + `go build ./...` 先通过再推进
9. **注释规范**：复杂逻辑关键决策点注释 WHY，不注释 WHAT；禁止逐行注释
10. **DDL 规范**：PostgreSQL 建表必须有 `COMMENT ON TABLE` + `COMMENT ON COLUMN`
11. **设计规范文档**：详细架构与命名规范参见 `docs/project-design-spec.md`
12. **npm 国内源**：已全局配置 `registry=https://registry.npmmirror.com`
13. **Docker 镜像**：PostgreSQL 必须使用 `pgvector/pgvector:pg16`（含 vector 扩展），不能用原版 `postgres:16`
14. **LLM Key 走环境变量**：`etc/aipivot-api.yaml` 的 `LLM.APIKeyEnv: "ARK_API_KEY"` 决定从哪个环境变量读密钥，
    yaml 中**不写明文 key**。未配置时服务仍能正常启动，只有对话端点降级返回 `code=1002 LLM 网关不可用`
15. **admin 角色需手动提升**：注册接口写死 `member`，管理台 / Flow / Skills 端点需 `role=admin`，
    本地开发用 SQL 提升：`UPDATE users SET role='admin' WHERE email='...';`
