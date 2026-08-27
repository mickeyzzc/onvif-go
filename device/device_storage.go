package device

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
)

// GetStorageConfigurations retrieves storage configurations. ONVIF Specification: GetStorageConfigurations operation.
func (s *Service) GetStorageConfigurations(ctx context.Context) ([]*StorageConfiguration, error) {
	type GetStorageConfigurationsBody struct {
		XMLName xml.Name `xml:"tds:GetStorageConfigurations"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetStorageConfigurationsResponse struct {
		XMLName               xml.Name                `xml:"GetStorageConfigurationsResponse"`
		StorageConfigurations []*StorageConfiguration `xml:"StorageConfigurations"`
	}

	request := GetStorageConfigurationsBody{
		Xmlns: Namespace,
	}
	var response GetStorageConfigurationsResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return nil, fmt.Errorf("GetStorageConfigurations failed: %w", err)
	}

	return response.StorageConfigurations, nil
}

// GetStorageConfiguration retrieves a storage configuration. ONVIF Specification: GetStorageConfiguration operation.
func (s *Service) GetStorageConfiguration(ctx context.Context, token string) (*StorageConfiguration, error) {
	type GetStorageConfigurationBody struct {
		XMLName xml.Name `xml:"tds:GetStorageConfiguration"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
		Token   string   `xml:"tds:Token"`
	}

	type GetStorageConfigurationResponse struct {
		XMLName              xml.Name              `xml:"GetStorageConfigurationResponse"`
		StorageConfiguration *StorageConfiguration `xml:"StorageConfiguration"`
	}

	request := GetStorageConfigurationBody{
		Xmlns: Namespace,
		Token: token,
	}
	var response GetStorageConfigurationResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return nil, fmt.Errorf("GetStorageConfiguration failed: %w", err)
	}

	return response.StorageConfiguration, nil
}

// CreateStorageConfiguration creates a storage configuration.
// ONVIF Specification: CreateStorageConfiguration operation.
func (s *Service) CreateStorageConfiguration(ctx context.Context, config *StorageConfiguration) (string, error) {
	type CreateStorageConfigurationBody struct {
		XMLName              xml.Name              `xml:"tds:CreateStorageConfiguration"`
		Xmlns                string                `xml:"xmlns:tds,attr"`
		StorageConfiguration *StorageConfiguration `xml:"tds:StorageConfiguration"`
	}

	type CreateStorageConfigurationResponse struct {
		XMLName xml.Name `xml:"CreateStorageConfigurationResponse"`
		Token   string   `xml:"Token"`
	}

	request := CreateStorageConfigurationBody{
		Xmlns:                Namespace,
		StorageConfiguration: config,
	}
	var response CreateStorageConfigurationResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return "", fmt.Errorf("CreateStorageConfiguration failed: %w", err)
	}

	return response.Token, nil
}

// SetStorageConfiguration sets a storage configuration. ONVIF Specification: SetStorageConfiguration operation.
func (s *Service) SetStorageConfiguration(ctx context.Context, config *StorageConfiguration) error {
	type SetStorageConfigurationBody struct {
		XMLName              xml.Name              `xml:"tds:SetStorageConfiguration"`
		Xmlns                string                `xml:"xmlns:tds,attr"`
		StorageConfiguration *StorageConfiguration `xml:"tds:StorageConfiguration"`
	}

	type SetStorageConfigurationResponse struct {
		XMLName xml.Name `xml:"SetStorageConfigurationResponse"`
	}

	request := SetStorageConfigurationBody{
		Xmlns:                Namespace,
		StorageConfiguration: config,
	}
	var response SetStorageConfigurationResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return fmt.Errorf("SetStorageConfiguration failed: %w", err)
	}

	return nil
}

// DeleteStorageConfiguration deletes a storage configuration.
// ONVIF Specification: DeleteStorageConfiguration operation.
func (s *Service) DeleteStorageConfiguration(ctx context.Context, token string) error {
	type DeleteStorageConfigurationBody struct {
		XMLName xml.Name `xml:"tds:DeleteStorageConfiguration"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
		Token   string   `xml:"tds:Token"`
	}

	type DeleteStorageConfigurationResponse struct {
		XMLName xml.Name `xml:"DeleteStorageConfigurationResponse"`
	}

	request := DeleteStorageConfigurationBody{
		Xmlns: Namespace,
		Token: token,
	}
	var response DeleteStorageConfigurationResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return fmt.Errorf("DeleteStorageConfiguration failed: %w", err)
	}

	return nil
}

// SetHashingAlgorithm sets the hashing algorithm. ONVIF Specification: SetHashingAlgorithm operation.
func (s *Service) SetHashingAlgorithm(ctx context.Context, algorithm string) error {
	type SetHashingAlgorithmBody struct {
		XMLName   xml.Name `xml:"tds:SetHashingAlgorithm"`
		Xmlns     string   `xml:"xmlns:tds,attr"`
		Algorithm string   `xml:"tds:Algorithm"`
	}

	type SetHashingAlgorithmResponse struct {
		XMLName xml.Name `xml:"SetHashingAlgorithmResponse"`
	}

	request := SetHashingAlgorithmBody{
		Xmlns:     Namespace,
		Algorithm: algorithm,
	}
	var response SetHashingAlgorithmResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return fmt.Errorf("SetHashingAlgorithm failed: %w", err)
	}

	return nil
}
