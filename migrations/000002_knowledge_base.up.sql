-- ============================================================
-- Migration: 000002_knowledge_base
-- 描述: 知识库模块（knowledge_bases / documents / document_chunks）
-- 说明: 支持 RAG 场景的知识管理，使用 pgvector 存储向量
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "vector";

-- -----------------------------------------------------------
-- knowledge_bases: 知识库表
-- -----------------------------------------------------------
CREATE TABLE knowledge_bases (
    id          BIGSERIAL    PRIMARY KEY,
    uuid        UUID         NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    tenant_id   BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    model       VARCHAR(100) NOT NULL DEFAULT 'text-embedding-3-small',
    dimension   INT          NOT NULL DEFAULT 1536,
    status      VARCHAR(20)  NOT NULL DEFAULT 'active',
    doc_count   INT          NOT NULL DEFAULT 0,
    chunk_count INT          NOT NULL DEFAULT 0,
    settings    JSONB        NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  knowledge_bases                IS '知识库表：每个租户可创建多个知识库，每个知识库使用独立的 Embedding 模型配置';
COMMENT ON COLUMN knowledge_bases.id             IS '自增主键';
COMMENT ON COLUMN knowledge_bases.uuid           IS '对外暴露的唯一标识（UUID v4）';
COMMENT ON COLUMN knowledge_bases.tenant_id      IS '所属租户 ID（级联删除）';
COMMENT ON COLUMN knowledge_bases.name           IS '知识库名称';
COMMENT ON COLUMN knowledge_bases.description    IS '知识库描述';
COMMENT ON COLUMN knowledge_bases.model          IS 'Embedding 模型名称';
COMMENT ON COLUMN knowledge_bases.dimension      IS '向量维度（与 Embedding 模型对应）';
COMMENT ON COLUMN knowledge_bases.status         IS '状态: active / archived';
COMMENT ON COLUMN knowledge_bases.doc_count      IS '文档数量（冗余计数，异步更新）';
COMMENT ON COLUMN knowledge_bases.chunk_count    IS '切块数量（冗余计数，异步更新）';
COMMENT ON COLUMN knowledge_bases.settings       IS '知识库配置（chunk_size / overlap / 检索策略等）';
COMMENT ON COLUMN knowledge_bases.created_at     IS '创建时间';
COMMENT ON COLUMN knowledge_bases.updated_at     IS '更新时间';

CREATE INDEX idx_kb_tenant_id ON knowledge_bases(tenant_id);
CREATE INDEX idx_kb_status    ON knowledge_bases(status);

-- -----------------------------------------------------------
-- documents: 文档表
-- -----------------------------------------------------------
CREATE TABLE documents (
    id              BIGSERIAL    PRIMARY KEY,
    uuid            UUID         NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    knowledge_base_id BIGINT     NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    tenant_id       BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            VARCHAR(500) NOT NULL,
    content_type    VARCHAR(50)  NOT NULL DEFAULT 'text/plain',
    file_size       BIGINT       NOT NULL DEFAULT 0,
    file_path       TEXT         NOT NULL DEFAULT '',
    chunk_count     INT          NOT NULL DEFAULT 0,
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending',
    error_msg       TEXT         NOT NULL DEFAULT '',
    metadata        JSONB        NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  documents                    IS '文档表：上传到知识库的原始文档，经过解析→切块→Embedding 后生成 document_chunks';
COMMENT ON COLUMN documents.id                 IS '自增主键';
COMMENT ON COLUMN documents.uuid               IS '对外暴露的唯一标识（UUID v4）';
COMMENT ON COLUMN documents.knowledge_base_id  IS '所属知识库 ID（级联删除）';
COMMENT ON COLUMN documents.tenant_id          IS '所属租户 ID（级联删除，冗余字段加速查询）';
COMMENT ON COLUMN documents.name               IS '文档名称（原始文件名）';
COMMENT ON COLUMN documents.content_type       IS '文档 MIME 类型：text/plain / application/pdf / text/markdown 等';
COMMENT ON COLUMN documents.file_size          IS '文件大小（字节）';
COMMENT ON COLUMN documents.file_path          IS '文件存储路径（MinIO/OSS）';
COMMENT ON COLUMN documents.chunk_count        IS '切块数量';
COMMENT ON COLUMN documents.status             IS '处理状态: pending / processing / completed / failed';
COMMENT ON COLUMN documents.error_msg          IS '处理失败时的错误信息';
COMMENT ON COLUMN documents.metadata           IS '文档元数据（标签、来源等）';
COMMENT ON COLUMN documents.created_at         IS '创建时间';
COMMENT ON COLUMN documents.updated_at         IS '更新时间';

CREATE INDEX idx_documents_kb_id     ON documents(knowledge_base_id);
CREATE INDEX idx_documents_tenant_id ON documents(tenant_id);
CREATE INDEX idx_documents_status    ON documents(status);

-- -----------------------------------------------------------
-- document_chunks: 文档切块表（含 pgvector 向量）
-- -----------------------------------------------------------
CREATE TABLE document_chunks (
    id              BIGSERIAL    PRIMARY KEY,
    uuid            UUID         NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    document_id     BIGINT       NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    knowledge_base_id BIGINT     NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    tenant_id       BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    chunk_index     INT          NOT NULL DEFAULT 0,
    content         TEXT         NOT NULL,
    token_count     INT          NOT NULL DEFAULT 0,
    embedding       vector(1536),
    metadata        JSONB        NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  document_chunks                    IS '文档切块表：文档被切块后每个片段对应一行，embedding 列存储向量供 pgvector 检索';
COMMENT ON COLUMN document_chunks.id                 IS '自增主键';
COMMENT ON COLUMN document_chunks.uuid               IS '对外暴露的唯一标识（UUID v4）';
COMMENT ON COLUMN document_chunks.document_id        IS '所属文档 ID（级联删除）';
COMMENT ON COLUMN document_chunks.knowledge_base_id  IS '所属知识库 ID（冗余字段，加速检索时按知识库过滤）';
COMMENT ON COLUMN document_chunks.tenant_id          IS '所属租户 ID（冗余字段，加速租户隔离查询）';
COMMENT ON COLUMN document_chunks.chunk_index        IS '切块在文档中的顺序索引（从 0 开始）';
COMMENT ON COLUMN document_chunks.content            IS '切块文本内容';
COMMENT ON COLUMN document_chunks.token_count        IS 'Token 数量（基于 Embedding 模型 tokenizer 计算）';
COMMENT ON COLUMN document_chunks.embedding          IS '向量表示（pgvector, 默认 1536 维）';
COMMENT ON COLUMN document_chunks.metadata           IS '切块元数据（页码、标题层级等）';
COMMENT ON COLUMN document_chunks.created_at         IS '创建时间';

CREATE INDEX idx_chunks_document_id ON document_chunks(document_id);
CREATE INDEX idx_chunks_kb_id       ON document_chunks(knowledge_base_id);
CREATE INDEX idx_chunks_tenant_id   ON document_chunks(tenant_id);

-- HNSW 向量索引：cosine 相似度，适用于 RAG 检索场景
CREATE INDEX idx_chunks_embedding ON document_chunks
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);
