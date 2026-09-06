import { describe, expect, it } from 'vitest'
import { createMemoryHistory, createRouter, type RouteRecordRaw } from 'vue-router'
import {
  buildWorkspaceNavigation,
  type WorkspaceNavigationCapabilities,
} from '../smirelWorkspaceNavigation'

const View = { template: '<div />' }

function makeRouter(routeNames: string[]) {
  const routes: RouteRecordRaw[] = routeNames.map((name) => ({
    path: `/${name.toLowerCase()}`,
    name,
    component: View,
  }))

  return createRouter({
    history: createMemoryHistory(),
    routes,
  })
}

function capabilities(overrides: Partial<WorkspaceNavigationCapabilities> = {}): WorkspaceNavigationCapabilities {
  return {
    isAdmin: false,
    isSimpleMode: false,
    paymentEnabled: true,
    riskControlEnabled: true,
    modelPlazaEnabled: true,
    availableChannelsEnabled: true,
    channelMonitorEnabled: true,
    affiliateEnabled: true,
    ...overrides,
  }
}

function labels(sections: ReturnType<typeof buildWorkspaceNavigation>): string[] {
  return sections.flatMap((section) => section.items.map((item) => item.label))
}

describe('buildWorkspaceNavigation', () => {
  it('only exposes user features backed by registered router routes', () => {
    const router = makeRouter(['Dashboard', 'Keys', 'ModelPlaza', 'Usage', 'Profile'])
    const sections = buildWorkspaceNavigation(router, capabilities())
    const visibleLabels = labels(sections)

    expect(visibleLabels).toEqual(['概览', 'API Keys', '模型与价格', '用量与日志', '账户设置'])
    expect(visibleLabels).not.toContain('我的订单')

    const modelItem = sections.flatMap((section) => section.items).find((item) => item.label === '模型与价格')
    expect(modelItem?.to).toContain('embedded=1')
  })

  it('respects simple-mode and feature gates instead of presenting unavailable admin actions', () => {
    const router = makeRouter([
      'AdminDashboard',
      'AdminUsers',
      'AdminUsage',
      'AdminAccounts',
      'AdminGroups',
      'AdminChannels',
      'AdminChannelMonitor',
      'AdminProxies',
      'AdminSubscriptions',
      'AdminPaymentDashboard',
      'AdminOrders',
      'AdminPaymentPlans',
      'AdminRedeem',
      'AdminPromoCodes',
      'AdminOps',
      'AdminAuditLogs',
      'AdminAnnouncements',
      'AdminRiskControl',
      'AdminSettings',
    ])

    const sections = buildWorkspaceNavigation(router, capabilities({
      isAdmin: true,
      isSimpleMode: true,
      paymentEnabled: false,
      riskControlEnabled: false,
    }))
    const visibleLabels = labels(sections)

    expect(visibleLabels).toContain('概览')
    expect(visibleLabels).toContain('上游账户')
    expect(visibleLabels).toContain('模型路由')
    expect(visibleLabels).toContain('运行状态')
    expect(visibleLabels).toContain('平台设置')

    expect(visibleLabels).not.toContain('路由策略')
    expect(visibleLabels).not.toContain('套餐管理')
    expect(visibleLabels).not.toContain('兑换码')
    expect(visibleLabels).not.toContain('营收概览')
    expect(visibleLabels).not.toContain('订单管理')
    expect(visibleLabels).not.toContain('套餐配置')
    expect(visibleLabels).not.toContain('风控规则')
  })

  it('drops empty navigation sections when none of their routes exist', () => {
    const router = makeRouter(['AdminDashboard'])
    const sections = buildWorkspaceNavigation(router, capabilities({ isAdmin: true }))

    expect(sections).toHaveLength(1)
    expect(sections[0]?.label).toBe('控制台')
    expect(labels(sections)).toEqual(['概览'])
  })
})
