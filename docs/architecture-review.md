# AIPivot AI 智能客服平台 — 后端架构评审与技术规划

基于产品需求文档、`project-design-spec.md` 六层架构规范和当前代码库现状，对 AIPivot 进行全面的需求评审、架构决策分析和技术选型规划。

---

## 一、需求评审：问题与建议

### 1.1 需求优先级问题

产品需求覆盖了 L1→L3 的完整路线，但**缺少对 MVP 边界的精确约束**：

| 问题 | 影响 | 建议 |
|------|------|------|
| MVP 功能列表偏大（Web Chat + RAG + 知识库管理 + 转人工 + 统计） | 1-2 人 1-2 月很难交付 | **MVP 砍到：单租户 Web Chat + RAG 问答 + 基础知识库上传**，转人工和统计放 P1.5 |
| 多渠道接入（微信/飞书/钉钉）在阶段 2 | 渠道适配工程量大 | 先做好 **Channel 抽象接口**，MVP 只实现 Web + API/Webhook |
| Agent 能力缺少具体 Skill 定义 | "查订单/创工单"需要对接具体业务系统 | MVP 阶段用 **Mock Skill + Function Calling 框架** 验证链路 |
| 多租户在 MVP 未提及但架构要求按 L3 设计 | 后补租户隔离代价极高 | **从数据库 schema 设计开始就预留 tenant_id** |

### 1.2 缺失的非功能需求

需求文档缺少以下关键项，需要补充：

- **并发与性能目标**：预期同时在线会话数、QPS 上限
- **数据安全与合规**：知识库数据隔离级别、对话数据保留策略
- **LLM 成本控制**：Token 用量限制、缓存策略、模型降级
- **可用性目标**：SLA 等级（99.9%？）、故障恢复时间

---

## 二、架构决策：单体优先，不用微服务

### 2.1 结论：**模块化单体（Modular Monolith）**

| 考量 | 微服务 | 模块化单体 | 决策 |
|------|--------|-----------|------|
| 开发人力 | 需要多团队 | 1-2 人可驾驭 | ✅ 单体 |
| 运维成本 | K8s/服务发现/链路追踪 | Docker Compose 即可 | ✅ 单体 |
| 部署复杂度 | 多镜像多配置 | 单二进制 | ✅ 单体 |
| 后续拆分 | — | 模块边界清晰可拆 | ✅ 单体先行 |

### 2.2 模块化边界设计

```
aipivot/
├── internal/
│   ├── modules/
│   │   ├── auth/          # 鉴权 & 租户（JWT + RBAC）
│   │   ├── chat/          # 对话管理（会话、消息、上下文）
│   │   ├── knowledge/     # 知识库（文档、切块、向量检索）
│   │   ├── rag/           # RAG 编排（检索 + 生成）
│   │   ├── agent/         # Agent/Skill/Flow 执行引擎
│   │   ├── channel/       # 多渠道适配（Web/API/微信等）
│   │   ├── humanagent/    # 人工客服协同
│   │   └── analytics/     # 数据统计与分析
│   ├── shared/            # 跨模块共享（middleware, errors, pagination）
│   ├── config/
│   ├── svc/
│   └── infra/             # 已有的基础设施层
├── pkg/                   # 可复用的工具包（llm client, vectordb client）
└── deploy/
```

每个 module 内部保持 go-zero 的 `handler → logic → domain → repo/dao` 分层结构，模块间**通过 Go interface 解耦**，不直接引用彼此内部类型。

---

## 三、核心技术选型

### 3.1 后端框架（保持 go-zero）

当前已使用 `go-zero v1.7.6`，**继续使用，不换框架**。理由：
- go-zero 的 API 代码生成、中间件、配置管理已经在用
- 内置熔断/限流/超时，契合 AI 客服场景的韧性需求
- gRPC 支持可为未来拆微服务预留

### 3.2 数据层

| 组件 | 选型 | 理由 |
|------|------|------|
| 关系库 | **PostgreSQL 16**（已有） | 支持 JSONB、全文检索、pgvector 扩展 |
| 向量数据库 | **pgvector** 扩展 (MVP) → **Milvus/Qdrant** (阶段2+) | MVP 用 pgvector 避免引入新组件；规模上来后迁移 |
| 缓存 | **Redis 7**（已有） | 会话上下文缓存、LLM 响应缓存、限流计数 |
| 对象存储 | **MinIO**（本地）/ **阿里云 OSS**（生产） | 知识库文档存储 |
| ORM | **GORM**（已有）| 继续使用，知识库/会话等模型都用 GORM |

### 3.3 AI/LLM 核心框架：Eino（CloudWeGo）

**核心选型：[cloudwego/eino](https://github.com/cloudwego/eino)** — 字节跳动 CloudWeGo 团队出品的 Go 原生 LLM 应用开发框架。

**为什么选 Eino 而不是自建或 LangChainGo：**

| 维度 | 自建 | LangChainGo | **Eino** |
|------|------|-------------|----------|
| 成熟度 | 从零开始 | 社区驱动，更新慢 | 字节生产验证，180+ releases |
| 组件生态 | 全部自写 | 有限 | ChatModel/Tool/Retriever/Embedding 官方实现（OpenAI/Claude/Ollama/通义等） |
| Agent 能力 | 自建 ReAct 循环 | 基础 | ADK：ChatModelAgent + DeepAgent + 多 Agent 协调 |
| 流式处理 | 自建 SSE 管道 | 有限 | 框架自动处理流拼接/分发/合并 |
| 人工介入 | 自建状态机 | 无 | **Interrupt/Resume** 原生支持 human-in-the-loop |
| 可观测性 | 自接 OTel | 无 | Callback Aspects 内置 tracing/metrics/logging 注入点 |
| 工作流编排 | 自建 | 链式 | **Graph Composition** — 类似 LangGraph 的 DAG 编排 |

**Eino 覆盖的产品需求映射：**

| 产品能力 | Eino 对应组件 |
|----------|---------------|
| RAG 知识检索 | `Retriever` + `Embedding` + Graph Composition |
| 多模型调度 | `ChatModel` 抽象 + eino-ext 多模型实现 |
| Agent 执行引擎 | `ChatModelAgent` + `Tool` 接口 + Function Calling |
| 多 Agent 协作 | `DeepAgent` + 子 Agent 委派 |
| 人工兜底/转人工 | `Interrupt/Resume` 机制 |
| 流式对话 | 内置 Stream Processing |
| 可观测追踪 | Callback Aspects（OnStart/OnEnd/OnError） |

### 3.4 AI/LLM 补充组件

| 能力 | 推荐方案 | 说明 |
|------|---------|------|
| LLM 多模型代理 | **One API** | [songquanpeng/one-api](https://github.com/songquanpeng/one-api) — 统一 OpenAI 格式，Eino 的 ChatModel 直接对接 |
| 文档解析 | **Unstructured** (Python sidecar) | PDF/Word/HTML 解析，Go 生态弱项，Docker sidecar 部署 |
| 文本切块 | Eino 内置 + **tiktoken-go** | [pkoukk/tiktoken-go](https://github.com/pkoukk/tiktoken-go) — Token 计数 |
| Embedding | Eino `Embedding` 组件 | 通过 eino-ext 对接通义/OpenAI embedding 模型 |
| 向量检索 | **pgvector** (MVP) | [pgvector/pgvector](https://github.com/pgvector/pgvector) — 实现 Eino `Retriever` 接口封装 |

### 3.5 实时通信

| 能力 | 选型 | 理由 |
|------|------|------|
| Chat 实时推送 | **WebSocket** (go-zero 原生支持) | 对话场景必须双向实时 |
| SSE (Server-Sent Events) | 用于 LLM 流式输出 | 大模型流式返回 token 时用 SSE 推送到前端 |
| 消息队列 | **Redis Streams** (MVP) → **NATS/Kafka** (阶段2+) | MVP 不引入额外 MQ 组件 |

### 3.6 前端技术栈

| 组件 | 选型 | 理由 |
|------|------|------|
| 管理后台 | **React + Vite + TailwindCSS + shadcn/ui** | 现代、组件丰富、适合 B 端后台 |
| Chat Widget | **React 独立组件** → 打包为可嵌入 JS | 客户网站嵌入 `<script>` 即可使用 |
| 状态管理 | **Zustand** | 轻量，适合中型项目 |
| HTTP 客户端 | **ky** 或 **axios** | — |
| 实时通信 | 原生 **WebSocket** + **EventSource (SSE)** | — |

### 3.7 可观测性（已有基础）

当前已有 Prometheus + Grafana + Jaeger + OpenTelemetry，**继续沿用，补充**：

- **LLM 调用追踪**：每次 LLM 请求作为独立 span，记录 model/tokens/latency
- **RAG 链路追踪**：检索 → 重排 → 生成 的全链路 trace
- **业务指标**：AI 解决率、转人工率、知识命中率

---

## 四、关键子系统设计要点

### 4.1 对话管理

```
用户消息 → Channel 适配 → 会话路由 → 意图识别 → RAG/Agent/转人工 → 响应生成 → 推送
```

- **会话状态机**：`created → active → waiting_human → resolved → closed`
- **上下文窗口**：Redis 缓存最近 N 轮对话，超出后摘要压缩
- **Fallback 策略**：RAG 置信度低 → 追问 → 仍低 → 转人工

### 4.2 知识库 & RAG（基于 Eino）

```
文档上传 → 解析(Unstructured) → 切块 → Eino Embedding → pgvector 存储
查询 → Eino Embedding → Eino Retriever(pgvector) → 重排(可选) → Eino Graph → ChatModel 生成
```

- **Retriever 实现**：封装 pgvector 查询为 Eino `Retriever` 接口
- **RAG Graph**：用 Eino `compose.NewGraph` 编排 `retrieve → rerank → prompt_assemble → generate` 流水线
- 切块策略：固定窗口 + 重叠（chunk_size=512, overlap=64 tokens）
- 检索：pgvector cosine similarity，Top-K=5
- 重排：MVP 不做，阶段 2 引入 cross-encoder reranker

### 4.3 Agent 引擎（基于 Eino ADK）

```go
// 使用 Eino 的 Tool 接口，而非自建 Skill 抽象
// 每个业务技能实现 tool.BaseTool 接口
agent, _ := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Model: chatModel,
    ToolsConfig: adk.ToolsConfig{
        ToolsNodeConfig: compose.ToolsNodeConfig{
            Tools: []tool.BaseTool{orderQueryTool, ticketCreateTool},
        },
    },
})
```

- **ChatModelAgent** 内置 ReAct 循环，自动决定何时调用 Tool、何时直接回复
- MVP 阶段实现 2-3 个示例 Tool（查订单、查天气等）
- 阶段 2 用 **DeepAgent** 支持多 Agent 协作 + 复杂任务分解
- 阶段 2 用 **Graph Composition** 构建确定性业务流程，暴露为 Agent Tool

### 4.4 多租户

- **数据隔离**：共享数据库 + `tenant_id` 字段（MVP 足够）
- **配置隔离**：每个租户独立配置 LLM 模型、知识库、Skill 集合
- **鉴权**：JWT + tenant_id claim，go-zero middleware 统一校验

---

## 五、开源框架全景对比

| 领域 | 方案 A | 方案 B | 推荐 |
|------|--------|--------|------|
| LLM 网关 | [One API](https://github.com/songquanpeng/one-api) (Go) | [LiteLLM](https://github.com/BerriAI/litellm) (Python) | **One API** — Go 生态，国内模型支持好 |
| 向量库 | [pgvector](https://github.com/pgvector/pgvector) | [Milvus](https://github.com/milvus-io/milvus) (Go) | **pgvector** 起步，Milvus 备选 |
| AI/RAG/Agent 框架 | [Eino](https://github.com/cloudwego/eino) (CloudWeGo/Go) | [LangChainGo](https://github.com/tmc/langchaingo) | **Eino** — 字节生产验证，Go 原生，组件生态完整，内置 Agent/RAG/Graph/流式/人工介入 |
| 文档解析 | [Unstructured](https://github.com/Unstructured-IO/unstructured) (Python) | [Apache Tika](https://tika.apache.org/) (Java) | **Unstructured** — Docker sidecar 部署 |
| WebSocket | go-zero 内置 | [gorilla/websocket](https://github.com/gorilla/websocket) | **gorilla/websocket** — 更灵活 |
| 任务队列 | Redis Streams | [Asynq](https://github.com/hibiken/asynq) (Go + Redis) | **Asynq** — 知识库文档处理异步任务 |
| 认证 | 自建 JWT | [Casdoor](https://github.com/casdoor/casdoor) (Go) | **自建 JWT** MVP 够用，规模化后考虑 Casdoor |
| 前端 UI | shadcn/ui + TailwindCSS | Ant Design | **shadcn/ui** — 更现代，定制性强 |

---

## 六、不需要的技术（避免过度工程）

- **Kubernetes** — Docker Compose 足够 MVP
- **独立消息队列（Kafka/RabbitMQ）** — Redis Streams / Asynq 覆盖
- **微服务拆分** — 模块化单体 + 清晰边界即可
- **GraphQL** — REST API 足以覆盖当前场景
- **独立 AI 编排框架（Dify/FastGPT）** — 你做的就是这类产品，核心链路必须自建
- **Elasticsearch** — pgvector + PostgreSQL 全文检索覆盖 MVP

---

## 七、建议实施路线

### Phase 0 — 基础补全（当前 → 1 周） ✅ 已完成
- [x] 数据库 migration 框架（golang-migrate）
- [x] 多租户基础表设计（tenants, users, api_keys）
- [x] JWT 鉴权中间件
- [x] 项目目录重构为模块化结构

### Phase 1 — MVP 核心（2-4 周） 🚧 进行中
- [x] 用户注册接口（auth.api + RegisterLogic 实现）
- [x] 知识库数据库表设计（migration 002: knowledge_bases / documents / document_chunks + pgvector HNSW 索引）
- [x] 对话数据库表设计（migration 003: conversations / messages）
- [x] 知识库模块完整 CRUD（.api 定义 + handler/logic/domain/repo/dao 全层实现）
- [x] 文档上传接口（multipart 文件上传，status=pending 异步处理桩）
- [x] 对话模块完整 CRUD + 发消息（.api 定义 + 全层实现，MVP stub AI 回复）
- [x] JWT 鉴权中间件集成（knowledge/chat 路由组通过 AuthMiddleware 保护）
- [x] GORM Gen 代码生成覆盖所有新表（KnowledgeBase/Document/DocumentChunk/Conversation/Message）
- [x] ServiceContext 完成全模块 DI 组装（Auth + Knowledge + Chat Repo 注入）
- [ ] 知识库模块：文档解析 → 切块 → Embedding → pgvector 存储（异步 pipeline）
- [ ] RAG 模块：Eino Retriever(pgvector) → Rerank → ChatModel 生成
- [x] Chat 模块：SSE 流式输出（POST /conversations/:convId/messages/stream）
- [x] 前端 Chat Widget 原型（React + Vite + TailwindCSS + Zustand + SSE 流式）
- [ ] LLM Gateway 集成（One API 或自建适配层）
- [ ] 管理后台：知识库 CRUD UI

### Phase 2 — 增强（4-8 周）
- [ ] Agent/Skill 框架 + Function Calling
- [ ] 人工客服转接
- [ ] 多渠道接入（API/Webhook 优先）
- [ ] 对话分析仪表盘
- [ ] LLM 成本追踪与限流

### Phase 3 — 平台化（8+ 周）
- [ ] 可视化 Flow 编辑器
- [ ] 多 Agent 协作
- [ ] 客户自助 Skill 注册
- [ ] 从 pgvector 迁移到 Milvus（如有需要）

---

## 八、与现有架构规范（project-design-spec.md）的结合分析

现有 `project-design-spec.md` 定义了成熟的六层架构（API → Handler → Logic → Domain → Repo/DAO → Infra），**直接复用**，但 AI 客服场景需要扩展：

| 规范能力 | 现状 | AI 客服扩展点 |
|----------|------|---------------|
| Handler/Logic/Repo 三层 | 完备 | 知识库/会话/租户等 CRUD 模块直接套用 |
| Domain Model + Assembler | 完备 | 新增 AI 领域模型（Conversation, Message, KnowledgeChunk, Skill） |
| GORM Gen 代码生成 | 完备 | pgvector 字段需自定义 GORM 类型（`vector(1536)`） |
| ServiceContext 依赖注入 | 完备 | 新增 Eino Agent/Retriever/ChatModel 实例注入 |
| 错误处理 (errorx) | 完备 | 新增 AI 相关错误码（LLM 超时、知识未命中、Token 超限） |
| Middleware 链 | 完备 | 新增 WebSocket 升级中间件、SSE 响应中间件 |
| 健康检查 | /healthz + /readyz | 新增 LLM Gateway 连通性检查、向量库检查 |
| goctl 代码生成 | .api 驱动 | REST API 继续用 goctl；**WebSocket/SSE 端点需手写 Handler** |

**需要突破现有规范的地方：**

1. **WebSocket/SSE 不走标准 Handler 模板** — go-zero 的 goctl 不生成 WebSocket handler，需手写并注册
2. **Eino 组件不走 Repo/DAO 层** — Eino 的 ChatModel/Retriever/Agent 是独立的 AI 组件层，注入 ServiceContext 但不经过 Repo 抽象
3. **异步任务（文档处理）** — Asynq 任务处理器不走 Handler→Logic 链路，需要独立的 Worker 入口
4. **流式响应** — LLM 生成是流式的，不适合 `httpx.OkJsonCtx` 标准响应模式

**建议的分层扩展：**
```
原有六层（CRUD 模块沿用）
├── API → Handler → Logic → Domain → Repo/DAO → PO

新增 AI 层（与 Logic 层平级协作）
├── Logic 调用 → Eino Agent/RAG Graph → Eino Components (ChatModel/Retriever/Tool)
│                                       → 通过 Eino Callback 接入 OTel tracing

新增异步层
├── Asynq Worker → 文档处理 Pipeline → Embedding → pgvector 写入
```

---

## 九、风险提示

1. **文档解析是 Go 弱项** — PDF/Word 解析必须依赖 Python sidecar（Unstructured），增加部署复杂度
2. **LLM 延迟不可控** — 必须做好超时、降级、缓存三件套；Eino 的 Callback Aspects 可统一记录
3. **知识库质量决定产品体验** — 切块策略和检索质量需要大量实验调优
4. **前端工作量不小** — 管理后台 + Chat Widget 至少占总工作量 40%
5. **One API 是第三方项目** — 如依赖它做 LLM 路由，需评估其稳定性或做好 fallback
6. **Eino 与 go-zero 集成** — 两者无直接冲突，但 Eino 的流式/回调模式与 go-zero 的同步 Handler 模式需要适配桥接

---

---

## 十、开发日志

### 2026-05-10 Phase 1 后端骨架搭建

**完成内容：**

1. **Auth 模块补全** — 新增 `/api/v1/auth/register` 注册接口，含邮箱唯一性校验、密码加密、default 租户绑定
2. **数据库 Migration** — 新增 `000002_knowledge_base`（pgvector extension + knowledge_bases / documents / document_chunks 含 HNSW 向量索引）和 `000003_conversations`（conversations / messages），完整 COMMENT ON 注释
3. **Knowledge 模块** — API 定义（8 个端点 CRUD + 文档上传/列表/删除）、handler/logic/domain(model+assembler)/repo/dao 全层实现
4. **Chat 模块** — API 定义（6 个端点：会话 CRUD + 发消息 + 消息历史）、全层实现；SendMessage 使用 stub AI 回复，为 RAG/LLM 集成预留接口
5. **基础设施** — AuthMiddleware 集成到 goctl 生成的 routes.go、ServiceContext 完成全模块 DI 组装、GORM Gen 覆盖 8 张表

**当前 API 端点清单：**

| 模块 | 方法 | 路径 | 鉴权 |
|------|------|------|------|
| Auth | POST | /api/v1/auth/login | ❌ |
| Auth | POST | /api/v1/auth/register | ❌ |
| Knowledge | POST | /api/v1/knowledge-bases | ✅ |
| Knowledge | GET | /api/v1/knowledge-bases | ✅ |
| Knowledge | GET | /api/v1/knowledge-bases/:id | ✅ |
| Knowledge | PUT | /api/v1/knowledge-bases/:id | ✅ |
| Knowledge | DELETE | /api/v1/knowledge-bases/:id | ✅ |
| Knowledge | POST | /api/v1/knowledge-bases/:kbId/documents | ✅ |
| Knowledge | GET | /api/v1/knowledge-bases/:kbId/documents | ✅ |
| Knowledge | DELETE | /api/v1/knowledge-bases/:kbId/documents/:id | ✅ |
| Chat | POST | /api/v1/conversations | ✅ |
| Chat | GET | /api/v1/conversations | ✅ |
| Chat | GET | /api/v1/conversations/:id | ✅ |
| Chat | PUT | /api/v1/conversations/:id/close | ✅ |
| Chat | POST | /api/v1/conversations/:convId/messages | ✅ |
| Chat | GET | /api/v1/conversations/:convId/messages | ✅ |
| Infra | GET | /healthz, /readyz, /metrics, /v1/ping | ❌ |

**下一步优先级：**
1. ~~P0: 文档异步处理 pipeline（Asynq + 切块 + Embedding + pgvector 写入）~~ ✅
2. ~~P0: Eino ChatModel + Retriever 集成，替换 SendMessage 中的 stub 回复~~ ✅
3. P1: SSE 流式输出
4. P1: 前端 Chat Widget 原型

### 2026-05-10 Phase 1 P0 — 文档 Pipeline + RAG 集成

**完成内容：**

1. **LLM 配置** — `config.LLMConf`（BaseURL/APIKey/ChatModel/EmbeddingModel/MaxTokens/Temperature）+ `WorkerConf`，YAML 配置完善
2. **pkg/llm** — OpenAI-compatible HTTP 客户端（兼容 One API/OpenAI/Azure），支持 `ChatCompletion` 和 `Embed/EmbedSingle`
3. **pkg/chunker** — 固定窗口+重叠文本切块器，支持段落边界优先切分、超长段落硬切分、句子边界回退
4. **DocumentChunk DAO/Repo** — pgvector 原生 SQL 批量写入（`BatchCreateWithEmbedding`）+ cosine 相似度搜索（`SimilaritySearch`），GORM Gen 查询用于计数/删除
5. **Asynq 异步 Worker** — `internal/worker/` 包含任务定义（`DocumentProcessPayload`）、处理器（`DocumentProcessor`）、Server 启动/关闭
6. **文档处理 Pipeline** — 上传文档 → Asynq 入队 → Worker 消费 → 读取内容 → 切块 → 批量 Embedding → pgvector 写入 → 更新文档状态和知识库计数
7. **RAG 模块** — `internal/modules/rag/Service`：retrieve（query embedding → pgvector Top-K）→ buildPrompt（system + context + history + question）→ LLM ChatCompletion
8. **SendMessage 集成** — 替换 stub 回复，现在调用 RAG.Answer，支持知识库关联会话的检索增强生成；无知识库时降级为纯 LLM 对话
9. **ServiceContext 扩展** — 新增 `LLMClient`/`RAGService`/`AsynqClient`/`DocumentChunkRepo`，完成 DI 组装和优雅关闭

**新增文件清单：**

| 路径 | 用途 |
|------|------|
| `pkg/llm/client.go` | OpenAI-compatible LLM/Embedding 客户端 |
| `pkg/chunker/chunker.go` | 文本切块器（段落+句子边界感知） |
| `internal/modules/knowledge/repo/dao/document_chunk_dao.go` | pgvector 批量写入 + 相似度搜索 |
| `internal/modules/knowledge/repo/document_chunk_repo.go` | DocumentChunk 仓储实现 |
| `internal/modules/rag/service.go` | RAG 编排服务 |
| `internal/worker/tasks.go` | Asynq 任务类型定义 |
| `internal/worker/document_processor.go` | 文档处理异步任务处理器 |
| `internal/worker/server.go` | Asynq Worker 启动/关闭 |

**下一步优先级：**
1. ~~P1: SSE 流式输出~~ ✅
2. ~~P1: 前端 Chat Widget 原型~~ ✅
3. P1: 管理后台知识库 CRUD UI
4. P2: LLM Gateway 多模型路由

### 2026-05-10 Phase 1 P1 — SSE 流式输出

**完成内容：**

1. **LLM 流式客户端** — `pkg/llm/stream.go` 新增 `ChatCompletionStream` 方法，解析 OpenAI SSE 协议（`data: {...}` + `[DONE]`），通过 channel 逐步推送增量 token
2. **RAG 流式编排** — `internal/modules/rag/stream.go` 新增 `AnswerStream` 方法：retrieve → prompt → streaming LLM generation，返回 `<-chan StreamEvent` + `StreamMeta`（来源引用）
3. **SSE Writer** — `internal/shared/sse/writer.go` 封装 SSE 事件写入，支持 `WriteEvent`/`WriteDone`/`WriteError`，自动设置 `text/event-stream` 头和 flush
4. **流式 Logic** — `internal/logic/chat/sendMessageStreamLogic.go`：校验会话 → 保存用户消息 → RAG 流式生成 → SSE 推送 delta → 保存 AI 回复
5. **手写 Handler + 路由** — `SendMessageStreamHandler` 绕过 goctl 标准 JSON 响应，直接操作 `http.ResponseWriter` 写 SSE；路由注册 `POST /api/v1/conversations/:convId/messages/stream`
6. **重构** — 提取 `buildChatHistory` 到 `internal/logic/chat/util.go`，供同步/流式 Logic 共享

**SSE 事件协议：**

```
event: message_start
data: {"messageId":"uuid","conversationId":123}

event: delta
data: {"content":"增量token"}

event: message_end
data: {"messageId":"uuid","model":"gpt-3.5-turbo","tokenCount":150,"latencyMs":2300,"sources":[...]}

event: done
data: [DONE]

event: error  (异常时)
data: {"code":1002,"msg":"AI 回复生成失败"}
```

**新增/修改文件清单：**

| 路径 | 用途 |
|------|------|
| `pkg/llm/stream.go` | **新文件** — `ChatCompletionStream` + 流式类型 |
| `internal/modules/rag/stream.go` | **新文件** — `AnswerStream` + `StreamMeta` |
| `internal/shared/sse/writer.go` | **新文件** — SSE 事件写入器 + 标准事件结构体 |
| `internal/logic/chat/util.go` | **新文件** — 提取 `buildChatHistory` 共享函数 |
| `internal/logic/chat/sendMessageStreamLogic.go` | **新文件** — 流式发送消息业务逻辑 |
| `internal/handler/chat/sendMessageStreamHandler.go` | **新文件** — SSE Handler（手写，非 goctl） |
| `internal/logic/chat/sendMessageLogic.go` | 删除 `buildChatHistory`（移至 util.go） |
| `internal/handler/routes.go` | 新增 `/conversations/:convId/messages/stream` 路由 |

### 2026-05-10 Phase 1 P1 — 前端 Chat Widget 原型

**完成内容：**

1. **项目初始化** — `web/` 目录，Vite 6 + React 18 + TypeScript + TailwindCSS 3，Vite proxy 转发 `/api` 到后端 8888 端口
2. **API 客户端** — `web/src/lib/api.ts`：全量 TypeScript 类型定义（API 响应、SSE 事件）、HTTP 请求封装、**POST SSE 流式解析**（fetch + ReadableStream 逐行解析 SSE 事件）
3. **状态管理** — Zustand stores：`auth.ts`（JWT + 用户信息 + localStorage 持久化）、`chat.ts`（会话列表、消息列表、流式消息状态机）
4. **登录页** — `LoginPage.tsx`：登录/注册双 Tab 切换，深色渐变背景 + 毛玻璃卡片，表单校验 + 错误提示
5. **聊天页** — `ChatPage.tsx` 完整实现：
   - **侧边栏**：会话列表 + 新建会话 + 用户信息 + 退出登录
   - **消息面板**：用户/AI 消息气泡（区分样式）、流式打字光标动画、RAG 来源引用标签、消息元信息（模型/耗时）
   - **输入区**：自动高度 textarea、Enter 发送 / Shift+Enter 换行、流式中禁用
   - **新建会话弹窗**：可选关联知识库、自定义标题
6. **SSE 集成** — 发送消息走 `POST /conversations/:convId/messages/stream`，实时渲染增量 token，流结束后提交完整消息到列表

**技术栈：**

| 组件 | 选型 |
|------|------|
| 构建工具 | Vite 6 |
| 框架 | React 18 + TypeScript 5 |
| 样式 | TailwindCSS 3 |
| 状态管理 | Zustand 5（persist middleware） |
| 图标 | Lucide React |
| API 代理 | Vite dev server proxy |

**新增文件清单：**

| 路径 | 用途 |
|------|------|
| `web/package.json` | 依赖管理 |
| `web/vite.config.ts` | Vite 配置 + API proxy |
| `web/tsconfig.json` | TypeScript 配置 |
| `web/tailwind.config.js` | TailwindCSS 配置 |
| `web/postcss.config.js` | PostCSS 配置 |
| `web/index.html` | 入口 HTML |
| `web/src/main.tsx` | React 入口 |
| `web/src/App.tsx` | 根组件（登录/聊天路由） |
| `web/src/index.css` | TailwindCSS + 自定义样式（滚动条、打字光标） |
| `web/src/lib/api.ts` | API 类型 + HTTP 客户端 + SSE 流解析 |
| `web/src/store/auth.ts` | 认证 Zustand store |
| `web/src/store/chat.ts` | 聊天 Zustand store |
| `web/src/pages/LoginPage.tsx` | 登录/注册页 |
| `web/src/pages/ChatPage.tsx` | 聊天主页（侧边栏 + 消息面板 + 输入区 + 新建弹窗） |

**下一步优先级：**
1. P1: 管理后台知识库 CRUD UI
2. P2: LLM Gateway 多模型路由
3. P2: Chat Widget 嵌入式 JS 打包

*文档版本：v1.4 | 更新日期：2026-05-10*
