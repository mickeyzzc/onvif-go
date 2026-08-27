package discovery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseScopes(t *testing.T) {
	tests := []struct {
		name   string
		scopes []string
		want   ScopeInfo
	}{
		{
			name: "all keys",
			scopes: []string{
				"onvif://www.onvif.org/name/Corridor%20Cam",
				"onvif://www.onvif.org/hardware/DS-2CD2143",
				"onvif://www.onvif.org/location/Floor%2F3",
			},
			want: ScopeInfo{Name: "Corridor Cam", Hardware: "DS-2CD2143", Location: "Floor/3"},
		},
		{
			name:   "later entry wins",
			scopes: []string{"onvif://www.onvif.org/name/Old", "onvif://www.onvif.org/name/New"},
			want:   ScopeInfo{Name: "New"},
		},
		{
			name:   "non-onvif scopes ignored",
			scopes: []string{"http://vendor.example/scope/1", "onvif://www.onvif.org/name/Cam", "global"},
			want:   ScopeInfo{Name: "Cam"},
		},
		{
			name:   "unknown keys ignored",
			scopes: []string{"onvif://www.onvif.org/type/videoEncoder", "onvif://www.onvif.org/name/C"},
			want:   ScopeInfo{Name: "C"},
		},
		{
			name:   "value-less entries ignored",
			scopes: []string{"onvif://www.onvif.org/name", "onvif://www.onvif.org/"},
			want:   ScopeInfo{},
		},
		{
			name:   "empty",
			scopes: nil,
			want:   ScopeInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseScopes(tt.scopes); got != tt.want {
				t.Errorf("ParseScopes() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestFilterONVIFDevices(t *testing.T) {
	onvifByTypes := &Device{
		EndpointRef: "types-cam",
		Types:       []string{"dp0:NetworkVideoTransmitter"},
	}
	onvifByPrefix := &Device{
		EndpointRef: "prefix-cam",
		Types:       []string{"tns:NetworkVideoTransmitter"},
	}
	onvifByScopesOnly := &Device{
		EndpointRef: "scopes-cam",
		Scopes:      []string{"onvif://www.onvif.org/name/Cam"},
	}
	synology := &Device{
		EndpointRef: "synology-dsm",
		Types:       []string{"tdn:SynologyNAS"},
		Scopes:      []string{"http://www.synology.com/"},
		XAddrs:      []string{"http://192.168.1.10:5000/"},
	}
	printer := &Device{
		EndpointRef: "printer",
		Types:       []string{"pwg:PrintBasic"},
	}

	got := FilterONVIFDevices([]*Device{synology, onvifByTypes, printer, onvifByPrefix, onvifByScopesOnly})

	wantRefs := []string{"types-cam", "prefix-cam", "scopes-cam"}
	if len(got) != len(wantRefs) {
		t.Fatalf("FilterONVIFDevices() kept %d devices (%v), want %v", len(got), refs(got), wantRefs)
	}

	for i, d := range got {
		if d.EndpointRef != wantRefs[i] {
			t.Errorf("kept[%d] = %q, want %q", i, d.EndpointRef, wantRefs[i])
		}
	}
}

func refs(devices []*Device) []string {
	out := make([]string, 0, len(devices))
	for _, d := range devices {
		out = append(out, d.EndpointRef)
	}

	return out
}

func TestDeviceStructuredScopeFields(t *testing.T) {
	d := parseDiscoveryDatagram([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<Hello xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery">
<EndpointReference><Address>urn:uuid:x</Address></EndpointReference>
<Scopes>onvif://www.onvif.org/name/Garage%20Cam onvif://www.onvif.org/hardware/HW-7</Scopes>
<XAddrs>http://1.2.3.4/onvif/device_service</XAddrs>
</Hello>
</s:Body></s:Envelope>`))

	if d == nil {
		t.Fatal("parse failed")
	}

	if d.Name != "Garage Cam" || d.Hardware != "HW-7" {
		t.Errorf("structured fields = %q/%q, want Garage Cam/HW-7", d.Name, d.Hardware)
	}

	if got := d.GetName(); got != "Garage Cam" {
		t.Errorf("GetName() = %q", got)
	}
}

func TestEnrichDevices(t *testing.T) {
	camera := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<tds:GetDeviceInformationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:Manufacturer>EnrichCam</tds:Manufacturer>
<tds:Model>E-1</tds:Model>
<tds:SerialNumber>ENR-001</tds:SerialNumber>
</tds:GetDeviceInformationResponse>
</s:Body></s:Envelope>`))
	}))
	t.Cleanup(camera.Close)

	unreachable := &Device{EndpointRef: "dead", XAddrs: []string{"http://127.0.0.1:1/onvif/device_service"}}
	live := &Device{EndpointRef: "live", XAddrs: []string{camera.URL + "/onvif/device_service"}}
	noAddr := &Device{EndpointRef: "noaddr"}
	preEnriched := &Device{
		EndpointRef: "cached",
		XAddrs:      []string{camera.URL + "/onvif/device_service"},
		Info:        &DeviceInfo{SerialNumber: "PRE"},
	}

	EnrichDevices(context.Background(), []*Device{unreachable, live, noAddr, preEnriched},
		WithEnrichTimeout(500*time.Millisecond), WithEnrichConcurrency(4))

	if live.Info == nil || live.Info.Manufacturer != "EnrichCam" || live.Info.SerialNumber != "ENR-001" {
		t.Errorf("live.Info = %+v, want EnrichCam/ENR-001", live.Info)
	}

	if unreachable.Info != nil {
		t.Errorf("unreachable.Info = %+v, want nil (skipped silently)", unreachable.Info)
	}

	if noAddr.Info != nil {
		t.Errorf("noAddr.Info = %+v, want nil", noAddr.Info)
	}

	if preEnriched.Info.SerialNumber != "PRE" {
		t.Errorf("pre-enriched device was overwritten: %+v", preEnriched.Info)
	}
}

func TestEnrichDevicesRespectsConcurrency(t *testing.T) {
	var inFlight, maxInFlight atomic.Int32

	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		cur := inFlight.Add(1)
		for {
			old := maxInFlight.Load()
			if cur <= old || maxInFlight.CompareAndSwap(old, cur) {
				break
			}
		}
		time.Sleep(80 * time.Millisecond)
		inFlight.Add(-1)
		w.Write([]byte(`<x:GetDeviceInformationResponse xmlns:x="http://www.onvif.org/ver10/device/wsdl"><x:SerialNumber>S</x:SerialNumber></x:GetDeviceInformationResponse>`))
	}))
	t.Cleanup(slow.Close)

	devices := make([]*Device, 8)
	for i := range devices {
		devices[i] = &Device{
			EndpointRef: "dev",
			XAddrs:      []string{slow.URL + "/onvif/device_service"},
		}
	}

	EnrichDevices(context.Background(), devices, WithEnrichConcurrency(2), WithEnrichTimeout(5*time.Second))

	if got := maxInFlight.Load(); got > 2 {
		t.Errorf("max in-flight requests = %d, want <= 2", got)
	}
}

func TestEnrichDevicesCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Must return promptly without touching devices.
	done := make(chan struct{})
	go func() {
		EnrichDevices(ctx, []*Device{{EndpointRef: "x", XAddrs: []string{"http://127.0.0.1:1/"}}})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("EnrichDevices did not return with a canceled context")
	}
}
