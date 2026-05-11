# AIPivot 框架改进方案

> 基于 80/100 评分结果，按优先级排列的改进计划。
> 架构定位：**DDD-Lite**（战略 DDD 划模块边界 + Repository 接口解耦 + logic 层即 Service 层）。

---

## P0 — 模块包结构重构（+3 分，难度 M）

当前 CRUD 模块（knowledge、auth、chat）每个拆了 4 个子包（domain/model、domain/assembler、repo、repo/dao），
但业务复杂度撑不起这套结构——领域模型只有 35 行，Repo 全是一行透传 DAO。
合并为 **1 个扁平包**，与已经写得很好的 agent/rag/webhook 风格对齐。

### 0.1 架构原则

- `logic/` 就是 Service 层（go-zero 设计），业务编排逻辑写在这里，不再额外建 Service
- `modules/` 是 logic 的支撑层，提供：Repository 接口+实现、校验函数、DTO 转换函数
- 复杂子系统（agent、rag、webhook）保持自封装风格，logic 委托调用

### 0.2 目标结构

**Before — 每模块 4 个包、6 个目录：**

```
modules/knowledge/
├── domain/
│   ├── assembler/              ← package assembler
│   │   ├── knowledge_base.go
│   │   └── document.go
│   └── model/                  ← package model
│       └── knowledge_base.go
└── repo/
    ├── dao/                    ← package dao
    │   ├── knowledge_base_dao.go
    │   ├── document_dao.go
    │   └── document_chunk_dao.go
    ├── interface.go            ← package repo
    ├── knowledge_base_repo.go
    ├── document_repo.go
    └── document_chunk_repo.go
```

**After — 1 个包、1 个目录：**

```
modules/knowledge/               ← package knowledge
├── knowledge.go                 ← 校验函数 + Repository 接口
├── assembler.go                 ← DTO 转换函数（Request→PO、PO→Show）
└── store.go                     ← Repository GORM 实现（合并原 repo + dao）
```

### 0.3 三个模块的具体改造清单

#### knowledge 模块

| 原文件 | 操作 | 目标 |
|--------|------|------|
| `domain/model/knowledge_base.go` | 校验逻辑迁入 `knowledge.go` 改为独立函数 | `ValidateName()`, `ResolveDimension()` |
| `domain/assembler/knowledge_base.go` | 迁入 `assembler.go` | `NewKnowledgeBasePo()`, `ToShowKB()`, `ToShowKBList()` |
| `domain/assembler/document.go` | 迁入 `assembler.go` | `ToShowDocument()`, `ToShowDocumentList()` |
| `repo/interface.go` | 接口定义迁入 `knowledge.go` | `KBRepository`, `DocumentRepository`, `DocChunkRepository` |
| `repo/knowledge_base_repo.go` + `repo/dao/knowledge_base_dao.go` | 合并为 `store.go` 中的一个 struct | `KBStore`（去掉透传层，直接操作 GORM Gen） |
| `repo/document_repo.go` + `repo/dao/document_dao.go` | 合并入 `store.go` | `DocumentStore` |
| `repo/document_chunk_repo.go` + `repo/dao/document_chunk_dao.go` | 合并入 `store.go` | `DocChunkStore` |
| 原 `domain/` 和 `repo/` 目录 | 删除 | — |

**knowledge.go 示例：**

```go
package knowledge

import "errors"

// ========== 校验 ==========

var EmbeddingModels = map[string]int{
    "text-embedding-3-small": 1536,
    "text-embedding-3-large": 3072,
    "text-embedding-ada-002": 1536,
}

func ValidateName(name string) error {
    if name == "" {
        return errors.New("知识库名称不能为空")
    }
    if len(name) > 255 {
        return errors.New("知识库名称不能超过 255 个字符")
    }
    return nil
}

func ResolveDimension(model string) int {
    if dim, ok := EmbeddingModels[model]; ok {
        return dim
    }
    return 1536
}

// ========== Repository 接口 ==========

type KBRepository interface {
    Create(ctx context.Context, kb *po.KnowledgeBase) error
    GetByID(ctx context.Context, id int64) (*po.KnowledgeBase, error)
    GetList(ctx context.Context, tenantID int64, page, pageSize int, name string) ([]*po.KnowledgeBase, int64, error)
    Update(ctx context.Context, id int64, updates map[string]any) error
    Delete(ctx context.Context, id int64) error
}

// DocumentRepository, DocChunkRepository 同理...
```

**store.go 示例（合并 repo + dao，去掉透传层）：**

```go
package knowledge

import (
    "context"
    "errors"

    "aipivot/internal/shared/po"
    "aipivot/internal/shared/query"

    "gorm.io/gorm"
)

type KBStore struct {
    q *query.Query
}

func NewKBStore(q *query.Query) *KBStore {
    return &KBStore{q: q}
}

func (s *KBStore) Create(ctx context.Context, kb *po.KnowledgeBase) error {
    return s.q.KnowledgeBase.WithContext(ctx).Create(kb)
}

func (s *KBStore) GetByID(ctx context.Context, id int64) (*po.KnowledgeBase, error) {
    kb := s.q.KnowledgeBase
    result, err := kb.WithContext(ctx).Where(kb.ID.Eq(id)).First()
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return result, err
}

// 其余方法同理，直接操作 GORM Gen，不再经过 DAO 中间层
// 错误不在 store 层打日志（去掉 logx.Errorf），让错误自然冒泡到 logic 层统一处理
```

**assembler.go 示例：**

```go
package knowledge

import (
    "aipivot/internal/shared/po"
    "aipivot/internal/types"

    "github.com/google/uuid"
)

func NewKnowledgeBasePo(req *types.CreateKnowledgeBaseRequest, tenantID int64) *po.KnowledgeBase {
    return &po.KnowledgeBase{
        UUID:        uuid.New().String(),
        TenantID:    tenantID,
        Name:        req.Name,
        Description: req.Description,
        Model:       req.EmbeddingModel,
        Dimension:   ResolveDimension(req.EmbeddingModel),
        Status:      "active",
    }
}

func ToShowKB(kb *po.KnowledgeBase) types.ShowKnowledgeBase {
    return types.ShowKnowledgeBase{
        ID: kb.ID, UUID: kb.UUID, Name: kb.Name,
        Description: kb.Description, Model: kb.Model,
        Dimension: kb.Dimension, Status: kb.Status,
        DocCount: kb.DocCount,
        CreatedAt: kb.CreatedAt.Unix(), UpdatedAt: kb.UpdatedAt.Unix(),
    }
}

func ToShowKBList(list []*po.KnowledgeBase) []types.ShowKnowledgeBase {
    result := make([]types.ShowKnowledgeBase, 0, len(list))
    for _, kb := range list {
        result = append(result, ToShowKB(kb))
    }
    return result
}

// Document 转换函数同理...
```

#### auth 模块

| 原文件 | 操作 | 目标 |
|--------|------|------|
| `domain/model/user.go` | 校验迁入 `auth.go`，EncryptPassword/CheckPasswordMatch 保留为独立函数 | `ValidateEmail()`, `ValidatePassword()`, `EncryptPassword()`, `CheckPasswordMatch()` |
| `domain/assembler/user.go` | 迁入 `assembler.go` | `NewUserPo()`, `ToShowUser()`, `ToLoginData()` |
| `repo/interface.go` | 接口迁入 `auth.go` | `UserRepository`, `TenantRepository`, `ApiKeyRepository` |
| `repo/*.go` + `repo/dao/*.go` | 合并为 `store.go` | `UserStore`, `TenantStore`, `ApiKeyStore` |
| `jwt.go` | 保持不变 | — |
| `middleware.go` | 保持不变 | — |

#### chat 模块

| 原文件 | 操作 | 目标 |
|--------|------|------|
| `domain/assembler/conversation.go` | 迁入 `assembler.go` | `ToShowConversation()`, `ToShowConversationList()` |
| `domain/assembler/message.go` | 迁入 `assembler.go` | `ToShowMessage()`, `ToShowMessageList()` |
| `repo/interface.go` | 接口迁入 `chat.go` | `ConversationRepository`, `MessageRepository` |
| `repo/*.go` + `repo/dao/*.go` | 合并为 `store.go` | `ConversationStore`, `MessageStore` |

### 0.4 Logic 层联动修改

Logic 文件的业务逻辑不变，只改 import 路径：

```go
// Before — 3 个 import path
import (
    "aipivot/internal/modules/knowledge/domain/assembler"
    "aipivot/internal/modules/knowledge/domain/model"
    kbRepo "aipivot/internal/modules/knowledge/repo"
)

// After — 1 个 import path
import (
    "aipivot/internal/modules/knowledge"
)

// 调用方式变化
// Before: assembler.KnowledgeBasePoToShow(kbPo)
// After:  knowledge.ToShowKB(kbPo)

// Before: model.NewKnowledgeBase(...) → kb.CheckName()
// After:  knowledge.ValidateName(req.Name)
```

### 0.5 ServiceContext 联动修改

```go
// Before: 构造链 Query → DAO → Repo（3 步）
knowledgeBaseDao := kbDao.NewKnowledgeBaseDao(q)
KnowledgeBaseRepo: kbRepo.NewKnowledgeBaseRepo(knowledgeBaseDao),

// After: 构造链 Query → Store（1 步）
KnowledgeBaseRepo: knowledge.NewKBStore(q),

// 字段类型变化
// Before: KnowledgeBaseRepo kbRepo.KnowledgeBaseRepository
// After:  KnowledgeBaseRepo knowledge.KBRepository
```

### 0.6 不动的模块

以下模块已经是扁平结构，**不做任何改动**：

- `modules/agent/` — 1 个包，Agent + Tool + Registry，已是最佳实践
- `modules/rag/` — 1 个包，Service + Stream，logic 委托调用
- `modules/channel/webhook/` — 1 个包，DeliveryService + Repository

---

## P1 — 中间件位置统一（+1 分，难度 S）

**问题：** 中间件实现分散在两处：

```
internal/middleware/authMiddleware.go       ← goctl 骨架，委托给 modules/auth
internal/middleware/apiKeyMiddleware.go     ← 完整实现在此（应该也是薄代理）
internal/modules/auth/middleware.go         ← JWT 实际逻辑在此
```

**方案：** 将 `apiKeyMiddleware.go` 的认证逻辑迁入 `internal/modules/auth/`，与 JWT 中间件对齐。

**调整后结构：**

```
internal/middleware/
  ├── authMiddleware.go         ← goctl 骨架，委托 → auth.JWTMiddleware
  └── apiKeyMiddleware.go       ← goctl 骨架，委托 → auth.APIKeyMiddleware

internal/modules/auth/
  ├── auth.go                   ← 校验 + Repository 接口（P0 已合并）
  ├── assembler.go              ← DTO 转换（P0 已合并）
  ├── store.go                  ← GORM 实现（P0 已合并）
  ├── jwt.go                    ← token 签发/解析
  ├── jwt_middleware.go         ← JWT 认证逻辑（从 middleware.go 重命名）
  ├── apikey_middleware.go      ← API Key 认证逻辑（从 middleware/ 迁入）
  └── context.go                ← ClaimsFromContext / TenantIDFromContext / APIKeyIDFromContext
```

**具体操作：**

1. 将 `internal/middleware/apiKeyMiddleware.go` 中的核心逻辑迁移到 `internal/modules/auth/apikey_middleware.go`
2. `internal/middleware/apiKeyMiddleware.go` 改为薄代理：

```go
package middleware

import (
    "net/http"
    "aipivot/internal/modules/auth"
)

type ApiKeyMiddleware struct {
    apiKeyRepo auth.ApiKeyRepository
}

func NewApiKeyMiddleware(apiKeyRepo auth.ApiKeyRepository) *ApiKeyMiddleware {
    return &ApiKeyMiddleware{apiKeyRepo: apiKeyRepo}
}

func (m *ApiKeyMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
    return auth.APIKeyMiddleware(m.apiKeyRepo)(next)
}
```

3. `internal/modules/auth/middleware.go` 重命名为 `jwt_middleware.go`
4. context 提取函数统一收敛到 `internal/modules/auth/context.go`

---

## P2 — 错误处理增强（+1 分，难度 S）

### 2.1 引入 error wrapping

**问题：** `BusinessError` 是扁平结构，无法追踪原始错误来源。且 store 层和 logic 层对同一个错误打两遍日志。

**方案：**

1. `BusinessError` 增加 `Cause` 字段 + `Unwrap()` 方法
2. store 层不打日志（P0 已去掉），错误冒泡到 logic 层
3. logic 层用 Wrap 构造器包装错误，全局 ErrorHandler 统一记录 Cause

```go
// internal/shared/errorx/errorx.go — 增强

type BusinessError struct {
    Code  int    `json:"code"`
    Msg   string `json:"msg"`
    Cause error  `json:"-"` // 不序列化给前端，仅用于日志和错误链追踪
}

func (e *BusinessError) Unwrap() error {
    return e.Cause
}

// Wrap 构造器：保留原始错误，对外暴露友好信息
func NewInternalErrorWrap(msg string, cause error) *BusinessError {
    return &BusinessError{Code: CodeFailed, Msg: msg, Cause: cause}
}

func NewNotFoundErrorWrap(msg string, cause error) *BusinessError {
    return &BusinessError{Code: CodeNotFound, Msg: msg, Cause: cause}
}
```

**Logic 层使用方式变化：**

```go
// Before — 手动打日志 + 丢弃原始错误
if err != nil {
    l.Logger.Errorf("GetKnowledgeBase err: %v", err)
    return nil, errorx.NewInternalError("查询知识库失败")
}

// After — 原始错误保留在链路中，全局 handler 自动记录
if err != nil {
    return nil, errorx.NewInternalErrorWrap("查询知识库失败", err)
}
```

**全局 ErrorHandler 同步调整：**

```go
// internal/shared/errorx/handler.go
var bizErr *BusinessError
if errors.As(err, &bizErr) {
    if bizErr.Cause != nil {
        logx.WithContext(ctx).Errorf("biz error [%d] %s: %v", bizErr.Code, bizErr.Msg, bizErr.Cause)
    }
    return http.StatusOK, bizErr
}
```

---

## P3 — 基础设施补全（+2 分）

### 3.1 Dockerfile 多阶段构建（难度 S）

```dockerfile
# ---- Build Stage ----
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/aipivot .

# ---- Runtime Stage ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/aipivot .
COPY etc/ etc/
COPY migrations/ migrations/
EXPOSE 8888
ENTRYPOINT ["./aipivot"]
```

### 3.2 CORS 中间件（难度 S）

**新增 `internal/middleware/cors.go`：**

```go
package middleware

import "net/http"

func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,X-API-Key")
        w.Header().Set("Access-Control-Max-Age", "86400")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next(w, r)
    }
}
```

**注册：** 在 `aipivot.go` 中 `server.Use(middleware.CORSMiddleware)` 添加为全局中间件。

### 3.3 Rate Limiting（难度 M）

基于 go-zero 内置的 `limit.NewTokenLimiter`，对 Open API 端点限流：

```go
package middleware

import (
    "net/http"
    "aipivot/internal/shared/errorx"
    "github.com/zeromicro/go-zero/core/limit"
    "github.com/zeromicro/go-zero/core/stores/redis"
    "github.com/zeromicro/go-zero/rest/httpx"
)

func NewRateLimitMiddleware(store *redis.Redis, rate, burst int) func(http.HandlerFunc) http.HandlerFunc {
    limiter := limit.NewTokenLimiter(rate, burst, store, "rate:open-api")
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests,
                    errorx.NewBusinessError(429, "rate limit exceeded"))
                return
            }
            next(w, r)
        }
    }
}
```

---

## P4 — 测试覆盖（+5 分，难度 M）

当前测试覆盖是最大短板。P0 重构后包结构更简单，补测成本更低。

### 4.1 测试基础设施（难度 S）

```
internal/testutil/
  ├── mock_service_context.go  ← 构建注入 mock Repo 的 ServiceContext
  └── fixtures.go              ← 通用测试 PO 工厂函数
```

**Makefile 扩展：**

```makefile
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
```

### 4.2 Store 层测试（难度 M）

使用 `github.com/DATA-DOG/go-sqlmock` 做 SQL 级 mock 测试：

```
internal/modules/auth/store_test.go           ← UserStore + ApiKeyStore
internal/modules/knowledge/store_test.go      ← KBStore + DocumentStore + DocChunkStore
internal/modules/chat/store_test.go           ← ConversationStore + MessageStore
```

### 4.3 Logic 层测试（难度 M）

为每个 Repository 接口生成 mock（`github.com/vektra/mockery`），注入 ServiceContext 测试业务逻辑：

```
internal/logic/auth/registerLogic_test.go          ← 注册校验 + 默认租户绑定
internal/logic/auth/loginLogic_test.go             ← 密码校验 + token 生成
internal/logic/knowledge/createKnowledgeBaseLogic_test.go
internal/logic/chat/sendMessageLogic_test.go       ← RAG 编排 + 消息持久化
```

---

## P5 — 业务指标增强（+1 分，难度 S）

在 `internal/observability/metrics.go` 中新增 AI 专属 Prometheus 指标：

```go
// 新增字段
llmCallTotal    *prometheus.CounterVec   // LLM 调用次数（model, status）
llmCallDuration *prometheus.HistogramVec // LLM 调用延迟（model）
llmTokensUsed   *prometheus.CounterVec   // token 消耗（model, type=prompt|completion）
ragRetrieveHits *prometheus.HistogramVec // RAG 检索命中数量（kb_id）
```

**埋点位置：**
- `rag/service.go` — retrieve 结束后记录命中数
- `pkg/llm/client.go` — ChatCompletion 返回后记录延迟和 token

---

## P6 — API 路径参数统一 + 设计文档同步（+1 分，难度 S）

### 6.1 路径参数命名统一

单资源路由统一为 `:id`，父子嵌套路由中父资源保留具名参数：

| 现状 | 调整后 |
|------|-------|
| `/conversations/:convId/messages` | `/conversations/:id/messages` |
| `/knowledge-bases/:kbId/documents` | `/knowledge-bases/:id/documents` |
| `/knowledge-bases/:kbId/documents/:id` | 保持不变 |
| `/webhook/:webhookId/inbound` | `/webhook/:id/inbound` |

修改 `api/*.api` → `make api` 重新生成。

### 6.2 更新 project-design-spec.md

1. 将 `hjm-admin` / `hjm-internal` 引用替换为 `aipivot` 单体架构
2. 删除双仓库设计部分，改为 `internal/shared/` 模式
3. 更新架构层次说明：logic = Service 层，modules = 支撑层（校验 + 转换 + 数据访问）
4. 更新 modules 包结构规范（1 个包 = knowledge.go + assembler.go + store.go）
5. 补充 DDD-Lite 定位说明

---

## 实施顺序总结

| 阶段 | 改进项 | 预计工作量 | 预期分值提升 |
|------|--------|-----------|-------------|
| **Phase 1** | **P0 模块包结构重构**（knowledge → auth → chat） | 1 天 | +3 |
| **Phase 2** | P1 中间件统一 + P2 错误处理增强 | 0.5 天 | +2 |
| **Phase 3** | P3.1 Dockerfile + P3.2 CORS | 0.5 天 | +1 |
| **Phase 4** | P4 测试基础设施 + Store/Logic 测试 | 1.5 天 | +5 |
| **Phase 5** | P5 业务指标 + P6 路径参数 + 文档同步 | 0.5 天 | +2 |
| **Phase 6** | P3.3 Rate Limit | 0.5 天 | +1 |

**预期总分：80 → 94 分（生产级水准）**

---

## 附录：什么时候升级到完整 DDD？

当前项目定位 **DDD-Lite**（战略 DDD + Repository 接口），适合当前业务复杂度。
出现以下信号时再引入战术 DDD 模式：

| 信号 | 引入什么 |
|------|---------|
| 跨实体业务规则（如"删 KB 必须回收 embedding 配额"） | 聚合根 |
| 复杂状态流转（会话有 10+ 状态） | 领域模型 struct + 状态模式 |
| 需要审计溯源 | 领域事件 |
| 多种创建方式（模板/导入/复制各有不同规则） | 工厂模式 |
| 单包超过 15 个文件或 2000 行 | 拆分为 domain/ + store/ 两个子包 |
