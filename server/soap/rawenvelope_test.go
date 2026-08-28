package soap

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/wsdiscovery"
)

// TestRawEnvelopePassthrough pins issue #39: a handler returning RawEnvelope
// has its bytes served as the complete response document — no extra
// Envelope/Body wrapping, no re-serialization. This is the channel a
// device-side WS-Discovery HTTP Probe handler needs (ProbeMatches is its
// own full envelope carrying WS-Addressing RelatesTo).
func TestRawEnvelopePassthrough(t *testing.T) {
	answer := wsdiscovery.BuildProbeMatches("abc-123", wsdiscovery.Match{
		EndpointRef:     "urn:uuid:test-device",
		Types:           "tds:Device",
		XAddrs:          "http://192.0.2.7/onvif/device_service",
		MetadataVersion: 1,
	})

	handler := NewHandler("", "")
	handler.RegisterContextHandler("Probe", func(_ *RequestContext, _ []byte) (interface{}, error) {
		return RawEnvelope(answer), nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <Probe xmlns="http://docs.oasis-open.org/ws-dd/ns/discovery/2009/01"/>
  </s:Body>
</s:Envelope>`))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}

	if got := rec.Header().Get("Content-Type"); got != "application/soap+xml; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want application/soap+xml; charset=utf-8", got)
	}

	if got := rec.Body.Bytes(); string(got) != string(answer) {
		t.Fatalf("response not verbatim:\n got: %q\nwant: %q", got, answer)
	}

	// Exactly one envelope: the handler's own.
	if n := strings.Count(rec.Body.String(), "<s:Envelope"); n != 1 {
		t.Fatalf("found %d <s:Envelope openings, want 1 (no double wrapping)", n)
	}

	if !strings.Contains(rec.Body.String(), "<a:RelatesTo>uuid:abc-123</a:RelatesTo>") {
		t.Fatal("RelatesTo missing from the raw answer")
	}
}

// TestRawEnvelopeNotRewrittenByPrefixMode ensures the explicit-prefix mode
// leaves RawEnvelope documents untouched too.
func TestRawEnvelopeNotRewrittenByPrefixMode(t *testing.T) {
	answer := []byte(`<?xml version="1.0"?><s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><ProbeMatches/></s:Body></s:Envelope>`)

	handler := NewHandlerWithOptions(HandlerOptions{ExplicitPrefixes: true})
	handler.RegisterHandler("Probe", func(_ []byte) (interface{}, error) {
		return RawEnvelope(answer), nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><Probe xmlns="http://docs.oasis-open.org/ws-dd/ns/discovery/2009/01"/></s:Body></s:Envelope>`))
	handler.ServeHTTP(rec, req)

	if got := rec.Body.Bytes(); string(got) != string(answer) {
		t.Fatalf("prefix mode rewrote RawEnvelope:\n got: %q\nwant: %q", got, answer)
	}
}
