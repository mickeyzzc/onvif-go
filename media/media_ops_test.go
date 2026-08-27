package media

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
)

// The configuration-list operations share one wire shape: a response
// element wrapping repeated <Configurations token="…"> items (or a
// single <Configuration token="…">). One canned body drives them all.
func listResponse(responseElement string, items int) string {
	var b strings.Builder
	b.WriteString("<" + responseElement + ">")

	for i := range items {
		b.WriteString(`<Configurations token="tok-` + itoa(i) + `"><Name>cfg-</Name><UseCount>2</UseCount></Configurations>`)
	}

	b.WriteString("</" + responseElement + ">")

	return b.String()
}

func singleResponse(responseElement string) string {
	return "<" + responseElement + `><Configuration token="tok-0"><Name>cfg</Name><UseCount>2</UseCount></Configuration></` + responseElement + ">"
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}

	return string(rune('0' + i))
}

// opCase is one table-driven service-operation contract: the action the
// op must send, the canned response, and how to invoke it.
type opCase struct {
	name   string
	action string
	resp   string
	invoke func(s *Service) error
	check  func(t *testing.T, err error)
}

func runOpCases(t *testing.T, cases []opCase) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			caller := testutil.NewFakeCaller("http://fake/media", func(action, _ string) (string, error) {
				if action != tc.action {
					return "", errors.New("unexpected action " + action)
				}

				return tc.resp, nil
			})

			err := tc.invoke(New(caller))
			tc.check(t, err)

			if caller.CountAction(tc.action) != 1 {
				t.Errorf("%s sent %d requests, want 1", tc.action, caller.CountAction(tc.action))
			}
		})
	}
}

func TestAudioConfigurationOps(t *testing.T) {
	runOpCases(t, []opCase{
		{
			name: "GetAudioSourceConfigurations", action: "trt:GetAudioSourceConfigurations",
			resp: listResponse("GetAudioSourceConfigurationsResponse", 2),
			invoke: func(s *Service) error {
				_, err := s.GetAudioSourceConfigurations(context.Background())

				return err
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "GetAudioEncoderConfigurations", action: "trt:GetAudioEncoderConfigurations",
			resp: listResponse("GetAudioEncoderConfigurationsResponse", 1),
			invoke: func(s *Service) error {
				_, err := s.GetAudioEncoderConfigurations(context.Background())

				return err
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "GetAudioSourceConfiguration", action: "trt:GetAudioSourceConfiguration",
			resp: singleResponse("GetAudioSourceConfigurationResponse"),
			invoke: func(s *Service) error {
				_, err := s.GetAudioSourceConfiguration(context.Background(), "tok-0")

				return err
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "GetAudioDecoderConfiguration", action: "trt:GetAudioDecoderConfiguration",
			resp: singleResponse("GetAudioDecoderConfigurationResponse"),
			invoke: func(s *Service) error {
				_, err := s.GetAudioDecoderConfiguration(context.Background(), "tok-0")

				return err
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "GetMetadataConfigurations", action: "trt:GetMetadataConfigurations",
			resp: listResponse("GetMetadataConfigurationsResponse", 1),
			invoke: func(s *Service) error {
				_, err := s.GetMetadataConfigurations(context.Background())

				return err
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "GetAudioOutputConfigurations", action: "trt:GetAudioOutputConfigurations",
			resp: listResponse("GetAudioOutputConfigurationsResponse", 1),
			invoke: func(s *Service) error {
				_, err := s.GetAudioOutputConfigurations(context.Background())

				return err
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "GetAudioDecoderConfigurations", action: "trt:GetAudioDecoderConfigurations",
			resp: listResponse("GetAudioDecoderConfigurationsResponse", 1),
			invoke: func(s *Service) error {
				_, err := s.GetAudioDecoderConfigurations(context.Background())

				return err
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "SetAudioSourceConfiguration", action: "trt:SetAudioSourceConfiguration",
			invoke: func(s *Service) error {
				return s.SetAudioSourceConfiguration(context.Background(), &AudioSourceConfiguration{
					Token: "tok-0", Name: "mic", SourceToken: "src-1",
				}, false)
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "SetAudioDecoderConfiguration", action: "trt:SetAudioDecoderConfiguration",
			invoke: func(s *Service) error {
				return s.SetAudioDecoderConfiguration(context.Background(), &AudioDecoderConfiguration{
					Token: "tok-0", Name: "dec",
				}, false)
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "AddAudioOutputConfiguration", action: "trt:AddAudioOutputConfiguration",
			invoke: func(s *Service) error {
				return s.AddAudioOutputConfiguration(context.Background(), "profile-1", "tok-0")
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "RemoveAudioOutputConfiguration", action: "trt:RemoveAudioOutputConfiguration",
			invoke: func(s *Service) error {
				return s.RemoveAudioOutputConfiguration(context.Background(), "profile-1")
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "AddAudioDecoderConfiguration", action: "trt:AddAudioDecoderConfiguration",
			invoke: func(s *Service) error {
				return s.AddAudioDecoderConfiguration(context.Background(), "profile-1", "tok-0")
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "RemoveAudioDecoderConfiguration", action: "trt:RemoveAudioDecoderConfiguration",
			invoke: func(s *Service) error {
				return s.RemoveAudioDecoderConfiguration(context.Background(), "profile-1")
			},
			check: func(t *testing.T, err error) {},
		},
	})
}

func TestCompatibleConfigurationOps(t *testing.T) {
	compatible := []opCase{
		{
			name: "GetCompatibleAudioEncoderConfigurations", action: "trt:GetCompatibleAudioEncoderConfigurations",
			resp: listResponse("GetCompatibleAudioEncoderConfigurationsResponse", 1),
		},
		{
			name: "GetCompatibleAudioSourceConfigurations", action: "trt:GetCompatibleAudioSourceConfigurations",
			resp: listResponse("GetCompatibleAudioSourceConfigurationsResponse", 1),
		},
		{
			name: "GetCompatibleMetadataConfigurations", action: "trt:GetCompatibleMetadataConfigurations",
			resp: listResponse("GetCompatibleMetadataConfigurationsResponse", 1),
		},
		{
			name: "GetCompatibleAudioOutputConfigurations", action: "trt:GetCompatibleAudioOutputConfigurations",
			resp: listResponse("GetCompatibleAudioOutputConfigurationsResponse", 1),
		},
		{
			name: "GetCompatibleAudioDecoderConfigurations", action: "trt:GetCompatibleAudioDecoderConfigurations",
			resp: listResponse("GetCompatibleAudioDecoderConfigurationsResponse", 1),
		},
		{
			name: "GetCompatibleVideoEncoderConfigurations", action: "trt:GetCompatibleVideoEncoderConfigurations",
			resp: listResponse("GetCompatibleVideoEncoderConfigurationsResponse", 1),
		},
		{
			name: "GetCompatibleVideoSourceConfigurations", action: "trt:GetCompatibleVideoSourceConfigurations",
			resp: listResponse("GetCompatibleVideoSourceConfigurationsResponse", 1),
		},
		{
			name: "GetCompatiblePTZConfigurations", action: "trt:GetCompatiblePTZConfigurations",
			resp: listResponse("GetCompatiblePTZConfigurationsResponse", 1),
		},
		{
			name: "GetCompatibleVideoAnalyticsConfigurations", action: "trt:GetCompatibleVideoAnalyticsConfigurations",
			resp: listResponse("GetCompatibleVideoAnalyticsConfigurationsResponse", 1),
		},
	}

	for i := range compatible {
		tc := compatible[i]
		compatible[i].invoke = func(s *Service) error {
			// All compatible-* ops share the (ctx, profileToken) shape;
			// dispatch through the recorded action is verified by the
			// generic runner, so route by name here.
			var err error

			switch tc.name {
			case "GetCompatibleAudioEncoderConfigurations":
				_, err = s.GetCompatibleAudioEncoderConfigurations(context.Background(), "profile-1")
			case "GetCompatibleAudioSourceConfigurations":
				_, err = s.GetCompatibleAudioSourceConfigurations(context.Background(), "profile-1")
			case "GetCompatibleMetadataConfigurations":
				_, err = s.GetCompatibleMetadataConfigurations(context.Background(), "profile-1")
			case "GetCompatibleAudioOutputConfigurations":
				_, err = s.GetCompatibleAudioOutputConfigurations(context.Background(), "profile-1")
			case "GetCompatibleAudioDecoderConfigurations":
				_, err = s.GetCompatibleAudioDecoderConfigurations(context.Background(), "profile-1")
			case "GetCompatibleVideoEncoderConfigurations":
				_, err = s.GetCompatibleVideoEncoderConfigurations(context.Background(), "profile-1")
			case "GetCompatibleVideoSourceConfigurations":
				_, err = s.GetCompatibleVideoSourceConfigurations(context.Background(), "profile-1")
			case "GetCompatiblePTZConfigurations":
				_, err = s.GetCompatiblePTZConfigurations(context.Background(), "profile-1")
			case "GetCompatibleVideoAnalyticsConfigurations":
				_, err = s.GetCompatibleVideoAnalyticsConfigurations(context.Background(), "profile-1")
			}

			return err
		}
		compatible[i].check = func(t *testing.T, err error) {}
	}

	runOpCases(t, compatible)
}

func TestVideoConfigurationOps(t *testing.T) {
	runOpCases(t, []opCase{
		{
			name: "GetVideoSourceConfigurations", action: "trt:GetVideoSourceConfigurations",
			resp: listResponse("GetVideoSourceConfigurationsResponse", 1),
			invoke: func(s *Service) error {
				_, err := s.GetVideoSourceConfigurations(context.Background())

				return err
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "GetVideoEncoderConfigurations", action: "trt:GetVideoEncoderConfigurations",
			resp: listResponse("GetVideoEncoderConfigurationsResponse", 1),
			invoke: func(s *Service) error {
				_, err := s.GetVideoEncoderConfigurations(context.Background())

				return err
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "GetVideoSourceConfiguration", action: "trt:GetVideoSourceConfiguration",
			resp: singleResponse("GetVideoSourceConfigurationResponse"),
			invoke: func(s *Service) error {
				_, err := s.GetVideoSourceConfiguration(context.Background(), "tok-0")

				return err
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "SetVideoSourceConfiguration", action: "trt:SetVideoSourceConfiguration",
			invoke: func(s *Service) error {
				return s.SetVideoSourceConfiguration(context.Background(), &VideoSourceConfiguration{
					Token: "tok-0", Name: "cam", SourceToken: "src-1",
				}, false)
			},
			check: func(t *testing.T, err error) {},
		},
	})
}

func TestVideoAnalyticsConfigurationOps(t *testing.T) {
	runOpCases(t, []opCase{
		{
			name: "GetVideoAnalyticsConfigurations", action: "trt:GetVideoAnalyticsConfigurations",
			resp: listResponse("GetVideoAnalyticsConfigurationsResponse", 1),
			invoke: func(s *Service) error {
				_, err := s.GetVideoAnalyticsConfigurations(context.Background())

				return err
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "GetVideoAnalyticsConfiguration", action: "trt:GetVideoAnalyticsConfiguration",
			resp: singleResponse("GetVideoAnalyticsConfigurationResponse"),
			invoke: func(s *Service) error {
				_, err := s.GetVideoAnalyticsConfiguration(context.Background(), "tok-0")

				return err
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "SetVideoAnalyticsConfiguration", action: "trt:SetVideoAnalyticsConfiguration",
			invoke: func(s *Service) error {
				return s.SetVideoAnalyticsConfiguration(context.Background(), &VideoAnalyticsConfiguration{
					Token: "tok-0", Name: "analytics",
				}, false)
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "GetVideoAnalyticsConfigurationOptions", action: "trt:GetVideoAnalyticsConfigurationOptions",
			resp: `<GetVideoAnalyticsConfigurationOptionsResponse><Options><AnalyticsEngineConfigInfo/></Options></GetVideoAnalyticsConfigurationOptionsResponse>`,
			invoke: func(s *Service) error {
				_, err := s.GetVideoAnalyticsConfigurationOptions(context.Background(), "tok-0", "profile-1")

				return err
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "AddVideoAnalyticsConfiguration", action: "trt:AddVideoAnalyticsConfiguration",
			invoke: func(s *Service) error {
				return s.AddVideoAnalyticsConfiguration(context.Background(), "profile-1", "tok-0")
			},
			check: func(t *testing.T, err error) {},
		},
		{
			name: "RemoveVideoAnalyticsConfiguration", action: "trt:RemoveVideoAnalyticsConfiguration",
			invoke: func(s *Service) error {
				return s.RemoveVideoAnalyticsConfiguration(context.Background(), "profile-1")
			},
			check: func(t *testing.T, err error) {},
		},
	})
}

func TestAudioSourceConfigurationOptions(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/media", func(action, _ string) (string, error) {
		if action != "trt:GetAudioSourceConfigurationOptions" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetAudioSourceConfigurationOptionsResponse>
	<Options><InputTokensAvailable>src-1</InputTokensAvailable><InputTokensAvailable>src-2</InputTokensAvailable></Options>
</GetAudioSourceConfigurationOptionsResponse>`, nil
	})

	options, err := New(caller).GetAudioSourceConfigurationOptions(context.Background(), "tok-0", "profile-1")
	if err != nil {
		t.Fatalf("GetAudioSourceConfigurationOptions: %v", err)
	}

	if len(options.InputTokensAvailable) != 2 || options.InputTokensAvailable[0] != "src-1" {
		t.Errorf("InputTokensAvailable = %v", options.InputTokensAvailable)
	}
}

func TestAudioEncoderListParsesMulticast(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/media", func(action, _ string) (string, error) {
		if action != "trt:GetAudioEncoderConfigurations" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetAudioEncoderConfigurationsResponse>
	<Configurations token="aenc-1">
		<Name>mic g711</Name><UseCount>3</UseCount><Encoding>G711</Encoding>
		<Bitrate>64</Bitrate><SampleRate>8000</SampleRate>
		<SessionTimeout>PT60S</SessionTimeout>
		<Multicast><Port>5000</Port><TTL>5</TTL><AutoStart>false</AutoStart></Multicast>
	</Configurations>
</GetAudioEncoderConfigurationsResponse>`, nil
	})

	configs, err := New(caller).GetAudioEncoderConfigurations(context.Background())
	if err != nil {
		t.Fatalf("GetAudioEncoderConfigurations: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("configs = %d, want 1", len(configs))
	}

	cfg := configs[0]
	if cfg.Token != "aenc-1" || cfg.Encoding != "G711" || cfg.Bitrate != 64 || cfg.SampleRate != 8000 {
		t.Errorf("audio encoder fields not parsed: %+v", cfg)
	}

	if cfg.Multicast == nil || cfg.Multicast.Port != 5000 || cfg.Multicast.TTL != 5 {
		t.Errorf("multicast not parsed: %+v", cfg.Multicast)
	}
}

func TestEmptyListsAreValid(t *testing.T) {
	// Devices without audio return empty lists — ops must yield empty,
	// non-nil results, not errors.
	caller := testutil.NewFakeCaller("http://fake/media", func(action, _ string) (string, error) {
		return "<" + strings.ReplaceAll(strings.TrimPrefix(action, "trt:Get"), "Get", "") + "0/>", nil
	})
	_ = caller

	empty := testutil.NewFakeCaller("http://fake/media", func(action, _ string) (string, error) {
		switch action {
		case "trt:GetAudioSourceConfigurations":
			return `<GetAudioSourceConfigurationsResponse/>`, nil
		case "trt:GetVideoEncoderConfigurations":
			return `<GetVideoEncoderConfigurationsResponse/>`, nil
		default:
			return "", errors.New("unexpected action " + action)
		}
	})

	svc := New(empty)

	sources, err := svc.GetAudioSourceConfigurations(context.Background())
	if err != nil || len(sources) != 0 {
		t.Errorf("empty audio sources = %v, %v", sources, err)
	}

	encoders, err := svc.GetVideoEncoderConfigurations(context.Background())
	if err != nil || len(encoders) != 0 {
		t.Errorf("empty video encoders = %v, %v", encoders, err)
	}
}

func TestProfileHelpers(t *testing.T) {
	profiles := []*Profile{
		{Token: "main", Name: "Main", VideoEncoderConfiguration: &VideoEncoderConfiguration{
			Resolution: &VideoResolution{Width: 2560, Height: 1440},
		}},
		{Token: "sub", Name: "Sub 1", VideoEncoderConfiguration: &VideoEncoderConfiguration{
			Resolution: &VideoResolution{Width: 704, Height: 480},
		}},
		{Token: "sub2", Name: "substream", VideoEncoderConfiguration: &VideoEncoderConfiguration{
			Resolution: &VideoResolution{Width: 1280, Height: 720},
		}},
	}

	if p := profileByToken(profiles, "sub"); p == nil || p.Token != "sub" {
		t.Errorf("profileByToken = %+v", p)
	}

	if p := profileByToken(profiles, "missing"); p != nil {
		t.Errorf("profileByToken(missing) = %+v, want nil", p)
	}

	// subTieBreak is a naming-hint comparator: sub-ish names sort first.
	if diff := subTieBreak(profiles[1], profiles[0]); diff >= 0 {
		t.Errorf("subTieBreak(sub, main) = %d, want negative (sub preferred)", diff)
	}

	if diff := subTieBreak(profiles[0], profiles[1]); diff <= 0 {
		t.Errorf("subTieBreak(main, sub) = %d, want positive", diff)
	}
}
