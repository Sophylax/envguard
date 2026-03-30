#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

make -C "$repo_root" build
docker run --rm \
  -v "$repo_root:/workspace" \
  alpine \
  sh -c 'rm -rf /workspace/.demo'
"$repo_root/demo/prepare-demo.sh"

docker build -f "$repo_root/demo/Dockerfile" -t envguard-demo-vhs "$repo_root"
docker run --rm \
  -v "$repo_root:/vhs" \
  -w /vhs/.demo/envguard-demo \
  envguard-demo-vhs \
  /vhs/demo/envguard.tape
