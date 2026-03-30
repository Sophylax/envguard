#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
demo_root="$repo_root/.demo/envguard-demo"

rm -rf "$demo_root"
mkdir -p "$demo_root"
cd "$demo_root"

git init -b main >/dev/null
git config user.name demo
git config user.email demo@example.com

cat > app.js <<'EOF'
const key = AKIA1234567890ABCDEF;
EOF

git add app.js

set +e
json_output="$("$repo_root/bin/envguard" check app.js --json 2>/dev/null)"
scan_status=$?
set -e

if [ "$scan_status" -eq 0 ]; then
  echo "expected demo scan to report a finding" >&2
  exit 1
fi

fingerprint="$(printf '%s\n' "$json_output" | sed -n 's/.*"fingerprint": "\([a-f0-9]\{16\}\)".*/\1/p' | head -n 1)"

if [ -z "$fingerprint" ]; then
  echo "failed to compute demo fingerprint" >&2
  exit 1
fi

rm -f app.js
git reset --hard >/dev/null

sed \
  -e "s|{{FINGERPRINT}}|$fingerprint|g" \
  "$repo_root/demo/envguard.tape.tmpl" > "$repo_root/demo/envguard.tape"
