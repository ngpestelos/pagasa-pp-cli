// Copyright 2026 Nestor G Pestelos Jr and contributors. Licensed under Apache-2.0. See LICENSE.
// cli-printing-press: novel-scaffold-test
// Novel command scaffold tests. Keep the wiring smoke test and add behavior cases as needed.

package cli

import (
	"bytes"
	"strings"
	"testing"
)

// TestNovelDigestHelpWires smoke-tests that the digest command
// resolves at runtime and renders useful --help output. Catches wiring
// regressions (missing AddCommand, panicking RunE on --help, etc.) before
// review. Keep this smoke test when adding behavior-specific cases.
func TestNovelDigestHelpWires(t *testing.T) {
	cmd := RootCmd()
	cmd.SetArgs([]string{"digest", "--help"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("digest --help error = %v (novel command not wired correctly?)", err)
	}
	help := out.String()
	for _, want := range []string{"Usage:", "digest"} {
		if !strings.Contains(help, want) {
			t.Fatalf("digest --help missing %q in output:\n%s", want, help)
		}
	}
}

// TestNowAndDigestNotMCPReadOnly guards issue #23: both commands call
// saveSnapshot (local SQLite) after live fetch, so they must not advertise
// mcp:read-only → readOnlyHint=true. Parallel to TestObsNotMCPReadOnly_HistoryIs.
// Issue #24: they must carry mcp:open-world so the walker sets openWorldHint.
func TestNowAndDigestNotMCPReadOnly(t *testing.T) {
	root := newRootCmd(&rootFlags{})
	for _, name := range []string{"now", "digest"} {
		cmd, _, err := root.Find([]string{name})
		if err != nil || cmd == nil {
			t.Fatalf("%s: %v", name, err)
		}
		if cmd.Annotations["mcp:read-only"] == "true" {
			t.Errorf("%s must not be mcp:read-only (saveSnapshot mutates local store)", name)
		}
		// local-write would set openWorldHint=false while these still scrape PAGASA.
		if cmd.Annotations["mcp:local-write"] == "true" {
			t.Errorf("%s must not be mcp:local-write (open-world HTTP; local-write forces openWorld=false)", name)
		}
		if cmd.Annotations["mcp:open-world"] != "true" {
			t.Errorf("%s must be mcp:open-world (outbound PAGASA HTTP, issue #24)", name)
		}
	}
}

// TestLiveScrapersMCPOpenWorld pins #24 annotations on pure network RO tools
// (walker tests cover hint mapping; this covers real RootCmd wiring).
func TestLiveScrapersMCPOpenWorld(t *testing.T) {
	root := newRootCmd(&rootFlags{})
	for _, name := range []string{"storm", "forecast", "watch", "approach"} {
		cmd, _, err := root.Find([]string{name})
		if err != nil || cmd == nil {
			t.Fatalf("%s: %v", name, err)
		}
		if cmd.Annotations["mcp:read-only"] != "true" {
			t.Errorf("%s: want mcp:read-only", name)
		}
		if cmd.Annotations["mcp:open-world"] != "true" {
			t.Errorf("%s: want mcp:open-world (issue #24)", name)
		}
	}
}
