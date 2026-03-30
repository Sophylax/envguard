# envguard

Zero-config pre-commit secret scanning with instant local feedback.

<!-- demo GIF here -->

## Install

### Homebrew

```bash
brew tap envguard/homebrew-tap
brew install envguard
```

### Scoop

```powershell
scoop bucket add sophylax https://github.com/sophylax/scoop-bucket
scoop install envguard
```

### Direct binary download

Download the matching archive from GitHub Releases and place `envguard` in your `PATH`.

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
| `exclude_paths` | `[]string` | `["testdata/**","**/*.test.js","vendor/**"]` | Glob patterns excluded from scanning. |
| `exclude_extensions` | `[]string` | `[".lock",".svg",".png"]` | File extensions excluded from scanning. |
| `custom_patterns` | `[]pattern` | `[]` | Extra regex rules added to the built-in pattern library. |
| `allow_test_fixtures` | `bool` | `false` | Skip entropy scanning for files under `testdata/`. |

Example:

```yaml
entropy_threshold: 4.5
min_length: 20
max_file_size_kb: 500
exclude_paths:
  - "testdata/**"
exclude_extensions:
  - ".png"
custom_patterns:
  - name: "Internal Token"
    pattern: "MYCO_[A-Z0-9]{32}"
    severity: "HIGH"
allow_test_fixtures: false
```

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

### `envguard install`

Installs the Git pre-commit hook in the current repository. If a non-envguard hook already exists, envguard prompts before prepending itself.

### `envguard uninstall`

Removes envguard from the pre-commit hook. If envguard was the only content, the hook file is deleted. Running it repeatedly is safe.

### `envguard allow <fingerprint>`

Adds a finding fingerprint to `.envguard-ignore` in the repository root.

### `envguard version`

Prints the build version injected at compile time.

## Allowlist Workflow

When envguard blocks a commit, it prints a stable fingerprint for each finding. Review the finding, confirm it is a false positive or accepted fixture, then add it:

```bash
envguard allow a3f9c2b1d8e04f11
git add .envguard-ignore
git commit -m "allow known test secret"
```

The `.envguard-ignore` file is newline-separated and intended to be committed so the team shares the same allowlist.

## Comparison

| Tool | Zero-config | Fast local pre-commit UX | Developer-local focus | CI / audit depth |
| --- | --- | --- | --- | --- |
| `envguard` | Yes | Yes | Yes | Limited by design |
| `gitleaks` | No | Good, but usually tuned for CI and policy workflows | Partial | Strong |
| `git-secrets` | No | Good | Yes | Narrower pattern coverage |

## Contributing

1. Fork the repo.
2. Run `make test`.
3. Add or update tests with your changes.
4. Open a pull request with a clear description and conventional commit title when possible.

## License

MIT
