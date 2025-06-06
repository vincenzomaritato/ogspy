# Changelog

All notable changes to **ogspy** will be documented in this file.  
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `ogspy serve` (HTTP server mode) – **experimental**.
- Webhook support for `monitor` (`--webhook-url` flag).

### Changed

- Bumped Go toolchain to 1.23.
- Improved diff rendering performance on high-frequency monitoring.

### Fixed

- Panic on malformed `<meta>` tags without `content` attribute.
- Incorrect MIME detection in `checkImage` for SVG images.

## [1.0.0] — 2025-06-06

### Added

- **inspect**: colourised table or JSON output.
- **validate**: essential & recommended tag validation (`--essentials`, `--semantic`).
- **monitor**: diff streaming (`colour | unified | JSON`).
- Advanced semantic checks (image resolution, HTTPS, aspect ratio, article fields).
- Concurrency: worker-pool for multiple URLs.
- Structured logging with `log/slog` (`--log-level`, `--log-json`).
- CI pipeline (lint, test, SLSA/Goreleaser) and Cosign-signed releases.
- Docker scratch images & Homebrew formula.
- Unit, integration and fuzz tests.

### Removed

- Inline `og:image` preview (now out-of-scope for CLI UX).

### Security

- Added `SECURITY.md` with private disclosure channel.
