-- ============================================================
-- Migration: 000008_widget_public_key
-- 描述: 为 Chat Widget 引入 public key 体系，并扩展会话表关联访客
-- 说明:
--   1) api_keys 增加 key_type / allowed_origins / knowledge_base_id
--      - key_type=public 的密钥可嵌入前端，受域名白名单保护
--      - 强制绑定单个知识库，限制 RAG 检索范围
--   2) conversations 增加 external_user_id，用于关联 Widget 访客
--      （不引入 widget_sessions 关联表，避免过度设计）
-- ============================================================

-- ---------- api_keys 扩展 ----------
ALTER TABLE api_keys
    ADD COLUMN key_type          VARCHAR(20) NOT NULL DEFAULT 'master',
    ADD COLUMN allowed_origins   TEXT[]      NOT NULL DEFAULT '{}',
    ADD COLUMN knowledge_base_id BIGINT      REFERENCES knowledge_bases(id) ON DELETE SET NULL;

COMMENT ON COLUMN api_keys.key_type          IS '密钥类型: master（管理员/服务端） / public（前端嵌入，Widget 专用）';
COMMENT ON COLUMN api_keys.allowed_origins   IS '允许的来源域名白名单（仅 public key 生效，严格匹配，空数组 fail-closed）';
COMMENT ON COLUMN api_keys.knowledge_base_id IS '绑定的知识库 ID（仅 public key 生效，强制限制 RAG 检索范围）';

CREATE INDEX idx_api_keys_key_type ON api_keys(key_type) WHERE key_type = 'public';

-- ---------- conversations 扩展 ----------
ALTER TABLE conversations
    ADD COLUMN external_user_id VARCHAR(100) NOT NULL DEFAULT '';

COMMENT ON COLUMN conversations.external_user_id IS '外部渠道用户标识（Widget 访客 ID / Webhook 外部用户 ID），空表示已登录用户';

CREATE INDEX idx_conv_external_user ON conversations(external_user_id) WHERE external_user_id != '';
