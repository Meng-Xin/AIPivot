-- ============================================================
-- Migration: 000001_init_tenants
-- 描述: 多租户基础表（tenants / users / api_keys）
-- 说明: 系统采用多租户隔离架构，所有业务数据通过 tenant_id 关联
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- -----------------------------------------------------------
-- tenants: 租户表
-- -----------------------------------------------------------
CREATE TABLE tenants (
    id          BIGSERIAL    PRIMARY KEY,
    uuid        UUID         NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    plan        VARCHAR(50)  NOT NULL DEFAULT 'free',
    status      VARCHAR(20)  NOT NULL DEFAULT 'active',
    settings    JSONB        NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  tenants              IS '租户表：每个 B2B 客户对应一个租户，是数据隔离的最小单元';
COMMENT ON COLUMN tenants.id           IS '自增主键';
COMMENT ON COLUMN tenants.uuid         IS '对外暴露的唯一标识（UUID v4）';
COMMENT ON COLUMN tenants.name         IS '租户名称';
COMMENT ON COLUMN tenants.slug         IS 'URL 友好标识符，全局唯一';
COMMENT ON COLUMN tenants.plan         IS '订阅计划: free / pro / enterprise';
COMMENT ON COLUMN tenants.status       IS '状态: active / suspended';
COMMENT ON COLUMN tenants.settings     IS '租户级配置（JSON 格式）';
COMMENT ON COLUMN tenants.created_at   IS '创建时间';
COMMENT ON COLUMN tenants.updated_at   IS '更新时间';

CREATE INDEX idx_tenants_slug   ON tenants(slug);
CREATE INDEX idx_tenants_status ON tenants(status);

-- -----------------------------------------------------------
-- users: 用户表
-- -----------------------------------------------------------
CREATE TABLE users (
    id          BIGSERIAL    PRIMARY KEY,
    uuid        UUID         NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    tenant_id   BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email       VARCHAR(255) NOT NULL,
    nick_name   VARCHAR(100) NOT NULL DEFAULT '',
    password    VARCHAR(255) NOT NULL,
    role        VARCHAR(50)  NOT NULL DEFAULT 'member',
    status      VARCHAR(20)  NOT NULL DEFAULT 'active',
    last_login  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, email)
);

COMMENT ON TABLE  users                IS '用户表：平台用户，归属于某个租户；同一租户内 email 唯一';
COMMENT ON COLUMN users.id             IS '自增主键';
COMMENT ON COLUMN users.uuid           IS '对外暴露的唯一标识（UUID v4）';
COMMENT ON COLUMN users.tenant_id      IS '所属租户 ID（级联删除）';
COMMENT ON COLUMN users.email          IS '登录邮箱（租户内唯一）';
COMMENT ON COLUMN users.nick_name      IS '用户昵称';
COMMENT ON COLUMN users.password       IS 'bcrypt 哈希密码';
COMMENT ON COLUMN users.role           IS '角色: admin / member';
COMMENT ON COLUMN users.status         IS '状态: active / disabled';
COMMENT ON COLUMN users.last_login     IS '最近登录时间';
COMMENT ON COLUMN users.created_at     IS '创建时间';
COMMENT ON COLUMN users.updated_at     IS '更新时间';

CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email     ON users(email);

-- -----------------------------------------------------------
-- api_keys: API 密钥表
-- -----------------------------------------------------------
CREATE TABLE api_keys (
    id          BIGSERIAL    PRIMARY KEY,
    tenant_id   BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    key_hash    VARCHAR(255) NOT NULL UNIQUE,
    key_prefix  VARCHAR(10)  NOT NULL,
    scopes      JSONB        NOT NULL DEFAULT '["chat"]',
    status      VARCHAR(20)  NOT NULL DEFAULT 'active',
    last_used   TIMESTAMPTZ,
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  api_keys             IS 'API 密钥表：租户的程序化访问凭证，仅存储哈希值，原始密钥只在创建时返回一次';
COMMENT ON COLUMN api_keys.id          IS '自增主键';
COMMENT ON COLUMN api_keys.tenant_id   IS '所属租户 ID（级联删除）';
COMMENT ON COLUMN api_keys.name        IS '密钥名称（便于管理识别）';
COMMENT ON COLUMN api_keys.key_hash    IS '密钥 SHA-256 哈希（唯一）';
COMMENT ON COLUMN api_keys.key_prefix  IS '密钥前缀（列表展示用，如 sk-abc...）';
COMMENT ON COLUMN api_keys.scopes      IS '权限范围（JSON 数组）';
COMMENT ON COLUMN api_keys.status      IS '状态: active / revoked';
COMMENT ON COLUMN api_keys.last_used   IS '最近使用时间';
COMMENT ON COLUMN api_keys.expires_at  IS '过期时间（空表示永不过期）';
COMMENT ON COLUMN api_keys.created_at  IS '创建时间';
COMMENT ON COLUMN api_keys.updated_at  IS '更新时间';

CREATE INDEX idx_api_keys_tenant_id ON api_keys(tenant_id);
CREATE INDEX idx_api_keys_key_hash  ON api_keys(key_hash);
CREATE INDEX idx_api_keys_status    ON api_keys(status);

-- -----------------------------------------------------------
-- 开发环境种子数据
-- -----------------------------------------------------------
INSERT INTO tenants (name, slug, plan, status)
VALUES ('Default Tenant', 'default', 'free', 'active');
