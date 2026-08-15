-- 回滚: 000008_widget_public_key

DROP INDEX IF EXISTS idx_conv_external_user;
ALTER TABLE conversations DROP COLUMN IF EXISTS external_user_id;

DROP INDEX IF EXISTS idx_api_keys_key_type;
ALTER TABLE api_keys
    DROP COLUMN IF EXISTS knowledge_base_id,
    DROP COLUMN IF EXISTS allowed_origins,
    DROP COLUMN IF EXISTS key_type;
