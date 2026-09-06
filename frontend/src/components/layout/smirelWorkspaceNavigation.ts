import type { Router } from 'vue-router'

export interface WorkspaceNavigationCapabilities {
  isAdmin: boolean
  isSimpleMode: boolean
  paymentEnabled: boolean
  riskControlEnabled: boolean
  modelPlazaEnabled: boolean
  availableChannelsEnabled: boolean
  channelMonitorEnabled: boolean
  affiliateEnabled: boolean
}

export interface WorkspaceNavigationItem {
  label: string
  to: string
  icon: string
}

export interface WorkspaceNavigationSection {
  label: string
  items: WorkspaceNavigationItem[]
}

type FeatureGate =
  | 'payment'
  | 'riskControl'
  | 'modelPlaza'
  | 'availableChannels'
  | 'channelMonitor'
  | 'affiliate'

interface NavigationDefinition {
  label: string
  routeName: string
  icon: string
  query?: Record<string, string>
  gate?: FeatureGate
  hideInSimpleMode?: boolean
}

interface NavigationSectionDefinition {
  label: string
  items: NavigationDefinition[]
}

const icons = {
  overview: 'M3.75 3.75h6.5v6.5h-6.5v-6.5zm10 0h6.5v6.5h-6.5v-6.5zm-10 10h6.5v6.5h-6.5v-6.5zm10 0h6.5v6.5h-6.5v-6.5z',
  users: 'M16.5 18.75a6 6 0 00-9 0M12 13.5a3.75 3.75 0 100-7.5 3.75 3.75 0 000 7.5zm7.5 5.25a4.5 4.5 0 00-3.2-4.31m.95-7.29a3 3 0 010 5.7',
  usage: 'M4 19V10m5.3 9V5m5.4 14v-7m5.3 7V8',
  accounts: 'M5 7.5h14M5 12h14M5 16.5h9M3.5 4.5h17v15h-17v-15z',
  routing: 'M5 5h5v5H5V5zm9 9h5v5h-5v-5zM10 7.5h3a3 3 0 013 3V14m-8 0v-1.5A2.5 2.5 0 0110.5 10H14',
  channels: 'M4 6.5h16M4 12h16M4 17.5h16M7 4v5m8 1v5m-5 0v5',
  health: 'M3 12h4l2-5 4 10 2.5-6H21',
  billing: 'M4 6.5h16v11H4v-11zm0 3.5h16M7 14h4',
  operations: 'M4 17l4-5 3 2 5-7 4 3M4 20h16',
  audit: 'M6 3.75h9l3 3V20.25H6V3.75zm8.5 0v3.5H18M9 11h6m-6 3h6m-6 3h4',
  settings: 'M12 8.75A3.25 3.25 0 1112 15.25 3.25 3.25 0 0112 8.75zm0-5.25v2m0 12.5v2m8.5-8.5h-2m-13 0h-2m14.51-6.01l-1.42 1.42M7.41 16.59l-1.42 1.42m12.02 0l-1.42-1.42M7.41 7.41L5.99 5.99',
  key: 'M15.5 7.5a4.5 4.5 0 11-8.2 2.6L3.5 14v3h3v3h3l4.2-4.2M15.5 7.5h.01',
  models: 'M12 3l8 4.5-8 4.5-8-4.5L12 3zm-8 9l8 4.5 8-4.5m-16 4.5L12 21l8-4.5',
  profile: 'M12 12a4 4 0 100-8 4 4 0 000 8zm-7 8a7 7 0 0114 0',
  network: 'M4 8.5h16M4 15.5h16M8 5l-4 3.5L8 12m8 0l4 3.5L16 19',
  announcement: 'M5 10.5v3h3l7 4V6.5l-7 4H5zm10 0h2.5a2 2 0 010 4H15',
  redeem: 'M4 8h16v10H4V8zm2-3h12v3H6V5zm6 3v10M8.5 12h1M14.5 12h1',
  promo: 'M4 7.5h16v9H4v-9zm4 0v-2h8v2M8 12h8',
}

const adminDefinitions: NavigationSectionDefinition[] = [
  {
    label: '控制台',
    items: [
      { label: '概览', routeName: 'AdminDashboard', icon: icons.overview },
      { label: '用户与权限', routeName: 'AdminUsers', icon: icons.users },
      { label: '平台用量', routeName: 'AdminUsage', icon: icons.usage },
    ],
  },
  {
    label: '路由与上游',
    items: [
      { label: '上游账户', routeName: 'AdminAccounts', icon: icons.accounts },
      { label: '路由策略', routeName: 'AdminGroups', icon: icons.routing, hideInSimpleMode: true },
      { label: '模型路由', routeName: 'AdminChannels', icon: icons.channels },
      { label: '路由健康', routeName: 'AdminChannelMonitor', icon: icons.health },
      { label: '网络出口', routeName: 'AdminProxies', icon: icons.network },
    ],
  },
  {
    label: '商业',
    items: [
      { label: '套餐管理', routeName: 'AdminSubscriptions', icon: icons.billing, hideInSimpleMode: true },
      { label: '营收概览', routeName: 'AdminPaymentDashboard', icon: icons.usage, gate: 'payment' },
      { label: '订单管理', routeName: 'AdminOrders', icon: icons.billing, gate: 'payment' },
      { label: '套餐配置', routeName: 'AdminPaymentPlans', icon: icons.settings, gate: 'payment' },
      { label: '兑换码', routeName: 'AdminRedeem', icon: icons.redeem, hideInSimpleMode: true },
      { label: '优惠码', routeName: 'AdminPromoCodes', icon: icons.promo },
    ],
  },
  {
    label: '系统',
    items: [
      { label: '运行状态', routeName: 'AdminOps', icon: icons.operations },
      { label: '审计日志', routeName: 'AdminAuditLogs', icon: icons.audit },
      { label: '公告管理', routeName: 'AdminAnnouncements', icon: icons.announcement },
      { label: '风控规则', routeName: 'AdminRiskControl', icon: icons.health, gate: 'riskControl' },
      { label: '平台设置', routeName: 'AdminSettings', icon: icons.settings },
    ],
  },
]

const userDefinitions: NavigationSectionDefinition[] = [
  {
    label: '工作区',
    items: [
      { label: '概览', routeName: 'Dashboard', icon: icons.overview },
      { label: 'API Keys', routeName: 'Keys', icon: icons.key },
      { label: '模型与价格', routeName: 'ModelPlaza', icon: icons.models, gate: 'modelPlaza', query: { embedded: '1' } },
      { label: '用量与日志', routeName: 'Usage', icon: icons.usage },
      { label: '可用模型', routeName: 'UserAvailableChannels', icon: icons.channels, gate: 'availableChannels' },
      { label: '服务状态', routeName: 'ChannelStatus', icon: icons.health, gate: 'channelMonitor' },
      { label: '批量生图指南', routeName: 'BatchImageGuide', icon: icons.models },
    ],
  },
  {
    label: '账单',
    items: [
      { label: '套餐与额度', routeName: 'Subscriptions', icon: icons.billing, hideInSimpleMode: true },
      { label: '充值与账单', routeName: 'PurchaseSubscription', icon: icons.billing, gate: 'payment' },
      { label: '我的订单', routeName: 'OrderList', icon: icons.audit, gate: 'payment' },
      { label: '兑换码', routeName: 'Redeem', icon: icons.redeem, hideInSimpleMode: true },
      { label: '邀请返利', routeName: 'Affiliate', icon: icons.users, gate: 'affiliate' },
    ],
  },
  {
    label: '账户',
    items: [
      { label: '账户设置', routeName: 'Profile', icon: icons.profile },
    ],
  },
]

function gateEnabled(gate: FeatureGate | undefined, capabilities: WorkspaceNavigationCapabilities): boolean {
  if (!gate) return true

  switch (gate) {
    case 'payment':
      return capabilities.paymentEnabled
    case 'riskControl':
      return capabilities.riskControlEnabled
    case 'modelPlaza':
      return capabilities.modelPlazaEnabled
    case 'availableChannels':
      return capabilities.availableChannelsEnabled
    case 'channelMonitor':
      return capabilities.channelMonitorEnabled
    case 'affiliate':
      return capabilities.affiliateEnabled
  }
}

export function buildWorkspaceNavigation(
  router: Router,
  capabilities: WorkspaceNavigationCapabilities,
): WorkspaceNavigationSection[] {
  const definitions = capabilities.isAdmin ? adminDefinitions : userDefinitions

  return definitions.flatMap((section) => {
    const items = section.items.flatMap<WorkspaceNavigationItem>((item) => {
      if (item.hideInSimpleMode && capabilities.isSimpleMode) return []
      if (!gateEnabled(item.gate, capabilities)) return []
      if (!router.hasRoute(item.routeName)) return []

      const resolved = router.resolve({
        name: item.routeName,
        query: item.query,
      })

      return [{
        label: item.label,
        to: resolved.fullPath,
        icon: item.icon,
      }]
    })

    if (!items.length) return []
    return [{ label: section.label, items }]
  })
}
