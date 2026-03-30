# Maintainers

This document defines the maintainer-facing workflow for releases, detection-rule changes, and general repository stewardship.

## Ownership

Before `v1.0.0`, release and repository ownership is maintained on a best-effort basis by the project maintainer(s) with write access to:

- `Sophylax/envguard`
- `Sophylax/homebrew-tap`
- `Sophylax/scoop-bucket`

Maintainers are responsible for:

- triaging issues and pull requests
- reviewing behavior changes that affect detection accuracy or hook reliability
- publishing tagged releases
- keeping release automation credentials and package publishing targets working

## Release Workflow

The release path is tag-driven through GitHub Actions and GoReleaser.

Before cutting a release:

- ensure `main` is green in CI
- confirm release notes and changelog inputs are acceptable
- verify docs and install instructions match the current publish targets
- confirm `GH_PAT` is present and still has access to the main repo, Homebrew tap, and Scoop bucket

To cut a release:

1. choose the semver tag
2. create and push the tag
3. monitor the `Release` workflow
4. verify GitHub release assets, checksums, Homebrew formula updates, and Scoop manifest updates
5. spot-check install flows from the published channels

If a release fails partway through, fix forward with a new tag unless there is a compelling reason to delete and recreate release state.

## Detection Rule Changes

Detection rules directly affect trust in `envguard`, so they should be changed conservatively.

For built-in pattern changes:

- require tests for true positives and at least one clear negative case
- document the behavior change in the pull request summary
- call out likely false-positive or false-negative tradeoffs
- avoid broad regex expansions without concrete fixture coverage

For entropy behavior changes:

- include regression coverage for both triggering and suppression behavior
- explain expected impact on noisy fixture data, generated files, and normal source text
- prefer narrowly scoped heuristics over threshold changes that shift behavior globally

For default config changes:

- treat them as product-surface changes, not silent refactors
- update `README.md`, `.envguard.yml.example`, and relevant tests in the same pull request

## Review Expectations

Maintainers should prioritize:

- hook correctness
- staged-file behavior
- fingerprint stability
- cross-platform behavior
- release and packaging integrity

Changes that affect the default scanning surface, allowlist semantics, or hook installation behavior should not merge without explicit review and verification notes.
