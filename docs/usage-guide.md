# AIPivot 本地使用手册

> 日常启动、联调、排错先看这份文档。架构与代码规范见 `CLAUDE.md` 和 `docs/project-design-spec.md`。
> 最后更新：2026-08-23（对应 commit：Flow 执行运行时 + Widget SDK 之后）

---

## 1. 环境要求

| 工具 | 版本要求 | 说明 |
|---|---|---|
| Go | ≥ 1.25 | 本机 1.27.0 |
| Docker Desktop | 任意近期版本 | 跑 PG/Redis/Jaeger/Prometheus/Grafana |
| Node.js | ≥ 18 | 本机 25.2.1 |
| goctl | ≥ 1.8.2 | 仅改 `.api` 时需要 |

**⚠️ C 盘空间**：C 盘经常满（< 100MB），构建/测试前必须把临时目录指到 F 盘：

```bash
export TMP=F:/tmp TMPDIR=F:/tmp GOTMPDIR=F:/tmp
```

---

## 2. 一键启动

### 2.1 启动依赖（Docker）

```bash
docker compose -f deploy/docker-compose.yml up -d
```

容器已常驻，一般不需要重启。查看状态：`docker ps`（postgres 和 redis 应显示 `healthy`）。

### 2.2 配置 LLM Key

项目根目录 `.env`（已被 .gitignore 忽略，不会提交）：

```
ARK_API_KEY=<你的火山方舟 key>
```

启动后端前必须加载它（bash）：

```bash
set -a && . ./.env && set +a
```

**不配 key 的后果**：服务照常启动，但所有对话端点返回 `{"code":1002,"msg":"LLM 网关不可用..."}`。
Key 的读取链路：`etc/aipivot-api.yaml` 的 `LLM.APIKeyEnv: "ARK_API_KEY"` → 环境变量，**yaml 里不写明文 key**。

### 2.3 启动后端（自动跑迁移 + Asynq worker）

```bash
export TMP=F:/tmp TMPDIR=F:/tmp GOTMPDIR=F:/tmp   # 可选，C 盘满时必需
go run aipivot.go -f etc/aipivot-api.yaml
# 或先构建再跑
go build -o bin/aipivot.exe . && ./bin/aipivot.exe -f etc/aipivot-api.yaml
```

启动成功的标志（日志）：

```
migrations applied: version=10 dirty=false
asynq worker started — concurrency=5
Starting aipivot-api at 0.0.0.0:8888...
```

### 2.4 启动前端（管理台）

```bash
make web-dev        # http://localhost:5173，/api 自动代理到 8888
```

### 2.5（可选）启动 Widget SDK 开发服务器

```bash
make widget-dev     # http://localhost:5174，端口固定（strictPort）
```

---

## 3. 登录账号

| 项 | 值 |
|---|---|
| 地址 | http://localhost:5173 |
| 邮箱 | `admin@aipivot.dev` |
| 密码 | `Aipivot@2026` |
| 角色 | admin（租户 1 Default Tenant） |

**注意**：`/api/v1/auth/register` 注册的新用户角色固定为 `member`，访问管理台（`/api/v1/admin/*`）、
Flow（`/api/v1/flows/*`）、Skills（`/api/v1/skills/*`）会 403。本地提升 admin：

```bash
docker exec aipivot-postgres psql -U aipivot -d aipivot \
  -c "UPDATE users SET role='admin' WHERE email='<你的邮箱>';"
```

历史账号 `lixiaoming@qq.com` 也是 admin，但密码是 bcrypt 密文，忘记密码就注册新号再提权。

---

## 4. 服务地址一览

| 服务 | 地址 | 凭证 |
|---|---|---|
| 管理台前端 | http://localhost:5173 | 见上表 |
| 后端 API | http://127.0.0.1:8888 | JWT（登录获取） |
| Jaeger UI | http://127.0.0.1:16686 | 无 |
| Prometheus | http://127.0.0.1:9090 | 无 |
| Grafana | http://127.0.0.1:3000 | admin / admin |
| PostgreSQL | 127.0.0.1:5432 | aipivot / aipivot（库名 aipivot） |
| Redis | 127.0.0.1:6379 | 无密码 |

常用探活：

```bash
curl http://127.0.0.1:8888/healthz   # 存活
curl http://127.0.0.1:8888/readyz    # PG + Redis 依赖状态
```

---

## 5. 管理台功能速览

前端 8 个页面：**Login / Chat / Knowledge / Flow / Skill / Webhook / Analytics / Admin**。

### 5.1 对话（Chat 页）

- 新建会话 → 输入消息 → SSE 流式返回（打字机效果）
- 走 `POST /api/v1/conversations/:id/messages/stream`
- 消息支持满意度评分（👍/👎，仅 Widget 渠道消息有评分按钮的完整闭环）

### 5.2 知识库（Knowledge 页）

- 创建知识库（可配置引导问答，最多 6 条，每条 ≤ 100 字）
- 上传文档：**当前仅支持纯文本 / Markdown**（`.txt` / `.md`），内容暂存 `documents.file_path`
- 上传后异步处理：`pending → processing → ready / failed`，由 Asynq worker 执行切块 + Embedding

**⚠️ 已知限制（Embedding 模型）**：配置的 `EmbeddingModel: "text-embedding-3-small"` 是 OpenAI 的模型名，
但 embedding 请求实际发到火山方舟（`pkg/llm/client.go` 的 `Embed()` 走全局 baseURL，未按模型 provider 分流），
Ark 上没有这个模型 → 文档处理必然 `failed`，RAG 检索 fail-soft（检索失败后直接裸 LLM 对话，功能不断）。

两个解决方向（涉及决策，未动）：
1. 改 `etc/aipivot-api.yaml` 的 `EmbeddingModel` 为 Ark 支持的模型（如 `doubao-embedding` 系列），
   同时 `EmbeddingDim` 要匹配 pgvector 列的 1536 维（migration 000002 固定，换维度要新迁移）
2. 给 `pkg/llm` 增加按 provider 分流的 embedding 客户端，继续用 OpenAI

### 5.3 可视化 Flow（Flow 页，需 admin）

- 拖拽编排节点（LLM / 条件分支等），`definition` 存 JSONB
- 试运行走 `POST /api/v1/flows/:id/run`（SSE：run_start / node_start / delta / node_end / run_end）
- 每次执行落 `flow_runs` 快照表，历史回放不受后续编辑影响

### 5.4 数据分析（Analytics 页）

7 个 KPI 卡片（含满意度），`/api/v1/analytics/overview` + `/daily`。

### 5.5 管理台（Admin 页，需 admin）

租户信息、用户增删改。

---

## 6. Widget 接入（访客悬浮聊天窗）

完整链路：**创建 public key → 嵌入 SDK → 访客自动建会话**。

### 6.1 创建 public key（pk_ 前缀）

在管理台 API Key 页或直接调接口（**必须绑定知识库 + Origin 白名单**）：

```bash
TOKEN=<登录返回的 JWT>
curl -X POST http://127.0.0.1:8888/api/v1/api-keys \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{
    "name": "我的网站 Widget",
    "keyType": "public",
    "knowledgeBaseId": 2,
    "allowedOrigins": ["http://localhost:5174", "https://www.yoursite.com"]
  }'
```

返回的 `data.key`（`pk-` 开头）**只展示一次**，丢了只能重建。

规则（fail-closed）：
- `allowedOrigins` **严格匹配** scheme+host+port，不支持通配符，空列表直接全拒
- public key 强制绑定 `knowledgeBaseId`（限制 RAG 检索范围）

### 6.2 构建与嵌入

```bash
make widget-build    # 产出 widget/dist/aipivot-widget.js（gzip ≈ 17KB）
```

客户网站一行接入：

```html
<script src="/path/to/aipivot-widget.js"></script>
<script>
  window.AIPivotWidget.init({
    publicKey: 'pk-xxxx',
    baseUrl: 'http://127.0.0.1:8888',   // 生产环境换成正式域名
  });
</script>
```

本地调试：`make widget-dev` 后访问 `http://localhost:5174/examples/minimal.html`，
把示例里的 `pk_REPLACE_ME` 换成你的 key。

### 6.3 访客侧行为

- SDK 自动生成 `visitorId`（localStorage 持久化），首次打开创建 widget 渠道会话（复用 `conversations.external_user_id`）
- 发消息走 `POST /api/v1/open/widget/sessions/:sessionToken/messages/stream`（SSE）
- 消息可评分：`PUT .../messages/:messageId/feedback`（up/down，UI 锁定后不可改）
- 引导问答来自绑定知识库的 `suggestedQuestions`

---

## 7. curl 快速验证（冒烟测试）

> ⚠️ Git Bash 里 curl 发**中文 body 会乱码**（GBK/UTF-8 转换问题），中文内容务必用
> `--data-binary @file.json` 从 UTF-8 文件读，不要 `-d '中文...'` 内联。纯 ASCII 不受影响。
> 同理 `-F "file=@/f/tmp/x.txt"` 的 MSYS 路径偶发解析失败，文件先拷到当前目录再用相对路径。

```bash
# 登录拿 token
TOKEN=$(curl -s -X POST http://127.0.0.1:8888/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@aipivot.dev","password":"Aipivot@2026"}' \
  | python -c "import sys,json;print(json.load(sys.stdin)['data']['token'])")
A="Authorization: Bearer $TOKEN"

# 会话列表（page/pageSize 必传！）
curl -s -H "$A" "http://127.0.0.1:8888/api/v1/conversations?page=1&pageSize=10"

# 建会话 + 流式对话（注意 :id 是数字 ID 不是 UUID）
CID=$(curl -s -X POST -H "$A" -H "Content-Type: application/json" \
  -d '{"title":"测试"}' http://127.0.0.1:8888/api/v1/conversations \
  | python -c "import sys,json;print(json.load(sys.stdin)['data']['id'])")
curl -s -N -X POST -H "$A" -H "Content-Type: application/json" \
  -d '{"content":"hello"}' \
  "http://127.0.0.1:8888/api/v1/conversations/$CID/messages/stream"
```

---

## 8. 常见错误对照表

| 现象 | 原因 | 解法 |
|---|---|---|
| `code=1002 LLM 网关不可用` | `ARK_API_KEY` 没加载 | 启动前 `set -a && . ./.env && set +a` |
| `field "page" is not set` | 列表端点分页参数必填 | URL 加 `?page=1&pageSize=10` |
| `strconv.ParseInt ... invalid syntax` | 路径参数传了 UUID | 全部路径参数（`:id`/`:convId`/`:kbId`）都是**数字自增 ID**，UUID 只用于 Widget sessionToken |
| `invalid input syntax for type json (SQLSTATE 22P02)` | JSONB 列写入空字符串 | Repo/Assembler 创建 PO 时给 JSONB 字段显式赋 `"{}"` / `"[]"`（2026-08 已修 KB/documents 路径） |
| 文档 `status=failed`，error 含 `InvalidEndpointOrModel.NotFound` | Embedding 模型名不在 Ark | 见 §5.2 已知限制 |
| Widget 接口 `403 origin not allowed` | Origin 白名单严格匹配失败 | 检查 scheme+host+**port** 完全一致；`http://localhost:5174` ≠ `http://127.0.0.1:5174` |
| `需要管理员权限`（403） | member 角色访问 admin 端点 | SQL 提权，见 §3 |
| go build 报错写不进临时目录 | C 盘满 | `export TMP=F:/tmp TMPDIR=F:/tmp GOTMPDIR=F:/tmp` |
| `make api` 后编译 redeclared | 重命名 handler 生成了重复文件 | 删掉旧 `xxxLogic.go` / `xxxHandler.go` 再生成 |
| 查库中文乱码 | psql 客户端编码 | 数据本身是好的，用 `\encoding utf8` 或前端看 |

---

## 9. 停止服务

```bash
# 后端（找到 PID）
tasklist | findstr aipivot
taskkill //PID <pid> //F

# 前端 / widget dev server：Ctrl+C（或同样 taskkill node.exe 前先确认 PID）

# 依赖容器（一般不用停）
docker compose -f deploy/docker-compose.yml stop
```

---

## 10. 当前环境的既有数据（本地联调产物）

| 数据 | 值 |
|---|---|
| 演示知识库 | id=2「演示知识库」，含 3 条引导问答 |
| public key | id=1「Widget Local Demo」，绑定 KB 2，白名单 localhost:5174 / 127.0.0.1:5174 |
| 文档 | id=1 faq.txt（failed，卡在 embedding 模型） |
| widget 会话 | conversation id=12（visitor-doc-demo） |

不需要可直接删：管理台 UI 操作，或 `DELETE /api/v1/knowledge-bases/2`、`PUT /api/v1/api-keys/1/revoke`。
