-- ============================================================
-- Migration: 000003_conversations (DOWN)
-- 描述: 回滚对话模块表
-- ============================================================

DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;
