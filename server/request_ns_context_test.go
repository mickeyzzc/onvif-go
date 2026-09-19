package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A client that declares its namespace prefixes on the ENVELOPE (legal
// SOAP; the bindings apply to the whole document) must reach the handler
// with resolvable namespaces. The innerxml body extraction drops those
// ancestor declarations, leaving unbound prefixes in the fragment handed
// to handlers.
func TestPrefixedRequestWithEnvelopeLevelDeclarations(t *testing.T) {
	cfg := createTestConfig()
	cfg.SupportPTZ = true

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/onvif/ptz_service", strings.NewReader(
		`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema">
  <s:Body>
    <tptz:ContinuousMove>
      <tptz:ProfileToken>profile_token_1</tptz:ProfileToken>
      <tptz:Velocity><tt:PanTilt x="0.5" y="0.0"/></tptz:Velocity>
    </tptz:ContinuousMove>
  </s:Body>
</s:Envelope>`))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
}
