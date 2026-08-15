import type { WidgetConfig } from './types';
import { generateVisitorId } from './utils/uuid';

const KEY_VISITOR = 'apw_visitor_id';
const KEY_SESSION = 'apw_session_token';

/** 持久化适配器，按 config.storage 选择 localStorage 或 sessionStorage。 */
function getStorage(config: WidgetConfig): Storage | null {
  const type = config.storage ?? 'localStorage';
  try {
    return type === 'sessionStorage' ? window.sessionStorage : window.localStorage;
  } catch {
    // 隐私模式下 Storage 可能不可用，降级为内存对象
    return null;
  }
}

/** 内存兜底存储（localStorage 被禁用时使用） */
const memoryStore = new Map<string, string>();

function read(storage: Storage | null, key: string): string | null {
  if (storage) {
    try {
      return storage.getItem(key);
    } catch {
      /* ignore */
    }
  }
  return memoryStore.get(key) ?? null;
}

function write(storage: Storage | null, key: string, value: string): void {
  if (storage) {
    try {
      storage.setItem(key, value);
      return;
    } catch {
      /* ignore */
    }
  }
  memoryStore.set(key, value);
}

/** 读取或生成 visitorId（跨会话保持，用于标识同一访客） */
export function getOrCreateVisitorId(config: WidgetConfig): string {
  const storage = getStorage(config);
  const existing = read(storage, KEY_VISITOR);
  if (existing) return existing;
  const id = generateVisitorId();
  write(storage, KEY_VISITOR, id);
  return id;
}

/** 缓存 sessionToken（= conversation UUID），刷新页面可恢复会话 */
export function saveSessionToken(config: WidgetConfig, token: string): void {
  write(getStorage(config), KEY_SESSION, token);
}

export function loadSessionToken(config: WidgetConfig): string | null {
  return read(getStorage(config), KEY_SESSION);
}

export function clearSessionToken(config: WidgetConfig): void {
  const storage = getStorage(config);
  if (storage) {
    try {
      storage.removeItem(KEY_SESSION);
      return;
    } catch {
      /* ignore */
    }
  }
  memoryStore.delete(KEY_SESSION);
}
