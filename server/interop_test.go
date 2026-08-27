package server

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	onvif "github.com/mickeyzzc/onvif-go/v2/onvif"
)

// TestClientServerInterop drives the library's own client against the
// simulator end to end: endpoint discovery through GetCapabilities must
// yield working service URLs (the client-IP echo makes the httptest
// loopback address self-consistent), and the media operations must parse
// the server's canonical-cased responses.
func TestClientServerInterop(t *testing.T) {
	// The advertised port is config.Port (a server-side concern); bind
	// the test listener first so both agree.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()

	config := createTestConfig()
	config.Port = listener.Addr().(*net.TCPAddr).Port

	s, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	mux := http.NewServeMux()
	s.registerDeviceService(mux)
	s.registerMediaService(mux)

	ts := httptest.NewUnstartedServer(mux)
	ts.Listener = listener
	ts.Start()
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := onvif.NewClient(
		ts.URL+"/onvif/device_service",
		onvif.WithCredentials("admin", "password"),
		onvif.WithTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("onvif.NewClient() error = %v", err)
	}

	if err := client.Initialize(ctx); err != nil {
		t.Fatalf("client.Initialize() error = %v", err)
	}

	info, err := client.Device().GetDeviceInformation(ctx)
	if err != nil {
		t.Fatalf("GetDeviceInformation() error = %v", err)
	}

	if info.Manufacturer != config.DeviceInfo.Manufacturer {
		t.Errorf("Manufacturer = %q, want %q", info.Manufacturer, config.DeviceInfo.Manufacturer)
	}

	profiles, err := client.Media().GetProfiles(ctx)
	if err != nil {
		t.Fatalf("GetProfiles() error = %v", err)
	}

	if len(profiles) == 0 {
		t.Fatal("GetProfiles() returned no profiles")
	}

	wantToken := config.Profiles[0].Token
	if profiles[0].Token != wantToken {
		t.Errorf("profile token = %q, want %q", profiles[0].Token, wantToken)
	}

	streamURI, err := client.Media().GetStreamURI(ctx, wantToken)
	if err != nil {
		t.Fatalf("GetStreamURI() error = %v", err)
	}

	// The httptest loopback client IP must be echoed into the RTSP URI.
	if !strings.HasPrefix(streamURI.URI, "rtsp://127.0.0.1:") {
		t.Errorf("stream URI = %q, want rtsp://127.0.0.1:… (client-IP echo)", streamURI.URI)
	}

	snapshotURI, err := client.Media().GetSnapshotURI(ctx, wantToken)
	if err != nil {
		t.Fatalf("GetSnapshotURI() error = %v", err)
	}

	if !strings.Contains(snapshotURI.URI, "/onvif/snapshot?profile=") {
		t.Errorf("snapshot URI = %q, want /onvif/snapshot?profile=… form", snapshotURI.URI)
	}
}
