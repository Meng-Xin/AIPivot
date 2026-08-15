import type { FunctionComponent as FC } from 'preact';

/** 三点跳动 typing 指示器，流式开始但还未收到首个 token 时显示。 */
export const TypingIndicator: FC = () => (
  <div className="apw-flex apw-gap-1 apw-py-2">
    <span className="apw-typing-dot" />
    <span className="apw-typing-dot" />
    <span className="apw-typing-dot" />
  </div>
);
