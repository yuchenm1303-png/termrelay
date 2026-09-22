/**
 * Shared shape and derivations for the per-request usage records served by
 * `GET /api/v1/usage` (backend DTO `dto.UsageLog`, handler/usage_handler.go).
 *
 * The list endpoint intentionally ships raw token buckets and the inbound
 * endpoint path only; it never ships a pre-summed `total_tokens` nor a generic
 * `endpoint` field. Every consumer must derive the total and the endpoint
 * label through the helpers below so the usage analytics pages cannot drift
 * from the API contract again.
 */
export interface UsageRecord {
  id?: number
  model?: string
  created_at?: string
  input_tokens?: number
  output_tokens?: number
  cache_creation_tokens?: number
  cache_read_tokens?: number
  cache_creation_5m_tokens?: number
  cache_creation_1h_tokens?: number
  inbound_endpoint?: string
  upstream_endpoint?: string
  total_cost?: number
  actual_cost?: number
  [key: string]: unknown
}

function toCount(value: unknown): number {
  const parsed = Number(value ?? 0)
  return Number.isFinite(parsed) ? parsed : 0
}

function toText(value: unknown): string {
  return typeof value === 'string' ? value.trim() : ''
}

/** Mirrors `service.UsageLog.TotalTokens()`: input + output + cache creation + cache read. */
export function usageTotalTokens(record: UsageRecord): number {
  return toCount(record.input_tokens)
    + toCount(record.output_tokens)
    + toCount(record.cache_creation_tokens)
    + toCount(record.cache_read_tokens)
}

/**
 * Client-facing endpoint path (`inbound_endpoint`), falling back to the
 * normalized upstream path when the inbound path was not recorded.
 * Returns an empty string when neither is available.
 */
export function usageEndpoint(record: UsageRecord): string {
  return toText(record.inbound_endpoint) || toText(record.upstream_endpoint)
}

/** Bucket label for the model / endpoint distribution panels. */
export function usageGroupLabel(record: UsageRecord, key: 'model' | 'endpoint'): string {
  if (key === 'endpoint') return usageEndpoint(record) || 'Unknown'
  return toText(record.model) || 'Unknown'
}
