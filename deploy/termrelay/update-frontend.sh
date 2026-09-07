#!/usr/bin/env bash
set -Eeuo pipefail

DEFAULT_BRANCH="release/smirel-commercial-dark"
BRANCH="$DEFAULT_BRANCH"
DRY_RUN=false
ROLLBACK=false
CONTAINER_NAME="${TERMRELAY_CONTAINER:-termrelay-app}"
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR/../.." rev-parse --show-toplevel 2>/dev/null || true)"
DATA_DIR="${TERMRELAY_DATA_DIR:-$SCRIPT_DIR/data/app}"
RELEASES_DIR=""
ACTIVE_LINK=""
PREVIOUS_LINK=""
TMP_ROOT=""

log() {
  printf '[frontend] %s\n' "$*"
}

die() {
  printf '[frontend] ERROR: %s\n' "$*" >&2
  exit 1
}

usage() {
  cat <<'EOF'
Usage:
  ./update-frontend.sh [--branch <branch>] [--dry-run]
  ./update-frontend.sh --rollback [--dry-run]

Options:
  --branch <branch>   UI source branch (default: release/smirel-commercial-dark)
  --dry-run           Resolve and print the action without building or switching UI
  --rollback          Atomically switch back to the previous frontend release
  -h, --help          Show this help

Environment:
  TERMRELAY_CONTAINER  Application container name (default: termrelay-app)
  TERMRELAY_DATA_DIR   Host app data directory (default: deploy/termrelay/data/app)
EOF
}

while (($#)); do
  case "$1" in
    --branch)
      (($# >= 2)) || die "--branch requires a value"
      BRANCH="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    --rollback)
      ROLLBACK=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown argument: $1"
      ;;
  esac
done

[[ -n "$REPO_ROOT" && -d "$REPO_ROOT/.git" ]] || die "TermRelay git repository not found from $SCRIPT_DIR"
for cmd in git docker tar mktemp readlink ln mv cp grep; do
  command -v "$cmd" >/dev/null 2>&1 || die "required command not found: $cmd"
done

docker info >/dev/null 2>&1 || die "Docker daemon is not available to the current user"

mkdir -p "$DATA_DIR"
[[ -w "$DATA_DIR" ]] || die "data directory is not writable: $DATA_DIR"
RELEASES_DIR="$DATA_DIR/frontend-releases"
ACTIVE_LINK="$DATA_DIR/frontend"
PREVIOUS_LINK="$DATA_DIR/frontend.previous"

validate_release_dir() {
  local release_dir="$1"
  [[ -s "$release_dir/index.html" ]] || return 1
  [[ -d "$release_dir/assets" ]] || return 1
  return 0
}

link_target_path() {
  local target="$1"
  if [[ "$target" = /* ]]; then
    printf '%s\n' "$target"
  else
    printf '%s/%s\n' "$DATA_DIR" "$target"
  fi
}

atomic_set_link() {
  local link_path="$1"
  local target="$2"
  local tmp_link="${link_path}.next.$$"
  rm -f "$tmp_link"
  ln -s "$target" "$tmp_link"
  mv -Tf "$tmp_link" "$link_path"
}

remove_active_link() {
  if [[ -L "$ACTIVE_LINK" ]]; then
    rm -f "$ACTIVE_LINK"
  fi
}

extract_asset_ref() {
  local release_dir="$1"
  grep -oE 'assets/[A-Za-z0-9._/-]+' "$release_dir/index.html" | head -n 1 || true
}

verify_runtime_release() {
  local release_dir="$1"
  local asset_ref
  local body

  validate_release_dir "$release_dir" || return 1
  docker inspect "$CONTAINER_NAME" >/dev/null 2>&1 || return 1
  [[ "$(docker inspect -f '{{.State.Running}}' "$CONTAINER_NAME" 2>/dev/null)" == "true" ]] || return 1

  docker exec "$CONTAINER_NAME" sh -c 'test -s /app/data/frontend/index.html' >/dev/null 2>&1 || return 1
  docker exec "$CONTAINER_NAME" wget -q -T 5 -O /dev/null http://127.0.0.1:8080/health >/dev/null 2>&1 || return 1

  asset_ref="$(extract_asset_ref "$release_dir")"
  [[ -n "$asset_ref" ]] || return 1
  body="$(docker exec "$CONTAINER_NAME" wget -q -T 5 -O - http://127.0.0.1:8080/ 2>/dev/null)" || return 1
  grep -Fq "$asset_ref" <<<"$body" || return 1
  return 0
}

rollback_frontend() {
  [[ -L "$ACTIVE_LINK" ]] || die "no active external frontend symlink: $ACTIVE_LINK"
  [[ -L "$PREVIOUS_LINK" ]] || die "no previous frontend release is recorded"

  local current_target previous_target current_dir previous_dir
  current_target="$(readlink "$ACTIVE_LINK")"
  previous_target="$(readlink "$PREVIOUS_LINK")"
  current_dir="$(link_target_path "$current_target")"
  previous_dir="$(link_target_path "$previous_target")"

  validate_release_dir "$previous_dir" || die "previous release is invalid: $previous_dir"

  log "rollback current:  $current_target"
  log "rollback target:   $previous_target"
  log "container:         $CONTAINER_NAME"
  if $DRY_RUN; then
    log "dry-run: no symlink will be changed"
    return 0
  fi

  atomic_set_link "$ACTIVE_LINK" "$previous_target"
  if ! verify_runtime_release "$previous_dir"; then
    atomic_set_link "$ACTIVE_LINK" "$current_target"
    die "rollback target failed runtime verification; restored current frontend"
  fi

  atomic_set_link "$PREVIOUS_LINK" "$current_target"
  log "rollback complete"
}

if $ROLLBACK; then
  rollback_frontend
  exit 0
fi

[[ -n "$BRANCH" ]] || die "branch must not be empty"
log "fetching UI branch: $BRANCH"
git -C "$REPO_ROOT" fetch --prune origin "$BRANCH"
UI_SHA="$(git -C "$REPO_ROOT" rev-parse "origin/$BRANCH^{commit}")"
UI_SHORT="${UI_SHA:0:12}"
RELEASE_DIR="$RELEASES_DIR/$UI_SHA"
RELATIVE_RELEASE_TARGET="frontend-releases/$UI_SHA"

CURRENT_TARGET=""
if [[ -L "$ACTIVE_LINK" ]]; then
  CURRENT_TARGET="$(readlink "$ACTIVE_LINK")"
elif [[ -e "$ACTIVE_LINK" ]]; then
  die "$ACTIVE_LINK exists but is not a symlink; refusing to overwrite it"
fi

log "branch:            $BRANCH"
log "UI commit:         $UI_SHA"
log "release directory: $RELEASE_DIR"
log "active link:       $ACTIVE_LINK"
log "current target:    ${CURRENT_TARGET:-embedded frontend}"
log "container:         $CONTAINER_NAME"

if $DRY_RUN; then
  if validate_release_dir "$RELEASE_DIR"; then
    log "dry-run: release already built; would atomically switch symlink"
  else
    log "dry-run: would build frontend only from origin/$BRANCH"
    log "dry-run: backend image/container, PostgreSQL, Redis and Caddy would not be changed"
  fi
  exit 0
fi

if [[ "$CURRENT_TARGET" == "$RELATIVE_RELEASE_TARGET" ]] && validate_release_dir "$RELEASE_DIR"; then
  if verify_runtime_release "$RELEASE_DIR"; then
    log "already running UI $UI_SHORT"
    exit 0
  fi
  die "active symlink points to $UI_SHORT but runtime verification failed"
fi

mkdir -p "$RELEASES_DIR"

if ! validate_release_dir "$RELEASE_DIR"; then
  TMP_ROOT="$(mktemp -d -t termrelay-frontend.XXXXXXXX)"
  cleanup() {
    if [[ -n "${TMP_ROOT:-}" && -d "$TMP_ROOT" ]]; then
      rm -rf "$TMP_ROOT"
    fi
  }
  trap cleanup EXIT

  SRC_DIR="$TMP_ROOT/src"
  mkdir -p "$SRC_DIR"
  log "exporting source at $UI_SHORT"
  git -C "$REPO_ROOT" archive --format=tar "origin/$BRANCH" | tar -xf - -C "$SRC_DIR"

  [[ -f "$SRC_DIR/frontend/package.json" ]] || die "frontend/package.json missing in $BRANCH"
  [[ -f "$SRC_DIR/frontend/pnpm-lock.yaml" ]] || die "frontend/pnpm-lock.yaml missing in $BRANCH"

  log "building frontend in isolated Node container"
  docker run --rm \
    -e HOST_UID="$(id -u)" \
    -e HOST_GID="$(id -g)" \
    -v "$SRC_DIR:/repo" \
    -v termrelay-pnpm-store:/root/.local/share/pnpm/store \
    -w /repo/frontend \
    node:24-alpine \
    sh -eu -c '
      corepack enable
      corepack prepare pnpm@9 --activate
      pnpm install --frozen-lockfile --prefer-offline
      VITE_STANDALONE=true VITE_API_BASE_URL=/api/v1 pnpm run build
      chown -R "$HOST_UID:$HOST_GID" dist
    '

  DIST_DIR="$SRC_DIR/frontend/dist"
  validate_release_dir "$DIST_DIR" || die "frontend build did not produce a valid dist bundle"
  if grep -R -Fq '__preview_api__' "$DIST_DIR"; then
    die "preview API marker found in production frontend build"
  fi

  STAGE_DIR="$RELEASES_DIR/.${UI_SHA}.tmp.$$"
  rm -rf "$STAGE_DIR"
  mkdir -p "$STAGE_DIR"
  cp -a "$DIST_DIR/." "$STAGE_DIR/"
  chmod -R a+rX "$STAGE_DIR"
  validate_release_dir "$STAGE_DIR" || die "staged frontend release is invalid"
  mv "$STAGE_DIR" "$RELEASE_DIR"
  log "built release: $UI_SHORT"
else
  log "reusing existing release: $UI_SHORT"
fi

OLD_TARGET="$CURRENT_TARGET"
atomic_set_link "$ACTIVE_LINK" "$RELATIVE_RELEASE_TARGET"

if ! verify_runtime_release "$RELEASE_DIR"; then
  log "new frontend failed runtime verification; rolling back" >&2
  if [[ -n "$OLD_TARGET" ]]; then
    atomic_set_link "$ACTIVE_LINK" "$OLD_TARGET"
  else
    remove_active_link
  fi
  die "frontend switch failed. The previous UI has been restored. If this is the first rollout, ensure the backend includes frontend hot-swap support."
fi

if [[ -n "$OLD_TARGET" ]]; then
  atomic_set_link "$PREVIOUS_LINK" "$OLD_TARGET"
fi

log "frontend update complete"
log "branch:    $BRANCH"
log "UI commit: $UI_SHA"
log "active:    $RELATIVE_RELEASE_TARGET"
log "backend:   unchanged"
log "database:  unchanged"
log "redis:     unchanged"
log "caddy:     unchanged"
