package deviceio

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
)

// Issue #114: the DeviceIO audio-output family. (The audio *decoder*
// configuration family lives in the Media service per the WSDLs and is
// already covered by the media package.)

func TestGetAudioOutputsReturnsTokens(t *testing.T) {
	svc := New(testutil.NewFakeCaller("http://fake/deviceio", func(action, _ string) (string, error) {
		if action != "tmd:GetAudioOutputs" {
			return "", errors.New("unexpected action " + action)
		}

		// tmd:Get/GetResponse: a bare token list.
		return `<GetAudioOutputsResponse>
	<Token>AudioOutput_1</Token>
	<Token>AudioOutput_2</Token>
</GetAudioOutputsResponse>`, nil
	}))

	tokens, err := svc.GetAudioOutputs(context.Background())
	if err != nil {
		t.Fatalf("GetAudioOutputs: %v", err)
	}

	if len(tokens) != 2 || tokens[0] != "AudioOutput_1" || tokens[1] != "AudioOutput_2" {
		t.Errorf("tokens = %v", tokens)
	}
}

func TestGetAudioOutputConfigurationParsesFields(t *testing.T) {
	svc := New(testutil.NewFakeCaller("http://fake/deviceio", func(action, reqXML string) (string, error) {
		if action != "tmd:GetAudioOutputConfiguration" {
			return "", errors.New("unexpected action " + action)
		}
		if !strings.Contains(reqXML, "<tmd:AudioOutputToken>AudioOutput_1</tmd:AudioOutputToken>") {
			t.Errorf("request missing audio output token:\n%s", reqXML)
		}

		return `<GetAudioOutputConfigurationResponse>
	<AudioOutputConfiguration token="aoc-1">
		<Name>Speaker</Name><UseCount>1</UseCount>
		<SourceToken>AudioSource_1</SourceToken>
		<OutputLevel>50</OutputLevel>
		<SendPrimacy>www.onvif.org/ver20/HalfDuplex/HTTP</SendPrimacy>
	</AudioOutputConfiguration>
</GetAudioOutputConfigurationResponse>`, nil
	}))

	cfg, err := svc.GetAudioOutputConfiguration(context.Background(), "AudioOutput_1")
	if err != nil {
		t.Fatalf("GetAudioOutputConfiguration: %v", err)
	}

	if cfg.Token != "aoc-1" || cfg.Name != "Speaker" || cfg.SourceToken != "AudioSource_1" {
		t.Errorf("config identity = %+v", cfg)
	}
	if cfg.OutputLevel != 50 {
		t.Errorf("OutputLevel = %d, want 50", cfg.OutputLevel)
	}
	if !strings.Contains(cfg.SendPrimacy, "HalfDuplex") {
		t.Errorf("SendPrimacy = %q", cfg.SendPrimacy)
	}
}

func TestGetAudioOutputConfigurationOptionsParsesRanges(t *testing.T) {
	svc := New(testutil.NewFakeCaller("http://fake/deviceio", func(action, _ string) (string, error) {
		if action != "tmd:GetAudioOutputConfigurationOptions" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetAudioOutputConfigurationOptionsResponse>
	<AudioOutputOptions>
		<OutputLevelsRange><Min>0</Min><Max>100</Max></OutputLevelsRange>
		<SendPrimacyOptions>www.onvif.org/ver20/HalfDuplex/HTTP www.onvif.org/ver20/FullDuplex/HTTP</SendPrimacyOptions>
	</AudioOutputOptions>
</GetAudioOutputConfigurationOptionsResponse>`, nil
	}))

	opts, err := svc.GetAudioOutputConfigurationOptions(context.Background(), "AudioOutput_1")
	if err != nil {
		t.Fatalf("GetAudioOutputConfigurationOptions: %v", err)
	}

	if opts.OutputLevelsRange == nil || opts.OutputLevelsRange.Min != 0 || opts.OutputLevelsRange.Max != 100 {
		t.Errorf("OutputLevelsRange = %+v, want [0,100]", opts.OutputLevelsRange)
	}
	if len(opts.SendPrimacyOptions) != 2 {
		t.Errorf("SendPrimacyOptions = %v, want 2 entries", opts.SendPrimacyOptions)
	}
}

func TestSetAudioOutputConfigurationEncodesWire(t *testing.T) {
	svc := New(testutil.NewFakeCaller("http://fake/deviceio", func(action, reqXML string) (string, error) {
		if action != "tmd:SetAudioOutputConfiguration" {
			return "", errors.New("unexpected action " + action)
		}
		for _, want := range []string{
			"<tmd:Configuration token=\"aoc-1\">",
			"<tt:SourceToken>AudioSource_1</tt:SourceToken>",
			"<tt:OutputLevel>75</tt:OutputLevel>",
			"<tmd:ForcePersistence>true</tmd:ForcePersistence>",
		} {
			if !strings.Contains(reqXML, want) {
				t.Errorf("request missing %q:\n%s", want, reqXML)
			}
		}

		return "<SetAudioOutputConfigurationResponse/>", nil
	}))

	err := svc.SetAudioOutputConfiguration(context.Background(), &AudioOutputConfiguration{
		Token:       "aoc-1",
		Name:        "Speaker",
		SourceToken: "AudioSource_1",
		OutputLevel: 75,
	}, true)
	if err != nil {
		t.Fatalf("SetAudioOutputConfiguration: %v", err)
	}
}

func TestAudioOutputOpsOnFaultCapableDevice(t *testing.T) {
	// Devices without audio answer a Sender fault; the call must surface
	// the transport's typed fault error, distinguishable from a network
	// error (the issue's classification requirement).
	svc := New(testutil.NewFakeCaller("http://fake/deviceio", func(action, _ string) (string, error) {
		return "", errors.New("sender fault: ter:ActionNotSupported")
	}))

	_, err := svc.GetAudioOutputs(context.Background())
	if err == nil {
		t.Fatal("expected the fault to surface as an error")
	}
	if !strings.Contains(err.Error(), "ActionNotSupported") {
		t.Errorf("error = %v, want ActionNotSupported classification", err)
	}
}
