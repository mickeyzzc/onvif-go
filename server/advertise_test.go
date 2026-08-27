package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// postSOAP drives a service endpoint registered on a fresh mux.
func postSOAP(t *testing.T, s *Server, endpoint, innerXML, remoteAddr string) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	s.registerDeviceService(mux)
	s.registerMediaService(mux)

	req := httptest.NewRequest(http.MethodPost, s.config.BasePath+endpoint, strings.NewReader(
		`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>`+innerXML+
			`</s:Body></s:Envelope>`))
	req.RemoteAddr = remoteAddr

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("%s status = %d, want 200; body: %s", endpoint, w.Code, w.Body.String())
	}

	return w
}

func TestXAddrEchoesClientIP(t *testing.T) {
	config := createTestConfig()
	s, _ := New(config)

	// GetCapabilities: every XAddr must carry the requesting client's IP.
	w := postSOAP(t, s, "/device_service",
		`<GetCapabilities xmlns="http://www.onvif.org/ver10/device/wsdl"/>`,
		"198.51.100.23:41000")

	body := w.Body.String()
	for _, service := range []string{"device_service", "media_service", "ptz_service", "imaging_service"} {
		if !strings.Contains(body, "http://198.51.100.23:") {
			t.Fatalf("XAddr does not echo client IP; body: %s", body)
		}

		if !strings.Contains(body, service) {
			t.Errorf("capabilities missing %s", service)
		}
	}

	// GetServices: same rule.
	w = postSOAP(t, s, "/device_service",
		`<GetServices xmlns="http://www.onvif.org/ver10/device/wsdl"><IncludeCapability>true</IncludeCapability></GetServices>`,
		"198.51.100.24:41001")

	if !strings.Contains(w.Body.String(), "http://198.51.100.24:") {
		t.Fatalf("GetServices XAddr does not echo client IP; body: %s", w.Body.String())
	}

	// Different peer gets its own address back.
	w = postSOAP(t, s, "/device_service",
		`<GetCapabilities xmlns="http://www.onvif.org/ver10/device/wsdl"/>`,
		"203.0.113.7:41002")

	if !strings.Contains(w.Body.String(), "http://203.0.113.7:") {
		t.Fatalf("XAddr does not follow the second client IP; body: %s", w.Body.String())
	}
}

func TestStreamURIAndSnapshotEchoClientIP(t *testing.T) {
	config := createTestConfig()
	s, _ := New(config)

	token := config.Profiles[0].Token

	w := postSOAP(t, s, "/media_service",
		`<GetStreamUri xmlns="http://www.onvif.org/ver10/media/wsdl"><ProfileToken>`+token+`</ProfileToken></GetStreamUri>`,
		"198.51.100.30:41010")

	if !strings.Contains(w.Body.String(), "rtsp://198.51.100.30:8554") {
		t.Fatalf("stream URI does not echo client IP; body: %s", w.Body.String())
	}

	w = postSOAP(t, s, "/media_service",
		`<GetSnapshotUri xmlns="http://www.onvif.org/ver10/media/wsdl"><ProfileToken>`+token+`</ProfileToken></GetSnapshotUri>`,
		"198.51.100.31:41011")

	if !strings.Contains(w.Body.String(), "http://198.51.100.31:") {
		t.Fatalf("snapshot URI does not echo client IP; body: %s", w.Body.String())
	}
}

func TestLegacyStreamURISpellingDispatches(t *testing.T) {
	config := createTestConfig()
	s, _ := New(config)

	token := config.Profiles[0].Token

	// The legacy GetStreamURI spelling must reach the canonical handler —
	// real-world clients send both forms.
	w := postSOAP(t, s, "/media_service",
		`<GetStreamURI xmlns="http://www.onvif.org/ver10/media/wsdl"><ProfileToken>`+token+`</ProfileToken></GetStreamURI>`,
		"198.51.100.32:41012")

	if !strings.Contains(w.Body.String(), "GetStreamUriResponse") {
		t.Fatalf("legacy spelling not dispatched or wrong response element; body: %s", w.Body.String())
	}
}

func TestAdvertiseHostOverride(t *testing.T) {
	config := createTestConfig()
	config.AdvertiseHost = "camera.example.org"
	s, _ := New(config)

	w := postSOAP(t, s, "/device_service",
		`<GetCapabilities xmlns="http://www.onvif.org/ver10/device/wsdl"/>`,
		"198.51.100.40:41020")

	if !strings.Contains(w.Body.String(), "http://camera.example.org:") {
		t.Fatalf("AdvertiseHost override not honored; body: %s", w.Body.String())
	}

	if strings.Contains(w.Body.String(), "http://198.51.100.40:") {
		t.Fatalf("client IP leaked despite AdvertiseHost override; body: %s", w.Body.String())
	}
}

func TestExplicitPrefixesThroughServer(t *testing.T) {
	config := createTestConfig()
	config.ExplicitPrefixes = true
	s, _ := New(config)

	token := config.Profiles[0].Token

	w := postSOAP(t, s, "/media_service",
		`<GetStreamUri xmlns="http://www.onvif.org/ver10/media/wsdl"><ProfileToken>`+token+`</ProfileToken></GetStreamUri>`,
		"198.51.100.50:41030")

	body := w.Body.String()
	for _, want := range []string{
		"<s:Envelope xmlns:s=\"http://www.w3.org/2003/05/soap-envelope\">",
		"<s:Body>",
		"<trt:GetStreamUriResponse xmlns:trt=\"http://www.onvif.org/ver10/media/wsdl\">",
		"<trt:MediaUri>",
		"<trt:Uri>rtsp://198.51.100.50:8554",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("prefixed response missing %q; body: %s", want, body)
		}
	}
}
