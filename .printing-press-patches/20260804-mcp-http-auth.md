# Patch: HTTP MCP auth + loopback default (20260804)

## Intent

Issue **#26** / security review #19 F6: streamable HTTP MCP must not listen
unauthenticated on all interfaces. Default transport stays **stdio** (no
socket). When `--transport http` is chosen:

1. Default `--addr` is `127.0.0.1:7777` (was `:7777`).
2. Non-loopback bind requires `--allow-remote`.
3. Every HTTP request requires `Authorization: Bearer <token>` from `--token`
   or `PP_MCP_TOKEN`; refuse start if token is empty.

## Files

| Path | Change |
|------|--------|
| `cmd/pagasa-pp-mcp/main.go` | flags, fail-closed HTTP branch, ListenAndServe + mux |
| `cmd/pagasa-pp-mcp/http_security.go` | loopback check, token resolve, bearer middleware |
| `cmd/pagasa-pp-mcp/http_security_test.go` | unit tests for bind/token/auth |

## On reprint / regen

- Re-apply HTTP fail-closed defaults and bearer middleware if the generator
  resets `cmd/*-pp-mcp/main.go` to bare `NewStreamableHTTPServer` + `Start`.
- Prefer upstream Printing Press template: loopback default + required token
  + `--allow-remote` so all printed CLIs inherit the same contract.
