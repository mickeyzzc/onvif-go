// Device security: users, remote user, access policy, certificates.

package security

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/mickeyzzc/onvif-go/v2/device"
	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/types"
)

// Common XML request/response types for device security operations.
// These are defined at package level to avoid repeated inline struct definitions.

// ipAddressFilterRequest is the common structure for IP address filter SOAP requests.
type ipAddressFilterRequest struct {
	Type        string                   `xml:"tds:Type"`
	IPv4Address []prefixedIPv4AddressXML `xml:"tds:IPv4Address,omitempty"`
	IPv6Address []prefixedIPv6AddressXML `xml:"tds:IPv6Address,omitempty"`
}

// prefixedIPv4AddressXML is the XML representation of a prefixed IPv4 address.
type prefixedIPv4AddressXML struct {
	Address      string `xml:"tds:Address"`
	PrefixLength int    `xml:"tds:PrefixLength"`
}

// prefixedIPv6AddressXML is the XML representation of a prefixed IPv6 address.
type prefixedIPv6AddressXML struct {
	Address      string `xml:"tds:Address"`
	PrefixLength int    `xml:"tds:PrefixLength"`
}

// buildIPAddressFilterRequest converts an IPAddressFilter to the XML request format.
// Pre-allocates slices for efficiency when the source length is known.
func buildIPAddressFilterRequest(filter *IPAddressFilter) ipAddressFilterRequest {
	req := ipAddressFilterRequest{
		Type: string(filter.Type),
	}

	// Pre-allocate slices with known capacity
	if len(filter.IPv4Address) > 0 {
		req.IPv4Address = make([]prefixedIPv4AddressXML, 0, len(filter.IPv4Address))
		for _, addr := range filter.IPv4Address {
			req.IPv4Address = append(req.IPv4Address, prefixedIPv4AddressXML{
				Address: addr.Address, PrefixLength: addr.PrefixLength,
			})
		}
	}

	if len(filter.IPv6Address) > 0 {
		req.IPv6Address = make([]prefixedIPv6AddressXML, 0, len(filter.IPv6Address))
		for _, addr := range filter.IPv6Address {
			req.IPv6Address = append(req.IPv6Address, prefixedIPv6AddressXML(addr))
		}
	}

	return req
}

// GetRemoteUser returns the configured remote user.
func (s *Service) GetRemoteUser(ctx context.Context) (*RemoteUser, error) {
	type getRemoteUserRequest struct {
		XMLName xml.Name `xml:"tds:GetRemoteUser"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type getRemoteUserResponse struct {
		XMLName    xml.Name `xml:"GetRemoteUserResponse"`
		RemoteUser *struct {
			Username           string `xml:"Username"`
			Password           string `xml:"Password"`
			UseDerivedPassword bool   `xml:"UseDerivedPassword"`
		} `xml:"RemoteUser"`
	}

	req := getRemoteUserRequest{
		Xmlns: device.Namespace,
	}

	var resp getRemoteUserResponse
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetRemoteUser failed: %w", err)
	}

	if resp.RemoteUser == nil {
		return nil, nil //nolint:nilnil // nil result = device has no remote user configured
	}

	return &RemoteUser{
		Username:           resp.RemoteUser.Username,
		Password:           resp.RemoteUser.Password,
		UseDerivedPassword: resp.RemoteUser.UseDerivedPassword,
	}, nil
}

// SetRemoteUser sets the remote user.
func (s *Service) SetRemoteUser(ctx context.Context, remoteUser *RemoteUser) error {
	type remoteUserXML struct {
		Username           string `xml:"tds:Username"`
		Password           string `xml:"tds:Password,omitempty"`
		UseDerivedPassword bool   `xml:"tds:UseDerivedPassword"`
	}

	type setRemoteUserRequest struct {
		XMLName    xml.Name       `xml:"tds:SetRemoteUser"`
		Xmlns      string         `xml:"xmlns:tds,attr"`
		RemoteUser *remoteUserXML `xml:"tds:RemoteUser,omitempty"`
	}

	req := setRemoteUserRequest{
		Xmlns: device.Namespace,
	}

	if remoteUser != nil {
		req.RemoteUser = &remoteUserXML{
			Username:           remoteUser.Username,
			Password:           remoteUser.Password,
			UseDerivedPassword: remoteUser.UseDerivedPassword,
		}
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetRemoteUser failed: %w", err)
	}

	return nil
}

// GetIPAddressFilter gets the IP address filter settings from a device.
func (s *Service) GetIPAddressFilter(ctx context.Context) (*IPAddressFilter, error) {
	type getIPAddressFilterRequest struct {
		XMLName xml.Name `xml:"tds:GetIPAddressFilter"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type prefixedAddressXML struct {
		Address      string `xml:"Address"`
		PrefixLength int    `xml:"PrefixLength"`
	}

	type getIPAddressFilterResponse struct {
		XMLName         xml.Name `xml:"GetIPAddressFilterResponse"`
		IPAddressFilter struct {
			Type        string               `xml:"Type"`
			IPv4Address []prefixedAddressXML `xml:"IPv4Address"`
			IPv6Address []prefixedAddressXML `xml:"IPv6Address"`
		} `xml:"IPAddressFilter"`
	}

	req := getIPAddressFilterRequest{
		Xmlns: device.Namespace,
	}

	var resp getIPAddressFilterResponse
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetIPAddressFilter failed: %w", err)
	}

	filter := &IPAddressFilter{
		Type: IPAddressFilterType(resp.IPAddressFilter.Type),
	}

	// Pre-allocate slices with known capacity
	if len(resp.IPAddressFilter.IPv4Address) > 0 {
		filter.IPv4Address = make([]types.PrefixedIPv4Address, 0, len(resp.IPAddressFilter.IPv4Address))
		for _, addr := range resp.IPAddressFilter.IPv4Address {
			filter.IPv4Address = append(filter.IPv4Address, types.PrefixedIPv4Address{
				Address: addr.Address, PrefixLength: addr.PrefixLength,
				Netmask: device.NetmaskFromPrefixLength(addr.PrefixLength),
			})
		}
	}

	if len(resp.IPAddressFilter.IPv6Address) > 0 {
		filter.IPv6Address = make([]types.PrefixedIPv6Address, 0, len(resp.IPAddressFilter.IPv6Address))
		for _, addr := range resp.IPAddressFilter.IPv6Address {
			filter.IPv6Address = append(filter.IPv6Address, types.PrefixedIPv6Address(addr))
		}
	}

	return filter, nil
}

// SetIPAddressFilter sets the IP address filter settings on a device.
func (s *Service) SetIPAddressFilter(ctx context.Context, filter *IPAddressFilter) error {
	type setIPAddressFilterRequest struct {
		XMLName         xml.Name               `xml:"tds:SetIPAddressFilter"`
		Xmlns           string                 `xml:"xmlns:tds,attr"`
		IPAddressFilter ipAddressFilterRequest `xml:"tds:IPAddressFilter"`
	}

	req := setIPAddressFilterRequest{
		Xmlns:           device.Namespace,
		IPAddressFilter: buildIPAddressFilterRequest(filter),
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetIPAddressFilter failed: %w", err)
	}

	return nil
}

// AddIPAddressFilter adds an IP filter address to a device.
func (s *Service) AddIPAddressFilter(ctx context.Context, filter *IPAddressFilter) error {
	type addIPAddressFilterRequest struct {
		XMLName         xml.Name               `xml:"tds:AddIPAddressFilter"`
		Xmlns           string                 `xml:"xmlns:tds,attr"`
		IPAddressFilter ipAddressFilterRequest `xml:"tds:IPAddressFilter"`
	}

	req := addIPAddressFilterRequest{
		Xmlns:           device.Namespace,
		IPAddressFilter: buildIPAddressFilterRequest(filter),
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("AddIPAddressFilter failed: %w", err)
	}

	return nil
}

// RemoveIPAddressFilter deletes an IP filter address from a device.
func (s *Service) RemoveIPAddressFilter(ctx context.Context, filter *IPAddressFilter) error {
	type removeIPAddressFilterRequest struct {
		XMLName         xml.Name               `xml:"tds:RemoveIPAddressFilter"`
		Xmlns           string                 `xml:"xmlns:tds,attr"`
		IPAddressFilter ipAddressFilterRequest `xml:"tds:IPAddressFilter"`
	}

	req := removeIPAddressFilterRequest{
		Xmlns:           device.Namespace,
		IPAddressFilter: buildIPAddressFilterRequest(filter),
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("RemoveIPAddressFilter failed: %w", err)
	}

	return nil
}

func (s *Service) GetPasswordComplexityConfiguration(ctx context.Context) (*PasswordComplexityConfiguration, error) {
	type getPasswordComplexityConfigurationRequest struct {
		XMLName xml.Name `xml:"tds:GetPasswordComplexityConfiguration"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type getPasswordComplexityConfigurationResponse struct {
		XMLName                   xml.Name `xml:"GetPasswordComplexityConfigurationResponse"`
		MinLen                    int      `xml:"MinLen"`
		Uppercase                 int      `xml:"Uppercase"`
		Number                    int      `xml:"Number"`
		SpecialChars              int      `xml:"SpecialChars"`
		BlockUsernameOccurrence   bool     `xml:"BlockUsernameOccurrence"`
		PolicyConfigurationLocked bool     `xml:"PolicyConfigurationLocked"`
	}

	req := getPasswordComplexityConfigurationRequest{
		Xmlns: device.Namespace,
	}

	var resp getPasswordComplexityConfigurationResponse
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetPasswordComplexityConfiguration failed: %w", err)
	}

	return &PasswordComplexityConfiguration{
		MinLen:                    resp.MinLen,
		Uppercase:                 resp.Uppercase,
		Number:                    resp.Number,
		SpecialChars:              resp.SpecialChars,
		BlockUsernameOccurrence:   resp.BlockUsernameOccurrence,
		PolicyConfigurationLocked: resp.PolicyConfigurationLocked,
	}, nil
}

// SetPasswordComplexityConfiguration allows setting of the password complexity configuration.
func (s *Service) SetPasswordComplexityConfiguration(
	ctx context.Context,
	config *PasswordComplexityConfiguration,
) error {
	type setPasswordComplexityConfigurationRequest struct {
		XMLName                   xml.Name `xml:"tds:SetPasswordComplexityConfiguration"`
		Xmlns                     string   `xml:"xmlns:tds,attr"`
		MinLen                    int      `xml:"tds:MinLen,omitempty"`
		Uppercase                 int      `xml:"tds:Uppercase,omitempty"`
		Number                    int      `xml:"tds:Number,omitempty"`
		SpecialChars              int      `xml:"tds:SpecialChars,omitempty"`
		BlockUsernameOccurrence   bool     `xml:"tds:BlockUsernameOccurrence,omitempty"`
		PolicyConfigurationLocked bool     `xml:"tds:PolicyConfigurationLocked,omitempty"`
	}

	req := setPasswordComplexityConfigurationRequest{
		Xmlns:                     device.Namespace,
		MinLen:                    config.MinLen,
		Uppercase:                 config.Uppercase,
		Number:                    config.Number,
		SpecialChars:              config.SpecialChars,
		BlockUsernameOccurrence:   config.BlockUsernameOccurrence,
		PolicyConfigurationLocked: config.PolicyConfigurationLocked,
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetPasswordComplexityConfiguration failed: %w", err)
	}

	return nil
}

// GetPasswordHistoryConfiguration retrieves the current password history configuration settings.
func (s *Service) GetPasswordHistoryConfiguration(ctx context.Context) (*PasswordHistoryConfiguration, error) {
	type getPasswordHistoryConfigurationRequest struct {
		XMLName xml.Name `xml:"tds:GetPasswordHistoryConfiguration"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type getPasswordHistoryConfigurationResponse struct {
		XMLName xml.Name `xml:"GetPasswordHistoryConfigurationResponse"`
		Enabled bool     `xml:"Enabled"`
		Length  int      `xml:"Length"`
	}

	req := getPasswordHistoryConfigurationRequest{
		Xmlns: device.Namespace,
	}

	var resp getPasswordHistoryConfigurationResponse
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetPasswordHistoryConfiguration failed: %w", err)
	}

	return &PasswordHistoryConfiguration{
		Enabled: resp.Enabled,
		Length:  resp.Length,
	}, nil
}

// SetPasswordHistoryConfiguration allows setting of the password history configuration.
func (s *Service) SetPasswordHistoryConfiguration(ctx context.Context, config *PasswordHistoryConfiguration) error {
	type setPasswordHistoryConfigurationRequest struct {
		XMLName xml.Name `xml:"tds:SetPasswordHistoryConfiguration"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
		Enabled bool     `xml:"tds:Enabled"`
		Length  int      `xml:"tds:Length"`
	}

	req := setPasswordHistoryConfigurationRequest{
		Xmlns:   device.Namespace,
		Enabled: config.Enabled,
		Length:  config.Length,
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetPasswordHistoryConfiguration failed: %w", err)
	}

	return nil
}

// GetAuthFailureWarningConfiguration retrieves the current authentication failure warning configuration.
func (s *Service) GetAuthFailureWarningConfiguration(ctx context.Context) (*AuthFailureWarningConfiguration, error) {
	type getAuthFailureWarningConfigurationRequest struct {
		XMLName xml.Name `xml:"tds:GetAuthFailureWarningConfiguration"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type getAuthFailureWarningConfigurationResponse struct {
		XMLName         xml.Name `xml:"GetAuthFailureWarningConfigurationResponse"`
		Enabled         bool     `xml:"Enabled"`
		MonitorPeriod   int      `xml:"MonitorPeriod"`
		MaxAuthFailures int      `xml:"MaxAuthFailures"`
	}

	req := getAuthFailureWarningConfigurationRequest{
		Xmlns: device.Namespace,
	}

	var resp getAuthFailureWarningConfigurationResponse
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAuthFailureWarningConfiguration failed: %w", err)
	}

	return &AuthFailureWarningConfiguration{
		Enabled:         resp.Enabled,
		MonitorPeriod:   resp.MonitorPeriod,
		MaxAuthFailures: resp.MaxAuthFailures,
	}, nil
}

// SetAuthFailureWarningConfiguration allows setting of the authentication failure warning configuration.
func (s *Service) SetAuthFailureWarningConfiguration(
	ctx context.Context,
	config *AuthFailureWarningConfiguration,
) error {
	type setAuthFailureWarningConfigurationRequest struct {
		XMLName         xml.Name `xml:"tds:SetAuthFailureWarningConfiguration"`
		Xmlns           string   `xml:"xmlns:tds,attr"`
		Enabled         bool     `xml:"tds:Enabled"`
		MonitorPeriod   int      `xml:"tds:MonitorPeriod"`
		MaxAuthFailures int      `xml:"tds:MaxAuthFailures"`
	}

	req := setAuthFailureWarningConfigurationRequest{
		Xmlns:           device.Namespace,
		Enabled:         config.Enabled,
		MonitorPeriod:   config.MonitorPeriod,
		MaxAuthFailures: config.MaxAuthFailures,
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetAuthFailureWarningConfiguration failed: %w", err)
	}

	return nil
}

// GetCertificates retrieves certificates. ONVIF Specification: GetCertificates operation.
func (s *Service) GetCertificates(ctx context.Context) ([]*Certificate, error) {
	type GetCertificatesBody struct {
		XMLName xml.Name `xml:"tds:GetCertificates"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetCertificatesResponse struct {
		XMLName      xml.Name       `xml:"GetCertificatesResponse"`
		Certificates []*Certificate `xml:"Certificate"`
	}

	request := GetCertificatesBody{
		Xmlns: device.Namespace,
	}
	var response GetCertificatesResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return nil, fmt.Errorf("GetCertificates failed: %w", err)
	}

	return response.Certificates, nil
}

// GetCACertificates retrieves CA certificates. ONVIF Specification: GetCACertificates operation.
func (s *Service) GetCACertificates(ctx context.Context) ([]*Certificate, error) {
	type GetCACertificatesBody struct {
		XMLName xml.Name `xml:"tds:GetCACertificates"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetCACertificatesResponse struct {
		XMLName      xml.Name       `xml:"GetCACertificatesResponse"`
		Certificates []*Certificate `xml:"Certificate"`
	}

	request := GetCACertificatesBody{
		Xmlns: device.Namespace,
	}
	var response GetCACertificatesResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return nil, fmt.Errorf("GetCACertificates failed: %w", err)
	}

	return response.Certificates, nil
}

// LoadCertificates loads certificates. ONVIF Specification: LoadCertificates operation.
func (s *Service) LoadCertificates(ctx context.Context, certificates []*Certificate) error {
	type LoadCertificatesBody struct {
		XMLName     xml.Name       `xml:"tds:LoadCertificates"`
		Xmlns       string         `xml:"xmlns:tds,attr"`
		Certificate []*Certificate `xml:"tds:Certificate"`
	}

	type LoadCertificatesResponse struct {
		XMLName xml.Name `xml:"LoadCertificatesResponse"`
	}

	request := LoadCertificatesBody{
		Xmlns:       device.Namespace,
		Certificate: certificates,
	}
	var response LoadCertificatesResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return fmt.Errorf("LoadCertificates failed: %w", err)
	}

	return nil
}

// LoadCACertificates loads CA certificates. ONVIF Specification: LoadCACertificates operation.
func (s *Service) LoadCACertificates(ctx context.Context, certificates []*Certificate) error {
	type LoadCACertificatesBody struct {
		XMLName     xml.Name       `xml:"tds:LoadCACertificates"`
		Xmlns       string         `xml:"xmlns:tds,attr"`
		Certificate []*Certificate `xml:"tds:Certificate"`
	}

	type LoadCACertificatesResponse struct {
		XMLName xml.Name `xml:"LoadCACertificatesResponse"`
	}

	request := LoadCACertificatesBody{
		Xmlns:       device.Namespace,
		Certificate: certificates,
	}
	var response LoadCACertificatesResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return fmt.Errorf("LoadCACertificates failed: %w", err)
	}

	return nil
}

// CreateCertificate creates a certificate. ONVIF Specification: CreateCertificate operation.
func (s *Service) CreateCertificate(
	ctx context.Context,
	certificateID, subject, validNotBefore, validNotAfter string,
) (*Certificate, error) {
	type CreateCertificateBody struct {
		XMLName        xml.Name `xml:"tds:CreateCertificate"`
		Xmlns          string   `xml:"xmlns:tds,attr"`
		CertificateID  string   `xml:"tds:CertificateID,omitempty"`
		Subject        string   `xml:"tds:Subject"`
		ValidNotBefore string   `xml:"tds:ValidNotBefore"`
		ValidNotAfter  string   `xml:"tds:ValidNotAfter"`
	}

	type CreateCertificateResponse struct {
		XMLName     xml.Name     `xml:"CreateCertificateResponse"`
		Certificate *Certificate `xml:"Certificate"`
	}

	request := CreateCertificateBody{
		Xmlns:          device.Namespace,
		CertificateID:  certificateID,
		Subject:        subject,
		ValidNotBefore: validNotBefore,
		ValidNotAfter:  validNotAfter,
	}
	var response CreateCertificateResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return nil, fmt.Errorf("CreateCertificate failed: %w", err)
	}

	return response.Certificate, nil
}

// DeleteCertificates deletes certificates. ONVIF Specification: DeleteCertificates operation.
func (s *Service) DeleteCertificates(ctx context.Context, certificateIDs []string) error {
	type DeleteCertificatesBody struct {
		XMLName       xml.Name `xml:"tds:DeleteCertificates"`
		Xmlns         string   `xml:"xmlns:tds,attr"`
		CertificateID []string `xml:"tds:CertificateID"`
	}

	type DeleteCertificatesResponse struct {
		XMLName xml.Name `xml:"DeleteCertificatesResponse"`
	}

	request := DeleteCertificatesBody{
		Xmlns:         device.Namespace,
		CertificateID: certificateIDs,
	}
	var response DeleteCertificatesResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return fmt.Errorf("DeleteCertificates failed: %w", err)
	}

	return nil
}

// GetCertificateInformation retrieves certificate information.
// ONVIF Specification: GetCertificateInformation operation.
func (s *Service) GetCertificateInformation(ctx context.Context, certificateID string) (*CertificateInformation, error) {
	type GetCertificateInformationBody struct {
		XMLName       xml.Name `xml:"tds:GetCertificateInformation"`
		Xmlns         string   `xml:"xmlns:tds,attr"`
		CertificateID string   `xml:"tds:CertificateID"`
	}

	type GetCertificateInformationResponse struct {
		XMLName                xml.Name                `xml:"GetCertificateInformationResponse"`
		CertificateInformation *CertificateInformation `xml:"CertificateInformation"`
	}

	request := GetCertificateInformationBody{
		Xmlns:         device.Namespace,
		CertificateID: certificateID,
	}
	var response GetCertificateInformationResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return nil, fmt.Errorf("GetCertificateInformation failed: %w", err)
	}

	return response.CertificateInformation, nil
}

// GetCertificatesStatus retrieves certificate status. ONVIF Specification: GetCertificatesStatus operation.
func (s *Service) GetCertificatesStatus(ctx context.Context) ([]*CertificateStatus, error) {
	type GetCertificatesStatusBody struct {
		XMLName xml.Name `xml:"tds:GetCertificatesStatus"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetCertificatesStatusResponse struct {
		XMLName           xml.Name             `xml:"GetCertificatesStatusResponse"`
		CertificateStatus []*CertificateStatus `xml:"CertificateStatus"`
	}

	request := GetCertificatesStatusBody{
		Xmlns: device.Namespace,
	}
	var response GetCertificatesStatusResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return nil, fmt.Errorf("GetCertificatesStatus failed: %w", err)
	}

	return response.CertificateStatus, nil
}

// SetCertificatesStatus sets certificate status. ONVIF Specification: SetCertificatesStatus operation.
func (s *Service) SetCertificatesStatus(ctx context.Context, statuses []*CertificateStatus) error {
	type SetCertificatesStatusBody struct {
		XMLName           xml.Name             `xml:"tds:SetCertificatesStatus"`
		Xmlns             string               `xml:"xmlns:tds,attr"`
		CertificateStatus []*CertificateStatus `xml:"tds:CertificateStatus"`
	}

	type SetCertificatesStatusResponse struct {
		XMLName xml.Name `xml:"SetCertificatesStatusResponse"`
	}

	request := SetCertificatesStatusBody{
		Xmlns:             device.Namespace,
		CertificateStatus: statuses,
	}
	var response SetCertificatesStatusResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return fmt.Errorf("SetCertificatesStatus failed: %w", err)
	}

	return nil
}

// GetPkcs10Request retrieves a PKCS10 certificate request. ONVIF Specification: GetPkcs10Request operation.
func (s *Service) GetPkcs10Request(
	ctx context.Context,
	certificateID, subject string,
	attributes *BinaryData,
) (*BinaryData, error) {
	type GetPkcs10RequestBody struct {
		XMLName       xml.Name    `xml:"tds:GetPkcs10Request"`
		Xmlns         string      `xml:"xmlns:tds,attr"`
		CertificateID string      `xml:"tds:CertificateID,omitempty"`
		Subject       string      `xml:"tds:Subject"`
		Attributes    *BinaryData `xml:"tds:Attributes,omitempty"`
	}

	type GetPkcs10RequestResponse struct {
		XMLName       xml.Name    `xml:"GetPkcs10RequestResponse"`
		Pkcs10Request *BinaryData `xml:"Pkcs10Request"`
	}

	request := GetPkcs10RequestBody{
		Xmlns:         device.Namespace,
		CertificateID: certificateID,
		Subject:       subject,
		Attributes:    attributes,
	}
	var response GetPkcs10RequestResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return nil, fmt.Errorf("GetPkcs10Request failed: %w", err)
	}

	return response.Pkcs10Request, nil
}

// LoadCertificateWithPrivateKey loads a certificate with its private key.
// ONVIF Specification: LoadCertificateWithPrivateKey operation.
func (s *Service) LoadCertificateWithPrivateKey(
	ctx context.Context,
	certificates []*Certificate,
	privateKey []*BinaryData,
	certificateIDs []string,
) error {
	type LoadCertificateWithPrivateKeyBody struct {
		XMLName                   xml.Name `xml:"tds:LoadCertificateWithPrivateKey"`
		Xmlns                     string   `xml:"xmlns:tds,attr"`
		CertificateWithPrivateKey []struct {
			CertificateID string       `xml:"CertificateID"`
			Certificate   *Certificate `xml:"Certificate"`
			PrivateKey    *BinaryData  `xml:"PrivateKey"`
		} `xml:"tds:CertificateWithPrivateKey"`
	}

	type LoadCertificateWithPrivateKeyResponse struct {
		XMLName xml.Name `xml:"LoadCertificateWithPrivateKeyResponse"`
	}

	request := LoadCertificateWithPrivateKeyBody{
		Xmlns: device.Namespace,
	}

	// Build certificate with private key array
	for i := range certificates {
		item := struct {
			CertificateID string       `xml:"CertificateID"`
			Certificate   *Certificate `xml:"Certificate"`
			PrivateKey    *BinaryData  `xml:"PrivateKey"`
		}{
			CertificateID: certificateIDs[i],
			Certificate:   certificates[i],
		}
		if i < len(privateKey) {
			item.PrivateKey = privateKey[i]
		}
		request.CertificateWithPrivateKey = append(request.CertificateWithPrivateKey, item)
	}

	var response LoadCertificateWithPrivateKeyResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return fmt.Errorf("LoadCertificateWithPrivateKey failed: %w", err)
	}

	return nil
}

// GetClientCertificateMode retrieves the client certificate mode.
// ONVIF Specification: GetClientCertificateMode operation.
func (s *Service) GetClientCertificateMode(ctx context.Context) (bool, error) {
	type GetClientCertificateModeBody struct {
		XMLName xml.Name `xml:"tds:GetClientCertificateMode"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetClientCertificateModeResponse struct {
		XMLName xml.Name `xml:"GetClientCertificateModeResponse"`
		Enabled bool     `xml:"Enabled"`
	}

	request := GetClientCertificateModeBody{
		Xmlns: device.Namespace,
	}
	var response GetClientCertificateModeResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return false, fmt.Errorf("GetClientCertificateMode failed: %w", err)
	}

	return response.Enabled, nil
}

// SetClientCertificateMode sets the client certificate mode. ONVIF Specification: SetClientCertificateMode operation.
func (s *Service) SetClientCertificateMode(ctx context.Context, enabled bool) error {
	type SetClientCertificateModeBody struct {
		XMLName xml.Name `xml:"tds:SetClientCertificateMode"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
		Enabled bool     `xml:"tds:Enabled"`
	}

	type SetClientCertificateModeResponse struct {
		XMLName xml.Name `xml:"SetClientCertificateModeResponse"`
	}

	request := SetClientCertificateModeBody{
		Xmlns:   device.Namespace,
		Enabled: enabled,
	}
	var response SetClientCertificateModeResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return fmt.Errorf("SetClientCertificateMode failed: %w", err)
	}

	return nil
}
