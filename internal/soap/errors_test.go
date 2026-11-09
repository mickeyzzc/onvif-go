package soap

// Unit tests for the client-side error classification helpers and the
// Username Created accessor (issue #40's tolerant decode).

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestFaultErrorMessage(t *testing.T) {
	fault := &FaultError{Code: "ter:Receiver", Reason: "boom"}
	if msg := fault.Error(); msg != "SOAP fault [ter:Receiver]: boom" {
		t.Errorf("Error() = %q", msg)
	}

	fault.Subcode = "ter:ActionNotSupported"
	if msg := fault.Error(); !strings.Contains(msg, "[ter:Receiver/ter:ActionNotSupported]") {
		t.Errorf("Error() with subcode = %q", msg)
	}
}

func TestFaultIsAuthFailure(t *testing.T) {
	yes := []*FaultError{
		{Code: "ter:NotAuthorized"},
		{Subcode: "ter:NotAuthorized"},
		{Reason: "Sender not Authorized"},
		{Reason: "the request was unauthorized"},
	}
	for _, f := range yes {
		if !f.IsAuthFailure() {
			t.Errorf("IsAuthFailure(%+v) = false, want true", f)
		}
	}

	no := &FaultError{Code: "ter:ActionNotSupported", Reason: "unknown action"}
	if no.IsAuthFailure() {
		t.Errorf("IsAuthFailure(%+v) = true, want false", no)
	}
}

func TestHTTPStatusError(t *testing.T) {
	err := &HTTPStatusError{Status: 401, Body: "denied"}

	if !errors.Is(err, ErrHTTPRequestFailed) {
		t.Error("Unwrap chain to ErrHTTPRequestFailed broken")
	}

	if !err.IsAuthFailure() {
		t.Error("401 must classify as auth failure")
	}

	if (&HTTPStatusError{Status: 500}).IsAuthFailure() {
		t.Error("500 must not classify as auth failure")
	}

	long := &HTTPStatusError{Status: 502, Body: strings.Repeat("x", maxErrorBodyLen+100)}
	if !strings.Contains(long.Error(), "...(truncated)") {
		t.Error("body not truncated in error message")
	}
}

func TestIsAuthFailureClassification(t *testing.T) {
	if !IsAuthFailure(&HTTPStatusError{Status: 403}) {
		t.Error("403 classified wrong")
	}

	if !IsAuthFailure(&FaultError{Code: "ter:NotAuthorized"}) {
		t.Error("fault classified wrong")
	}

	if !IsAuthFailure(errors.Join(errors.New("wrap"), &HTTPStatusError{Status: 401})) {
		t.Error("joined chain classified wrong")
	}

	if IsAuthFailure(errors.New("boom")) {
		t.Error("generic error classified as auth failure")
	}
}

func TestTruncateBody(t *testing.T) {
	short := "body"
	if truncateBody(short) != short {
		t.Error("short body altered")
	}

	long := strings.Repeat("y", maxErrorBodyLen+1)
	if got := truncateBody(long); len(got) != maxErrorBodyLen+len("...(truncated)") {
		t.Errorf("truncated length = %d", len(got))
	}
}

func TestUsernameTokenCreatedValue(t *testing.T) {
	token := &UsernameToken{}
	if token.CreatedValue() != "" {
		t.Error("empty token must yield empty created value")
	}

	token.Created = "canonical"
	if token.CreatedValue() != "canonical" {
		t.Error("canonical value not preferred")
	}

	token.CreatedVariant = "variant"
	if token.CreatedValue() != "canonical" {
		t.Error("canonical must win over variant")
	}

	token.Created = ""
	if token.CreatedValue() != "variant" {
		t.Error("variant not used when canonical empty")
	}
}

func TestClientSetters(t *testing.T) {
	client := NewClient(nil, "", "")

	client.SetAuthMode(AuthModePasswordText)
	if client.authMode != AuthModePasswordText {
		t.Error("SetAuthMode not applied")
	}

	calls := 0
	client.SetDebug(true, func(string, ...interface{}) { calls++ })
	client.logDebugf("hello")
	if calls != 1 {
		t.Error("debug logger not invoked")
	}

	client.SetDebug(false, nil)
	client.logDebugf("quiet") // must not panic with nil logger

	client.SetClockSkew(90 * time.Second)
	if client.clockSkew != 90*time.Second {
		t.Error("SetClockSkew not applied")
	}
}

// TestParseFaultLenientShapes pins the classification of abbreviated fault
// envelopes real devices emit: the fault code/subcode hoisted directly under
// Fault without the SOAP 1.2 <Code>/<Subcode> wrappers. Strict-path XML
// decoding sees none of those fields; without a lenient fallback the fault
// classifies as an empty FaultError and a definitive credential rejection
// turns ambiguous downstream (regression class from MiBeeNvr's credential
// prober: HTTP 500 + s:Fault/s:Value/s:Subcode was classified correctly via
// the legacy raw-body sniff and lost after fault classification landed).
func TestParseFaultLenientShapes(t *testing.T) {
	// parseFault receives the envelope Body's inner XML, so the Fault element
	// is the root of the input.
	body := []byte(`<s:Fault><s:Value>s:Sender</s:Value>` +
		`<s:Subcode><s:Value>ter:NotAuthorized</s:Value></s:Subcode>` +
		`</s:Fault>`)

	fault := parseFault(body, 500)
	if fault == nil {
		t.Fatal("parseFault returned nil for an abbreviated fault envelope")
	}

	if fault.Code != "s:Sender" {
		t.Errorf("Code = %q, want s:Sender", fault.Code)
	}

	if fault.Subcode != "ter:NotAuthorized" {
		t.Errorf("Subcode = %q, want ter:NotAuthorized", fault.Subcode)
	}

	if !fault.IsAuthFailure() {
		t.Error("IsAuthFailure = false, want true for ter:NotAuthorized")
	}

	if !IsAuthFailure(fault) {
		t.Error("package-level IsAuthFailure = false, want true")
	}
}

// TestParseFaultLenientReasonText: the abbreviated shape may still carry a
// SOAP 1.2 Reason/Text outside the strict Code hierarchy.
func TestParseFaultLenientReasonText(t *testing.T) {
	body := []byte(`<s:Fault><s:Code><s:Value>s:Sender</s:Value></s:Code>` +
		`<s:Reason><s:Text xml:lang="en">Sender not authorized</s:Text></s:Reason>` +
		`</s:Fault>`)

	fault := parseFault(body, 200)
	if fault == nil {
		t.Fatal("parseFault returned nil")
	}

	if fault.Reason == "" {
		t.Error("Reason empty; xml:lang-attributed Text must still be captured")
	}

	if !fault.IsAuthFailure() {
		t.Error("IsAuthFailure = false, want true via the reason phrasing")
	}
}
