import { defineConfig } from 'vite';
import preact from '@preact/preset-vite';

// Widget SDK 构建配置：
// - lib 模式 + IIFE 输出，挂载到 window.AIPivotWidget
// - CSS 内联到 JS（cssCodeSplit:false + inlineDynamicImports），客户网站只需一行 <script>
// - 不打包 Preact 到全局（externalize 由 preset 内部处理 jsx-runtime）
export default defineConfig({
  plugins: [preact()],
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
