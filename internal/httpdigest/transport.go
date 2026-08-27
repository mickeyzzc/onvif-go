// Package httpdigest implements HTTP Digest authentication (RFC 7616,
// the subset cameras use) as an http.RoundTripper, shared by the
// library's file-download path.
package httpdigest

import (
	"crypto/md5" //nolint:gosec // MD5 required for ONVIF digest auth
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// nonceSize matches the WS-Security nonce length used across the library.
const nonceSize = 16

// Transport implements digest authentication over a base transport.
type Transport struct {
	Transport *http.Transport
	Username  string
	Password  string
	nc        int
	ncMu      sync.Mutex // Protects nc field from concurrent access
}

// RoundTrip implements http.RoundTripper with digest auth support.
func (d *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// First request without auth to get the challenge
	resp, err := d.Transport.RoundTrip(req)
	if err != nil {
		return resp, fmt.Errorf("transport round trip failed: %w", err)
	}

	// If we get 401, handle digest auth challenge
	if resp.StatusCode == http.StatusUnauthorized {
		// Read the WWW-Authenticate header
		authHeader := resp.Header.Get("WWW-Authenticate")
		if strings.Contains(authHeader, "Digest") {
			// Parse digest challenge and create auth header
			authHeaderValue := d.createDigestAuthHeader(req, authHeader)

			// Create new request with auth header
			newReq := req.Clone(req.Context())
			newReq.Header.Set("Authorization", authHeaderValue)

			// Retry with auth
			resp, err = d.Transport.RoundTrip(newReq)
			if err != nil {
				return resp, fmt.Errorf("transport round trip with auth failed: %w", err)
			}

			return resp, nil
		}
	}

	return resp, nil
}

// createDigestAuthHeader creates a digest auth header from the challenge.
func (d *Transport) createDigestAuthHeader(req *http.Request, authHeader string) string {
	// Simple digest auth implementation - parse challenge and create response
	// This is a basic implementation that handles most ONVIF cameras

	// Extract digest parameters from WWW-Authenticate header
	realm := extractParam(authHeader, "realm")
	nonce := extractParam(authHeader, "nonce")
	qop := extractParam(authHeader, "qop")
	uri := req.URL.Path
	if req.URL.RawQuery != "" {
		uri += "?" + req.URL.RawQuery
	}

	// Generate response hash
	ha1 := md5Hash(d.Username + ":" + realm + ":" + d.Password)

	method := req.Method
	ha2 := md5Hash(method + ":" + uri)

	// Increment nonce count atomically to prevent race conditions
	// HTTP transports must be safe for concurrent use
	d.ncMu.Lock()
	d.nc++
	nc := d.nc
	d.ncMu.Unlock()
	ncStr := fmt.Sprintf("%08x", nc)
	cnonce := generateNonce()

	var responseStr string
	if qop == "auth" {
		responseStr = md5Hash(ha1 + ":" + nonce + ":" + ncStr + ":" + cnonce + ":auth:" + ha2)
	} else {
		responseStr = md5Hash(ha1 + ":" + nonce + ":" + ha2)
	}

	// Build Authorization header
	authHeaderValue := fmt.Sprintf(`Digest username=%q, realm=%q, nonce=%q, uri=%q, response=%q`,
		d.Username, realm, nonce, uri, responseStr)

	if qop == "auth" {
		authHeaderValue += fmt.Sprintf(`, opaque=%q, qop=%s, nc=%s, cnonce=%q`,
			extractParam(authHeader, "opaque"), qop, ncStr, cnonce)
	}

	return authHeaderValue
}

// Helper functions for digest auth.
func extractParam(authHeader, param string) string {
	prefix := param + `="`
	idx := strings.Index(authHeader, prefix)
	if idx == -1 {
		return ""
	}
	start := idx + len(prefix)
	end := strings.Index(authHeader[start:], `"`)
	if end == -1 {
		return ""
	}

	return authHeader[start : start+end]
}

func md5Hash(s string) string {
	h := md5.New() //nolint:gosec // MD5 required for ONVIF digest auth
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

// generateNonce generates a cryptographically secure random nonce for digest authentication.
func generateNonce() string {
	bytes := make([]byte, nonceSize)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based nonce if crypto/rand fails (shouldn't happen)
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	return hex.EncodeToString(bytes)
}
