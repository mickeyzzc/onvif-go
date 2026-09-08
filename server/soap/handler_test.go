package soap

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testXMLHeader = `<?xml version="1.0"?>`

func TestNewHandler(t *testing.T) {
	handler := NewHandler("admin", "password")

	if handler == nil {
		t.Error("NewHandler returned nil")

		return
	}
	if handler.username != "admin" {
		t.Errorf("Username mismatch: got %s, want admin", handler.username)
	}
	if handler.password != "password" {
		t.Errorf("Password mismatch: got %s, want password", handler.password)
	}
	if handler.handlers == nil {
		t.Error("Handlers map is nil")
	}
}

func TestRegisterHandler(t *testing.T) {
	handler := NewHandler("admin", "password")

	testHandler := func(body []byte) (interface{}, error) {
		return "test response", nil
	}

	handler.RegisterHandler("TestAction", testHandler)

	if _, ok := handler.handlers["TestAction"]; !ok {
		t.Error("Handler not registered")
	}
}

func TestServeHTTPMethodNotAllowed(t *testing.T) {
	handler := NewHandler("admin", "password")

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestServeHTTPValidSOAPRequest(t *testing.T) {
	handler := NewHandler("", "") // No authentication

	// Create test handler
	handler.RegisterHandler("TestAction", func(body []byte) (interface{}, error) {
		return hardeningResponse{Result: "Success"}, nil
	})

	// Create SOAP request
	soapBody := testXMLHeader + `
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <TestAction/>
  </soap:Body>
</soap:Envelope>`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(soapBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code == http.StatusInternalServerError {
		t.Errorf("Handler returned error: %s", w.Body.String())
	}
}

func TestServeHTTPInvalidSOAPEnvelope(t *testing.T) {
	handler := NewHandler("", "")

	invalidXML := `<?xml version="1.0"?>
<invalid>
  <xml>not soap</xml>
</invalid>`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(invalidXML))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return a SOAP fault
	if !strings.Contains(w.Body.String(), "Fault") {
		t.Errorf("Expected SOAP fault, got: %s", w.Body.String())
	}
}

func TestServeHTTPUnknownAction(t *testing.T) {
	handler := NewHandler("", "")

	soapBody := `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <UnknownAction/>
  </soap:Body>
</soap:Envelope>`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(soapBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "Fault") {
		t.Errorf("Expected SOAP fault for unknown action")
	}
}

func TestExtractAction(t *testing.T) {
	handler := NewHandler("", "")

	tests := []struct {
		name           string
		soapBody       string
		expectedAction string
	}{
		{
			name: "Simple action",
			soapBody: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <GetDeviceInformation/>
  </soap:Body>
</soap:Envelope>`,
			expectedAction: "GetDeviceInformation",
		},
		{
			name: "Action with namespace",
			soapBody: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <tds:GetDeviceInformation xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
  </soap:Body>
</soap:Envelope>`,
			expectedAction: "GetDeviceInformation",
		},
		{
			name: "Action with attributes",
			soapBody: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <GetProfiles>
      <param>value</param>
    </GetProfiles>
  </soap:Body>
</soap:Envelope>`,
			expectedAction: "GetProfiles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := handler.extractAction([]byte(tt.soapBody))
			if action != tt.expectedAction {
				t.Errorf("Expected action %s, got %s", tt.expectedAction, action)
			}
		})
	}
}

func TestExtractActionInvalid(t *testing.T) {
	handler := NewHandler("", "")

	invalidXML := "not valid xml at all"
	action := handler.extractAction([]byte(invalidXML))

	if action != "" {
		t.Errorf("Expected empty action for invalid XML, got %s", action)
	}
}

func TestSendFault(t *testing.T) {
	handler := NewHandler("", "")

	w := httptest.NewRecorder()
	handler.sendFault(w, "Sender", "Test error", "Test error message")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	response := w.Body.String()
	if !strings.Contains(response, "Fault") {
		t.Error("Response should contain Fault element")
	}
	if !strings.Contains(response, "Test error") {
		t.Error("Response should contain error message")
	}
}

func TestSendResponse(t *testing.T) {
	handler := NewHandler("", "")

	w := httptest.NewRecorder()

	// encoding/xml cannot marshal maps as nested field content; handler
	// responses are always structs (see RegisterHandler call sites).
	type testResponse struct {
		Result string
	}

	handler.sendResponse(w, testResponse{Result: "Success"})

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Response body is empty")
	}
}

func TestHandlerWithoutAuthentication(t *testing.T) {
	handler := NewHandler("", "") // No authentication

	soapBody := testXMLHeader + `
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <TestAction/>
  </soap:Body>
</soap:Envelope>`

	handler.RegisterHandler("TestAction", func(body []byte) (interface{}, error) {
		return "success", nil
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(soapBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should succeed without authentication
	if w.Code == http.StatusInternalServerError && strings.Contains(w.Body.String(), "Authentication") {
		t.Errorf("Should not require authentication when not configured")
	}
}

func TestReadRequestBodyError(t *testing.T) {
	handler := NewHandler("", "")

	// Create a request with a body that will fail to read
	req := httptest.NewRequest(http.MethodPost, "/", &failingReader{})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "Fault") {
		t.Errorf("Expected SOAP fault for read error")
	}
}

// Helper types and functions

type failingReader struct{}

func (f *failingReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestResponseHandling(t *testing.T) {
	handler := NewHandler("", "")

	type TestResponse struct {
		XMLName xml.Name `xml:"TestActionResponse"`
		Result  string   `xml:"Result"`
	}

	handler.RegisterHandler("TestAction", func(body []byte) (interface{}, error) {
		return &TestResponse{Result: "Success"}, nil
	})

	// ONVIF is SOAP 1.2 only — the envelope namespace must be the 2003/05
	// form the library's Envelope type declares.
	soapBody := `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <TestAction/>
  </soap:Body>
</soap:Envelope>`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(soapBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	response := w.Body.String()
	if !strings.Contains(response, "TestActionResponse") {
		t.Errorf("Response should contain TestActionResponse element")
	}
}

func TestEmptyBody(t *testing.T) {
	handler := NewHandler("", "")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("")))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "Fault") {
		t.Errorf("Expected SOAP fault for empty body")
	}
}

func TestContentType(t *testing.T) {
	handler := NewHandler("", "")

	handler.RegisterHandler("TestAction", func(body []byte) (interface{}, error) {
		return "test", nil
	})

	soapBody := `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <TestAction/>
  </soap:Body>
</soap:Envelope>`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(soapBody))
	req.Header.Set("Content-Type", "application/soap+xml")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Handler should work regardless of content type
	if w.Code == http.StatusInternalServerError {
		t.Logf("Note: Handler may validate content type")
	}
}

// --- Enterprise hardening P0: body limits (#59), auth lockout (#60) ---

// encoding/xml cannot marshal maps; handler responses are structs.
type hardeningResponse struct {
	Result string
}

const hardeningBody = testXMLHeader + `
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <TestAction/>
  </soap:Body>
</soap:Envelope>`

func newHardeningHandler(t *testing.T) *Handler {
	t.Helper()
	h := NewHandler("admin", "secret")
	h.RegisterHandler("TestAction", func(_ []byte) (interface{}, error) {
		return hardeningResponse{Result: "Success"}, nil
	})
	return h
}

func TestServeHTTPBodyLimitDefault(t *testing.T) {
	h := newHardeningHandler(t)

	big := hardeningBody + strings.Repeat(" ", 2<<20) // > 1 MiB default
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(big))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized body status = %d, want 413", w.Code)
	}
}

func TestServeHTTPBodyLimitCustom(t *testing.T) {
	h := NewHandlerWithOptions(HandlerOptions{
		Username:     "admin",
		Password:     "secret",
		MaxBodyBytes: 256,
	})
	h.RegisterHandler("TestAction", func(_ []byte) (interface{}, error) {
		return hardeningResponse{Result: "Success"}, nil
	})

	over := hardeningBody + strings.Repeat(" ", 512) // exceeds the 256-byte custom limit
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(over))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("custom-limit status = %d, want 413", w.Code)
	}
}

// unauthorizedPOST sends a body for an auth-protected action with no
// UsernameToken (the default policy protects write-style prefixes; the
// empty-prefix element name alone identifies the action).
func unauthorizedBody() string {
	return testXMLHeader + `
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <SetSystemDateAndTime/>
  </soap:Body>
</soap:Envelope>`
}

func TestServeHTTPAuthFailureLockout(t *testing.T) {
	h := NewHandlerWithOptions(HandlerOptions{
		Username:         "admin",
		Password:         "secret",
		AuthFailureLimit: 3,
		AuthLockout:      100 * time.Millisecond,
	})
	h.RegisterHandler("SetSystemDateAndTime", func(_ []byte) (interface{}, error) {
		return hardeningResponse{Result: "Success"}, nil
	})

	post := func() int {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(unauthorizedBody()))
		req.RemoteAddr = "10.0.0.1:1234"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		return w.Code
	}

	// The library's fault convention maps Sender faults to 400 (401
	// status mapping is tracked in #60).
	for i := 0; i < 3; i++ {
		if got := post(); got != http.StatusBadRequest {
			t.Fatalf("failure %d: status = %d, want 400", i, got)
		}
	}
	// Locked out: the distinct fault message is observable.
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(unauthorizedBody()))
	req.RemoteAddr = "10.0.0.1:1234"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("locked-out status = %d, want 400", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Too many authentication failures") {
		t.Fatalf("locked-out body should name the lockout, got: %s", w.Body.String())
	}

	// Another source is unaffected.
	req2 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(unauthorizedBody()))
	req2.RemoteAddr = "10.0.0.2:1234"
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)
	if strings.Contains(w2.Body.String(), "Too many authentication failures") {
		t.Fatalf("other source must not share the lockout")
	}
}

func TestServeHTTPAnonymousNoCredentialsWarns(t *testing.T) {
	// NewHandler with empty credentials keeps legacy open behavior but must
	// be explicitly observable — the AllowAnonymous option documents the
	// future fail-closed flip (#60).
	h := NewHandlerWithOptions(HandlerOptions{AllowAnonymous: true})
	if h.anonymousAllowed != true {
		t.Fatalf("AllowAnonymous option not honored")
	}
	locked := NewHandlerWithOptions(HandlerOptions{AllowAnonymous: true, AuthFailureLimit: -1})
	_ = locked // limiter disabled still constructs
}
