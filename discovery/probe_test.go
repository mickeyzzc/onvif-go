package discovery

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/wsdiscovery"
)

const probeMatchesBody = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<wsd:ProbeMatches xmlns:wsd="http://schemas.xmlsoap.org/ws/2005/04/discovery" xmlns:dp0="http://www.onvif.org/ver10/network/wsdl">
<wsd:ProbeMatch>
<wsa:EndpointReference xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing">
<wsa:Address>urn:uuid:2108cbb2-8591-4b4a-a316-a5d5abf1e2ee</wsa:Address></wsa:EndpointReference>
<wsd:Types>dp0:NetworkVideoTransmitter</wsd:Types>
<wsd:Scopes>onvif://www.onvif.org/name/ProbeCam onvif://www.onvif.org/hardware/PWR-01</wsd:Scopes>
<wsd:XAddrs>http://192.168.9.9:80/onvif/device_service</wsd:XAddrs>
<wsd:MetadataVersion>1</wsd:MetadataVersion>
</wsd:ProbeMatch>
</wsd:ProbeMatches>
</s:Body></s:Envelope>`

const gdiBody = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<tds:GetDeviceInformationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:Manufacturer>ProbeCam</tds:Manufacturer>
<tds:Model>PC-1</tds:Model>
<tds:FirmwareVersion>2.1</tds:FirmwareVersion>
<tds:SerialNumber>SN-123456</tds:SerialNumber>
</tds:GetDeviceInformationResponse>
</s:Body></s:Envelope>`

func TestProbeEndpointViaWSDiscoveryOverHTTP(t *testing.T) {
	var sawProbe bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		sawProbe = len(body) > 0 && httpProbeEnvelope(string(body))
		w.Write([]byte(probeMatchesBody))
	}))
	t.Cleanup(server.Close)

	host, port := hostPort(t, server.URL)

	dev := ProbeEndpoint(context.Background(), host, port, 2*time.Second)
	if dev == nil {
		t.Fatal("ProbeEndpoint() = nil, want device")
	}

	if !sawProbe {
		t.Error("server did not see a Probe envelope")
	}

	if dev.EndpointRef != "urn:uuid:2108cbb2-8591-4b4a-a316-a5d5abf1e2ee" {
		t.Errorf("EndpointRef = %q", dev.EndpointRef)
	}

	if len(dev.XAddrs) != 1 || dev.XAddrs[0] != "http://192.168.9.9:80/onvif/device_service" {
		t.Errorf("XAddrs = %v", dev.XAddrs)
	}
}

func TestProbeEndpointFallsBackToGetDeviceInformation(t *testing.T) {
	// Rejects the WS-Discovery probe with 405 but answers GDI.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if httpProbeEnvelope(string(body)) {
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		w.Write([]byte(gdiBody))
	}))
	t.Cleanup(server.Close)

	host, port := hostPort(t, server.URL)

	dev := ProbeEndpoint(context.Background(), host, port, 2*time.Second)
	if dev == nil {
		t.Fatal("ProbeEndpoint() = nil, want device via GDI fallback")
	}

	if dev.Info == nil || dev.Info.Manufacturer != "ProbeCam" || dev.Info.SerialNumber != "SN-123456" {
		t.Errorf("Info = %+v, want ProbeCam/SN-123456", dev.Info)
	}

	if len(dev.XAddrs) != 1 {
		t.Errorf("XAddrs = %v, want the probed URL", dev.XAddrs)
	}
}

func TestProbeEndpointGDIRequiresAuth(t *testing.T) {
	// 401 on everything: authenticated devices are "not found" for probing.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	host, port := hostPort(t, server.URL)

	if dev := ProbeEndpoint(context.Background(), host, port, 2*time.Second); dev != nil {
		t.Errorf("ProbeEndpoint() = %+v, want nil for a 401 device", dev)
	}
}

func TestProbeEndpointMalformedXML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("<this is not xml"))
	}))
	t.Cleanup(server.Close)

	host, port := hostPort(t, server.URL)

	if dev := ProbeEndpoint(context.Background(), host, port, 2*time.Second); dev != nil {
		t.Errorf("ProbeEndpoint() = %+v, want nil for garbage responses", dev)
	}
}

func TestProbeEndpointFaultOnGDI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<s:Fault><s:Code><s:Value>s:Sender</s:Value></s:Code>
<s:Reason><s:Text>Not Authorized</s:Text></s:Reason></s:Fault>
</s:Body></s:Envelope>`))
	}))
	t.Cleanup(server.Close)

	host, port := hostPort(t, server.URL)

	if dev := ProbeEndpoint(context.Background(), host, port, 2*time.Second); dev != nil {
		t.Errorf("ProbeEndpoint() = %+v, want nil for a faulting device", dev)
	}
}

func TestProbeEndpointTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.Write([]byte(gdiBody))
	}))
	t.Cleanup(server.Close)

	host, port := hostPort(t, server.URL)

	start := time.Now()
	dev := ProbeEndpoint(context.Background(), host, port, 50*time.Millisecond)
	if dev != nil {
		t.Errorf("ProbeEndpoint() = %+v, want nil on timeout", dev)
	}

	if elapsed := time.Since(start); elapsed > 450*time.Millisecond {
		t.Errorf("probe took %v despite 50ms per-attempt budget", elapsed)
	}
}

func TestProbeEndpointInvalidArgs(t *testing.T) {
	if dev := ProbeEndpoint(context.Background(), "", 80, time.Second); dev != nil {
		t.Error("empty host accepted")
	}

	if dev := ProbeEndpoint(context.Background(), "1.2.3.4", 0, time.Second); dev != nil {
		t.Error("port 0 accepted")
	}

	if dev := ProbeEndpoint(context.Background(), "1.2.3.4", 70000, time.Second); dev != nil {
		t.Error("port 70000 accepted")
	}
}

func TestProbeSerialFindsSerialAcrossNamespaces(t *testing.T) {
	// Nonstandard namespace prefix — the regex extractor must not care.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<vendor:GetDeviceInformationResponse xmlns:vendor="http://vendor.example/onvif">
<vendor:Manufacturer>Vendor</vendor:Manufacturer>
<vendor:SerialNumber>  GB-SER-99  </vendor:SerialNumber>
</vendor:GetDeviceInformationResponse>
</s:Body></s:Envelope>`))
	}))
	t.Cleanup(server.Close)

	host, port := hostPort(t, server.URL)

	serial, ok := ProbeSerial(context.Background(), host, []int{port})
	if !ok {
		t.Fatal("ProbeSerial() ok = false, want serial")
	}

	if serial != "GB-SER-99" {
		t.Errorf("serial = %q, want GB-SER-99 (trimmed)", serial)
	}
}

func TestProbeSerialScansPorts(t *testing.T) {
	// First port is a non-ONVIF server; second carries the camera.
	nonONVIF := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(nonONVIF.Close)

	camera := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(gdiBody))
	}))
	t.Cleanup(camera.Close)

	_, deadPort := hostPort(t, nonONVIF.URL)
	camHost, camPort := hostPort(t, camera.URL)

	serial, ok := ProbeSerial(context.Background(), camHost, []int{deadPort, camPort})
	if !ok {
		t.Fatal("ProbeSerial() ok = false, want serial from second port")
	}

	if serial != "SN-123456" {
		t.Errorf("serial = %q, want SN-123456", serial)
	}
}

func TestProbeSerialNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("hello world"))
	}))
	t.Cleanup(server.Close)

	host, port := hostPort(t, server.URL)

	serial, ok := ProbeSerial(context.Background(), host, []int{port})
	if ok || serial != "" {
		t.Errorf("ProbeSerial() = (%q, %v), want empty/not-found", serial, ok)
	}
}

func TestProbeSerialContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	serial, ok := ProbeSerial(ctx, "192.0.2.1", DefaultProbePorts())
	if ok || serial != "" {
		t.Errorf("ProbeSerial(canceled ctx) = (%q, %v), want empty/not-found", serial, ok)
	}
}

// hostPort splits an httptest URL.
func hostPort(t *testing.T, rawURL string) (string, int) {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("bad test URL %q: %v", rawURL, err)
	}

	host, portStr, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		t.Fatalf("bad test host %q: %v", parsed.Host, err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("bad port %q: %v", portStr, err)
	}

	return host, port
}

// httpProbeEnvelope reports whether the body carries a WS-Discovery Probe.
func httpProbeEnvelope(body string) bool {
	return strings.Contains(body, ":Probe") || strings.Contains(body, "<Probe ")
}

func TestParseProbeResponseDirect(t *testing.T) {
	answer := wsdiscovery.BuildProbeMatches("probe-1", wsdiscovery.Match{
		EndpointRef:     "urn:uuid:dev-1",
		Types:           "tds:Device dp0:NetworkVideoTransmitter",
		Scopes:          "onvif://www.onvif.org/name/TestCam onvif://www.onvif.org/location/Roof",
		XAddrs:          "http://192.0.2.9/onvif/device_service",
		MetadataVersion: 1,
	})

	device, err := parseProbeResponse(answer)
	if err != nil {
		t.Fatal(err)
	}

	if device.EndpointRef != "urn:uuid:dev-1" {
		t.Errorf("EndpointRef = %q", device.EndpointRef)
	}

	if len(device.XAddrs) != 1 || device.XAddrs[0] != "http://192.0.2.9/onvif/device_service" {
		t.Errorf("XAddrs = %v", device.XAddrs)
	}

	if device.Name != "TestCam" || device.Location != "Roof" {
		t.Errorf("scope fields = %q/%q", device.Name, device.Location)
	}

	if _, err := parseProbeResponse([]byte("not a probe matches")); err == nil {
		t.Error("garbage accepted")
	}
}
