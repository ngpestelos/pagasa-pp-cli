# Patch: block MCP shell-out `--rate-limit` (20260804)

## Intent

Issue **#25** / security review #19 F5: agents must not pass `--rate-limit 0` (or any override) through mirrored Cobra MCP tools. Typed MCP tools hardcode polite rate 2; shell-out must keep the CLI default (2) by blocking the root flag.

## Files

| Path | Change |
|------|--------|
| `internal/mcp/cobratree/shellout.go` | `"rate-limit": true` in `blockedRootFlags` |
| `internal/mcp/cobratree/shellout_test.go` | Live blocklist pin + `TestBlockedRootFlags_IncludesRateLimit` |

## On reprint / regen

- Re-apply `rate-limit` to `blockedRootFlags` if the generator resets the map.
- Prefer upstream Printing Press template blocklist includes `rate-limit` so all printed CLIs inherit the politeness contract.
