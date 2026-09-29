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

// A dual-face device serves BOTH Media1 (GetCapabilities ver10 Media
// XAddr) and Media2 (GetServices ver20/media XAddr at a DIFFERENT
// address). The ver20/media entry in GetServices is the authoritative
// Media2 address — Media2 calls must not ride the Media1 endpoint when
// the device advertises a dedicated one (the Media1 endpoint dispatches
// by action local name, so a tr2 GetStreamUri posted there gets the
// Media1-shaped answer).
func TestInitializePrefersGetServicesMedia2XAddrOnDualFaceDevices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml")

		switch {
		case strings.Contains(string(body), "GetCapabilities"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<tds:GetCapabilitiesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:Capabilities>
<tds:Device><tds:XAddr>http://device/onvif/device_service</tds:XAddr></tds:Device>
<tds:Media><tds:XAddr>http://device/onvif/media_service</tds:XAddr></tds:Media>
</tds:Capabilities>
</tds:GetCapabilitiesResponse>
</s:Body></s:Envelope>`))
		case strings.Contains(string(body), "GetServices"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<tds:GetServicesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:Service>
<tds:Namespace>http://www.onvif.org/ver10/media/wsdl</tds:Namespace>
<tds:XAddr>http://device/onvif/media_service</tds:XAddr>
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

	base := strings.TrimSuffix(srv.URL, "/")
	if got, want := c.EndpointFor(api.ServiceMedia), base+"/onvif/media_service"; got != want {
		t.Errorf("media1 endpoint = %q, want %q", got, want)
	}
	if got, want := c.EndpointFor(api.ServiceMedia2), base+"/onvif/media2_service"; got != want {
		t.Errorf("media2 endpoint = %q, want %q (the dedicated ver20/media address from GetServices)", got, want)
	}
}
