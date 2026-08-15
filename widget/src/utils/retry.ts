/**
 * 指数退避重试工具。
 * 用于 Widget 网络抖动场景（如创建会话/拉取历史失败），流式接口不重试（避免重复消耗 token）。
 */

export interface RetryOptions {
  /** 最大重试次数（默认 2，总尝试 = 1 + retries） */
  retries?: number;
  /** 初始退避毫秒（默认 500ms） */
  baseDelay?: number;
  /** 最大退避毫秒（默认 4000ms） */
  maxDelay?: number;
  /** 判断错误是否可重试（默认所有错误都重试） */
  shouldRetry?: (err: unknown) => boolean;
}

const DEFAULT_OPTS: Required<Omit<RetryOptions, 'shouldRetry'>> = {
  retries: 2,
  baseDelay: 500,
  maxDelay: 4000,
};

export async function withRetry<T>(
  fn: () => Promise<T>,
  opts: RetryOptions = {}
): Promise<T> {
  const config = { ...DEFAULT_OPTS, ...opts };
  const shouldRetry = opts.shouldRetry ?? (() => true);

  let lastErr: unknown = null;
  for (let attempt = 0; attempt <= config.retries; attempt++) {
    try {
      return await fn();
    } catch (err) {
      lastErr = err;
      if (attempt >= config.retries || !shouldRetry(err)) {
        break;
      }
      // 指数退避：baseDelay * 2^attempt，附加 10% 抖动避免雪崩
      const delay = Math.min(config.baseDelay * Math.pow(2, attempt), config.maxDelay);
      const jitter = delay * (0.9 + Math.random() * 0.2);
      await new Promise((r) => setTimeout(r, jitter));
    }
  }
  throw lastErr instanceof Error ? lastErr : new Error('retry exhausted');
}
