# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

**How this file is maintained** (this public fork, not printing-press-library automation):

1. **During work** — every user-visible change lands under `[Unreleased]` in the same PR as the code (categories below).
2. **On release** — move `[Unreleased]` items into a new `## [X.Y.Z] - YYYY-MM-DD` section (Manila calendar date), leave `[Unreleased]` empty (or with a placeholder comment), tag `vX.Y.Z`, and paste that section into the GitHub Release body.
3. **Do not** rely on GoReleaser’s auto-changelog (`changelog.disable: true` in `.goreleaser.yaml`); this file is the SSOT.

### Categories

| Section | Use for |
|---------|---------|
| **Added** | New commands, flags, surfaces, install paths |
| **Changed** | Behavior or defaults that already existed |
| **Fixed** | Bugs (non-security) |
| **Security** | Auth, integrity, MCP hint honesty, supply chain |
| **Deprecated** | Still works; will remove later |
| **Removed** | Gone |

---

## [Unreleased]

### Security

- **install.sh**: prebuilt path verifies release `checksums.txt` SHA256 before extract; extracts only `pagasa-pp-cli` + `pagasa-pp-mcp`; fails closed on mismatch or missing checksum entry ([#31](https://github.com/ngpestelos/pagasa-pp-cli/pull/31), [#21](https://github.com/ngpestelos/pagasa-pp-cli/issues/21)).
- **install.sh**: harden fail-closed path under `set -euo pipefail` (awk field match + loud `die` on missing entry); soft-skip prebuilt when no digest tool; reject non-regular (symlink) extract members; `go install` installs CLI and MCP ([#32](https://github.com/ngpestelos/pagasa-pp-cli/pull/32)).

### Fixed

- **docs**: document `obs --capture` vs `--limit` footgun in README ([#20](https://github.com/ngpestelos/pagasa-pp-cli/issues/20)).

<!-- Open work lands here in the same PR as the code. Example: #33 (now/digest MCP annotations) when it merges. -->

---

## [0.1.4] - 2026-08-03

### Added

- **obs**: PAGASA Automated Weather Stations (bagong host) — live table, `--station` filter, `--capture` (local `aws_obs` + 14-day prune), `obs history` ([#17](https://github.com/ngpestelos/pagasa-pp-cli/issues/17), [#18](https://github.com/ngpestelos/pagasa-pp-cli/pull/18)).

### Fixed

- Review hardening for AWS obs: empty history JSON `[]`, `GetNoCache` for AWS fetch, `LIKE` + `ESCAPE`, `--capture` incompatible with `--limit`, honest MCP annotations on `obs` / `obs history`.

### Changed

- Dependencies: `modernc.org/sqlite` 1.55.0, `github.com/mark3labs/mcp-go` 0.57.0.

---

## [0.1.3] - 2026-07-26

### Added

- **storm**: center, intensity, movement, forecast track, and `wind_signals` from bulletin HTML.
- **watch**: match locality against HTML `wind_signals` (`signal`, `signal_matched`); unmatched is “not confirmed,” not clear.
- **digest / storm / watch / approach**: independent pages fetched in parallel.
- **approach**: bulletin eye/center coords when synopsis lacks them.

### Fixed

- **version**: local installs no longer stuck on `0.0.0-dev` (module/VCS/`git describe` fallback; Makefile ldflags) ([#13](https://github.com/ngpestelos/pagasa-pp-cli/pull/13)).
- Skill densify (v1.3.0) and family-safety hard rule for empty wind signals.

---

## [0.1.2] - 2026-07-23

### Fixed

- Release: drop deprecated `brews:` / homebrew tap from GoReleaser (failed CI when tap repo missing; red run even when binaries uploaded).

---

## [0.1.1] - 2026-07-23

### Added

- One-shot fleet installer (`scripts/install.sh`); prefers prebuilt release tarball over on-host compile.

### Fixed

- Release: prebuilt binaries via goreleaser CI for fleet download (avoid compiling `modernc.org/sqlite` on small hosts).
- Release: full-module ldflags path so released binaries embed `--version` correctly.

---

## [0.1.0] - 2026-07-23

### Added

- First public release — agent-native CLI for PAGASA public weather.
- Commands: `now`, `forecast`, `storm`, `approach`, `watch`, `digest`, `history`, `drift`.
- Local SQLite mirror for synopsis history and forecast drift.
- MCP companion binary (`pagasa-pp-mcp`).

---

[Unreleased]: https://github.com/ngpestelos/pagasa-pp-cli/compare/v0.1.4...HEAD
[0.1.4]: https://github.com/ngpestelos/pagasa-pp-cli/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/ngpestelos/pagasa-pp-cli/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/ngpestelos/pagasa-pp-cli/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/ngpestelos/pagasa-pp-cli/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/ngpestelos/pagasa-pp-cli/releases/tag/v0.1.0
