// Copyright 2026 Nestor G Pestelos Jr and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import "testing"

func TestNormalizeVersion(t *testing.T) {
	cases := map[string]string{
		"":        "",
		"0.1.2":   "0.1.2",
		"v0.1.2":  "0.1.2",
		" v1.0 ":  "1.0",
		"(devel)": "(devel)",
	}
	for in, want := range cases {
		if got := normalizeVersion(in); got != want {
			t.Fatalf("normalizeVersion(%q) = %q; want %q", in, got, want)
		}
	}
}

func TestResolveVersionPrefersLdflags(t *testing.T) {
	if got := resolveVersion("0.1.2"); got != "0.1.2" {
		t.Fatalf("ldflags version: got %q", got)
	}
	if got := resolveVersion("v0.1.2"); got != "0.1.2" {
		t.Fatalf("ldflags with v: got %q", got)
	}
}
