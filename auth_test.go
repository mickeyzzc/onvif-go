package onvif

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// authFlavor describes how a test server decides which authentication a
// request carried, mirroring the real-device compatibility matrix from
// issue #1.
type authFlavor string

const (
	flavorDigestOnly authFlavor = "digest-only"
	flavorTextOnly   authFlavor = "text-only"
	flavorBasicOnly  authFlavor = "basic-only"
	flavorNoneOnly   authFlavor = "none-only"
	flavorRejectAll  authFlavor = "reject-all"
)

// classifiedAuth records what one incoming request looked like.
type classifiedAuth struct {
	hasDigest bool
	hasText   bool
	hasBasic  bool
	hasNone   bool
}

// classify inspects a SOAP request and reports which auth shapes it carried.
func classify(r *http.Request) classifiedAuth {
	var c classifiedAuth

	body, _ := io.ReadAll(r.Body)
	bodyStr := string(body)

	c.hasDigest = strings.Contains(bodyStr, "#PasswordDigest")
	c.hasText = strings.Contains(bodyStr, "#PasswordText") &&
		!strings.Contains(bodyStr, "#PasswordDigest")
	if user, pass, ok := r.BasicAuth(); ok && user != "" && pass != "" {
		c.hasBasic = true
	}

	c.hasNone = !c.hasDigest && !c.hasText && !c.hasBasic

	return c
}

// deviceInfoBody is a minimal GetDeviceInformationResponse envelope.
const deviceInfoBody = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <tds:GetDeviceInformationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
      <tds:Manufacturer>TestCam</tds:Manufacturer>
      <tds:Model>T-1000</tds:Model>
      <tds:FirmwareVersion>1.0</tds:FirmwareVersion>
      <tds:SerialNumber>SER-42</tds:SerialNumber>
    </tds:GetDeviceInformationResponse>
  </s:Body>
</s:Envelope>`

const notAuthorizedFaultBody = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <s:Fault>
      <s:Code>
        <s:Value>s:Sender</s:Value>
        <s:Subcode>
          <s:Value>ter:NotAuthorized</s:Value>
        </s:Subcode>
      </s:Code>
      <s:Reason><s:Text xml:lang="en">Sender not authorized</s:Text></s:Reason>
    </s:Fault>
  </s:Body>
</s:Envelope>`

// newAuthTestServer starts a device that accepts exactly one auth flavor and
// counts the classified attempts it saw.
func newAuthTestServer(t *testing.T, flavor authFlavor) (*httptest.Server, *authAttemptLog) {
	t.Helper()

	attempts := &authAttemptLog{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := classify(r)
		attempts.add(c)

		accepted := false
		switch flavor {
		case flavorDigestOnly:
			accepted = c.hasDigest
		case flavorTextOnly:
			accepted = c.hasText
		case flavorBasicOnly:
			accepted = c.hasBasic
		case flavorNoneOnly:
			accepted = c.hasNone
		case flavorRejectAll:
			accepted = false
		}

		if !accepted {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(notAuthorizedFaultBody))

			return
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(deviceInfoBody))
	}))
	t.Cleanup(server.Close)

	return server, attempts
}

// authAttemptLog is a concurrency-safe log of classified requests.
type authAttemptLog struct {
	mu   sync.Mutex
	logs []classifiedAuth
}

func (l *authAttemptLog) add(c classifiedAuth) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, c)
}

func (l *authAttemptLog) snapshot() []classifiedAuth {
	l.mu.Lock()
	defer l.mu.Unlock()

	return append([]classifiedAuth(nil), l.logs...)
}

// newAuthTestClient builds a client pointed at the test server.
func newAuthTestClient(t *testing.T, server *httptest.Server, opts ...ClientOption) *Client {
	t.Helper()

	allOpts := append([]ClientOption{
		WithCredentials("admin", "secret"),
		WithHTTPClient(server.Client()),
	}, opts...)

	client, err := NewClient(server.URL, allOpts...)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	return client
}

func TestAuthDefaultBehaviorIsDigest(t *testing.T) {
	server, attempts := newAuthTestServer(t, flavorDigestOnly)
	client := newAuthTestClient(t, server)

	info, err := client.Device().GetDeviceInformation(context.Background())
	if err != nil {
		t.Fatalf("GetDeviceInformation() error = %v", err)
	}

	if info.Manufacturer != "TestCam" {
		t.Errorf("Manufacturer = %q, want %q", info.Manufacturer, "TestCam")
	}

	for i, a := range attempts.snapshot() {
		if !a.hasDigest {
			t.Errorf("attempt %d was not digest-authenticated: %+v", i, a)
		}
	}
}

func TestAuthModePasswordText(t *testing.T) {
	server, _ := newAuthTestServer(t, flavorTextOnly)
	client := newAuthTestClient(t, server, WithAuthMode(AuthPasswordText))

	if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
		t.Fatalf("GetDeviceInformation() error = %v", err)
	}
}

func TestAuthModeHTTPBasic(t *testing.T) {
	server, _ := newAuthTestServer(t, flavorBasicOnly)
	client := newAuthTestClient(t, server, WithAuthMode(AuthHTTPBasic))

	if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
		t.Fatalf("GetDeviceInformation() error = %v", err)
	}
}

func TestAuthModeNoneSendsNoAuthHeaders(t *testing.T) {
	server, _ := newAuthTestServer(t, flavorNoneOnly)
	// Credentials remain configured; AuthNone must suppress them anyway
	// (ESP32-style devices reject every auth-bearing request).
	client := newAuthTestClient(t, server, WithAuthMode(AuthNone))

	if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
		t.Fatalf("GetDeviceInformation() error = %v", err)
	}
}

func TestAuthFallbackLadderAndSticky(t *testing.T) {
	// Device accepts only HTTP Basic; the ladder is digest -> basic.
	server, attempts := newAuthTestServer(t, flavorBasicOnly)
	client := newAuthTestClient(t, server,
		WithAuthFallback(AuthHTTPBasic, AuthNone))

	// First call: digest rejected, basic accepted.
	if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
		t.Fatalf("GetDeviceInformation() (ladder) error = %v", err)
	}

	first := attempts.snapshot()
	if len(first) != 2 { //nolint:mnd // digest attempt + basic attempt
		t.Fatalf("first call made %d attempts, want 2: %+v", len(first), first)
	}
	if !first[0].hasDigest || !first[1].hasBasic {
		t.Errorf("ladder order wrong: %+v", first)
	}

	// Sticky: the second call must go straight to basic.
	if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
		t.Fatalf("GetDeviceInformation() (sticky) error = %v", err)
	}

	if got := client.AuthLadderMode(); got != AuthHTTPBasic {
		t.Errorf("AuthLadderMode() = %q, want %q", got, AuthHTTPBasic)
	}

	second := attempts.snapshot()[len(first):]
	if len(second) != 1 || !second[0].hasBasic {
		t.Errorf("sticky call made %+v, want a single basic attempt", second)
	}
}

func TestAuthFallbackExhaustedIsErrUnauthorized(t *testing.T) {
	server, _ := newAuthTestServer(t, flavorRejectAll)
	client := newAuthTestClient(t, server,
		WithAuthFallback(AuthPasswordText, AuthHTTPBasic))

	_, err := client.Device().GetDeviceInformation(context.Background())
	if err == nil {
		t.Fatal("GetDeviceInformation() succeeded, want auth failure")
	}

	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("errors.Is(err, ErrUnauthorized) = false, err = %v", err)
	}
}

func TestErrUnauthorizedOnPlainHTTP401(t *testing.T) {
	server, _ := newAuthTestServer(t, flavorRejectAll)
	client := newAuthTestClient(t, server) // no fallback configured

	_, err := client.Device().GetDeviceInformation(context.Background())
	if err == nil {
		t.Fatal("GetDeviceInformation() succeeded, want auth failure")
	}

	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("errors.Is(err, ErrUnauthorized) = false, err = %v", err)
	}
}

func TestErrUnauthorizedOnFaultWithHTTP200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(notAuthorizedFaultBody)) // HTTP 200 + fault
	}))
	t.Cleanup(server.Close)

	client := newAuthTestClient(t, server)

	_, err := client.Device().GetDeviceInformation(context.Background())
	if err == nil {
		t.Fatal("GetDeviceInformation() succeeded, want fault error")
	}

	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("errors.Is(err, ErrUnauthorized) = false, err = %v", err)
	}
}

func TestVoidOperationSurfacesFaultWithHTTP200(t *testing.T) {
	// Regression: void operations (response struct without payload fields)
	// previously reported success for 200-with-Fault responses.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(notAuthorizedFaultBody))
	}))
	t.Cleanup(server.Close)

	client := newAuthTestClient(t, server, WithAuthMode(AuthNone))

	err := client.Events().Unsubscribe(context.Background(), server.URL+"/subscription")
	if err == nil {
		t.Fatal("Unsubscribe() succeeded on a fault response, want error")
	}

	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("errors.Is(err, ErrUnauthorized) = false, err = %v", err)
	}
}

func TestSetCredentialsResetsStickyAuth(t *testing.T) {
	server, attempts := newAuthTestServer(t, flavorBasicOnly)
	client := newAuthTestClient(t, server, WithAuthFallback(AuthHTTPBasic))

	if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
		t.Fatalf("GetDeviceInformation() error = %v", err)
	}

	if got := client.AuthLadderMode(); got != AuthHTTPBasic {
		t.Fatalf("AuthLadderMode() = %q before reset, want %q", got, AuthHTTPBasic)
	}

	client.SetCredentials("admin", "newsecret")
	if got := client.AuthLadderMode(); got != AuthDigest {
		t.Errorf("AuthLadderMode() = %q after SetCredentials, want primary %q", got, AuthDigest)
	}

	// Ladder must run again for the new credentials.
	if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
		t.Fatalf("GetDeviceInformation() error = %v", err)
	}

	logs := attempts.snapshot()
	if len(logs) < 3 { //nolint:mnd // digest+basic, then digest+basic again
		t.Errorf("expected ladder to re-run after credential change, attempts = %+v", logs)
	}
}

func TestNewClientRejectsUnknownAuthMode(t *testing.T) {
	_, err := NewClient("http://192.168.1.100", WithAuthMode("bogus"))
	if err == nil {
		t.Fatal("NewClient() accepted unknown auth mode, want error")
	}

	_, err = NewClient("http://192.168.1.100", WithAuthFallback("bogus"))
	if err == nil {
		t.Fatal("NewClient() accepted unknown fallback mode, want error")
	}
}

func TestNonAuthErrorsAreNotRetried(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	t.Cleanup(server.Close)

	client := newAuthTestClient(t, server, WithAuthFallback(AuthHTTPBasic))

	_, err := client.Device().GetDeviceInformation(context.Background())
	if err == nil {
		t.Fatal("GetDeviceInformation() succeeded, want error")
	}

	if calls != 1 {
		t.Errorf("server called %d times, want 1 (non-auth errors must not re-run the ladder)", calls)
	}

	if errors.Is(err, ErrUnauthorized) {
		t.Errorf("500 must not classify as ErrUnauthorized, err = %v", err)
	}
}
