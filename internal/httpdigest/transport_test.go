package httpdigest

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	testUsername = "admin"
	testRealm    = "test-realm"
	testOpaque   = "test-opaque"
)

// TestDigestAuthTransport tests the digest authentication transport.
func TestDigestAuthTransport(t *testing.T) {
	nonce := "test-nonce"
	realm := testRealm
	opaque := testOpaque

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Digest ") {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(
				`Digest realm=%q, nonce=%q, opaque=%q, qop="auth"`,
				realm, nonce, opaque))
			w.WriteHeader(http.StatusUnauthorized)

			return
		}
		// Verify digest auth header contains required fields
		if !strings.Contains(authHeader, `username="`+testUsername+`"`) {
			t.Error("Digest auth header missing username")
		}
		if !strings.Contains(authHeader, `realm="`+realm+`"`) {
			t.Error("Digest auth header missing realm")
		}
		if !strings.Contains(authHeader, `nonce="`+nonce+`"`) {
			t.Error("Digest auth header missing nonce")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer server.Close()

	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
	}

	digestClient := &http.Client{
		Transport: &Transport{
			Transport: tr,
			Username:  testUsername,
			Password:  "password",
		},
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, http.NoBody)
	if err != nil {
		t.Fatalf("NewRequest() failed: %v", err)
	}

	resp, err := digestClient.Do(req)
	if err != nil {
		t.Fatalf("Do() failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}

// TestExtractParam tests the extractParam helper function.
// TestExtractParam tests the extractParam helper function.
func TestExtractParam(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		param      string
		expected   string
	}{
		{
			name:       "extract realm",
			authHeader: `Digest realm="` + testRealm + `", nonce="123"`,
			param:      "realm",
			expected:   testRealm,
		},
		{
			name:       "extract nonce",
			authHeader: `Digest realm="test", nonce="abc123"`,
			param:      "nonce",
			expected:   "abc123",
		},
		{
			name:       "extract qop",
			authHeader: `Digest realm="test", qop="auth"`,
			param:      "qop",
			expected:   "auth",
		},
		{
			name:       "missing param",
			authHeader: `Digest realm="test"`,
			param:      "nonce",
			expected:   "",
		},
		{
			name:       "empty header",
			authHeader: "",
			param:      "realm",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractParam(tt.authHeader, tt.param)
			if result != tt.expected {
				t.Errorf("extractParam() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestGenerateNonce tests nonce generation.
// TestGenerateNonce tests nonce generation.
func TestGenerateNonce(t *testing.T) {
	// Generate multiple nonces and verify they're different and valid hex
	nonces := make(map[string]bool)
	for range 10 {
		nonce := generateNonce()
		if len(nonce) != nonceSize*2 { // hex encoding doubles the length
			t.Errorf("generateNonce() length = %d, want %d", len(nonce), nonceSize*2)
		}
		// Verify it's valid hex
		_, err := hex.DecodeString(nonce)
		if err != nil {
			t.Errorf("generateNonce() returned invalid hex: %v", err)
		}
		nonces[nonce] = true
	}

	// Verify nonces are unique (very unlikely to collide with crypto/rand)
	if len(nonces) < 10 {
		t.Error("generateNonce() generated duplicate nonces")
	}
}

// TestMd5Hash tests MD5 hash function.
// This verifies that the nc field is properly protected from race conditions.
func TestDigestAuthTransportConcurrency(t *testing.T) {
	nonce := "test-nonce"
	realm := testRealm
	opaque := testOpaque

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Digest ") {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(
				`Digest realm=%q, nonce=%q, opaque=%q, qop="auth"`,
				realm, nonce, opaque))
			w.WriteHeader(http.StatusUnauthorized)

			return
		}
		// Verify nc (nonce count) is present and valid
		if !strings.Contains(authHeader, "nc=") {
			t.Error("Digest auth header missing nc (nonce count)")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer server.Close()

	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
	}

	// Create a single transport instance that will be used concurrently
	digestTransport := &Transport{
		Transport: tr,
		Username:  testUsername,
		Password:  "password",
	}

	digestClient := &http.Client{
		Transport: digestTransport,
		Timeout:   30 * time.Second,
	}

	// Make concurrent requests to verify no race conditions
	const numRequests = 10
	done := make(chan bool, numRequests)
	errCh := make(chan error, numRequests)

	for i := range numRequests {
		go func(id int) {
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, http.NoBody)
			if err != nil {
				errCh <- fmt.Errorf("request %d: %w", id, errors.New("request creation failed"))
				done <- true

				return
			}

			resp, err := digestClient.Do(req)
			if err != nil {
				errCh <- fmt.Errorf("request %d: %w", id, errors.New("request execution failed"))
				done <- true

				return
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				errCh <- fmt.Errorf("request %d: expected 200, got %d", id, resp.StatusCode)
			}
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for range numRequests {
		<-done
	}

	// Check for errors
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Error(err)
		}
	}

	// Verify that nc was incremented correctly (should be at least numRequests)
	// Note: Each request triggers 2 RoundTrip calls (initial + retry with auth),
	// so nc should be at least numRequests
	digestTransport.ncMu.Lock()
	finalNC := digestTransport.nc
	digestTransport.ncMu.Unlock()

	if finalNC < numRequests {
		t.Errorf("Expected nc >= %d, got %d", numRequests, finalNC)
	}
}

func TestMd5Hash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // Expected MD5 hash in hex
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name:     "simple string",
			input:    "test",
			expected: "098f6bcd4621d373cade4e832627b4f6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := md5Hash(tt.input)
			if result != tt.expected {
				t.Errorf("md5Hash(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestErrorTypes tests error type checking.
