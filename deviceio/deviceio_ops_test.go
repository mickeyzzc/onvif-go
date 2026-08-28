package deviceio

// Wire-contract tests for the device-io operations that had zero
// coverage, through internal/testutil.FakeCaller (tmd: request prefix).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
)

func newDeviceIOService(t *testing.T, wantAction, resp string) *Service {
	t.Helper()

	return New(testutil.NewFakeCaller("http://fake/device", func(action, _ string) (string, error) {
		trimmed := strings.TrimPrefix(strings.TrimPrefix(action, "tmd:"), "tds:")
		if trimmed != wantAction {
			return "", errors.New("unexpected action " + action)
		}

		return resp, nil
	}))
}

func TestDeviceIOOpsWireContract(t *testing.T) {
	cases := []struct {
		name   string
		action string
		resp   string
		invoke func(s *Service) error
	}{
		{"GetDigitalInputConfigurationOptions", "GetDigitalInputConfigurationOptions", `<GetDigitalInputConfigurationOptionsResponse/>`, func(s *Service) error {
			_, err := s.GetDigitalInputConfigurationOptions(context.Background(), "di-1")

			return err
		}},
		{"SetDigitalInputConfigurations", "SetDigitalInputConfigurations", `<SetDigitalInputConfigurationsResponse/>`, func(s *Service) error {
			return s.SetDigitalInputConfigurations(context.Background(), []*DigitalInput{{Token: "di-1"}})
		}},
		{"GetVideoOutputs", "GetVideoOutputs", `<GetVideoOutputsResponse/>`, func(s *Service) error {
			_, err := s.GetVideoOutputs(context.Background())

			return err
		}},
		{"GetSerialPorts", "GetSerialPorts", `<GetSerialPortsResponse/>`, func(s *Service) error {
			_, err := s.GetSerialPorts(context.Background())

			return err
		}},
		{"GetSerialPortConfiguration", "GetSerialPortConfiguration", `<GetSerialPortConfigurationResponse/>`, func(s *Service) error {
			_, err := s.GetSerialPortConfiguration(context.Background(), "sp-1")

			return err
		}},
		{"GetSerialPortConfigurationOptions", "GetSerialPortConfigurationOptions", `<GetSerialPortConfigurationOptionsResponse/>`, func(s *Service) error {
			_, err := s.GetSerialPortConfigurationOptions(context.Background(), "sp-1")

			return err
		}},
		{"SetSerialPortConfiguration", "SetSerialPortConfiguration", `<SetSerialPortConfigurationResponse/>`, func(s *Service) error {
			return s.SetSerialPortConfiguration(context.Background(), &SerialPortConfiguration{Token: "sp-1"})
		}},
		{"SendReceiveSerialCommand", "SendReceiveSerialCommand", `<SendReceiveSerialCommandResponse/>`, func(s *Service) error {
			_, err := s.SendReceiveSerialCommand(context.Background(), "sp-1", []byte("ping"), 2, 4)

			return err
		}},
		{"GetVideoOutputConfiguration", "GetVideoOutputConfiguration", `<GetVideoOutputConfigurationResponse/>`, func(s *Service) error {
			_, err := s.GetVideoOutputConfiguration(context.Background(), "vo-1")

			return err
		}},
		{"GetVideoOutputConfigurationOptions", "GetVideoOutputConfigurationOptions", `<GetVideoOutputConfigurationOptionsResponse/>`, func(s *Service) error {
			_, err := s.GetVideoOutputConfigurationOptions(context.Background(), "vo-1")

			return err
		}},
		{"SetVideoOutputConfiguration", "SetVideoOutputConfiguration", `<SetVideoOutputConfigurationResponse/>`, func(s *Service) error {
			return s.SetVideoOutputConfiguration(context.Background(), &VideoOutputConfiguration{Token: "vo-1"})
		}},
		{"GetRelayOutputOptions", "GetRelayOutputOptions", `<GetRelayOutputOptionsResponse/>`, func(s *Service) error {
			_, err := s.GetRelayOutputOptions(context.Background(), "ro-1")

			return err
		}},
		{"GetRelayOutputs", "GetRelayOutputs", `<GetRelayOutputsResponse/>`, func(s *Service) error {
			_, err := s.GetRelayOutputs(context.Background())

			return err
		}},
		{"SetRelayOutputSettings", "SetRelayOutputSettings", `<SetRelayOutputSettingsResponse/>`, func(s *Service) error {
			return s.SetRelayOutputSettings(context.Background(), "ro-1", &RelayOutputSettings{})
		}},
		{"SetRelayOutputState", "SetRelayOutputState", `<SetRelayOutputStateResponse/>`, func(s *Service) error {
			return s.SetRelayOutputState(context.Background(), "ro-1", RelayLogicalState("active"))
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.invoke(newDeviceIOService(t, tc.action, tc.resp)); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
		})
	}
}

func TestGetRelayOutputsMapping(t *testing.T) {
	s := newDeviceIOService(t, "GetRelayOutputs",
		`<GetRelayOutputsResponse><RelayOutputs token="ro-1"><Properties><Mode>Monostable</Mode><DelayTime>1</DelayTime></Properties></RelayOutputs></GetRelayOutputsResponse>`)

	relays, err := s.GetRelayOutputs(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(relays) != 1 || relays[0].Token != "ro-1" {
		t.Fatalf("mapping wrong: %+v", relays)
	}
}
