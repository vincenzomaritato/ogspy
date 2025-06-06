# Security Policy

## Supported Versions

| Version | Supported |
| ------- | --------- |
| `1.x`   | ✅ Yes    |
| `< 1.0` | ❌ No     |

Only the latest minor/patch release of each supported major version receives security updates.

## Reporting a Vulnerability

Please **do not open public GitHub issues** for security problems.

1. Email **<hello@vmaritato.com>** with the subject line `ogspy security <short summary>`.
2. Provide a clear, reproducible proof-of-concept or description.
3. Include any CVE reference if you already have one.

You will receive an acknowledgment within **48 hours** (working days).

## Disclosure Process

1. We confirm the vulnerability and develop a fix.
2. We prepare a coordinated disclosure date, usually within **90 days**.
3. On the release date we tag a new version, publish signed binaries and update the changelog.
4. A CVE will be requested (or updated) if the severity warrants it.
5. Credit is given to the reporter unless anonymity is requested.

## GPG / OIDC

All release artefacts are signed with [Sigstore Cosign] for supply-chain integrity.  
If encrypted email is preferred, request a temporary PGP key via the address above.

Thank you for helping keep **ogspy** and its users safe!
