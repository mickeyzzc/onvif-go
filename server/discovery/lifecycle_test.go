package discovery

import (
	"context"
	"strings"
	"testing"
	"time"
)

// waitFor is the bounded-wait guard for channel-based lifecycle tests.
func waitFor(t *testing.T, what string, ch <-chan struct{}) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatalf("%s: timed out waiting (possible deadlock)", what)
	}
}

func TestResponderStartStopLifecycle(t *testing.T) {
	responder := NewResponder(Config{EndpointRef: "urn:uuid:lc-1"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := responder.Start(ctx); err != nil {
		t.Skipf("host cannot join the WS-Discovery multicast group: %v", err)
	}

	// Double start is rejected.
	if err := responder.Start(ctx); err == nil {
		t.Error("second Start must fail")
	}

	responder.Stop()
	waitFor(t, "Done after Stop", responder.Done())

	// Stop is idempotent.
	responder.Stop()
}

func TestResponderBadInterfaceErrors(t *testing.T) {
	responder := NewResponder(Config{EndpointRef: "urn:uuid:lc-2", Interface: "no-such-iface-xyz"})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := responder.Start(ctx); err == nil {
		t.Error("Start with unknown interface must fail")
	}
}

func TestResponderDefaults(t *testing.T) {
	responder := NewResponder(Config{})

	// Defaults: ONVIF types, port 80, conventional device path, generated UUID.
	match := responder.matchFor(t.Context(), "198.51.100.9")

	if match.EndpointRef == "" {
		t.Error("EndpointRef default missing")
	}

	// Derived XAddrs: the device's own address toward the peer (#38 —
	// never the requester's), default port and device path.
	if !strings.HasSuffix(match.XAddrs, ":80/onvif/device_service") {
		t.Errorf("derived XAddrs = %q, want default port/path suffix", match.XAddrs)
	}

	if strings.Contains(match.XAddrs, "198.51.100.9") {
		t.Errorf("derived XAddrs = %q echoes the requester", match.XAddrs)
	}

	if !isLocalAddress(t, match.XAddrs) {
		t.Errorf("derived XAddrs = %q is not a local device address", match.XAddrs)
	}

	if match.Types == "" || match.MetadataVersion == 0 {
		t.Errorf("types/metadata defaults: %+v", match)
	}
}

func TestRandomUUIDShape(t *testing.T) {
	seen := make(map[string]bool)

	for range 100 {
		id := randomUUID()

		if len(id) != 36 || id[8] != '-' || id[13] != '-' {
			t.Fatalf("UUID shape broken: %q", id)
		}

		if seen[id] {
			t.Fatalf("UUID collision: %q", id)
		}

		seen[id] = true
	}
}
