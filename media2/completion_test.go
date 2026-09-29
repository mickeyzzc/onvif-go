package media2_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
	"github.com/mickeyzzc/onvif-go/v2/media2"
	"github.com/mickeyzzc/onvif-go/v2/types"
)

// Issue #110: the Profile-T path — SetSynchronizationPoint,
// GetVideoEncoderInstances, GetSnapshotUri, multicast control, service
// capabilities, audio/metadata configuration families.

func newSvc(handler func(action, reqXML string) (string, error)) *media2.Service {
	return media2.New(testutil.NewFakeCaller("http://fake/media2", handler))
}

func TestSetSynchronizationPoint(t *testing.T) {
	var gotAction, gotBody string
	svc := newSvc(func(action, reqXML string) (string, error) {
		gotAction, gotBody = action, reqXML
		return "<SetSynchronizationPointResponse/>", nil
	})

	if err := svc.SetSynchronizationPoint(context.Background(), "profile-1"); err != nil {
		t.Fatalf("SetSynchronizationPoint: %v", err)
	}

	if gotAction != "tr2:SetSynchronizationPoint" {
		t.Errorf("action = %q", gotAction)
	}
	if !strings.Contains(gotBody, "<tr2:ProfileToken>profile-1</tr2:ProfileToken>") {
		t.Errorf("body missing profile token:\n%s", gotBody)
	}
}

func TestGetVideoEncoderInstancesParsesCodecBreakdown(t *testing.T) {
	svc := newSvc(func(action, _ string) (string, error) {
		if action != "tr2:GetVideoEncoderInstances" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetVideoEncoderInstancesResponse>
	<Info>
		<Codec><Encoding>H264</Encoding><Number>2</Number></Codec>
		<Codec><Encoding>H265</Encoding><Number>1</Number></Codec>
		<Total>3</Total>
	</Info>
</GetVideoEncoderInstancesResponse>`, nil
	})

	info, err := svc.GetVideoEncoderInstances(context.Background(), "vsc-1")
	if err != nil {
		t.Fatalf("GetVideoEncoderInstances: %v", err)
	}

	if info.Total != 3 {
		t.Errorf("Total = %d, want 3", info.Total)
	}
	if len(info.Codecs) != 2 || info.Codecs[0].Encoding != "H264" || info.Codecs[0].Number != 2 {
		t.Errorf("codecs = %+v", info.Codecs)
	}
}

func TestGetSnapshotUriReturnsPlainUri(t *testing.T) {
	svc := newSvc(func(action, _ string) (string, error) {
		if action != "tr2:GetSnapshotUri" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetSnapshotUriResponse><Uri>http://192.168.1.10/onvif-http/snapshot?Profile_1</Uri></GetSnapshotUriResponse>`, nil
	})

	uri, err := svc.GetSnapshotUri(context.Background(), "Profile_1")
	if err != nil {
		t.Fatalf("GetSnapshotUri: %v", err)
	}

	if !strings.HasPrefix(uri, "http://192.168.1.10/onvif-http/snapshot") {
		t.Errorf("uri = %q", uri)
	}
}

func TestMulticastStreamingControl(t *testing.T) {
	for _, tc := range []struct {
		name   string
		action string
		call   func(s *media2.Service) error
	}{
		{
			name:   "start",
			action: "tr2:StartMulticastStreaming",
			call: func(s *media2.Service) error {
				return s.StartMulticastStreaming(context.Background(), "profile-1")
			},
		},
		{
			name:   "stop",
			action: "tr2:StopMulticastStreaming",
			call: func(s *media2.Service) error {
				return s.StopMulticastStreaming(context.Background(), "profile-1")
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := newSvc(func(action, reqXML string) (string, error) {
				if action != tc.action {
					return "", errors.New("unexpected action " + action)
				}
				if !strings.Contains(reqXML, "<tr2:ProfileToken>profile-1</tr2:ProfileToken>") {
					t.Errorf("body missing profile token:\n%s", reqXML)
				}

				return "<" + strings.TrimPrefix(tc.action, "tr2:") + "Response/>", nil
			})

			if err := tc.call(svc); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
		})
	}
}

func TestGetServiceCapabilitiesParsesFlags(t *testing.T) {
	svc := newSvc(func(action, _ string) (string, error) {
		if action != "tr2:GetServiceCapabilities" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetServiceCapabilitiesResponse>
	<Capabilities SnapshotUri="true" Rotation="false" OSD="true">
		<ProfileCapabilities MaximumNumberOfProfiles="16"/>
		<StreamingCapabilities RTSPStreaming="true"/>
	</Capabilities>
</GetServiceCapabilitiesResponse>`, nil
	})

	caps, err := svc.GetServiceCapabilities(context.Background())
	if err != nil {
		t.Fatalf("GetServiceCapabilities: %v", err)
	}

	if !caps.SnapshotUri || caps.Rotation || !caps.OSD {
		t.Errorf("flags = snapshot %v rotation %v osd %v", caps.SnapshotUri, caps.Rotation, caps.OSD)
	}
	if caps.MaximumNumberOfProfiles != 16 {
		t.Errorf("MaximumNumberOfProfiles = %d, want 16", caps.MaximumNumberOfProfiles)
	}
	if !caps.RTSPStreaming {
		t.Error("RTSPStreaming = false, want true")
	}
}

const audioEncoderConfigurationsResponse = `<GetAudioEncoderConfigurationsResponse>
	<Configurations token="aec-1">
		<Name>Audio in</Name><UseCount>1</UseCount>
		<Encoding>audio/PCMA</Encoding>
		<Bitrate>64</Bitrate>
		<SampleRate>8</SampleRate>
	</Configurations>
</GetAudioEncoderConfigurationsResponse>`

func TestGetAudioEncoderConfigurations(t *testing.T) {
	svc := newSvc(func(action, reqXML string) (string, error) {
		if action != "tr2:GetAudioEncoderConfigurations" {
			return "", errors.New("unexpected action " + action)
		}
		if !strings.Contains(reqXML, "<tr2:ConfigurationToken>aec-1</tr2:ConfigurationToken>") {
			t.Errorf("body missing configuration token filter:\n%s", reqXML)
		}

		return audioEncoderConfigurationsResponse, nil
	})

	cfgs, err := svc.GetAudioEncoderConfigurations(context.Background(), "aec-1")
	if err != nil {
		t.Fatalf("GetAudioEncoderConfigurations: %v", err)
	}

	if len(cfgs) != 1 {
		t.Fatalf("configurations = %d, want 1", len(cfgs))
	}
	cfg := cfgs[0]
	if cfg.Token != "aec-1" || cfg.Encoding != "audio/PCMA" || cfg.Bitrate != 64 || cfg.SampleRate != 8 {
		t.Errorf("configuration = %+v", cfg)
	}
}

const metadataConfigurationsResponse = `<GetMetadataConfigurationsResponse>
	<Configurations token="mdc-1">
		<Name>Metadata</Name><UseCount>2</UseCount>
		<PTZStatus><Status>true</Status><Position>true</Position></PTZStatus>
		<Events><Filter/></Events>
		<Analytics>false</Analytics>
	</Configurations>
</GetMetadataConfigurationsResponse>`

func TestGetMetadataConfigurations(t *testing.T) {
	svc := newSvc(func(action, _ string) (string, error) {
		if action != "tr2:GetMetadataConfigurations" {
			return "", errors.New("unexpected action " + action)
		}

		return metadataConfigurationsResponse, nil
	})

	cfgs, err := svc.GetMetadataConfigurations(context.Background(), "")
	if err != nil {
		t.Fatalf("GetMetadataConfigurations: %v", err)
	}

	if len(cfgs) != 1 {
		t.Fatalf("configurations = %d, want 1", len(cfgs))
	}
	cfg := cfgs[0]
	if cfg.Token != "mdc-1" || cfg.Name != "Metadata" {
		t.Errorf("configuration = %+v", cfg)
	}
	if cfg.PTZStatus == nil || !cfg.PTZStatus.Status || !cfg.PTZStatus.Position {
		t.Errorf("PTZStatus = %+v, want status+position", cfg.PTZStatus)
	}
	if cfg.Analytics {
		t.Error("Analytics = true, want false")
	}
}

func TestMedia2EndpointConstant(t *testing.T) {
	// The service identifier the client wires endpoints for.
	if !strings.Contains(media2.Namespace, "ver20/media/wsdl") {
		t.Errorf("Namespace = %q", media2.Namespace)
	}
	_ = api.ServiceMedia2
	_ = types.ErrServiceNotSupported
}
