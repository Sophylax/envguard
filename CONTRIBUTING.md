# Contributing to envguard

Thanks for contributing to `envguard`.

## Before You Start

- Open or reference an issue before starting larger work so behavior changes have clear scope.
- Keep changes focused. Small, reviewable pull requests are preferred over mixed refactors.
- Add or update tests for behavior changes and bug fixes.
- Use conventional commit prefixes when practical, such as `fix:`, `feat:`, `docs:`, `test:`, or `ci:`.

## Development Setup

Requirements:

- Go 1.22+
- Git

Useful commands:

```bash
make test
make build
make build-all
```

`envguard` is designed to ship as a static binary with `CGO_ENABLED=0`. Keep new code compatible with that constraint.

## Coding Expectations

- Wrap returned errors with `fmt.Errorf("context: %w", err)`.
- Do not add runtime dependencies without clear justification.
- Preserve the zero-config local UX. Changes that add setup friction should be justified explicitly in the PR.
- Prefer repo-relative, cross-platform behavior over machine-specific assumptions.

## Tests

Before opening a pull request, run:

```bash
make test
```

Run `make build-all` when changing build, release, hook, or platform-sensitive behavior.

If you change hook installation, staged file handling, config loading, or output formatting, add regression coverage where possible.

## Pull Requests

Each pull request should include:

- a short problem statement
- a concise summary of the approach
- test coverage or manual verification notes
- linked issue references when applicable

Keep screenshots or demo assets limited to changes that actually affect user-facing output.

Maintainer-facing release and rule-update policy lives in [MAINTAINERS.md](MAINTAINERS.md).

## Scope and Support

`envguard` currently targets:

- Go 1.22+
- Linux
- macOS
- Windows

The project is optimized for developer-local pre-commit scanning. CI-oriented policy engines and hosted scanning workflows are out of scope unless the project direction changes explicitly.
