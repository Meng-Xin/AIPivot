import type { FunctionComponent as FC } from 'preact';
import { useEffect, useRef } from 'preact/hooks';
import type { WidgetMessage } from '../types';
import { MessageBubble } from './MessageBubble';
import { TypingIndicator } from './TypingIndicator';

interface MessageListProps {
  messages: WidgetMessage[];
  welcome: string;
  primary: string;
  isStreaming: boolean;
  /** 引导问答列表（仅首屏空消息列表时展示） */
  suggestedQuestions?: string[];
  /** 评分回调 */
  onRate?: (message: WidgetMessage, rating: 'up' | 'down', feedback?: string) => void;
  /** 点击引导问答 chip 时触发发送 */
  onSuggestedClick?: (question: string) => void;
}

/** 消息列表，自动滚动到底部；流式未收到首个 token 时显示 typing 指示器。 */
export const MessageList: FC<MessageListProps> = ({
  messages,
  welcome,
  primary,
  isStreaming,
  suggestedQuestions,
  onRate,
  onSuggestedClick,
}) => {
  const endRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' });
  }, [messages, isStreaming]);

  // 是否显示 typing 指示器：流式中且最后一条 assistant 消息内容为空
  const showTyping =
    isStreaming &&
    messages.length > 0 &&
    messages[messages.length - 1].role === 'assistant' &&
    messages[messages.length - 1].content === '';

  const showSuggestions = messages.length === 0 && (suggestedQuestions?.length ?? 0) > 0;

  return (
    <div className="apw-flex-1 apw-overflow-y-auto apw-px-3 apw-py-3 apw-scroll">
      {!messages.length && (
        <div className="apw-text-center apw-text-gray-500 apw-text-sm apw-mt-8 apw-px-4">
          {welcome}
        </div>
      )}
      {showSuggestions && (
        <div className="apw-flex apw-flex-col apw-gap-1.5 apw-mt-3 apw-px-2">
          {(suggestedQuestions ?? []).slice(0, 6).map((q, i) => (
            <button
              key={`${q}-${i}`}
              type="button"
              onClick={() => onSuggestedClick?.(q)}
              className="apw-text-left apw-text-xs apw-px-2.5 apw-py-1.5 apw-rounded-lg apw-border apw-border-gray-200 apw-bg-gray-50 apw-text-gray-700 hover:apw-bg-white apw-transition"
              style={{ borderColor: 'transparent' }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLButtonElement).style.borderColor = primary;
                (e.currentTarget as HTMLButtonElement).style.color = primary;
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLButtonElement).style.borderColor = 'transparent';
                (e.currentTarget as HTMLButtonElement).style.color = '#374151';
              }}
            >
              {q}
            </button>
          ))}
        </div>
      )}
      {messages.map((m) => (
        <MessageBubble key={m.uuid} message={m} primary={primary} onRate={onRate} />
      ))}
      {showTyping && (
        <div className="apw-flex apw-justify-start apw-my-2 apw-ml-1">
          <div className="apw-px-3 apw-py-2 apw-rounded-lg apw-rounded-bl-sm" style={{ background: '#f3f4f6' }}>
            <TypingIndicator />
          </div>
        </div>
      )}
      <div ref={endRef} />
    </div>
  );
};
