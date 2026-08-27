package onvif

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestGetStreamURIWithOptionsRequestShape asserts the StreamSetup fields the
// device receives for every supported stream × protocol combination. Devices
// such as ESP32 firmwares branch on the requested Protocol, so the wire
// content is part of the contract (issue #2).
func TestGetStreamURIWithOptionsRequestShape(t *testing.T) {
	combos := []struct {
		stream   string
		protocol string
	}{
		{StreamRTPUnicast, ProtocolRTSP},
		{StreamRTPUnicast, ProtocolHTTP},
		{StreamRTPUnicast, ProtocolUDP},
		{StreamRTPMulticast, ProtocolRTSP},
		{StreamRTPMulticast, ProtocolHTTP},
		{StreamRTPMulticast, ProtocolUDP},
	}

	for _, combo := range combos {
		name := combo.stream + "/" + combo.protocol
		t.Run(name, func(t *testing.T) {
			var lastRequest string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				lastRequest = string(body)
				w.Header().Set("Content-Type", "application/soap+xml")
				_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<trt:GetStreamUriResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema">
<trt:MediaUri><tt:Uri>rtsp://device/stream</tt:Uri></trt:MediaUri>
</trt:GetStreamUriResponse></s:Body></s:Envelope>`))
			}))
			t.Cleanup(server.Close)

			client, err := NewClient(server.URL, WithHTTPClient(server.Client()))
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			uri, err := client.Media().GetStreamURIWithOptions(
				context.Background(), "profile-1",
				StreamSetup{Stream: combo.stream, Transport: &Transport{Protocol: combo.protocol}})
			if err != nil {
				t.Fatalf("GetStreamURIWithOptions() error = %v", err)
			}

			if uri.URI != "rtsp://device/stream" {
				t.Errorf("URI = %q, want rtsp://device/stream", uri.URI)
			}

			if !strings.Contains(lastRequest, "<tt:Stream>"+combo.stream+"</tt:Stream>") {
				t.Errorf("request missing <tt:Stream>%s</tt:Stream>: %s", combo.stream, lastRequest)
			}

			if !strings.Contains(lastRequest, "<tt:Protocol>"+combo.protocol+"</tt:Protocol>") {
				t.Errorf("request missing <tt:Protocol>%s</tt:Protocol>: %s", combo.protocol, lastRequest)
			}
		})
	}
}

// TestGetStreamURIDefaultStillUnicastRTSP pins the legacy GetStreamURI wire
// behavior (issue #2: zero breakage for the existing entry point).
func TestGetStreamURIDefaultStillUnicastRTSP(t *testing.T) {
	var lastRequest string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		lastRequest = string(body)
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<trt:GetStreamUriResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema">
<trt:MediaUri><tt:Uri>rtsp://device/main</tt:Uri></trt:MediaUri>
</trt:GetStreamUriResponse></s:Body></s:Envelope>`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if _, err := client.Media().GetStreamURI(context.Background(), "profile-1"); err != nil {
		t.Fatalf("GetStreamURI() error = %v", err)
	}

	if !strings.Contains(lastRequest, "<tt:Stream>RTP-Unicast</tt:Stream>") ||
		!strings.Contains(lastRequest, "<tt:Protocol>RTSP</tt:Protocol>") {
		t.Errorf("default GetStreamURI request changed shape: %s", lastRequest)
	}
}

func TestGetStreamURIWithOptionsRejectsInvalidSetup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("device must not be contacted for invalid setups")
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	invalid := []StreamSetup{
		{Stream: "RTP-CarrierPigeon", Transport: &Transport{Protocol: ProtocolRTSP}},
		{Stream: "", Transport: &Transport{Protocol: ProtocolRTSP}},
		{Stream: StreamRTPUnicast, Transport: &Transport{Protocol: "FTP"}},
	}

	for _, setup := range invalid {
		if _, err := client.Media().GetStreamURIWithOptions(context.Background(), "p", setup); err == nil {
			t.Errorf("GetStreamURIWithOptions(%+v) succeeded, want validation error", setup)
		}
	}
}

// TestGetStreamURIWithOptionsDefaultsToRTSP verifies that a nil Transport
// (or an empty protocol) defaults to RTSP instead of erroring.
func TestGetStreamURIWithOptionsDefaultsToRTSP(t *testing.T) {
	for _, setup := range []StreamSetup{
		{Stream: StreamRTPUnicast},
		{Stream: StreamRTPUnicast, Transport: &Transport{}},
	} {
		if err := setup.validate(); err != nil {
			t.Fatalf("validate(%+v) error = %v, want RTSP default", setup, err)
		}

		if setup.Transport.Protocol != ProtocolRTSP {
			t.Errorf("validate(%+v) left protocol %q, want %q", setup, setup.Transport.Protocol, ProtocolRTSP)
		}
	}
}
