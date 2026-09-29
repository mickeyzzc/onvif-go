package onvif

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
)

// Issue #110: a Media2-only (Profile T) device advertises no ver10 Media
// XAddr in GetCapabilities. Initialize must then fall back to GetServices
// and wire the ver20/media endpoint so client.Media2() works.
func TestInitializeDiscoversMedia2OnlyEndpointViaGetServices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml")

		switch {
		case strings.Contains(string(body), "GetCapabilities"):
			// No ver10 Media section at all — a Media2-only device.
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<tds:GetCapabilitiesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:Capabilities>
<tds:Device><tds:XAddr>http://device/onvif/device_service</tds:XAddr></tds:Device>
</tds:Capabilities>
</tds:GetCapabilitiesResponse>
</s:Body></s:Envelope>`))
		case strings.Contains(string(body), "GetServices"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<tds:GetServicesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:Service>
<tds:Namespace>http://www.onvif.org/ver10/device/wsdl</tds:Namespace>
<tds:XAddr>http://device/onvif/device_service</tds:XAddr>
</tds:Service>
<tds:Service>
<tds:Namespace>http://www.onvif.org/ver20/media/wsdl</tds:Namespace>
<tds:XAddr>http://device/onvif/media2_service</tds:XAddr>
</tds:Service>
</tds:GetServicesResponse>
</s:Body></s:Envelope>`))
		default:
			t.Errorf("unexpected request: %.120s", body)
		}
	}))
	t.Cleanup(srv.Close)

	c, err := NewClient(srv.URL, WithCredentials("u", "p"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if err := c.Initialize(context.Background()); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	got := c.EndpointFor(api.ServiceMedia2)
	base := strings.TrimSuffix(srv.URL, "/")
	want := base + "/onvif/media2_service"
	if got != want {
		t.Errorf("media2 endpoint = %q, want %q (host-fixed from the advertised XAddr)", got, want)
	}
}
