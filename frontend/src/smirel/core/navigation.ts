export interface NavItem {
  path: string
  name: string
  label: string
  feature: string
  short: string
}

export const userNavigation: NavItem[] = [
  { path: '/dashboard', name: 'Dashboard', label: 'Overview', feature: 'dashboard', short: 'OV' },
  { path: '/keys', name: 'Keys', label: 'API Keys', feature: 'keys', short: 'AK' },
  { path: '/usage', name: 'Usage', label: '用量与日志', feature: 'usage', short: 'US' },
  { path: '/subscriptions', name: 'Subscriptions', label: '订阅', feature: 'subscriptions', short: 'SB' },
  { path: '/purchase', name: 'Purchase', label: '购买服务', feature: 'purchase', short: 'BY' },
  { path: '/orders', name: 'Orders', label: '订单', feature: 'orders', short: 'OR' },
  { path: '/redeem', name: 'Redeem', label: '兑换码', feature: 'redeem', short: 'RD' },
  { path: '/affiliate', name: 'Affiliate', label: '邀请返利', feature: 'affiliate', short: 'AF' },
  { path: '/available-channels', name: 'AvailableChannels', label: '可用渠道', feature: 'channels', short: 'CH' },
  { path: '/monitor', name: 'Monitor', label: '服务状态', feature: 'monitor', short: 'ST' },
  { path: '/batch-image', name: 'BatchImage', label: '批量图片', feature: 'batch-image', short: 'BI' },
  { path: '/profile', name: 'Profile', label: '账户设置', feature: 'profile', short: 'AC' },
]

export const adminNavigation: NavItem[] = [
  { path: '/admin/dashboard', name: 'AdminDashboard', label: 'Overview', feature: 'admin-dashboard', short: 'OV' },
  { path: '/admin/users', name: 'AdminUsers', label: '用户', feature: 'admin-users', short: 'UR' },
  { path: '/admin/accounts', name: 'AdminAccounts', label: '上游账户', feature: 'admin-accounts', short: 'UP' },
  { path: '/admin/groups', name: 'AdminGroups', label: '分组与模型', feature: 'admin-groups', short: 'GP' },
  { path: '/admin/channels/pricing', name: 'AdminChannels', label: '渠道与价格', feature: 'admin-channels', short: 'PR' },
  { path: '/admin/channels/monitor', name: 'AdminChannelMonitor', label: '渠道监控', feature: 'admin-monitor', short: 'CM' },
  { path: '/admin/usage', name: 'AdminUsage', label: '用量记录', feature: 'admin-usage', short: 'US' },
  { path: '/admin/ops', name: 'AdminOps', label: '运行监控', feature: 'admin-ops', short: 'OP' },
  { path: '/admin/audit-logs', name: 'AdminAudit', label: '审计日志', feature: 'admin-audit', short: 'AL' },
  { path: '/admin/prompt-audit', name: 'AdminPromptAudit', label: 'Prompt 审计', feature: 'admin-prompt-audit', short: 'PA' },
  { path: '/admin/subscriptions', name: 'AdminSubscriptions', label: '订阅管理', feature: 'admin-subscriptions', short: 'SB' },
  { path: '/admin/orders/dashboard', name: 'AdminPaymentDashboard', label: '支付概览', feature: 'admin-payment-dashboard', short: 'PD' },
  { path: '/admin/orders', name: 'AdminOrders', label: '订单管理', feature: 'admin-orders', short: 'OR' },
  { path: '/admin/orders/plans', name: 'AdminPaymentPlans', label: '订阅计划', feature: 'admin-plans', short: 'PL' },
  { path: '/admin/redeem', name: 'AdminRedeem', label: '兑换码', feature: 'admin-redeem', short: 'RD' },
  { path: '/admin/promo-codes', name: 'AdminPromo', label: '优惠码', feature: 'admin-promo', short: 'PC' },
  { path: '/admin/affiliates/invites', name: 'AdminAffiliateInvites', label: '邀请记录', feature: 'admin-affiliate-invites', short: 'AI' },
  { path: '/admin/affiliates/rebates', name: 'AdminAffiliateRebates', label: '返利记录', feature: 'admin-affiliate-rebates', short: 'AR' },
  { path: '/admin/affiliates/transfers', name: 'AdminAffiliateTransfers', label: '转账记录', feature: 'admin-affiliate-transfers', short: 'AT' },
  { path: '/admin/proxies', name: 'AdminProxies', label: '代理', feature: 'admin-proxies', short: 'PX' },
  { path: '/admin/announcements', name: 'AdminAnnouncements', label: '公告', feature: 'admin-announcements', short: 'AN' },
  { path: '/admin/risk-control', name: 'AdminRiskControl', label: '风控', feature: 'admin-risk', short: 'RC' },
  { path: '/admin/backup', name: 'AdminBackup', label: '备份', feature: 'admin-backup', short: 'BK' },
  { path: '/admin/settings', name: 'AdminSettings', label: '平台设置', feature: 'admin-settings', short: 'ST' },
]
