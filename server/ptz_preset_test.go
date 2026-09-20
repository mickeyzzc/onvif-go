package server

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
	"github.com/mickeyzzc/onvif-go/v2/server/simulator"
)

// The PTZ preset/configuration family added to close the parity gap with
// the Rust twin: SetPreset / RemovePreset (behind an optional
// PTZPresetWriter provider), GetConfigurations / GetNodes (static,
// configuration-driven), and GetPresets serving the mutable store.

// authedPTZPost mounts all services and posts an authenticated SOAP
// request to the PTZ endpoint — Set/Remove/Go actions are protected by
// the default auth policy.
func authedPTZPost(t *testing.T, s *Server, innerXML string) string {
	t.Helper()

	mux := http.NewServeMux()
	s.RegisterServices(mux)

	envelope := `<?xml version="1.0" encoding="UTF-8"?>` +
		`<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">` +
		`<Header><Security xmlns="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">` +
		`<UsernameToken><Username>admin</Username>` +
		`<Password Type="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordText">` +
		`password</Password></UsernameToken></Security></Header>` +
		`<Body>` + innerXML + `</Body></Envelope>`

	req := httptest.NewRequest(http.MethodPost, s.config.BasePath+"/ptz_service", strings.NewReader(envelope))
	req.RemoteAddr = "127.0.0.1:41000"
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	return w.Body.String()
}

func ptzRequest(action, profileToken, extra string) string {
	return `<` + action + ` xmlns="` + nsPTZWSDL + `"><ProfileToken>` +
		profileToken + `</ProfileToken>` + extra + `</` + action + `>`
}

// responseFragment slices the `<root …>…</root>` element out of a full
// SOAP envelope so it can be decoded namespace-strictly.
func responseFragment(body, root string) string {
	start := strings.Index(body, "<"+root)
	end := strings.Index(body, "</"+root+">")
	if start < 0 || end < 0 {
		return ""
	}

	return body[start : end+len(root)+3]
}

// TestPTZPresetLifecycle proves the mutable store end to end over SOAP:
// set (generated token) → visible in GetPresets → reachable via
// GotoPreset → update by token → removal drops it.
func TestPTZPresetLifecycle(t *testing.T) {
	srv, err := New(createTestConfig())
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	// Set with a name only — the simulator generates the token.
	body := authedPTZPost(t, srv, ptzRequest("SetPreset", "profile_token_1",
		`<PresetName>Gate</PresetName>`))
	if !strings.Contains(body, "<SetPresetResponse") {
		t.Fatalf("SetPreset did not answer; body:\n%s", body)
	}

	var setResp struct {
		XMLName     xml.Name
		PresetToken *string `xml:"PresetToken"`
	}
	if err := xml.Unmarshal([]byte(responseFragment(body, "SetPresetResponse")), &setResp); err != nil {
		t.Fatalf("decode SetPresetResponse: %v\n%s", err, body)
	}
	if setResp.PresetToken == nil || *setResp.PresetToken == "" {
		t.Fatalf("SetPresetResponse carries no generated token:\n%s", body)
	}
	token := *setResp.PresetToken

	// The ns-strict decode above populates PresetToken only if the
	// element resolved to the PTZ WSDL namespace.
	if setResp.XMLName.Space != nsPTZWSDL {
		t.Fatalf("SetPresetResponse root ns = %q, want %q", setResp.XMLName.Space, nsPTZWSDL)
	}

	// GetPresets serves the live store: the new preset is there.
	body = authedPTZPost(t, srv, ptzRequest("GetPresets", "profile_token_1", ""))
	if !strings.Contains(body, ">Gate<") || !strings.Contains(body, `token="`+token+`"`) {
		t.Fatalf("GetPresets after SetPreset missing the new preset:\n%s", body)
	}

	// GotoPreset resolves through the store as well.
	body = authedPTZPost(t, srv, ptzRequest("GotoPreset", "profile_token_1",
		`<PresetToken>`+token+`</PresetToken>`))
	if !strings.Contains(body, "<GotoPresetResponse") {
		t.Fatalf("GotoPreset to the set preset failed:\n%s", body)
	}

	// Set again with the explicit token — updates in place, no growth.
	body = authedPTZPost(t, srv, ptzRequest("SetPreset", "profile_token_1",
		`<PresetName>Gate-East</PresetName><PresetToken>`+token+`</PresetToken>`))
	if !strings.Contains(body, ">"+token+"<") {
		t.Fatalf("SetPreset update did not echo the explicit token:\n%s", body)
	}

	body = authedPTZPost(t, srv, ptzRequest("GetPresets", "profile_token_1", ""))
	if !strings.Contains(body, ">Gate-East<") {
		t.Fatalf("GetPresets after update missing the renamed preset:\n%s", body)
	}

	// Remove it; GetPresets no longer lists it, GotoPreset faults.
	body = authedPTZPost(t, srv, ptzRequest("RemovePreset", "profile_token_1",
		`<PresetToken>`+token+`</PresetToken>`))
	if !strings.Contains(body, "<RemovePresetResponse") {
		t.Fatalf("RemovePreset failed:\n%s", body)
	}

	body = authedPTZPost(t, srv, ptzRequest("GetPresets", "profile_token_1", ""))
	if strings.Contains(body, ">Gate-East<") {
		t.Fatalf("GetPresets still lists the removed preset:\n%s", body)
	}

	body = authedPTZPost(t, srv, ptzRequest("GotoPreset", "profile_token_1",
		`<PresetToken>`+token+`</PresetToken>`))
	if !strings.Contains(body, "Fault") {
		t.Fatalf("GotoPreset to a removed preset should fault:\n%s", body)
	}
}

// TestPTZSetRemoveNotServedWithoutWriter: a read-only PTZ provider does
// not get the write actions registered at all.
func TestPTZSetRemoveNotServedWithoutWriter(t *testing.T) {
	cfg := createTestConfig()

	sim := simulator.New(cfg.Profiles, provider.DeviceInfo{})
	srv, err := New(cfg, WithPTZProvider(readOnlyPTZ{sim}))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	body := authedPTZPost(t, srv, ptzRequest("SetPreset", "profile_token_1",
		`<PresetName>Gate</PresetName>`))
	if !strings.Contains(body, "Fault") || strings.Contains(body, "SetPresetResponse") {
		t.Fatalf("SetPreset must fault without a PTZPresetWriter provider:\n%s", body)
	}
}

// readOnlyPTZ hides the simulator's optional interfaces behind a named
// field (embedding would promote them right back).
type readOnlyPTZ struct {
	sim *simulator.Simulator
}

func (r readOnlyPTZ) ContinuousMove(profileToken string, velocity provider.PTZVector, timeout string) error {
	return r.sim.ContinuousMove(profileToken, velocity, timeout)
}

func (r readOnlyPTZ) AbsoluteMove(profileToken string, position provider.PTZVector) error {
	return r.sim.AbsoluteMove(profileToken, position)
}

func (r readOnlyPTZ) RelativeMove(profileToken string, translation provider.PTZVector) error {
	return r.sim.RelativeMove(profileToken, translation)
}

func (r readOnlyPTZ) Stop(profileToken string, panTilt, zoom bool) error {
	return r.sim.Stop(profileToken, panTilt, zoom)
}

func (r readOnlyPTZ) Status(profileToken string) (provider.PTZState, error) {
	return r.sim.Status(profileToken)
}

func (r readOnlyPTZ) GotoPreset(profileToken, presetToken string) error {
	return r.sim.GotoPreset(profileToken, presetToken)
}

// TestGetConfigurationsWire pins the response namespaces per the WSDL
// ground truth: the PTZConfiguration wrapper resolves to the PTZ WSDL,
// its children to ver10/schema — values only populate on exact-match
// decode.
func TestGetConfigurationsWire(t *testing.T) {
	srv, err := New(createTestConfig())
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	body := authedPTZPost(t, srv, `<GetConfigurations xmlns="`+nsPTZWSDL+`"/>`)
	if !strings.Contains(body, "<GetConfigurationsResponse") {
		t.Fatalf("GetConfigurations did not answer; body:\n%s", body)
	}

	var resp struct {
		XMLName xml.Name
		Configs []struct {
			XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl PTZConfiguration"`
			Token   string   `xml:"token,attr"`
			Name    string   `xml:"http://www.onvif.org/ver10/schema Name"`
			Node    *string  `xml:"http://www.onvif.org/ver10/schema NodeToken"`
			PanTilt *struct {
				Range *struct {
					XRange *struct {
						Min *float64 `xml:"http://www.onvif.org/ver10/schema Min"`
						Max *float64 `xml:"http://www.onvif.org/ver10/schema Max"`
					} `xml:"http://www.onvif.org/ver10/schema XRange"`
				} `xml:"http://www.onvif.org/ver10/schema Range"`
			} `xml:"http://www.onvif.org/ver10/schema PanTiltLimits"`
		} `xml:"http://www.onvif.org/ver20/ptz/wsdl PTZConfiguration"`
	}
	if err := xml.Unmarshal([]byte(responseFragment(body, "GetConfigurationsResponse")), &resp); err != nil {
		t.Fatalf("decode GetConfigurationsResponse: %v\n%s", err, body)
	}

	if len(resp.Configs) != 1 {
		t.Fatalf("GetConfigurations returned %d configurations, want 1:\n%s", len(resp.Configs), body)
	}

	cfg := resp.Configs[0]
	if cfg.Token != "profile_token_1" {
		t.Errorf("configuration token = %q, want profile_token_1", cfg.Token)
	}
	if cfg.Node == nil || *cfg.Node != "ptz_node_1" {
		t.Errorf("NodeToken did not resolve in ver10/schema: %+v", cfg.Node)
	}
	if cfg.PanTilt == nil || cfg.PanTilt.Range == nil ||
		cfg.PanTilt.Range.XRange == nil || cfg.PanTilt.Range.XRange.Min == nil ||
		*cfg.PanTilt.Range.XRange.Min != -360 || *cfg.PanTilt.Range.XRange.Max != 360 {
		t.Errorf("pan limits did not resolve with values: %+v", cfg.PanTilt)
	}
}

// TestGetNodesWire pins the PTZNode shape: the token attribute (per
// tt:DeviceEntity, not the twin's NodeToken spelling), schema-typed
// children, and movement spaces gated on the capability flags.
func TestGetNodesWire(t *testing.T) {
	cfg := createTestConfig()
	cfg.Profiles[0].PTZ.SupportsAbsolute = true
	cfg.Profiles[0].PTZ.SupportsContinuous = true

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	body := authedPTZPost(t, srv, `<GetNodes xmlns="`+nsPTZWSDL+`"/>`)
	if !strings.Contains(body, "<GetNodesResponse") {
		t.Fatalf("GetNodes did not answer; body:\n%s", body)
	}

	var resp struct {
		Nodes []struct {
			Token      string  `xml:"token,attr"`
			Name       *string `xml:"http://www.onvif.org/ver10/schema Name"`
			MaxPresets *int    `xml:"http://www.onvif.org/ver10/schema MaximumNumberOfPresets"`
			Spaces     *struct {
				AbsPanTilt *struct {
					URI *string `xml:"http://www.onvif.org/ver10/schema URI"`
				} `xml:"http://www.onvif.org/ver10/schema AbsolutePanTiltPositionSpace"`
				ContZoom *struct {
					URI *string `xml:"http://www.onvif.org/ver10/schema URI"`
				} `xml:"http://www.onvif.org/ver10/schema ContinuousZoomVelocitySpace"`
			} `xml:"http://www.onvif.org/ver10/schema SupportedPTZSpaces"`
		} `xml:"http://www.onvif.org/ver20/ptz/wsdl PTZNode"`
	}
	if err := xml.Unmarshal([]byte(responseFragment(body, "GetNodesResponse")), &resp); err != nil {
		t.Fatalf("decode GetNodesResponse: %v\n%s", err, body)
	}

	if len(resp.Nodes) != 1 {
		t.Fatalf("GetNodes returned %d nodes, want 1:\n%s", len(resp.Nodes), body)
	}

	node := resp.Nodes[0]
	if node.Token != "ptz_node_1" {
		t.Errorf("node token attribute = %q, want ptz_node_1", node.Token)
	}
	if node.Name == nil || *node.Name != "ptz_node_1" {
		t.Errorf("node Name did not resolve in ver10/schema: %+v", node.Name)
	}
	if node.MaxPresets == nil || *node.MaxPresets != maxPresetsPerNode {
		t.Errorf("MaximumNumberOfPresets = %+v, want %d", node.MaxPresets, maxPresetsPerNode)
	}
	if node.Spaces == nil || node.Spaces.AbsPanTilt == nil || node.Spaces.AbsPanTilt.URI == nil {
		t.Errorf("absolute pan/tilt space missing while SupportsAbsolute is set:\n%s", body)
	}
	if node.Spaces == nil || node.Spaces.ContZoom == nil || node.Spaces.ContZoom.URI == nil {
		t.Errorf("continuous zoom space missing while SupportsContinuous is set:\n%s", body)
	}

	// Capability flags off → no spaces advertised.
	cfg2 := createTestConfig()
	srv2, err := New(cfg2)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	body2 := authedPTZPost(t, srv2, `<GetNodes xmlns="`+nsPTZWSDL+`"/>`)
	if strings.Contains(body2, "PositionSpace") || strings.Contains(body2, "VelocitySpace") {
		t.Fatalf("spaces advertised without capability flags:\n%s", body2)
	}
}

// TestSimulatorPresetStoreIsolation pins that the reader returns copies:
// callers cannot mutate the simulator state through the slice they get.
func TestSimulatorPresetStoreIsolation(t *testing.T) {
	sim := simulator.New(createTestConfig().Profiles, provider.DeviceInfo{})

	presets, err := sim.Presets("profile_token_1")
	if err != nil {
		t.Fatalf("Presets(): %v", err)
	}

	if len(presets) != 0 {
		t.Fatalf("expected the seeded (empty) store, got %d presets", len(presets))
	}

	token, err := sim.SetPreset("profile_token_1", "Gate", "")
	if err != nil {
		t.Fatalf("SetPreset(): %v", err)
	}

	presets, _ = sim.Presets("profile_token_1")
	presets[0].Name = "mutated-outside"

	again, _ := sim.Presets("profile_token_1")
	if again[0].Name != "Gate" {
		t.Fatalf("store mutated through the returned slice: %q", again[0].Name)
	}

	if err := sim.RemovePreset("profile_token_1", token); err != nil {
		t.Fatalf("RemovePreset(): %v", err)
	}

	if err := sim.RemovePreset("profile_token_1", fmt.Sprintf("%s-x", token)); err == nil {
		t.Fatal("removing an unknown preset must fail")
	}
}
