-- Migration: 000010_flow_runs (down)
DROP INDEX IF EXISTS idx_flow_runs_flow_created;
DROP INDEX IF EXISTS idx_flow_runs_tenant_id;
DROP TABLE IF EXISTS flow_runs;
