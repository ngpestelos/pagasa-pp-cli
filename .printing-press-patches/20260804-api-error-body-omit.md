# Patch: omit upstream HTTP bodies from APIError.Error() (20260804)

## Intent

Issue **#27** / security review #19 F7: failed HTTP responses must not embed
body snippets into `err.Error()` paths that flow to `--json` / `--agent` /
MCP tool errors (token noise + sensitive edge content).

## Files

| Path | Change |
|------|--------|
| `internal/client/client.go` | `APIError.Error()` status+class only; `httpStatusClass`; Body field retained for doctor |
| `internal/client/client_test.go` | assert Error() omits HTML body |
| `internal/cli/helpers.go` | `writeAPIErrorEnvelope` adds `status`, never body |
| `internal/cli/helpers_api_error_test.go` | machine envelope leak test |

## On reprint / regen

- Re-apply Error() body omission if the generator restores `: " + e.Body`.
- Prefer upstream Printing Press: never append response body to `APIError.Error()`.
