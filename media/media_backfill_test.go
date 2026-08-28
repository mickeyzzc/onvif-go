package media

// Wire-contract backfill for the media operations that had zero
// coverage: profiles/analytics management, encoder/source modes, audio
// chains, OSD, and multicast/synchronization controls — table-driven
// through internal/testutil.FakeCaller (trt: prefix).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
)

func newMediaOpsService(t *testing.T, wantAction, resp string) (*Service, *testutil.FakeCaller) {
	t.Helper()

	caller := testutil.NewFakeCaller("http://fake/media", func(action, _ string) (string, error) {
		if strings.TrimPrefix(action, "trt:") != wantAction {
			return "", errors.New("unexpected action " + action)
		}

		return resp, nil
	})

	return New(caller), caller
}

func TestMediaOpsWireContract(t *testing.T) {
	cases := []struct {
		name   string
		action string
		resp   string
		invoke func(s *Service) error
	}{
		// --- stream controls (media_stream.go) ---
		{"GetStreamURI", "GetStreamUri", `<GetStreamUriResponse><MediaUri><Uri>rtsp://x/main</Uri></MediaUri></GetStreamUriResponse>`, func(s *Service) error {
			_, err := s.GetStreamURI(context.Background(), "profile-1")

			return err
		}},
		{"GetStreamURIWithOptions", "GetStreamUri", `<GetStreamUriResponse><MediaUri><Uri>rtsp://x/main</Uri></MediaUri></GetStreamUriResponse>`, func(s *Service) error {
			_, err := s.GetStreamURIWithOptions(context.Background(), "profile-1", StreamSetup{Stream: "RTP-Unicast"})

			return err
		}},
		{"GetSnapshotURI", "GetSnapshotUri", `<GetSnapshotUriResponse><MediaUri><Uri>http://x/snap</Uri></MediaUri></GetSnapshotUriResponse>`, func(s *Service) error {
			_, err := s.GetSnapshotURI(context.Background(), "profile-1")

			return err
		}},
		{"SetSynchronizationPoint", "SetSynchronizationPoint", `<SetSynchronizationPointResponse/>`, func(s *Service) error {
			return s.SetSynchronizationPoint(context.Background(), "profile-1")
		}},
		{"StartMulticastStreaming", "StartMulticastStreaming", `<StartMulticastStreamingResponse/>`, func(s *Service) error {
			return s.StartMulticastStreaming(context.Background(), "profile-1")
		}},
		{"StopMulticastStreaming", "StopMulticastStreaming", `<StopMulticastStreamingResponse/>`, func(s *Service) error {
			return s.StopMulticastStreaming(context.Background(), "profile-1")
		}},

		// --- encoder / video source (media_encoder.go) ---
		{"GetVideoEncoderConfiguration", "GetVideoEncoderConfiguration", `<GetVideoEncoderConfigurationResponse/>`, func(s *Service) error {
			_, err := s.GetVideoEncoderConfiguration(context.Background(), "enc-1")

			return err
		}},
		{"GetVideoSources", "GetVideoSources", `<GetVideoSourcesResponse/>`, func(s *Service) error {
			_, err := s.GetVideoSources(context.Background())

			return err
		}},
		{"SetVideoEncoderConfiguration", "SetVideoEncoderConfiguration", `<SetVideoEncoderConfigurationResponse/>`, func(s *Service) error {
			return s.SetVideoEncoderConfiguration(context.Background(), &VideoEncoderConfiguration{Token: "enc-1"}, false)
		}},
		{"GetVideoEncoderConfigurationOptions", "GetVideoEncoderConfigurationOptions", `<GetVideoEncoderConfigurationOptionsResponse/>`, func(s *Service) error {
			_, err := s.GetVideoEncoderConfigurationOptions(context.Background(), "enc-1")

			return err
		}},
		{"GetVideoSourceModes", "GetVideoSourceModes", `<GetVideoSourceModesResponse/>`, func(s *Service) error {
			_, err := s.GetVideoSourceModes(context.Background(), "src-1")

			return err
		}},
		{"SetVideoSourceMode", "SetVideoSourceMode", `<SetVideoSourceModeResponse/>`, func(s *Service) error {
			return s.SetVideoSourceMode(context.Background(), "src-1", "mode-1")
		}},
		{"AddVideoEncoderConfiguration", "AddVideoEncoderConfiguration", `<AddVideoEncoderConfigurationResponse/>`, func(s *Service) error {
			return s.AddVideoEncoderConfiguration(context.Background(), "profile-1", "enc-1")
		}},
		{"RemoveVideoEncoderConfiguration", "RemoveVideoEncoderConfiguration", `<RemoveVideoEncoderConfigurationResponse/>`, func(s *Service) error {
			return s.RemoveVideoEncoderConfiguration(context.Background(), "profile-1")
		}},
		{"AddVideoSourceConfiguration", "AddVideoSourceConfiguration", `<AddVideoSourceConfigurationResponse/>`, func(s *Service) error {
			return s.AddVideoSourceConfiguration(context.Background(), "profile-1", "vsc-1")
		}},
		{"RemoveVideoSourceConfiguration", "RemoveVideoSourceConfiguration", `<RemoveVideoSourceConfigurationResponse/>`, func(s *Service) error {
			return s.RemoveVideoSourceConfiguration(context.Background(), "profile-1")
		}},
		{"GetGuaranteedNumberOfVideoEncoderInstances", "GetGuaranteedNumberOfVideoEncoderInstances", `<GetGuaranteedNumberOfVideoEncoderInstancesResponse/>`, func(s *Service) error {
			_, err := s.GetGuaranteedNumberOfVideoEncoderInstances(context.Background(), "src-1")

			return err
		}},
		{"GetVideoSourceConfigurations", "GetVideoSourceConfigurations", `<GetVideoSourceConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetVideoSourceConfigurations(context.Background())

			return err
		}},
		{"GetVideoEncoderConfigurations", "GetVideoEncoderConfigurations", `<GetVideoEncoderConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetVideoEncoderConfigurations(context.Background())

			return err
		}},
		{"GetVideoSourceConfiguration", "GetVideoSourceConfiguration", `<GetVideoSourceConfigurationResponse/>`, func(s *Service) error {
			_, err := s.GetVideoSourceConfiguration(context.Background(), "vsc-1")

			return err
		}},
		{"GetVideoSourceConfigurationOptions", "GetVideoSourceConfigurationOptions", `<GetVideoSourceConfigurationOptionsResponse/>`, func(s *Service) error {
			_, err := s.GetVideoSourceConfigurationOptions(context.Background(), "vsc-1", "profile-1")

			return err
		}},
		{"SetVideoSourceConfiguration", "SetVideoSourceConfiguration", `<SetVideoSourceConfigurationResponse/>`, func(s *Service) error {
			return s.SetVideoSourceConfiguration(context.Background(), &VideoSourceConfiguration{Token: "vsc-1"}, false)
		}},
		{"GetCompatibleVideoEncoderConfigurations", "GetCompatibleVideoEncoderConfigurations", `<GetCompatibleVideoEncoderConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetCompatibleVideoEncoderConfigurations(context.Background(), "profile-1")

			return err
		}},
		{"GetCompatibleVideoSourceConfigurations", "GetCompatibleVideoSourceConfigurations", `<GetCompatibleVideoSourceConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetCompatibleVideoSourceConfigurations(context.Background(), "profile-1")

			return err
		}},

		// --- profiles / analytics (media_profiles.go) ---
		{"GetProfiles", "GetProfiles", `<GetProfilesResponse/>`, func(s *Service) error {
			_, err := s.GetProfiles(context.Background())

			return err
		}},
		{"CreateProfile", "CreateProfile", `<CreateProfileResponse/>`, func(s *Service) error {
			_, err := s.CreateProfile(context.Background(), "main", "token-1")

			return err
		}},
		{"DeleteProfile", "DeleteProfile", `<DeleteProfileResponse/>`, func(s *Service) error {
			return s.DeleteProfile(context.Background(), "profile-1")
		}},
		{"GetMediaServiceCapabilities", "GetServiceCapabilities", `<GetServiceCapabilitiesResponse/>`, func(s *Service) error {
			_, err := s.GetMediaServiceCapabilities(context.Background())

			return err
		}},
		{"GetProfile", "GetProfile", `<GetProfileResponse/>`, func(s *Service) error {
			_, err := s.GetProfile(context.Background(), "profile-1")

			return err
		}},
		{"SetProfile", "SetProfile", `<SetProfileResponse/>`, func(s *Service) error {
			return s.SetProfile(context.Background(), &Profile{Token: "profile-1"})
		}},
		{"AddPTZConfiguration", "AddPTZConfiguration", `<AddPTZConfigurationResponse/>`, func(s *Service) error {
			return s.AddPTZConfiguration(context.Background(), "profile-1", "ptz-1")
		}},
		{"RemovePTZConfiguration", "RemovePTZConfiguration", `<RemovePTZConfigurationResponse/>`, func(s *Service) error {
			return s.RemovePTZConfiguration(context.Background(), "profile-1")
		}},
		{"GetCompatiblePTZConfigurations", "GetCompatiblePTZConfigurations", `<GetCompatiblePTZConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetCompatiblePTZConfigurations(context.Background(), "profile-1")

			return err
		}},
		{"GetVideoAnalyticsConfigurations", "GetVideoAnalyticsConfigurations", `<GetVideoAnalyticsConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetVideoAnalyticsConfigurations(context.Background())

			return err
		}},
		{"GetVideoAnalyticsConfiguration", "GetVideoAnalyticsConfiguration", `<GetVideoAnalyticsConfigurationResponse/>`, func(s *Service) error {
			_, err := s.GetVideoAnalyticsConfiguration(context.Background(), "vac-1")

			return err
		}},
		{"GetCompatibleVideoAnalyticsConfigurations", "GetCompatibleVideoAnalyticsConfigurations", `<GetCompatibleVideoAnalyticsConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetCompatibleVideoAnalyticsConfigurations(context.Background(), "profile-1")

			return err
		}},
		{"SetVideoAnalyticsConfiguration", "SetVideoAnalyticsConfiguration", `<SetVideoAnalyticsConfigurationResponse/>`, func(s *Service) error {
			return s.SetVideoAnalyticsConfiguration(context.Background(), &VideoAnalyticsConfiguration{Token: "vac-1"}, false)
		}},
		{"GetVideoAnalyticsConfigurationOptions", "GetVideoAnalyticsConfigurationOptions", `<GetVideoAnalyticsConfigurationOptionsResponse/>`, func(s *Service) error {
			_, err := s.GetVideoAnalyticsConfigurationOptions(context.Background(), "vac-1", "profile-1")

			return err
		}},
		{"AddVideoAnalyticsConfiguration", "AddVideoAnalyticsConfiguration", `<AddVideoAnalyticsConfigurationResponse/>`, func(s *Service) error {
			return s.AddVideoAnalyticsConfiguration(context.Background(), "profile-1", "vac-1")
		}},
		{"RemoveVideoAnalyticsConfiguration", "RemoveVideoAnalyticsConfiguration", `<RemoveVideoAnalyticsConfigurationResponse/>`, func(s *Service) error {
			return s.RemoveVideoAnalyticsConfiguration(context.Background(), "profile-1")
		}},

		// --- audio chains (media_audio.go) ---
		{"GetAudioSources", "GetAudioSources", `<GetAudioSourcesResponse/>`, func(s *Service) error {
			_, err := s.GetAudioSources(context.Background())

			return err
		}},
		{"GetAudioOutputs", "GetAudioOutputs", `<GetAudioOutputsResponse/>`, func(s *Service) error {
			_, err := s.GetAudioOutputs(context.Background())

			return err
		}},
		{"GetAudioEncoderConfiguration", "GetAudioEncoderConfiguration", `<GetAudioEncoderConfigurationResponse/>`, func(s *Service) error {
			_, err := s.GetAudioEncoderConfiguration(context.Background(), "aec-1")

			return err
		}},
		{"SetAudioEncoderConfiguration", "SetAudioEncoderConfiguration", `<SetAudioEncoderConfigurationResponse/>`, func(s *Service) error {
			return s.SetAudioEncoderConfiguration(context.Background(), &AudioEncoderConfiguration{Token: "aec-1"}, false)
		}},
		{"GetMetadataConfiguration", "GetMetadataConfiguration", `<GetMetadataConfigurationResponse/>`, func(s *Service) error {
			_, err := s.GetMetadataConfiguration(context.Background(), "mc-1")

			return err
		}},
		{"SetMetadataConfiguration", "SetMetadataConfiguration", `<SetMetadataConfigurationResponse/>`, func(s *Service) error {
			return s.SetMetadataConfiguration(context.Background(), &MetadataConfiguration{Token: "mc-1"}, false)
		}},
		{"AddAudioEncoderConfiguration", "AddAudioEncoderConfiguration", `<AddAudioEncoderConfigurationResponse/>`, func(s *Service) error {
			return s.AddAudioEncoderConfiguration(context.Background(), "profile-1", "aec-1")
		}},
		{"RemoveAudioEncoderConfiguration", "RemoveAudioEncoderConfiguration", `<RemoveAudioEncoderConfigurationResponse/>`, func(s *Service) error {
			return s.RemoveAudioEncoderConfiguration(context.Background(), "profile-1")
		}},
		{"AddAudioSourceConfiguration", "AddAudioSourceConfiguration", `<AddAudioSourceConfigurationResponse/>`, func(s *Service) error {
			return s.AddAudioSourceConfiguration(context.Background(), "profile-1", "asc-1")
		}},
		{"RemoveAudioSourceConfiguration", "RemoveAudioSourceConfiguration", `<RemoveAudioSourceConfigurationResponse/>`, func(s *Service) error {
			return s.RemoveAudioSourceConfiguration(context.Background(), "profile-1")
		}},
		{"AddMetadataConfiguration", "AddMetadataConfiguration", `<AddMetadataConfigurationResponse/>`, func(s *Service) error {
			return s.AddMetadataConfiguration(context.Background(), "profile-1", "mc-1")
		}},
		{"RemoveMetadataConfiguration", "RemoveMetadataConfiguration", `<RemoveMetadataConfigurationResponse/>`, func(s *Service) error {
			return s.RemoveMetadataConfiguration(context.Background(), "profile-1")
		}},
		{"GetAudioEncoderConfigurationOptions", "GetAudioEncoderConfigurationOptions", `<GetAudioEncoderConfigurationOptionsResponse/>`, func(s *Service) error {
			_, err := s.GetAudioEncoderConfigurationOptions(context.Background(), "aec-1", "profile-1")

			return err
		}},
		{"GetMetadataConfigurationOptions", "GetMetadataConfigurationOptions", `<GetMetadataConfigurationOptionsResponse/>`, func(s *Service) error {
			_, err := s.GetMetadataConfigurationOptions(context.Background(), "mc-1", "profile-1")

			return err
		}},
		{"GetAudioOutputConfiguration", "GetAudioOutputConfiguration", `<GetAudioOutputConfigurationResponse/>`, func(s *Service) error {
			_, err := s.GetAudioOutputConfiguration(context.Background(), "aoc-1")

			return err
		}},
		{"SetAudioOutputConfiguration", "SetAudioOutputConfiguration", `<SetAudioOutputConfigurationResponse/>`, func(s *Service) error {
			return s.SetAudioOutputConfiguration(context.Background(), &AudioOutputConfiguration{Token: "aoc-1"}, false)
		}},
		{"GetAudioOutputConfigurationOptions", "GetAudioOutputConfigurationOptions", `<GetAudioOutputConfigurationOptionsResponse/>`, func(s *Service) error {
			_, err := s.GetAudioOutputConfigurationOptions(context.Background(), "aoc-1")

			return err
		}},
		{"GetAudioDecoderConfigurationOptions", "GetAudioDecoderConfigurationOptions", `<GetAudioDecoderConfigurationOptionsResponse/>`, func(s *Service) error {
			_, err := s.GetAudioDecoderConfigurationOptions(context.Background(), "adc-1")

			return err
		}},
		{"GetAudioSourceConfigurations", "GetAudioSourceConfigurations", `<GetAudioSourceConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetAudioSourceConfigurations(context.Background())

			return err
		}},
		{"GetAudioEncoderConfigurations", "GetAudioEncoderConfigurations", `<GetAudioEncoderConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetAudioEncoderConfigurations(context.Background())

			return err
		}},
		{"GetAudioSourceConfiguration", "GetAudioSourceConfiguration", `<GetAudioSourceConfigurationResponse/>`, func(s *Service) error {
			_, err := s.GetAudioSourceConfiguration(context.Background(), "asc-1")

			return err
		}},
		{"GetAudioSourceConfigurationOptions", "GetAudioSourceConfigurationOptions", `<GetAudioSourceConfigurationOptionsResponse/>`, func(s *Service) error {
			_, err := s.GetAudioSourceConfigurationOptions(context.Background(), "asc-1", "profile-1")

			return err
		}},
		{"SetAudioSourceConfiguration", "SetAudioSourceConfiguration", `<SetAudioSourceConfigurationResponse/>`, func(s *Service) error {
			return s.SetAudioSourceConfiguration(context.Background(), &AudioSourceConfiguration{Token: "asc-1"}, false)
		}},
		{"GetCompatibleAudioEncoderConfigurations", "GetCompatibleAudioEncoderConfigurations", `<GetCompatibleAudioEncoderConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetCompatibleAudioEncoderConfigurations(context.Background(), "profile-1")

			return err
		}},
		{"GetCompatibleAudioSourceConfigurations", "GetCompatibleAudioSourceConfigurations", `<GetCompatibleAudioSourceConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetCompatibleAudioSourceConfigurations(context.Background(), "profile-1")

			return err
		}},
		{"GetCompatibleMetadataConfigurations", "GetCompatibleMetadataConfigurations", `<GetCompatibleMetadataConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetCompatibleMetadataConfigurations(context.Background(), "profile-1")

			return err
		}},
		{"GetCompatibleAudioOutputConfigurations", "GetCompatibleAudioOutputConfigurations", `<GetCompatibleAudioOutputConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetCompatibleAudioOutputConfigurations(context.Background(), "profile-1")

			return err
		}},
		{"GetCompatibleAudioDecoderConfigurations", "GetCompatibleAudioDecoderConfigurations", `<GetCompatibleAudioDecoderConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetCompatibleAudioDecoderConfigurations(context.Background(), "profile-1")

			return err
		}},
		{"GetMetadataConfigurations", "GetMetadataConfigurations", `<GetMetadataConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetMetadataConfigurations(context.Background())

			return err
		}},
		{"GetAudioOutputConfigurations", "GetAudioOutputConfigurations", `<GetAudioOutputConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetAudioOutputConfigurations(context.Background())

			return err
		}},
		{"GetAudioDecoderConfigurations", "GetAudioDecoderConfigurations", `<GetAudioDecoderConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetAudioDecoderConfigurations(context.Background())

			return err
		}},
		{"GetAudioDecoderConfiguration", "GetAudioDecoderConfiguration", `<GetAudioDecoderConfigurationResponse/>`, func(s *Service) error {
			_, err := s.GetAudioDecoderConfiguration(context.Background(), "adc-1")

			return err
		}},
		{"SetAudioDecoderConfiguration", "SetAudioDecoderConfiguration", `<SetAudioDecoderConfigurationResponse/>`, func(s *Service) error {
			return s.SetAudioDecoderConfiguration(context.Background(), &AudioDecoderConfiguration{Token: "adc-1"}, false)
		}},
		{"AddAudioOutputConfiguration", "AddAudioOutputConfiguration", `<AddAudioOutputConfigurationResponse/>`, func(s *Service) error {
			return s.AddAudioOutputConfiguration(context.Background(), "profile-1", "aoc-1")
		}},
		{"RemoveAudioOutputConfiguration", "RemoveAudioOutputConfiguration", `<RemoveAudioOutputConfigurationResponse/>`, func(s *Service) error {
			return s.RemoveAudioOutputConfiguration(context.Background(), "profile-1")
		}},
		{"AddAudioDecoderConfiguration", "AddAudioDecoderConfiguration", `<AddAudioDecoderConfigurationResponse/>`, func(s *Service) error {
			return s.AddAudioDecoderConfiguration(context.Background(), "profile-1", "adc-1")
		}},
		{"RemoveAudioDecoderConfiguration", "RemoveAudioDecoderConfiguration", `<RemoveAudioDecoderConfigurationResponse/>`, func(s *Service) error {
			return s.RemoveAudioDecoderConfiguration(context.Background(), "profile-1")
		}},

		// --- OSD (media_osd.go) ---
		{"GetOSDs", "GetOSDs", `<GetOSDsResponse/>`, func(s *Service) error {
			_, err := s.GetOSDs(context.Background(), "")

			return err
		}},
		{"GetOSD", "GetOSD", `<GetOSDResponse/>`, func(s *Service) error {
			_, err := s.GetOSD(context.Background(), "osd-1")

			return err
		}},
		{"SetOSD", "SetOSD", `<SetOSDResponse/>`, func(s *Service) error {
			return s.SetOSD(context.Background(), &OSDConfiguration{Token: "osd-1"})
		}},
		{"CreateOSD", "CreateOSD", `<CreateOSDResponse/>`, func(s *Service) error {
			_, err := s.CreateOSD(context.Background(), "video-1", &OSDConfiguration{})

			return err
		}},
		{"DeleteOSD", "DeleteOSD", `<DeleteOSDResponse/>`, func(s *Service) error {
			return s.DeleteOSD(context.Background(), "osd-1")
		}},
		{"GetOSDOptions", "GetOSDOptions", `<GetOSDOptionsResponse/>`, func(s *Service) error {
			_, err := s.GetOSDOptions(context.Background(), "video-1")

			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, _ := newMediaOpsService(t, tc.action, tc.resp)

			if err := tc.invoke(s); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
		})
	}
}

func TestGetStreamURIMapping(t *testing.T) {
	s, _ := newMediaOpsService(t, "GetStreamUri",
		`<GetStreamUriResponse><MediaUri><Uri>rtsp://192.0.2.9:554/main</Uri><InvalidAfterConnect>false</InvalidAfterConnect><InvalidAfterReboot>true</InvalidAfterReboot></MediaUri></GetStreamUriResponse>`)

	uri, err := s.GetStreamURI(context.Background(), "profile-1")
	if err != nil {
		t.Fatal(err)
	}

	if uri.URI != "rtsp://192.0.2.9:554/main" || uri.InvalidAfterReboot != true {
		t.Fatalf("mapping wrong: %+v", uri)
	}
}
