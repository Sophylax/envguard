# envguard

[![CI](https://github.com/Sophylax/envguard/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/Sophylax/envguard/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Sophylax/envguard)](https://github.com/Sophylax/envguard/releases)
[![Go Version](https://img.shields.io/badge/go-1.22%2B-00ADD8)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

You meant to commit code, not a key.

`envguard` adds secret scanning to Git at the one moment it matters most: right before the commit leaves your machine.

One command installs a local pre-commit hook that blocks likely secrets with instant feedback, no CI round-trip, no YAML tax.

![envguard terminal demo](assets/envguard-demo.gif)

## Quick Start

```bash
envguard install
git add .
git commit -m "ship it"
```

## Install

### Homebrew

```bash
brew tap Sophylax/homebrew-tap
brew install envguard
```

### Scoop

```powershell
scoop bucket add Sophylax https://github.com/Sophylax/scoop-bucket
scoop install envguard
```

### Direct binary download

Download the matching archive for your platform from [GitHub Releases](https://github.com/Sophylax/envguard/releases), extract it, and place `envguard` in your `PATH`.

### Go install

```bash
go install github.com/sophylax/envguard@latest
```

## How It Works

Most secret scanners live in CI. By then the leak already happened.

`envguard` is built for the opposite direction: fast local feedback before the push, before the PR, before the cleanup commit.

The pattern engine catches known token shapes and sensitive assignment patterns such as AWS keys, bearer tokens, Slack tokens, private keys, inline database credentials, and staged `.env` files. Built-in rules work out of the box, and `.envguard.yml` can add custom regex patterns without code changes.

The entropy engine catches what pattern matching misses: high-randomness strings that don't follow any vendor format.

## Config Reference

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `entropy_threshold` | `float64` | `4.5` | Minimum Shannon entropy required to report a token. |
| `min_length` | `int` | `20` | Minimum token length considered by entropy scanning. |
| `max_file_size_kb` | `int` | `500` | Skip files larger than this limit with a warning. |
| `exclude_paths` | `[]string` | `["**/*.test.js","vendor/**"]` | Glob patterns excluded from scanning. |
| `exclude_extensions` | `[]string` | `[".lock",".svg",".png"]` | File extensions excluded from scanning. |
| `entropy_exclude_paths` | `[]string` | `["testdata/**"]` | Glob patterns that skip entropy scanning only while keeping pattern matching enabled for files that are not excluded by `exclude_paths`. |
| `custom_patterns` | `[]pattern` | `[]` | Extra regex rules added to the built-in pattern library. |

Example:

```yaml
entropy_threshold: 4.5
min_length: 20
max_file_size_kb: 500
exclude_paths:
  - "**/*.test.js"
  - "vendor/**"
exclude_extensions:
  - ".lock"
  - ".png"
  - ".svg"
entropy_exclude_paths:
  - "testdata/**"
custom_patterns:
  - name: "Internal Token"
    pattern: "MYCO_[A-Z0-9]{32}"
    severity: "HIGH"
```

`exclude_paths` takes priority. Paths matched there are skipped entirely, including entropy scanning.

## CLI Reference

### `envguard check [path]`

Scan staged files by default, or a file or directory when `path` is provided.

Flags:
- `--all`: scan the entire working tree.
- `--json`: output findings as a JSON array.
- `--severity HIGH|MEDIUM|LOW`: filter displayed findings by severity.

Exit code:
- `0`: clean scan.
- `1`: findings detected.

Notes:
- when no path is provided, `envguard` scans staged files from `git diff --cached --name-only`
- files larger than `max_file_size_kb` are skipped with a warning instead of failing the scan

### `envguard install`

Install the Git pre-commit hook in the current repository. If a non-envguard hook already exists, `envguard` prompts before prepending itself in interactive use and fails fast in non-interactive contexts unless `--yes` is passed.

Flags:
- `-y`, `--yes`: prepend envguard to an existing foreign hook without prompting.

### `envguard uninstall`

Remove `envguard` from the pre-commit hook. If `envguard` was the only content, the hook file is deleted. Running it repeatedly is safe.

### `envguard allow <fingerprint>`

Add a finding fingerprint to `.envguard-ignore` in the repository root.

### `envguard version`

Print the resolved build or module version.

## Allowlist Workflow

When `envguard` blocks a commit, it prints a stable fingerprint for each finding. If the finding is a false positive or an intentional fixture, allow it explicitly:

```bash
envguard allow a3f9c2b1d8e04f11
git add .envguard-ignore
git commit -m "allow known test secret"
```

The `.envguard-ignore` file is newline-separated and intended to be committed so the team shares the same allowlist.

## Comparison

If you already use Gitleaks or `git-secrets`, the useful way to think about `envguard` is not as a replacement for CI scanning, but as the fast local checkpoint that stops the mistake before CI ever has to complain.

| Tool | Zero-config | Fast local pre-commit UX | Developer-local focus | CI / audit depth |
| --- | --- | --- | --- | --- |
| `envguard` | Yes | Yes | Yes | Limited by design |
| `gitleaks` | No | Good, but usually tuned for CI and policy workflows | Partial | Strong |
| `git-secrets` | No | Good | Yes | Narrower pattern coverage |

## Project Policies

See [CONTRIBUTING.md](CONTRIBUTING.md) · [SECURITY.md](SECURITY.md) · [SUPPORT.md](SUPPORT.md) · [MAINTAINERS.md](MAINTAINERS.md).

## License

MIT
