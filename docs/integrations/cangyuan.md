# 沧元算力作为 TermRelay 上游（阶段一：对话 API）

> 状态：开发分支上的接入预设，**尚未部署生产、尚未配置真实 Key、尚未验证付费调用**。

官方文档：https://ai.cangyuansuanli.cn/docs/relay

## 已实现的接入方式

在管理员后台「上游账号 → 新增账号」：

1. 选择 **OpenAI** 协议，**沧元算力 · New API** 上游供应商。
2. 基础地址默认填 `https://ai.cangyuansuanli.cn`（对应的模型列表端点是 `/v1/models`），使用平台控制台生成的 API Key；密钥**不要**提交到源码、Issue、PR 或聊天记录。
3. 点击「同步模型」，通过服务端执行带密钥的只读模型列表探测。没有真实密钥时无法确认当前账号可用模型。
4. **只手动勾选确认支持 OpenAI Chat Completions 且已核对定价的对话模型**。不同厂商的模型名不意味着上游协议也不同；此渠道始终属于 OpenAI 分组。
5. 账号初始**不自动绑定调度分组**，以免在价格/调用验证前承接生产请求。
6. 接通后在独立测试环境小额验证模型响应、流式、失败计费、上下游用量和销售价格，再将该账号绑定目标 OpenAI 分组。

上游配置保留在已有 `openai / apikey` 数据模型中，不新增协议类型：
- `credentials.base_url`: `https://ai.cangyuansuanli.cn`
- `credentials.model_mapping`: 只保存手动勾选的模型 ID → 原始 ID 的映射
- `extra.upstream_vendor`: `cangyuan`（仅作为来源标记）
- `extra.openai_passthrough`: `false`（**必须**保持关闭，透传会绕过模型白名单）
- `extra.openai_responses_mode`: `force_chat_completions`（对话入口直转 `/v1/chat/completions`）

网关已支持 `GET /v1/models`、OpenAI API Key 鉴权以及直转 Chat Completions；不需要另建一个「Anthropic」分组来承载兼容渠道里的 Claude 模型。

## 生产上线前检查

- [ ] 从真实沧元令牌执行 `GET /v1/models`，确认型号与账号权限。只读，不产生模型推理费用。
- [ ] 确认 `POST /v1/chat/completions` 的非流式和流式响应格式与计费字段。
- [ ] 确认各模型价格和计费币种，设置 TermRelay 销售价、成本口径和利润保护。
- [ ] 检查 401/403、429、5xx、超时、上游断流等失败场景，确保**没有成功交付时不重复或错误扣费**。
- [ ] 先用隔离的测试用户、测试 API Key 和测试分组走一次链路，再启用生产分组。
- [ ] 验证现有供应商、调度、计费流程无回归后再决定是否合并、部署。

## 暂不开放的模型类型

沧元平台同时提供图片、视频、音乐和异步任务等产品。这些模型可能使用 `/v1/images/*`、`/v1/videos`、`/v1/3d/generations` 或独立任务轮询，且可按张、条、秒或 Token 计费。**仅创建 OpenAI API Key 上游账号并不会让这些接口自动可用。**

阶段二需要按照 [沧元 Relay 文档](https://ai.cangyuansuanli.cn/docs/relay) 的 Skill 说明逐一适配：
- 生成请求/响应结构、资源上传、异步任务 ID 和查询（不可盲目复用同步 Chat Completions）；
- 按具体模型的费率/分辨率/时长处理价格、用量和成功状态；
- 失败、超时或任务轮询时不重复结算，并处理失败或待完成任务；
- 使用正式令牌可见的模型 ID 与官方单模型文档，不根据市场宣传推断支持范围。

接入期间**不要修改正在运行的生产容器、数据库、环境文件，也不要将 API Key 写入 Git**。

## 与官方 Relay 文档核对的多模态协议（2026-11-17）

官方参考：[沧元 Relay 指南](https://ai.cangyuansuanli.cn/docs/relay) 和公开的 `/docs-static/getting-started/clients/relay.md`。这些是**第二阶段的验收依据**，并不表示下列能力在 TermRelay 已经交付：

| 能力 | 上游提交 | 上游查询 | 成功后返回 | 计费注意 |
|---|---|---|---|---|
| 图像生成 | `POST /v1/images/generations`，`async:true` | `GET /v1/images/generations/{task_id}` | `data[0].url` | 按张、按档、按量，需匹配模型页 |
| 图像编辑 | `POST /v1/images/edits`，`async:true` | `GET /v1/images/edits/{task_id}` | `data[0].url` | 参考图和蒙版需可访问的 HTTPS |
| 视频 | `POST /v1/videos` | `GET /v1/videos/{task_id}` | `video_url` 或 `data[0].url` | 以模型定价为准，不能擅自再乘视频秒数 |
| 音乐 | `POST /v1/music` | `GET /v1/music/{task_id}` | `music_url` 数组 | `extra` 与结果数组逐项对应 |
| 三维 | `POST /v1/3d/generations` | `GET /v1/3d/generations/{task_id}` | `data[].url` | 异步按实际任务完结结算 |

异步任务通用状态为 `queued`、`in_progress`、`completed`、`failed`；建议 5–10 秒轮询，视频任务客户端总等待需支持 30 分钟以上。不得把任务创建成功当成模型生成成功，也不得重复提交同一任务或根据 `GET /content` 拼结果链接。

官方明确：只调用 `GET /v1/models` 能确认**令牌权限内**的模型，`/api/pricing` 则是价格源，二者不可混为一谈。价格可能含 `n`、`vs`、`tok`、`vref` 等分档变量；原价和本站用户实际扣费必须分开计算。未经模型逐档验证不可上架，也不可沿用 OpenAI Token 计费给按张/条模型。

**当前实现的保护**：沧元预设被限定为经过明确选择的对话模型，不接收普通 OpenAI Images 同步分流；媒体能力必须新建独立调度与费用结算流程后再开放。不得因为模型 ID 在 `/v1/models` 中出现就直接视为可售。
