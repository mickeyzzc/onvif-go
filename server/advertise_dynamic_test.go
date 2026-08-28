package server

// Tests for the dynamic advertise-host sources (#45): DHCP environments
// change the device IP at runtime; advertised URLs must follow without a
// process restart.

import (
	"strings"
	"sync/atomic"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

const streamURIBody = `<GetStreamURI><ProfileToken>profile_token_1</ProfileToken></GetStreamURI>`

func newAdvertisedServer(t *testing.T, mutate func(*Config)) (*Server, *soap.RequestContext) {
	t.Helper()

	config := createTestConfig()
	if mutate != nil {
		mutate(config)
	}

	srv, err := New(config, WithStreamURIProvider(staticStreamProvider{
		info: streamInfoFor("profile_token_1"),
	}))
	if err != nil {
		t.Fatal(err)
	}

	return srv, &soap.RequestContext{RemoteIP: "192.0.2.9"}
}

// TestAdvertiseHostProviderDynamic pins the acceptance: after the device
// IP changes (provider returns the new value), GetStreamUri and
// GetCapabilities immediately advertise the new address — no restart.
func TestAdvertiseHostProviderDynamic(t *testing.T) {
	var host atomic.Value
	host.Store("192.0.2.10")

	srv, rc := newAdvertisedServer(t, func(c *Config) {
		c.AdvertiseHostProvider = func() string { return host.Load().(string) }
	})

	uri := streamURIFrom(t, srv, rc)
	if !strings.Contains(uri, "rtsp://192.0.2.10:") {
		t.Fatalf("before change: URI = %q", uri)
	}

	host.Store("192.0.2.99") // DHCP renewed a new lease

	uri = streamURIFrom(t, srv, rc)
	if !strings.Contains(uri, "rtsp://192.0.2.99:") {
		t.Fatalf("after change: URI = %q, want immediate 192.0.2.99", uri)
	}

	capsHost := capabilitiesXAddrHost(t, srv, rc)
	if capsHost != "192.0.2.99" {
		t.Fatalf("GetCapabilities host = %q, want 192.0.2.99", capsHost)
	}
}

// TestAdvertiseHostProviderWinsOverStatic: a provider result takes
// precedence over the static Config.AdvertiseHost (#45 requirement).
func TestAdvertiseHostProviderWinsOverStatic(t *testing.T) {
	srv, rc := newAdvertisedServer(t, func(c *Config) {
		c.AdvertiseHost = "static.example.org"
		c.AdvertiseHostProvider = func() string { return "dynamic.example.org" }
	})

	if uri := streamURIFrom(t, srv, rc); !strings.Contains(uri, "rtsp://dynamic.example.org:") {
		t.Fatalf("URI = %q, want provider host", uri)
	}
}

// TestSetAdvertiseHostRuntime: the setter updates the advertised host at
// runtime, replacing any configured provider.
func TestSetAdvertiseHostRuntime(t *testing.T) {
	srv, rc := newAdvertisedServer(t, func(c *Config) {
		c.AdvertiseHostProvider = func() string { return "old.example.org" }
	})

	srv.SetAdvertiseHost("10.0.0.5")
	if uri := streamURIFrom(t, srv, rc); !strings.Contains(uri, "rtsp://10.0.0.5:") {
		t.Fatalf("after SetAdvertiseHost: URI = %q", uri)
	}

	srv.SetAdvertiseHost("10.0.0.6")
	if uri := streamURIFrom(t, srv, rc); !strings.Contains(uri, "rtsp://10.0.0.6:") {
		t.Fatalf("after second SetAdvertiseHost: URI = %q", uri)
	}
}

// TestSetAdvertiseHostProviderRuntime: installing a provider at runtime
// restores dynamic resolution.
func TestSetAdvertiseHostProviderRuntime(t *testing.T) {
	var host atomic.Value
	host.Store("a.example.org")

	srv, rc := newAdvertisedServer(t, nil)

	srv.SetAdvertiseHostProvider(func() string { return host.Load().(string) })
	if uri := streamURIFrom(t, srv, rc); !strings.Contains(uri, "rtsp://a.example.org:") {
		t.Fatalf("URI = %q", uri)
	}

	host.Store("b.example.org")
	if uri := streamURIFrom(t, srv, rc); !strings.Contains(uri, "rtsp://b.example.org:") {
		t.Fatalf("URI after provider change = %q", uri)
	}
}

// TestAdvertiseHostFallsBackWithoutSources: no provider, no static
// override → the requester echo / config host path is unchanged.
func TestAdvertiseHostFallsBackWithoutSources(t *testing.T) {
	srv, rc := newAdvertisedServer(t, nil)

	if uri := streamURIFrom(t, srv, rc); !strings.Contains(uri, "rtsp://192.0.2.9:") {
		t.Fatalf("URI = %q, want requester-echo fallback", uri)
	}
}

func streamInfoFor(string) provider.StreamInfo {
	return provider.StreamInfo{RTSPPath: "/main"}
}

func streamURIFrom(t *testing.T, srv *Server, rc *soap.RequestContext) string {
	t.Helper()

	resp, err := srv.HandleGetStreamUri(rc, []byte(streamURIBody))
	if err != nil {
		t.Fatal(err)
	}

	return resp.(*GetStreamUriResponse).MediaUri.URI
}

func capabilitiesXAddrHost(t *testing.T, srv *Server, rc *soap.RequestContext) string {
	t.Helper()

	resp, err := srv.HandleGetCapabilities(rc, []byte(`<GetCapabilities/>`))
	if err != nil {
		t.Fatal(err)
	}

	caps, ok := resp.(*GetCapabilitiesResponse)
	if !ok || caps.Capabilities == nil {
		t.Fatalf("response shape: %T %+v", resp, resp)
	}

	media := caps.Capabilities.Media
	if media == nil || media.XAddr == "" {
		t.Fatalf("media capabilities missing: %+v", caps.Capabilities)
	}

	addr := media.XAddr
	if i := strings.Index(addr, "://"); i >= 0 {
		addr = addr[i+3:]
	}

	if i := strings.LastIndexByte(addr, ':'); i >= 0 {
		addr = addr[:i]
	}

	return addr
}
