# Security Policy

## Supported versions

Security fixes are released for the latest minor version. The previous minor version receives fixes for high and critical severity issues for three months after the next minor version is released.

| Version | Supported                         |
| ------- | --------------------------------- |
| 0.11.x  | Yes                               |
| 0.10.x  | High and critical fixes only      |
| < 0.10  | No                                |

Pre-releases (`-pre.*` tags) are not supported. Upgrade to the matching stable release.

## Reporting a vulnerability

**Do not report security vulnerabilities through public GitHub issues, discussions or pull requests.**

Report them privately through one of these channels:

- **GitHub private vulnerability reporting** (preferred): open a report from the [Security tab](https://github.com/terralist/terralist/security/advisories/new) of the repository.
- **Email**: [team@terralist.io](mailto:team@terralist.io).

Include as much of the following as you can:

- The affected version or commit, and the relevant configuration (OAuth provider, storage and database backends).
- A description of the vulnerability and its impact, including whether it is reachable without authentication.
- Steps to reproduce, or a proof of concept.
- Any suggested fix or mitigation.

## What to expect

- We acknowledge your report within **3 business days**.
- We send an initial assessment, including whether we accept the report and its severity, within **7 days**.
- We aim to release a fix within **90 days** of the report, and keep you updated on progress until then.

## Disclosure policy

We follow coordinated disclosure:

1. We confirm the vulnerability, determine the affected versions and prepare a fix in a private fork.
2. We release fixed versions for every supported version line that is affected.
3. We publish a [GitHub Security Advisory](https://github.com/terralist/terralist/security/advisories) and request a CVE for the issue. The release notes of the fixed versions link to the advisory.
4. We credit the reporter in the advisory, unless they prefer to stay anonymous.

Please keep the details private until the advisory is published, or until 90 days have passed since your report, whichever comes first. If you need a different timeline, tell us in your report.

## Hardening

For guidance on deploying Terralist securely, see the [security guide](https://www.terralist.io/user-guide/security/).
