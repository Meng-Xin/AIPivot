import type { FunctionComponent as FC } from 'preact';
import { useMemo, useState } from 'preact/hooks';
import type { WidgetMessage } from '../types';
import { renderMarkdownSafe } from '../utils/escape';

interface MessageBubbleProps {
  message: WidgetMessage;
  primary: string;
  /** 评分回调（仅 assistant 已完成消息会传入；undefined 表示禁用评分 UI） */
  onRate?: (message: WidgetMessage, rating: 'up' | 'down', feedback?: string) => void;
}

/** 单条消息气泡。assistant 流式生成中追加打字光标；assistant 完成态展示评分按钮。 */
export const MessageBubble: FC<MessageBubbleProps> = ({ message, primary, onRate }) => {
  const isUser = message.role === 'user';
  const html = useMemo(
    () => (isUser ? message.content : renderMarkdownSafe(message.content)),
    [isUser, message.content]
  );

  const bubbleStyle = isUser
    ? ({ background: primary, color: '#fff' } as const)
    : ({ background: '#f3f4f6', color: '#111827' } as const);

  // 仅 assistant + 非流式 + 注入了 onRate 时显示评分
  const canRate = !isUser && !message.streaming && !!onRate;
  const rating = message.rating;

  return (
    <div className={`apw-flex apw-w-full apw-my-2 ${isUser ? 'apw-justify-end' : 'apw-justify-start'}`}>
      <div className="apw-flex apw-flex-col apw-max-w-[85%]">
        <div
          className={`apw-px-3 apw-py-2 apw-rounded-lg apw-text-sm apw-leading-relaxed apw-break-words ${
            isUser ? 'apw-rounded-br-sm' : 'apw-rounded-bl-sm'
          } ${message.streaming ? 'apw-typing-cursor' : ''}`}
          style={bubbleStyle}
          // 内容已通过 renderMarkdownSafe 转义后再做受控替换，可安全使用 dangerouslySetInnerHTML
          dangerouslySetInnerHTML={{ __html: html }}
        />
        {canRate && <RatingBar message={message} primary={primary} onRate={onRate!} rating={rating} />}
      </div>
    </div>
  );
};

/** 评分按钮条：👍 / 👎，👎 弹出可选文字反馈。锁定语义：评分后不可更改。 */
const RatingBar: FC<{
  message: WidgetMessage;
  primary: string;
  rating?: 'up' | 'down';
  onRate: (message: WidgetMessage, rating: 'up' | 'down', feedback?: string) => void;
}> = ({ message, primary, rating, onRate }) => {
  // 文字反馈输入态（仅 👎 点击后激活）
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState('');

  if (rating) {
    // 已评分：仅展示被选中的高亮图标，锁定 UI
    return (
      <div className="apw-flex apw-items-center apw-gap-1 apw-mt-1 apw-ml-1">
        <RatingIcon
          kind={rating}
          active
          color={primary}
        />
        <span className="apw-text-[11px] apw-text-gray-400">
          {rating === 'up' ? '已反馈：有用' : message.ratingFeedback ? '已反馈' : '已反馈：需改进'}
        </span>
      </div>
    );
  }

  const handleThumbsUp = () => onRate(message, 'up', undefined);
  const handleThumbsDown = () => setEditing(true);
  const submitFeedback = (skip: boolean) => {
    const feedback = skip ? undefined : draft.trim();
    onRate(message, 'down', feedback);
    setEditing(false);
    setDraft('');
  };

  return (
    <div className="apw-mt-1 apw-ml-1">
      <div className="apw-flex apw-items-center apw-gap-1">
        <RatingIcon kind="up" color={primary} onClick={handleThumbsUp} title="有用" />
        <RatingIcon kind="down" color={primary} onClick={handleThumbsDown} title="需改进" />
      </div>
      {editing && (
        <div className="apw-mt-1 apw-p-2 apw-rounded-md apw-bg-gray-50 apw-border apw-border-gray-200">
          <textarea
            className="apw-w-full apw-text-xs apw-p-1.5 apw-resize-none apw-rounded apw-border apw-border-gray-200 apw-outline-none focus:apw-border-gray-300"
            placeholder="请告诉我们哪里做得不好（可选）"
            rows={2}
            maxlength={500}
            value={draft}
            onInput={(e) => setDraft((e.target as HTMLTextAreaElement).value)}
          />
          <div className="apw-flex apw-justify-end apw-gap-1 apw-mt-1">
            <button
              type="button"
              className="apw-text-[11px] apw-px-2 apw-py-0.5 apw-rounded apw-text-gray-500 hover:apw-bg-gray-200"
              onClick={() => submitFeedback(true)}
            >
              跳过
            </button>
            <button
              type="button"
              className="apw-text-[11px] apw-px-2 apw-py-0.5 apw-rounded apw-text-white"
              style={{ background: primary }}
              onClick={() => submitFeedback(false)}
            >
              提交
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

const RatingIcon: FC<{
  kind: 'up' | 'down';
  color: string;
  active?: boolean;
  onClick?: () => void;
  title?: string;
}> = ({ kind, color, active, onClick, title }) => {
  const cursor = onClick ? 'pointer' : 'default';
  const fill = active ? color : 'none';
  const stroke = active ? color : '#9ca3af';
  const path =
    kind === 'up'
      ? 'M14 9V5a3 3 0 0 0-3-3l-4 9v11h11.28a2 2 0 0 0 2-1.7l1.38-9a2 2 0 0 0-2-2.3zM7 22H4a2 2 0 0 1-2-2v-7a2 2 0 0 1 2-2h3'
      : 'M10 15v4a3 3 0 0 0 3 3l4-9V2H5.72a2 2 0 0 0-2 1.7l-1.38 9a2 2 0 0 0 2 2.3zm7-13h2.67A2.31 2.31 0 0 1 22 4v7a2.31 2.31 0 0 1-2.33 2H17';
  return (
    <button
      type="button"
      title={title}
      onClick={onClick}
      style={{ cursor, lineHeight: 0, padding: '2px' }}
      className="apw-rounded hover:apw-bg-gray-100"
    >
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill={fill}
        stroke={stroke}
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d={path} />
      </svg>
    </button>
  );
};
