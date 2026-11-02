package server

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
)

// startOnZero starts a server bound to an ephemeral port and waits for
// the listener; the returned cancel stops it.
func startOnZero(t *testing.T, cfg *Config) (*Server, context.CancelFunc) {
	t.Helper()
	cfg.Host = "127.0.0.1"
	cfg.Port = 0
	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start(ctx) }()

	deadline := time.Now().Add(2 * time.Second)
	for srv.ListenAddr() == "" {
		if time.Now().After(deadline) {
			cancel()
			t.Fatal("server never bound a listener")
		}
		time.Sleep(2 * time.Millisecond)
	}

	t.Cleanup(func() {
		cancel()
		select {
		case <-errCh:
		case <-time.After(2 * time.Second):
			t.Error("server did not stop within 2s of cancel")
		}
	})

	return srv, cancel
}

// Issue #161: with Config.Port == 0 the server binds a kernel-assigned
// port, and every advertisement — service XAddrs in GetCapabilities and
// GetServices, the snapshot URI — must carry that actual port, never :0.
func TestPortZeroAdvertisesActualPort(t *testing.T) {
	srv, _ := startOnZero(t, createTestConfig())

	listenPort := srv.ListenAddr()[strings.LastIndex(srv.ListenAddr(), ":")+1:]
	if listenPort == "" || listenPort == "0" {
		t.Fatalf("listener port %q is not an ephemeral port", listenPort)
	}
	if got := srv.advertisedPort(); got == 0 {
		t.Fatalf("advertisedPort() = 0 with a bound listener, want %s", listenPort)
	} else if got != atoi(t, listenPort) {
		t.Fatalf("advertisedPort() = %d, want %s", got, listenPort)
	}

	// GetCapabilities: the device/media XAddrs ride the actual port.
	capsResp, err := srv.HandleGetCapabilities(nil, nil)
	if err != nil {
		t.Fatalf("HandleGetCapabilities() error = %v", err)
	}
	for _, xaddr := range []string{capsResp.(*GetCapabilitiesResponse).Capabilities.Device.XAddr} {
		if strings.Contains(xaddr, ":0/") {
			t.Errorf("capabilities XAddr advertises :0: %s", xaddr)
		}
		if !strings.HasSuffix(xaddr, ":"+listenPort+"/onvif/device_service") {
			t.Errorf("capabilities XAddr %s does not use listener port %s", xaddr, listenPort)
		}
	}

	// GetServices: same contract (the flow a real client follows).
	svcs, err := srv.HandleGetServices(nil, nil)
	if err != nil {
		t.Fatalf("HandleGetServices() error = %v", err)
	}
	for _, svc := range svcs.(*GetServicesResponse).Service {
		if strings.Contains(svc.XAddr, ":0/") {
			t.Errorf("service %s advertises :0: %s", svc.Namespace, svc.XAddr)
		}
		if !strings.Contains(svc.XAddr, ":"+listenPort+"/") {
			t.Errorf("service XAddr %s does not use listener port %s", svc.XAddr, listenPort)
		}
	}

	// Snapshot URI: same shared helper.
	snap, err := srv.HandleGetSnapshotUri(nil, []byte(`<GetSnapshotURI><ProfileToken>profile_token_1</ProfileToken></GetSnapshotURI>`))
	if err != nil {
		t.Fatalf("HandleGetSnapshotUri() error = %v", err)
	}
	snapURI := snap.(*GetSnapshotUriResponse).MediaUri.URI
	if strings.Contains(snapURI, ":0/") {
		t.Errorf("snapshot URI advertises :0: %s", snapURI)
	}
	if !strings.Contains(snapURI, ":"+listenPort+"/") {
		t.Errorf("snapshot URI %s does not use listener port %s", snapURI, listenPort)
	}
}

// A configured (non-zero) port is published verbatim — the #161 fix must
// not disturb the normal path.
func TestConfiguredPortAdvertisedVerbatim(t *testing.T) {
	config := createTestConfig() // Port 8080
	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if got := srv.advertisedPort(); got != 8080 {
		t.Fatalf("advertisedPort() = %d, want 8080", got)
	}
}

func atoi(t *testing.T, s string) int {
	t.Helper()
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			t.Fatalf("non-numeric port %q", s)
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// Issue #162: the simulator answers GetMetadataUri with the deterministic
// per-profile path, both media faces.
func TestHandleGetMetadataUri(t *testing.T) {
	config := createTestConfig()
	srv, _ := New(config)

	resp, err := srv.HandleGetMetadataUri(nil, []byte(`<GetMetadataUri><ProfileToken>profile_token_1</ProfileToken></GetMetadataUri>`))
	if err != nil {
		t.Fatalf("HandleGetMetadataUri() error = %v", err)
	}
	uri := resp.(*GetMetadataUriResponse).MediaUri.URI
	if !strings.HasPrefix(uri, "rtsp://127.0.0.1:8554/metadata0") {
		t.Fatalf("metadata URI = %q, want rtsp://127.0.0.1:8554/metadata0…", uri)
	}

	// Unknown token → the profile fault, same as GetStreamUri.
	if _, err := srv.HandleGetMetadataUri(nil, []byte(`<GetMetadataUri><ProfileToken>nope</ProfileToken></GetMetadataUri>`)); !errors.Is(err, provider.ErrProfileNotFound) {
		t.Fatalf("unknown token error = %v, want ErrProfileNotFound", err)
	}

	// The pinned override wins verbatim (MetadataURISetter on the sim).
	setter, ok := srv.metadata.(provider.MetadataURISetter)
	if !ok {
		t.Fatal("simulator does not implement MetadataURISetter")
	}
	if err := setter.SetMetadataURI("profile_token_1", "rtsp://lab:7554/meta/custom"); err != nil {
		t.Fatalf("SetMetadataURI() error = %v", err)
	}
	resp, err = srv.HandleGetMetadataUri(nil, []byte(`<GetMetadataUri><ProfileToken>profile_token_1</ProfileToken></GetMetadataUri>`))
	if err != nil {
		t.Fatalf("HandleGetMetadataUri() after override error = %v", err)
	}
	if got := resp.(*GetMetadataUriResponse).MediaUri.URI; got != "rtsp://lab:7554/meta/custom" {
		t.Fatalf("override metadata URI = %q, want the pinned value", got)
	}
}

func TestMedia2GetMetadataUri(t *testing.T) {
	config := createTestConfig()
	config.SupportMedia2 = true
	srv, _ := New(config)

	resp, err := srv.HandleMedia2GetMetadataUri(nil, []byte(`<GetMetadataUri><ProfileToken>profile_token_1</ProfileToken></GetMetadataUri>`))
	if err != nil {
		t.Fatalf("HandleMedia2GetMetadataUri() error = %v", err)
	}
	if got := resp.(media2GetMetadataUriResponse).Uri; !strings.HasPrefix(got, "rtsp://127.0.0.1:8554/metadata0") {
		t.Fatalf("Media2 metadata URI = %q, want rtsp://127.0.0.1:8554/metadata0…", got)
	}
}

// A host-supplied metadata provider replaces the simulator's answers.
func TestWithMetadataURIProvider(t *testing.T) {
	fake := &fakeMetadataProvider{info: provider.MetadataInfo{RTSPPath: "/cameralab/meta", RTSPPort: 7554}}
	config := createTestConfig()
	srv, err := New(config, WithMetadataURIProvider(fake))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, err := srv.HandleGetMetadataUri(nil, []byte(`<GetMetadataUri><ProfileToken>profile_token_1</ProfileToken></GetMetadataUri>`))
	if err != nil {
		t.Fatalf("HandleGetMetadataUri() error = %v", err)
	}
	if got := resp.(*GetMetadataUriResponse).MediaUri.URI; got != "rtsp://127.0.0.1:7554/cameralab/meta" {
		t.Fatalf("custom provider metadata URI = %q", got)
	}
	if fake.lastToken != "profile_token_1" {
		t.Fatalf("provider got token %q", fake.lastToken)
	}
}

type fakeMetadataProvider struct {
	info      provider.MetadataInfo
	lastToken string
}

func (f *fakeMetadataProvider) MetadataURI(profileToken string) (provider.MetadataInfo, error) {
	f.lastToken = profileToken
	return f.info, nil
}
