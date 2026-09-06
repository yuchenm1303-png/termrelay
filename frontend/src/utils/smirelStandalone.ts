import type { PublicSettings } from '@/types'

const SMIREL_PUBLIC_BASE = String(import.meta.env.BASE_URL || '/').replace(/\/?$/, '/')

export const SMIREL_SITE_NAME = 'Smirel API'
export const SMIREL_SITE_LOGO = `${SMIREL_PUBLIC_BASE}smirel-logo.png`
export const SMIREL_SITE_SUBTITLE = 'Unified AI API Gateway'
export const SMIREL_API_BASE_URL = 'https://api.smirel.com/v1'

const API_KEY_HIDDEN_COLUMNS_KEY = 'api-key-hidden-columns'
const SMIREL_KEY_COLUMNS_POLICY_KEY = 'smirel-api-key-columns-v1'
const UPSTREAM_DEFAULT_HIDDEN_KEY_COLUMNS = ['id', 'rate_limit', 'last_used_at', 'last_used_ip']
const SMIREL_DEFAULT_HIDDEN_KEY_COLUMNS = ['group', 'current_concurrency']

export function isStandaloneSmirelFrontend(): boolean {
  if (import.meta.env.VITE_SMIREL_STANDALONE === 'true') return true
  if (typeof window === 'undefined') return false
  const host = window.location.hostname.toLowerCase()
  return host === 'relay.smirel.com' || host.endsWith('.vercel.app')
}

/**
 * Apply only customer-facing policy owned by the standalone Smirel frontend.
 * Backend feature flags and auth capabilities remain authoritative.
 *
 * Keeping this transformation in one place prevents direct `/settings/public`
 * consumers from bypassing the brand/API-endpoint policy used by the app store.
 */
export function applyStandaloneSmirelPublicSettings(
  settings: PublicSettings,
): PublicSettings {
  if (!isStandaloneSmirelFrontend()) return settings

  return {
    ...settings,
    site_name: SMIREL_SITE_NAME,
    site_logo: SMIREL_SITE_LOGO,
    site_subtitle: SMIREL_SITE_SUBTITLE,
    api_base_url: SMIREL_API_BASE_URL,
    hide_ccs_import_button: true,
    home_content: '',
    compact_home_enabled: false,
  }
}

/**
 * Migrate first-party UI preferences once without taking control away from the
 * user afterwards. The upstream key table foregrounds scheduler internals
 * (`group`, live concurrency). Smirel keeps those columns available in Column
 * Settings but does not show them by default on the customer-facing console.
 */
export function applyStandaloneSmirelPreferences(): void {
  if (!isStandaloneSmirelFrontend()) return

  try {
    if (localStorage.getItem(SMIREL_KEY_COLUMNS_POLICY_KEY) === '1') return

    const stored = localStorage.getItem(API_KEY_HIDDEN_COLUMNS_KEY)
    let hiddenColumns: string[]

    if (stored) {
      const parsed = JSON.parse(stored) as unknown
      hiddenColumns = Array.isArray(parsed)
        ? parsed.filter((value): value is string => typeof value === 'string')
        : []
    } else {
      hiddenColumns = [...UPSTREAM_DEFAULT_HIDDEN_KEY_COLUMNS]
    }

    for (const column of SMIREL_DEFAULT_HIDDEN_KEY_COLUMNS) {
      if (!hiddenColumns.includes(column)) hiddenColumns.push(column)
    }

    localStorage.setItem(API_KEY_HIDDEN_COLUMNS_KEY, JSON.stringify(hiddenColumns))
    localStorage.setItem(SMIREL_KEY_COLUMNS_POLICY_KEY, '1')
  } catch {
    // Storage can be unavailable in hardened/private browsing environments.
    // The table will simply fall back to its component defaults in that case.
  }
}
