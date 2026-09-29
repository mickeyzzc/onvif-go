package imaging

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
	"github.com/mickeyzzc/onvif-go/v2/types"
)

// Issue #113: spec-named Stop, GetServiceCapabilities, imaging presets.

func TestGetImagingServiceCapabilities(t *testing.T) {
	svc := New(testutil.NewFakeCaller("http://fake/imaging", func(action, _ string) (string, error) {
		if action != "timg:GetServiceCapabilities" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetServiceCapabilitiesResponse>
	<Capabilities ImageStabilization="false" Presets="true" AdaptablePreset="false"/>
</GetServiceCapabilitiesResponse>`, nil
	}))

	caps, err := svc.GetImagingServiceCapabilities(context.Background())
	if err != nil {
		t.Fatalf("GetImagingServiceCapabilities: %v", err)
	}

	if caps.ImageStabilization || !caps.Presets || caps.AdaptablePreset {
		t.Errorf("caps = %+v", caps)
	}
}

func TestStopIsTheSpecNamedFocusStop(t *testing.T) {
	var gotAction, gotBody string
	svc := New(testutil.NewFakeCaller("http://fake/imaging", func(action, reqXML string) (string, error) {
		gotAction, gotBody = action, reqXML
		return "<StopResponse/>", nil
	}))

	if err := svc.Stop(context.Background(), "src-1"); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	if gotAction != "timg:Stop" {
		t.Errorf("action = %q, want timg:Stop (the spec op; iris/zoom stop does not exist in ver20 imaging)", gotAction)
	}
	if !strings.Contains(gotBody, "<timg:VideoSourceToken>src-1</timg:VideoSourceToken>") {
		t.Errorf("body missing video source token:\n%s", gotBody)
	}
}

func TestGetPresetsParsesPresetList(t *testing.T) {
	svc := New(testutil.NewFakeCaller("http://fake/imaging", func(action, _ string) (string, error) {
		if action != "timg:GetPresets" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetPresetsResponse>
	<Preset token="preset-1"><Name>Day mode</Name><Type>Custom</Type></Preset>
	<Preset token="preset-2"><Name>Night mode</Name><Type>Night</Type></Preset>
</GetPresetsResponse>`, nil
	}))

	presets, err := svc.GetPresets(context.Background(), "src-1")
	if err != nil {
		t.Fatalf("GetPresets: %v", err)
	}

	if len(presets) != 2 {
		t.Fatalf("presets = %d, want 2", len(presets))
	}
	if presets[0].Token != "preset-1" || presets[0].Name != "Day mode" || presets[0].Type != "Custom" {
		t.Errorf("preset[0] = %+v", presets[0])
	}
}

func TestSetCurrentPreset(t *testing.T) {
	svc := New(testutil.NewFakeCaller("http://fake/imaging", func(action, reqXML string) (string, error) {
		if action != "timg:SetCurrentPreset" {
			return "", errors.New("unexpected action " + action)
		}
		for _, want := range []string{
			"<timg:VideoSourceToken>src-1</timg:VideoSourceToken>",
			"<timg:PresetToken>preset-2</timg:PresetToken>",
		} {
			if !strings.Contains(reqXML, want) {
				t.Errorf("request missing %q:\n%s", want, reqXML)
			}
		}

		return "<SetCurrentPresetResponse/>", nil
	}))

	if err := svc.SetCurrentPreset(context.Background(), "src-1", "preset-2"); err != nil {
		t.Fatalf("SetCurrentPreset: %v", err)
	}
}

func TestCompletionOpsNotSupportedWithoutEndpoint(t *testing.T) {
	svc := New(testutil.NewFakeCaller("", func(string, string) (string, error) { return "", nil }))

	if _, err := svc.GetImagingServiceCapabilities(context.Background()); !errors.Is(err, types.ErrServiceNotSupported) {
		t.Errorf("GetImagingServiceCapabilities error = %v", err)
	}

	if err := svc.Stop(context.Background(), "src-1"); !errors.Is(err, types.ErrServiceNotSupported) {
		t.Errorf("Stop error = %v", err)
	}

	if _, err := svc.GetPresets(context.Background(), "src-1"); !errors.Is(err, types.ErrServiceNotSupported) {
		t.Errorf("GetPresets error = %v", err)
	}

	if err := svc.SetCurrentPreset(context.Background(), "src-1", "p"); !errors.Is(err, types.ErrServiceNotSupported) {
		t.Errorf("SetCurrentPreset error = %v", err)
	}
}
