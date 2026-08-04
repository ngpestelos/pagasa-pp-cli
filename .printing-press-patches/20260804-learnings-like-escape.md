# Patch: ListLearnings LIKE ESCAPE + escapeLike (20260804)

## Intent

Issue **#29** / security review #19 F9: `ListLearnings` must not treat `%` / `_` in the filter needle as SQL LIKE wildcards. Bound params (not SQLi); residual risk is over-broad local matches. Match `ListAWSObs` (`escapeLike` + `ESCAPE '\'`).

## Files

| Path | Change |
|------|--------|
| `internal/store/learnings.go` | `learningsLikeFilter` → `LIKE ? ESCAPE '\'` + `escapeLike(needle)`; `ListLearnings` uses it after `NormalizeQuery` |
| `internal/store/learnings_like_test.go` | `TestEscapeLike`, filter shape, metachar seed rows, production substring path |
| `CHANGELOG.md` | Unreleased Security bullet for #29 |

## On reprint / regen

- If generator restores bare `query_pattern LIKE ?` with `"%"+NormalizeQuery(...)+"%"`, re-apply `learningsLikeFilter` (or inline ESCAPE + `escapeLike`).
- Prefer upstream store template: substring LIKE filters always ESCAPE + escape metacharacters.
- `escapeLike` lives in `aws_obs.go` (same package); keep one helper.
