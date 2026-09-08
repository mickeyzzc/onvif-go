package soap

import (
	"bytes"
	"crypto/sha1" //nolint:gosec // SHA1 is the ONVIF digest formula, not a choice
	"encoding/base64"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// FuzzServeHTTP pins the server-side "hostile request never panics, and a
// protected action is never dispatched without a valid UsernameToken"
// invariants (issue #61). The seed corpus runs on every normal `go test`;
// extended fuzzing is `go test -fuzz`.
func FuzzServeHTTP(f *testing.F) {
	const (
		nonce   = "fuzz-nonce"
		created = "2026-01-02T03:04:05Z"
	)
	hash := sha1.New() //nolint:gosec // ONVIF digest formula
	hash.Write([]byte(nonce))
	hash.Write([]byte(created))
	hash.Write([]byte("secret"))
	digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	nonceB64 := base64.StdEncoding.EncodeToString([]byte(nonce))

	envelope := func(header, action string) string {
		return `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Header>` + header + `</s:Header>
  <s:Body>
    <` + action + ` xmlns="http://www.onvif.org/ver10/device/wsdl"/>
  </s:Body>
</s:Envelope>`
	}
	digestHeader := func(user, pass string) string {
		return `<wsse:Security xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
      <wsse:UsernameToken>
        <wsse:Username>` + user + `</wsse:Username>
        <wsse:Password Type="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest">` + pass + `</wsse:Password>
        <wsse:Nonce EncodingType="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary">` + nonceB64 + `</wsse:Nonce>
        <wsse:Created>` + created + `</wsse:Created>
      </wsse:UsernameToken>
    </wsse:Security>`
	}

	f.Add([]byte(envelope(digestHeader("admin", digest), "SetScopes")))                                                                                                                           // valid digest token, protected action
	f.Add([]byte(envelope(digestHeader("admin", "AAAA"), "SetScopes")))                                                                                                                           // wrong digest
	f.Add([]byte(envelope(digestHeader("", ""), "SetScopes")))                                                                                                                                    // empty token fields
	f.Add([]byte(envelope(digestHeader("admin", digest), "GetScopes")))                                                                                                                           // valid token, unprotected action
	f.Add([]byte(envelope("<wsse:Security xmlns:wsse=\"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd\"><wsse:UsernameToken/></wsse:Security>", "SetScopes"))) // empty token element
	f.Add([]byte(envelope("", "SetScopes")))                                                                                                                                                      // no header at all
	f.Add([]byte("<s:Envelope><s:Body>"))
	f.Add([]byte{0x00, 0xff, 0xfe})

	h := NewHandlerWithOptions(HandlerOptions{
		Username:         "admin",
		Password:         "secret",
		AuthFailureLimit: -1, // disable per-source lockout: fuzz inputs are adversarial by design
	})
	var protectedInvoked bool
	h.RegisterHandler("SetScopes", func(_ []byte) (interface{}, error) {
		protectedInvoked = true
		return nil, nil
	})
	h.RegisterHandler("GetScopes", func(_ []byte) (interface{}, error) {
		return nil, nil
	})

	f.Fuzz(func(t *testing.T, body []byte) {
		protectedInvoked = false
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.RemoteAddr = "192.0.2.1:1234"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if !protectedInvoked {
			return
		}
		// The protected action ran: re-decode the same envelope the server
		// saw and confirm authentication genuinely accepted it. This is the
		// "malformed token never reaches a protected handler" invariant —
		// dispatch without authenticate() would only be reachable by a
		// parser/auth bypass.
		var env requestEnvelope
		if err := xml.Unmarshal(body, &env); err != nil {
			t.Fatalf("protected action dispatched on unparseable envelope: %v", err)
		}
		if !h.authenticate(env.Header) {
			t.Fatal("protected action dispatched without a valid UsernameToken")
		}
		if strings.Contains(rec.Body.String(), "internal") && rec.Code >= 500 {
			t.Fatalf("server error on hostile input: %d %s", rec.Code, rec.Body.String())
		}
	})
}
