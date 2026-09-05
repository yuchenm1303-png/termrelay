type LocaleCode = 'en' | 'zh'

type MessageTree = Record<string, unknown>

export const smirelLocaleOverrides: Record<LocaleCode, MessageTree> = {
  zh: {
    common: {
      noGroupsAvailable: '暂无可用线路',
    },
    nav: {
      dashboard: '概览',
      apiKeys: 'API 密钥',
      usage: '用量与日志',
      availableChannels: '可用模型',
      modelPlaza: '模型与价格',
      channelStatus: '服务状态',
      mySubscriptions: '订阅与额度',
      buySubscription: '充值与套餐',
      profile: '账户设置',
      users: '用户与权限',
      groups: '路由策略',
      channels: '模型路由',
      channelManagement: '模型路由',
      channelPricing: '模型与计费',
      channelMonitor: '路由健康',
      accounts: '上游账户',
      proxies: '网络出口',
      subscriptions: '套餐管理',
      ops: '运行状态',
      settings: '平台设置',
      securityAudit: '安全与审计',
      auditLogs: '审计日志',
      announcements: '公告管理',
      orderManagement: '订单管理',
      paymentDashboard: '收入概览',
      paymentPlans: '套餐配置',
      docs: 'API 文档',
    },
    keys: {
      group: 'API 路由',
      groupLabel: 'API 路由',
      allGroups: '全部线路',
      selectGroup: '选择线路',
      searchGroup: '搜索线路',
      noGroup: '未选择线路',
      noGroupFound: '未找到匹配线路',
      clickToChangeGroup: '点击切换线路',
      currentConcurrency: '实时并发',
    },
    onboarding: {
      admin: {
        welcome: {
          title: '👋 欢迎使用 Smirel API',
          description: '<div style="line-height: 1.8;"><p style="margin-bottom: 16px;">Smirel API 是统一的 AI API 网关与运营平台。</p><p style="margin-bottom: 12px;"><b>🎯 核心能力：</b></p><ul style="margin-left: 20px; margin-bottom: 16px;"><li>🧭 <b>路由策略</b> - 组织模型线路与计费策略</li><li>🔗 <b>上游账户</b> - 管理多个 AI 服务账户</li><li>🔑 <b>API 密钥</b> - 为客户提供独立访问凭证</li><li>📊 <b>用量与计费</b> - 查看调用、额度与成本</li></ul><p style="color: #0f9f8f; font-weight: 600;">接下来完成基础配置即可开始提供 API 服务 →</p></div>',
        },
        groupManage: {
          title: '🧭 第一步：配置路由策略',
          description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;"><b>什么是路由策略？</b></p><p style="margin-bottom: 12px;">它决定 API Key 可以使用哪一类上游能力，以及对应的计费和访问规则。</p><ul style="margin-left: 20px; margin-bottom: 12px; font-size: 13px;"><li>可绑定多个上游账户实现服务调度</li><li>可设置独立的计费倍率与额度规则</li><li>可控制公开或专属访问</li></ul><p style="margin-top: 16px; color: #0f9f8f; font-weight: 600;">👉 从左侧“路由策略”开始配置</p></div>',
        },
      },
      user: {
        welcome: {
          title: '👋 欢迎使用 Smirel API',
          description: '<div style="line-height: 1.8;"><p style="margin-bottom: 16px;">欢迎来到 Smirel API 开发者平台。</p><p style="margin-bottom: 12px;"><b>快速开始：</b></p><ul style="margin-left: 20px; margin-bottom: 16px;"><li>创建自己的 API Key</li><li>复制 Smirel Base URL</li><li>使用 OpenAI-compatible SDK 发起请求</li></ul><p style="color: #0f9f8f; font-weight: 600;">只需要几步即可开始调用 →</p></div>',
        },
      },
    },
  },
  en: {
    common: {
      noGroupsAvailable: 'No routes available',
    },
    nav: {
      dashboard: 'Overview',
      apiKeys: 'API Keys',
      usage: 'Usage & Logs',
      availableChannels: 'Available Models',
      modelPlaza: 'Models & Pricing',
      channelStatus: 'Service Status',
      mySubscriptions: 'Plans & Quota',
      buySubscription: 'Billing & Plans',
      profile: 'Account Settings',
      users: 'Users & Access',
      groups: 'Routing Policies',
      channels: 'Model Routing',
      channelManagement: 'Model Routing',
      channelPricing: 'Models & Billing',
      channelMonitor: 'Route Health',
      accounts: 'Upstream Accounts',
      proxies: 'Egress Network',
      subscriptions: 'Plan Management',
      ops: 'Operations',
      settings: 'Platform Settings',
      securityAudit: 'Security & Audit',
      auditLogs: 'Audit Logs',
      announcements: 'Announcements',
      orderManagement: 'Orders',
      paymentDashboard: 'Revenue Overview',
      paymentPlans: 'Plan Configuration',
      docs: 'API Docs',
    },
    keys: {
      group: 'API Route',
      groupLabel: 'API Route',
      allGroups: 'All routes',
      selectGroup: 'Select route',
      searchGroup: 'Search routes',
      noGroup: 'No route selected',
      noGroupFound: 'No matching routes',
      clickToChangeGroup: 'Click to change route',
      currentConcurrency: 'Live Concurrency',
    },
    onboarding: {
      admin: {
        welcome: {
          title: '👋 Welcome to Smirel API',
          description: '<div style="line-height: 1.8;"><p style="margin-bottom: 16px;">Smirel API is a unified AI API gateway and operations platform.</p><p style="margin-bottom: 12px;"><b>Core capabilities:</b></p><ul style="margin-left: 20px; margin-bottom: 16px;"><li><b>Routing policies</b> - Organize model routes and billing rules</li><li><b>Upstream accounts</b> - Manage AI service accounts</li><li><b>API keys</b> - Issue independent customer credentials</li><li><b>Usage & billing</b> - Track requests, quota and cost</li></ul><p style="color: #0f9f8f; font-weight: 600;">Complete the basic configuration to start serving API traffic →</p></div>',
        },
        groupManage: {
          title: '🧭 Step 1: Configure Routing',
          description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;"><b>What is a routing policy?</b></p><p style="margin-bottom: 12px;">It defines which upstream capability an API key can use and the associated billing and access rules.</p><ul style="margin-left: 20px; margin-bottom: 12px; font-size: 13px;"><li>Bind multiple upstream accounts for service routing</li><li>Configure billing multipliers and quota rules</li><li>Control public or exclusive access</li></ul><p style="margin-top: 16px; color: #0f9f8f; font-weight: 600;">👉 Start from “Routing Policies” in the sidebar</p></div>',
        },
      },
      user: {
        welcome: {
          title: '👋 Welcome to Smirel API',
          description: '<div style="line-height: 1.8;"><p style="margin-bottom: 16px;">Welcome to the Smirel API developer platform.</p><p style="margin-bottom: 12px;"><b>Quick start:</b></p><ul style="margin-left: 20px; margin-bottom: 16px;"><li>Create your API key</li><li>Copy the Smirel Base URL</li><li>Send requests with an OpenAI-compatible SDK</li></ul><p style="color: #0f9f8f; font-weight: 600;">You can be ready in just a few steps →</p></div>',
        },
      },
    },
  },
}
