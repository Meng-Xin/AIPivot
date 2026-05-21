// ==================== API Types ====================

export interface ApiResponse<T = null> {
  code: number;
  msg: string;
  timestamp: number;
  data: T;
}

export interface LoginData {
  token: string;
  expireAt: number;
  user: ShowUser;
}

export interface ShowUser {
  id: number;
  uuid: string;
  tenantId: number;
  email: string;
  nickName: string;
  role: string;
  status: string;
  lastLogin: number;
  createdAt: number;
}

export interface ShowConversation {
  id: number;
  uuid: string;
  knowledgeBaseId?: number;
  title: string;
  model?: string;
  status: string;
  channel: string;
  messageCount: number;
  createdAt: number;
  updatedAt: number;
}

export interface ShowModel {
  id: string;
  name: string;
  type: "chat" | "embedding";
  provider?: string;
  maxTokens?: number;
  isDefault: boolean;
}

export interface ModelListData {
  chatModels: ShowModel[];
  embeddingModels: ShowModel[];
}

export interface ShowMessage {
  id: number;
  uuid: string;
  role: string;
  content: string;
  contentType: string;
  tokenCount?: number;
  model?: string;
  latencyMs?: number;
  sources?: string[];
  createdAt: number;
}

export interface ListData<T> {
  total: number;
  list: T[];
}

export interface ShowKnowledgeBase {
  id: number;
  uuid: string;
  name: string;
  description: string;
  documentCount: number;
  chunkCount: number;
  status: string;
  createdAt: number;
  updatedAt: number;
}

export interface ShowDocument {
  id: number;
  uuid: string;
  name: string;
  contentType: string;
  fileSize: number;
  chunkCount: number;
  status: string;
  errorMsg?: string;
  createdAt: number;
  updatedAt: number;
}

// ==================== SSE Event Types ====================

export interface SSEMessageStart {
  messageId: string;
  conversationId: number;
}

export interface SSEDelta {
  content: string;
}

export interface SSEMessageEnd {
  messageId: string;
  model: string;
  tokenCount: number;
  latencyMs: number;
  sources?: string[];
}

export interface SSEError {
  code: number;
  msg: string;
}

// ==================== HTTP Client ====================

const BASE = "";

function getHeaders(token?: string): HeadersInit {
  const h: Record<string, string> = { "Content-Type": "application/json" };
  if (token) h["Authorization"] = `Bearer ${token}`;
  return h;
}

async function request<T>(
  method: string,
  path: string,
  token?: string,
  body?: unknown
): Promise<ApiResponse<T>> {
  const res = await fetch(`${BASE}${path}`, {
    method,
    headers: getHeaders(token),
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`HTTP ${res.status}: ${text}`);
  }
  return res.json();
}

// ==================== Auth API ====================

export async function login(email: string, password: string) {
  return request<LoginData>("POST", "/api/v1/auth/login", undefined, {
    email,
    password,
  });
}

export async function register(
  nickName: string,
  email: string,
  password: string
) {
  return request<null>("POST", "/api/v1/auth/register", undefined, {
    nick_name: nickName,
    email,
    password,
  });
}

// ==================== Conversation API ====================

export async function createConversation(
  token: string,
  opts?: { knowledgeBaseId?: number; title?: string; model?: string }
) {
  return request<ShowConversation>("POST", "/api/v1/conversations", token, opts);
}

// ==================== Models API ====================

export async function listModels(token: string) {
  return request<ModelListData>("GET", "/api/v1/models", token);
}

export async function listConversations(
  token: string,
  page = 1,
  pageSize = 20,
  status?: string
) {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
  });
  if (status) params.set("status", status);
  return request<ListData<ShowConversation>>(
    "GET",
    `/api/v1/conversations?${params}`,
    token
  );
}

export async function getConversation(token: string, id: number) {
  return request<ShowConversation>(
    "GET",
    `/api/v1/conversations/${id}`,
    token
  );
}

export async function closeConversation(token: string, id: number) {
  return request<null>("PUT", `/api/v1/conversations/${id}/close`, token);
}

// ==================== Message API ====================

export async function listMessages(
  token: string,
  convId: number,
  page = 1,
  pageSize = 50
) {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
  });
  return request<ListData<ShowMessage>>(
    "GET",
    `/api/v1/conversations/${convId}/messages?${params}`,
    token
  );
}

export async function sendMessage(
  token: string,
  convId: number,
  content: string
) {
  return request<ShowMessage>(
    "POST",
    `/api/v1/conversations/${convId}/messages`,
    token,
    { content, contentType: "text" }
  );
}

// ==================== SSE Stream ====================

export type StreamCallback = {
  onStart?: (data: SSEMessageStart) => void;
  onDelta?: (data: SSEDelta) => void;
  onEnd?: (data: SSEMessageEnd) => void;
  onError?: (data: SSEError) => void;
};

/**
 * sendMessageStream 通过 POST 请求发送消息并接收 SSE 流式响应。
 * 因为 EventSource 只支持 GET，这里用 fetch + ReadableStream 解析 SSE。
 */
export async function sendMessageStream(
  token: string,
  convId: number,
  content: string,
  callbacks: StreamCallback,
  signal?: AbortSignal
): Promise<void> {
  const res = await fetch(
    `${BASE}/api/v1/conversations/${convId}/messages/stream`,
    {
      method: "POST",
      headers: getHeaders(token),
      body: JSON.stringify({ content, contentType: "text" }),
      signal,
    }
  );

  if (!res.ok || !res.body) {
    const text = await res.text();
    callbacks.onError?.({ code: res.status, msg: text });
    return;
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });

    // 解析 SSE 事件：以双换行分隔
    const parts = buffer.split("\n\n");
    buffer = parts.pop() ?? "";

    for (const part of parts) {
      const lines = part.trim().split("\n");
      let eventName = "";
      let eventData = "";

      for (const line of lines) {
        if (line.startsWith("event: ")) {
          eventName = line.slice(7);
        } else if (line.startsWith("data: ")) {
          eventData = line.slice(6);
        }
      }

      if (!eventName || !eventData) continue;

      if (eventName === "done") break;

      try {
        const parsed = JSON.parse(eventData);
        switch (eventName) {
          case "message_start":
            callbacks.onStart?.(parsed as SSEMessageStart);
            break;
          case "delta":
            callbacks.onDelta?.(parsed as SSEDelta);
            break;
          case "message_end":
            callbacks.onEnd?.(parsed as SSEMessageEnd);
            break;
          case "error":
            callbacks.onError?.(parsed as SSEError);
            break;
        }
      } catch {
        // 忽略非 JSON 数据（如 [DONE]）
      }
    }
  }
}

// ==================== Knowledge Base API ====================

export async function listKnowledgeBases(
  token: string,
  page = 1,
  pageSize = 50,
  name?: string
) {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
  });
  if (name) params.set("name", name);
  return request<ListData<ShowKnowledgeBase>>(
    "GET",
    `/api/v1/knowledge-bases?${params}`,
    token
  );
}

export async function createKnowledgeBase(
  token: string,
  data: { name: string; description?: string; model?: string }
) {
  return request<ShowKnowledgeBase>(
    "POST",
    "/api/v1/knowledge-bases",
    token,
    data
  );
}

export async function getKnowledgeBase(token: string, id: number) {
  return request<ShowKnowledgeBase>(
    "GET",
    `/api/v1/knowledge-bases/${id}`,
    token
  );
}

export async function updateKnowledgeBase(
  token: string,
  id: number,
  data: { name?: string; description?: string }
) {
  return request<null>("PUT", `/api/v1/knowledge-bases/${id}`, token, data);
}

export async function deleteKnowledgeBase(token: string, id: number) {
  return request<null>("DELETE", `/api/v1/knowledge-bases/${id}`, token);
}

// ==================== Document API ====================

export async function listDocuments(
  token: string,
  kbId: number,
  page = 1,
  pageSize = 20,
  status?: string
) {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
  });
  if (status) params.set("status", status);
  return request<ListData<ShowDocument>>(
    "GET",
    `/api/v1/knowledge-bases/${kbId}/documents?${params}`,
    token
  );
}

export async function uploadDocument(
  token: string,
  kbId: number,
  file: File
): Promise<ApiResponse<ShowDocument>> {
  const form = new FormData();
  form.append("file", file);
  const res = await fetch(
    `${BASE}/api/v1/knowledge-bases/${kbId}/documents`,
    {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: form,
    }
  );
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`HTTP ${res.status}: ${text}`);
  }
  return res.json();
}

export async function deleteDocument(
  token: string,
  kbId: number,
  docId: number
) {
  return request<null>(
    "DELETE",
    `/api/v1/knowledge-bases/${kbId}/documents/${docId}`,
    token
  );
}

// ==================== Analytics API ====================

export interface ConvStatusStat {
  status: string;
  count: number;
}

export interface ChannelStat {
  channel: string;
  count: number;
}

export interface ModelUsageStat {
  model: string;
  messageCount: number;
  tokenCount: number;
  estimatedCost: number;
}

export interface AnalyticsOverviewData {
  totalConversations: number;
  activeConversations: number;
  totalMessages: number;
  totalTokens: number;
  estimatedCost: number;
  aiResolveRate: number;
  escalationRate: number;
  byStatus: ConvStatusStat[];
  byChannel: ChannelStat[];
  modelUsage: ModelUsageStat[];
}

export interface DailyStatPoint {
  date: string;
  conversationCount: number;
  messageCount: number;
  tokenCount: number;
}

export async function getAnalyticsOverview(token: string) {
  return request<AnalyticsOverviewData>(
    "GET",
    "/api/v1/analytics/overview",
    token
  );
}

export async function getAnalyticsDaily(token: string, days = 7) {
  return request<DailyStatPoint[]>(
    "GET",
    `/api/v1/analytics/daily?days=${days}`,
    token
  );
}
