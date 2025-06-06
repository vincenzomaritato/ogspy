# ogspy

[![CI](https://github.com/vincenzomaritato/ogspy/actions/workflows/ci.yml/badge.svg)](https://github.com/vincenzomaritato/ogspy/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vincenzomaritato/ogspy)](https://goreportcard.com/report/github.com/vincenzomaritato/ogspy)
[![Latest Release](https://img.shields.io/github/v/release/vincenzomaritato/ogspy?logo=github)](https://github.com/vincenzomaritato/ogspy/releases)

> A single-binary CLI to **inspect, validate & monitor Open Graph metadata**.  
> Built for developers, SEO specialists and content creators who want their links to shine on every social platform.

## Features

- **Inspect** — Pretty-prints all `og:*` tags or raw JSON.
- **Validate** — Ensures mandatory and recommended tags are present, exits ≠ 0 on failure.
- **Semantic checks** — Image resolution ≥ 1200×630 px, HTTPS only, correct aspect ratio, article attributes, etc.
- **Monitor** — Watches a URL at any interval, streaming diffs in colour, unified or JSON.
- **Structured logging** — Text or ND-JSON via `log/slog`; SLSA provenance and Cosign signatures for every release.
- **Cross-platform** — Static builds for Linux, macOS (AMD64 & ARM64) and Windows.

## Quick install

| Method          | Command                                                                                       |
| --------------- | --------------------------------------------------------------------------------------------- |
| **Homebrew**    | `brew install vincenzomaritato/tap/ogspy`                                                     |
| **cURL script** | `curl -fsSL https://raw.githubusercontent.com/vincenzomaritato/ogspy/main/install.sh \| bash` |
| **Go install**  | `go install github.com/vincenzomaritato/ogspy@latest`                                         |

> All binaries/tarballs are Cosign-signed and checksummed (SHA-256).

## Usage

```bash
# Inspect in a colourful table
ogspy inspect https://example.com

# JSON output (useful in CI)
ogspy inspect -j https://example.com | jq .

# Validate only essential tags
ogspy validate -e https://example.com

# Full validation with semantic checks
ogspy validate -s https://example.com

# Monitor every 5 minutes, diff as unified text
ogspy monitor -i 300 -u https://example.com
```

Run `ogspy --help` or `ogspy <command> --help` for every flag.

### Development

```bash
# Lint + test

make lint test

# Build all target binaries

make all # or: make linux-arm64, etc.

# Snapshot release artefacts under ./build

make snapshot
```

### Tests & fuzzing

```bash
go test -v ./...
go test -run=Fuzz -fuzz=FuzzParseOG -fuzztime=30s
```

## Contributing

Pull requests are welcome ❤️
Please read [CONTRIBUTING.md](CONTRIBUTING.md) and open an issue before large changes.

## Security

Found a vulnerability?
Please do not open a public issue.
Email [hello@vmaritato.com](mailto://hello@vmaritato.com) for responsible disclosure instructions.

## License

ogspy is released under the MIT license (see [LICENSE](LICENSE)).

<div align="center">
  <sub>Crafted with Go • © 2025 Vincenzo Maritato — <https://vmaritato.com></sub>
</div>
