package discovery

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	clientdiscovery "github.com/mickeyzzc/onvif-go/v2/discovery"
	"github.com/mickeyzzc/onvif-go/v2/wsdiscovery"
)

// captureSend records outgoing datagrams for deterministic assertions.
type captureSend struct {
	replies [][]byte
	targets []*net.UDPAddr
}

func (c *captureSend) send(data []byte, to *net.UDPAddr) {
	c.replies = append(c.replies, data)
	c.targets = append(c.targets, to)
}

func TestResponderAnswersMulticastProbe(t *testing.T) {
	responder := NewResponder(Config{
		EndpointRef: "urn:uuid:resp-1",
		Types:       []string{"tds:Device", "dp0:NetworkVideoTransmitter"},
		Scopes:      []string{"onvif://www.onvif.org/name/TestCam", "onvif://www.onvif.org/location/Roof"},
		Port:        9090,
	})

	probe := wsdiscovery.BuildProbe("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	src := &net.UDPAddr{IP: net.ParseIP("198.51.100.7"), Port: 52123}

	capture := &captureSend{}
	responder.handleDatagram(t.Context(), probe, src, capture.send)

	if len(capture.replies) != 1 {
		t.Fatalf("replies = %d, want 1", len(capture.replies))
	}

	if capture.targets[0].String() != src.String() {
		t.Errorf("reply sent to %v, want unicast to prober %v", capture.targets[0], src)
	}

	// The client's own parser must understand the answer.
	devices, err := parseClientDevices(capture.replies[0])
	if err != nil {
		t.Fatalf("client-side parse: %v", err)
	}

	device := devices[0]
	if device.EndpointRef != "urn:uuid:resp-1" {
		t.Errorf("EndpointRef = %q", device.EndpointRef)
	}

	// #38: derived XAddrs must point at THIS device (a local interface
	// address), never echo the requester — an echoing XAddrs makes NVR
	// consumers register themselves as the camera.
	if len(device.XAddrs) != 1 {
		t.Fatalf("XAddrs = %v, want exactly one", device.XAddrs)
	}

	xaddrHost := device.XAddrs[0]
	if strings.Contains(xaddrHost, "198.51.100.7") {
		t.Errorf("XAddrs = %q echoes the requester IP", xaddrHost)
	}

	if !isLocalAddress(t, xaddrHost) {
		t.Errorf("XAddrs = %q does not point at a local device address", xaddrHost)
	}

	if !strings.Contains(xaddrHost, ":9090/onvif/device_service") {
		t.Errorf("XAddrs = %q missing configured port/path", xaddrHost)
	}

	if device.Name != "TestCam" || device.Location != "Roof" {
		t.Errorf("scopes not surfaced: name=%q location=%q", device.Name, device.Location)
	}
}

// TestDerivedXAddrsNeverEchoRequester pins #38 at the derivation layer:
// for any peer, an unset XAddrs config derives the device's own source
// address toward that peer (loopback when the peer is loopback) — the
// requester's address never appears.
func TestDerivedXAddrsNeverEchoRequester(t *testing.T) {
	responder := NewResponder(Config{Port: 8080, DevicePath: "/onvif/device_service"})

	for _, peer := range []string{"198.51.100.7", "127.0.0.1"} {
		match := responder.matchFor(t.Context(), peer)
		if match.XAddrs == "" {
			t.Fatalf("peer %s: empty derived XAddrs", peer)
		}

		if strings.Contains(match.XAddrs, peer) && peer != "127.0.0.1" {
			t.Errorf("peer %s: XAddrs %q echoes the requester", peer, match.XAddrs)
		}

		if !isLocalAddress(t, match.XAddrs) {
			t.Errorf("peer %s: XAddrs %q is not a local device address", peer, match.XAddrs)
		}
	}
}

// isLocalAddress reports whether the host part of an http://… XAddr (or
// host:port token) belongs to one of this host's interface addresses.
func isLocalAddress(t *testing.T, xaddr string) bool {
	t.Helper()

	u := strings.TrimPrefix(xaddr, "http://")
	host := u

	if h, _, err := net.SplitHostPort(strings.TrimSuffix(u, "/onvif/device_service")); err == nil {
		host = h
	} else if i := strings.IndexByte(u, '/'); i >= 0 {
		host = u[:i]
	} else if i := strings.LastIndexByte(u, ':'); i >= 0 {
		host = u[:i]
	}

	ip := net.ParseIP(host)
	if ip == nil {
		t.Fatalf("cannot parse host out of %q", xaddr)
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Fatalf("list interface addrs: %v", err)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && ipnet.IP.Equal(ip) {
			return true
		}
	}

	return false
}

func TestResponderIgnoresNonProbesAndTypeFilters(t *testing.T) {
	responder := NewResponder(Config{EndpointRef: "urn:uuid:resp-2", Port: 8080})
	src := &net.UDPAddr{IP: net.IPv4(10, 0, 0, 1), Port: 1234}

	capture := &captureSend{}

	// Hello (someone else's announcement) — ignored.
	responder.handleDatagram(t.Context(), wsdiscovery.BuildHello(wsdiscovery.Match{EndpointRef: "urn:uuid:x"}), src, capture.send)

	// Probe for printers only — filtered out.
	printerProbe := strings.Replace(string(wsdiscovery.BuildProbe("p1")),
		"dp0:NetworkVideoTransmitter", "dn:NetworkPrinter", 1)
	responder.handleDatagram(t.Context(), []byte(printerProbe), src, capture.send)

	if len(capture.replies) != 0 {
		t.Errorf("replies = %d, want 0 (ignored/filtered)", len(capture.replies))
	}
}

func TestResponderStaticXAddrs(t *testing.T) {
	responder := NewResponder(Config{
		EndpointRef: "urn:uuid:resp-3",
		XAddrs:      []string{"http://camera.example.org/onvif/device_service"},
	})

	src := &net.UDPAddr{IP: net.IPv4(10, 1, 2, 3), Port: 9}
	capture := &captureSend{}
	responder.handleDatagram(t.Context(), wsdiscovery.BuildProbe("fixed-id"), src, capture.send)

	devices, err := parseClientDevices(capture.replies[0])
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if devices[0].XAddrs[0] != "http://camera.example.org/onvif/device_service" {
		t.Errorf("static XAddrs not honored: %v", devices[0].XAddrs)
	}
}

func TestResponderHTTPProbe(t *testing.T) {
	responder := NewResponder(Config{
		EndpointRef: "urn:uuid:resp-http",
		Scopes:      []string{"onvif://www.onvif.org/name/HTTPCam"},
		Port:        9090,
	})

	ts := httptest.NewServer(responder)
	defer ts.Close()

	// Client-side directed probe (Probe-over-HTTP) must discover us.
	req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(string(wsdiscovery.BuildProbe("http-probe-1"))))
	if err != nil {
		t.Fatal(err)
	}

	req.RemoteAddr = "" // httptest sets it; ensure default handling
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("directed probe failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/soap+xml") {
		t.Errorf("Content-Type = %q", ct)
	}

	body := make([]byte, 65536)
	n, _ := resp.Body.Read(body)

	devices, err := parseClientDevices(body[:n])
	if err != nil {
		t.Fatalf("client parse of HTTP answer: %v", err)
	}

	if devices[0].EndpointRef != "urn:uuid:resp-http" {
		t.Errorf("EndpointRef = %q", devices[0].EndpointRef)
	}
}

func TestResponderHTTPRejectsGarbage(t *testing.T) {
	responder := NewResponder(Config{EndpointRef: "urn:uuid:resp-http-2"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("this is not soap"))
	req.RemoteAddr = "192.0.2.9:5555"
	responder.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("garbage POST status = %d, want 400", w.Code)
	}

	w2 := httptest.NewRecorder()
	responder.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/", nil))
	if w2.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET status = %d, want 405", w2.Code)
	}
}

// TestResponderDiscoveredByClient runs the full bootstrap loop when the
// host supports multicast: discovery.Discover must find the responder.
func TestResponderDiscoveredByClient(t *testing.T) {
	responder := NewResponder(Config{
		EndpointRef: "urn:uuid:resp-bootstrap",
		Scopes:      []string{"onvif://www.onvif.org/name/Bootstrap"},
		Port:        8888,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := responder.Start(ctx); err != nil {
		t.Skipf("host cannot join the WS-Discovery multicast group: %v", err)
	}
	defer responder.Stop()

	// Give the responder's socket a moment to come up.
	time.Sleep(200 * time.Millisecond)

	devices, err := clientdiscovery.Discover(ctx, 2*time.Second)
	if err != nil {
		t.Skipf("multicast discovery unavailable on this host: %v", err)
	}

	for _, device := range devices {
		if device.EndpointRef == "urn:uuid:resp-bootstrap" {
			return // found ourselves — bootstrap proven
		}
	}

	t.Skip("bootstrap: responder not seen (multicast delivery unavailable on this host/network)")
}

// parseClientDevices routes an answer through the client package's own
// parser — the interop guarantee.
func parseClientDevices(data []byte) ([]*clientdiscovery.Device, error) {
	var wrapper struct {
		Devices []*clientdiscovery.Device
	}
	_ = wrapper

	matches, err := wsdiscovery.ParseProbeMatches(data)
	if err != nil {
		return nil, err
	}

	devices := make([]*clientdiscovery.Device, 0, len(matches))
	for i := range matches {
		m := matches[i]
		device := &clientdiscovery.Device{
			EndpointRef:     m.EndpointRef,
			XAddrs:          splitSpaces(m.XAddrs),
			Types:           splitSpaces(m.Types),
			Scopes:          splitSpaces(m.Scopes),
			MetadataVersion: m.MetadataVersion,
		}
		device.Name, device.Hardware, device.Location = scopeFields(device.Scopes)
		devices = append(devices, device)
	}

	return devices, nil
}

func splitSpaces(s string) []string {
	fields := strings.Fields(strings.TrimSpace(s))
	if fields == nil {
		return []string{}
	}

	return fields
}

// scopeFields mirrors the client's scope extraction (name/hardware/
// location) without importing unexported helpers.
func scopeFields(scopes []string) (name, hardware, location string) {
	for _, scope := range scopes {
		switch {
		case strings.Contains(scope, "/name/"):
			name = lastSegment(scope)
		case strings.Contains(scope, "/hardware/"):
			hardware = lastSegment(scope)
		case strings.Contains(scope, "/location/"):
			location = lastSegment(scope)
		}
	}

	return name, hardware, location
}

func lastSegment(s string) string {
	parts := strings.Split(s, "/")

	return parts[len(parts)-1]
}
