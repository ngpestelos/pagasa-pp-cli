# Patch: hand-maintained CHANGELOG.md (20260804)

## Intent

This public fork maintains **`CHANGELOG.md`** as the release SSOT (Keep a Changelog + semver tags). GoReleaser auto-changelog stays disabled; GitHub Release bodies are pasted from the versioned section.

## Files

| Path | Role |
|------|------|
| `CHANGELOG.md` | Ledger: `[Unreleased]` + `## [X.Y.Z] - date` |
| `AGENTS.md` § Release Ledger | Process for agents/reviewers |
| `README.md` § Changelog | Link for humans |

## On reprint / regen

- Prefer **preserve** existing `CHANGELOG.md` (merge history; never wipe released sections).
- Re-apply AGENTS/README process text if a regen restores the library “do not hand-edit” wording.
- Do not reintroduce “changelog is only printing-press-library automation” as the sole guidance for this fork.

## Related

- Fleet distribution skill `ship-go-cli-fleet` documents the same bar for other PP CLIs.
