# TermRelay production deployment

This directory is the production deployment package for the customized `custom-ui` branch. It runs four isolated services:

- `termrelay`: the customized Go backend with the Vue frontend embedded
- `postgres`: persistent application data
- `redis`: sessions, queues and distributed gateway state
- `caddy`: public HTTP/HTTPS ingress with streaming and WebSocket proxying

The application container is not exposed directly to the host. All public traffic enters through Caddy.

## 1. Server requirements

Recommended first-stage VPS:

- Ubuntu 24.04 LTS or another current Linux distribution
- 2 vCPU minimum; 4 vCPU recommended
- 4 GB RAM minimum; 8 GB recommended for sustained traffic
- 30 GB SSD minimum
- Docker Engine 24+
- Docker Compose v2+
- A domain pointed to the VPS for automatic HTTPS

Open inbound TCP ports `22`, `80`, and `443`, plus UDP `443` for HTTP/3. Do not expose PostgreSQL or Redis.

## 2. Build the customized image

The workflow `.github/workflows/build-termrelay-image.yml` publishes this branch to:

```text
ghcr.io/yuchenm1303-png/termrelay:custom-ui
```

After its first successful run, make the package public in GitHub Packages. When the package remains private, log in on the server before pulling:

```bash
echo "$GITHUB_TOKEN" | docker login ghcr.io -u yuchenm1303-png --password-stdin
```

Use a token with `read:packages`; do not store it in this repository.

## 3. Prepare the server

Clone the repository and enter this directory:

```bash
git clone --branch custom-ui https://github.com/yuchenm1303-png/termrelay.git
cd termrelay/deploy/termrelay
bash prepare.sh
```

The preparation script creates `.env`, generates independent PostgreSQL, Redis, admin, JWT and TOTP secrets, applies restrictive file permissions, and creates persistent data directories.

Edit the generated environment file:

```bash
nano .env
```

At minimum, set:

```dotenv
TERMRELAY_DOMAIN=relay.example.com
ADMIN_EMAIL=your-admin@example.com
```

Point the domain's `A` record to the VPS before starting. For temporary IP-only testing, set `TERMRELAY_DOMAIN=http://SERVER_IP`; OAuth and passkeys should be tested again after moving to the final HTTPS domain.

## 4. Start the stack

```bash
docker compose pull
docker compose up -d
docker compose ps
```

Follow startup logs:

```bash
docker compose logs -f termrelay
```

Caddy obtains and renews the TLS certificate automatically when DNS and firewall rules are correct.

## 5. Verify the deployment

Check the public health endpoint:

```bash
curl -fsS https://relay.example.com/health
```

Then open the domain in a browser and sign in with `ADMIN_EMAIL` and the generated administrator password printed by `prepare.sh`.

Useful checks:

```bash
# Service state
docker compose ps

# Application logs
docker compose logs --tail=200 termrelay

# Reverse-proxy and certificate logs
docker compose logs --tail=200 caddy

# Database health
docker compose exec postgres pg_isready -U termrelay -d termrelay

# Redis health
docker compose exec redis redis-cli ping
```

## 6. Connect the first OpenAI OAuth account

After administrator login:

1. Open **Upstream Accounts**.
2. Choose **Connect upstream account**.
3. Select **OpenAI** and **OAuth**.
4. Complete authorization using the final HTTPS domain.
5. Run the built-in connectivity test.
6. Keep the account schedulable only after the test succeeds.

The current first-stage deployment uses `RUN_MODE=simple`, which is intended for private operation and skips SaaS billing flows. Change it to `standard` only when public users, balances, subscriptions or payment features are deliberately enabled.

## 7. Create and test a downstream key

Create a key in **API Key Control**, assign it to the OpenAI group, then test the unified endpoint:

```bash
curl https://relay.example.com/v1/responses \
  -H "Authorization: Bearer YOUR_TERMRELAY_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5",
    "input": "Return exactly: TermRelay online"
  }'
```

Use a model exposed by the connected upstream account. Do not paste the upstream OAuth token into clients; clients should receive only TermRelay-issued keys.

## 8. Update and rollback

Update to the newest `custom-ui` image:

```bash
git pull origin custom-ui
docker compose pull termrelay
docker compose up -d
```

Every image build also receives an immutable SHA tag such as `custom-ui-abcdef0`. For rollback, set `TERMRELAY_IMAGE` in `.env` to a known SHA tag and recreate the application:

```bash
docker compose pull termrelay
docker compose up -d termrelay
```

## 9. Backup

Create a consistent backup window:

```bash
docker compose stop termrelay caddy
tar czf "termrelay-backup-$(date +%F-%H%M).tar.gz" .env Caddyfile data/
docker compose start termrelay caddy
```

Keep backups encrypted and outside the VPS. The `.env` file and `data/app` directory contain sensitive authentication material.

## 10. Security notes

- Never publish `.env`, database files, Redis data or application data.
- Keep the application service internal; expose only Caddy.
- Caddy overwrites forwarded-client headers before proxying to TermRelay.
- HTTPS-only upstream URLs and private-host blocking are enabled in the production compose file.
- Enable TOTP for the administrator immediately after the first login.
- Use separate downstream API keys for each device or client and revoke them individually.
- Review upstream provider terms before using OAuth-backed relay functionality.
