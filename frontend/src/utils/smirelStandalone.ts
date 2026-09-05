import type { PublicSettings } from '@/types'

export const SMIREL_SITE_NAME = 'Smirel API'
export const SMIREL_SITE_LOGO = '/smirel-mark.svg'
export const SMIREL_SITE_SUBTITLE = 'Unified AI API Gateway'
export const SMIREL_API_BASE_URL = 'https://api.smirel.com/v1'

export function isStandaloneSmirelFrontend(): boolean {
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
