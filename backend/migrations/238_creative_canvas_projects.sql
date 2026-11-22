-- 238: User-owned creative canvas documents.
-- Keep media binaries in separate, private asset storage. Only structured node
-- graph metadata and prompts live in this table. Never store API credentials.
CREATE TABLE IF NOT EXISTS creative_canvas_projects (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    id VARCHAR(80) NOT NULL,
    title VARCHAR(90) NOT NULL,
    document JSONB NOT NULL,
    version BIGINT NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, id),
    CONSTRAINT creative_canvas_project_object CHECK (jsonb_typeof(document) = 'object'),
    CONSTRAINT creative_canvas_project_bytes CHECK (octet_length(document::text) <= 500000)
);

CREATE INDEX IF NOT EXISTS creative_canvas_projects_by_owner_updated
    ON creative_canvas_projects(user_id, updated_at DESC);

COMMENT ON TABLE creative_canvas_projects IS
    'Tenant-scoped creative node graph snapshots (no API keys or binary media).';

