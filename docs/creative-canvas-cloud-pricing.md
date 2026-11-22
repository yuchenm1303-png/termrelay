# AI 创作画布 · Preview 02

本阶段在 [Preview 01](creative-canvas-foundation.md) 之上加入云端项目保存和经过审核的媒体报价。当前为**独立开发分支 PR #58**，未合并或部署到生产。

## 云端项目持久化

- `backend/migrations/238_creative_canvas_projects.sql` 新增 `creative_canvas_projects`，主键 `(user_id, id)`，外键绑定 `users(id)`。
- 路径（全都要求现有 JWT 登录、用户面板全局限流、管理面审计中间件）：
  - `GET /api/v1/creative/projects`：仅列当前登录用户的项目元数据。
  - `GET /api/v1/creative/projects/:id`：仅返回当前用户拥有的项目。
  - `PUT /api/v1/creative/projects/:id`：提交 `{ project, expected_version }`。首次保存用 0；其后每次携带已读到的 `version`。版本改变返回 409，**不会覆盖云端较新的文档**。
  - `DELETE /api/v1/creative/projects/:id?version=N`：删除时同样必须匹配已知版本。
- 限制：每用户最多 20 个项目、每文档 500KB、每项目最多 120 个节点和 250 条连线。服务端严格验证字段、图合法性和 HTTPS 媒体链接；禁止凭证字段混入持久化。
- 前端仍保留本地草稿。云端断网时本地继续编辑；有版本冲突时停止自动同步，并保留本机内容。重新连接后发现同 ID 的云端和本机存在差异，会**另存本机恢复副本**，保留本地已付费任务编号和本机素材引用。
- 上传的二进制图片/视频目前仍在浏览器 IndexedDB；**云端保存的是节点结构、提示词与任务编号，不代表媒体二进制已同步**。清理浏览器缓存可能丢失仅本机上传的素材。

## 生成前精确价格与不可变更承诺

- `POST /api/v1/creative/media/quote` 以当前 JWT 用户拥有的 API Key ID、实际模型、生成种类和完整请求参数为输入。
- 只有合法、有效、未过期、分组已授权、存在可调度账号、并与现有管理员审核价规则 **唯一精确匹配**的请求才返回 `approved: true` 和 USD 单价/合计、计价版本、档位。**不会展示上游成本或 API Key**。
- UI 在用户确认付费前重新实时核价，不使用预存旧报价。服务端收到 `X-Canvas-Quoted-Sale-USD` 与 `X-Canvas-Quoted-Pricing-Version` 后，**对同一请求在冻结余额之前重新计算**。若价格、档位、版本或可调度媒体账号变动则返回拒绝，不落入其他上游的计费通道。
- 不确认、无可用已审核价格、无有效 Key、无已授权媒体账号，则客户端 **不发送付费生成 POST**。
- 仍由现有 TermRelay 媒体托管系统负责冻结、结算与失败退款；本阶段不修改任何正式生产价格规则或调度开关。

## 图片 → 视频的显式首帧映射

视频节点连接到已完成且有真实任务编号、公共 HTTPS 结果的图片节点后，使用 `options` 内的 `"canvas_use_input_image": true` 显式要求用该图片作首帧。客户端移除这一画布专用标志，向既有沧元视频接口传递 `first_image_url`。服务器验证引用 URL 为公开 HTTPS，并且仅当**相应首帧存在性和其他档位参数匹配的审核价格**已启用时允许提交。未审核模型或没有上游参数支持时，不会尝试强行生成。

## 尚未完成

- 画布图片/视频二进制的云端私有素材库、跨设备同步、素材配额与生命周期清理。
- 视频模型能力目录的可视化选择（目前用精确参数 JSON 和后台已审核规则约束）。
- 批量自动运行节点、完整 DAG 队列执行与多人协作。
- iPad Safari、桌面 Chrome 的真实交互截图与生产小额计费验收。
- 管理员人工审核开启可收费模型之前，普通用户的付费按钮继续关闭。

## 验证

- 云端、画布、报价客户端 Vitest 用例，后端强校验及 SQL CAS 单测。
- GitHub CI、Security Scan、Production Control。
- 本地 Docker 前端 Vite + Vue-TSC 全量构建。
- **不要因为 PR 测试全绿而自动启用既有沧元测试账号或收费规则。**
