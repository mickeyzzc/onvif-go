package server

// WithMetrics (issue #66): hooks wired at the Server level flow through
// to the SOAP handlers — an embedded request fires SoapRequest with the
// dispatched action.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/metrics"
)

type serverCountingHooks struct {
	metrics.NoopHooks

	mu       sync.Mutex
	requests []string
}

func (h *serverCountingHooks) SoapRequest(action string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.requests = append(h.requests, action)
}

func (h *serverCountingHooks) counted() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.requests
}

func TestWithMetricsPlumbsToSOAPHandlers(t *testing.T) {
	hooks := &serverCountingHooks{}
	srv, err := New(createTestConfig(), WithMetrics(hooks))
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	srv.RegisterServices(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/onvif/device_service", "application/soap+xml", strings.NewReader(deviceInfoProbe))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	counted := hooks.counted()
	if len(counted) != 1 || counted[0] != "GetDeviceInformation" {
		t.Fatalf("counted requests = %v, want [GetDeviceInformation]", counted)
	}
}
