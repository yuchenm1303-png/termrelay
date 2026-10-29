import type { InjectionKey, Ref } from 'vue'
import { adminNavigationGroups, neutralWorkspacePaths, userNavigationGroups, type NavGroupSpec } from './navigation'

// Pages reveal async sections after BOTH real data and the active route settle.
export const workspaceRouteSettledKey: InjectionKey<Ref<boolean>> = Symbol('workspaceRouteSettled')

/* How a workspace navigation should feel:
 * - sibling:   same sidebar section (充值/订阅 → 我的订单) — a lateral tab-like slide.
 * - section:   a different sidebar section — vertical travel following the sidebar.
 * - workspace: user ↔ admin switch — a depth change, no direction.
 * direction is +1 when the target sits below/after the origin in the sidebar. */
export type RouteMotionKind = 'sibling' | 'section' | 'workspace'
export interface RouteMotion { kind: RouteMotionKind; direction: 1 | -1 }
export interface RouteMotionEndpoint { path: string; feature?: unknown }

type WorkspaceSide = 'user' | 'admin'

const isNeutral = (path: string) =>
  neutralWorkspacePaths.some((neutral) => path === neutral || path.startsWith(`${neutral}/`))
const sideOf = (path: string): WorkspaceSide => (path.startsWith('/admin') ? 'admin' : 'user')

function locate(groups: NavGroupSpec[], feature: unknown) {
  let order = 0
  for (const [group, spec] of groups.entries()) {
    const index = spec.features.indexOf(String(feature))
    if (index >= 0) return { group, order: order + index }
    order += spec.features.length
  }
  return null
}

export function resolveRouteMotion(to: RouteMotionEndpoint, from: RouteMotionEndpoint | null): RouteMotion {
  if (!from) return { kind: 'section', direction: 1 }
  // Neutral routes adopt the side of the route they are paired with.
  const fromSide = isNeutral(from.path) ? sideOf(to.path) : sideOf(from.path)
  const toSide = isNeutral(to.path) ? fromSide : sideOf(to.path)
  if (fromSide !== toSide) return { kind: 'workspace', direction: toSide === 'admin' ? 1 : -1 }

  const groups = toSide === 'admin' ? adminNavigationGroups : userNavigationGroups
  const origin = locate(groups, from.feature)
  const target = locate(groups, to.feature)
  if (!origin || !target) return { kind: 'section', direction: 1 }
  const direction = target.order >= origin.order ? 1 : -1
  return { kind: origin.group === target.group ? 'sibling' : 'section', direction }
}
