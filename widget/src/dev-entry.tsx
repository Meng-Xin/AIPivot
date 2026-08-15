/**
 * 本地开发入口：直接在主文档挂载 Widget（绕过 Shadow DOM 便于调试）。
 * 在 index.html 中通过 <script type="module" src="/src/dev-entry.tsx"> 加载。
 *
 * 使用前请填入你自己的 publicKey（管理后台创建 public key）与后端 baseUrl。
 */
import { init } from './index.tsx';

init({
  publicKey: 'pk_REPLACE_ME',
  baseUrl: 'http://127.0.0.1:8888',
  title: '智能客服',
  welcome: '您好，请问有什么可以帮您？',
  theme: { primary: '#4f46e5' },
});
