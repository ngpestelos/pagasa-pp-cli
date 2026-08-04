// Copyright 2026 Nestor G Pestelos Jr and contributors. Licensed under Apache-2.0. See LICENSE.

package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestIsLoopbackBind(t *testing.T) {
	t.Parallel()
	cases := []struct {
		addr string
		want bool
	}{
		{"127.0.0.1:7777", true},
		{"localhost:7777", true},
		{"LOCALHOST:9", true},
		{"[::1]:7777", true},
		{":7777", false},
		{"0.0.0.0:7777", false},
		{"192.168.1.10:7777", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isLoopbackBind(tc.addr); got != tc.want {
			t.Errorf("isLoopbackBind(%q)=%v want %v", tc.addr, got, tc.want)
		}
	}
}

func TestValidateHTTPListen(t *testing.T) {
	t.Parallel()
	if err := validateHTTPListen("127.0.0.1:7777", false); err != nil {
		t.Fatalf("loopback without allow-remote: %v", err)
	}
	if err := validateHTTPListen(":7777", false); err == nil {
		t.Fatal("expected error for all-interfaces without --allow-remote")
	}
	if err := validateHTTPListen(":7777", true); err != nil {
		t.Fatalf("all-interfaces with allow-remote: %v", err)
	}
	if err := validateHTTPListen("", false); err == nil {
		t.Fatal("expected empty addr error")
	}
}

func TestRequireHTTPToken(t *testing.T) {
	t.Parallel()
	if err := requireHTTPToken(""); err == nil {
		t.Fatal("empty token should fail")
	}
	if err := requireHTTPToken("   "); err == nil {
		t.Fatal("whitespace token should fail")
	}
	if err := requireHTTPToken("secret"); err != nil {
		t.Fatalf("non-empty: %v", err)
	}
}

func TestResolveHTTPToken_FlagWinsOverEnv(t *testing.T) {
	t.Setenv(envMCPToken, "from-env")
	if got := resolveHTTPToken("from-flag"); got != "from-flag" {
		t.Fatalf("got %q", got)
	}
	if got := resolveHTTPToken(""); got != "from-env" {
		t.Fatalf("env fallback got %q", got)
	}
	_ = os.Unsetenv(envMCPToken)
	if got := resolveHTTPToken(""); got != "" {
		t.Fatalf("empty got %q", got)
	}
}

func TestExtractBearerToken(t *testing.T) {
	t.Parallel()
	h := http.Header{}
	if extractBearerToken(h) != "" {
		t.Fatal("empty header")
	}
	h.Set("Authorization", "Bearer abc")
	if got := extractBearerToken(h); got != "abc" {
		t.Fatalf("got %q", got)
	}
	h.Set("Authorization", "bearer xyz")
	if got := extractBearerToken(h); got != "xyz" {
		t.Fatalf("case-insensitive got %q", got)
	}
	h.Set("Authorization", "Basic abc")
	if extractBearerToken(h) != "" {
		t.Fatal("non-bearer should be empty")
	}
}

func TestBearerAuthMiddleware(t *testing.T) {
	t.Parallel()
	const secret = "test-secret-token"
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	h := bearerAuthMiddleware(secret, inner)

	// missing auth
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing auth: status %d", rec.Code)
	}
	if rec.Header().Get("WWW-Authenticate") == "" {
		t.Fatal("expected WWW-Authenticate")
	}

	// wrong token
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token: status %d", rec.Code)
	}

	// correct token
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer "+secret)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("good token: status %d body %q", rec.Code, rec.Body.String())
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "ok" {
		t.Fatalf("body %q", body)
	}
}

func TestNewHTTPMCPHandler_PathAndAuth(t *testing.T) {
	t.Parallel()
	const secret = "path-secret"
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("mcp"))
	})
	h := newHTTPMCPHandler(secret, inner)

	// wrong path
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+secret)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("root path status %d", rec.Code)
	}

	// correct path + auth
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer "+secret)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "mcp" {
		t.Fatalf("status %d body %q", rec.Code, rec.Body.String())
	}
}
