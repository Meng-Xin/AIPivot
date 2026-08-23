import { defineConfig } from 'vite';
import preact from '@preact/preset-vite';

// Widget SDK 构建配置：
// - lib 模式 + IIFE 输出，挂载到 window.AIPivotWidget
// - CSS 内联到 JS（cssCodeSplit:false + inlineDynamicImports），客户网站只需一行 <script>
// - 不打包 Preact 到全局（externalize 由 preset 内部处理 jsx-runtime）
export default defineConfig({
  plugins: [preact()],
  // 端口固定在 5174：public key 的 allowed_origins 是严格匹配 + fail-closed，
  // 端口漂移会让 Widget 直接吃 403，所以用 strictPort 让冲突尽早暴露。
  server: {
    port: 5174,
    strictPort: true,
  },
  build: {
    lib: {
      entry: 'src/index.tsx',
      name: 'AIPivotWidget',
      formats: ['iife'],
      fileName: () => 'aipivot-widget.js',
    },
    cssCodeSplit: false,
    minify: 'esbuild',
    sourcemap: false,
    outDir: 'dist',
  },
  define: {
    'process.env.NODE_ENV': JSON.stringify('production'),
  },
});
