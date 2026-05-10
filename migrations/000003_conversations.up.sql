-- ============================================================
-- Migration: 000003_conversations
-- 描述: 对话模块（conversations / messages）
-- 说明: 支持 AI 客服会话管理，消息存储与上下文追踪
-- ============================================================

-- -----------------------------------------------------------
-- conversations: 会话表
-- -----------------------------------------------------------
CREATE TABLE conversations (
    id              BIGSERIAL    PRIMARY KEY,
    uuid            UUID         NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    tenant_id       BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id         BIGINT       REFERENCES users(id) ON DELETE SET NULL,
    knowledge_base_id BIGINT     REFERENCES knowledge_bases(id) ON DELETE SET NULL,
    title           VARCHAR(500) NOT NULL DEFAULT '',
    status          VARCHAR(20)  NOT NULL DEFAULT 'active',
    channel         VARCHAR(50)  NOT NULL DEFAULT 'web',
    message_count   INT          NOT NULL DEFAULT 0,
    summary         TEXT         NOT NULL DEFAULT '',
    metadata        JSONB        NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    closed_at       TIMESTAMPTZ
);

COMMENT ON TABLE  conversations                    IS '会话表：一次用户与 AI 的完整对话，可关联知识库、可转人工';
COMMENT ON COLUMN conversations.id                 IS '自增主键';
COMMENT ON COLUMN conversations.uuid               IS '对外暴露的唯一标识（UUID v4）';
COMMENT ON COLUMN conversations.tenant_id          IS '所属租户 ID（级联删除）';
COMMENT ON COLUMN conversations.user_id            IS '发起用户 ID（匿名对话可为空）';
COMMENT ON COLUMN conversations.knowledge_base_id  IS '关联知识库 ID（决定 RAG 检索范围）';
COMMENT ON COLUMN conversations.title              IS '会话标题（由 LLM 自动生成或用户指定）';
COMMENT ON COLUMN conversations.status             IS '状态: active / waiting_human / resolved / closed';
COMMENT ON COLUMN conversations.channel            IS '接入渠道: web / api / wechat / feishu';
COMMENT ON COLUMN conversations.message_count      IS '消息数量（冗余计数）';
COMMENT ON COLUMN conversations.summary            IS '会话摘要（上下文压缩后的摘要文本）';
COMMENT ON COLUMN conversations.metadata           IS '会话元数据（客户端信息、标签等）';
COMMENT ON COLUMN conversations.created_at         IS '创建时间';
COMMENT ON COLUMN conversations.updated_at         IS '更新时间';
COMMENT ON COLUMN conversations.closed_at          IS '关闭时间';

CREATE INDEX idx_conv_tenant_id  ON conversations(tenant_id);
CREATE INDEX idx_conv_user_id    ON conversations(user_id);
CREATE INDEX idx_conv_status     ON conversations(status);
CREATE INDEX idx_conv_created_at ON conversations(created_at DESC);

-- -----------------------------------------------------------
-- messages: 消息表
-- -----------------------------------------------------------
CREATE TABLE messages (
    id              BIGSERIAL    PRIMARY KEY,
    uuid            UUID         NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    conversation_id BIGINT       NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    tenant_id       BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role            VARCHAR(20)  NOT NULL DEFAULT 'user',
    content         TEXT         NOT NULL,
    content_type    VARCHAR(20)  NOT NULL DEFAULT 'text',
    token_count     INT          NOT NULL DEFAULT 0,
    model           VARCHAR(100) NOT NULL DEFAULT '',
    latency_ms      INT          NOT NULL DEFAULT 0,
    sources         JSONB        NOT NULL DEFAULT '[]',
    metadata        JSONB        NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  messages                    IS '消息表：会话中的每条消息，包括用户消息、AI 回复、系统消息';
COMMENT ON COLUMN messages.id                 IS '自增主键';
COMMENT ON COLUMN messages.uuid               IS '对外暴露的唯一标识（UUID v4）';
COMMENT ON COLUMN messages.conversation_id    IS '所属会话 ID（级联删除）';
COMMENT ON COLUMN messages.tenant_id          IS '所属租户 ID（冗余字段，加速查询）';
COMMENT ON COLUMN messages.role               IS '角色: user / assistant / system';
COMMENT ON COLUMN messages.content            IS '消息内容';
COMMENT ON COLUMN messages.content_type       IS '内容类型: text / image / file';
COMMENT ON COLUMN messages.token_count        IS 'Token 消耗数量';
COMMENT ON COLUMN messages.model              IS '生成此消息的 LLM 模型名称（仅 assistant 消息）';
COMMENT ON COLUMN messages.latency_ms         IS 'LLM 响应延迟（毫秒，仅 assistant 消息）';
COMMENT ON COLUMN messages.sources            IS '知识来源引用（JSON 数组，RAG 检索命中的 chunk UUID 列表）';
COMMENT ON COLUMN messages.metadata           IS '消息元数据（评价、标注等）';
COMMENT ON COLUMN messages.created_at         IS '创建时间';

CREATE INDEX idx_messages_conv_id    ON messages(conversation_id);
CREATE INDEX idx_messages_tenant_id  ON messages(tenant_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
