# envguard

[![CI](https://github.com/Sophylax/envguard/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/Sophylax/envguard/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Sophylax/envguard)](https://github.com/Sophylax/envguard/releases)
[![Go Version](https://img.shields.io/badge/go-1.22%2B-00ADD8)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Zero-config pre-commit secret scanning with instant local feedback.

`envguard` installs as a local Git pre-commit hook and blocks commits that contain likely secrets before they ever leave your machine. It is designed for fast developer-local feedback, not CI policy enforcement.

![envguard terminal demo](assets/envguard-demo.gif)

## Install

Choose the install path that fits your environment.

### Homebrew

```bash
brew tap Sophylax/homebrew-tap
brew install envguard
```

Homebrew formula:
- https://github.com/Sophylax/homebrew-tap/blob/main/envguard.rb

### Scoop

```powershell
scoop bucket add Sophylax https://github.com/Sophylax/scoop-bucket
scoop install envguard
```

Scoop manifest:
- https://github.com/Sophylax/scoop-bucket/blob/main/envguard.json

### Direct binary download

Download the matching archive for your platform from [GitHub Releases](https://github.com/Sophylax/envguard/releases), extract it, and place `envguard` in your `PATH`.

### Go install

```bash
go install github.com/sophylax/envguard@latest
```

## Quick Start

```bash
envguard install
git add .
git commit -m "ship it"
```

If you need to install over an existing non-envguard pre-commit hook in automation:

```bash
envguard install --yes
```

## How It Works

`envguard` runs in your local Git pre-commit hook and scans only the content relevant to your commit by default. It combines two detection engines so it catches both obvious credentials and secrets that do not follow a known vendor format.

The pattern engine looks for known token layouts and sensitive assignment shapes such as AWS keys, bearer tokens, Slack tokens, private keys, inline database credentials, and staged `.env` files. Built-in rules ship out of the box, and `.envguard.yml` can append custom regex patterns without changing code.

The entropy engine tokenizes each scanned line, measures Shannon entropy, and flags suspicious high-randomness values that exceed the configured threshold. This catches opaque secrets that look like random blobs, while simple heuristics suppress normal lowercase words and other low-risk text.

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

`exclude_paths` is applied before scanning starts. If a path is excluded there, `entropy_exclude_paths` never runs for it. By default, `testdata/**` stays included for pattern scanning and is excluded from entropy scanning through `entropy_exclude_paths`.

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

Print the build version injected at compile time.

## Allowlist Workflow

When envguard blocks a commit, it prints a stable fingerprint for each finding. Review the finding, confirm it is a false positive or accepted fixture, then add it:

```bash
envguard allow a3f9c2b1d8e04f11
git add .envguard-ignore
git commit -m "allow known test secret"
```

The `.envguard-ignore` file is newline-separated and intended to be committed so the team shares the same allowlist.

## Release Channels

- GitHub Releases: https://github.com/Sophylax/envguard/releases
- Homebrew tap: https://github.com/Sophylax/homebrew-tap
- Scoop bucket: https://github.com/Sophylax/scoop-bucket

## Comparison

| Tool | Zero-config | Fast local pre-commit UX | Developer-local focus | CI / audit depth |
| --- | --- | --- | --- | --- |
| `envguard` | Yes | Yes | Yes | Limited by design |
| `gitleaks` | No | Good, but usually tuned for CI and policy workflows | Partial | Strong |
| `git-secrets` | No | Good | Yes | Narrower pattern coverage |

## Project Policies

- [CONTRIBUTING.md](CONTRIBUTING.md)
- [SECURITY.md](SECURITY.md)
- [SUPPORT.md](SUPPORT.md)

## License

MIT
