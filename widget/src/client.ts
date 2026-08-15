import type { StreamCallbacks, WidgetConfig } from './types';
import { parseSSEStream } from './utils/sse';
import { withRetry } from './utils/retry';

/**
 * WidgetClient — 封装与后端 Widget Open API 的所有交互。
 * 所有方法均通过 public key（X-API-Key）鉴权，由后端 ApiKeyMiddleware 校验域名白名单。
 */
export class WidgetClient {
  constructor(private config: WidgetConfig) {}

  private get headers(): Record<string, string> {
    return {
      'Content-Type': 'application/json',
      'X-API-Key': this.config.publicKey,
    };
  }

  private url(path: string): string {
    return `${this.config.baseUrl}${path}`;
  }

  /** 创建访客会话，返回 sessionToken（= conversation UUID） */
  async createSession(visitorId: string, title?: string): Promise<{
    sessionToken: string;
    conversationId: number;
    visitorId: string;
    createdAt: number;
    suggestedQuestions?: string[];
  }> {
    return withRetry(async () => {
      const res = await fetch(this.url('/api/v1/open/widget/sessions'), {
        method: 'POST',
        headers: this.headers,
        body: JSON.stringify({ visitorId, title }),
      });
      const json = await res.json();
      if (!res.ok || json.code !== 0) {
        throw new Error(json.msg || `创建会话失败 (HTTP ${res.status})`);
      }
      return json.data;
    });
  }

  /** 流式发送消息（SSE），通过回调推送增量 token */
  async sendMessageStream(
    sessionToken: string,
    content: string,
    callbacks: StreamCallbacks,
    signal?: AbortSignal
  ): Promise<void> {
    const res = await fetch(
      this.url(`/api/v1/open/widget/sessions/${encodeURIComponent(sessionToken)}/messages/stream`),
      {
        method: 'POST',
        headers: this.headers,
        body: JSON.stringify({ content, contentType: 'text' }),
        signal,
      }
    );
    await parseSSEStream(res, callbacks);
  }

  /** 拉取历史消息 */
  async listMessages(sessionToken: string, page = 1, pageSize = 50): Promise<
    Array<{
      uuid: string;
      role: string;
      content: string;
      contentType: string;
      tokenCount?: number;
      model?: string;
      sources?: string[];
      rating?: 'up' | 'down' | '';
      ratingFeedback?: string;
      createdAt: number;
    }>
  > {
    return withRetry(async () => {
      const url = this.url(
        `/api/v1/open/widget/sessions/${encodeURIComponent(sessionToken)}/messages?page=${page}&pageSize=${pageSize}`
      );
      const res = await fetch(url, { headers: this.headers });
      const json = await res.json();
      if (!res.ok || json.code !== 0) {
        throw new Error(json.msg || `拉取历史失败 (HTTP ${res.status})`);
      }
      return json.data ?? [];
    });
  }

  /**
   * 提交消息满意度评分。
   * 第一波为锁定语义：前端调用一次即锁定 UI；后端允许覆盖但前端不会发起第二次。
   * 失败时抛出 Error，调用方负责本地状态回滚。
   */
  async rateMessage(
    sessionToken: string,
    messageId: string,
    rating: 'up' | 'down',
    feedback?: string
  ): Promise<void> {
    const res = await fetch(
      this.url(
        `/api/v1/open/widget/sessions/${encodeURIComponent(sessionToken)}/messages/${encodeURIComponent(messageId)}/feedback`
      ),
      {
        method: 'PUT',
        headers: this.headers,
        body: JSON.stringify({ rating, feedback }),
      }
    );
    const json = await res.json().catch(() => ({}));
    if (!res.ok || json.code !== 0) {
      throw new Error(json.msg || `评分失败 (HTTP ${res.status})`);
    }
  }
}
