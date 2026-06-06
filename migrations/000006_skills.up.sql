-- Migration: 000006_skills
-- Description: 租户自定义工具（Skill）表
-- 每个租户可定义任意数量的 HTTP 回调工具，由 Agent 在对话中调用

CREATE TABLE IF NOT EXISTS skills (
    id          BIGSERIAL    PRIMARY KEY,
    tenant_id   BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    parameters  JSONB        NOT NULL DEFAULT '{}',
    endpoint    VARCHAR(500) NOT NULL DEFAULT '',
    method      VARCHAR(10)  NOT NULL DEFAULT 'POST',
    headers     JSONB        NOT NULL DEFAULT '{}',
    timeout_ms  INT          NOT NULL DEFAULT 5000,
    enabled     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

COMMENT ON TABLE  skills                  IS '租户自定义工具表：Agent Function Calling 的 HTTP 回调技能';
COMMENT ON COLUMN skills.id              IS '自增主键';
COMMENT ON COLUMN skills.tenant_id       IS '所属租户 ID（级联删除）';
COMMENT ON COLUMN skills.name            IS '工具唯一名称（租户内唯一，对应 function.name）';
COMMENT ON COLUMN skills.description     IS '工具描述，帮助 LLM 判断何时调用';
COMMENT ON COLUMN skills.parameters      IS 'JSON Schema 参数定义（type/properties/required），传给 LLM 的 function parameters';
COMMENT ON COLUMN skills.endpoint        IS 'HTTP 回调端点 URL';
COMMENT ON COLUMN skills.method          IS 'HTTP 方法: GET / POST';
COMMENT ON COLUMN skills.headers         IS '附加请求头 JSON 对象（如认证 Token）';
COMMENT ON COLUMN skills.timeout_ms      IS '请求超时毫秒数（默认 5000）';
COMMENT ON COLUMN skills.enabled         IS '是否启用（false 时不注册到 Agent）';
COMMENT ON COLUMN skills.created_at      IS '创建时间';
COMMENT ON COLUMN skills.updated_at      IS '更新时间';

CREATE INDEX IF NOT EXISTS idx_skills_tenant_id ON skills(tenant_id);
CREATE INDEX IF NOT EXISTS idx_skills_tenant_enabled ON skills(tenant_id, enabled);
