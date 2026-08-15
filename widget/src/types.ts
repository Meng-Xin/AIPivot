// Widget SDK 类型定义

/** Widget 初始化配置 */
export interface WidgetConfig {
  /** public key（pk_ 前缀，可在管理后台创建） */
  publicKey: string;
  /** AIPivot 后端基地址，如 https://api.example.com */
  baseUrl: string;
  /** 聊天面板标题（默认 "在线客服"） */
  title?: string;
  /** 欢迎语（默认 "您好，有什么可以帮您？"） */
  welcome?: string;
  /** 悬浮按钮距离右下角的偏移（px） */
  offset?: { bottom?: number; right?: number };
  /** 主题色（HEX/RGB），覆盖默认 #4f46e5 */
  theme?: { primary?: string };
  /** 持久化存储类型（默认 localStorage） */
  storage?: 'localStorage' | 'sessionStorage';
  /** 输入框占位符（默认 "请输入消息..."） */
  placeholder?: string;
}

/** Widget 消息（前端展示用） */
export interface WidgetMessage {
  uuid: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  contentType: string;
  tokenCount?: number;
  model?: string;
  sources?: string[];
  /** 访客评分：up / down / undefined=未评分 */
  rating?: 'up' | 'down';
  /** 负评文字反馈（rating=down 时可能有值） */
  ratingFeedback?: string;
  createdAt: number;
  /** 是否处于流式生成中（仅 assistant 临时消息） */
  streaming?: boolean;
}

/** SSE 流式回调 */
export interface StreamCallbacks {
  onStart?: (data: { messageId: string; conversationId: number }) => void;
  onDelta?: (data: { content: string }) => void;
  onEnd?: (data: {
    messageId: string;
    model: string;
    tokenCount: number;
    latencyMs: number;
    sources: string[];
  }) => void;
  onError?: (data: { code: number; msg: string }) => void;
}

/** SDK 暴露给宿主页面的句柄 */
export interface WidgetHandle {
  /** 打开聊天面板 */
  open(): void;
  /** 关闭聊天面板 */
  close(): void;
  /** 切换面板开合 */
  toggle(): void;
  /** 销毁 Widget（移除 DOM 与事件监听） */
  destroy(): void;
  /** 主动发送一条消息 */
  send(text: string): Promise<void>;
}
