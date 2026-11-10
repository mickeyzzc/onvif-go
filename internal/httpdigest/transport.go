// Package httpdigest implements HTTP Digest authentication (RFC 7616,
// the subset cameras use) as an http.RoundTripper, shared by the
// library's file-download path.
package httpdigest

import (
	"crypto/md5" //nolint:gosec // MD5 required for ONVIF digest auth
	"crypto/rand"
	"crypto/sha256"
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
		// Read the WWW-Authenticate header(s). Devices may emit several
		// challenges at once (Hikvision SHA-256-capable firmware answers
		// with both an MD5 line and an algorithm="SHA-256" line); prefer
		// the strongest, RFC 7616 §5.2.1-style.
		authHeader := pickChallenge(resp.Header.Values("WWW-Authenticate"))
		if authHeader != "" {
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

// pickChallenge selects the strongest digest challenge: an
// algorithm=SHA-256 one when offered (RFC 7616), otherwise the first
// plain Digest line.
func pickChallenge(challenges []string) string {
	var firstDigest string
	for _, ch := range challenges {
		if !strings.Contains(ch, "Digest") && !strings.Contains(ch, "digest") {
			continue
		}
		if firstDigest == "" {
			firstDigest = ch
		}
		if strings.Contains(ch, "SHA-256") || strings.Contains(ch, "sha-256") {
			return ch
		}
	}
	return firstDigest
}

// createDigestAuthHeader creates a digest auth header from the challenge.
// The hash algorithm follows the challenge's algorithm parameter: MD5 by
// default, SHA-256 when the device demands it (RFC 7616; Hikvision
// SHA-256-capable firmware rejects the MD5 response outright).
func (d *Transport) createDigestAuthHeader(req *http.Request, authHeader string) string {
	// Extract digest parameters from WWW-Authenticate header
	realm := extractParam(authHeader, "realm")
	nonce := extractParam(authHeader, "nonce")
	qop := extractParam(authHeader, "qop")
	algorithm := extractParamBare(authHeader, "algorithm")
	uri := req.URL.Path
	if req.URL.RawQuery != "" {
		uri += "?" + req.URL.RawQuery
	}

	// Generate response hash
	ha1 := hashHex(algorithm, d.Username+":"+realm+":"+d.Password)

	method := req.Method
	ha2 := hashHex(algorithm, method+":"+uri)

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
		responseStr = hashHex(algorithm, ha1+":"+nonce+":"+ncStr+":"+cnonce+":auth:"+ha2)
	} else {
		responseStr = hashHex(algorithm, ha1+":"+nonce+":"+ha2)
	}

	// Build Authorization header
	authHeaderValue := fmt.Sprintf(`Digest username=%q, realm=%q, nonce=%q, uri=%q, response=%q`,
		d.Username, realm, nonce, uri, responseStr)

	// RFC 7616 §3.5: the client MUST echo the algorithm it used when the
	// challenge specified one (servers key their nonce on it).
	if algorithm != "" {
		authHeaderValue += ", algorithm=" + algorithm
	}

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

// extractParamBare reads an unquoted (or quoted) parameter value —
// algorithm=SHA-256 ships unquoted on several firmwares.
func extractParamBare(authHeader, param string) string {
	if v := extractParam(authHeader, param); v != "" {
		return v
	}
	idx := strings.Index(authHeader, param+"=")
	if idx == -1 {
		return ""
	}
	rest := authHeader[idx+len(param)+1:]
	end := strings.IndexAny(rest, ", \t")
	if end == -1 {
		end = len(rest)
	}
	return strings.Trim(strings.TrimSpace(rest[:end]), `"`)
}

// hashHex digests s with the challenge's algorithm: SHA-256 when asked,
// MD5 otherwise (the ONVIF default).
func hashHex(algorithm, s string) string {
	switch strings.ToUpper(strings.TrimSpace(algorithm)) {
	case "SHA-256", "SHA256":
		h := sha256.New()
		h.Write([]byte(s))
		return hex.EncodeToString(h.Sum(nil))
	default:
		return md5Hash(s)
	}
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
