import { useEffect, useRef, useState, useCallback } from "react";
import {
  Bot,
  Send,
  Plus,
  MessageSquare,
  LogOut,
  Loader2,
  X,
  BookOpen,
  Clock,
  ChevronDown,
} from "lucide-react";
import { useAuthStore } from "../store/auth";
import { useChatStore } from "../store/chat";
import {
  listConversations,
  listMessages,
  createConversation,
  sendMessageStream,
  listKnowledgeBases,
} from "../lib/api";
import type {
  ShowConversation,
  ShowMessage,
  ShowKnowledgeBase,
} from "../lib/api";

// ==================== ChatPage ====================

export default function ChatPage() {
  const token = useAuthStore((s) => s.token)!;
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);

  const {
    conversations,
    setConversations,
    activeConvId,
    setActiveConvId,
    messages,
    setMessages,
    addMessage,
    streamingContent,
    isStreaming,
    onStreamStart,
    onStreamDelta,
    onStreamEnd,
    onStreamReset,
    commitStreamMessage,
    incrementMessageCount,
  } = useChatStore();

  const [input, setInput] = useState("");
  const [loadingConvs, setLoadingConvs] = useState(false);
  const [loadingMsgs, setLoadingMsgs] = useState(false);
  const [showNewChat, setShowNewChat] = useState(false);
  const [knowledgeBases, setKnowledgeBases] = useState<ShowKnowledgeBase[]>([]);
  const [selectedKbId, setSelectedKbId] = useState<number | undefined>();
  const [newTitle, setNewTitle] = useState("");

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const abortRef = useRef<AbortController | null>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  // 滚动到底部
  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, []);

  useEffect(scrollToBottom, [messages, streamingContent, scrollToBottom]);

  // 加载会话列表
  useEffect(() => {
    const load = async () => {
      setLoadingConvs(true);
      try {
        const res = await listConversations(token);
        if (res.code === 0) setConversations(res.data.list);
      } finally {
        setLoadingConvs(false);
      }
    };
    load();
  }, [token, setConversations]);

  // 切换会话时加载消息
  useEffect(() => {
    if (!activeConvId) {
      setMessages([]);
      return;
    }
    const load = async () => {
      setLoadingMsgs(true);
      try {
        const res = await listMessages(token, activeConvId);
        if (res.code === 0) setMessages(res.data.list);
      } finally {
        setLoadingMsgs(false);
      }
    };
    load();
  }, [activeConvId, token, setMessages]);

  // 创建新会话
  const handleCreateConversation = async () => {
    try {
      const res = await createConversation(token, {
        knowledgeBaseId: selectedKbId,
        title: newTitle || undefined,
      });
      if (res.code === 0) {
        setConversations([res.data, ...conversations]);
        setActiveConvId(res.data.id);
        setShowNewChat(false);
        setNewTitle("");
        setSelectedKbId(undefined);
      }
    } catch (err) {
      console.error("创建会话失败:", err);
    }
  };

  // 加载知识库列表（创建会话弹窗时）
  useEffect(() => {
    if (!showNewChat) return;
    listKnowledgeBases(token).then((res) => {
      if (res.code === 0) setKnowledgeBases(res.data.list);
    });
  }, [showNewChat, token]);

  // 发送消息（SSE 流式）
  const handleSend = async () => {
    const content = input.trim();
    if (!content || !activeConvId || isStreaming) return;

    setInput("");

    // 立即展示用户消息
    const userMsg: ShowMessage = {
      id: Date.now(),
      uuid: crypto.randomUUID(),
      role: "user",
      content,
      contentType: "text",
      createdAt: Date.now(),
    };
    addMessage(userMsg);
    incrementMessageCount(activeConvId);

    // 启动 SSE 流
    abortRef.current = new AbortController();
    try {
      await sendMessageStream(
        token,
        activeConvId,
        content,
        {
          onStart: onStreamStart,
          onDelta: onStreamDelta,
          onEnd: (data) => {
            onStreamEnd(data);
            commitStreamMessage(data);
            incrementMessageCount(activeConvId);
          },
          onError: (data) => {
            console.error("SSE error:", data);
            onStreamReset();
          },
        },
        abortRef.current.signal
      );
    } catch (err) {
      if ((err as Error).name !== "AbortError") {
        console.error("Stream failed:", err);
      }
      onStreamReset();
    }
  };

  // Ctrl+Enter 或 Enter 发送
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  // 格式化时间
  const formatTime = (ts: number) => {
    const d = new Date(ts);
    const now = new Date();
    if (d.toDateString() === now.toDateString()) {
      return d.toLocaleTimeString("zh-CN", {
        hour: "2-digit",
        minute: "2-digit",
      });
    }
    return d.toLocaleDateString("zh-CN", {
      month: "short",
      day: "numeric",
    });
  };

  return (
    <div className="flex h-full bg-slate-50">
      {/* ========== 侧边栏 ========== */}
      <aside className="flex w-72 flex-col border-r border-slate-200 bg-white">
        {/* 顶部 */}
        <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-600">
              <Bot className="h-4 w-4 text-white" />
            </div>
            <span className="text-sm font-semibold text-slate-800">
              AIPivot
            </span>
          </div>
          <button
            onClick={() => setShowNewChat(true)}
            className="rounded-lg p-1.5 text-slate-400 transition hover:bg-indigo-50 hover:text-indigo-600"
            title="新建会话"
          >
            <Plus className="h-5 w-5" />
          </button>
        </div>

        {/* 会话列表 */}
        <div className="flex-1 overflow-y-auto scrollbar-thin">
          {loadingConvs ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-5 w-5 animate-spin text-slate-300" />
            </div>
          ) : conversations.length === 0 ? (
            <div className="px-4 py-12 text-center">
              <MessageSquare className="mx-auto mb-2 h-8 w-8 text-slate-200" />
              <p className="text-xs text-slate-400">暂无会话</p>
              <button
                onClick={() => setShowNewChat(true)}
                className="mt-3 text-xs font-medium text-indigo-600 hover:underline"
              >
                创建第一个会话
              </button>
            </div>
          ) : (
            <div className="p-2">
              {conversations.map((conv) => (
                <ConversationItem
                  key={conv.id}
                  conv={conv}
                  active={conv.id === activeConvId}
                  onClick={() => setActiveConvId(conv.id)}
                  formatTime={formatTime}
                />
              ))}
            </div>
          )}
        </div>

        {/* 底部用户信息 */}
        <div className="border-t border-slate-100 px-4 py-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 overflow-hidden">
              <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-indigo-100 text-xs font-medium text-indigo-600">
                {user?.nickName?.charAt(0) ?? "U"}
              </div>
              <div className="min-w-0">
                <p className="truncate text-sm font-medium text-slate-700">
                  {user?.nickName}
                </p>
                <p className="truncate text-xs text-slate-400">
                  {user?.email}
                </p>
              </div>
            </div>
            <button
              onClick={logout}
              className="rounded-lg p-1.5 text-slate-400 transition hover:bg-red-50 hover:text-red-500"
              title="退出登录"
            >
              <LogOut className="h-4 w-4" />
            </button>
          </div>
        </div>
      </aside>

      {/* ========== 主聊天区 ========== */}
      <main className="flex flex-1 flex-col">
        {!activeConvId ? (
          <EmptyState onNew={() => setShowNewChat(true)} />
        ) : (
          <>
            {/* 头部 */}
            <ChatHeader
              conv={conversations.find((c) => c.id === activeConvId)}
            />

            {/* 消息列表 */}
            <div className="flex-1 overflow-y-auto px-4 py-6 scrollbar-thin">
              {loadingMsgs ? (
                <div className="flex items-center justify-center py-20">
                  <Loader2 className="h-6 w-6 animate-spin text-slate-300" />
                </div>
              ) : messages.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-20 text-slate-400">
                  <Bot className="mb-3 h-12 w-12 text-slate-200" />
                  <p className="text-sm">发送消息开始对话</p>
                </div>
              ) : (
                <div className="mx-auto max-w-3xl space-y-4">
                  {messages.map((msg) => (
                    <MessageBubble
                      key={msg.uuid}
                      msg={msg}
                      formatTime={formatTime}
                    />
                  ))}
                  {/* 流式消息 */}
                  {isStreaming && (
                    <StreamingBubble content={streamingContent} />
                  )}
                  <div ref={messagesEndRef} />
                </div>
              )}
            </div>

            {/* 输入框 */}
            <div className="border-t border-slate-200 bg-white px-4 py-3">
              <div className="mx-auto flex max-w-3xl items-end gap-3">
                <textarea
                  ref={inputRef}
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="输入消息… (Enter 发送, Shift+Enter 换行)"
                  rows={1}
                  disabled={isStreaming}
                  className="max-h-32 flex-1 resize-none rounded-xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-sm text-slate-800 outline-none transition placeholder:text-slate-400 focus:border-indigo-400 focus:bg-white focus:ring-2 focus:ring-indigo-100 disabled:opacity-60"
                  style={{
                    height: "auto",
                    minHeight: "40px",
                    overflow: "hidden",
                  }}
                  onInput={(e) => {
                    const t = e.target as HTMLTextAreaElement;
                    t.style.height = "auto";
                    t.style.height = Math.min(t.scrollHeight, 128) + "px";
                  }}
                />
                <button
                  onClick={handleSend}
                  disabled={!input.trim() || isStreaming}
                  className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-indigo-600 text-white transition hover:bg-indigo-500 disabled:opacity-40 disabled:hover:bg-indigo-600"
                >
                  {isStreaming ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Send className="h-4 w-4" />
                  )}
                </button>
              </div>
            </div>
          </>
        )}
      </main>

      {/* ========== 新建会话弹窗 ========== */}
      {showNewChat && (
        <NewChatModal
          knowledgeBases={knowledgeBases}
          selectedKbId={selectedKbId}
          setSelectedKbId={setSelectedKbId}
          newTitle={newTitle}
          setNewTitle={setNewTitle}
          onClose={() => setShowNewChat(false)}
          onCreate={handleCreateConversation}
        />
      )}
    </div>
  );
}

// ==================== 子组件 ====================

function ConversationItem({
  conv,
  active,
  onClick,
  formatTime,
}: {
  conv: ShowConversation;
  active: boolean;
  onClick: () => void;
  formatTime: (ts: number) => string;
}) {
  return (
    <button
      onClick={onClick}
      className={`flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left transition ${
        active
          ? "bg-indigo-50 text-indigo-700"
          : "text-slate-600 hover:bg-slate-50"
      }`}
    >
      <MessageSquare
        className={`h-4 w-4 shrink-0 ${
          active ? "text-indigo-500" : "text-slate-400"
        }`}
      />
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">
          {conv.title || `会话 #${conv.id}`}
        </p>
        <div className="mt-0.5 flex items-center gap-2">
          <span className="text-xs text-slate-400">
            {conv.messageCount} 条消息
          </span>
          <span className="text-xs text-slate-300">·</span>
          <span className="text-xs text-slate-400">
            {formatTime(conv.updatedAt)}
          </span>
        </div>
      </div>
      {conv.status === "active" && (
        <span className="h-2 w-2 shrink-0 rounded-full bg-green-400" />
      )}
    </button>
  );
}

function ChatHeader({ conv }: { conv?: ShowConversation }) {
  if (!conv) return null;
  return (
    <div className="flex items-center justify-between border-b border-slate-200 bg-white px-6 py-3">
      <div>
        <h2 className="text-sm font-semibold text-slate-800">
          {conv.title || `会话 #${conv.id}`}
        </h2>
        <div className="mt-0.5 flex items-center gap-2 text-xs text-slate-400">
          {conv.knowledgeBaseId ? (
            <>
              <BookOpen className="h-3 w-3" />
              <span>知识库 #{conv.knowledgeBaseId}</span>
            </>
          ) : (
            <span>自由对话</span>
          )}
          <span>·</span>
          <span
            className={`capitalize ${
              conv.status === "active" ? "text-green-500" : "text-slate-400"
            }`}
          >
            {conv.status}
          </span>
        </div>
      </div>
    </div>
  );
}

function MessageBubble({
  msg,
  formatTime,
}: {
  msg: ShowMessage;
  formatTime: (ts: number) => string;
}) {
  const isUser = msg.role === "user";
  return (
    <div className={`flex gap-3 ${isUser ? "flex-row-reverse" : ""}`}>
      {/* 头像 */}
      <div
        className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-medium ${
          isUser
            ? "bg-indigo-100 text-indigo-600"
            : "bg-emerald-100 text-emerald-600"
        }`}
      >
        {isUser ? "你" : <Bot className="h-4 w-4" />}
      </div>

      {/* 气泡 */}
      <div className={`max-w-[75%] ${isUser ? "text-right" : ""}`}>
        <div
          className={`inline-block rounded-2xl px-4 py-2.5 text-sm leading-relaxed ${
            isUser
              ? "rounded-tr-md bg-indigo-600 text-white"
              : "rounded-tl-md bg-white text-slate-700 shadow-sm ring-1 ring-slate-100"
          }`}
        >
          <p className="whitespace-pre-wrap">{msg.content}</p>
        </div>

        {/* 元信息 */}
        <div
          className={`mt-1 flex items-center gap-2 text-xs text-slate-400 ${
            isUser ? "justify-end" : ""
          }`}
        >
          <Clock className="h-3 w-3" />
          <span>{formatTime(msg.createdAt)}</span>
          {msg.model && (
            <>
              <span>·</span>
              <span>{msg.model}</span>
            </>
          )}
          {msg.latencyMs ? (
            <>
              <span>·</span>
              <span>{(msg.latencyMs / 1000).toFixed(1)}s</span>
            </>
          ) : null}
        </div>

        {/* RAG 来源引用 */}
        {msg.sources && msg.sources.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1">
            {msg.sources.map((src, i) => (
              <span
                key={i}
                className="inline-flex items-center gap-1 rounded-md bg-slate-100 px-2 py-0.5 text-xs text-slate-500"
              >
                <BookOpen className="h-3 w-3" />
                {src}
              </span>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function StreamingBubble({ content }: { content: string }) {
  return (
    <div className="flex gap-3">
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-emerald-100 text-emerald-600">
        <Bot className="h-4 w-4" />
      </div>
      <div className="max-w-[75%]">
        <div className="inline-block rounded-2xl rounded-tl-md bg-white px-4 py-2.5 text-sm leading-relaxed text-slate-700 shadow-sm ring-1 ring-slate-100">
          <p className="whitespace-pre-wrap">
            {content || (
              <span className="text-slate-400">思考中...</span>
            )}
            {content && <span className="typing-cursor" />}
          </p>
        </div>
      </div>
    </div>
  );
}

function EmptyState({ onNew }: { onNew: () => void }) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center">
      <div className="mb-6 flex h-20 w-20 items-center justify-center rounded-2xl bg-indigo-50">
        <Bot className="h-10 w-10 text-indigo-400" />
      </div>
      <h2 className="mb-2 text-lg font-semibold text-slate-700">
        欢迎使用 AIPivot
      </h2>
      <p className="mb-6 max-w-sm text-center text-sm text-slate-400">
        选择左侧已有会话继续对话，或创建新会话开始体验 AI 智能客服。
      </p>
      <button
        onClick={onNew}
        className="flex items-center gap-2 rounded-xl bg-indigo-600 px-5 py-2.5 text-sm font-medium text-white transition hover:bg-indigo-500"
      >
        <Plus className="h-4 w-4" />
        新建会话
      </button>
    </div>
  );
}

function NewChatModal({
  knowledgeBases,
  selectedKbId,
  setSelectedKbId,
  newTitle,
  setNewTitle,
  onClose,
  onCreate,
}: {
  knowledgeBases: ShowKnowledgeBase[];
  selectedKbId: number | undefined;
  setSelectedKbId: (id: number | undefined) => void;
  newTitle: string;
  setNewTitle: (v: string) => void;
  onClose: () => void;
  onCreate: () => void;
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl">
        <div className="mb-5 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-slate-800">新建会话</h3>
          <button
            onClick={onClose}
            className="rounded-lg p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="space-y-4">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              会话标题（可选）
            </label>
            <input
              type="text"
              value={newTitle}
              onChange={(e) => setNewTitle(e.target.value)}
              placeholder="例如：产品咨询"
              className="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm outline-none transition focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100"
            />
          </div>

          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              关联知识库（可选）
            </label>
            <div className="relative">
              <select
                value={selectedKbId ?? ""}
                onChange={(e) =>
                  setSelectedKbId(
                    e.target.value ? Number(e.target.value) : undefined
                  )
                }
                className="w-full appearance-none rounded-lg border border-slate-200 bg-white px-3 py-2 pr-9 text-sm outline-none transition focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100"
              >
                <option value="">不关联（自由对话）</option>
                {knowledgeBases.map((kb) => (
                  <option key={kb.id} value={kb.id}>
                    {kb.name} ({kb.documentCount} 篇文档)
                  </option>
                ))}
              </select>
              <ChevronDown className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
            </div>
          </div>
        </div>

        <div className="mt-6 flex justify-end gap-3">
          <button
            onClick={onClose}
            className="rounded-lg px-4 py-2 text-sm font-medium text-slate-600 transition hover:bg-slate-100"
          >
            取消
          </button>
          <button
            onClick={onCreate}
            className="flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-500"
          >
            <Plus className="h-4 w-4" />
            创建
          </button>
        </div>
      </div>
    </div>
  );
}
