package server

import (
	"context"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	onvif "github.com/mickeyzzc/onvif-go/v2/onvif"
	"github.com/mickeyzzc/onvif-go/v2/ptz"
)

// Conformance loopback: the library's own client driven against the
// simulator over real HTTP, across the full served operation matrix. The
// namespace contract suites pin single wire shapes; these tests prove the
// client can actually read what the server writes (and vice versa) — the
// class of mismatch issue #90 exposed.

// startConformanceServer boots a simulator-backed server with every
// service enabled plus a PTZ preset, and connects a client through the
// real GetCapabilities discovery dance.
func startConformanceServer(t *testing.T) (*httptest.Server, *onvif.Client, *Config) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	config := createTestConfig()
	config.Port = listener.Addr().(*net.TCPAddr).Port
	config.SupportPTZ = true
	config.SupportImaging = true
	config.SupportEvents = true
	config.Scopes = []string{
		"onvif://www.onvif.org/name/ConformanceCam",
		"onvif://www.onvif.org/type/video_encoder",
	}
	config.Profiles[0].PTZ.Presets = []Preset{{
		Token: "preset_gate", Name: "Gate",
		Position: PTZPosition{Pan: 10, Tilt: -5, Zoom: 2},
	}}

	s, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ts := httptest.NewUnstartedServer(s.Handler())
	ts.Listener = listener
	ts.Start()
	t.Cleanup(ts.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)

	client, err := onvif.NewClient(
		ts.URL+"/onvif/device_service",
		onvif.WithCredentials("admin", "password"),
		onvif.WithTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("onvif.NewClient() error = %v", err)
	}

	if err := client.Initialize(ctx); err != nil {
		t.Fatalf("client.Initialize() error = %v", err)
	}

	return ts, client, config
}

func TestConformanceDeviceMatrix(t *testing.T) {
	_, client, config := startConformanceServer(t)
	ctx := context.Background()

	services, err := client.Device().GetServices(ctx, false)
	if err != nil {
		t.Fatalf("GetServices: %v", err)
	}

	got := map[string]string{}
	for _, svc := range services {
		got[svc.Namespace] = svc.XAddr
	}

	for _, ns := range []string{
		"http://www.onvif.org/ver10/device/wsdl",
		"http://www.onvif.org/ver10/media/wsdl",
		"http://www.onvif.org/ver20/ptz/wsdl",
		"http://www.onvif.org/ver20/imaging/wsdl",
		"http://www.onvif.org/ver10/events/wsdl",
	} {
		if got[ns] == "" {
			t.Errorf("GetServices missing %s (have %v)", ns, got)
		}
	}

	scopes, err := client.Device().GetScopes(ctx)
	if err != nil {
		t.Fatalf("GetScopes: %v", err)
	}

	scopeItems := map[string]bool{}
	for _, sc := range scopes {
		scopeItems[sc.ScopeItem] = true
	}

	for _, want := range config.Scopes {
		if !scopeItems[want] {
			t.Errorf("GetScopes missing %q (parsed %v)", want, scopeItems)
		}
	}

	dt, err := client.Device().FixedGetSystemDateAndTime(ctx)
	if err != nil {
		t.Fatalf("FixedGetSystemDateAndTime: %v", err)
	}

	if dt.UTCDateTime == nil || dt.UTCDateTime.Date.Year < 2026 {
		t.Errorf("GetSystemDateAndTime produced an implausible UTC time: %+v", dt.UTCDateTime)
	}

	caps, err := client.Device().GetCapabilities(ctx)
	if err != nil {
		t.Fatalf("GetCapabilities: %v", err)
	}

	for name, xaddr := range map[string]string{
		"Device":  caps.Device.XAddr,
		"Media":   caps.Media.XAddr,
		"PTZ":     caps.PTZ.XAddr,
		"Imaging": caps.Imaging.XAddr,
		"Events":  caps.Events.XAddr,
	} {
		if xaddr == "" {
			t.Errorf("GetCapabilities %s XAddr empty", name)
		}
	}

	msg, err := client.Device().SystemReboot(ctx)
	if err != nil {
		t.Fatalf("SystemReboot: %v", err)
	}

	if !strings.Contains(msg, "reboot") {
		t.Errorf("SystemReboot message = %q, want a reboot note", msg)
	}
}

func TestConformancePTZMatrix(t *testing.T) {
	_, client, config := startConformanceServer(t)
	ctx := context.Background()
	token := config.Profiles[0].Token

	speed := &ptz.PTZSpeed{PanTilt: &ptz.Vector2D{X: 0.5, Y: 0}}
	if err := client.PTZ().ContinuousMove(ctx, token, speed, nil); err != nil {
		t.Fatalf("ContinuousMove: %v", err)
	}

	status, err := client.PTZ().GetStatus(ctx, token)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}

	if status.MoveStatus == nil || status.MoveStatus.PanTilt != "MOVING" {
		t.Errorf("MoveStatus.PanTilt after ContinuousMove = %+v, want MOVING", status.MoveStatus)
	}

	if err := client.PTZ().Stop(ctx, token, true, true); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	status, err = client.PTZ().GetStatus(ctx, token)
	if err != nil {
		t.Fatalf("GetStatus after Stop: %v", err)
	}

	if status.MoveStatus == nil || status.MoveStatus.PanTilt != "IDLE" {
		t.Errorf("MoveStatus.PanTilt after Stop = %+v, want IDLE", status.MoveStatus)
	}

	position := &ptz.PTZVector{
		PanTilt: &ptz.Vector2D{X: 10, Y: -5},
		Zoom:    &ptz.Vector1D{X: 2},
	}
	if err := client.PTZ().AbsoluteMove(ctx, token, position, nil); err != nil {
		t.Fatalf("AbsoluteMove: %v", err)
	}

	presets, err := client.PTZ().GetPresets(ctx, token)
	if err != nil {
		t.Fatalf("GetPresets: %v", err)
	}

	if len(presets) != 1 {
		t.Fatalf("GetPresets count = %d, want 1: %+v", len(presets), presets)
	}

	preset := presets[0]
	if preset.Token != "preset_gate" || preset.Name != "Gate" {
		t.Errorf("preset = %q/%q, want preset_gate/Gate", preset.Token, preset.Name)
	}

	if preset.PTZPosition == nil ||
		preset.PTZPosition.PanTilt == nil || preset.PTZPosition.PanTilt.X != 10 ||
		preset.PTZPosition.Zoom == nil || preset.PTZPosition.Zoom.X != 2 {
		t.Errorf("preset position not parsed (tt: vectors): %+v", preset.PTZPosition)
	}

	if err := client.PTZ().GotoPreset(ctx, token, preset.Token, nil); err != nil {
		t.Fatalf("GotoPreset: %v", err)
	}

	// The parity closure: preset writes + configuration/node enumeration
	// (SetPreset/RemovePreset behind PTZPresetWriter; GetConfigurations
	// and GetNodes from the profile configuration).
	configs, err := client.PTZ().GetConfigurations(ctx)
	if err != nil {
		t.Fatalf("GetConfigurations: %v", err)
	}

	if len(configs) != 1 || configs[0].NodeToken != config.Profiles[0].PTZ.NodeToken {
		t.Fatalf("GetConfigurations = %+v, want one entry with node %s",
			configs, config.Profiles[0].PTZ.NodeToken)
	}

	if configs[0].PanTiltLimits == nil || configs[0].PanTiltLimits.Range.XRange == nil {
		t.Errorf("GetConfigurations pan limits not parsed: %+v", configs[0].PanTiltLimits)
	}

	newToken, err := client.PTZ().SetPreset(ctx, token, "Entrance", "")
	if err != nil {
		t.Fatalf("SetPreset (generated): %v", err)
	}

	if newToken == "" {
		t.Fatal("SetPreset returned an empty token")
	}

	if _, err := client.PTZ().SetPreset(ctx, token, "Entrance-East", newToken); err != nil {
		t.Fatalf("SetPreset (update by token): %v", err)
	}

	presets, err = client.PTZ().GetPresets(ctx, token)
	if err != nil {
		t.Fatalf("GetPresets after SetPreset: %v", err)
	}

	if len(presets) != 2 || presets[1].Token != newToken || presets[1].Name != "Entrance-East" {
		t.Fatalf("presets after SetPreset = %+v, want the configured one plus renamed %s", presets, newToken)
	}

	if err := client.PTZ().GotoPreset(ctx, token, newToken, nil); err != nil {
		t.Fatalf("GotoPreset to the generated preset: %v", err)
	}

	if err := client.PTZ().RemovePreset(ctx, token, newToken); err != nil {
		t.Fatalf("RemovePreset: %v", err)
	}

	presets, err = client.PTZ().GetPresets(ctx, token)
	if err != nil {
		t.Fatalf("GetPresets after RemovePreset: %v", err)
	}

	if len(presets) != 1 || presets[0].Token != "preset_gate" {
		t.Fatalf("presets after RemovePreset = %+v, want only preset_gate back", presets)
	}
}

func TestConformanceImagingMatrix(t *testing.T) {
	_, client, config := startConformanceServer(t)
	ctx := context.Background()
	source := config.Profiles[0].VideoSource.Token

	settings, err := client.Imaging().GetImagingSettings(ctx, source)
	if err != nil {
		t.Fatalf("GetImagingSettings: %v", err)
	}

	brightness := 42.0
	settings.Brightness = &brightness
	if err := client.Imaging().SetImagingSettings(ctx, source, settings, false); err != nil {
		t.Fatalf("SetImagingSettings: %v", err)
	}

	reread, err := client.Imaging().GetImagingSettings(ctx, source)
	if err != nil {
		t.Fatalf("GetImagingSettings (reread): %v", err)
	}

	if reread.Brightness == nil || *reread.Brightness != 42 {
		t.Errorf("Brightness round-trip = %v, want 42", reread.Brightness)
	}

	options, err := client.Imaging().GetOptions(ctx, source)
	if err != nil {
		t.Fatalf("GetOptions: %v", err)
	}

	if options == nil || options.Brightness == nil {
		t.Fatalf("GetOptions missing brightness range: %+v", options)
	}
}
