package imaging

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
	"github.com/mickeyzzc/onvif-go/v2/types"
)

const imagingSettingsResponse = `<GetImagingSettingsResponse>
	<ImagingSettings>
		<Brightness>52.5</Brightness>
		<Contrast>40</Contrast>
		<Sharpness>11</Sharpness>
		<IrCutFilter>AUTO</IrCutFilter>
		<BacklightCompensation><Mode>OFF</Mode><Level>0</Level></BacklightCompensation>
		<Exposure>
			<Mode>AUTO</Mode><Priority>FrameRate</Priority>
			<MinExposureTime>1</MinExposureTime><MaxExposureTime>10000</MaxExposureTime>
			<ExposureTime>120</ExposureTime><Gain>50</Gain>
		</Exposure>
		<Focus><AutoFocusMode>AUTO</AutoFocusMode><DefaultSpeed>0.5</DefaultSpeed></Focus>
		<WideDynamicRange><Mode>OFF</Mode><Level>0</Level></WideDynamicRange>
		<WhiteBalance><Mode>AUTO</Mode><CrGain>64</CrGain><CbGain>52</CbGain></WhiteBalance>
	</ImagingSettings>
</GetImagingSettingsResponse>`

const imagingOptionsResponse = `<GetOptionsResponse>
	<ImagingOptions>
		<Brightness><Min>0</Min><Max>100</Max></Brightness>
		<Contrast><Min>0</Min><Max>100</Max></Contrast>
		<BacklightCompensation><Mode>OFF</Mode><Mode>ON</Mode><Level><Min>0</Min><Max>100</Max></Level></BacklightCompensation>
		<Exposure><Mode>AUTO</Mode><Mode>MANUAL</Mode><Priority>LowNoise</Priority></Exposure>
		<Focus><AutoFocusModes>AUTO</AutoFocusModes><DefaultSpeed><Min>0</Min><Max>1</Max></DefaultSpeed></Focus>
	</ImagingOptions>
</GetOptionsResponse>`

const moveOptionsResponse = `<GetMoveOptionsResponse>
	<MoveOptions>
		<Absolute><Position><Min>0</Min><Max>1</Max></Position><Speed><Min>0</Min><Max>1</Max></Speed></Absolute>
		<Relative><Distance><Min>-1</Min><Max>1</Max></Distance></Relative>
		<Continuous><Speed><Min>-1</Min><Max>1</Max></Speed></Continuous>
	</MoveOptions>
</GetMoveOptionsResponse>`

const imagingStatusResponse = `<GetStatusResponse>
	<Status><FocusStatus><Position>0.5</Position><MoveStatus>IDLE</MoveStatus><Error>NO_ERROR</Error></FocusStatus></Status>
</GetStatusResponse>`

func TestGetImagingSettingsParses(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/imaging", func(action, reqXML string) (string, error) {
		if action != "timg:GetImagingSettings" {
			return "", errors.New("unexpected action " + action)
		}

		if !strings.Contains(reqXML, "video-source-1") {
			t.Errorf("request misses token: %s", reqXML)
		}

		return imagingSettingsResponse, nil
	})

	settings, err := New(caller).GetImagingSettings(context.Background(), "video-source-1")
	if err != nil {
		t.Fatalf("GetImagingSettings: %v", err)
	}

	if settings.Brightness == nil || *settings.Brightness != 52.5 {
		t.Errorf("Brightness = %v", settings.Brightness)
	}

	if settings.Exposure == nil || settings.Exposure.Mode != "AUTO" || settings.Exposure.MaxExposureTime != 10000 {
		t.Errorf("Exposure = %+v", settings.Exposure)
	}

	if settings.WhiteBalance == nil || settings.WhiteBalance.CrGain != 64 {
		t.Errorf("WhiteBalance = %+v", settings.WhiteBalance)
	}

	if settings.IrCutFilter == nil || *settings.IrCutFilter != "AUTO" {
		t.Errorf("IrCutFilter = %v", settings.IrCutFilter)
	}
}

func TestSetImagingSettingsSendsDelta(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/imaging", func(action, reqXML string) (string, error) {
		if action != "timg:SetImagingSettings" {
			return "", errors.New("unexpected action " + action)
		}

		for _, want := range []string{"video-source-2", "70", "AUTO"} {
			if !strings.Contains(reqXML, want) {
				t.Errorf("request misses %q: %s", want, reqXML)
			}
		}

		return "", nil
	})

	brightness := 70.0
	ircut := "AUTO"
	err := New(caller).SetImagingSettings(context.Background(), "video-source-2", &ImagingSettings{
		Brightness:  &brightness,
		IrCutFilter: &ircut,
	}, false)
	if err != nil {
		t.Fatalf("SetImagingSettings: %v", err)
	}
}

func TestMoveEncodesFocus(t *testing.T) {
	tests := []struct {
		name   string
		focus  *FocusMove
		wantIn []string
	}{
		{
			name:   "absolute",
			focus:  &FocusMove{Absolute: &AbsoluteFocus{Position: 0.75, Speed: ptrFloat64(0.5)}},
			wantIn: []string{"0.75"},
		},
		{
			name:   "relative",
			focus:  &FocusMove{Relative: &RelativeFocus{Distance: -0.1}},
			wantIn: []string{"-0.1"},
		},
		{
			name:   "continuous",
			focus:  &FocusMove{Continuous: &ContinuousFocus{Speed: 0.9}},
			wantIn: []string{"0.9"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caller := testutil.NewFakeCaller("http://fake/imaging", func(action, _ string) (string, error) {
				if action != "timg:Move" {
					return "", errors.New("unexpected action " + action)
				}

				return "", nil
			})

			if err := New(caller).Move(context.Background(), "vs-1", tt.focus); err != nil {
				t.Fatalf("Move: %v", err)
			}

			body := caller.Requests()[0].Body
			for _, want := range tt.wantIn {
				if !strings.Contains(body, want) {
					t.Errorf("Move body misses %q: %s", want, body)
				}
			}
		})
	}
}

func ptrFloat64(v float64) *float64 { return &v }

func TestGetOptionsParses(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/imaging", func(action, _ string) (string, error) {
		if action != "timg:GetOptions" {
			return "", errors.New("unexpected action " + action)
		}

		return imagingOptionsResponse, nil
	})

	options, err := New(caller).GetOptions(context.Background(), "vs-1")
	if err != nil {
		t.Fatalf("GetOptions: %v", err)
	}

	if options.Brightness == nil || options.Brightness.Min != 0 || options.Brightness.Max != 100 {
		t.Errorf("Brightness options = %+v", options.Brightness)
	}

	if options.Exposure == nil || len(options.Exposure.Mode) != 2 {
		t.Errorf("Exposure modes = %+v", options.Exposure)
	}
}

func TestGetMoveOptionsParses(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/imaging", func(action, _ string) (string, error) {
		if action != "timg:GetMoveOptions" {
			return "", errors.New("unexpected action " + action)
		}

		return moveOptionsResponse, nil
	})

	options, err := New(caller).GetMoveOptions(context.Background(), "vs-1")
	if err != nil {
		t.Fatalf("GetMoveOptions: %v", err)
	}

	if options.Absolute == nil || options.Absolute.Position.Max != 1 || options.Absolute.Speed.Max != 1 {
		t.Errorf("Absolute options = %+v", options.Absolute)
	}

	if options.Continuous == nil || options.Continuous.Speed.Min != -1 {
		t.Errorf("Continuous options = %+v", options.Continuous)
	}
}

func TestStopFocusAndStatus(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/imaging", func(action, _ string) (string, error) {
		switch action {
		case "timg:Stop":
			return "", nil
		case "timg:GetStatus":
			return imagingStatusResponse, nil
		default:
			return "", errors.New("unexpected action " + action)
		}
	})

	svc := New(caller)

	if err := svc.StopFocus(context.Background(), "vs-1"); err != nil {
		t.Fatalf("StopFocus: %v", err)
	}

	status, err := svc.GetImagingStatus(context.Background(), "vs-1")
	if err != nil {
		t.Fatalf("GetImagingStatus: %v", err)
	}

	if status.FocusStatus == nil || status.FocusStatus.MoveStatus != "IDLE" || status.FocusStatus.Position != 0.5 {
		t.Errorf("FocusStatus = %+v", status.FocusStatus)
	}
}

func TestImagingNotSupportedWithoutEndpoint(t *testing.T) {
	svc := New(testutil.NewFakeCaller("", func(string, string) (string, error) { return "", nil }))

	if _, err := svc.GetImagingSettings(context.Background(), "vs"); !errors.Is(err, types.ErrServiceNotSupported) {
		t.Errorf("GetImagingSettings error = %v, want ErrServiceNotSupported", err)
	}

	if err := svc.StopFocus(context.Background(), "vs"); !errors.Is(err, types.ErrServiceNotSupported) {
		t.Errorf("StopFocus error = %v, want ErrServiceNotSupported", err)
	}
}
