/**
 * Canonical layout components.
 *
 * Public portal routes own their first-level composition, authenticated routes
 * use AppLayout, and auth routes use AuthLayout. Historical AppHeader/AppSidebar
 * and SmirelConsole generations have been retired.
 */

export { default as AppLayout } from './AppLayout.vue'
export { default as AuthLayout } from './AuthLayout.vue'
export { default as TablePageLayout } from './TablePageLayout.vue'
export { default as SmirelWorkspaceSidebar } from './SmirelWorkspaceSidebar.vue'
export { default as SmirelWorkspaceTopbar } from './SmirelWorkspaceTopbar.vue'
