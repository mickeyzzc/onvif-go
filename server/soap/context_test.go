package soap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRequestContextPopulated(t *testing.T) {
	h := NewHandler("", "")

	var gotCtx *RequestContext
	h.RegisterContextHandler("GetDeviceInformation", func(rc *RequestContext, _ []byte) (interface{}, error) {
		gotCtx = rc

		return RawXML("<Done/>"), nil
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><tds:GetDeviceInformation xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/></s:Body></s:Envelope>`))
	req.RemoteAddr = "203.0.113.9:41234"

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if gotCtx == nil {
		t.Fatal("handler not reached")
	}

	if gotCtx.Action != "GetDeviceInformation" {
		t.Errorf("Action = %q, want GetDeviceInformation", gotCtx.Action)
	}

	if gotCtx.RemoteIP != "203.0.113.9" {
		t.Errorf("RemoteIP = %q, want 203.0.113.9", gotCtx.RemoteIP)
	}

	if gotCtx.Request != req {
		t.Error("Request not carried through")
	}

	if gotCtx.Context() != req.Context() {
		t.Error("Context() must return the request context")
	}
}

func TestRequestContextCancellationVisible(t *testing.T) {
	h := NewHandler("", "")

	deadline := time.Now().Add(-time.Second) // already expired

	var sawDone bool
	h.RegisterContextHandler("Slow", func(rc *RequestContext, _ []byte) (interface{}, error) {
		select {
		case <-rc.Context().Done():
			sawDone = true
		case <-time.After(2 * time.Second):
			t.Error("cancellation not observed in time")
		}

		return RawXML("<Done/>"), nil
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><Slow/></s:Body></s:Envelope>`))

	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if !sawDone {
		t.Error("handler context was not cancelled")
	}
}

func TestNilRequestContextHelpers(t *testing.T) {
	var rc *RequestContext

	if rc.Context() == nil {
		t.Error("Context() on nil RequestContext must return context.Background")
	}
}

func TestRemoteIPExtraction(t *testing.T) {
	tests := []struct {
		remoteAddr string
		want       string
	}{
		{remoteAddr: "192.0.2.4:8080", want: "192.0.2.4"},
		{remoteAddr: "[2001:db8::1]:8080", want: "2001:db8::1"},
		{remoteAddr: "192.0.2.4", want: "192.0.2.4"},
		{remoteAddr: "", want: ""},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.RemoteAddr = tt.remoteAddr

		if got := remoteIP(req); got != tt.want {
			t.Errorf("remoteIP(%q) = %q, want %q", tt.remoteAddr, got, tt.want)
		}
	}
}

func TestLegacyMessageHandlerWrapper(t *testing.T) {
	h := NewHandler("", "")

	var legacyCalled bool
	h.RegisterHandler("Legacy", func(body []byte) (interface{}, error) {
		legacyCalled = true

		if string(body) != `<Legacy token="x"/>` {
			t.Errorf("legacy handler body = %q", string(body))
		}

		return RawXML("<Done/>"), nil
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><Legacy token="x"/></s:Body></s:Envelope>`))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if !legacyCalled {
		t.Fatalf("legacy handler not reached; body: %s", w.Body.String())
	}

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRequestBodyPassesAsInnerXML(t *testing.T) {
	h := NewHandler("", "")

	h.RegisterContextHandler("GetStreamUri", func(_ *RequestContext, body []byte) (interface{}, error) {
		var req struct {
			ProfileToken string `xml:"ProfileToken"`
		}
		if err := ParseRequest(body, &req); err != nil {
			return nil, err
		}

		if req.ProfileToken != "profile_1" {
			t.Errorf("ProfileToken = %q, want profile_1 (request decoding through the transport)", req.ProfileToken)
		}

		return RawXML("<Done/>"), nil
	})

	soapReq := `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <GetStreamUri xmlns="http://www.onvif.org/ver10/media/wsdl">
      <StreamSetup><Stream>RTP_Unicast</Stream><Transport><Protocol>RTSP</Protocol></Transport></StreamSetup>
      <ProfileToken>profile_1</ProfileToken>
    </GetStreamUri>
  </s:Body>
</s:Envelope>`

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(soapReq)))

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
}
