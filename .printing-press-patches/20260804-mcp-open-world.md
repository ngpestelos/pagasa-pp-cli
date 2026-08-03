# Patch: MCP openWorldHint for novel network tools (20260804)

## Intent

Issue **#24** / #19 F4: Cobra walker set `readOnlyHint` without `openWorldHint` for novel scrapers, so hosts could treat PAGASA HTTP as local-only reads.

## Changes

| Path | Change |
|------|--------|
| `internal/mcp/cobratree/classify.go` | `mcp:open-world` + `isMCPOpenWorld` |
| `internal/mcp/cobratree/walker.go` | RO+open-world / RO+local / may-write+open-world / local-write tiers |
| `internal/cli/{forecast,storm,approach,watch,weather,now,digest,obs}.go` | annotate network tools |
| tests | `TestRegisterAll_OpenWorldHints`, extend now/digest annotation test |

## On reprint

- Keep `OpenWorldAnnotation` and walker branches.
- Re-apply `mcp:open-world` on live scrapers if annotations reset.
