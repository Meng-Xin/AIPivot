import type { FunctionComponent as FC } from 'preact';
import { useMemo } from 'preact/hooks';

interface LauncherProps {
  open: boolean;
  primary: string;
  offset: { bottom: number; right: number };
  onClick: () => void;
}

/** 悬浮启动按钮（右下角圆形）。open=true 时变为关闭图标。 */
export const Launcher: FC<LauncherProps> = ({ open, primary, offset, onClick }) => {
  const style = useMemo(
    () => ({
      background: primary,
      bottom: `${offset.bottom}px`,
      right: `${offset.right}px`,
    }) as CSSStyleDeclaration,
    [primary, offset]
  );

  return (
    <button
      type="button"
      className="apw-launcher"
      style={style as unknown as Record<string, string>}
      onClick={onClick}
      aria-label={open ? '关闭聊天' : '打开聊天'}
    >
      {open ? (
        // X 图标（关闭）
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round">
          <line x1="6" y1="6" x2="18" y2="18" />
          <line x1="18" y1="6" x2="6" y2="18" />
        </svg>
      ) : (
        // 聊天气泡图标（打开）
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z" />
        </svg>
      )}
      <style>{`
        .apw-launcher {
          position: fixed;
          width: 56px;
          height: 56px;
          border-radius: 50%;
          border: none;
          cursor: pointer;
          box-shadow: 0 4px 12px rgba(0,0,0,0.2);
          display: flex;
          align-items: center;
          justify-content: center;
          transition: transform 0.15s ease, box-shadow 0.15s ease;
        }
        .apw-launcher:hover { transform: scale(1.05); box-shadow: 0 6px 16px rgba(0,0,0,0.25); }
      `}</style>
    </button>
  );
};
