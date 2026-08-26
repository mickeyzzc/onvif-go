package onvif

import (
	"context"
	"encoding/xml"
	"fmt"
)

// Request/response types hoisted from method bodies.

type CreateOSD struct {
	XMLName                       xml.Name `xml:"trt:CreateOSD"`
	Xmlns                         string   `xml:"xmlns:trt,attr"`
	Xmlnst                        string   `xml:"xmlns:tt,attr"`
	VideoSourceConfigurationToken string   `xml:"trt:VideoSourceConfigurationToken"`
	OSD                           struct {
		Token string `xml:"token,attr,omitempty"`
	} `xml:"trt:OSD"`
}

type CreateOSDResponse struct {
	XMLName xml.Name `xml:"CreateOSDResponse"`
	OSD     struct {
		Token string `xml:"token,attr"`
	} `xml:"OSD"`
}

type DeleteOSD struct {
	XMLName  xml.Name `xml:"trt:DeleteOSD"`
	Xmlns    string   `xml:"xmlns:trt,attr"`
	OSDToken string   `xml:"trt:OSDToken"`
}

type GetOSD struct {
	XMLName  xml.Name `xml:"trt:GetOSD"`
	Xmlns    string   `xml:"xmlns:trt,attr"`
	OSDToken string   `xml:"trt:OSDToken"`
}

type GetOSDOptions struct {
	XMLName            xml.Name `xml:"trt:GetOSDOptions"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
}

type GetOSDOptionsResponse struct {
	XMLName xml.Name `xml:"GetOSDOptionsResponse"`
	Options struct {
		MaximumNumberOfOSDs int `xml:"MaximumNumberOfOSDs"`
	} `xml:"Options"`
}

type GetOSDResponse struct {
	XMLName xml.Name `xml:"GetOSDResponse"`
	OSD     struct {
		Token string `xml:"token,attr"`
	} `xml:"OSD"`
}

type GetOSDs struct {
	XMLName            xml.Name `xml:"trt:GetOSDs"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
}

type GetOSDsResponse struct {
	XMLName xml.Name `xml:"GetOSDsResponse"`
	OSDs    []struct {
		Token string `xml:"token,attr"`
	} `xml:"OSDs"`
}

type SetOSD struct {
	XMLName xml.Name `xml:"trt:SetOSD"`
	Xmlns   string   `xml:"xmlns:trt,attr"`
	Xmlnst  string   `xml:"xmlns:tt,attr"`
	OSD     struct {
		Token string `xml:"token,attr"`
	} `xml:"trt:OSD"`
}

func (s *MediaService) GetOSDs(ctx context.Context, configurationToken string) ([]*OSDConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetOSDs{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}

	var resp GetOSDsResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetOSDs failed: %w", err)
	}

	osds := make([]*OSDConfiguration, len(resp.OSDs))
	for i, o := range resp.OSDs {
		osds[i] = &OSDConfiguration{
			Token: o.Token,
		}
	}

	return osds, nil
}

func (s *MediaService) GetOSD(ctx context.Context, osdToken string) (*OSDConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetOSD{
		Xmlns:    mediaNamespace,
		OSDToken: osdToken,
	}

	var resp GetOSDResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetOSD failed: %w", err)
	}

	return &OSDConfiguration{
		Token: resp.OSD.Token,
	}, nil
}

func (s *MediaService) SetOSD(ctx context.Context, osd *OSDConfiguration) error {
	endpoint := s.getMediaEndpoint()

	req := SetOSD{
		Xmlns:  mediaNamespace,
		Xmlnst: "http://www.onvif.org/ver10/schema",
	}
	req.OSD.Token = osd.Token

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetOSD failed: %w", err)
	}

	return nil
}

func (s *MediaService) CreateOSD(
	ctx context.Context,
	videoSourceConfigurationToken string,
	osd *OSDConfiguration,
) (*OSDConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := CreateOSD{
		Xmlns:                         mediaNamespace,
		Xmlnst:                        "http://www.onvif.org/ver10/schema",
		VideoSourceConfigurationToken: videoSourceConfigurationToken,
	}
	if osd != nil && osd.Token != "" {
		req.OSD.Token = osd.Token
	}

	var resp CreateOSDResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("CreateOSD failed: %w", err)
	}

	return &OSDConfiguration{
		Token: resp.OSD.Token,
	}, nil
}

func (s *MediaService) DeleteOSD(ctx context.Context, osdToken string) error {
	endpoint := s.getMediaEndpoint()

	req := DeleteOSD{
		Xmlns:    mediaNamespace,
		OSDToken: osdToken,
	}

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("DeleteOSD failed: %w", err)
	}

	return nil
}

func (s *MediaService) GetOSDOptions(ctx context.Context, configurationToken string) (*OSDConfigurationOptions, error) {
	endpoint := s.getMediaEndpoint()

	req := GetOSDOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}

	var resp GetOSDOptionsResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetOSDOptions failed: %w", err)
	}

	return &OSDConfigurationOptions{
		MaximumNumberOfOSDs: resp.Options.MaximumNumberOfOSDs,
	}, nil
}
