package onvif

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// capsTestServer counts GetCapabilities requests and answers with a minimal
// valid capabilities envelope (or a fault when faulting is set).
func capsTestServer(t *testing.T, faulting *atomic.Bool) (*httptest.Server, *atomic.Int32) {
	t.Helper()

	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests.Add(1)

		if !strings.Contains(string(body), "GetCapabilities") {
			t.Errorf("unexpected non-GetCapabilities request: %.100s", body)
		}

		if faulting != nil && faulting.Load() {
			w.Header().Set("Content-Type", "application/soap+xml")
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<s:Fault><s:Code><s:Value>s:Receiver</s:Value></s:Code>
<s:Reason><s:Text>Not Implemented</s:Text></s:Reason></s:Fault>
</s:Body></s:Envelope>`))

			return
		}

		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<tds:GetCapabilitiesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:Capabilities>
<tds:Media><tds:XAddr>http://device/onvif/media_service</tds:XAddr></tds:Media>
<tds:PTZ><tds:XAddr>http://device/onvif/ptz_service</tds:XAddr></tds:PTZ>
</tds:Capabilities>
</tds:GetCapabilitiesResponse>
</s:Body></s:Envelope>`))
	}))
	t.Cleanup(server.Close)

	return server, &requests
}

func newCapsClient(t *testing.T, server *httptest.Server, opts ...ClientOption) *Client {
	t.Helper()

	allOpts := append([]ClientOption{WithHTTPClient(server.Client())}, opts...)

	client, err := NewClient(server.URL, allOpts...)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	return client
}

func TestGetCapabilitiesCachedHitsCache(t *testing.T) {
	server, requests := capsTestServer(t, nil)
	client := newCapsClient(t, server)

	first, err := client.Device().GetCapabilitiesCached(context.Background())
	if err != nil {
		t.Fatalf("first GetCapabilitiesCached() error = %v", err)
	}

	second, err := client.Device().GetCapabilitiesCached(context.Background())
	if err != nil {
		t.Fatalf("second GetCapabilitiesCached() error = %v", err)
	}

	if got := requests.Load(); got != 1 {
		t.Errorf("device saw %d requests, want 1 (second call must be a cache hit)", got)
	}

	if first != second {
		t.Error("cache returned different pointers for the same entry")
	}

	if first.Media == nil {
		t.Error("capabilities did not parse (Media == nil)")
	}
}

func TestInvalidateCapabilitiesCacheRefetches(t *testing.T) {
	server, requests := capsTestServer(t, nil)
	client := newCapsClient(t, server)

	if _, err := client.Device().GetCapabilitiesCached(context.Background()); err != nil {
		t.Fatalf("first call error = %v", err)
	}

	client.InvalidateCapabilitiesCache()

	if _, err := client.Device().GetCapabilitiesCached(context.Background()); err != nil {
		t.Fatalf("post-invalidate call error = %v", err)
	}

	if got := requests.Load(); got != 2 {
		t.Errorf("device saw %d requests, want 2 after invalidation", got)
	}
}

func TestGetCapabilitiesCachedMinimalFallback(t *testing.T) {
	var faulting atomic.Bool
	faulting.Store(true)
	server, requests := capsTestServer(t, &faulting)
	client := newCapsClient(t, server, WithMinimalCapsFallback())

	caps, err := client.Device().GetCapabilitiesCached(context.Background())
	if err != nil {
		t.Fatalf("GetCapabilitiesCached() with minimal fallback error = %v", err)
	}

	if caps == nil {
		t.Fatal("minimal capability set is nil")
	}

	if caps.Media != nil {
		t.Error("minimal capability set must not advertise services")
	}

	// The degraded result must be cached: no re-hammering the device.
	if _, err := client.Device().GetCapabilitiesCached(context.Background()); err != nil {
		t.Fatalf("cached degraded call error = %v", err)
	}

	if got := requests.Load(); got != 1 {
		t.Errorf("device saw %d requests, want 1 (degraded result must be cached)", got)
	}
}

func TestGetCapabilitiesCachedDefaultIsConservative(t *testing.T) {
	var faulting atomic.Bool
	faulting.Store(true)
	server, requests := capsTestServer(t, &faulting)
	client := newCapsClient(t, server) // no fallback option

	if _, err := client.Device().GetCapabilitiesCached(context.Background()); err == nil {
		t.Fatal("GetCapabilitiesCached() succeeded against a faulting device, want error")
	}

	// Not cached: the failure must surface again on the next call.
	if _, err := client.Device().GetCapabilitiesCached(context.Background()); err == nil {
		t.Fatal("second call succeeded, want repeated error")
	}

	if got := requests.Load(); got != 2 {
		t.Errorf("device saw %d requests, want 2 (failures must not be cached by default)", got)
	}
}

func TestGetCapabilitiesCachedConcurrentSingleFlight(t *testing.T) {
	server, requests := capsTestServer(t, nil)
	client := newCapsClient(t, server)

	const callers = 10

	var wg sync.WaitGroup
	errs := make([]error, callers)
	start := make(chan struct{})

	for i := range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := client.Device().GetCapabilitiesCached(context.Background())
			errs[i] = err
		}()
	}

	close(start)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent caller %d error = %v", i, err)
		}
	}

	if got := requests.Load(); got != 1 {
		t.Errorf("device saw %d requests for %d concurrent callers, want 1", got, callers)
	}
}
