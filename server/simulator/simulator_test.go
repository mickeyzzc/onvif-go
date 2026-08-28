package simulator

import (
	"bytes"
	"image/jpeg"
	"testing"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
)

func testProfiles() []provider.ProfileConfig {
	return []provider.ProfileConfig{
		{
			Token: "profile-1",
			Name:  "Main",
			VideoSource: provider.VideoSourceConfig{
				Token:      "vs-1",
				Resolution: provider.Resolution{Width: 640, Height: 480},
			},
			PTZ: &provider.PTZConfig{},
			Snapshot: provider.SnapshotConfig{
				Enabled:    true,
				Resolution: provider.Resolution{Width: 320, Height: 240},
			},
		},
		{
			Token: "profile-2",
			Name:  "NoPTZ",
			VideoSource: provider.VideoSourceConfig{
				Token:      "vs-2",
				Resolution: provider.Resolution{Width: 640, Height: 480},
			},
			Snapshot: provider.SnapshotConfig{Enabled: false},
		},
	}
}

func newSim() *Simulator {
	return New(testProfiles(), provider.DeviceInfo{Manufacturer: "T", Model: "M"})
}

func TestStreamLifecycle(t *testing.T) {
	sim := newSim()

	info, err := sim.Stream("profile-1")
	if err != nil || info.RTSPPath != "/stream0" || info.OverrideURI != "" {
		t.Fatalf("Stream = %+v, %v", info, err)
	}

	if _, err := sim.Stream("nope"); err == nil {
		t.Error("unknown profile must error")
	}

	if err := sim.SetStreamURI("profile-1", "rtsp://pinned/stream"); err != nil {
		t.Fatalf("SetStreamURI: %v", err)
	}

	info, _ = sim.Stream("profile-1")
	if info.OverrideURI != "rtsp://pinned/stream" {
		t.Errorf("override not stored: %+v", info)
	}

	if err := sim.SetStreamURI("nope", "x"); err == nil {
		t.Error("SetStreamURI on unknown profile must error")
	}
}

func TestSnapshotCachingAndGuards(t *testing.T) {
	sim := newSim()

	first, err := sim.Snapshot("profile-1")
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	if first.ContentType != "" {
		t.Errorf("simulator snapshot ContentType = %q, want empty (JPEG default)", first.ContentType)
	}

	cfg, err := jpeg.DecodeConfig(bytes.NewReader(first.Data))
	if err != nil || cfg.Width != 320 || cfg.Height != 240 {
		t.Errorf("snapshot not valid JPEG at configured size: %v, %+v", err, cfg)
	}

	second, _ := sim.Snapshot("profile-1")
	if !bytes.Equal(first.Data, second.Data) {
		t.Error("snapshot must be cached per profile")
	}

	if _, err := sim.Snapshot("profile-2"); err == nil {
		t.Error("disabled snapshot must error")
	}

	if _, err := sim.Snapshot("nope"); err == nil {
		t.Error("unknown profile snapshot must error")
	}
}

func TestPTZOpsAndClamps(t *testing.T) {
	sim := newSim()

	if err := sim.ContinuousMove("profile-1", provider.PTZVector{
		PanTilt: &provider.Vector2D{X: 1, Y: 1},
	}, "PT1S"); err != nil {
		t.Fatalf("ContinuousMove: %v", err)
	}

	state, _ := sim.Status("profile-1")
	if !state.Moving || !state.PanMoving {
		t.Errorf("moving state = %+v", state)
	}

	if err := sim.Stop("profile-1", true, true); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	state, _ = sim.Status("profile-1")
	if state.Moving {
		t.Errorf("stopped state = %+v", state)
	}

	// Relative move clamps to the pan/tilt envelope.
	if err := sim.RelativeMove("profile-1", provider.PTZVector{
		PanTilt: &provider.Vector2D{X: 1000, Y: -1000},
	}); err != nil {
		t.Fatalf("RelativeMove: %v", err)
	}

	state, _ = sim.Status("profile-1")
	if state.Position.Pan > 180 || state.Position.Tilt < -90 {
		t.Errorf("clamp failed: %+v", state.Position)
	}

	if err := sim.ContinuousMove("profile-2", provider.PTZVector{}, ""); err == nil {
		t.Error("PTZ on non-PTZ profile must error")
	}

	if _, err := sim.Status("profile-2"); err == nil {
		t.Error("Status on non-PTZ profile must error")
	}
}

func TestAbsoluteMoveCompletionTimer(t *testing.T) {
	sim := newSim()

	if err := sim.AbsoluteMove("profile-1", provider.PTZVector{
		PanTilt: &provider.Vector2D{X: 30, Y: -10},
		Zoom:    &provider.Vector1D{X: 0.5},
	}); err != nil {
		t.Fatalf("AbsoluteMove: %v", err)
	}

	state, _ := sim.Status("profile-1")
	if state.Position.Pan != 30 || state.Position.Zoom != 0.5 {
		t.Fatalf("position not applied: %+v", state.Position)
	}

	if !state.Moving {
		t.Fatal("must report moving right after the move")
	}

	// The completion timer flips the flags off after ptzMoveDuration.
	deadline := time.Now().Add(2 * time.Second)
	for state.Moving && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
		state, _ = sim.Status("profile-1")
	}

	if state.Moving {
		t.Error("completion timer did not stop movement (stuck state)")
	}
}

func TestGotoPresetResolvesConfiguration(t *testing.T) {
	profiles := testProfiles()
	profiles[0].PTZ.Presets = []provider.Preset{
		{Token: "home", Name: "Home", Position: provider.PTZPosition{Pan: 5, Tilt: 6, Zoom: 7}},
	}
	sim := New(profiles, provider.DeviceInfo{})

	if err := sim.GotoPreset("profile-1", "home"); err != nil {
		t.Fatalf("GotoPreset: %v", err)
	}

	state, _ := sim.Status("profile-1")
	if state.Position.Pan != 5 || state.Position.Zoom != 7 {
		t.Errorf("preset not applied: %+v", state.Position)
	}

	if err := sim.GotoPreset("profile-1", "missing"); err == nil {
		t.Error("missing preset must error")
	}

	if err := sim.GotoPreset("profile-2", "home"); err == nil {
		t.Error("GotoPreset on non-PTZ profile must error")
	}
}

func TestImagingRoundTrip(t *testing.T) {
	sim := newSim()

	settings, err := sim.ImagingSettings("vs-1")
	if err != nil {
		t.Fatalf("ImagingSettings: %v", err)
	}

	if settings.Brightness == nil || *settings.Brightness != 50 {
		t.Errorf("default brightness = %v", settings.Brightness)
	}

	newBrightness := 72.5
	if err := sim.SetImagingSettings("vs-1", &provider.ImagingSettings{Brightness: &newBrightness}); err != nil {
		t.Fatalf("SetImagingSettings: %v", err)
	}

	settings, _ = sim.ImagingSettings("vs-1")
	if *settings.Brightness != 72.5 {
		t.Errorf("brightness not persisted: %v", *settings.Brightness)
	}

	if _, err := sim.ImagingSettings("nope"); err == nil {
		t.Error("unknown source must error")
	}

	if _, err := sim.ImagingOptions("nope"); err == nil {
		t.Error("unknown source options must error")
	}

	options, err := sim.ImagingOptions("vs-1")
	if err != nil || options.Brightness == nil {
		t.Errorf("ImagingOptions = %+v, %v", options, err)
	}

	// Focus move with clamping.
	if err := sim.MoveFocus("vs-1", &provider.FocusMove{
		Relative: &provider.RelativeFocus{Distance: 0.9},
	}); err != nil {
		t.Fatalf("MoveFocus: %v", err)
	}

	state, ok := sim.ImagingStateOf("vs-1")
	if !ok || state.Focus.CurrentPos > 1 {
		t.Errorf("focus not clamped: %+v", state.Focus)
	}
}

func TestDeviceInfoAndProfileLookup(t *testing.T) {
	sim := newSim()

	if info := sim.DeviceInfo(); info.Manufacturer != "T" || info.Model != "M" {
		t.Errorf("DeviceInfo = %+v", info)
	}

	if p, ok := sim.Profile("profile-2"); !ok || p.Name != "NoPTZ" {
		t.Errorf("Profile = %+v, %v", p, ok)
	}

	if _, ok := sim.Profile("nope"); ok {
		t.Error("unknown profile must not resolve")
	}
}

// Compile-time provider interface checks.
var (
	_ provider.DeviceInfoProvider = (*Simulator)(nil)
	_ provider.StreamURIProvider  = (*Simulator)(nil)
	_ provider.StreamURISetter    = (*Simulator)(nil)
	_ provider.SnapshotProvider   = (*Simulator)(nil)
	_ provider.ImagingProvider    = (*Simulator)(nil)
	_ provider.PTZProvider        = (*Simulator)(nil)
)
