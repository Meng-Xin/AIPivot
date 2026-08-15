import type { FunctionComponent as FC } from 'preact';
import { useEffect, useState } from 'preact/hooks';
import { ChatPanel } from './components/ChatPanel';
import { Launcher } from './components/Launcher';
import { WidgetClient } from './client';
import type { WidgetConfig, WidgetHandle, WidgetMessage } from './types';
import { normalizeConfig } from './config';
import { useWidgetStore } from './store';
import { clearSessionToken, getOrCreateVisitorId, loadSessionToken, saveSessionToken } from './storage';

interface WidgetRootProps {
  rawConfig: WidgetConfig;
  handleRef: { current: WidgetHandle | null };
}

/**
 * WidgetRoot — Preact 根组件。
 * 负责：会话初始化（首次创建或恢复）、发送消息（调用 SSE 流式 API）、UI 编排。
 */
export const WidgetRoot: FC<WidgetRootProps> = ({ rawConfig, handleRef }) => {
  const config = normalizeConfig(rawConfig);
  const [client] = useState(() => new WidgetClient(config));

  const store = useWidgetStore;

  // 初始化：恢复或创建会话
  useEffect(() => {
    const init = async () => {
      const visitorId = getOrCreateVisitorId(config);
      const existingToken = loadSessionToken(config);

      if (existingToken) {
        // 尝试恢复历史消息
        try {
          const msgs = await client.listMessages(existingToken);
          const widgetMsgs: WidgetMessage[] = msgs.map((m) => ({
            uuid: m.uuid,
            role: m.role as WidgetMessage['role'],
            content: m.content,
            contentType: m.contentType,
            tokenCount: m.tokenCount,
            model: m.model,
            sources: m.sources,
            rating: m.rating === 'up' || m.rating === 'down' ? m.rating : undefined,
            ratingFeedback: m.ratingFeedback,
            createdAt: m.createdAt * 1000,
          }));
          store.getState().setSession(existingToken, visitorId);
          store.getState().loadHistory(widgetMsgs);
          return;
        } catch {
          // token 失效（会话被删/租户变化），清除后重新创建
          clearSessionToken(config);
        }
      }

      try {
        const session = await client.createSession(visitorId);
        saveSessionToken(config, session.sessionToken);
        store.getState().setSession(session.sessionToken, visitorId);
        // 引导问答：建会话时由后端从绑定的 KB 配置返回
        store.getState().setSuggestedQuestions(session.suggestedQuestions ?? []);
      } catch (e) {
        store.getState().setError((e as Error).message);
      }
    };
    void init();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 发送消息编排
  const sendMessage = async (text: string) => {
    const state = store.getState();
    if (!state.sessionToken) {
      store.getState().setError('会话尚未就绪，请稍候');
      return;
    }
    if (state.isStreaming) return;

    store.getState().setError(null);
    store.getState().appendUserMessage(text);
    const assistantUuid = store.getState().startAssistantStream();

    await client.sendMessageStream(
      state.sessionToken,
      text,
      {
        onDelta: (d) => store.getState().appendDelta(assistantUuid, d.content),
        onEnd: (meta) =>
          store.getState().finalizeAssistant(assistantUuid, {
            uuid: meta.messageId || assistantUuid,
            model: meta.model,
            tokenCount: meta.tokenCount,
            sources: meta.sources,
          }),
        onError: (err) => {
          // 出错时移除空的 assistant 占位，避免遗留空气泡
          const cur = store.getState().messages.find((m) => m.uuid === assistantUuid);
          if (cur && cur.content === '') {
            store.getState().removeMessage(assistantUuid);
          } else {
            store.getState().finalizeAssistant(assistantUuid, {});
          }
          store.getState().setError(err.msg || 'AI 回复失败');
          store.getState().setStreaming(false);
        },
      }
    );
  };

  // 暴露 handle
  useEffect(() => {
    handleRef.current = {
      open: () => store.getState().open(),
      close: () => store.getState().close(),
      toggle: () => store.getState().toggle(),
      send: (t: string) => sendMessage(t),
      destroy: () => {
        /* 由 index.ts 顶层处理卸载 */
      },
    };
  }, [handleRef]);

  /**
   * 评分回调：本地先乐观更新 → 调 API → 失败回滚。
   * 第一波为锁定语义，一旦评分本地状态机不允许再点击（MessageBubble 控制按钮渲染）。
   */
  const handleRate = async (message: WidgetMessage, rating: 'up' | 'down', feedback?: string) => {
    const token = store.getState().sessionToken;
    if (!token) return;
    const prevRating = message.rating;
    const prevFeedback = message.ratingFeedback;
    store.getState().setMessageRating(message.uuid, rating, feedback);
    try {
      await client.rateMessage(token, message.uuid, rating, feedback);
    } catch (e) {
      // 回滚
      store.getState().setMessageRating(message.uuid, prevRating as 'up' | 'down' | undefined, prevFeedback);
      store.getState().setError((e as Error).message);
    }
  };

  const isOpen = useWidgetStore((s) => s.isOpen);
  const messages = useWidgetStore((s) => s.messages);
  const isStreaming = useWidgetStore((s) => s.isStreaming);
  const error = useWidgetStore((s) => s.error);
  const suggestedQuestions = useWidgetStore((s) => s.suggestedQuestions);

  return (
    <div className="apw-root">
      {isOpen && (
        <ChatPanel
          title={config.title}
          welcome={config.welcome}
          placeholder={config.placeholder}
          primary={config.theme.primary}
          messages={messages}
          isStreaming={isStreaming}
          error={error}
          suggestedQuestions={suggestedQuestions}
          onSend={sendMessage}
          onClose={() => store.getState().close()}
          onRate={handleRate}
        />
      )}
      <Launcher
        open={isOpen}
        primary={config.theme.primary}
        offset={config.offset}
        onClick={() => store.getState().toggle()}
      />
    </div>
  );
};
