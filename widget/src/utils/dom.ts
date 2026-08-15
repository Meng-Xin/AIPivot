/**
 * Shadow DOM 创建与样式注入工具。
 * 使用 closed mode 防止宿主页面通过 element.shadowRoot 探测/篡改 Widget 内部。
 */

let injectedCss = '';

/** 由 Vite 构建时调用，将打包后的 CSS 字符串注入到 Shadow DOM 中。 */
export function setWidgetCss(css: string): void {
  injectedCss = css;
}

/** 创建宿主容器（position: fixed 挂到 body），并附加 closed Shadow Root。 */
export function createShadowContainer(): { host: HTMLElement; shadow: ShadowRoot } {
  const host = document.createElement('div');
  host.setAttribute('data-aipivot-widget', '');
  // 宿主元素零样式，所有定位通过内部容器实现
  host.style.cssText = 'all: initial; position: fixed; z-index: 2147483000;';
  document.body.appendChild(host);

  const shadow = host.attachShadow({ mode: 'closed' });

  if (injectedCss) {
    const styleEl = document.createElement('style');
    styleEl.textContent = injectedCss;
    shadow.appendChild(styleEl);
  }

  return { host, shadow };
}

/** 移除宿主容器及其 Shadow Root */
export function removeShadowContainer(host: HTMLElement): void {
  if (host.parentNode) {
    host.parentNode.removeChild(host);
  }
}
