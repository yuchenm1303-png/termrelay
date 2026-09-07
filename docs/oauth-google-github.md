# Smirel Google / GitHub OAuth 接入清单

Google 和 GitHub OAuth 的代码链路已经预埋完成。接入时只需要创建官方 OAuth 应用、填写凭据并启用 Provider。

## 已有代码链路

- 登录/注册页会根据公开设置显示 Google / GitHub 按钮。
- 后端已有 OAuth start、callback 和新用户 complete-registration 路由。
- OAuth `state` 使用 HttpOnly Cookie 校验，Client Secret 只在后端使用。
- Google 只接受 `email_verified=true` 的邮箱。
- GitHub 会读取 `/user/emails`，只接受 verified email。
- 已存在的 OAuth 身份可直接登录；新身份会进入 Smirel 的待注册流程并设置本地密码。
- 登录成功后复用 Smirel 现有 access token / refresh token 会话体系。

## Google

在 Google Cloud 创建 OAuth 2.0 Web application。

需要填写：

- Enabled: `true`
- Client ID
- Client Secret
- Redirect URL: `https://YOUR_DOMAIN/api/v1/auth/oauth/google/callback`
- Frontend Redirect URL: `/auth/oauth/callback`

Google 后台的 Authorized redirect URI 必须与上面的 Redirect URL 完全一致。

代码已经提供默认 Google authorize/token/userinfo 地址和默认 scope，因此正常接入不需要手工修改这些端点。

## GitHub

在 GitHub Developer Settings 创建 OAuth App。

需要填写：

- Enabled: `true`
- Client ID
- Client Secret
- Redirect URL: `https://YOUR_DOMAIN/api/v1/auth/oauth/github/callback`
- Frontend Redirect URL: `/auth/oauth/callback`

GitHub OAuth App 的 Authorization callback URL 必须与上面的 Redirect URL 完全一致。

代码已经提供默认 GitHub authorize/token/userinfo/emails 地址和所需 scope，因此正常接入不需要手工修改这些端点。

## 配置位置

后端已经支持两组 Provider 配置：

- `github_oauth`
- `google_oauth`

也支持系统 Setting 覆盖配置文件中的值。Provider 只有在 `enabled=true` 且 Client ID、Client Secret 均存在时才会向前端公开为可用登录方式。

Client Secret 不要写入前端代码，也不要提交真实 Secret 到 GitHub。

## 接入后检查

1. 打开登录页，确认对应 Google / GitHub 按钮出现。
2. 点击按钮，确认跳到官方授权页。
3. 授权后确认回到 `/api/v1/auth/oauth/<provider>/callback`。
4. 已有账号确认能直接进入工作区。
5. 新账号确认能进入 `/auth/oauth/callback` 完成 Smirel 密码设置并登录。
6. 最后检查 access token / refresh token 与普通邮箱登录一致可用。
