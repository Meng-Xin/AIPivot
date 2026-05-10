-- ============================================================
-- Migration: 000002_knowledge_base (DOWN)
-- 描述: 回滚知识库模块表
-- ============================================================

DROP TABLE IF EXISTS document_chunks;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS knowledge_bases;
