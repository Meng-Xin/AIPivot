# AIPivot Chat Widget SDK

可嵌入到任意网站的悬浮聊天组件，通过 public key（`pk_` 前缀）安全接入 AIPivot 后端，访客对话持久化到服务端 `conversations` / `messages` 表。

## 特性

- **单文件 IIFE 输出**：`dist/aipivot-widget.js`（gzip ≈ 16KB），客户网站 `<script>` 一行接入
- **Shadow DOM 隔离**：`mode: 'closed'`，样式与宿主页面互不污染
- **SSE 流式**：基于 fetch + ReadableStream 解析（EventSource 不支持 POST + 自定义 header）
- **会话持久化**：访客 `sessionToken`（= conversation UUID）存 localStorage，刷新页面可恢复历史
- **XSS 防护**：LLM 返回内容经过 `escapeHtml` + 受控 markdown 替换
- **Preact 内核**：~3KB gzipped，API 与 React 兼容
- **主题色可定制**：通过 `--apw-primary` CSS 变量覆盖

## 接入

### 1. 创建 public key

在管理后台创建 API Key，指定：
- `keyType: "public"`（生成 `pk_` 前缀密钥）
- `allowedOrigins: ["https://your-website.com"]`（严格匹配域名白名单）
- `knowledgeBaseId: <kb-id>`（绑定知识库，限制 RAG 检索范围）

### 2. 嵌入到网站

```html
<script src="https://cdn.example.com/aipivot-widget.js"></script>
<script>
  AIPivotWidget.init({
    publicKey: 'pk_xxxxxxxxxxxx',
    baseUrl: 'https://api.example.com',
    title: '在线客服',
    theme: { primary: '#4f46e5' },
  });
</script>
```

完整配置项见 [types.ts](./src/types.ts) 中的 `WidgetConfig`。

## 句柄 API

`init()` 返回的句柄支持编程控制：

```js
const handle = AIPivotWidget.init({ /* ... */ });
handle.open();           // 打开面板
handle.close();          // 关闭面板
handle.toggle();         // 切换开合
handle.send('你好');     // 主动发送一条消息
handle.destroy();        // 销毁 Widget（移除 DOM）
```

## 本地开发

```bash
# 在项目根目录
make widget-install    # 安装依赖
make widget-dev        # 启动开发服务器（5174）
make widget-build      # 构建生产包到 dist/aipivot-widget.js
```

开发模式下，编辑 `src/dev-entry.tsx` 填入你的 `publicKey` 与后端 `baseUrl`，浏览器访问 `http://127.0.0.1:5174` 即可看到悬浮按钮。

## 后端端点

SDK 调用以下 Widget Open API 端点（由后端 `ApiKeyMiddleware` 通过 `X-API-Key` 头鉴权 + Origin 白名单校验）：

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/v1/open/widget/sessions` | 创建访客会话 |
| GET  | `/api/v1/open/widget/sessions/:sessionToken/messages` | 拉取历史消息 |
| POST | `/api/v1/open/widget/sessions/:sessionToken/messages/stream` | 流式发送消息（SSE） |

## 安全注意事项

- **public key 域名白名单是唯一防线**：必须在创建时正确配置 `allowedOrigins`，后端严格匹配（不支持通配符，空数组 fail-closed）
- **sessionToken 不放 URL query**：仅出现在 path 与后端校验 `conv.ExternalUserID == visitorId`
- **滑窗限流兜底**：后端对每个访客默认每分钟 10 次请求，超出返回 `code:429`，可通过 `RateLimitConf.WidgetVisitorLimit` 调整
- **Nginx/CDN 配置**：流式响应需关闭缓冲（`proxy_buffering off;`），后端已设置 `X-Accel-Buffering: no`
