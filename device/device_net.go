package device

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
)

// GetZeroConfiguration gets the zero-configuration from a device.
func (s *Service) GetZeroConfiguration(ctx context.Context) (*NetworkZeroConfiguration, error) {
	type getZeroConfigurationRequest struct {
		XMLName xml.Name `xml:"tds:GetZeroConfiguration"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type getZeroConfigurationResponse struct {
		XMLName           xml.Name `xml:"GetZeroConfigurationResponse"`
		ZeroConfiguration struct {
			InterfaceToken string   `xml:"InterfaceToken"`
			Enabled        bool     `xml:"Enabled"`
			Addresses      []string `xml:"Addresses"`
		} `xml:"ZeroConfiguration"`
	}

	req := getZeroConfigurationRequest{
		Xmlns: Namespace,
	}

	var resp getZeroConfigurationResponse
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetZeroConfiguration failed: %w", err)
	}

	return &NetworkZeroConfiguration{
		InterfaceToken: resp.ZeroConfiguration.InterfaceToken,
		Enabled:        resp.ZeroConfiguration.Enabled,
		Addresses:      resp.ZeroConfiguration.Addresses,
	}, nil
}

// SetZeroConfiguration sets the zero-configuration.
// SetZeroConfiguration sets the zero-configuration.
func (s *Service) SetZeroConfiguration(ctx context.Context, interfaceToken string, enabled bool) error {
	type setZeroConfigurationRequest struct {
		XMLName        xml.Name `xml:"tds:SetZeroConfiguration"`
		Xmlns          string   `xml:"xmlns:tds,attr"`
		InterfaceToken string   `xml:"tds:InterfaceToken"`
		Enabled        bool     `xml:"tds:Enabled"`
	}

	req := setZeroConfigurationRequest{
		Xmlns:          Namespace,
		InterfaceToken: interfaceToken,
		Enabled:        enabled,
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetZeroConfiguration failed: %w", err)
	}

	return nil
}

// GetDynamicDNS gets the dynamic DNS settings from a device.
// GetDynamicDNS gets the dynamic DNS settings from a device.
func (s *Service) GetDynamicDNS(ctx context.Context) (*DynamicDNSInformation, error) {
	type getDynamicDNSRequest struct {
		XMLName xml.Name `xml:"tds:GetDynamicDNS"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type getDynamicDNSResponse struct {
		XMLName               xml.Name `xml:"GetDynamicDNSResponse"`
		DynamicDNSInformation struct {
			Type string `xml:"Type"`
			Name string `xml:"Name"`
			TTL  string `xml:"TTL"`
		} `xml:"DynamicDNSInformation"`
	}

	req := getDynamicDNSRequest{
		Xmlns: Namespace,
	}

	var resp getDynamicDNSResponse
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetDynamicDNS failed: %w", err)
	}

	return &DynamicDNSInformation{
		Type: DynamicDNSType(resp.DynamicDNSInformation.Type),
		Name: resp.DynamicDNSInformation.Name,
		// TTL would need duration parsing
	}, nil
}

// SetDynamicDNS sets the dynamic DNS settings on a device.
// SetDynamicDNS sets the dynamic DNS settings on a device.
func (s *Service) SetDynamicDNS(ctx context.Context, dnsType DynamicDNSType, name string) error {
	type setDynamicDNSRequest struct {
		XMLName xml.Name       `xml:"tds:SetDynamicDNS"`
		Xmlns   string         `xml:"xmlns:tds,attr"`
		Type    DynamicDNSType `xml:"tds:Type"`
		Name    string         `xml:"tds:Name,omitempty"`
	}

	req := setDynamicDNSRequest{
		Xmlns: Namespace,
		Type:  dnsType,
		Name:  name,
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetDynamicDNS failed: %w", err)
	}

	return nil
}

// GetPasswordComplexityConfiguration retrieves the current password complexity configuration settings.
