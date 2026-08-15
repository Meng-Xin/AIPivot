import type { FunctionComponent as FC } from 'preact';
import type { WidgetMessage } from '../types';
import { InputArea } from './InputArea';
import { MessageList } from './MessageList';

interface ChatPanelProps {
  title: string;
  welcome: string;
  placeholder: string;
  primary: string;
  messages: WidgetMessage[];
  isStreaming: boolean;
  error: string | null;
  suggestedQuestions: string[];
  onSend: (text: string) => void;
  onClose: () => void;
  onRate: (message: WidgetMessage, rating: 'up' | 'down', feedback?: string) => void;
}

/** 聊天面板容器：标题栏 + 消息列表 + 输入区。 */
export const ChatPanel: FC<ChatPanelProps> = ({
  title,
  welcome,
  placeholder,
  primary,
  messages,
  isStreaming,
  error,
  suggestedQuestions,
  onSend,
  onClose,
  onRate,
}) => (
  <div
    className="apw-panel-enter apw-fixed apw-bg-white apw-rounded-xl apw-shadow-2xl apw-flex apw-flex-col apw-overflow-hidden"
    style={{
      bottom: '96px',
      right: '24px',
      width: '360px',
      maxWidth: 'calc(100vw - 32px)',
      height: '520px',
      maxHeight: 'calc(100vh - 128px)',
    }}
  >
    {/* 标题栏 */}
    <div
      className="apw-px-4 apw-py-3 apw-flex apw-items-center apw-justify-between apw-text-white"
      style={{ background: primary }}
    >
      <div className="apw-flex apw-items-center apw-gap-2">
        <span className="apw-font-medium apw-text-sm">{title}</span>
        <span className="apw-w-2 apw-h-2 apw-rounded-full apw-bg-green-400" />
      </div>
      <button
        type="button"
        className="apw-text-white/80 hover:apw-text-white"
        onClick={onClose}
        aria-label="关闭"
      >
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
          <line x1="6" y1="6" x2="18" y2="18" />
          <line x1="18" y1="6" x2="6" y2="18" />
        </svg>
      </button>
    </div>

    {/* 错误提示 */}
    {error && (
      <div className="apw-px-3 apw-py-2 apw-bg-red-50 apw-text-red-600 apw-text-xs apw-text-center">
        {error}
      </div>
    )}

    <MessageList
      messages={messages}
      welcome={welcome}
      primary={primary}
      isStreaming={isStreaming}
      suggestedQuestions={suggestedQuestions}
      onRate={onRate}
      onSuggestedClick={onSend}
    />

    <InputArea
      placeholder={placeholder}
      primary={primary}
      disabled={isStreaming}
      onSend={onSend}
    />
  </div>
);
