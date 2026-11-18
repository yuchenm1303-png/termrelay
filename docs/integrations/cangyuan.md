# 沧元算力作为 TermRelay 上游（阶段一：对话 API）

> 状态：开发分支上的接入预设，**尚未部署生产、尚未配置真实 Key、尚未验证付费调用**。

官方文档：https://ai.cangyuansuanli.cn/docs/relay

## 已实现的接入方式

在管理员后台「上游账号 → 新增账号」：

1. 选择 **OpenAI** 协议，**沧元算力 · New API** 上游供应商。
2. 基础地址默认填 `https://ai.cangyuansuanli.cn`（对应的模型列表端点是 `/v1/models`），使用平台控制台生成的 API Key；密钥**不要**提交到源码、Issue、PR 或聊天记录。
3. 点击「同步模型」，通过服务端执行带密钥的只读模型列表探测。没有真实密钥时无法确认当前账号可用模型。
4. **只手动勾选确认支持 OpenAI Chat Completions 且已核对定价的对话模型**。不同厂商的模型名不意味着上游协议也不同；此渠道始终属于 OpenAI 分组。
5. 账号初始**不自动绑定调度分组且禁用调度**，以免被无分组兜底调度选中。完成价格、失败扣费和真实调用验证后，在管理后台手动绑定测试/生产分组，并明确点击「启用调度」。
6. 接通后在独立测试环境小额验证模型响应、流式、失败计费、上下游用量和销售价格，再将该账号绑定目标 OpenAI 分组。

上游配置保留在已有 `openai / apikey` 数据模型中，不新增协议类型：
- `credentials.base_url`: `https://ai.cangyuansuanli.cn`
- `credentials.model_mapping`: 只保存手动勾选的模型 ID → 原始 ID 的映射
- `extra.upstream_vendor`: `cangyuan`（仅作为来源标记）
- `extra.openai_passthrough`: `false`（**必须**保持关闭，透传会绕过模型白名单）
- `extra.openai_responses_mode`: `force_chat_completions`（对话入口直转 `/v1/chat/completions`）

网关已支持 `GET /v1/models`、OpenAI API Key 鉴权以及直转 Chat Completions；不需要另建一个「Anthropic」分组来承载兼容渠道里的 Claude 模型。

## 模型与价格双重筛选（阶段一新增）

新增账号选择「沧元算力 · New API」并点击「同步模型」时，后台现在按下列顺序执行：

1. 通过管理员输入的密钥只读调用沧元 `GET /v1/models`，获取**该密钥实际有权限**的模型。
2. 从固定公开地址 `https://ai.cangyuansuanli.cn/api/pricing` 读取最新版价格元数据，**不向价格接口发送 API Key**。
3. 只返回同时存在于上述两处、`supported_endpoint_types` 含 `openai` 且 `quota_type=0`（Token 用量）的**候选对话模型**。图片、视频、音乐、3D、按请求计费和缺乏价格分类的模型会被过滤。
4. 价格表读取失败或格式异常时**拒绝显示未核价候选模型**，不能默默回退为自动开放全模型目录。
5. 后台显示被过滤的模型数与每个候选模型是否涉及分档表达式；**不执行上游分档表达式，不导入上游分组倍率，也不自动设置售价**。

**重要边界**：`supported_endpoint_types=openai` 仅代表官方目录声明兼容 OpenAI 协议，不足以证明模型一定支持 Chat Completions 的特定格式（例如某些语音类型号可能复用 OpenAI 协议）。管理员仍需按模型官方文档核对路由、价格并小额实际测试。分档模型不能直接使用单价冒充全部档位。在 API Key 和用量测试完成之前，账户会保持暂停调度。

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

## 图片与视频专用适配（开发中，不可直接销售）

本分支现包含 `internal/service/cangyuan_media_client.go`，实现了一个**独立、尚未绑定公开路由及收费的**异步客户端：

- 文生图：`POST /v1/images/generations`，强制 `async:true`，同路径 `GET /v1/images/generations/{task_id}`；
- 图片编辑：`POST /v1/images/edits`，同路径 `GET /v1/images/edits/{task_id}`；
- 视频：`POST /v1/videos`，同路径 `GET /v1/videos/{task_id}`；
- 状态支持 `queued` / `in_progress` / `completed` / `failed`；完成时必须有图像或视频链接，否则按异常处理；
- 所有 POST 仅提交一次，即使连接异常也不自动重新生成；状态轮询只调用 GET；轮询必须由调用者设置有效截止时间；
- 上游请求只使用预设的官方 HTTPS 域名，禁止自动跟随重定向携带密钥，任务 ID 严格校验；
- 图片编辑、图生视频参考素材要求公网 HTTPS URL；本地路径、base64、私网/本机地址被提前拒绝。上传素材到公网对象存储的步骤尚未实现；
- `GET /v1/models` 与公开价格表的交集现在分别展示**对话、图片和视频候选模型**，媒体仅做只读预览，不会自动创建渠道或开启调度。

**重要**：独立异步客户端的测试使用模拟 HTTP 响应，尚未使用真实 API Key 发起收费生成。它没有接入 TermRelay 公共图片/视频路由，不能作为可用的最终上线能力。

后续可发布媒体渠道前必须补齐：

1. Redis/数据库持久化任务 ID 与用户、API Key、账户、分组、具体模型绑定。查询任务时不能泄漏其他用户任务信息；
2. 真正实现 `POST /v1/images/generations`、`/edits` 的同步与异步双模式，以及 `GET` 原路径查询；视频使用 `POST /v1/videos` 和 `GET /v1/videos/{id}`，不能借用不兼容的 Grok/Sora 路由；
3. 按供应商成本 `per_request`、`tiered_expr`、`n`、时长、清晰度等建立**明确的计费规则**，本站销售倍率单独计算。未实现表达式解析的价格不得上架；
4. 使用幂等结算记录，**只在确认完成并有产物后扣费一次**。失败任务不扣费；对于超时/断网但上游仍在生成的任务，不得重新提交或在未知状态下多次扣费；
5. 图像参考图 HTTPS 上传、视频素材权限、超时重试、30 分钟以上轮询预算与跨实例恢复；在隔离测试组用真实密钥验证扣费、失败、任务重试等。


## 任务持久化与单次提交保障（开发中，不是可收费上线）

新增 `internal/service/cangyuan_media_job.go` 与 `internal/repository/cangyuan_media_job_store.go`。这两层目前供服务端独立编排与模拟测试使用，尚**未**接入公开路由：

1. 提交前通过 Redis `SETNX` 预留本站 `cyjob_...` 任务，保存用户 ID、调用 API Key ID、上游账号 ID、模型 ID、图片/视频类型、创建时间、有效期；**不保存上游密钥或 prompt**。
2. 必须先通过 Redis WATCH/CAS 将任务从 `reserved` 原子变为 `submitting`，仅抢占成功的进程可以执行一次上游 POST。并行调用不会造成双重生成。
3. 上游返回任务 ID 后，CAS 持久化 `queued` / `in_progress` 等状态和上游任务号；随后只进行 `GET` 查询，不重复 POST。
4. 网络断开、上游是否接收 POST 不明时进入 `submit_unknown`；客户端不自动重试。若进程在上游响应后、写回 Redis 前崩溃，可能需要人工与上游账单核对，不能自动再次发起生成。
5. 只有任务拥有者（同一用户、同一 API Key）可以读取或推进自己的任务；终态 `completed`、`failed` 不允许被后到的 `in_progress` 回退。
6. Redis 使用版本比较和保留原 TTL 的更新。默认保留 48 小时，轮询不会无休止续期。
7. `billing_state=disabled` 被写入记录，并在 CAS 时强制不可更改。**现阶段不向用户扣费**，也不声称已经实现上线所需的“成功后只结算一次”。

需继续实现的功能是将该任务注册表接入经过模型权限检查的实际网关路由，补齐任务结果归属、售卖档位价格和数据库唯一键账单事务，使扣费与账单记录原子完成。测试已涵盖身份隔离、并发版本冲突、禁止重复 POST、网络异常状态、过期与正常完成路径。


## 本地分档售价、数据库幂等结算与客户端重试键（开发中）

当前实现的内部接口已经可用来编排和测试媒体任务，但**还没有注册对外收费 API**，因为生产前还必须接好用户分组权限、余额预留、完整用量日志和定价配置页面。

- **冻结价格**：任务创建时可选保存 `CangyuanMediaPriceSnapshot`，包含运营后台审核的本地 USD 售价、上游成本、数量、档位和价格版本。只有 `verified_by_admin=true` 且报价合法，任务才会标记为待结算；无报价的旧测试任务保持 `billing_state=disabled`。
- **价格不可回改**：Redis CAS 校验冻结报价的每一个字段，不允许生成后降低、提高售价，也不能从没有报价的任务凭空启用计费。本站售价独立于沧元价格表，不复制对方的分组倍率。
- **成功才准扣费**：`SettleCompleted` 只处理 `completed` 且实际结果数量与价格快照一致、所有结果链接均为公网 HTTPS 的任务；对 `queued`、`failed`、`submit_unknown`、无报价任务拒绝结算。
- **SQL 而非 Redis 防重复扣费**：使用既有 `UsageBillingRepository.Apply` 的事务与 `usage_billing_dedup` 唯一键，同一本站任务固定使用 `request_id=cangyuan-media:cyjob_...`；SQL 扣费已提交但 Redis `settled` 标记写回失败时，重试相同账单不会重复扣费。
- **按客户端请求去重**：新增 `ReserveIdempotent` 与 Redis `CreateIdempotent`，要求付费媒体 POST 的调用方提交 `Idempotency-Key`。基于用户、API Key 和请求键进行 Redis 原子任务预留；相同请求重试只返回原任务；同一键换模型、参数或价格则拒绝。仅保存 SHA-256 指纹和任务号，不存原始 prompt。去重映射与任务记录同生命周期（默认 48 小时）。**这仍需公开网关接入才会生效**。
- **状态不回退**：`queued → in_progress → completed/failed`，迟到的 `queued` 不会覆盖 `in_progress`，重复状态不重复写 Redis。上游结果网址必须是公网 HTTPS，避免恶意或无效产物触发结算。

**上线前尚缺**：站内可信定价管理界面、请求参数与档位匹配、任务与账号路由、API Key 分组权限、生成前余额/额度预留、完整用量/退款日志、取消/超时任务策略，以及真实 API Key 的隔离小额付费测试。当前结算接口没有被任何外部路由调用，因此不会对用户产生扣费。
