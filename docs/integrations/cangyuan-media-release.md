# 沧元算力媒体上游：配置与联调

> 本文适用于 `feat/cangyuan-upstream-integration` 分支 / PR #53，基于 2026-11-18 的开发进度。代码已接入媒体提交和恢复 Worker，但完整 CI、正式部署与真实付费联调结果必须单独确认。**不要把此文当作生产验收证明。**

官方接口文档：https://ai.cangyuansuanli.cn/docs/relay

## 一、管理端准备

1. 在「上游账号」新建 **沧元算力 · New API**，使用官方域名 `https://ai.cangyuansuanli.cn` 和**新生成**的专用上游密钥。请勿把上游密钥发到聊天、Issue、PR 或源码。
2. 同步官方 `GET /v1/models` 和公开价格表，分别选择该账号实际授权的**图片**或**视频**模型。保存成媒体专用账号（`cangyuan_media_only: true`），而非普通 Chat Completions 账号。
3. 将该账号**明确绑定**到目标 OpenAI 协议分组；图片请求还需要管理员启用分组的**图片生成权限**。启用账号调度；未绑定分组、未启用调度或没有模型白名单的媒体账号不会被选择。
4. 在管理员价格 API `PUT /api/v1/admin/cangyuan-media/prices/{rule-id}` 中，为每一组明确的模型、图片张数、分辨率或视频时长等建立**真实审核过**的 USD 售价与成本、档位版本。写入需要管理员身份及 step-up 2FA；先用 `enabled:false` 检查，确认后再启用。**不要直接使用示例数字充当官方报价**。
5. 使用普通**余额计费**测试用户创建关联该分组的 API Key 并充值足够余额。现阶段不支持订阅模式媒体扣费。

## 二、客户端调用

本文示例按 JSON 请求；图片编辑/图生视频参考素材应提供公网可达 HTTPS URL。请先核对该模型的具体字段及价格档位。一次生成始终使用一个稳定且唯一的 `Idempotency-Key`，网络超时重试时**沿用原键**。

图片生成：

```bash
curl -X POST "https://YOUR_TERMRELAY_HOST/v1/images/generations" \
  -H "Authorization: Bearer YOUR_TERMRELAY_API_KEY" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: YOUR_UNIQUE_REQUEST_ID" \
  -d '{"model":"YOUR_VERIFIED_IMAGE_MODEL","prompt":"a quiet harbor at sunrise","n":1,"size":"1024x1024","async":true}'
```

图片编辑：`POST /v1/images/edits`；视频生成：`POST /v1/videos`。视频请求必须与后台价格档位（如 duration、resolution）逐项一致。客户端提交成功返回 HTTP 202，`id` / `task_id` 是本站 `cyjob_...` 任务号，**并不表示上游已经生成成功**。

结果查询（保持同一个 TermRelay 用户和 API Key）：

```bash
curl "https://YOUR_TERMRELAY_HOST/v1/images/generations/cyjob_YOUR_TASK_ID" \
  -H "Authorization: Bearer YOUR_TERMRELAY_API_KEY"
```

图片编辑：`GET /v1/images/edits/{cyjob_id}`；视频：`GET /v1/videos/{cyjob_id}`。生成完成但本站结算尚未确认时，接口不返回媒体文件 URL；成功结算后才返回 `data[].url` 或 `video_url`。

## 三、内部账单与恢复流程

- 首次请求由服务端挑选**绑定当前分组、可调度且授权该模型的媒体专用账号**；从管理员报价表取得唯一匹配价格，生成持久化任务记录并固定完整请求摘要。
- SQL 事务冻结**审核售价的全额**，成功后才抢占唯一一次上游 POST。重复 Idempotency-Key 返回原任务，不再次冻结或提交。
- 服务启动的 `CangyuanMediaRecoveryWorker` 每隔约 6 秒检查 Redis 待处理任务索引，跨实例 Redis 租约阻止重复处理。对于排队/进行中的任务只调用官方 GET；完成且产物有效时 capture 冻结金额并幂等更新额度；确定失败时释放余额。
- 明确被上游拒绝的请求（部分确定性 4xx）直接进入 failed 并释放余额；网络断开、5xx、无法确定是否成功接收的响应进入人工核对路径，**不得自动重发 POST 或误退款**。
- 冻结中的任务和重试映射不会因为原始 48 小时 TTL 自动过期；结算或退款终态确认后才恢复 48 小时保留期。

## 四、尚须完成的上线验收

1. 核对 GitHub Actions **Go 单测、集成测试、lint、安全扫描**全部通过，并在隔离环境验证数据库迁移和账户鉴权。
2. 验证真实密钥的模型列表、实际小额媒体生成、上游计费金额、失败退款、限流与异常网络恢复，尤其检查视频时长/分辨率价格档位。
3. 完善资金恢复与审计：Redis 丢库/逐出后的 SQL 独立恢复记录、按任务回溯的完整用量日志、管理员结算异常查看。
4. 检查订阅模式（目前显式不支持）、账号级并发资源控制、额度限制与标准 API 的同步/异步兼容要求。当前仅以异步任务为主。
5. 在成功测试前不要对正式用户展示“已全部接通”或“失败绝不会扣款”的承诺；无论服务器是否已有测试数据，真实上游密钥都应通过受控管理后台录入。

测试期可以将该功能单独限定在一个测试分组，其他模型仍使用原来的 OpenAI / Grok 网关，不应把所有 OpenAI 图片请求强制转发到沧元。
