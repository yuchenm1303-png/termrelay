-- 237: Keep the display-only /v1/models configuration separate from the
-- admission allowlist. Migration 236 renamed models_list_config into
-- model_allowlist, but Ent and the group admin UI still persist the former.
-- Fresh installations consequently failed any INSERT/SELECT against groups.
--
-- For databases which only have the renamed column, restore the old column
-- with its prior data rather than silently clearing a configured model list.
-- If both columns already exist, never overwrite either configuration.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_attribute
         WHERE attrelid = 'groups'::regclass
           AND attname = 'models_list_config'
           AND NOT attisdropped
    ) THEN
        ALTER TABLE groups
            ADD COLUMN models_list_config JSONB NOT NULL DEFAULT '{}'::jsonb;
        UPDATE groups
           SET models_list_config = COALESCE(model_allowlist, '{}'::jsonb);
    END IF;
END
$$;

ALTER TABLE groups
    ALTER COLUMN models_list_config SET DEFAULT '{}'::jsonb,
    ALTER COLUMN models_list_config SET NOT NULL;

COMMENT ON COLUMN groups.models_list_config IS
    'Display-only custom /v1/models list; does not grant model access';
