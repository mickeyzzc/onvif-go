package discovery

// Metrics seam (issue #66): every answered WS-Discovery Probe — multicast
// datagram or directed HTTP — counts DiscoveryProbeAnswered; unanswered
// probes (wrong types, malformed datagrams) count nothing.

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/wsdiscovery"
)

type probeCounter struct {
	mu     sync.Mutex
	probes int
}

func (c *probeCounter) SoapRequest(string)      {}
func (c *probeCounter) SoapFault(string)        {}
func (c *probeCounter) AuthFail()               {}
func (c *probeCounter) AuthLockout()            {}
func (c *probeCounter) DiscoveryProbeAnswered() { c.mu.Lock(); c.probes++; c.mu.Unlock() }

func (c *probeCounter) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.probes
}

func TestMetricsDatagramProbeAnswered(t *testing.T) {
	counter := &probeCounter{}
	responder := NewResponder(Config{
		EndpointRef: "urn:uuid:resp-metrics-1",
		XAddrs:      []string{"http://camera.example.org/onvif/device_service"},
		Metrics:     counter,
	})

	src := &net.UDPAddr{IP: net.IPv4(10, 1, 2, 3), Port: 9}
	capture := &captureSend{}
	responder.handleDatagram(t.Context(), wsdiscovery.BuildProbe("probe-1"), src, capture.send)

	if len(capture.replies) != 1 {
		t.Fatalf("replies = %d, want 1", len(capture.replies))
	}
	if got := counter.count(); got != 1 {
		t.Errorf("probes counted = %d, want 1", got)
	}
}

func TestMetricsDatagramNonMatchingTypesCountNothing(t *testing.T) {
	counter := &probeCounter{}
	responder := NewResponder(Config{
		EndpointRef: "urn:uuid:resp-metrics-2",
		Types:       []string{"tds:Device"},
		Metrics:     counter,
	})

	// The default BuildProbe asks for NetworkVideoTransmitter etc.; with
	// the responder only answering tds:Device probes this must be skipped.
	src := &net.UDPAddr{IP: net.IPv4(10, 1, 2, 3), Port: 9}
	capture := &captureSend{}
	responder.handleDatagram(t.Context(), wsdiscovery.BuildProbe("probe-2"), src, capture.send)

	if len(capture.replies) != 0 {
		t.Fatalf("replies = %d, want 0", len(capture.replies))
	}
	if got := counter.count(); got != 0 {
		t.Errorf("probes counted = %d, want 0", got)
	}
}

func TestMetricsHTTPProbeAnswered(t *testing.T) {
	counter := &probeCounter{}
	responder := NewResponder(Config{
		EndpointRef: "urn:uuid:resp-metrics-3",
		Port:        9090,
		Metrics:     counter,
	})

	ts := httptest.NewServer(responder)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(string(wsdiscovery.BuildProbe("http-probe-1"))))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if got := counter.count(); got != 1 {
		t.Errorf("probes counted = %d, want 1", got)
	}
}
