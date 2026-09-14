#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
overlay="$repo_root/deploy/docker-compose.termrelay.yml"
goreleaser="$repo_root/.goreleaser.yaml"
goreleaser_simple="$repo_root/.goreleaser.simple.yaml"

if [ ! -f "$overlay" ]; then
  echo "missing TermRelay production compose overlay" >&2
  exit 1
fi

for config in "$goreleaser" "$goreleaser_simple"; do
  if [ ! -f "$config" ]; then
    echo "missing GoReleaser configuration: $config" >&2
    exit 1
  fi
done

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

# Full and simple releases must both publish the same owner-controlled,
# versioned image path used by production. Floating :latest may also exist for
# convenience, but production is never instructed to use it.
for config in "$goreleaser" "$goreleaser_simple"; do
  if ! grep -Fq 'ghcr.io/{{ .Env.GITHUB_REPO_OWNER_LOWER }}/sub2api:{{ .Version }}' "$config"; then
    echo "GoReleaser config must publish the versioned owner-controlled GHCR image: $config" >&2
    exit 1
  fi
done

printf 'TermRelay production image source test passed\n'
