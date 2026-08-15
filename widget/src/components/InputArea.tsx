import type { FunctionComponent as FC } from 'preact';
import { useEffect, useRef, useState } from 'preact/hooks';

interface InputAreaProps {
  placeholder: string;
  primary: string;
  disabled: boolean;
  onSend: (text: string) => void;
}

/** 输入区：自适应高度的 textarea + 发送按钮。Enter 发送，Shift+Enter 换行。 */
export const InputArea: FC<InputAreaProps> = ({ placeholder, primary, disabled, onSend }) => {
  const [value, setValue] = useState('');
  const taRef = useRef<HTMLTextAreaElement>(null);

  // 自适应高度
  useEffect(() => {
    const ta = taRef.current;
    if (!ta) return;
    ta.style.height = 'auto';
    ta.style.height = `${Math.min(ta.scrollHeight, 120)}px`;
  }, [value]);

  const submit = () => {
    const text = value.trim();
    if (!text || disabled) return;
    onSend(text);
    setValue('');
  };

  const onKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      submit();
    }
  };

  return (
    <div className="apw-border-t apw-border-gray-200 apw-p-3 apw-flex apw-items-end apw-gap-2">
      <textarea
        ref={taRef}
        rows={1}
        className="apw-flex-1 apw-resize-none apw-border apw-border-gray-300 apw-rounded-lg apw-px-3 apw-py-2 apw-text-sm apw-outline-none apw-focus:border-transparent apw-max-h-[120px]"
        style={{ focus: { borderColor: primary } as never }}
        placeholder={placeholder}
        value={value}
        disabled={disabled}
        onInput={(e) => setValue((e.target as HTMLTextAreaElement).value)}
        onKeyDown={onKeyDown}
      />
      <button
        type="button"
        className="apw-px-4 apw-py-2 apw-rounded-lg apw-text-white apw-text-sm apw-font-medium apw-disabled:opacity-50"
        style={{ background: primary }}
        disabled={disabled || !value.trim()}
        onClick={submit}
      >
        发送
      </button>
    </div>
  );
};
