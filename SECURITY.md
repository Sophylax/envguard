# Security Policy

## Supported Versions

Before `v1.0.0`, security fixes are applied on the latest supported `main`/release line rather than maintained across multiple historical branches.

## Reporting a Vulnerability

Do not open a public GitHub issue for undisclosed vulnerabilities.

Instead, report security-sensitive issues through a private GitHub security advisory:

- https://github.com/Sophylax/envguard/security/advisories/new

Please include:

- affected version or commit
- reproduction details
- impact assessment
- any suggested mitigation if known

You should receive an acknowledgement within a reasonable best-effort window. Public disclosure should wait until a fix or mitigation is available.

## Scope Notes

Security reports are especially relevant for:

- false negatives that allow obvious secrets through the hook
- allowlist bypasses
- hook execution paths that can be subverted unexpectedly
- release or packaging integrity issues

General detection gaps, feature requests, and false positives should use the normal issue tracker unless they expose a clear security vulnerability.
