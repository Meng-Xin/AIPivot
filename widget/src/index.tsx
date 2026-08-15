import { render } from 'preact';
import type { WidgetConfig, WidgetHandle } from './types';
import { WidgetRoot } from './widget';
import { createShadowContainer, removeShadowContainer, setWidgetCss } from './utils/dom';

// Vite 在 build 时将所有 CSS 合并为一个字符串，通过 ?inline 后缀导入
// @ts-ignore - Vite 专用后缀
import widgetCss from './styles/index.css?inline';

setWidgetCss(widgetCss as string);

/**
 * 初始化 Widget SDK。
 * 在宿主页面创建 Shadow DOM 容器并挂载 Preact 应用，返回操作句柄。
 *
 * @example
 *   const handle = AIPivotWidget.init({
 *     publicKey: 'pk_xxx',
 *     baseUrl: 'https://api.example.com',
 *   });
 *   handle.open();
 */
export function init(config: WidgetConfig): WidgetHandle {
  const handleRef: { current: WidgetHandle | null } = { current: null };

  const { host, shadow } = createShadowContainer();

  // 在 Shadow Root 内创建挂载点
  const mountPoint = document.createElement('div');
  mountPoint.setAttribute('class', 'apw-mount');
  shadow.appendChild(mountPoint);

  // 注入主题色 CSS 变量（覆盖 Tailwind 中 var(--apw-primary) 的默认值）
  if (config.theme?.primary) {
    mountPoint.style.setProperty('--apw-primary', config.theme.primary);
  }

  render(<WidgetRoot rawConfig={config} handleRef={handleRef} />, mountPoint);

  const handle: WidgetHandle = {
    open: () => handleRef.current?.open(),
    close: () => handleRef.current?.close(),
    toggle: () => handleRef.current?.toggle(),
    send: (text) => handleRef.current?.send(text),
    destroy: () => {
      render(null, mountPoint);
      removeShadowContainer(host);
    },
  };
  handleRef.current = handle;
  return handle;
}

// 挂载到全局，便于 <script> 接入
if (typeof window !== 'undefined') {
  (window as unknown as { AIPivotWidget: { init: typeof init } }).AIPivotWidget = { init };
}
