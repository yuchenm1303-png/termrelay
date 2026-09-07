# Smirel Google / GitHub Login Setup

The Smirel branch already contains the runtime OAuth flow for Google and GitHub. This document is the final provider-side handoff: create the OAuth apps, enter the credentials, register the callback URLs, and enable the providers.

## Existing runtime flow

Smirel already provides:

- Google and GitHub login buttons on the auth page.
- Backend OAuth start and callback routes.
- OAuth `state` validation and HttpOnly callback cookies.
- Authorization-code exchange on the backend.
- Verified-email lookup.
- Existing-account matching / OAuth identity binding.
- New-user pending registration flow when required.
- Reuse of the existing Smirel access-token / refresh-token session system.
- Frontend callback handling at `/auth/oauth/callback`.

No provider client secret is required in frontend code.

## GitHub

Create a GitHub OAuth App and configure:

- Client ID: save to the Smirel GitHub OAuth settings.
- Client Secret: save to the Smirel GitHub OAuth settings.
- Authorization callback URL:

  `https://YOUR_DOMAIN/api/v1/auth/oauth/github/callback`

Recommended scope:

`read:user user:email`

The backend calls GitHub's email endpoint and requires a verified email address.

After the credentials and callback URL are saved, enable GitHub OAuth in system settings.

## Google

Create a Google OAuth 2.0 Web application and configure:

- Client ID: save to the Smirel Google OAuth settings.
- Client Secret: save to the Smirel Google OAuth settings.
- Authorized redirect URI:

  `https://YOUR_DOMAIN/api/v1/auth/oauth/google/callback`

Recommended scopes:

`openid email profile`

The backend requires the Google userinfo response to contain a verified email address.

After the credentials and redirect URI are saved, enable Google OAuth in system settings.

## Frontend callback

Both providers use the existing Smirel frontend callback page:

`/auth/oauth/callback`

There is no need to create another frontend callback route.

## Provider settings already exposed by the backend

GitHub:

- `github_oauth_enabled`
- `github_oauth_client_id`
- `github_oauth_client_secret`
- `github_oauth_redirect_url`
- `github_oauth_frontend_redirect_url`

Google:

- `google_oauth_enabled`
- `google_oauth_client_id`
- `google_oauth_client_secret`
- `google_oauth_redirect_url`
- `google_oauth_frontend_redirect_url`

The settings read API exposes only whether a client secret is configured; it does not return the secret itself.

## Smoke test

1. Configure Client ID and Client Secret.
2. Register the exact backend callback URL at the provider.
3. Enable the provider in Smirel settings.
4. Open the Smirel login page and click Google or GitHub.
5. Complete provider consent.
6. Confirm the browser returns through `/api/v1/auth/oauth/<provider>/callback` and then `/auth/oauth/callback`.
7. Existing users should enter the normal Smirel session. New users should follow the existing pending-registration flow when registration completion is required.

Production deployments should use HTTPS so OAuth callback cookies are handled securely.

## Reference template

See `deploy/smirel-oauth.example.yaml` for a ready-to-fill provider template.
