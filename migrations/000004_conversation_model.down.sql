-- ============================================================
-- Migration: 000004_conversation_model (rollback)
-- ============================================================

ALTER TABLE conversations DROP COLUMN IF EXISTS model;
