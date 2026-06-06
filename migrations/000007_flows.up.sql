-- Migration: 000007_flows
-- Description: 可视化 Flow 流程定义表
-- 每个租户可维护多个流程，definition 存储前端画布节点/连线/配置 JSON

CREATE TABLE IF NOT EXISTS flows (
    id          BIGSERIAL    PRIMARY KEY,
    uuid        UUID         NOT NULL UNIQUE,
    tenant_id   BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        VARCHAR(120) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    definition  JSONB        NOT NULL DEFAULT '{}',
    status      VARCHAR(20)  NOT NULL DEFAULT 'draft',
    version     INT          NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

COMMENT ON TABLE flows                 IS '租户可视化 Flow 定义表';
COMMENT ON COLUMN flows.id             IS '自增主键';
COMMENT ON COLUMN flows.uuid           IS '流程 UUID，对外展示使用';
COMMENT ON COLUMN flows.tenant_id      IS '所属租户 ID（级联删除）';
COMMENT ON COLUMN flows.name           IS '流程名称，租户内唯一';
COMMENT ON COLUMN flows.description    IS '流程描述';
COMMENT ON COLUMN flows.definition     IS 'Flow 画布定义 JSON，包含节点、连线、视图配置';
COMMENT ON COLUMN flows.status         IS '流程状态：draft / published / archived';
COMMENT ON COLUMN flows.version        IS '流程版本，每次更新递增';
COMMENT ON COLUMN flows.created_at     IS '创建时间';
COMMENT ON COLUMN flows.updated_at     IS '更新时间';

CREATE INDEX IF NOT EXISTS idx_flows_tenant_id ON flows(tenant_id);
CREATE INDEX IF NOT EXISTS idx_flows_tenant_status ON flows(tenant_id, status);
