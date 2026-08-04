# Patch: OpenReadOnly file: URI path safety (20260804)

## Intent

Issue **#28** / #19 F8: `OpenReadOnly` built DSN as `"file:" + dbPath + "?mode=ro&..."`. Path characters significant to URI (`?`, `#`) could alter query parameters so `mode=ro` may not apply.

## Changes

| Path | Change |
|------|--------|
| `internal/store/store.go` | `readOnlySQLiteDSN` / `sqliteFileURI` via `url.URL`; reject `?` in path (modernc cannot open `%3F`); preflight probe shares helpers |
| tests | DSN `#` escape + reject `?`; OpenReadOnly integration for both |

## On reprint

- Keep URL-based DSN construction; do not reintroduce string concat for `file:` + path + query.
- Keep hard reject on `?` in db paths unless driver gains working percent-decode for `%3F`.
