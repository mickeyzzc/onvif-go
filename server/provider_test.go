package server

import (
	"bytes"
	"errors"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
)

// stubProviders is a hardware-style backend: fixed values, recorded
// calls — the shape a MiBee-Eye-class embedder would bring.
type stubProviders struct {
	info    provider.DeviceInfo
	streams map[string]provider.StreamInfo
	jpegs   map[string][]byte

	imagingSettings map[string]*provider.ImagingSettings
	imagingCalls    []string
	ptzMoves        []string
	ptzStatus       map[string]provider.PTZState
}

func (p *stubProviders) DeviceInfo() provider.DeviceInfo { return p.info }

func (p *stubProviders) Stream(token string) (provider.StreamInfo, error) {
	info, ok := p.streams[token]
	if !ok {
		return provider.StreamInfo{}, errors.New("no such stream")
	}

	return info, nil
}

func (p *stubProviders) Snapshot(token string) ([]byte, error) {
	data, ok := p.jpegs[token]
	if !ok {
		return nil, errors.New("no snapshot")
	}

	return data, nil
}

func (p *stubProviders) ImagingSettings(token string) (*provider.ImagingSettings, error) {
	settings, ok := p.imagingSettings[token]
	if !ok {
		return nil, errors.New("no such source")
	}

	return settings, nil
}

func (p *stubProviders) SetImagingSettings(token string, settings *provider.ImagingSettings) error {
	p.imagingCalls = append(p.imagingCalls, "set:"+token)

	return nil
}

func (p *stubProviders) ImagingOptions(string) (*provider.ImagingOptions, error) {
	return &provider.ImagingOptions{Brightness: &provider.FloatRange{Min: 0, Max: 42}}, nil
}

func (p *stubProviders) MoveFocus(token string, _ *provider.FocusMove) error {
	p.imagingCalls = append(p.imagingCalls, "move:"+token)

	return nil
}

func (p *stubProviders) ContinuousMove(token string, _ provider.PTZVector, _ string) error {
	p.ptzMoves = append(p.ptzMoves, "continuous:"+token)

	return nil
}

func (p *stubProviders) AbsoluteMove(token string, _ provider.PTZVector) error {
	p.ptzMoves = append(p.ptzMoves, "absolute:"+token)

	return nil
}

func (p *stubProviders) RelativeMove(token string, _ provider.PTZVector) error {
	return nil
}

func (p *stubProviders) Stop(string, bool, bool) error { return nil }

func (p *stubProviders) Status(token string) (provider.PTZState, error) {
	state, ok := p.ptzStatus[token]
	if !ok {
		return provider.PTZState{}, errors.New("no ptz")
	}

	return state, nil
}

func (p *stubProviders) GotoPreset(string, string) error { return nil }

func newStubbedServer(t *testing.T) (*Server, *stubProviders) {
	t.Helper()

	stub := &stubProviders{
		info: provider.DeviceInfo{
			Manufacturer:    "MiBee",
			Model:           "Eye",
			FirmwareVersion: "9.9",
			SerialNumber:    "EYE-1",
			HardwareID:      "EYE-HW",
		},
		streams: map[string]provider.StreamInfo{},
		jpegs:   map[string][]byte{},
	}

	config := createTestConfig()
	for _, profile := range config.Profiles {
		stub.streams[profile.Token] = provider.StreamInfo{
			OverrideURI: "rtsp://hardware.example/" + profile.Token,
		}
		stub.jpegs[profile.Token] = []byte("\xff\xd8\xff\xe0HARDWARE-JPEG\xff\xd9")
	}
	stub.imagingSettings = map[string]*provider.ImagingSettings{
		config.Profiles[0].VideoSource.Token: {Brightness: ptrFloat(7)},
	}
	stub.ptzStatus = map[string]provider.PTZState{
		config.Profiles[0].Token: {Position: provider.PTZPosition{Pan: 11, Tilt: 12, Zoom: 13}},
	}

	server, err := New(config,
		WithDeviceInfoProvider(stub),
		WithStreamURIProvider(stub),
		WithSnapshotProvider(stub),
		WithImagingProvider(stub),
		WithPTZProvider(stub),
	)
	if err != nil {
		t.Fatalf("New with providers: %v", err)
	}

	return server, stub
}

func ptrFloat(v float64) *float64 { return &v }

func TestProviderDeviceInformation(t *testing.T) {
	server, _ := newStubbedServer(t)

	resp, err := server.HandleGetDeviceInformation(nil, nil)
	if err != nil {
		t.Fatalf("HandleGetDeviceInformation: %v", err)
	}

	info := resp.(*GetDeviceInformationResponse)
	if info.Manufacturer != "MiBee" || info.Model != "Eye" || info.SerialNumber != "EYE-1" {
		t.Errorf("device info not served by provider: %+v", info)
	}
}

func TestProviderStreamURIVerbatim(t *testing.T) {
	server, _ := newStubbedServer(t)

	token := server.config.Profiles[0].Token
	body := []byte(`<GetStreamUri xmlns="http://www.onvif.org/ver10/media/wsdl"><ProfileToken>` + token + `</ProfileToken></GetStreamUri>`)

	resp, err := server.HandleGetStreamUri(nil, body)
	if err != nil {
		t.Fatalf("HandleGetStreamUri: %v", err)
	}

	uri := resp.(*GetStreamUriResponse)
	if uri.MediaUri.URI != "rtsp://hardware.example/"+token {
		t.Errorf("stream URI = %q, want the provider override verbatim", uri.MediaUri.URI)
	}
}

func TestProviderSnapshotEndpoint(t *testing.T) {
	server, stub := newStubbedServer(t)

	mux := http.NewServeMux()
	mux.HandleFunc(server.config.BasePath+"/snapshot", server.handleSnapshot)

	token := server.config.Profiles[0].Token
	req := httptest.NewRequest(http.MethodGet, "/onvif/snapshot?profile="+token, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("snapshot status = %d, want 200", w.Code)
	}

	if !bytes.Equal(w.Body.Bytes(), stub.jpegs[token]) {
		t.Errorf("snapshot bytes not served verbatim by provider")
	}

	if ct := w.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Errorf("Content-Type = %q, want image/jpeg", ct)
	}
}

func TestProviderImagingDelegation(t *testing.T) {
	server, stub := newStubbedServer(t)

	token := server.config.Profiles[0].VideoSource.Token
	body := []byte(`<GetImagingSettings xmlns="http://www.onvif.org/ver20/imaging/wsdl"><VideoSourceToken>` + token + `</VideoSourceToken></GetImagingSettings>`)

	resp, err := server.HandleGetImagingSettings(nil, body)
	if err != nil {
		t.Fatalf("HandleGetImagingSettings: %v", err)
	}

	settings := resp.(*GetImagingSettingsResponse).ImagingSettings
	if settings.Brightness == nil || *settings.Brightness != 7 {
		t.Errorf("imaging settings not served by provider: %+v", settings)
	}

	setBody := []byte(`<SetImagingSettings xmlns="http://www.onvif.org/ver20/imaging/wsdl"><VideoSourceToken>` +
		token + `</VideoSourceToken><ImagingSettings><Contrast>3</Contrast></ImagingSettings></SetImagingSettings>`)

	if _, err := server.HandleSetImagingSettings(nil, setBody); err != nil {
		t.Fatalf("HandleSetImagingSettings: %v", err)
	}

	if len(stub.imagingCalls) != 1 || stub.imagingCalls[0] != "set:"+token {
		t.Errorf("SetImagingSettings not delegated: %v", stub.imagingCalls)
	}
}

func TestProviderPTZDelegation(t *testing.T) {
	server, stub := newStubbedServer(t)

	token := server.config.Profiles[0].Token
	body := []byte(`<ContinuousMove xmlns="http://www.onvif.org/ver20/ptz/wsdl"><ProfileToken>` + token + `</ProfileToken></ContinuousMove>`)

	if _, err := server.HandleContinuousMove(nil, body); err != nil {
		t.Fatalf("HandleContinuousMove: %v", err)
	}

	if len(stub.ptzMoves) != 1 || stub.ptzMoves[0] != "continuous:"+token {
		t.Errorf("ContinuousMove not delegated: %v", stub.ptzMoves)
	}

	statusBody := []byte(`<GetStatus xmlns="http://www.onvif.org/ver20/ptz/wsdl"><ProfileToken>` + token + `</ProfileToken></GetStatus>`)

	resp, err := server.HandleGetStatus(nil, statusBody)
	if err != nil {
		t.Fatalf("HandleGetStatus: %v", err)
	}

	status := resp.(*GetStatusResponse).PTZStatus
	if status.Position.PanTilt == nil || status.Position.PanTilt.X != 11 {
		t.Errorf("status not served by provider: %+v", status)
	}
}

func TestSimulatorSnapshotValidJPEG(t *testing.T) {
	config := createTestConfig()
	server, _ := New(config) // no options: simulator default

	token := config.Profiles[0].Token

	mux := http.NewServeMux()
	mux.HandleFunc(config.BasePath+"/snapshot", server.handleSnapshot)

	req := httptest.NewRequest(http.MethodGet, "/onvif/snapshot?profile="+token, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	cfg, err := jpeg.DecodeConfig(bytes.NewReader(w.Body.Bytes()))
	if err != nil {
		t.Fatalf("simulator snapshot is not valid JPEG: %v", err)
	}

	if cfg.Width != config.Profiles[0].Snapshot.Resolution.Width {
		t.Errorf("snapshot width = %d, want %d", cfg.Width, config.Profiles[0].Snapshot.Resolution.Width)
	}

	// Disabled snapshot -> 501.
	disabled := createTestConfig()
	disabled.Profiles[0].Snapshot.Enabled = false
	server2, _ := New(disabled)
	mux2 := http.NewServeMux()
	mux2.HandleFunc(disabled.BasePath+"/snapshot", server2.handleSnapshot)

	req2 := httptest.NewRequest(http.MethodGet, "/onvif/snapshot?profile="+disabled.Profiles[0].Token, nil)
	w2 := httptest.NewRecorder()
	mux2.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotImplemented {
		t.Errorf("disabled snapshot status = %d, want 501", w2.Code)
	}
}

func TestNonSetterStreamProviderRejectsUpdate(t *testing.T) {
	server, _ := newStubbedServer(t)

	token := server.config.Profiles[0].Token
	if err := server.UpdateStreamURI(token, "rtsp://example/new"); err == nil {
		t.Error("UpdateStreamURI must fail for providers without StreamURISetter")
	}
}

func TestSimulatorPresetGotoThroughProvider(t *testing.T) {
	config := DefaultConfig()
	server, _ := New(config)

	token := config.Profiles[0].Token
	preset := config.Profiles[0].PTZ.Presets[1]

	body := []byte(`<GotoPreset xmlns="http://www.onvif.org/ver20/ptz/wsdl"><ProfileToken>` +
		token + `</ProfileToken><PresetToken>` + preset.Token + `</PresetToken></GotoPreset>`)

	if _, err := server.HandleGotoPreset(nil, body); err != nil {
		t.Fatalf("HandleGotoPreset: %v", err)
	}

	state, ok := server.GetPTZState(token)
	if !ok {
		t.Fatal("PTZ state missing")
	}

	if state.Position.Pan != preset.Position.Pan || state.Position.Tilt != preset.Position.Tilt {
		t.Errorf("preset position not applied: %+v want %+v", state.Position, preset.Position)
	}
}
