import { create } from "zustand";
import type {
  ShowConversation,
  ShowMessage,
  SSEMessageStart,
  SSEDelta,
  SSEMessageEnd,
} from "../lib/api";

interface ChatState {
  conversations: ShowConversation[];
  activeConvId: number | null;
  messages: ShowMessage[];
  // 流式消息状态
  streamingMessageId: string | null;
  streamingContent: string;
  isStreaming: boolean;

  setConversations: (list: ShowConversation[]) => void;
  setActiveConvId: (id: number | null) => void;
  setMessages: (list: ShowMessage[]) => void;
  addMessage: (msg: ShowMessage) => void;
  // 流式回调
  onStreamStart: (data: SSEMessageStart) => void;
  onStreamDelta: (data: SSEDelta) => void;
  onStreamEnd: (data: SSEMessageEnd) => void;
  onStreamReset: () => void;
  // 当流结束后将完整消息加入列表
  commitStreamMessage: (meta: SSEMessageEnd) => void;
  // 更新会话列表中的消息计数
  incrementMessageCount: (convId: number, count?: number) => void;
}

export const useChatStore = create<ChatState>()((set, get) => ({
  conversations: [],
  activeConvId: null,
  messages: [],
  streamingMessageId: null,
  streamingContent: "",
  isStreaming: false,

  setConversations: (list) => set({ conversations: list }),
  setActiveConvId: (id) => set({ activeConvId: id }),
  setMessages: (list) => set({ messages: list }),
  addMessage: (msg) => set((s) => ({ messages: [...s.messages, msg] })),

  onStreamStart: (data) =>
    set({
      streamingMessageId: data.messageId,
      streamingContent: "",
      isStreaming: true,
    }),

  onStreamDelta: (data) =>
    set((s) => ({
      streamingContent: s.streamingContent + data.content,
    })),

  onStreamEnd: (_data) =>
    set({ isStreaming: false }),

  onStreamReset: () =>
    set({
      streamingMessageId: null,
      streamingContent: "",
      isStreaming: false,
    }),

  commitStreamMessage: (meta) => {
    const { streamingContent, streamingMessageId } = get();
    if (!streamingMessageId) return;

    const msg: ShowMessage = {
      id: 0,
      uuid: streamingMessageId,
      role: "assistant",
      content: streamingContent,
      contentType: "text",
      tokenCount: meta.tokenCount,
      model: meta.model,
      latencyMs: meta.latencyMs,
      sources: meta.sources,
      createdAt: Date.now(),
    };

    set((s) => ({
      messages: [...s.messages, msg],
      streamingMessageId: null,
      streamingContent: "",
      isStreaming: false,
    }));
  },

  incrementMessageCount: (convId, count = 1) =>
    set((s) => ({
      conversations: s.conversations.map((c) =>
        c.id === convId ? { ...c, messageCount: c.messageCount + count } : c
      ),
    })),
}));
