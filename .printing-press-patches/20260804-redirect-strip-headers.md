# Patch: strip secrets on cross-host redirects (20260804)

## Intent

Issue **#22** / #19 F2: `CheckRedirect` only `Del("Authorization")` on cross-host hops. Comment claimed custom API-key headers must also be stripped — they were not. `Config.Headers` could follow an open redirect / partner handoff.

## Changes

| Path | Change |
|------|--------|
| `internal/client/client.go` | Cross-host: Del Authorization, Cookie, Proxy-Authorization + every `Config.Headers` key |
| `internal/client/client_test.go` | Cross-host strip regression + same-host keep auth/custom headers |

## On reprint

- Keep cross-host strip of `Config.Headers` keys (not only standard auth headers).
- Keep both unit tests if the generator rewrites CheckRedirect.
