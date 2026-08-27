package onvif

import (
	"context"
	"encoding/xml"
	"fmt"
)

// NetworkInterfaceConfig is the desired configuration for one interface in
// SetNetworkInterfaces.
type NetworkInterfaceConfig struct {
	// Enabled toggles the interface itself.
	Enabled bool
	// IPv4Enabled toggles the interface's IPv4 stack.
	IPv4Enabled bool
	// DHCP enables DHCP address acquisition. When false, ManualAddress and
	// ManualPrefixLength define the static address.
	DHCP bool
	// ManualAddress is the static IPv4 address (ignored when DHCP is true).
	ManualAddress string
	// ManualPrefixLength is the static address prefix length, e.g. 24 for a
	// /24 (ignored when DHCP is true).
	ManualPrefixLength int
}

// SetNetworkInterfaces applies network interface configuration for one
// interface token (the ONVIF operation is per-interface despite its plural
// name). NOTE: most devices apply interface changes only after a reboot —
// check the returned rebootNeeded flag and call SystemReboot accordingly.
func (s *DeviceService) SetNetworkInterfaces(
	ctx context.Context,
	token string,
	cfg NetworkInterfaceConfig,
) (rebootNeeded bool, err error) {
	if token == "" {
		return false, fmt.Errorf("SetNetworkInterfaces: %w: interface token is empty", ErrInvalidParameter)
	}

	if !cfg.DHCP && cfg.ManualAddress == "" {
		return false, fmt.Errorf("SetNetworkInterfaces: %w: static configuration needs a manual address",
			ErrInvalidParameter)
	}

	type manualEntry struct {
		Address      string `xml:"tt:Address"`
		PrefixLength int    `xml:"tt:PrefixLength"`
	}

	type ipv4Config struct {
		Manual *manualEntry `xml:"tt:Manual,omitempty"`
		DHCP   bool         `xml:"tt:DHCP"`
	}

	type ipv4 struct {
		Enabled bool       `xml:"tt:Enabled"`
		Config  ipv4Config `xml:"tt:Config"`
	}

	type networkInterface struct {
		Enabled bool `xml:"tt:Enabled"`
		IPv4    ipv4 `xml:"tt:IPv4"`
	}

	type SetNetworkInterfaces struct {
		XMLName          xml.Name         `xml:"tds:SetNetworkInterfaces"`
		Xmlns            string           `xml:"xmlns:tds,attr"`
		Xmlnst           string           `xml:"xmlns:tt,attr"`
		InterfaceToken   string           `xml:"tds:InterfaceToken"`
		NetworkInterface networkInterface `xml:"tds:NetworkInterface"`
	}

	type SetNetworkInterfacesResponse struct {
		XMLName      xml.Name `xml:"SetNetworkInterfacesResponse"`
		RebootNeeded bool     `xml:"RebootNeeded"`
	}

	req := SetNetworkInterfaces{
		Xmlns:          deviceNamespace,
		Xmlnst:         "http://www.onvif.org/ver10/schema",
		InterfaceToken: token,
		NetworkInterface: networkInterface{
			Enabled: cfg.Enabled,
			IPv4: ipv4{
				Enabled: cfg.IPv4Enabled,
				Config:  ipv4Config{DHCP: cfg.DHCP},
			},
		},
	}

	if !cfg.DHCP {
		req.NetworkInterface.IPv4.Config.Manual = &manualEntry{
			Address:      cfg.ManualAddress,
			PrefixLength: cfg.ManualPrefixLength,
		}
	}

	var resp SetNetworkInterfacesResponse

	if err := s.client.call(ctx, s.client.endpoint, "", req, &resp); err != nil {
		return false, fmt.Errorf("SetNetworkInterfaces failed: %w", err)
	}

	return resp.RebootNeeded, nil
}

// NetmaskFromPrefixLength converts a prefix length (0-32) to a dotted netmask
// ("255.255.255.0" for 24). Returns "" for out-of-range prefixes. A
// convenience so every consumer does not reimplement the conversion.
func NetmaskFromPrefixLength(prefixLength int) string {
	if prefixLength < 0 || prefixLength > 32 {
		return ""
	}

	var mask uint32 = 0xFFFFFFFF
	mask <<= 32 - prefixLength

	if prefixLength == 0 {
		mask = 0
	}

	return fmt.Sprintf("%d.%d.%d.%d",
		byte(mask>>24), byte(mask>>16), byte(mask>>8), byte(mask))
}
