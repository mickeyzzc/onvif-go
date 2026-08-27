package media

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
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

func (s *Service) GetOSDs(ctx context.Context, configurationToken string) ([]*OSDConfiguration, error) {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := GetOSDs{
		Xmlns: Namespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}

	var resp GetOSDsResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
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

func (s *Service) GetOSD(ctx context.Context, osdToken string) (*OSDConfiguration, error) {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := GetOSD{
		Xmlns:    Namespace,
		OSDToken: osdToken,
	}

	var resp GetOSDResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetOSD failed: %w", err)
	}

	return &OSDConfiguration{
		Token: resp.OSD.Token,
	}, nil
}

func (s *Service) SetOSD(ctx context.Context, osd *OSDConfiguration) error {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := SetOSD{
		Xmlns:  Namespace,
		Xmlnst: "http://www.onvif.org/ver10/schema",
	}
	req.OSD.Token = osd.Token

	if err := s.c.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetOSD failed: %w", err)
	}

	return nil
}

func (s *Service) CreateOSD(
	ctx context.Context,
	videoSourceConfigurationToken string,
	osd *OSDConfiguration,
) (*OSDConfiguration, error) {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := CreateOSD{
		Xmlns:                         Namespace,
		Xmlnst:                        "http://www.onvif.org/ver10/schema",
		VideoSourceConfigurationToken: videoSourceConfigurationToken,
	}
	if osd != nil && osd.Token != "" {
		req.OSD.Token = osd.Token
	}

	var resp CreateOSDResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("CreateOSD failed: %w", err)
	}

	return &OSDConfiguration{
		Token: resp.OSD.Token,
	}, nil
}

func (s *Service) DeleteOSD(ctx context.Context, osdToken string) error {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := DeleteOSD{
		Xmlns:    Namespace,
		OSDToken: osdToken,
	}

	if err := s.c.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("DeleteOSD failed: %w", err)
	}

	return nil
}

func (s *Service) GetOSDOptions(ctx context.Context, configurationToken string) (*OSDConfigurationOptions, error) {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := GetOSDOptions{
		Xmlns: Namespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}

	var resp GetOSDOptionsResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetOSDOptions failed: %w", err)
	}

	return &OSDConfigurationOptions{
		MaximumNumberOfOSDs: resp.Options.MaximumNumberOfOSDs,
	}, nil
}
