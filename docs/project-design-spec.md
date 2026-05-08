# HJM-Admin 项目设计规范与开发指导文档

> 本文档基于 `hjm-admin` + `hjm-internal` 工程实际代码总结而成，作为后续新模块开发的标准规范和指导思想。

---

## 一、工程总体架构

### 1.1 多仓库 + Go Workspace 协作模式

项目采用 **双仓库** 设计，通过 `go.work` 实现本地联合开发：

```
f:\work\
├── hjm-admin/          # 业务服务（REST API 层）
│   └── go.work         # Go Workspace，引用 hjm-internal
└── hjm-internal/       # 基础设施共享层（数据模型、代码生成、序列化）
```

**设计意图：**
- `hjm-internal` 作为**跨服务共享包**，承载所有数据库模型（PO）、GORM Gen 查询代码、序列化工具等基础设施
- `hjm-admin` 作为**具体业务服务**，仅关注业务逻辑，通过 `go.work` 引用共享包
- 未来新增微服务（如 `hjm-user`、`hjm-order`）可复用同一个 `hjm-internal`，避免重复定义数据模型

**规范要求：**
- 所有 PO（Persistent Object）定义 **必须** 放在 `hjm-internal/infra/data/po/` 中
- 所有 GORM Gen 查询代码 **必须** 放在 `hjm-internal/infra/data/{serviceName}Query/` 中
- 业务服务 **禁止** 直接定义数据库表结构，必须依赖 `hjm-internal`

### 1.2 技术栈选型

| 层面 | 技术选型 | 说明 |
|------|---------|------|
| HTTP 框架 | go-zero (rest) | 高性能 REST 框架，内置参数校验、路由分组 |
| ORM | GORM + GORM Gen | 类型安全的数据库操作，自动代码生成 |
| 数据库 | PostgreSQL | 支持 JSONB 等高级类型 |
| API 定义 | goctl .api 文件 | API-First 开发，自动生成 handler/types/logic 脚手架 |
| 配置管理 | go-zero conf | 统一配置解析，嵌入 rest.RestConf 并扩展自定义字段 |
| 缓存 | Redis + go-redis v9 | 缓存与会话存储 |
| 链路追踪 | OpenTelemetry + Jaeger | OTLP gRPC 导出，分布式链路追踪 |
| 指标采集 | Prometheus + Grafana | HTTP 请求指标、依赖状态指标、可视化面板 |
| 本地开发环境 | Docker Compose | 一键编排 PG/Redis/Jaeger/Prometheus/Grafana |
| API 文档 | Swagger (goctl 生成) | 自动化文档输出 |

---

## 二、分层架构设计（核心）

### 2.1 六层架构总览

```
┌──────────────────────────────────────────────────────┐
│  API 定义层    api/*.api                              │  ← 接口契约，API-First
├──────────────────────────────────────────────────────┤
│  Handler 层    internal/handler/{group}/              │  ← HTTP 入口，参数解析
├──────────────────────────────────────────────────────┤
│  Logic 层      internal/logic/{group}/                │  ← 业务编排，核心逻辑
├──────────────────────────────────────────────────────┤
│  Domain 层     internal/domain/                       │  ← 领域模型 + 校验 + 转换
│                ├── model/     领域模型与校验规则       │
│                └── assembler/ DTO 转换器              │
├──────────────────────────────────────────────────────┤
│  Repo 层       internal/repo/                         │  ← 仓储聚合，组合 DAO
│                └── dao/       数据访问对象             │  ← 最底层数据操作
├──────────────────────────────────────────────────────┤
│  Infrastructure  hjm-internal/infra/                  │  ← 跨服务共享基础设施
│                  ├── data/po/         持久化对象       │
│                  ├── data/adminQuery/ GORM Gen 查询    │
│                  ├── resource/        数据库工具       │
│                  └── serialize/       自定义序列化     │
└──────────────────────────────────────────────────────┘
```

### 2.2 请求处理完整流程

```
HTTP Request
    │
    ▼
[Handler] ─── httpx.Parse(r, &req) 参数绑定与 tag 级校验
    │
    ▼
[Logic]   ─── assembler 转换为 Domain Model → 业务校验 → 调用 Repo → 构造响应
    │
    ▼
[Domain]  ─── model.CheckXxx() 领域规则校验 / assembler 做 DTO ↔ PO 转换
    │
    ▼
[Repo]    ─── 聚合一个或多个 DAO，提供仓储级入口
    │
    ▼
[DAO]     ─── 使用 GORM Gen Query 执行数据库操作
    │
    ▼
[PO/Query]─── hjm-internal 自动生成的持久化对象与类型安全查询
```

---

## 三、各层设计规范

### 3.1 API 定义层（`api/`）

**目录结构：**
```
api/
├── entry.api          # 入口文件，import 所有子模块
├── comm.api           # 通用类型定义（CommResponse 等）
├── user.api           # 用户模块 API
├── role.api           # 角色模块 API
├── permission.api     # 权限模块 API
└── auth.api           # 认证模块 API
```

**规范要求：**

1. **入口文件模式**：`entry.api` 仅做 import 汇总，不定义任何类型或路由
   ```api
   syntax = "v1"
   import (
       "user.api"
       "auth.api"
       "role.api"
       "permission.api"
   )
   ```

2. **通用响应抽取**：所有模块共用 `CommResponse`，定义在 `comm.api` 中
   ```api
   type CommResponse {
       Code      int32  `json:"code"`
       Msg       string `json:"msg"`
       Timestamp int64  `json:"timestamp"`
   }
   ```

3. **模块文件规范**：每个 `.api` 文件包含完整的 info 元数据
   ```api
   info (
       title:   "用户管理接口"
       desc:    "用户管理相关接口，包括注册、登录、用户信息管理"
       author:  "mengxin"
       email:   "mengxin@example.com"
       version: "1.0.0"
   )
   ```

4. **路由分组**：使用 `@server` 指令按业务模块分组，统一前缀
   ```api
   @server (
       tags:   "用户管理"
       group:  "user"           // 对应 handler/user/ 和 logic/user/ 目录
       prefix: /api/v1/admin    // RESTful 版本化前缀
   )
   ```

5. **类型设计 — 请求与展示分离**：
   - **完整模型**（如 `UserBase`）：包含所有字段，用于内部引用
   - **创建请求**（如 `CreateUserRequest`）：仅包含创建所需字段
   - **展示模型**（如 `ShowUserBase`）：脱敏后的对外展示字段，**排除** password、secret、deleted_at 等敏感字段
   - **列表响应**：嵌入 `CommResponse` + 独立 Data 结构体（含 `Total` + `List`）

6. **参数校验 Tag**：充分利用 go-zero 内置校验
   - 分页：`form:"page,range=[1:)"`、`form:"page_size,range=[0:50]"`
   - 可选字段：`json:"nick_name,optional"` 或 `form:"name,optional"`
   - 路径参数：`path:"id"`

7. **RESTful 路由设计**：
   - CRUD：`POST /user`、`GET /user`、`GET /user/:id`、`PUT /user/:id`、`DELETE /user/:id`
   - 子资源：`POST /user/:id/roles`、`GET /user/:id/roles`、`DELETE /user/:id/roles/:role_id`

### 3.2 Handler 层（`internal/handler/`）

**特征：** 由 `goctl` 自动生成，标记为 `Safe to edit`。

**标准模板：**
```go
func CreateUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.CreateUserRequest
        if err := httpx.Parse(r, &req); err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }
        l := user.NewCreateUserLogic(r.Context(), svcCtx)
        resp, err := l.CreateUser(&req)
        if err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
        } else {
            httpx.OkJsonCtx(r.Context(), w, resp)
        }
    }
}
```

**规范要求：**
- Handler **只做** 三件事：参数绑定 → 调用 Logic → 返回响应
- **禁止** 在 Handler 中编写业务逻辑
- 错误一律通过 `httpx.ErrorCtx` 传递，由全局错误处理器统一格式化
- 按 `@server group` 分子目录存放（`handler/user/`、`handler/role/` 等）

### 3.3 Logic 层（`internal/logic/`）

**特征：** 业务逻辑核心层，所有编排逻辑在此实现。

**标准结构：**
```go
type CreateUserLogic struct {
    logx.Logger                    // 嵌入日志器
    ctx    context.Context         // 请求上下文
    svcCtx *svc.ServiceContext     // 服务依赖容器
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
    return &CreateUserLogic{
        Logger: logx.WithContext(ctx),  // 带链路追踪的日志
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}
```

**业务编排标准流程：**
```
1. assembler 将 Request → Domain Model
2. Domain Model 校验（CheckXxx 方法）
3. [可选] 业务前置检查（唯一性、权限等，通过 Repo 查询）
4. [可选] assembler 将 Domain Model → PO
5. 调用 Repo 方法执行数据操作
6. [可选] assembler 将 PO → ShowXxx
7. 构造统一响应返回
```

**Logic 调用 Repo 示例：**
```go
func (l *CreateUserLogic) CreateUser(req *types.CreateUserRequest) (resp *types.CommResponse, err error) {
    // 1. Request → Domain Model
    modelUser := assembler.CreateUserRequestToModelUser(req)
    // 2. 领域校验
    if err = modelUser.CheckEmail(); err != nil {
        return nil, errorx.NewInternalError(err.Error())   // 直接复用 Domain 错误描述
    }
    // 3. 加密 + 转换为 PO
    encryptedPwd, _ := modelUser.EncryptPassword()
    userPo := assembler.ModelUserToUserBasePo(modelUser, encryptedPwd)
    // 4. 调用 Repo（不直接访问 DAO）
    if err = l.svcCtx.UserBaseRepo.Create(l.ctx, userPo); err != nil {
        return nil, errorx.NewInternalError("用户创建失败")
    }
    // 5. 构造响应
    return &types.CommResponse{Code: 0, Msg: "注册成功", Timestamp: time.Now().Unix()}, nil
}
```

**规范要求：**
- 每个 Logic **必须** 嵌入 `logx.Logger`，使用 `l.Logger.Errorf(...)` 记录错误
- 业务错误使用 `errorx.NewInternalError(msg)` 或 `errorx.NewBusinessError(code, msg)`
- Domain `CheckXxx()` 返回的错误信息直接传递给 `errorx.NewInternalError(err.Error())`，**禁止** 重复硬编码相同提示
- Logic 通过 **Repo 接口方法** 操作数据，**禁止** 穿透 Repo 直接访问 DAO，**禁止** 直接操作 SQL 或 GORM
- 对象转换 **必须** 通过 assembler 完成，**禁止** 在 Logic 中手动构造 PO
- 成功响应统一格式：`CommResponse{Code: 0, Msg: "xxx成功", Timestamp: time.Now().Unix()}`
- 按 group 分子目录存放，与 handler 一一对应

### 3.4 Domain 层（`internal/domain/`）

#### 3.4.1 领域模型（`domain/model/`）

**设计思想：** 领域模型是**纯业务规则**的载体，不依赖框架、不关心持久化。

**标准模式：**
```go
// 业务常量：正则等校验规则在包级别定义
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// 领域模型：只包含业务属性
type UserBase struct {
    NickName string
    Email    string
    Password string
}

// 校验方法：以 Check 为前缀，返回 error
func (u *UserBase) CheckEmail() error { ... }

// 行为方法：领域内的业务操作
func (u *UserBase) EncryptPassword() (string, error) { ... }
func (u *UserBase) CheckPasswordMatch(hashedPassword string) bool { ... }
```

**规范要求：**
- 每个核心业务实体对应一个领域模型文件（`user_base.go`、`role.go`、`permission.go`）
- 校验方法命名：`Check{FieldName}()`，返回 `error`
- 正则表达式 **必须** 使用 `regexp.MustCompile` 预编译为包级变量
- 领域模型 **不包含** ID、时间戳等持久化字段——这些属于 PO
- 业务工具方法（如 `IsBuiltinRole`）可作为包级函数提供

#### 3.4.2 转换器（`domain/assembler/`）

**设计思想：** 实现不同层对象之间的转换，隔离层间数据耦合。

**三类转换函数（必须完整覆盖）：**
```go
// ① Request → Domain Model：{Action}RequestToModel{Entity}
func CreateUserRequestToModelUser(req *types.CreateUserRequest) *model.UserBase

// ② Domain Model → PO：Model{Entity}To{Entity}Po
func ModelUserToUserBasePo(m *model.UserBase, encryptedPwd string) *po.UserBase {
    return &po.UserBase{
        NickName: m.NickName,
        Email:    m.Email,
        Password: encryptedPwd,
    }
}

// ③ PO → ShowType：{Entity}PoTo{ShowType}
func UserBasePoToShowUserBase(userBase *po.UserBase) types.ShowUserBase
```

**规范要求：**
- 转换函数按**实体**组织文件（`assembler/user_base.go`）
- **三个方向必须完整覆盖**：Request → Model、Model → PO、PO → Show，**禁止** 在 Logic 中手动构造 PO
- 时间字段在 assembler 中完成 `time.Time → int64 (Unix)` 的转换
- 敏感字段在 PO → Show 转换时**必须过滤**（如 password 不出现在 ShowUserBase 中）
- 加密后的密码等派生值通过函数参数传入，assembler 不承担加密等业务逻辑

### 3.5 Repo + DAO 层（`internal/repo/`）

#### 3.5.1 Repo 仓储层

**设计思想：** Repo 是 DAO 的**聚合容器**，对外暴露**业务语义**的方法，封装底层 DAO 细节。Logic 层 **禁止** 直接穿透 Repo 访问 DAO 字段。

**接口定义（推荐）：** 为每个 Repo 定义接口，便于单元测试 Mock：

```go
// internal/repo/interface.go — Repo 接口定义
type UserRepository interface {
    Create(ctx context.Context, user *po.UserBase) error
    GetByID(ctx context.Context, id int64) (*po.UserBase, error)
    GetList(ctx context.Context, page, pageSize int) ([]*po.UserBase, int64, error)
    Update(ctx context.Context, id int64, updates map[string]any) error
    Delete(ctx context.Context, id int64) error
}

type RoleRepository interface {
    Create(ctx context.Context, role *po.Role) error
    GetByID(ctx context.Context, id int64) (*po.Role, error)
    GetByCode(ctx context.Context, code string) (*po.Role, error)
    GetList(ctx context.Context, page, pageSize int, name, code string) ([]*po.Role, int64, error)
    Delete(ctx context.Context, id int64) error
    // 跨 DAO 的聚合操作
    CreateWithPermissions(ctx context.Context, role *po.Role, permIDs []int64) error
}
```

**实现示例：** Repo 封装 DAO，对 Logic 层屏蔽内部结构：

```go
// internal/repo/role_repo.go
type RoleRepo struct {
    roleDao           *dao.RoleDao
    rolePermissionDao *dao.RolePermissionDao
}

// 对外暴露业务方法，而非暴露 DAO 字段
func (r *RoleRepo) Create(ctx context.Context, role *po.Role) error {
    return r.roleDao.CreateRole(ctx, role)
}

func (r *RoleRepo) GetByCode(ctx context.Context, code string) (*po.Role, error) {
    return r.roleDao.GetRoleByCode(ctx, code)
}
```

**事务管理模式：** 涉及多个 DAO 的原子操作 **必须** 在 Repo 层封装事务：

```go
// 跨 DAO 事务示例：创建角色并分配权限
func (r *RoleRepo) CreateWithPermissions(ctx context.Context, role *po.Role, permIDs []int64) error {
    return r.db.Transaction(func(tx *adminQuery.Query) error {
        roleDao := r.roleDao.WithTx(tx)
        rpDao := r.rolePermissionDao.WithTx(tx)
        if err := roleDao.CreateRole(ctx, role); err != nil {
            return err
        }
        rps := make([]*po.RolePermission, 0, len(permIDs))
        for _, pid := range permIDs {
            rps = append(rps, &po.RolePermission{RoleID: role.ID, PermissionID: pid})
        }
        return rpDao.BatchCreateRolePermissions(ctx, rps)
    })
}
```

**规范要求：**
- 每个业务实体一个 Repo 文件（`user_base_repo.go`、`role_repo.go`）
- Repo **必须** 定义对应的接口（放在 `repo/interface.go`），ServiceContext 依赖接口类型
- Repo 对外暴露业务语义方法，**禁止** Logic 层直接访问 `repo.XxxDao` 字段
- 关联表的 DAO（如 `RolePermissionDao`）挂在主实体的 Repo 下
- 涉及多个 DAO 的写操作 **必须** 通过 Repo 封装事务，使用 `WithTx` 传递事务实例

#### 3.5.2 DAO 数据访问层

**设计思想：** DAO 是最底层的数据库操作封装，使用 GORM Gen 的类型安全查询。

**标准模式：**
```go
type UserBaseDao struct {
    db *adminQuery.Query       // GORM Gen 生成的类型安全 Query
}

func NewUserBaseDao(db *adminQuery.Query) *UserBaseDao {
    return &UserBaseDao{db: db}
}

// 事务支持：返回绑定到事务的新实例
func (d *UserBaseDao) WithTx(tx *adminQuery.Query) *UserBaseDao {
    return &UserBaseDao{db: tx}
}

// CRUD 操作示例 — 日志使用 logx.WithContext(ctx) 按请求获取
func (d *UserBaseDao) CreateUser(ctx context.Context, user *po.UserBase) error {
    err := d.db.UserBase.WithContext(ctx).Create(user)
    if err != nil {
        logx.WithContext(ctx).Errorf("CreateUser err: %v", err)
        return err
    }
    return nil
}
```

> **注意：** DAO 中的日志 **必须** 使用 `logx.WithContext(ctx)` 按请求上下文获取，而非保存在 struct 字段中。struct 字段的 `logx.Logger` 零值无法携带链路信息，会导致日志丢失 trace-id。

**规范要求：**
- 每个数据库表对应一个 DAO 文件
- **所有方法** 必须接收 `context.Context` 参数（支持链路追踪和超时控制）
- **所有方法** 必须使用 `WithContext(ctx)` 执行查询
- 错误日志统一格式：`logx.WithContext(ctx).Errorf("{MethodName} err: %v", err)`
- 每个 DAO **必须** 提供 `WithTx` 方法支持事务传递
- 分页查询统一模式：返回 `([]*po.Xxx, int64, error)` 即 `(列表, 总数, 错误)`
- 更新操作使用 `map[string]any` 支持部分更新
- 条件过滤使用 GORM Gen 的链式 `Where` 构建

### 3.6 ServiceContext 依赖注入（`internal/svc/`）

**设计思想：** `ServiceContext` 是全局的**依赖注入容器**，在服务启动时一次性初始化所有依赖。

```go
type ServiceContext struct {
    Config         config.Config
    UserBaseRepo   repo.UserRepository       // 接口类型，便于 Mock 测试
    RoleRepo       repo.RoleRepository
    PermissionRepo repo.PermissionRepository
    UserRoleRepo   repo.UserRoleRepository
}

func NewServiceContext(c config.Config) *ServiceContext {
    db := resource.InitDatabase(c.DataBase.BuildDSN())     // 1. 初始化数据库连接
    adminQuery.SetDefault(db)                                // 2. 设置 GORM Gen 默认实例
    adminQueryDB := adminQuery.Use(db)                       // 3. 获取类型安全 Query
    return &ServiceContext{
        Config:       c,
        UserBaseRepo: repo.NewUserbaseRepo(                  // 4. 逐层组装：Query → DAO → Repo
            dao.NewUserBaseDao(adminQueryDB)),
        RoleRepo: repo.NewRoleRepo(
            dao.NewRoleDao(adminQueryDB),
            dao.NewRolePermissionDao(adminQueryDB)),
        // ...
    }
}
```

**组装顺序：** `DB Connection → GORM Gen Query → DAO → Repo(实现接口) → ServiceContext`

**规范要求：**
- 所有 Repo 的实例化 **必须** 在 `NewServiceContext` 中完成
- ServiceContext 中的 Repo 字段 **必须** 使用接口类型（`repo.XxxRepository`），不使用具体 struct 指针
- 新增业务模块时，先在 `repo/interface.go` 定义接口，再实现并在此处注册
- 数据库资源初始化放在 `svc/resource/` 子包中
- 单元测试时可直接构造 `ServiceContext{XxxRepo: mockImpl}` 注入 Mock 实现，无需启动数据库

### 3.7 错误处理（`internal/common/errorx/`）

**设计思想：** 统一错误响应格式，将错误分为**业务错误**（可预期）和**系统错误**（不可预期）两类，分别处理。

```go
// 错误类型
type BusinessError struct {
    Code int    `json:"code"`
    Msg  string `json:"msg"`
}

// 两种构造方式
errorx.NewInternalError("操作失败")              // Code=-1，通用内部错误
errorx.NewBusinessError(400, "角色编码已存在")    // 自定义业务错误码

// 全局错误拦截（在 main 中注册）
func HandleError() {
    httpx.SetErrorHandlerCtx(func(ctx context.Context, err error) (int, any) {
        // 业务错误：已通过 errorx 包装，返回 HTTP 200 + JSON code
        var bizErr *BusinessError
        if errors.As(err, &bizErr) {
            return http.StatusOK, bizErr
        }
        // 系统错误：未预期的异常，返回 HTTP 500，隐藏内部细节
        logx.WithContext(ctx).Errorf("unexpected error: %v", err)
        return http.StatusInternalServerError, &BusinessError{
            Code: CodeFailed,
            Msg:  "服务器内部错误",
        }
    })
}
```

**错误码规范：**

| code | 含义 | 使用场景 |
|------|------|---------|
| `0` | 成功 | 所有正常响应 |
| `-1` | 通用失败 | `NewInternalError`，参数校验失败、操作失败等 |
| `> 0` | 业务错误码 | `NewBusinessError`，如 400=重复数据、403=权限不足 |

**规范要求：**
- 业务错误（参数不合法、规则不满足）返回 HTTP 200 + JSON `code != 0`
- 系统错误（数据库崩溃、未预期 panic）返回 HTTP 500，**禁止** 将原始错误信息暴露给前端
- Logic 层抛出错误时 **必须** 使用 `errorx` 包装，**禁止** 直接返回原始 error
- Domain 层 `CheckXxx()` 返回的 error 信息可直接传递给 `errorx.NewInternalError(err.Error())`，**禁止** 在 Logic 中重复硬编码相同的错误提示

### 3.8 中间件层（Middleware）

**设计思想：** 通过 go-zero 的 `@server` 指令和自定义中间件实现横切关注点（认证、日志、限流等），与业务逻辑解耦。

**JWT 认证（推荐模式）：** 需要鉴权的路由组在 `.api` 文件中声明 `jwt` 指令：

```api
@server (
    jwt:    Auth                // 开启 JWT 中间件，对应 Config 中的 Auth 字段
    group:  "user"
    prefix: /api/v1/admin
)
service adminService {
    @handler GetUserDetail
    get /user/:id (GetUserDetailRequest) returns (GetUserDetailResponse)
}
```

对应 Config 中添加：
```go
type Config struct {
    rest.RestConf
    Auth struct {
        AccessSecret string
        AccessExpire int64
    }
    DataBase Postgres
}
```

**不需要鉴权的路由**（如登录）单独放在无 `jwt` 的 `@server` 块中。

**自定义中间件注册：**
```go
// 在 adminservice.go 中注册全局中间件
server := rest.MustNewServer(c.RestConf)
server.Use(middleware.RequestIdMiddleware)    // 请求ID
server.Use(middleware.AccessLogMiddleware)    // 访问日志
```

**规范要求：**
- 公开接口（登录、注册）与鉴权接口 **必须** 分属不同的 `@server` 块
- 中间件代码放在 `internal/middleware/` 目录下
- 全局中间件在 `adminservice.go` 中注册，路由级中间件通过 `.api` 指令声明

---

## 四、共享基础设施层（`hjm-internal`）

### 4.1 职责划分

```
hjm-internal/
├── config/             # 基础设施配置（go-zero conf 驱动）
├── infra/
│   ├── data/
│   │   ├── po/             # 持久化对象（GORM Gen 自动生成，DO NOT EDIT）
│   │   └── adminQuery/     # 类型安全查询代码（GORM Gen 自动生成，DO NOT EDIT）
│   └── resource/
│       └── database.go     # GORM Gen 代码生成工具
├── serialize/          # 自定义序列化类型（JSONB 映射）
├── docs/               # 数据库迁移脚本等文档
└── main.go             # 代码生成入口
```

### 4.2 GORM Gen 代码生成流程

```
1. 在 PostgreSQL 中创建/修改表结构
2. 配置 hjm-internal/etc/config-{env}.yaml 中的数据库连接和生成路径
3. 运行 hjm-internal/main.go（go run main.go --env=dev）
4. 自动生成：
   - infra/data/po/*.gen.go       → 表结构到 Go struct 的映射
   - infra/data/adminQuery/*.go   → 类型安全的查询 API
5. 业务服务通过 go.work 引用生成结果
```

### 4.3 自定义序列化（JSONB 支持）

对于 PostgreSQL JSONB 字段，需要在 `serialize/` 中定义对应的 Go 类型，并实现 `driver.Valuer` 和 `sql.Scanner` 接口：

```go
type UserSummary struct {
    Bio   string `json:"bio"`
    Level int    `json:"level"`
}

func (u UserSummary) Value() (driver.Value, error)  { return json.Marshal(u) }
func (u *UserSummary) Scan(value any) error          { ... }
```

在 `main.go` 中通过 `FieldCustomField` 配置 JSONB 字段到自定义类型的映射：
```go
customFields := []resource.FieldCustomField{
    {TableName: "user_base", ColumnName: "summary", FieldType: "serialize.UserSummary"},
}
```

---

## 五、配置管理规范

### 5.1 业务服务配置（go-zero conf）

```yaml
# etc/adminservice.yaml
Name: admin_service
Host: 0.0.0.0
Port: 8888

DataBase:
  Host: localhost
  Port: 5432
  User: postgres
  Password: postgres123
  DBName: admin
  SSLMode: disable
  TimeZone: Asia/Shanghai
```

**规范要求：**
- 通过 `Config` struct 嵌入 `rest.RestConf` 继承 go-zero 标准配置
- 自定义配置（如数据库）在 `Config` 中扩展独立 struct
- 支持多环境配置：`adminservice.yaml`（开发）、`adminservice-prod.yaml`（生产）
- DSN 构建逻辑封装在 Config struct 的方法中（如 `Postgres.BuildDSN()`）

### 5.2 数据库连接池配置

```go
// svc/resource/gorm.go
sqlDB.SetMaxIdleConns(20)            // 空闲连接数
sqlDB.SetMaxOpenConns(100)           // 最大连接数
sqlDB.SetConnMaxLifetime(time.Second * 30)  // 连接最大生命周期
```

---

## 六、代码生成与自动化

### 6.1 Makefile 命令

```makefile
make api       # goctl api go -api ./api/entry.api -dir ./
make swagger   # goctl api swagger --api ... --dir ./swagger
make all       # 执行 api + swagger
```

### 6.2 代码生成边界

| 文件/目录 | 生成方式 | 可否编辑 |
|-----------|---------|---------|
| `internal/types/types.go` | goctl 生成 | ❌ **禁止编辑** |
| `internal/handler/routes.go` | goctl 生成 | ❌ **禁止编辑** |
| `internal/handler/{group}/*.go` | goctl 脚手架 | ✅ 可编辑（通常无需改动） |
| `internal/logic/{group}/*.go` | goctl 脚手架 | ✅ **主要编辑区** |
| `internal/domain/**` | 手动编写 | ✅ 手动维护 |
| `internal/repo/**` | 手动编写 | ✅ 手动维护 |
| `hjm-internal/infra/data/po/**` | GORM Gen 生成 | ❌ **禁止编辑** |
| `hjm-internal/infra/data/adminQuery/**` | GORM Gen 生成 | ❌ **禁止编辑** |

---

## 七、新模块开发 SOP（标准操作流程）

以新增「部门管理」模块为例：

### Step 1：数据库建表
在 PostgreSQL 中创建 `department` 表。

### Step 2：生成 PO 和 Query
```bash
cd hjm-internal
go run main.go --env=dev
# → 自动生成 infra/data/po/department.gen.go
# → 自动生成 infra/data/adminQuery/department.gen.go
```

### Step 3：定义 API
创建 `api/department.api`：
```api
syntax = "v1"
import "comm.api"

info (
    title:   "部门管理接口"
    author:  "mengxin"
    version: "1.0.0"
)

type CreateDepartmentRequest { ... }
type ShowDepartment { ... }
// ... 其他类型

@server (
    tags:   "部门管理"
    group:  "department"
    prefix: /api/v1/admin
)
service adminService {
    @handler CreateDepartment
    post /department (CreateDepartmentRequest) returns (CommResponse)
    // ...
}
```

在 `entry.api` 中添加 import：
```api
import "department.api"
```

### Step 4：生成代码
```bash
make api
# → 自动生成 handler/department/*.go
# → 自动生成 logic/department/*.go
# → 更新 types/types.go 和 handler/routes.go
```

### Step 5：编写 Domain 层
- 新建 `internal/domain/model/department.go`，定义领域模型和 Check 方法
- 新建 `internal/domain/assembler/department.go`，定义转换函数

### Step 6：编写 DAO
- 新建 `internal/repo/dao/department_dao.go`

### Step 7：定义 Repo 接口并实现
- 在 `internal/repo/interface.go` 中添加 `DepartmentRepository` 接口定义
- 新建 `internal/repo/department_repo.go`，实现该接口，内部聚合 DAO

### Step 8：注册依赖
在 `internal/svc/servicecontext.go` 的 `ServiceContext` 中添加 `DepartmentRepo` 字段（接口类型）并初始化。

### Step 9：实现 Logic
在 `internal/logic/department/` 下各 logic 文件中实现业务逻辑。

### Step 10：生成文档
```bash
make swagger
```

---

## 八、命名规范总结

| 类别 | 规范 | 示例 |
|------|------|------|
| API 文件 | 小写模块名.api | `user.api`、`role.api` |
| Handler | {Action}{Entity}Handler | `CreateUserHandler` |
| Logic | {Action}{Entity}Logic | `CreateUserLogic` |
| DAO | {Entity}Dao | `UserBaseDao`、`RolePermissionDao` |
| Repo 接口 | {Entity}Repository | `UserRepository`、`RoleRepository` |
| Repo 实现 | {Entity}Repo | `UserbaseRepo`、`RoleRepo` |
| Domain Model | {Entity} | `UserBase`、`Role`、`Permission` |
| Assembler Req→Model | {Action}RequestToModel{Entity} | `CreateUserRequestToModelUser` |
| Assembler Model→PO | Model{Entity}To{Entity}Po | `ModelUserToUserBasePo` |
| Assembler PO→Show | {Entity}PoTo{ShowType} | `UserBasePoToShowUserBase` |
| 校验方法 | Check{Field} | `CheckEmail()`、`CheckCode()` |
| PO (自动生成) | {TableName} | `po.UserBase`、`po.Role` |
| 展示类型 | Show{Entity} | `ShowUserBase`、`ShowRole` |
| 请求类型 | {Action}{Entity}Request | `CreateUserRequest`、`GetUserListRequest` |
| 响应类型 | {Action}{Entity}Response | `GetUserListResponse`、`GetRoleDetailResponse` |
| DAO 文件 | {snake_case}_dao.go | `user_base_dao.go` |
| Repo 接口文件 | interface.go | `repo/interface.go` |
| Logic 文件 | {lowercase_action}{entity}logic.go | `createuserlogic.go` |
| 中间件 | {Purpose}Middleware | `RequestIdMiddleware`、`AccessLogMiddleware` |

---

## 九、运行时基础设施规范

> 基于 AIPivot 项目实践总结，覆盖可观测性、健康检查、服务生命周期、本地开发环境等运行时关注点。与前述业务架构规范互补，共同构成完整的工程规范体系。

### 9.1 可观测性架构总览

```
┌──────────────────────────────────────────────────────────────┐
│  HTTP Request                                                 │
│      │                                                        │
│      ▼                                                        │
│  [Middleware]  RequestID → Tracing Span → Metrics → Log       │
│      │                                                        │
│      ▼                                                        │
│  [Handler → Logic → Repo → DAO]  业务处理                      │
│      │                                                        │
│      ▼                                                        │
│  [Middleware]  记录状态码、耗时、设置 Span 属性                   │
└──────────┬───────────────┬──────────────────┬─────────────────┘
           │               │                  │
           ▼               ▼                  ▼
       [Jaeger]      [Prometheus]         [stdout]
       链路追踪        指标采集           结构化日志
           │               │
           ▼               ▼
       Jaeger UI      Grafana Dashboard
```

**三大支柱：**
- **Tracing**：OpenTelemetry SDK + OTLP gRPC → Jaeger，每个请求自动生成 TraceID
- **Metrics**：Prometheus 客户端，自定义指标 + `/metrics` 端点暴露 + Grafana 可视化
- **Logging**：go-zero logx 结构化日志，自动携带 RequestID 和 TraceID

### 9.2 链路追踪（OpenTelemetry + Jaeger）

**初始化模式：**
```go
// internal/observability/tracing.go
func InitTracing(ctx context.Context, conf config.TelemetryConf) (func(context.Context) error, error) {
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpointURL(conf.JaegerEndpoint),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, fmt.Errorf("create otlp trace exporter: %w", err)
    }

    sampler := sdktrace.AlwaysSample()
    if conf.SampleRatio >= 0 && conf.SampleRatio < 1 {
        sampler = sdktrace.TraceIDRatioBased(conf.SampleRatio)
    }

    provider := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),     // 生产环境用异步批量导出
        sdktrace.WithSampler(sampler),
        sdktrace.WithResource(resource.NewWithAttributes("",
            attribute.String("service.name", conf.ServiceName),
            attribute.String("deployment.environment", conf.Environment),
        )),
    )
    otel.SetTracerProvider(provider)
    otel.SetTextMapPropagator(propagation.TraceContext{})
    return provider.Shutdown, nil  // 返回关闭函数，由 ServiceContext 管理
}
```

**配置结构体：**
```go
type TelemetryConf struct {
    ServiceName    string
    Environment    string
    JaegerEndpoint string
    SampleRatio    float64 `json:",default=1"`  // 1=全采样, 0~1=按比例采样
}
```

**规范要求：**
- InitTracing **必须** 返回 `shutdown` 函数，由 ServiceContext 在服务关闭时调用
- 生产环境 **必须** 使用 `WithBatcher` 异步批量导出，开发环境可用 `WithSyncer` 同步导出
- 采样率通过配置文件控制，开发环境设为 `1.0`，生产环境按流量调整
- 错误处理通过 `otel.SetErrorHandler` 统一记录到 logx

### 9.3 指标采集（Prometheus）

**自定义指标注册模式：**
```go
// internal/observability/metrics.go
type Metrics struct {
    registry        *prometheus.Registry
    httpRequests    *prometheus.CounterVec      // 请求计数
    httpDuration    *prometheus.HistogramVec    // 请求耗时分布
    dependencyReady *prometheus.GaugeVec        // 依赖就绪状态
}

func NewMetrics(registry *prometheus.Registry) *Metrics {
    metrics := &Metrics{
        registry: registry,
        httpRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
            Name: "{service}_http_requests_total",
            Help: "Total number of HTTP requests.",
        }, []string{"method", "path", "status"}),
        httpDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
            Name:    "{service}_http_request_duration_seconds",
            Help:    "HTTP request duration in seconds.",
            Buckets: prometheus.DefBuckets,
        }, []string{"method", "path", "status"}),
        dependencyReady: prometheus.NewGaugeVec(prometheus.GaugeOpts{
            Name: "{service}_dependency_ready",
            Help: "Dependency readiness state, 1=ready, 0=not ready.",
        }, []string{"dependency"}),
    }
    registry.MustRegister(metrics.httpRequests, metrics.httpDuration, metrics.dependencyReady)
    return metrics
}
```

**规范要求：**
- 使用**私有 Registry**（`prometheus.NewRegistry()`）而非全局默认 Registry，避免指标污染
- 建议额外注册 `prometheus.NewGoCollector()` 和 `prometheus.NewProcessCollector()` 获取 GC、goroutine 等运行时指标
- 指标命名遵循 Prometheus 规范：`{service}_{subsystem}_{name}_{unit}`
- 标签（label）控制在 3-5 个以内，避免基数爆炸
- `/metrics` 端点通过独立 Handler 暴露，使用 `promhttp.HandlerFor(registry, ...)`

### 9.4 可观测性中间件链

**中间件职责（单一中间件统一处理所有横切关注点）：**
```
Request → [生成/提取 RequestID] → [提取 Trace 上下文] → [创建 Span]
       → [执行业务] → [记录指标] → [设置 Span 属性] → [写入访问日志] → Response
```

**标准实现模式：**
```go
// internal/observability/middleware.go
func Middleware(metrics *Metrics, serviceName string) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            // 1. RequestID: 从 Header 提取或自动生成 UUID
            requestID := r.Header.Get(RequestIDHeader)
            if requestID == "" {
                requestID = uuid.NewString()
            }
            // 2. Trace 上下文传播（支持跨服务传递）
            ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
            ctx = ContextWithRequestID(ctx, requestID)
            // 3. 创建 Span
            ctx, span := otel.Tracer(serviceName).Start(ctx, "HTTP "+r.Method+" "+r.URL.Path)
            defer span.End()
            // 4. 包装 ResponseWriter 捕获状态码
            recorder := newStatusRecorder(w)
            recorder.Header().Set(RequestIDHeader, requestID)
            // 5. 执行业务处理
            next(recorder, r.WithContext(ctx))
            // 6. 记录 Span 属性
            duration := time.Since(start)
            span.SetAttributes(
                attribute.String("http.method", r.Method),
                attribute.String("http.route", r.URL.Path),
                attribute.Int("http.status_code", recorder.status),
            )
            if recorder.status >= http.StatusBadRequest {
                span.SetStatus(codes.Error, http.StatusText(recorder.status))
            }
            // 7. 记录 Prometheus 指标
            metrics.RecordHTTPRequest(r.Method, r.URL.Path, recorder.status, duration)
            // 8. 结构化访问日志（自动携带 TraceID）
            logx.WithContext(ctx).Infow("http_request",
                logx.Field("request_id", requestID),
                logx.Field("method", r.Method),
                logx.Field("path", r.URL.Path),
                logx.Field("status", recorder.status),
                logx.Field("duration_ms", duration.Milliseconds()),
            )
        }
    }
}
```

**RequestID 上下文传递模式：**
```go
// internal/observability/context.go
const RequestIDHeader = "X-Request-Id"

func ContextWithRequestID(ctx context.Context, id string) context.Context  // 写入 ctx
func RequestIDFromContext(ctx context.Context) string                       // 从 ctx 读取
func TraceIDFromContext(ctx context.Context) string                         // 从 Span 中提取
```

**规范要求：**
- 中间件通过 `server.Use()` 全局注册，对所有路由生效
- RequestID **必须** 同时存入 Context 和 Response Header，便于客户端排查
- StatusRecorder 包装 `http.ResponseWriter` 捕获状态码，**禁止** 在 Handler 中手动记录指标
- 日志 **必须** 使用 `logx.WithContext(ctx)` 确保自动携带 TraceID
- HTTP 指标在中间件层统一记录，业务层 **禁止** 手动记录 HTTP 指标

### 9.5 健康检查规范

**端点设计：**

| 端点 | 用途 | HTTP 状态码 | 适用场景 |
|------|------|------------|---------|
| `GET /healthz` | 存活检查 | 200 = 存活 | K8s livenessProbe |
| `GET /readyz` | 就绪检查 | 200 = 就绪, 503 = 未就绪 | K8s readinessProbe |

**依赖检查模式：**
```go
// internal/infra/health.go
type DependencyCheck struct {
    Name  string
    Check func(ctx context.Context) error  // 返回 nil 表示健康
}

type ReadinessResult struct {
    Status       string             // "ready" 或 "not_ready"
    Ready        bool
    Dependencies []DependencyStatus
}

func CheckDependencies(ctx context.Context, checks []DependencyCheck) ReadinessResult
```

**依赖检查注册（在 ServiceContext 中）：**
```go
HealthChecks: []infra.DependencyCheck{
    infra.CheckPostgres(db),    // 执行 SELECT 1 验证连接
    infra.CheckRedis(client),   // 执行 PING 验证连接
}
```

**规范要求：**
- `/healthz` 是纯存活探针，**禁止** 检查外部依赖，直接返回 `{"status":"ok"}`
- `/readyz` 遍历所有 `DependencyCheck`，任一失败返回 HTTP 503 + 详细依赖状态
- 每个外部依赖（数据库、缓存、消息队列等）**必须** 注册对应的 `DependencyCheck`
- ReadyHandler **应该** 同步更新 Prometheus 依赖状态指标（`metrics.SetDependencyReady`）
- 健康检查路由 **不经过** JWT 等认证中间件

### 9.6 服务生命周期管理

**初始化顺序（ServiceContext 构造）：**
```
1. InitTracing        → 链路追踪（最先初始化，使后续操作可追踪）
2. NewPostgres        → 数据库连接池
3. NewRedis           → 缓存连接
4. NewMetrics         → 指标注册
5. 注册 HealthChecks  → 依赖健康检查
6. 构造 Shutdown 函数 → 资源关闭回收
```

**优雅关闭模式（服务入口）：**
```go
// aipivot.go
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
```

**Shutdown 实现（按依赖关系逆序关闭）：**
```go
// internal/svc/servicecontext.go
Shutdown: func(ctx context.Context) error {
    var shutdownErr error
    if err := tracingShutdown(ctx); err != nil {   // 1. 先刷新 tracing 导出缓冲
        shutdownErr = fmt.Errorf("shutdown tracing: %w", err)
    }
    if err := redisClient.Close(); err != nil {     // 2. 关闭缓存连接
        shutdownErr = fmt.Errorf("close redis: %w", err)
    }
    if err := infra.ClosePostgres(db); err != nil { // 3. 关闭数据库连接池
        shutdownErr = fmt.Errorf("close postgres: %w", err)
    }
    return shutdownErr
}
```

**规范要求：**
- ServiceContext **必须** 包含 `Shutdown func(context.Context) error` 字段
- 关闭顺序与初始化顺序**相反**：先刷新可观测性数据 → 关闭缓存 → 关闭数据库
- Shutdown 超时设为 **5 秒**，防止关闭阻塞导致进程被强杀
- 服务入口通过 `defer` 确保即使 panic 也能执行清理
- 所有资源清理错误 **必须** 记录日志，但不应阻断其他资源的关闭

### 9.7 Redis 集成规范

**初始化模式：**
```go
// internal/infra/redis.go
func NewRedis(conf config.RedisConf) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     conf.Addr,
        Password: conf.Password,
        DB:       conf.DB,
    })
}
```

**配置结构体：**
```go
type RedisConf struct {
    Addr     string
    Password string `json:",optional"`
    DB       int
}
```

**健康检查注册：**
```go
func CheckRedis(client *redis.Client) DependencyCheck {
    return DependencyCheck{
        Name:  "redis",
        Check: func(ctx context.Context) error { return client.Ping(ctx).Err() },
    }
}
```

**规范要求：**
- Redis 客户端通过 `infra.NewRedis()` 初始化，注入 ServiceContext
- **必须** 注册健康检查，纳入 `/readyz` 探测
- **必须** 在 Shutdown 中调用 `Close()` 释放连接
- 业务层通过 `svcCtx.Redis` 访问，**禁止** 使用全局变量

### 9.8 本地开发环境（Docker Compose）

**标准服务编排：**

| 服务 | 镜像 | 端口 | 用途 |
|------|------|------|------|
| PostgreSQL | `postgres:16` | 5432 | 数据库 |
| Redis | `redis:7` | 6379 | 缓存 |
| Jaeger | `jaegertracing/all-in-one` | 16686 (UI) / 4317 (OTLP gRPC) | 链路追踪 |
| Prometheus | `prom/prometheus` | 9090 | 指标采集 |
| Grafana | `grafana/grafana` | 3000 (admin/admin) | 指标可视化 |

**目录结构：**
```
deploy/
├── docker-compose.yml              # 服务编排
├── prometheus/
│   └── prometheus.yml              # 抓取配置
└── grafana/
    ├── datasources/
    │   └── datasources.yml         # 自动配置 Prometheus 数据源
    └── dashboards/
        ├── dashboard.yml           # Dashboard provisioning 配置
        └── http-service.json       # 预置 HTTP 服务面板
```

**规范要求：**
- 每个有状态服务 **必须** 配置 `healthcheck`（PG 用 `pg_isready`，Redis 用 `redis-cli ping`）
- 容器名统一前缀：`{project}-{service}`（如 `aipivot-postgres`、`aipivot-redis`）
- Prometheus 配置文件挂载到 `deploy/prometheus/prometheus.yml`，抓取目标使用 `host.docker.internal:{port}`
- Grafana 通过 provisioning 自动配置 datasource 和 dashboard，**禁止** 手动 UI 配置
- 开发环境配置文件中的服务地址统一使用 `127.0.0.1`
- 一键启动命令：`docker compose -f deploy/docker-compose.yml up -d`

---

## 十、设计亮点总结

1. **API-First 开发**：以 `.api` 文件为单一事实来源（Single Source of Truth），通过 goctl 驱动代码生成，保证接口定义与实现的一致性

2. **领域驱动校验**：校验规则封装在 Domain Model 中（而非分散在 Logic 或 Handler），实现业务规则的集中管理和复用

3. **四层数据对象隔离**：`types (DTO)` → `domain/model` → `po (持久化)` → `Show (展示)`，通过 assembler 三方向全覆盖转换，严格隔离各层关注点

4. **GORM Gen 类型安全**：完全消除手写 SQL 字符串，编译期发现查询错误

5. **共享基础设施仓库**：`hjm-internal` 可被多个微服务复用，统一数据模型和工具函数

6. **Repo 接口抽象 + 事务封装**：Repo 对外暴露业务语义接口，封装 DAO 细节和跨 DAO 事务；ServiceContext 依赖接口类型，支持 Mock 测试

7. **错误分级处理**：业务错误（HTTP 200 + code）与系统错误（HTTP 500）分层处理，兼顾前端对接便利性与运维可观测性

8. **依赖注入链路清晰**：`DB → Query → DAO → Repo(接口) → ServiceContext → Logic`，每一层职责单一、可替换、可测试

9. **中间件与鉴权分离**：公开路由与鉴权路由通过 `@server` 块隔离，JWT 等横切关注点声明式配置

10. **自动化工具链**：Makefile 封装 goctl 命令，一键生成代码和文档

11. **可观测性三件套**：Tracing（OpenTelemetry → Jaeger）+ Metrics（Prometheus + Grafana）+ 结构化日志（logx），每个 HTTP 请求自动覆盖链路追踪、指标采集和访问日志

12. **健康检查双端点**：`/healthz` 存活 + `/readyz` 就绪（含依赖探测），符合 K8s 容器化部署标准

13. **服务生命周期管理**：ServiceContext 统一管理初始化与优雅关闭，Shutdown 按依赖逆序释放资源，5 秒超时防阻塞

14. **一键开发环境**：Docker Compose 编排全套依赖（PG/Redis/Jaeger/Prometheus/Grafana），Grafana 自动 provisioning 数据源与面板

---

*文档版本：v3.0 | 更新日期：2026-05-08 | 基于 hjm-admin 业务架构 + AIPivot 运行时基础设施实践整合*
