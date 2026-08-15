import type { WidgetConfig } from './types';

/** 默认配置（与 WidgetConfig 中的可选字段一一对应） */
export const DEFAULT_CONFIG = {
  title: '在线客服',
  welcome: '您好，有什么可以帮您？',
  placeholder: '请输入消息...',
  storage: 'localStorage' as const,
  offset: { bottom: 24, right: 24 },
  theme: { primary: '#4f46e5' },
};

/** 校验必填配置并合并默认值。校验失败抛出 Error，由 init 捕获后打印到 console。 */
export function normalizeConfig(input: WidgetConfig): Required<WidgetConfig> {
  if (!input || typeof input !== 'object') {
    throw new Error('AIPivotWidget: config 不能为空');
  }
  if (!input.publicKey || !input.publicKey.startsWith('pk_')) {
    throw new Error('AIPivotWidget: publicKey 必须为 pk_ 前缀的 public key');
  }
  if (!input.baseUrl) {
    throw new Error('AIPivotWidget: baseUrl 不能为空');
  }

  return {
    publicKey: input.publicKey,
    baseUrl: input.baseUrl.replace(/\/$/, ''), // 去尾斜杠，避免拼接出 //
    title: input.title ?? DEFAULT_CONFIG.title,
    welcome: input.welcome ?? DEFAULT_CONFIG.welcome,
    placeholder: input.placeholder ?? DEFAULT_CONFIG.placeholder,
    storage: input.storage ?? DEFAULT_CONFIG.storage,
    offset: {
      bottom: input.offset?.bottom ?? DEFAULT_CONFIG.offset.bottom,
      right: input.offset?.right ?? DEFAULT_CONFIG.offset.right,
    },
    theme: {
      primary: input.theme?.primary ?? DEFAULT_CONFIG.theme.primary,
    },
  };
}
