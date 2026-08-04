# Patch: MCP teach/playbook path allowlist (20260804)

## Intent

Issue **#30** / security review #19 F10: MCP-local-write teach/playbook tools must not
ReadFile arbitrary paths or open `--db` outside the app data/state directories.
CLI TTY retains full path flexibility.

## Files

| Path | Change |
|------|--------|
| `internal/cli/mcp_path_guard.go` | `rejectPathOutsideMCPAppDirs`, `resolveLearnDBPath`, root containment |
| `internal/cli/mcp_path_guard_test.go` | MCP vs CLI surface, data/state allow, traversal reject |
| `internal/cli/teach_playbook.go` | gate file paths in `resolvePlaybookInputs`; `resolveLearnDBPath` on open |
| `internal/cli/teach.go` | `resolveLearnDBPath` on learn DB opens |
| `internal/cli/learnings_*.go`, `root.go` | same DB gate for mirrored MCP learn tools |
| `CHANGELOG.md` | Unreleased Security #30 |

## On reprint / regen

- Re-apply path allowlist before any `os.ReadFile` / `store.Open*` on agent-supplied paths when `LearnEventSurface()==mcp`.
- Prefer upstream: shared helper for "path under XDG data/state" used by all MCP-exposed file flags.
