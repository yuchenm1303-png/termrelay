import type { InjectionKey, Ref } from 'vue'

// Pages reveal async sections after BOTH real data and the active route settle.
export const workspaceRouteSettledKey: InjectionKey<Ref<boolean>> = Symbol('workspaceRouteSettled')
