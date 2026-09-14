#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
overlay="$repo_root/deploy/docker-compose.termrelay.yml"

if [ ! -f "$overlay" ]; then
  echo "missing TermRelay production compose overlay" >&2
  exit 1
fi

if ! grep -Fq 'image: ${TERMRELAY_IMAGE_REF:?' "$overlay"; then
  echo "TermRelay production overlay must require an explicit TERMRELAY_IMAGE_REF" >&2
  exit 1
fi

if grep -Eq 'weishaw/sub2api:(latest|main)' "$overlay"; then
  echo "TermRelay production overlay must not pin production to a mutable upstream image" >&2
  exit 1
fi

if ! grep -Fq 'ghcr.io/yuchenm1303-png/sub2api' "$overlay"; then
  echo "TermRelay production overlay should document the repository-owner GHCR image source" >&2
  exit 1
fi

printf 'TermRelay production image source test passed\n'
