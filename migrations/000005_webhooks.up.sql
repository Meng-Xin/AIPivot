-- ============================================================
-- Migration: 000005_webhooks
-- 描述: Webhook 配置表，支持出站事件推送和入站消息接收
-- 说明: 每个租户可注册多个 Webhook，指定监听事件类型和回调 URL
-- ============================================================

CREATE TABLE webhooks (
    id            BIGSERIAL    PRIMARY KEY,
    uuid          UUID         NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    tenant_id     BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name          VARCHAR(255) NOT NULL,
    url           TEXT         NOT NULL,
    secret        VARCHAR(255) NOT NULL DEFAULT '',
    events        JSONB        NOT NULL DEFAULT '["message.created"]',
    channel_type  VARCHAR(50)  NOT NULL DEFAULT 'webhook',
    status        VARCHAR(20)  NOT NULL DEFAULT 'active',
    retry_count   INT          NOT NULL DEFAULT 3,
    timeout_ms    INT          NOT NULL DEFAULT 5000,
    last_error    TEXT         NOT NULL DEFAULT '',
    last_trigger  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  webhooks                IS 'Webhook 配置表：租户注册的回调端点，用于出站事件推送和第三方平台集成';
COMMENT ON COLUMN webhooks.id             IS '自增主键';
COMMENT ON COLUMN webhooks.uuid           IS '对外暴露的唯一标识（UUID v4）';
COMMENT ON COLUMN webhooks.tenant_id      IS '所属租户 ID（级联删除）';
COMMENT ON COLUMN webhooks.name           IS 'Webhook 名称（便于管理识别）';
COMMENT ON COLUMN webhooks.url            IS '回调 URL（HTTPS 推荐）';
COMMENT ON COLUMN webhooks.secret         IS '签名密钥（用于 HMAC-SHA256 签名验证，空表示不签名）';
COMMENT ON COLUMN webhooks.events         IS '订阅事件类型列表（JSON 数组），如 message.created / conversation.closed';
COMMENT ON COLUMN webhooks.channel_type   IS '关联渠道类型: webhook / wechat / feishu / dingtalk';
COMMENT ON COLUMN webhooks.status         IS '状态: active / disabled';
COMMENT ON COLUMN webhooks.retry_count    IS '失败重试次数上限';
COMMENT ON COLUMN webhooks.timeout_ms     IS '请求超时毫秒数';
COMMENT ON COLUMN webhooks.last_error     IS '最近一次推送的错误信息（成功时清空）';
COMMENT ON COLUMN webhooks.last_trigger   IS '最近一次触发时间';
COMMENT ON COLUMN webhooks.created_at     IS '创建时间';
COMMENT ON COLUMN webhooks.updated_at     IS '更新时间';

CREATE INDEX idx_webhooks_tenant_id     ON webhooks(tenant_id);
CREATE INDEX idx_webhooks_status        ON webhooks(status);
CREATE INDEX idx_webhooks_channel_type  ON webhooks(channel_type);
