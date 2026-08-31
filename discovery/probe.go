package discovery

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/wsdiscovery"
)

// DefaultProbePorts are the ports commonly serving ONVIF device_service.
//
// It is a function (not a package var) so callers always get a fresh
// slice — appending to a shared package-level var would mutate the scan
// set process-wide.
func DefaultProbePorts() []int {
	return []int{80, 8080, 8000}
}

// DeviceInfo carries the identity fields an unauthenticated
// GetDeviceInformation probe can extract.
type DeviceInfo struct {
	Manufacturer    string
	Model           string
	FirmwareVersion string
	HardwareID      string
	SerialNumber    string
}

// deviceServiceURL builds the conventional ONVIF device service URL.
func deviceServiceURL(host string, port int) string {
	return fmt.Sprintf("http://%s/onvif/device_service", net.JoinHostPort(host, strconv.Itoa(port)))
}

// postSOAP posts a raw SOAP envelope and returns the response body. Status
// errors are returned for caller-classified handling (401/405 usually mean
// "needs auth" or "not ONVIF", both non-fatal for probing).
func postSOAP(ctx context.Context, client *http.Client, url, envelope string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(envelope))
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxProbeResponseBytes))
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return body, resp.StatusCode, nil
}

// maxProbeResponseBytes caps how much of a probe response is read; probes hit
// arbitrary hosts and a non-ONVIF server may stream forever.
const maxProbeResponseBytes = 1 << 20 // 1 MiB

// probeViaWSDiscovery POSTs a WS-Discovery Probe to the device service URL
// over HTTP. Some devices answer it even when they ignore UDP probes.
func probeViaWSDiscovery(ctx context.Context, client *http.Client, url string) (*ProbeMatch, error) {
	body, status, err := postSOAP(ctx, client, url,
		string(wsdiscovery.BuildProbe(generateUUID())))
	if err != nil || status != http.StatusOK {
		return nil, fmt.Errorf("probe via HTTP failed (status %d): %w", status, err)
	}

	matches, err := wsdiscovery.ParseProbeMatches(body)
	if err != nil {
		if errors.Is(err, wsdiscovery.ErrNoMatches) {
			return nil, fmt.Errorf("%w (HTTP form)", ErrNoProbeMatches)
		}

		return nil, err
	}

	return &matches[0], nil
}

// getDeviceInformationRequest is the most widely accepted unauthenticated
// ONVIF probe: minimal devices that reject WS-Discovery still answer it.
const getDeviceInformationRequest = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
<s:Body>
<tds:GetDeviceInformation xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
</s:Body>
</s:Envelope>`

// fetchDeviceInfo sends an unauthenticated GetDeviceInformation to url.
func fetchDeviceInfo(ctx context.Context, client *http.Client, url string) (*DeviceInfo, error) {
	body, status, err := postSOAP(ctx, client, url, getDeviceInformationRequest)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("GetDeviceInformation returned status %d", status)
	}

	var envelope struct {
		Body struct {
			GetDeviceInformationResponse struct {
				Manufacturer    string `xml:"Manufacturer"`
				Model           string `xml:"Model"`
				FirmwareVersion string `xml:"FirmwareVersion"`
				SerialNumber    string `xml:"SerialNumber"`
				HardwareID      string `xml:"HardwareId"`
			} `xml:"GetDeviceInformationResponse"`
			Fault struct {
				Code   string `xml:"Code>Value"`
				Reason string `xml:"Reason>Text"`
			} `xml:"Fault"`
		} `xml:"Body"`
	}

	if err := xml.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GetDeviceInformation response: %w", err)
	}

	resp := envelope.Body.GetDeviceInformationResponse
	if resp.Manufacturer == "" && resp.SerialNumber == "" && resp.Model == "" &&
		envelope.Body.Fault.Code != "" {
		return nil, fmt.Errorf("device faulted on GetDeviceInformation: %s %s",
			envelope.Body.Fault.Code, envelope.Body.Fault.Reason)
	}

	return &DeviceInfo{
		Manufacturer:    resp.Manufacturer,
		Model:           resp.Model,
		FirmwareVersion: resp.FirmwareVersion,
		SerialNumber:    resp.SerialNumber,
		HardwareID:      resp.HardwareID,
	}, nil
}

// serialNumberPattern matches a SerialNumber element regardless of the
// namespace prefix a particular firmware picked.
var serialNumberPattern = regexp.MustCompile(`(?i)<[A-Za-z0-9_.-]*:?SerialNumber[^>]*>\s*([^<]+?)\s*</`)

// extractSerialNumber pulls the serial out of a raw GetDeviceInformation
// response without any namespace assumptions.
func extractSerialNumber(body []byte) (string, bool) {
	m := serialNumberPattern.FindSubmatch(body)
	if m == nil {
		return "", false
	}

	return string(m[1]), true
}

// probeHTTPClient builds the probing HTTP client: short timeout, no
// redirects (a stray router UI must not turn a probe into a login page hit).
func probeHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// ProbeEndpoint directs an ONVIF probe at one known address over pure HTTP —
// no UDP, no multicast — making it work across subnets and for devices that
// never answer WS-Discovery probes. Two strategies are tried in order:
//
//  1. a WS-Discovery Probe envelope POSTed to http://host:port/onvif/device_service;
//  2. an unauthenticated GetDeviceInformation (the most widely accepted probe).
//
// A nil return (with nil error) means "this address does not speak ONVIF or
// is offline" — probing cannot distinguish the two, and both are simply
// "not found". Malformed responses are contained: a garbage answer yields
// nil, never a panic.
func ProbeEndpoint(ctx context.Context, host string, port int, timeout time.Duration) (device *Device) {
	if host == "" || port <= 0 || port > 65535 {
		return nil
	}

	if timeout <= 0 {
		timeout = defaultProbeTimeout
	}

	// Malformed responses from arbitrary hosts must not panic the caller.
	defer func() {
		if recover() != nil {
			device = nil
		}
	}()

	client := probeHTTPClient(timeout)
	url := deviceServiceURL(host, port)

	// Strategy 1: WS-Discovery Probe over HTTP.
	if match, err := probeViaWSDiscovery(ctx, client, url); err == nil {
		dev := &Device{
			EndpointRef:     match.EndpointRef,
			XAddrs:          parseSpaceSeparated(match.XAddrs),
			Types:           parseSpaceSeparated(match.Types),
			Scopes:          parseSpaceSeparated(match.Scopes),
			MetadataVersion: match.MetadataVersion,
		}

		if len(dev.XAddrs) == 0 {
			dev.XAddrs = []string{url}
		}

		return dev
	}

	// Strategy 2: unauthenticated GetDeviceInformation.
	info, err := fetchDeviceInfo(ctx, client, url)
	if err != nil || (info.Manufacturer == "" && info.SerialNumber == "" && info.Model == "") {
		return nil
	}

	return &Device{
		XAddrs: []string{url},
		Info:   info,
	}
}

// defaultProbeTimeout matches the short single-shot budget a rediscovery
// engine needs (NVRs use ~1.2s per attempt).
const defaultProbeTimeout = 1200 * time.Millisecond

// ProbeSerial scans the given ports (DefaultProbePorts when empty) for an
// ONVIF device and returns its serial number, extracted without any
// namespace assumptions — the serial is stable across protocols, which makes
// it the identity anchor for correlating the same physical camera exposed
// through both GB28181 and ONVIF. ok is false when no port yields a serial.
func ProbeSerial(ctx context.Context, host string, ports []int) (serial string, ok bool) {
	if len(ports) == 0 {
		ports = DefaultProbePorts()
	}

	if host == "" {
		return "", false
	}

	defer func() {
		if recover() != nil {
			serial, ok = "", false
		}
	}()

	client := probeHTTPClient(defaultProbeTimeout)

	for _, port := range ports {
		if err := ctx.Err(); err != nil {
			return "", false
		}

		body, status, err := postSOAP(ctx, client, deviceServiceURL(host, port), getDeviceInformationRequest)
		if err != nil || status != http.StatusOK {
			continue
		}

		if s, found := extractSerialNumber(body); found {
			return s, true
		}
	}

	return "", false
}
