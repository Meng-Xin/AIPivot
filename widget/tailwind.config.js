/** @type {import('tailwindcss').Config} */
export default {
  // 仅扫描 widget 自身源码，避免打包多余样式
  content: ['./src/**/*.{ts,tsx}'],
  // 不使用 Tailwind 默认 reset（避免污染宿主页面，即使有 Shadow DOM 也保持纯净）
  corePlugins: {
    preflight: false,
  },
  theme: {
    extend: {
      colors: {
        // 通过 CSS 变量覆盖主题色
        widget: {
          primary: 'var(--apw-primary, #4f46e5)',
        },
      },
    },
  },
  plugins: [],
};
