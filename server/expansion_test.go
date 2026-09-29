package server

import (
	"context"
	"encoding/xml"
	"net"
	"net/http/httptest"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/media"
	"github.com/mickeyzzc/onvif-go/v2/onvif"
	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// Issue #115: the virtual camera grows DeviceIO (relays / digital inputs),
// the media OSD loop, the audio configuration family (empty sets), and a
// minimal Media2 face — every surface the client packages can already
// call, so NVR logic can be tested without real hardware.

// startExpansionServer boots the simulator with DeviceIO + Media2 enabled
// and a client wired through real discovery.
func startExpansionServer(t *testing.T) (*httptest.Server, *onvif.Client, *Config) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	config := createTestConfig()
	config.Port = listener.Addr().(*net.TCPAddr).Port
	config.SupportDeviceIO = true
	config.SupportMedia2 = true

	s, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ts := httptest.NewUnstartedServer(s.Handler())
	ts.Listener = listener
	ts.Start()
	t.Cleanup(ts.Close)

	client, err := onvif.NewClient(ts.URL, onvif.WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Initialize(context.Background()); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	// Media2 rides the media1 endpoint by default; pin the dedicated
	// tr2 face like a Media2-discovering client would (GetServices probe
	// or explicit pin).
	client.SetServiceEndpoint(api.ServiceMedia2, ts.URL+config.BasePath+"/media2_service")

	return ts, client, config
}

func TestDeviceIORelayLoopback(t *testing.T) {
	_, client, _ := startExpansionServer(t)
	ctx := context.Background()

	outputs, err := client.DeviceIO().GetRelayOutputs(ctx)
	if err != nil {
		t.Fatalf("GetRelayOutputs: %v", err)
	}
	if len(outputs) == 0 {
		t.Fatal("simulator must expose at least one relay output")
	}
	first := outputs[0]
	if first.Token == "" {
		t.Errorf("relay token empty: %+v", first)
	}

	// NVR alarm-linkage flow: activate, verify, release.
	if err := client.DeviceIO().SetRelayOutputState(ctx, first.Token, "active"); err != nil {
		t.Fatalf("SetRelayOutputState(active): %v", err)
	}
	outputs, err = client.DeviceIO().GetRelayOutputs(ctx)
	if err != nil {
		t.Fatalf("GetRelayOutputs after set: %v", err)
	}
	for _, o := range outputs {
		if o.Token == first.Token && string(o.LogicalState) != "active" {
			t.Errorf("relay %s state = %v, want active", o.Token, o.LogicalState)
		}
	}

	inputs, err := client.DeviceIO().GetDigitalInputs(ctx)
	if err != nil {
		t.Fatalf("GetDigitalInputs: %v", err)
	}
	if len(inputs) == 0 {
		t.Error("simulator must expose at least one digital input")
	}
}

func TestDeviceIOUnknownRelayRejected(t *testing.T) {
	_, client, _ := startExpansionServer(t)

	err := client.DeviceIO().SetRelayOutputState(context.Background(), "no-such-relay", "active")
	if err == nil {
		t.Fatal("unknown relay token must fail")
	}
}

func TestMedia2Loopback(t *testing.T) {
	_, client, config := startExpansionServer(t)
	ctx := context.Background()

	profiles, err := client.Media2().GetProfiles(ctx, "", nil)
	if err != nil {
		t.Fatalf("Media2 GetProfiles: %v", err)
	}
	if len(profiles) != len(config.Profiles) {
		t.Fatalf("profiles = %d, want %d", len(profiles), len(config.Profiles))
	}
	if profiles[0].Token != config.Profiles[0].Token {
		t.Errorf("profile token = %q, want %q", profiles[0].Token, config.Profiles[0].Token)
	}
	if profiles[0].VideoEncoder == nil || profiles[0].VideoEncoder.Encoding == "" {
		t.Errorf("video encoder configuration missing: %+v", profiles[0].VideoEncoder)
	}

	uri, err := client.Media2().GetStreamUri(ctx, "rtsp", config.Profiles[0].Token)
	if err != nil {
		t.Fatalf("Media2 GetStreamUri: %v", err)
	}
	if uri == "" {
		t.Error("stream URI empty")
	}

	if err := client.Media2().SetSynchronizationPoint(ctx, config.Profiles[0].Token); err != nil {
		t.Fatalf("Media2 SetSynchronizationPoint: %v", err)
	}
}

func TestMedia2NotMountedWhenDisabled(t *testing.T) {
	// Default config keeps SupportMedia2 off — the endpoint 404s and
	// GetServices does not advertise the ver20/media namespace.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	config := createTestConfig()
	config.Port = listener.Addr().(*net.TCPAddr).Port

	s, err := New(config)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts := httptest.NewUnstartedServer(s.Handler())
	ts.Listener = listener
	ts.Start()
	t.Cleanup(ts.Close)

	client, err := onvif.NewClient(ts.URL, onvif.WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Initialize(context.Background()); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	services, err := client.Device().GetServices(context.Background(), false)
	if err != nil {
		t.Fatalf("GetServices: %v", err)
	}
	for _, svc := range services {
		if svc.Namespace == "http://www.onvif.org/ver20/media/wsdl" {
			t.Error("media2 advertised while SupportMedia2=false")
		}
	}
}

func TestMediaOSDAndAudioEmptySurfaces(t *testing.T) {
	_, client, _ := startExpansionServer(t)
	ctx := context.Background()

	// Audio family: valid empty sets, not ActionNotSupported faults.
	sources, err := client.Media().GetAudioSources(ctx)
	if err != nil {
		t.Fatalf("GetAudioSources: %v", err)
	}
	if len(sources) != 0 {
		t.Errorf("audio sources = %d, want 0 (no audio hardware)", len(sources))
	}

	outputs, err := client.Media().GetAudioOutputs(ctx)
	if err != nil {
		t.Fatalf("GetAudioOutputs: %v", err)
	}
	if len(outputs) != 0 {
		t.Errorf("audio outputs = %d, want 0", len(outputs))
	}
}

// Namespace-strict shape of the Media2 profile response: the
// Configurations wrapper and its children are tr2-local elements, the
// configuration bodies resolve to ver10/schema (the #93 ground-truth
// rules applied to the new face).
type nsMedia2Profiles struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/media/wsdl GetProfilesResponse"`
	Profile []struct {
		Token string `xml:"token,attr"`
		Name  string `xml:"http://www.onvif.org/ver10/schema Name"`
		Conf  struct {
			VideoSource struct {
				Token string `xml:"token,attr"`
			} `xml:"http://www.onvif.org/ver20/media/wsdl VideoSource"`
			VideoEncoder struct {
				Token      string `xml:"token,attr"`
				Encoding   string `xml:"http://www.onvif.org/ver10/schema Encoding"`
				Resolution struct {
					Width  int `xml:"http://www.onvif.org/ver10/schema Width"`
					Height int `xml:"http://www.onvif.org/ver10/schema Height"`
				} `xml:"http://www.onvif.org/ver10/schema Resolution"`
			} `xml:"http://www.onvif.org/ver20/media/wsdl VideoEncoder"`
		} `xml:"http://www.onvif.org/ver20/media/wsdl Configurations"`
	} `xml:"http://www.onvif.org/ver20/media/wsdl Profiles"`
}

func TestMedia2GetProfilesWireNamespaces(t *testing.T) {
	s, err := New(createTestConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	resp, err := s.HandleMedia2GetProfiles(&soap.RequestContext{RemoteIP: "192.0.2.9"}, testBody(t, &struct {
		XMLName xml.Name `xml:"tr2:GetProfiles"`
	}{}))
	if err != nil {
		t.Fatalf("HandleMedia2GetProfiles: %v", err)
	}

	raw, err := xml.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var parsed nsMedia2Profiles
	if err := xml.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("namespace-strict decode failed (wrong ns on a child?): %v\n%s", err, raw)
	}

	if len(parsed.Profile) == 0 {
		t.Fatalf("no profiles parsed:\n%s", raw)
	}
	p := parsed.Profile[0]
	if p.Name == "" || p.Conf.VideoEncoder.Encoding == "" || p.Conf.VideoEncoder.Resolution.Width == 0 {
		t.Errorf("profile fields did not survive ns-strict decode: %+v\n%s", p, raw)
	}
}

func TestOSDCreateSetDeleteLoop(t *testing.T) {
	_, client, cfg := startExpansionServer(t)
	ctx := context.Background()
	vscToken := cfg.Profiles[0].VideoSource.Token

	osds, err := client.Media().GetOSDs(ctx, "")
	if err != nil {
		t.Fatalf("GetOSDs: %v", err)
	}
	if len(osds) != 0 {
		t.Fatalf("fresh simulator OSDs = %d, want 0", len(osds))
	}

	created, err := client.Media().CreateOSD(ctx, vscToken, &media.OSDConfiguration{})
	if err != nil {
		t.Fatalf("CreateOSD: %v", err)
	}
	if created.Token == "" {
		t.Fatal("CreateOSD returned no token")
	}

	osds, err = client.Media().GetOSDs(ctx, "")
	if err != nil {
		t.Fatalf("GetOSDs after create: %v", err)
	}
	if len(osds) != 1 || osds[0].Token != created.Token {
		t.Fatalf("OSDs after create = %+v, want one with token %q", osds, created.Token)
	}

	if err := client.Media().SetOSD(ctx, created); err != nil {
		t.Fatalf("SetOSD: %v", err)
	}

	if err := client.Media().DeleteOSD(ctx, created.Token); err != nil {
		t.Fatalf("DeleteOSD: %v", err)
	}

	osds, err = client.Media().GetOSDs(ctx, "")
	if err != nil {
		t.Fatalf("GetOSDs after delete: %v", err)
	}
	if len(osds) != 0 {
		t.Fatalf("OSDs after delete = %d, want 0", len(osds))
	}
}
