-- Migration: 000010_flow_runs
-- Description: Flow 执行历史表
-- 每次试运行/触发执行落一条记录，node_results 保存全量节点执行快照（含 flow_version，definition 后续编辑不污染历史回放）

CREATE TABLE IF NOT EXISTS flow_runs (
    id           BIGSERIAL    PRIMARY KEY,
    uuid         UUID         NOT NULL UNIQUE,
    tenant_id    BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    flow_id      BIGINT       NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
    flow_version INT          NOT NULL DEFAULT 1,
    status       VARCHAR(20)  NOT NULL DEFAULT 'running',
    trigger_type VARCHAR(20)  NOT NULL DEFAULT 'manual',
    input        JSONB        NOT NULL DEFAULT '{}',
    output       TEXT         NOT NULL DEFAULT '',
    node_results JSONB        NOT NULL DEFAULT '[]',
    error        TEXT         NOT NULL DEFAULT '',
    total_ms     INT          NOT NULL DEFAULT 0,
    token_count  INT          NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE flow_runs                  IS 'Flow 执行历史表';
COMMENT ON COLUMN flow_runs.id              IS '自增主键';
COMMENT ON COLUMN flow_runs.uuid            IS '执行记录 UUID，对外展示使用';
COMMENT ON COLUMN flow_runs.tenant_id       IS '所属租户 ID（级联删除）';
COMMENT ON COLUMN flow_runs.flow_id         IS '关联 Flow ID（级联删除）';
COMMENT ON COLUMN flow_runs.flow_version    IS '执行时的 Flow 版本快照';
COMMENT ON COLUMN flow_runs.status          IS '执行状态：running / success / failed / timeout';
COMMENT ON COLUMN flow_runs.trigger_type    IS '触发方式：manual（一期仅手动试运行）';
COMMENT ON COLUMN flow_runs.input           IS '执行输入 JSON（message + variables）';
COMMENT ON COLUMN flow_runs.output          IS '最终输出文本';
COMMENT ON COLUMN flow_runs.node_results    IS '节点执行结果快照数组';
COMMENT ON COLUMN flow_runs.error           IS '失败原因';
COMMENT ON COLUMN flow_runs.total_ms        IS '总耗时（毫秒）';
COMMENT ON COLUMN flow_runs.token_count     IS 'Token 消耗总量';
COMMENT ON COLUMN flow_runs.created_at      IS '创建时间';
COMMENT ON COLUMN flow_runs.updated_at      IS '更新时间';

CREATE INDEX IF NOT EXISTS idx_flow_runs_tenant_id ON flow_runs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_flow_runs_flow_created ON flow_runs(flow_id, created_at DESC);
