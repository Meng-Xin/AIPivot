/** XSS 防护：HTML 实体转义，用于将 LLM 返回内容安全插入 DOM。 */
export function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

/**
 * 极简 markdown 渲染：仅支持 **粗体**、*斜体*、`行内代码`、换行。
 * 输出前先 escapeHtml，再应用受控替换，避免 onerror 等事件注入。
 * 不支持 raw HTML 标签（任何 < 都会被转义）。
 */
export function renderMarkdownSafe(src: string): string {
  const escaped = escapeHtml(src);
  return escaped
    .replace(/`([^`]+)`/g, '<code class="apw-code">$1</code>')
    .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
    .replace(/\*([^*]+)\*/g, '<em>$1</em>')
    .replace(/\n/g, '<br/>');
}
