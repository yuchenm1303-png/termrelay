# Smirel frontend hot-swap

This deployment path separates routine Smirel UI releases from the TermRelay backend image.

## Why this exists

The normal TermRelay release image builds the Vue frontend and embeds `frontend/dist` into the Go binary with `//go:embed all:dist`. That is a good immutable fallback, but it means a CSS/Vue-only change would normally require rebuilding and recreating the application container.

The hot-swap server keeps that embedded bundle as the fallback and adds one runtime source:

```text
/app/data/frontend -> /app/data/frontend-releases/<ui-commit>
```

The production compose already mounts the host `./data/app` directory at `/app/data`, so no Caddy, PostgreSQL or Redis change is required.

`data/public` is not used for full frontend releases. It remains the legacy per-file override directory.

## One-time backend rollout

The currently running backend must include `feature/frontend-hot-swap` (or the equivalent merged commits) once before `update-frontend.sh` can activate external bundles.

For the 2026-09-07 production state, the feature branch was created directly from backend commit:

```text
0231ad7ae31362e544db29ba08e1b70c42628259
```

That matches the `termrelay:0231ad7a` backend baseline. Deploying the hot-swap backend is a one-time backend change. Keep the existing backend rollback image and rollback record until this capability has been verified in production.

Do not modify Caddy for this feature. In particular, no single-file Caddy bind mount replacement or reload is necessary.

## Routine frontend update

From `deploy/termrelay`:

```bash
./update-frontend.sh --dry-run
./update-frontend.sh
```

The default UI source is:

```text
release/smirel-commercial-dark
```

A different UI branch can be selected explicitly:

```bash
./update-frontend.sh --branch release/smirel-ui-next
```

The updater:

1. fetches only the requested UI branch;
2. resolves its immutable Git commit;
3. exports the source without checking out or changing the server worktree;
4. builds only `frontend/` in an isolated `node:24-alpine` container with `VITE_STANDALONE=true`;
5. rejects preview builds containing the `__preview_api__` marker;
6. stores the bundle under `data/app/frontend-releases/<ui-commit>`;
7. atomically switches `data/app/frontend` to the new release;
8. verifies the running `termrelay-app` health endpoint and confirms the root HTML references an asset from the new release;
9. automatically restores the previous frontend state if verification fails.

It does not rebuild or recreate `termrelay-app`, PostgreSQL, Redis or Caddy.

## Frontend rollback

```bash
./update-frontend.sh --rollback --dry-run
./update-frontend.sh --rollback
```

The previous state may be another external frontend release or the original embedded frontend. Rollback switches the state atomically and verifies it before recording success.

Because Vite assets are content-hashed, old release directories are intentionally retained. This also keeps assets available to browser tabs that were opened immediately before a release switch.

## Production paths

With the standard TermRelay production package under `/opt/termrelay/deploy/termrelay`:

```text
/opt/termrelay/deploy/termrelay/data/app/frontend
/opt/termrelay/deploy/termrelay/data/app/frontend.previous
/opt/termrelay/deploy/termrelay/data/app/frontend-releases/<ui-commit>
```

Inside the application container the active frontend is visible at:

```text
/app/data/frontend
```

## Safety model

- Embedded frontend remains the final fallback.
- API paths always bypass the external frontend middleware.
- `index.html` is served with existing public-settings injection and CSP nonce handling.
- Fingerprinted Vite assets keep immutable cache headers.
- No database migration is involved.
- No Caddy reload is involved.
- No Docker Compose `down`, volume deletion or global prune is used.
