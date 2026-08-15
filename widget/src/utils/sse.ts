import type { StreamCallbacks } from '../types';

/**
 * 通过 fetch + ReadableStream 解析 SSE 流（EventSource 不支持 POST + 自定义 header）。
 * 提炼自 web/src/lib/api.ts 第 260-339 行。
 */
export async function parseSSEStream(
  res: Response,
  callbacks: StreamCallbacks
): Promise<void> {
  if (!res.ok || !res.body) {
    const text = await res.text().catch(() => '');
    let code = res.status;
    let msg = text;
    try {
      const j = JSON.parse(text);
      code = j.code ?? code;
      msg = j.msg ?? text;
    } catch {
      /* keep raw */
    }
    callbacks.onError?.({ code, msg: msg || `HTTP ${res.status}` });
    return;
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });

      // SSE 事件以双换行分隔
      const parts = buffer.split('\n\n');
      buffer = parts.pop() ?? '';

      for (const part of parts) {
        const lines = part.trim().split('\n');
        let eventName = '';
        let eventData = '';

        for (const line of lines) {
          if (line.startsWith('event: ')) {
            eventName = line.slice(7).trim();
          } else if (line.startsWith('data: ')) {
            eventData = line.slice(6);
          }
        }

        if (!eventName || !eventData) continue;
        if (eventName === 'done') return;

        try {
          const parsed = JSON.parse(eventData);
          switch (eventName) {
            case 'message_start':
              callbacks.onStart?.(parsed);
              break;
            case 'delta':
              callbacks.onDelta?.(parsed);
              break;
            case 'message_end':
              callbacks.onEnd?.(parsed);
              break;
            case 'error':
              callbacks.onError?.(parsed);
              break;
          }
        } catch {
          // 忽略非 JSON 数据
        }
      }
    }
  } finally {
    try {
      reader.releaseLock();
    } catch {
      /* ignore */
    }
  }
}
