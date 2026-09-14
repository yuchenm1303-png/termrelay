#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
overlay="$repo_root/deploy/docker-compose.termrelay.yml"
goreleaser="$repo_root/.goreleaser.yaml"

if [ ! -f "$overlay" ]; then
  echo "missing TermRelay production compose overlay" >&2
  exit 1
fi

if [ ! -f "$goreleaser" ]; then
  echo "missing GoReleaser configuration" >&2
  exit 1
fi

if ! grep -Fq 'image: ${TERMRELAY_IMAGE_REF:?' "$overlay"; then
  echo "TermRelay production overlay must require an explicit TERMRELAY_IMAGE_REF" >&2
  exit 1
fi

# Inspect actual Compose image directives only. Documentation comments may name
# unsafe examples such as an upstream :latest tag in order to warn against it.
if grep -Eq '^[[:space:]]*image:[[:space:]]*.*weishaw/sub2api:(latest|main)([[:space:]]|$)' "$overlay"; then
  echo "TermRelay production overlay must not pin production to a mutable upstream image" >&2
  exit 1
fi

if ! grep -Fq 'ghcr.io/yuchenm1303-png/sub2api' "$overlay"; then
  echo "TermRelay production overlay should document the repository-owner GHCR image source" >&2
  exit 1
fi

# The deployment overlay and the publisher must agree on the same image path.
# This catches a future rename that changes only one side and would make a
# documented production release impossible to pull.
if ! grep -Fq 'ghcr.io/{{ .Env.GITHUB_REPO_OWNER_LOWER }}/sub2api:{{ .Version }}' "$goreleaser"; then
  echo "GoReleaser must publish the versioned owner-controlled GHCR image used by production" >&2
  exit 1
fi

printf 'TermRelay production image source test passed\n'
