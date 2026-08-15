-- 回滚 000009_widget_feedback_suggested

DROP INDEX IF EXISTS idx_messages_tenant_rating;

ALTER TABLE messages
    DROP COLUMN IF EXISTS rating_feedback,
    DROP COLUMN IF EXISTS rating;

ALTER TABLE knowledge_bases
    DROP COLUMN IF EXISTS suggested_questions;
