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

// TestDigestSHA256Challenge (RFC 7616): a server demanding algorithm=
// SHA-256 must be answered with SHA-256 hashes and the algorithm echoed;
// Hikvision SHA-256-capable firmware rejects the MD5 response outright
// (found during the #1120 zero-touch E2E — the device offers both an MD5
// and a SHA-256 challenge and only honors the SHA-256 answer).
func TestDigestSHA256Challenge(t *testing.T) {
	const nonce = "sha256-nonce"
	const realm = testRealm

	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			// Dual challenge, exactly the live shape: MD5 line + SHA-256 line.
			w.Header().Add("WWW-Authenticate",
				`Digest qop="auth", realm="`+realm+`", nonce="md5-nonce", stale="FALSE"`)
			w.Header().Add("WWW-Authenticate",
				`Digest qop="auth", realm="`+realm+`", nonce="`+nonce+`", algorithm="SHA-256", stale="FALSE"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	tr := &Transport{
		Transport: server.Client().Transport.(*http.Transport),
		Username:  testUsername,
		Password:  "secret",
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Post(server.URL+"/onvif/device_service", "application/soap+xml", strings.NewReader("<body/>"))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 after digest auth", resp.StatusCode)
	}

	// The answer must target the SHA-256 challenge (its nonce), echo the
	// algorithm, and carry a SHA-256 response hash.
	if !strings.Contains(gotAuth, `nonce="sha256-nonce"`) {
		t.Errorf("authorization did not answer the SHA-256 challenge: %s", gotAuth)
	}
	if !strings.Contains(gotAuth, `algorithm=SHA-256`) {
		t.Errorf("algorithm not echoed: %s", gotAuth)
	}

	// Known-answer: HA1 = SHA256(user:realm:pass); response uses the qop=auth
	// chain. Recompute with the cnonce/nc the transport generated by
	// re-deriving from the header fields.
	nc := extractParamBare(gotAuth, "nc")
	cnonce := extractParamBare(gotAuth, "cnonce")
	response := extractParamBare(gotAuth, "response")
	uri := extractParamBare(gotAuth, "uri")
	ha1 := hashHex("SHA-256", testUsername+":"+realm+":secret")
	ha2 := hashHex("SHA-256", "POST:"+uri)
	want := hashHex("SHA-256", ha1+":"+nonce+":"+nc+":"+cnonce+":auth:"+ha2)
	if response != want {
		t.Errorf("response hash mismatch:\n got %s\nwant %s", response, want)
	}
}

// TestDigestMD5ChallengeUnchanged: a plain (no algorithm) challenge keeps
// the legacy MD5 behavior — the response header carries no algorithm echo.
func TestDigestMD5ChallengeUnchanged(t *testing.T) {
	const nonce = "plain-nonce"
	const realm = testRealm

	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.Header().Set("WWW-Authenticate",
				`Digest qop="auth", realm="`+realm+`", nonce="`+nonce+`", opaque="`+testOpaque+`"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	tr := &Transport{
		Transport: server.Client().Transport.(*http.Transport),
		Username:  testUsername,
		Password:  "secret",
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(server.URL + "/x")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if strings.Contains(gotAuth, "algorithm=") {
		t.Errorf("plain challenge must not produce an algorithm echo: %s", gotAuth)
	}

	response := extractParamBare(gotAuth, "response")
	nc := extractParamBare(gotAuth, "nc")
	cnonce := extractParamBare(gotAuth, "cnonce")
	uri := extractParamBare(gotAuth, "uri")
	ha1 := md5Hash(testUsername + ":" + realm + ":secret")
	ha2 := md5Hash("GET:" + uri)
	want := md5Hash(ha1 + ":" + nonce + ":" + nc + ":" + cnonce + ":auth:" + ha2)
	if response != want {
		t.Errorf("MD5 response hash mismatch:\n got %s\nwant %s", response, want)
	}
}

// TestPickChallenge: the SHA-256 line wins when several are offered.
func TestPickChallenge(t *testing.T) {
	md5Line := `Digest qop="auth", realm="r", nonce="m"`
	shaLine := `Digest qop="auth", realm="r", nonce="s", algorithm="SHA-256"`
	if got := pickChallenge([]string{md5Line, shaLine}); got != shaLine {
		t.Errorf("pickChallenge chose the MD5 line; want SHA-256")
	}
	if got := pickChallenge([]string{md5Line}); got != md5Line {
		t.Errorf("pickChallenge lost the only line")
	}
}
