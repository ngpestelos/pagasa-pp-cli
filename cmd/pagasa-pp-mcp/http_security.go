// Copyright 2026 Nestor G Pestelos Jr and contributors. Licensed under Apache-2.0. See LICENSE.
// Security hardening for streamable HTTP MCP transport (issue #26).

package main

import (
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"
)

// Env / defaults for HTTP MCP transport. Token uses PP_MCP_* so fleet
// supervisors share one name with PP_MCP_TRANSPORT.
const (
	envMCPToken     = "PP_MCP_TOKEN"
	defaultHTTPAddr = "127.0.0.1:7777"
	mcpHTTPPath     = "/mcp"
	authHeaderName  = "Authorization"
	wwwAuthenticate = `Bearer realm="pagasa-pp-mcp"`
)

// resolveHTTPToken returns flag value if non-empty, else PP_MCP_TOKEN env.
func resolveHTTPToken(flagToken string) string {
	if strings.TrimSpace(flagToken) != "" {
		return flagToken
	}
	return os.Getenv(envMCPToken)
}

// isLoopbackBind reports whether addr is a loopback listen address.
// Accepts host:port, :port (all interfaces — not loopback), bare host, or
// IPv6 bracket forms. Empty host in host:port means 0.0.0.0 (not loopback).
func isLoopbackBind(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// Bare host or malformed — try as host only.
		host = strings.Trim(addr, "[]")
		if host == "" {
			return false
		}
	}
	if host == "" {
		// ":7777" → all interfaces
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	return ip.IsLoopback()
}

// validateHTTPListen enforces loopback-only bind unless allowRemote is set.
func validateHTTPListen(addr string, allowRemote bool) error {
	if strings.TrimSpace(addr) == "" {
		return fmt.Errorf("http --addr must not be empty")
	}
	if isLoopbackBind(addr) {
		return nil
	}
	if !allowRemote {
		return fmt.Errorf("http --addr %q is not loopback; pass --allow-remote to bind on all interfaces or a non-loopback host (and still require --token / %s)", addr, envMCPToken)
	}
	return nil
}

// requireHTTPToken fails closed when HTTP transport has no shared secret.
func requireHTTPToken(token string) error {
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("http transport requires a non-empty bearer token via --token or %s (stdio needs no token)", envMCPToken)
	}
	return nil
}

// constantTimeTokenEqual compares bearer secrets without early length leak
// beyond unequal-length false (Go subtle requires equal lengths).
func constantTimeTokenEqual(got, want string) bool {
	if len(got) != len(want) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

// extractBearerToken parses Authorization: Bearer <token>. Empty if missing/malformed.
// Prefix match is case-insensitive ("Bearer", "bearer", …).
func extractBearerToken(h http.Header) string {
	raw := h.Get(authHeaderName)
	if raw == "" {
		return ""
	}
	const prefix = "bearer "
	if len(raw) < len(prefix) || !strings.EqualFold(raw[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(raw[len(prefix):])
}

// bearerAuthMiddleware requires Authorization: Bearer <token> on every request.
// On failure returns 401 without invoking next (no tool side effects).
func bearerAuthMiddleware(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := extractBearerToken(r.Header)
		if !constantTimeTokenEqual(got, token) {
			w.Header().Set("WWW-Authenticate", wwwAuthenticate)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// newHTTPMCPHandler mounts the streamable MCP handler at /mcp behind bearer auth.
func newHTTPMCPHandler(token string, mcp http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(mcpHTTPPath, bearerAuthMiddleware(token, mcp))
	// Explicit 401 for other paths would leak less about routing; 404 is fine.
	return mux
}
