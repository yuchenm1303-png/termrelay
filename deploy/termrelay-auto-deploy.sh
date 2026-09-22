#!/usr/bin/env bash
set -Eeuo pipefail
umask 027

ROOT="${TERMRELAY_ROOT:-/srv/termrelay}"
REPO="${TERMRELAY_REPO:-$ROOT/app}"
STATE_DIR="${TERMRELAY_DEPLOY_STATE_DIR:-$ROOT/deploy-state}"
WORKTREE_ROOT="${TERMRELAY_WORKTREE_ROOT:-$ROOT/autodeploy/worktrees}"
BACKUP_ROOT="${TERMRELAY_DEPLOY_BACKUP_ROOT:-$ROOT/backups/auto-deploy}"
COMPOSE_FILE="${TERMRELAY_COMPOSE_FILE:-$ROOT/docker-compose.yml}"
COMPOSE_OVERLAY="${TERMRELAY_COMPOSE_OVERLAY:-$ROOT/docker-compose.termrelay.yml}"
ENV_FILE="${TERMRELAY_ENV_FILE:-$ROOT/.env}"
HEALTH_URL="${TERMRELAY_HEALTH_URL:-http://127.0.0.1:8080/health}"
PUBLIC_HEALTH_URL="${TERMRELAY_PUBLIC_HEALTH_URL:-https://muxway.dev/health}"
GITHUB_REPO="${TERMRELAY_GITHUB_REPO:-yuchenm1303-png/termrelay}"
REQUIRED_WORKFLOWS="${TERMRELAY_REQUIRED_WORKFLOWS:-CI,Security Scan,Production Control}"
MIN_FREE_KB="${TERMRELAY_MIN_FREE_KB:-3145728}"

force=false
check_only=false
target_override=""
case "${1:-}" in
  --force) force=true ;;
  --check-only) check_only=true ;;
  --target)
    target_override="${2:-}"
    [[ -n "$target_override" ]] || { echo "missing SHA after --target" >&2; exit 2; }
    force=true
    ;;
  "")
    ;;
  *)
    echo "usage: $0 [--check-only|--force|--target <sha>]" >&2
    exit 2
    ;;
esac

mkdir -p "$STATE_DIR" "$WORKTREE_ROOT" "$BACKUP_ROOT"
exec 9>"$STATE_DIR/deploy.lock"
if ! flock -n 9; then
  echo "[auto-deploy] another deployment is already running"
  exit 0
fi

log() {
  printf '[%s] [auto-deploy] %s\n' "$(date -u +'%Y-%m-%dT%H:%M:%SZ')" "$*"
}

fail() {
  log "ERROR: $*"
  exit 1
}

for cmd in git docker curl python3 flock; do
  command -v "$cmd" >/dev/null 2>&1 || fail "required command not found: $cmd"
done
[[ -d "$REPO/.git" || -f "$REPO/.git" ]] || fail "repository missing: $REPO"
[[ -f "$COMPOSE_FILE" ]] || fail "compose file missing: $COMPOSE_FILE"
[[ -f "$COMPOSE_OVERLAY" ]] || fail "compose overlay missing: $COMPOSE_OVERLAY"
[[ -f "$ENV_FILE" ]] || fail "environment file missing: $ENV_FILE"

if [[ -n "$(git -C "$REPO" status --porcelain)" ]]; then
  fail "production source tree is dirty; refusing automated reset/deploy"
fi

log "fetching origin/main"
git -C "$REPO" fetch --quiet --prune origin main

if [[ -n "$target_override" ]]; then
  target="$(git -C "$REPO" rev-parse "$target_override^{commit}")"
else
  target="$(git -C "$REPO" rev-parse origin/main)"
fi
short="${target:0:12}"
last_success=""
last_failed=""
[[ -f "$STATE_DIR/last-successful-sha" ]] && last_success="$(tr -d '[:space:]' < "$STATE_DIR/last-successful-sha")"
[[ -f "$STATE_DIR/last-failed-sha" ]] && last_failed="$(tr -d '[:space:]' < "$STATE_DIR/last-failed-sha")"

log "target=$target last_success=${last_success:-none} last_failed=${last_failed:-none}"

if [[ "$force" != true && "$target" == "$last_success" ]]; then
  log "already deployed"
  exit 0
fi
if [[ "$force" != true && "$target" == "$last_failed" ]]; then
  log "target previously failed; waiting for a new main commit or manual --force"
  exit 0
fi

if [[ "$check_only" == true ]]; then
  exit 0
fi

# Do not deploy a main commit until the repository's required workflows have
# completed successfully. This keeps the server from racing ahead of GitHub CI.
if [[ -z "$target_override" ]]; then
  actions_json="$(mktemp)"
  trap 'rm -f "$actions_json"' EXIT
  api="https://api.github.com/repos/$GITHUB_REPO/actions/runs?branch=main&head_sha=$target&per_page=100"
  if ! curl -fsSL --max-time 20       -H 'Accept: application/vnd.github+json'       -H 'X-GitHub-Api-Version: 2022-11-28'       "$api" -o "$actions_json"; then
    log "GitHub Actions status unavailable; deferring deployment"
    exit 0
  fi
  if ! python3 - "$actions_json" "$REQUIRED_WORKFLOWS" <<'PY'
import json, sys
data=json.load(open(sys.argv[1], encoding='utf-8'))
required=[x.strip() for x in sys.argv[2].split(',') if x.strip()]
runs=[
    run for run in (data.get('workflow_runs') or [])
    if str(run.get('event') or '') == 'push' and str(run.get('head_branch') or '') == 'main'
]
by_name={}
for run in runs:
    name=str(run.get('name') or '')
    old=by_name.get(name)
    # Keep the newest push run for each workflow. Never let an older completed
    # run hide a newer in-progress or failed run for the same commit.
    if old is None or int(run.get('run_number') or 0) > int(old.get('run_number') or 0):
        by_name[name]=run
missing=[]
bad=[]
waiting=[]
for name in required:
    run=by_name.get(name)
    if not run:
        missing.append(name)
    elif run.get('status') != 'completed':
        waiting.append(name)
    elif run.get('conclusion') != 'success':
        bad.append((name, run.get('conclusion')))
if missing or waiting or bad:
    if missing: print('missing workflows: ' + ', '.join(missing), file=sys.stderr)
    if waiting: print('workflows still running: ' + ', '.join(waiting), file=sys.stderr)
    if bad: print('failed workflows: ' + ', '.join(f'{n}={c}' for n,c in bad), file=sys.stderr)
    raise SystemExit(1)
PY
  then
    log "required GitHub checks are not green yet; deferring deployment"
    exit 0
  fi
fi

# Compose files are operational configuration, not application artifacts.
# Refuse automatic deployment if main changes them without a manual production
# review, instead of silently changing databases, ports, volumes or services.
if ! git -C "$REPO" show "$target:deploy/docker-compose.local.yml" | cmp -s - "$COMPOSE_FILE"; then
  fail "production docker-compose.yml differs from target main; manual compose review required"
fi
if ! git -C "$REPO" show "$target:deploy/docker-compose.termrelay.yml" | cmp -s - "$COMPOSE_OVERLAY"; then
  fail "production compose overlay differs from target main; manual compose review required"
fi

free_kb="$(df -Pk "$ROOT" | awk 'NR==2 {print $4}')"
if [[ "$free_kb" =~ ^[0-9]+$ ]] && (( free_kb < MIN_FREE_KB )); then
  log "free disk below threshold (${free_kb} KiB); pruning Docker build cache"
  docker builder prune -af >/dev/null 2>&1 || true
  free_kb="$(df -Pk "$ROOT" | awk 'NR==2 {print $4}')"
fi
if [[ "$free_kb" =~ ^[0-9]+$ ]] && (( free_kb < 2097152 )); then
  fail "insufficient free disk after cache prune: ${free_kb} KiB"
fi

worktree="$WORKTREE_ROOT/$target"
if [[ -e "$worktree" ]]; then
  git -C "$REPO" worktree remove --force "$worktree" >/dev/null 2>&1 || rm -rf "$worktree"
fi
git -C "$REPO" worktree prune
git -C "$REPO" worktree add --detach "$worktree" "$target" >/dev/null
cleanup_worktree() {
  git -C "$REPO" worktree remove --force "$worktree" >/dev/null 2>&1 || rm -rf "$worktree"
}
trap 'cleanup_worktree' EXIT

image="termrelay-local:git-$short"
build_date="$(date -u +'%Y-%m-%dT%H:%M:%SZ')"
log "building $image"
docker build   -f "$worktree/deploy/Dockerfile"   --build-arg VERSION="main-$short"   --build-arg COMMIT="$target"   --build-arg DATE="$build_date"   -t "$image"   "$worktree"

timestamp="$(date -u +'%Y%m%dT%H%M%SZ')"
snapshot_dir="$BACKUP_ROOT/$timestamp-$short"
mkdir -p "$snapshot_dir"
cp -p "$ENV_FILE" "$snapshot_dir/.env.before"

log "creating pre-deploy PostgreSQL backup"
COMPOSE_FILE="$COMPOSE_FILE" BACKUP_DIR="$snapshot_dir/postgres" BACKUP_RETENTION_DAYS=30   "$worktree/deploy/backup-postgres.sh"

previous_ref="$(awk -F= '$1=="TERMRELAY_IMAGE_REF"{sub(/^[^=]*=/,""); print; exit}' "$ENV_FILE")"
if [[ -z "$previous_ref" ]]; then
  previous_ref="$(docker inspect -f '{{.Config.Image}}' sub2api 2>/dev/null || true)"
fi
[[ -n "$previous_ref" ]] || fail "could not determine previous image reference"
printf '%s\n' "$previous_ref" > "$snapshot_dir/previous-image-ref"

set_image_ref() {
  local value="$1"
  if grep -q '^TERMRELAY_IMAGE_REF=' "$ENV_FILE"; then
    sed -i "s|^TERMRELAY_IMAGE_REF=.*$|TERMRELAY_IMAGE_REF=$value|" "$ENV_FILE"
  else
    printf '\nTERMRELAY_IMAGE_REF=%s\n' "$value" >> "$ENV_FILE"
  fi
}

compose() {
  docker compose     --project-directory "$ROOT"     --env-file "$ENV_FILE"     -f "$COMPOSE_FILE"     -f "$COMPOSE_OVERLAY"     "$@"
}

wait_healthy() {
  local tries="${1:-75}" status
  for _ in $(seq 1 "$tries"); do
    status="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' sub2api 2>/dev/null || true)"
    if [[ "$status" == "healthy" ]]; then
      curl -fsS --max-time 10 "$HEALTH_URL" >/dev/null && return 0
    fi
    if [[ "$status" == "unhealthy" || "$status" == "exited" || "$status" == "dead" ]]; then
      return 1
    fi
    sleep 2
  done
  return 1
}

runtime_frontend="$ROOT/data/frontend"
retired_frontend=""
rollback() {
  local rc="$?"
  trap - ERR
  log "deployment failed; rolling back to $previous_ref"
  cp -p "$snapshot_dir/.env.before" "$ENV_FILE"
  if [[ -n "$retired_frontend" && -d "$retired_frontend" && ! -e "$runtime_frontend" ]]; then
    mv "$retired_frontend" "$runtime_frontend" || true
  fi
  compose up -d --no-deps --force-recreate sub2api || true
  if wait_healthy 60; then
    log "rollback healthy"
  else
    log "CRITICAL: rollback did not become healthy"
  fi
  printf '%s\n' "$target" > "$STATE_DIR/last-failed-sha"
  docker logs --tail 120 sub2api 2>&1 > "$snapshot_dir/failed-container.log" || true
  exit "$rc"
}
trap rollback ERR

set_image_ref "$image"
log "starting target container"
compose up -d --no-deps --force-recreate sub2api
wait_healthy 75

running_ref="$(docker inspect -f '{{.Config.Image}}' sub2api)"
if [[ "$running_ref" != "$image" ]]; then
  log "container started unexpected image: $running_ref"
  false
fi

curl -fsS --max-time 15 "$PUBLIC_HEALTH_URL" >/dev/null

# Retire the temporary external frontend override after the new backend/image is
# healthy. From this point the frontend is the one embedded in the same commit.
if [[ -f "$runtime_frontend/index.html" ]]; then
  retired_frontend="$snapshot_dir/runtime-frontend.previous"
  log "retiring runtime frontend override"
  mv "$runtime_frontend" "$retired_frontend"
  curl -fsS --max-time 15 http://127.0.0.1:8080/ >/dev/null
fi

# Keep the checked-out production source at the exact commit that is running.
git -C "$REPO" reset --hard "$target" >/dev/null
printf '%s\n' "$target" > "$STATE_DIR/last-successful-sha"
printf '%s\n' "$previous_ref" > "$STATE_DIR/previous-image-ref"
rm -f "$STATE_DIR/last-failed-sha"

trap - ERR
log "deployment succeeded: $target ($image)"

# Adopt a validated source-controlled deploy script for the next run. The
# currently executing shell keeps this invocation stable while the installed
# path is atomically replaced for future cron executions.
self_source="$worktree/deploy/termrelay-auto-deploy.sh"
if [[ -f "$self_source" ]]; then
  bash -n "$self_source"
  install -m 0750 "$self_source" "$ROOT/autodeploy/.termrelay-auto-deploy.sh.next"
  mv "$ROOT/autodeploy/.termrelay-auto-deploy.sh.next" "$ROOT/autodeploy/termrelay-auto-deploy.sh"
fi

cleanup_worktree
trap - EXIT

# Keep disk use bounded without removing tagged rollback images.
docker image prune -f >/dev/null 2>&1 || true
