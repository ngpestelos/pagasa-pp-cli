# Version resolve — 2026-07-27

## Problem

`var version = "0.0.0-dev"` only got overridden by goreleaser ldflags. Local
`go install` / `go build` printed `0.0.0-dev` forever.

## Fix

1. `internal/cli/version.go` — `resolveVersion`: ldflags → module `Main.Version`
   (pseudo-version) → VCS revision → `0.0.0-dev`. Exported `Version()`.
2. MCP uses `cli.Version()` (drop `main.version` ldflag).
3. Makefile injects `git describe` via ldflags on build/install.
4. Goreleaser MCP ldflags retarget to `internal/cli.version`.
