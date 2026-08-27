package ptz

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
	"github.com/mickeyzzc/onvif-go/v2/types"
)

// canned PTZ device responses.
const (
	statusResponse = `<GetStatusResponse>
	<PTZStatus>
		<Position>
			<PanTilt x="11.5" y="-20.25" space="http://www.onvif.org/ver10/tptz/PanTiltSpaces/PositionGenericSpace"/>
			<Zoom x="0.75" space="http://www.onvif.org/ver10/tptz/ZoomSpaces/PositionGenericSpace"/>
		</Position>
		<MoveStatus><PanTilt>MOVING</PanTilt><Zoom>IDLE</Zoom></MoveStatus>
		<Error>NO_ERROR</Error>
		<UtcTime>2026-08-28T10:00:00Z</UtcTime>
	</PTZStatus>
</GetStatusResponse>`

	presetsResponse = `<GetPresetsResponse>
	<Preset token="preset-1"><Name>Home</Name>
		<PTZPosition><PanTilt x="1" y="2"/><Zoom x="3"/></PTZPosition></Preset>
	<Preset token="preset-2"><Name>Gate</Name></Preset>
</GetPresetsResponse>`

	setPresetResponse = `<SetPresetResponse><PresetToken>new-preset-9</PresetToken></SetPresetResponse>`

	configurationResponse = `<GetConfigurationsResponse>
	<PTZConfiguration token="ptz-conf-1">
		<Name>Main PTZ</Name><UseCount>4</UseCount><NodeToken>ptz-node-1</NodeToken>
		<DefaultAbsolutePantTiltPositionSpace>http://example/abs</DefaultAbsolutePantTiltPositionSpace>
		<DefaultSpeed><PanTilt x="0.5" y="0.5"/><Zoom x="0.25"/></DefaultSpeed>
		<PanTiltLimits><Range><URI>http://example/pt</URI><XRange><Min>-180</Min><Max>180</Max></XRange><YRange><Min>-90</Min><Max>90</Max></YRange></Range></PanTiltLimits>
	</PTZConfiguration>
</GetConfigurationsResponse>`
)

func TestMovesEncodeSpeedAndPosition(t *testing.T) {
	tests := []struct {
		name    string
		action  string
		invoke  func(s *Service) error
		wantSub []string
	}{
		{
			name:   "ContinuousMove",
			action: "tptz:ContinuousMove",
			invoke: func(s *Service) error {
				timeout := "PT5S"

				return s.ContinuousMove(context.Background(), "profile-1", &PTZSpeed{
					PanTilt: &Vector2D{X: 0.5, Y: -0.5},
					Zoom:    &Vector1D{X: 0.25},
				}, &timeout)
			},
			wantSub: []string{`x="0.5"`, `y="-0.5"`, "PT5S", "profile-1"},
		},
		{
			name:   "AbsoluteMove",
			action: "tptz:AbsoluteMove",
			invoke: func(s *Service) error {
				return s.AbsoluteMove(context.Background(), "profile-1", &PTZVector{
					PanTilt: &Vector2D{X: 10, Y: 20},
					Zoom:    &Vector1D{X: 3},
				}, nil)
			},
			wantSub: []string{`x="10"`, `x="3"`, "profile-1"},
		},
		{
			name:   "RelativeMove",
			action: "tptz:RelativeMove",
			invoke: func(s *Service) error {
				return s.RelativeMove(context.Background(), "profile-1", &PTZVector{
					PanTilt: &Vector2D{X: -1, Y: 2},
				}, nil)
			},
			wantSub: []string{`x="-1"`, `y="2"`},
		},
		{
			name:    "StopAll",
			action:  "tptz:Stop",
			invoke:  func(s *Service) error { return s.Stop(context.Background(), "profile-1", true, true) },
			wantSub: []string{"profile-1"},
		},
		{
			name:    "StopPanTiltOnly",
			action:  "tptz:Stop",
			invoke:  func(s *Service) error { return s.Stop(context.Background(), "profile-1", true, false) },
			wantSub: []string{"profile-1"},
		},
		{
			name:   "GotoPreset",
			action: "tptz:GotoPreset",
			invoke: func(s *Service) error {
				return s.GotoPreset(context.Background(), "profile-1", "preset-1", nil)
			},
			wantSub: []string{"preset-1", "profile-1"},
		},
		{
			name:    "GotoHomePosition",
			action:  "tptz:GotoHomePosition",
			invoke:  func(s *Service) error { return s.GotoHomePosition(context.Background(), "profile-1", nil) },
			wantSub: []string{"profile-1"},
		},
		{
			name:    "SetHomePosition",
			action:  "tptz:SetHomePosition",
			invoke:  func(s *Service) error { return s.SetHomePosition(context.Background(), "profile-1") },
			wantSub: []string{"profile-1"},
		},
		{
			name:   "RemovePreset",
			action: "tptz:RemovePreset",
			invoke: func(s *Service) error {
				return s.RemovePreset(context.Background(), "profile-1", "preset-2")
			},
			wantSub: []string{"preset-2"},
		},
		{
			name:   "AddAudioOutput-style void op: SetVideoEncoder is media; here AddPreset via SetPreset",
			action: "tptz:SetPreset",
			invoke: func(s *Service) error {
				_, err := s.SetPreset(context.Background(), "profile-1", "Gate", "")

				return err
			},
			wantSub: []string{"Gate"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caller := testutil.NewFakeCaller("http://fake/ptz", func(action, _ string) (string, error) {
				if action != tt.action {
					return "", errors.New("unexpected action " + action)
				}

				if action == "tptz:SetPreset" {
					return setPresetResponse, nil
				}

				return "", nil
			})

			if err := tt.invoke(New(caller)); err != nil {
				t.Fatalf("op error: %v", err)
			}

			requests := caller.Requests()
			if len(requests) != 1 || requests[0].Action != tt.action {
				t.Fatalf("requests = %+v, want one %s", requests, tt.action)
			}

			for _, sub := range tt.wantSub {
				if !strings.Contains(requests[0].Body, sub) {
					t.Errorf("%s body misses %q: %s", tt.action, sub, requests[0].Body)
				}
			}
		})
	}
}

func TestGetStatusParses(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/ptz", func(action, _ string) (string, error) {
		if action != "tptz:GetStatus" {
			return "", errors.New("unexpected action " + action)
		}

		return statusResponse, nil
	})

	status, err := New(caller).GetStatus(context.Background(), "profile-1")
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}

	if status.Position == nil || status.Position.PanTilt == nil || status.Position.PanTilt.X != 11.5 || status.Position.PanTilt.Y != -20.25 {
		t.Errorf("pan/tilt not parsed: %+v", status.Position)
	}

	if status.Position.Zoom == nil || status.Position.Zoom.X != 0.75 {
		t.Errorf("zoom not parsed: %+v", status.Position.Zoom)
	}

	if status.MoveStatus == nil || status.MoveStatus.PanTilt != "MOVING" || status.MoveStatus.Zoom != "IDLE" {
		t.Errorf("move status not parsed: %+v", status.MoveStatus)
	}
}

func TestGetPresetsParses(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/ptz", func(action, _ string) (string, error) {
		if action != "tptz:GetPresets" {
			return "", errors.New("unexpected action " + action)
		}

		return presetsResponse, nil
	})

	presets, err := New(caller).GetPresets(context.Background(), "profile-1")
	if err != nil {
		t.Fatalf("GetPresets: %v", err)
	}

	if len(presets) != 2 {
		t.Fatalf("presets = %d, want 2", len(presets))
	}

	if presets[0].Token != "preset-1" || presets[0].Name != "Home" {
		t.Errorf("preset[0] = %+v", presets[0])
	}

	if presets[0].PTZPosition == nil || presets[0].PTZPosition.PanTilt == nil || presets[0].PTZPosition.PanTilt.X != 1 {
		t.Errorf("preset position not parsed: %+v", presets[0].PTZPosition)
	}

	if presets[1].PTZPosition != nil {
		t.Errorf("preset without position should stay nil: %+v", presets[1].PTZPosition)
	}
}

func TestSetPresetReturnsToken(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/ptz", func(action, _ string) (string, error) {
		if action != "tptz:SetPreset" {
			return "", errors.New("unexpected action " + action)
		}

		return setPresetResponse, nil
	})

	token, err := New(caller).SetPreset(context.Background(), "profile-1", "Gate", "")
	if err != nil {
		t.Fatalf("SetPreset: %v", err)
	}

	if token != "new-preset-9" {
		t.Errorf("token = %q, want new-preset-9", token)
	}
}

func TestGetConfigurationsParses(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/ptz", func(action, _ string) (string, error) {
		switch action {
		case "tptz:GetConfigurations":
			return configurationResponse, nil
		case "tptz:GetConfiguration":
			return strings.ReplaceAll(configurationResponse,
				"GetConfigurationsResponse", "GetConfigurationResponse"), nil
		default:
			return "", errors.New("unexpected action " + action)
		}
	})

	svc := New(caller)

	configs, err := svc.GetConfigurations(context.Background())
	if err != nil {
		t.Fatalf("GetConfigurations: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("configs = %d, want 1", len(configs))
	}

	cfg := configs[0]
	if cfg.Token != "ptz-conf-1" || cfg.Name != "Main PTZ" || cfg.NodeToken != "ptz-node-1" {
		t.Errorf("configuration header not parsed: %+v", cfg)
	}

	one, err := svc.GetConfiguration(context.Background(), "ptz-conf-1")
	if err != nil {
		t.Fatalf("GetConfiguration: %v", err)
	}

	if one.Token != "ptz-conf-1" {
		t.Errorf("GetConfiguration token = %q", one.Token)
	}
}

func TestServiceNotSupportedWithoutEndpoint(t *testing.T) {
	svc := New(testutil.NewFakeCaller("", func(string, string) (string, error) { return "", nil }))

	if err := svc.ContinuousMove(context.Background(), "p", nil, nil); !errors.Is(err, types.ErrServiceNotSupported) {
		t.Errorf("ContinuousMove error = %v, want ErrServiceNotSupported", err)
	}

	if _, err := svc.GetStatus(context.Background(), "p"); !errors.Is(err, types.ErrServiceNotSupported) {
		t.Errorf("GetStatus error = %v, want ErrServiceNotSupported", err)
	}
}

func TestConvertersNilSafety(t *testing.T) {
	if convertToPTZVectorXML(nil) != nil {
		t.Error("nil vector must convert to nil")
	}

	if convertToPTZSpeedXML(nil) != nil {
		t.Error("nil speed must convert to nil")
	}

	v := convertToPTZVectorXML(&PTZVector{PanTilt: &Vector2D{X: 1, Y: 2, Space: "s"}, Zoom: &Vector1D{X: 3}})
	if v.PanTilt == nil || v.PanTilt.Space != "s" || v.Zoom.X != 3 {
		t.Errorf("vector conversion = %+v", v)
	}
}

// compile-time interface check.
var _ api.Caller = (*testutil.FakeCaller)(nil)
