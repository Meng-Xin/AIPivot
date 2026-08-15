-- ============================================================
-- Migration: 000009_widget_feedback_suggested
-- 描述: Widget 第一波 —— 消息满意度评分 + 引导问答
-- 说明:
--   1) messages 增加 rating / rating_feedback
--      - rating: up / down / 空，二选一锁定语义（第一波不支持撤销）
--      - rating_feedback: 负评文字反馈（仅 rating=down 时可能填）
--      - 部分索引加速 Analytics 聚合，仅覆盖已评分消息
--   2) knowledge_bases 增加 suggested_questions
--      - 引导问答列表（JSON 数组），运营在管理后台维护
--      - 建会话时返回给 Widget 渲染首屏快捷回复 chip
-- ============================================================

-- ---------- messages 满意度评分 ----------
ALTER TABLE messages
    ADD COLUMN rating          VARCHAR(8) NOT NULL DEFAULT '' CHECK (rating IN ('', 'up', 'down')),
    ADD COLUMN rating_feedback TEXT       NOT NULL DEFAULT '';

COMMENT ON COLUMN messages.rating          IS '访客评分：up=满意 / down=不满意 / 空=未评分';
COMMENT ON COLUMN messages.rating_feedback IS '负评文字反馈（仅 rating=down 时可能有值）';

-- 部分索引：仅对已评分消息建索引，加速 Analytics 聚合
CREATE INDEX idx_messages_tenant_rating ON messages (tenant_id, rating) WHERE rating <> '';

-- ---------- knowledge_bases 引导问答 ----------
ALTER TABLE knowledge_bases
    ADD COLUMN suggested_questions JSONB NOT NULL DEFAULT '[]';

COMMENT ON COLUMN knowledge_bases.suggested_questions IS '引导问答 / 快捷回复问题列表（JSON 数组，最多 6 条，每条 ≤ 100 字）';
