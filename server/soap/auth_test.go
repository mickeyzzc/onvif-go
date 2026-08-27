package soap

import (
	"crypto/sha1" //nolint:gosec // SHA1 required by the ONVIF digest formula
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// buildAuthedRequest builds a SOAP 1.2 POST whose body invokes action and
// whose header carries a WS-Security UsernameToken. passwordForm selects
// the credential variant: "digest", "text", or "" for no security header.
func buildAuthedRequest(t *testing.T, action, username, password, passwordForm string) *http.Request {
	t.Helper()

	security := ""
	switch passwordForm {
	case "digest":
		nonce := "nonce-bytes"
		created := "2026-01-02T03:04:05Z"
		hash := sha1.New() //nolint:gosec // SHA1 required by the ONVIF digest formula
		hash.Write([]byte(nonce))
		hash.Write([]byte(created))
		hash.Write([]byte(password))
		digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))
		security = `<s:Header>
    <wsse:Security xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
      <wsse:UsernameToken>
        <wsse:Username>` + username + `</wsse:Username>
        <wsse:Password Type="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest">` + digest + `</wsse:Password>
        <wsse:Nonce EncodingType="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary">` + base64.StdEncoding.EncodeToString([]byte(nonce)) + `</wsse:Nonce>
        <wsu:Created xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-utility-1.0.xsd">` + created + `</wsu:Created>
      </wsse:UsernameToken>
    </wsse:Security>
  </s:Header>`
	case "text":
		security = `<s:Header>
    <wsse:Security xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
      <wsse:UsernameToken>
        <wsse:Username>` + username + `</wsse:Username>
        <wsse:Password Type="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordText">` + password + `</wsse:Password>
      </wsse:UsernameToken>
    </wsse:Security>
  </s:Header>`
	case "none":
		security = ""
	default:
		t.Fatalf("unknown password form %q", passwordForm)
	}

	soapBody := `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  ` + security + `
  <s:Body>
    <` + action + ` xmlns="http://www.onvif.org/ver10/device/wsdl"/>
  </s:Body>
</s:Envelope>`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(soapBody))
	req.RemoteAddr = "192.0.2.10:1234"

	return req
}

func TestRequiresAuthPolicyTable(t *testing.T) {
	tests := []struct {
		name        string
		configure   func(h *Handler)
		action      string
		requireAuth bool
	}{
		{
			name:        "no credentials means fully open",
			configure:   func(h *Handler) { h.username, h.password = "", "" },
			action:      "SetVideoEncoderConfiguration",
			requireAuth: false,
		},
		{
			name:        "default policy protects Set prefix",
			configure:   func(h *Handler) { h.username, h.password = "admin", "pass" },
			action:      "SetVideoEncoderConfiguration",
			requireAuth: true,
		},
		{
			name:        "default policy protects Remove prefix",
			configure:   func(h *Handler) { h.username, h.password = "admin", "pass" },
			action:      "RemoveScopes",
			requireAuth: true,
		},
		{
			name:        "default policy protects Create prefix",
			configure:   func(h *Handler) { h.username, h.password = "admin", "pass" },
			action:      "CreateUsers",
			requireAuth: true,
		},
		{
			name:        "default policy protects Go prefix",
			configure:   func(h *Handler) { h.username, h.password = "admin", "pass" },
			action:      "GotoPreset",
			requireAuth: true,
		},
		{
			name:        "default policy leaves reads open",
			configure:   func(h *Handler) { h.username, h.password = "admin", "pass" },
			action:      "GetDeviceInformation",
			requireAuth: false,
		},
		{
			name: "exact action list",
			configure: func(h *Handler) {
				h.username, h.password = "admin", "pass"
				h.auth = &AuthPolicy{Actions: []string{"SystemReboot"}}
			},
			action:      "SystemReboot",
			requireAuth: true,
		},
		{
			name: "strict mode protects reads too",
			configure: func(h *Handler) {
				h.username, h.password = "admin", "pass"
				h.auth = &AuthPolicy{All: true}
			},
			action:      "GetDeviceInformation",
			requireAuth: true,
		},
		{
			name: "custom prefixes replace defaults",
			configure: func(h *Handler) {
				h.username, h.password = "admin", "pass"
				h.auth = &AuthPolicy{Prefixes: []string{"Reboot"}}
			},
			action:      "SetVideoEncoderConfiguration",
			requireAuth: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler("", "")
			tt.configure(h)

			if got := h.requiresAuth(tt.action); got != tt.requireAuth {
				t.Errorf("requiresAuth(%q) = %v, want %v", tt.action, got, tt.requireAuth)
			}
		})
	}
}

func TestAuthEndToEnd(t *testing.T) {
	tests := []struct {
		name   string
		action string
		form   string // digest | text | none
		wantOK bool
	}{
		{name: "read action without credentials is open", action: "GetDeviceInformation", form: "none", wantOK: true},
		{name: "write action without credentials is rejected", action: "SetScopes", form: "none", wantOK: false},
		{name: "write action with valid digest passes", action: "SetScopes", form: "digest", wantOK: true},
		{name: "write action with valid password text passes", action: "SetScopes", form: "text", wantOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler("admin", "secret")

			var served bool
			h.RegisterContextHandler(tt.action, func(_ *RequestContext, _ []byte) (interface{}, error) {
				served = true

				return RawXML("<Done/>"), nil
			})

			w := httptest.NewRecorder()
			h.ServeHTTP(w, buildAuthedRequest(t, tt.action, "admin", "secret", tt.form))

			if tt.wantOK {
				if w.Code != http.StatusOK {
					t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
				}

				if !served {
					t.Error("handler was not reached")
				}

				return
			}

			if served {
				t.Error("handler must not be reached for rejected request")
			}

			if !strings.Contains(w.Body.String(), "not authorized") {
				t.Errorf("expected auth fault, got: %s", w.Body.String())
			}
		})
	}
}

func TestAuthenticateVariants(t *testing.T) {
	tests := []struct {
		name   string
		policy *AuthPolicy
		form   string
		user   string
		pass   string // password used to build the token
		want   bool
	}{
		{name: "valid digest", form: "digest", user: "admin", pass: "secret", want: true},
		{name: "wrong password digest", form: "digest", user: "admin", pass: "wrong", want: false},
		{name: "wrong username digest", form: "digest", user: "root", pass: "secret", want: false},
		{name: "valid password text", form: "text", user: "admin", pass: "secret", want: true},
		{
			name:   "password text rejected when policy disables it",
			policy: &AuthPolicy{AllowPasswordText: false},
			form:   "text", user: "admin", pass: "secret", want: false,
		},
		{name: "missing header", form: "none", user: "admin", pass: "secret", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandlerWithOptions(HandlerOptions{
				Username: "admin",
				Password: "secret",
				Auth:     tt.policy,
			})

			// Authenticate is exercised through the transport on a
			// protected action so the header path is realistic.
			h.RegisterContextHandler("SetScopes", func(_ *RequestContext, _ []byte) (interface{}, error) {
				return RawXML("<Done/>"), nil
			})

			w := httptest.NewRecorder()
			h.ServeHTTP(w, buildAuthedRequest(t, "SetScopes", tt.user, tt.pass, tt.form))

			got := w.Code == http.StatusOK
			if got != tt.want {
				t.Errorf("authenticated = %v, want %v; status %d body %s", got, tt.want, w.Code, w.Body.String())
			}
		})
	}
}

func TestLegacyActionSpellingAccepted(t *testing.T) {
	h := NewHandler("admin", "secret")

	var served bool
	h.RegisterContextHandler("GetStreamUri", func(rc *RequestContext, _ []byte) (interface{}, error) {
		served = true

		if rc.Action != "GetStreamUri" {
			t.Errorf("RequestContext.Action = %q, want canonical spelling", rc.Action)
		}

		return RawXML("<Done/>"), nil
	})

	w := httptest.NewRecorder()
	h.ServeHTTP(w, buildAuthedRequest(t, "GetStreamURI", "admin", "secret", "none"))

	if !served {
		t.Fatalf("legacy spelling not dispatched; body: %s", w.Body.String())
	}

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}
