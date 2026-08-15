import { create } from 'zustand';
import type { WidgetMessage } from './types';

/**
 * Widget 状态机（提炼自 web/src/store/chat.ts）。
 * 仅管理 UI 状态（面板开合、消息列表、流式状态），不直接调用 client —— 由组件层编排。
 */
interface WidgetState {
  // 面板
  isOpen: boolean;
  hasInteracted: boolean; // 用户是否已发起过对话（控制欢迎语显示）

  // 会话
  sessionToken: string | null;
  visitorId: string | null;

  // 消息
  messages: WidgetMessage[];

  // 引导问答（建会话时由后端从 KB 配置返回）
  suggestedQuestions: string[];

  // 流式状态
  isStreaming: boolean;
  error: string | null;

  // Actions
  open: () => void;
  close: () => void;
  toggle: () => void;
  setSession: (token: string, visitorId: string) => void;
  setSuggestedQuestions: (qs: string[]) => void;
  loadHistory: (msgs: WidgetMessage[]) => void;
  appendUserMessage: (content: string) => string; // 返回临时 uuid
  startAssistantStream: () => string; // 返回 assistant 占位 uuid
  appendDelta: (uuid: string, delta: string) => void;
  finalizeAssistant: (uuid: string, patch: Partial<WidgetMessage>) => void;
  removeMessage: (uuid: string) => void;
  setMessageRating: (uuid: string, rating: 'up' | 'down', feedback?: string) => void;
  setError: (msg: string | null) => void;
  setStreaming: (v: boolean) => void;
  reset: () => void;
}

export const useWidgetStore = create<WidgetState>((set) => ({
  isOpen: false,
  hasInteracted: false,
  sessionToken: null,
  visitorId: null,
  messages: [],
  suggestedQuestions: [],
  isStreaming: false,
  error: null,

  open: () => set({ isOpen: true }),
  close: () => set({ isOpen: false }),
  toggle: () => set((s) => ({ isOpen: !s.isOpen })),

  setSession: (token, visitorId) => set({ sessionToken: token, visitorId }),

  setSuggestedQuestions: (qs) => set({ suggestedQuestions: qs }),

  loadHistory: (msgs) => set({ messages: msgs, hasInteracted: msgs.length > 0 }),

  appendUserMessage: (content) => {
    const uuid = `u-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
    const msg: WidgetMessage = {
      uuid,
      role: 'user',
      content,
      contentType: 'text',
      createdAt: Date.now(),
    };
    set((s) => ({ messages: [...s.messages, msg], hasInteracted: true }));
    return uuid;
  },

  startAssistantStream: () => {
    const uuid = `a-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
    const msg: WidgetMessage = {
      uuid,
      role: 'assistant',
      content: '',
      contentType: 'text',
      createdAt: Date.now(),
      streaming: true,
    };
    set((s) => ({ messages: [...s.messages, msg], isStreaming: true }));
    return uuid;
  },

  appendDelta: (uuid, delta) =>
    set((s) => ({
      messages: s.messages.map((m) =>
        m.uuid === uuid ? { ...m, content: m.content + delta } : m
      ),
    })),

  finalizeAssistant: (uuid, patch) =>
    set((s) => ({
      messages: s.messages.map((m) =>
        m.uuid === uuid ? { ...m, ...patch, streaming: false } : m
      ),
      isStreaming: false,
    })),

  removeMessage: (uuid) =>
    set((s) => ({ messages: s.messages.filter((m) => m.uuid !== uuid) })),

  setMessageRating: (uuid, rating, feedback) =>
    set((s) => ({
      messages: s.messages.map((m) =>
        m.uuid === uuid ? { ...m, rating, ratingFeedback: feedback } : m
      ),
    })),

  setError: (msg) => set({ error: msg }),
  setStreaming: (v) => set({ isStreaming: v }),

  reset: () =>
    set({
      messages: [],
      hasInteracted: false,
      isStreaming: false,
      error: null,
      suggestedQuestions: [],
    }),
}));
