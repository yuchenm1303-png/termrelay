-- Align usage_logs with repository ON CONFLICT (request_id, api_key_id).
-- This migration intentionally uses CONCURRENTLY to avoid blocking the hot usage_logs table.

CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_logs_request_api_key_unique
    ON usage_logs (request_id, api_key_id);
