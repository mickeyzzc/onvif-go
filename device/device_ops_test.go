package device

// Wire-contract tests for the device-service operations that had zero
// coverage — table-driven through internal/testutil.FakeCaller. Most ops
// assert the action name and error-free mapping from a minimal response;
// a few carry content where the mapping is worth pinning.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
	"github.com/mickeyzzc/onvif-go/v2/ptz"
	"github.com/mickeyzzc/onvif-go/v2/types"
)

func newDeviceOpsService(t *testing.T, wantAction, resp string) (*Service, *testutil.FakeCaller) {
	t.Helper()

	caller := testutil.NewFakeCaller("http://fake/device_service", func(action, _ string) (string, error) {
		if strings.TrimPrefix(strings.TrimPrefix(action, "tds:"), "tds:") != wantAction {
			return "", errors.New("unexpected action " + action)
		}

		return resp, nil
	})

	return New(caller), caller
}

// TestDeviceOpsWireContract drives every zero-coverage operation with a
// minimal valid response; the assertions pin the action name and the
// no-error decode path.
func TestDeviceOpsWireContract(t *testing.T) {
	cases := []struct {
		name   string
		action string
		resp   string
		invoke func(s *Service) error
	}{
		{"GetZeroConfiguration", "GetZeroConfiguration", `<GetZeroConfigurationResponse/>`, func(s *Service) error {
			_, err := s.GetZeroConfiguration(context.Background())

			return err
		}},
		{"SetZeroConfiguration", "SetZeroConfiguration", `<SetZeroConfigurationResponse/>`, func(s *Service) error {
			return s.SetZeroConfiguration(context.Background(), "eth0", true)
		}},
		{"GetDynamicDNS", "GetDynamicDNS", `<GetDynamicDNSResponse/>`, func(s *Service) error {
			_, err := s.GetDynamicDNS(context.Background())

			return err
		}},
		{"SetDynamicDNS", "SetDynamicDNS", `<SetDynamicDNSResponse/>`, func(s *Service) error {
			return s.SetDynamicDNS(context.Background(), DynamicDNSType("Client"), "dyn.example.org")
		}},
		{"GetHostname", "GetHostname", `<GetHostnameResponse/>`, func(s *Service) error {
			_, err := s.GetHostname(context.Background())

			return err
		}},
		{"SetHostname", "SetHostname", `<SetHostnameResponse/>`, func(s *Service) error {
			return s.SetHostname(context.Background(), "cam1")
		}},
		{"GetDNS", "GetDNS", `<GetDNSResponse/>`, func(s *Service) error {
			_, err := s.GetDNS(context.Background())

			return err
		}},
		{"SetDNS", "SetDNS", `<SetDNSResponse/>`, func(s *Service) error {
			return s.SetDNS(context.Background(), false, nil, nil)
		}},
		{"GetNTP", "GetNTP", `<GetNTPResponse/>`, func(s *Service) error {
			_, err := s.GetNTP(context.Background())

			return err
		}},
		{"SetNTP", "SetNTP", `<SetNTPResponse/>`, func(s *Service) error {
			return s.SetNTP(context.Background(), true, nil)
		}},
		{"SetHostnameFromDHCP", "SetHostnameFromDHCP", `<SetHostnameFromDHCPResponse/>`, func(s *Service) error {
			_, err := s.SetHostnameFromDHCP(context.Background(), true)

			return err
		}},
		{"FixedGetSystemDateAndTime", "GetSystemDateAndTime", `<GetSystemDateAndTimeResponse/>`, func(s *Service) error {
			_, err := s.FixedGetSystemDateAndTime(context.Background())

			return err
		}},
		{"SetSystemDateAndTime", "SetSystemDateAndTime", `<SetSystemDateAndTimeResponse/>`, func(s *Service) error {
			return s.SetSystemDateAndTime(context.Background(), &SystemDateTime{})
		}},
		{"GetSystemLog", "GetSystemLog", `<GetSystemLogResponse/>`, func(s *Service) error {
			_, err := s.GetSystemLog(context.Background(), SystemLogType("System"))

			return err
		}},
		{"GetSystemBackup", "GetSystemBackup", `<GetSystemBackupResponse/>`, func(s *Service) error {
			_, err := s.GetSystemBackup(context.Background())

			return err
		}},
		{"RestoreSystem", "RestoreSystem", `<RestoreSystemResponse/>`, func(s *Service) error {
			return s.RestoreSystem(context.Background(), nil)
		}},
		{"GetSystemUris", "GetSystemUris", `<GetSystemUrisResponse/>`, func(s *Service) error {
			_, _, _, err := s.GetSystemUris(context.Background())

			return err
		}},
		{"GetSystemSupportInformation", "GetSystemSupportInformation", `<GetSystemSupportInformationResponse/>`, func(s *Service) error {
			_, err := s.GetSystemSupportInformation(context.Background())

			return err
		}},
		{"SetSystemFactoryDefault", "SetSystemFactoryDefault", `<SetSystemFactoryDefaultResponse/>`, func(s *Service) error {
			return s.SetSystemFactoryDefault(context.Background(), FactoryDefaultType("Hard"))
		}},
		{"StartFirmwareUpgrade", "StartFirmwareUpgrade", `<StartFirmwareUpgradeResponse/>`, func(s *Service) error {
			_, _, _, err := s.StartFirmwareUpgrade(context.Background())

			return err
		}},
		{"StartSystemRestore", "StartSystemRestore", `<StartSystemRestoreResponse/>`, func(s *Service) error {
			_, _, err := s.StartSystemRestore(context.Background())

			return err
		}},
		{"SetNetworkInterfaces", "SetNetworkInterfaces", `<SetNetworkInterfacesResponse/>`, func(s *Service) error {
			_, err := s.SetNetworkInterfaces(context.Background(), "eth0", NetworkInterfaceConfig{DHCP: true})

			return err
		}},
		{"GetSystemDateAndTime", "GetSystemDateAndTime", `<GetSystemDateAndTimeResponse/>`, func(s *Service) error {
			_, err := s.GetSystemDateAndTime(context.Background())

			return err
		}},
		{"GetNetworkInterfaces", "GetNetworkInterfaces", `<GetNetworkInterfacesResponse/>`, func(s *Service) error {
			_, err := s.GetNetworkInterfaces(context.Background())

			return err
		}},
		{"GetScopes", "GetScopes", `<GetScopesResponse/>`, func(s *Service) error {
			_, err := s.GetScopes(context.Background())

			return err
		}},
		{"GetDeviceServiceCapabilities", "GetServiceCapabilities", `<GetServiceCapabilitiesResponse/>`, func(s *Service) error {
			_, err := s.GetDeviceServiceCapabilities(context.Background())

			return err
		}},
		{"GetDiscoveryMode", "GetDiscoveryMode", `<GetDiscoveryModeResponse/>`, func(s *Service) error {
			_, err := s.GetDiscoveryMode(context.Background())

			return err
		}},
		{"SetDiscoveryMode", "SetDiscoveryMode", `<SetDiscoveryModeResponse/>`, func(s *Service) error {
			return s.SetDiscoveryMode(context.Background(), DiscoveryModeDiscoverable)
		}},
		{"GetRemoteDiscoveryMode", "GetRemoteDiscoveryMode", `<GetRemoteDiscoveryModeResponse/>`, func(s *Service) error {
			_, err := s.GetRemoteDiscoveryMode(context.Background())

			return err
		}},
		{"SetRemoteDiscoveryMode", "SetRemoteDiscoveryMode", `<SetRemoteDiscoveryModeResponse/>`, func(s *Service) error {
			return s.SetRemoteDiscoveryMode(context.Background(), DiscoveryModeDiscoverable)
		}},
		{"GetEndpointReference", "GetEndpointReference", `<GetEndpointReferenceResponse/>`, func(s *Service) error {
			_, err := s.GetEndpointReference(context.Background())

			return err
		}},
		{"GetNetworkProtocols", "GetNetworkProtocols", `<GetNetworkProtocolsResponse/>`, func(s *Service) error {
			_, err := s.GetNetworkProtocols(context.Background())

			return err
		}},
		{"SetNetworkProtocols", "SetNetworkProtocols", `<SetNetworkProtocolsResponse/>`, func(s *Service) error {
			return s.SetNetworkProtocols(context.Background(), nil)
		}},
		{"GetNetworkDefaultGateway", "GetNetworkDefaultGateway", `<GetNetworkDefaultGatewayResponse/>`, func(s *Service) error {
			_, err := s.GetNetworkDefaultGateway(context.Background())

			return err
		}},
		{"SetNetworkDefaultGateway", "SetNetworkDefaultGateway", `<SetNetworkDefaultGatewayResponse/>`, func(s *Service) error {
			return s.SetNetworkDefaultGateway(context.Background(), &NetworkGateway{})
		}},
		{"GetGeoLocation", "GetGeoLocation", `<GetGeoLocationResponse/>`, func(s *Service) error {
			_, err := s.GetGeoLocation(context.Background())

			return err
		}},
		{"SetGeoLocation", "SetGeoLocation", `<SetGeoLocationResponse/>`, func(s *Service) error {
			return s.SetGeoLocation(context.Background(), nil)
		}},
		{"DeleteGeoLocation", "DeleteGeoLocation", `<DeleteGeoLocationResponse/>`, func(s *Service) error {
			return s.DeleteGeoLocation(context.Background(), nil)
		}},
		{"GetDPAddresses", "GetDPAddresses", `<GetDPAddressesResponse/>`, func(s *Service) error {
			_, err := s.GetDPAddresses(context.Background())

			return err
		}},
		{"SetDPAddresses", "SetDPAddresses", `<SetDPAddressesResponse/>`, func(s *Service) error {
			return s.SetDPAddresses(context.Background(), nil)
		}},
		{"GetWsdlURL", "GetWsdlUrl", `<GetWsdlUrlResponse/>`, func(s *Service) error {
			_, err := s.GetWsdlURL(context.Background())

			return err
		}},
		{"AddScopes", "AddScopes", `<AddScopesResponse/>`, func(s *Service) error {
			return s.AddScopes(context.Background(), []string{"onvif://www.onvif.org/type/video_encoder"})
		}},
		{"RemoveScopes", "RemoveScopes", `<RemoveScopesResponse/>`, func(s *Service) error {
			_, err := s.RemoveScopes(context.Background(), nil)

			return err
		}},
		{"SetScopes", "SetScopes", `<SetScopesResponse/>`, func(s *Service) error {
			return s.SetScopes(context.Background(), nil)
		}},
		{"SendAuxiliaryCommand", "SendAuxiliaryCommand", `<SendAuxiliaryCommandResponse/>`, func(s *Service) error {
			_, err := s.SendAuxiliaryCommand(context.Background(), ptz.AuxiliaryData("wiper"))

			return err
		}},
		{"GetStorageConfigurations", "GetStorageConfigurations", `<GetStorageConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetStorageConfigurations(context.Background())

			return err
		}},
		{"GetStorageConfiguration", "GetStorageConfiguration", `<GetStorageConfigurationResponse/>`, func(s *Service) error {
			_, err := s.GetStorageConfiguration(context.Background(), "tok")

			return err
		}},
		{"CreateStorageConfiguration", "CreateStorageConfiguration", `<CreateStorageConfigurationResponse/>`, func(s *Service) error {
			_, err := s.CreateStorageConfiguration(context.Background(), &StorageConfiguration{})

			return err
		}},
		{"SetStorageConfiguration", "SetStorageConfiguration", `<SetStorageConfigurationResponse/>`, func(s *Service) error {
			return s.SetStorageConfiguration(context.Background(), &StorageConfiguration{})
		}},
		{"DeleteStorageConfiguration", "DeleteStorageConfiguration", `<DeleteStorageConfigurationResponse/>`, func(s *Service) error {
			return s.DeleteStorageConfiguration(context.Background(), "tok")
		}},
		{"SetHashingAlgorithm", "SetHashingAlgorithm", `<SetHashingAlgorithmResponse/>`, func(s *Service) error {
			return s.SetHashingAlgorithm(context.Background(), "SHA-256")
		}},
		{"GetDot11Capabilities", "GetDot11Capabilities", `<GetDot11CapabilitiesResponse/>`, func(s *Service) error {
			_, err := s.GetDot11Capabilities(context.Background())

			return err
		}},
		{"GetDot11Status", "GetDot11Status", `<GetDot11StatusResponse/>`, func(s *Service) error {
			_, err := s.GetDot11Status(context.Background(), "wifi0")

			return err
		}},
		{"GetDot1XConfiguration", "GetDot1XConfiguration", `<GetDot1XConfigurationResponse/>`, func(s *Service) error {
			_, err := s.GetDot1XConfiguration(context.Background(), "dot1x-1")

			return err
		}},
		{"GetDot1XConfigurations", "GetDot1XConfigurations", `<GetDot1XConfigurationsResponse/>`, func(s *Service) error {
			_, err := s.GetDot1XConfigurations(context.Background())

			return err
		}},
		{"SetDot1XConfiguration", "SetDot1XConfiguration", `<SetDot1XConfigurationResponse/>`, func(s *Service) error {
			return s.SetDot1XConfiguration(context.Background(), &Dot1XConfiguration{})
		}},
		{"CreateDot1XConfiguration", "CreateDot1XConfiguration", `<CreateDot1XConfigurationResponse/>`, func(s *Service) error {
			return s.CreateDot1XConfiguration(context.Background(), &Dot1XConfiguration{})
		}},
		{"DeleteDot1XConfiguration", "DeleteDot1XConfiguration", `<DeleteDot1XConfigurationResponse/>`, func(s *Service) error {
			return s.DeleteDot1XConfiguration(context.Background(), "dot1x-1")
		}},
		{"ScanAvailableDot11Networks", "ScanAvailableDot11Networks", `<ScanAvailableDot11NetworksResponse/>`, func(s *Service) error {
			_, err := s.ScanAvailableDot11Networks(context.Background(), "wifi0")

			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, _ := newDeviceOpsService(t, tc.action, tc.resp)

			if err := tc.invoke(s); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
		})
	}
}

func TestGetDeviceInformationMapping(t *testing.T) {
	s, _ := newDeviceOpsService(t, "GetDeviceInformation",
		`<GetDeviceInformationResponse><Manufacturer>Acme</Manufacturer><Model>CamX</Model><FirmwareVersion>1.2.3</FirmwareVersion><SerialNumber>SN-1</SerialNumber><HardwareId>HW-1</HardwareId></GetDeviceInformationResponse>`)

	info, err := s.GetDeviceInformation(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if info.Manufacturer != "Acme" || info.SerialNumber != "SN-1" {
		t.Fatalf("mapping wrong: %+v", info)
	}
}

func TestSystemRebootWire(t *testing.T) {
	s, _ := newDeviceOpsService(t, "SystemReboot",
		`<SystemRebootResponse><Message>rebooting</Message></SystemRebootResponse>`)

	msg, err := s.SystemReboot(context.Background())
	if err != nil || msg != "rebooting" {
		t.Fatalf("msg = %q err %v", msg, err)
	}
}

// TestCapabilitiesCache: the cached accessor fetches once and serves
// the cache afterwards; InvalidateCapsCache forces a refetch.
func TestCapabilitiesCache(t *testing.T) {
	calls := 0

	caller := testutil.NewFakeCaller("http://fake/device_service", func(action, _ string) (string, error) {
		if strings.TrimPrefix(action, "tds:") != "GetCapabilities" && action != "GetCapabilities" {
			return "", errors.New("unexpected action " + action)
		}

		calls++

		return `<GetCapabilitiesResponse/>`, nil
	})

	s := New(caller)

	if _, err := s.GetCapabilitiesCached(context.Background()); err != nil {
		t.Fatal(err)
	}

	if _, err := s.GetCapabilitiesCached(context.Background()); err != nil {
		t.Fatal(err)
	}

	if calls != 1 {
		t.Fatalf("wire calls = %d, want 1 (second read must hit the cache)", calls)
	}

	s.InvalidateCapsCache()

	if _, err := s.GetCapabilitiesCached(context.Background()); err != nil {
		t.Fatal(err)
	}

	if calls != 2 {
		t.Fatalf("wire calls after invalidate = %d, want 2", calls)
	}
}

// TestSetNetworkInterfacesValidation pins the parameter validation
// branches without any wire traffic.
func TestSetNetworkInterfacesValidation(t *testing.T) {
	s := New(testutil.NewFakeCaller("", func(_, _ string) (string, error) { return "", nil }))

	if _, err := s.SetNetworkInterfaces(context.Background(), "", NetworkInterfaceConfig{DHCP: true}); err == nil {
		t.Error("empty token accepted")
	}

	if _, err := s.SetNetworkInterfaces(context.Background(), "eth0", NetworkInterfaceConfig{DHCP: false}); err == nil {
		t.Error("static config without manual address accepted")
	} else if !errors.Is(err, types.ErrInvalidParameter) {
		t.Error("validation must return types.ErrInvalidParameter")
	}
}

func TestGetServicesMapping(t *testing.T) {
	s, _ := newDeviceOpsService(t, "GetServices",
		`<GetServicesResponse><Service><Namespace>http://www.onvif.org/ver10/device/wsdl</Namespace><XAddr>http://192.0.2.9/onvif/device_service</XAddr></Service></GetServicesResponse>`)

	services, err := s.GetServices(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}

	if len(services) != 1 || services[0].Namespace != "http://www.onvif.org/ver10/device/wsdl" {
		t.Fatalf("mapping wrong: %+v", services)
	}
}

func TestNewWithFallback(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake", func(_, _ string) (string, error) { return "", nil })

	s := NewWithFallback(caller, true)
	if s == nil || !s.minimalCapsFallback {
		t.Fatal("NewWithFallback(caller, true) must set the fallback flag")
	}

	if NewWithFallback(caller, false).minimalCapsFallback {
		t.Fatal("NewWithFallback(caller, false) must clear the flag")
	}
}

func TestNetmaskFromPrefixLength(t *testing.T) {
	cases := map[int]string{
		0:  "0.0.0.0",
		8:  "255.0.0.0",
		24: "255.255.255.0",
		32: "255.255.255.255",
		-1: "",
		33: "",
	}

	for prefix, want := range cases {
		if got := NetmaskFromPrefixLength(prefix); got != want {
			t.Errorf("NetmaskFromPrefixLength(%d) = %q, want %q", prefix, got, want)
		}
	}
}
