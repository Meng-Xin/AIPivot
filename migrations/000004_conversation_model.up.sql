-- ============================================================
-- Migration: 000004_conversation_model
-- 描述: 为会话表增加 model 字段，支持 per-conversation 模型选择
-- 说明: LLM Gateway 多模型路由功能，允许每个会话指定不同的聊天模型
-- ============================================================

ALTER TABLE conversations
    ADD COLUMN model VARCHAR(100) NOT NULL DEFAULT '';

COMMENT ON COLUMN conversations.model IS '聊天模型标识（如 gpt-4o），空字符串表示使用系统默认模型';
