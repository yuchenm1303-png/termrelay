/**
 * Canonical layout components.
 *
 * Public/utility routes use SmirelPortalShell, authenticated routes use
 * AppLayout, and auth routes use AuthLayout. Historical App/Console/Glass
 * generations are retired.
 */

export { default as AppLayout } from './AppLayout.vue'
export { default as AuthLayout } from './AuthLayout.vue'
export { default as TablePageLayout } from './TablePageLayout.vue'
export { default as SmirelPortalShell } from './SmirelPortalShell.vue'
export { default as SmirelWorkspaceSidebar } from './SmirelWorkspaceSidebar.vue'
export { default as SmirelWorkspaceTopbar } from './SmirelWorkspaceTopbar.vue'
