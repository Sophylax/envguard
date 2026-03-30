#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
tmp_repo="$(mktemp -d)"
trap 'rm -rf "$tmp_repo"' EXIT

cd "$tmp_repo"
git init -b main >/dev/null
git config user.name "envguard-ci"
git config user.email "envguard-ci@example.com"

cat > secret.js <<'EOF'
const key = "AKIA1234567890ABCDEF";
EOF

git add secret.js
PATH="$repo_root/bin:$PATH" "$repo_root/bin/envguard" install >/dev/null

set +e
commit_output="$(PATH="$repo_root/bin:$PATH" git commit -m "blocked" 2>&1)"
commit_status=$?
set -e

if [ "$commit_status" -eq 0 ]; then
  echo "expected commit with staged secret to be blocked"
  exit 1
fi

printf '%s\n' "$commit_output" | grep -q "Commit blocked."

set +e
json_output="$(PATH="$repo_root/bin:$PATH" "$repo_root/bin/envguard" check secret.js --json)"
json_status=$?
set -e

if [ "$json_status" -eq 0 ]; then
  echo "expected JSON check to report findings before allowlisting"
  exit 1
fi

fingerprint="$(printf '%s\n' "$json_output" | sed -n 's/.*"fingerprint": "\([a-f0-9]\{16\}\)".*/\1/p' | head -n 1)"
if [ -z "$fingerprint" ]; then
  echo "failed to extract finding fingerprint from JSON output"
  exit 1
fi

PATH="$repo_root/bin:$PATH" "$repo_root/bin/envguard" allow "$fingerprint" >/dev/null
git add .envguard-ignore
PATH="$repo_root/bin:$PATH" git commit -m "allow known fixture" >/dev/null 2>&1

printf '\necho foreign-hook\n' >> .git/hooks/pre-commit
PATH="$repo_root/bin:$PATH" "$repo_root/bin/envguard" uninstall >/dev/null

grep -q "foreign-hook" .git/hooks/pre-commit
if grep -q "envguard check" .git/hooks/pre-commit; then
  echo "envguard uninstall left hook content behind"
  exit 1
fi
