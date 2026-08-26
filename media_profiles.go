package onvif

import (
	"context"
	"encoding/xml"
	"fmt"
)

const mediaNamespace = "http://www.onvif.org/ver10/media/wsdl"

// Request/response types hoisted from method bodies.

type AddPTZConfiguration struct {
	XMLName            xml.Name `xml:"trt:AddPTZConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ProfileToken       string   `xml:"trt:ProfileToken"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type AddVideoAnalyticsConfiguration struct {
	XMLName            xml.Name `xml:"trt:AddVideoAnalyticsConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ProfileToken       string   `xml:"trt:ProfileToken"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type CreateProfile struct {
	XMLName xml.Name `xml:"trt:CreateProfile"`
	Xmlns   string   `xml:"xmlns:trt,attr"`
	Name    string   `xml:"trt:Name"`
	Token   *string  `xml:"trt:Token,omitempty"`
}

type CreateProfileResponse struct {
	XMLName xml.Name `xml:"CreateProfileResponse"`
	Profile struct {
		Token string `xml:"token,attr"`
		Name  string `xml:"Name"`
	} `xml:"Profile"`
}

type DeleteProfile struct {
	XMLName      xml.Name `xml:"trt:DeleteProfile"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type GetCompatiblePTZConfigurations struct {
	XMLName      xml.Name `xml:"trt:GetCompatiblePTZConfigurations"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type GetCompatiblePTZConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetCompatiblePTZConfigurationsResponse"`
	Configurations []struct {
		Token     string `xml:"token,attr"`
		Name      string `xml:"Name"`
		UseCount  int    `xml:"UseCount"`
		NodeToken string `xml:"NodeToken"`
	} `xml:"Configurations"`
}

type GetCompatibleVideoAnalyticsConfigurations struct {
	XMLName      xml.Name `xml:"trt:GetCompatibleVideoAnalyticsConfigurations"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type GetCompatibleVideoAnalyticsConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetCompatibleVideoAnalyticsConfigurationsResponse"`
	Configurations []struct {
		Token    string `xml:"token,attr"`
		Name     string `xml:"Name"`
		UseCount int    `xml:"UseCount"`
	} `xml:"Configurations"`
}

type GetProfile struct {
	XMLName      xml.Name `xml:"trt:GetProfile"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type GetProfileResponse struct {
	XMLName xml.Name `xml:"GetProfileResponse"`
	Profile struct {
		Token string `xml:"token,attr"`
		Name  string `xml:"Name"`
	} `xml:"Profile"`
}

type GetProfiles struct {
	XMLName xml.Name `xml:"trt:GetProfiles"`
	Xmlns   string   `xml:"xmlns:trt,attr"`
}

type GetProfilesResponse struct {
	XMLName  xml.Name `xml:"GetProfilesResponse"`
	Profiles []struct {
		Token                    string `xml:"token,attr"`
		Name                     string `xml:"Name"`
		VideoSourceConfiguration *struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			SourceToken string `xml:"SourceToken"`
			Bounds      *struct {
				X      int `xml:"x,attr"`
				Y      int `xml:"y,attr"`
				Width  int `xml:"width,attr"`
				Height int `xml:"height,attr"`
			} `xml:"Bounds"`
		} `xml:"VideoSourceConfiguration"`
		VideoEncoderConfiguration *struct {
			Token      string `xml:"token,attr"`
			Name       string `xml:"Name"`
			UseCount   int    `xml:"UseCount"`
			Encoding   string `xml:"Encoding"`
			Resolution *struct {
				Width  int `xml:"Width"`
				Height int `xml:"Height"`
			} `xml:"Resolution"`
			Quality     float64 `xml:"Quality"`
			RateControl *struct {
				FrameRateLimit   int `xml:"FrameRateLimit"`
				EncodingInterval int `xml:"EncodingInterval"`
				BitrateLimit     int `xml:"BitrateLimit"`
			} `xml:"RateControl"`
		} `xml:"VideoEncoderConfiguration"`
		PTZConfiguration *struct {
			Token     string `xml:"token,attr"`
			Name      string `xml:"Name"`
			UseCount  int    `xml:"UseCount"`
			NodeToken string `xml:"NodeToken"`
		} `xml:"PTZConfiguration"`
	} `xml:"Profiles"`
}

type GetServiceCapabilities struct {
	XMLName xml.Name `xml:"trt:GetServiceCapabilities"`
	Xmlns   string   `xml:"xmlns:trt,attr"`
}

type GetServiceCapabilitiesResponse struct {
	XMLName      xml.Name `xml:"GetServiceCapabilitiesResponse"`
	Capabilities struct {
		SnapshotURI         bool `xml:"SnapshotUri,attr"`
		Rotation            bool `xml:"Rotation,attr"`
		VideoSourceMode     bool `xml:"VideoSourceMode,attr"`
		OSD                 bool `xml:"OSD,attr"`
		TemporaryOSDText    bool `xml:"TemporaryOSDText,attr"`
		EXICompression      bool `xml:"EXICompression,attr"`
		ProfileCapabilities *struct {
			MaximumNumberOfProfiles int `xml:"MaximumNumberOfProfiles,attr"`
		} `xml:"ProfileCapabilities"`
		StreamingCapabilities *struct {
			RTPMulticast bool `xml:"RTPMulticast,attr"`
			RTPTCP       bool `xml:"RTP_TCP,attr"`
			RTPRTSPTCP   bool `xml:"RTP_RTSP_TCP,attr"`
		} `xml:"StreamingCapabilities"`
	} `xml:"Capabilities"`
}

type GetVideoAnalyticsConfiguration struct {
	XMLName            xml.Name `xml:"trt:GetVideoAnalyticsConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type GetVideoAnalyticsConfigurationOptions struct {
	XMLName            xml.Name `xml:"trt:GetVideoAnalyticsConfigurationOptions"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
	ProfileToken       string   `xml:"trt:ProfileToken,omitempty"`
}

type GetVideoAnalyticsConfigurationOptionsResponse struct {
	XMLName xml.Name `xml:"GetVideoAnalyticsConfigurationOptionsResponse"`
	Options struct{} `xml:"Options"`
}

type GetVideoAnalyticsConfigurationResponse struct {
	XMLName       xml.Name `xml:"GetVideoAnalyticsConfigurationResponse"`
	Configuration struct {
		Token    string `xml:"token,attr"`
		Name     string `xml:"Name"`
		UseCount int    `xml:"UseCount"`
	} `xml:"Configuration"`
}

type GetVideoAnalyticsConfigurations struct {
	XMLName xml.Name `xml:"trt:GetVideoAnalyticsConfigurations"`
	Xmlns   string   `xml:"xmlns:trt,attr"`
}

type GetVideoAnalyticsConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetVideoAnalyticsConfigurationsResponse"`
	Configurations []struct {
		Token    string `xml:"token,attr"`
		Name     string `xml:"Name"`
		UseCount int    `xml:"UseCount"`
	} `xml:"Configurations"`
}

type RemovePTZConfiguration struct {
	XMLName      xml.Name `xml:"trt:RemovePTZConfiguration"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type RemoveVideoAnalyticsConfiguration struct {
	XMLName      xml.Name `xml:"trt:RemoveVideoAnalyticsConfiguration"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type SetProfile struct {
	XMLName xml.Name `xml:"trt:SetProfile"`
	Xmlns   string   `xml:"xmlns:trt,attr"`
	Xmlnst  string   `xml:"xmlns:tt,attr"`
	Profile struct {
		Token string `xml:"token,attr"`
		Name  string `xml:"tt:Name"`
	} `xml:"trt:Profile"`
}

type SetVideoAnalyticsConfiguration struct {
	XMLName       xml.Name `xml:"trt:SetVideoAnalyticsConfiguration"`
	Xmlns         string   `xml:"xmlns:trt,attr"`
	Xmlnst        string   `xml:"xmlns:tt,attr"`
	Configuration struct {
		Token    string `xml:"token,attr"`
		Name     string `xml:"tt:Name"`
		UseCount int    `xml:"tt:UseCount"`
	} `xml:"trt:Configuration"`
	ForcePersistence bool `xml:"trt:ForcePersistence"`
}

func (s *MediaService) getMediaEndpoint() string {
	if s.client.mediaEndpoint != "" {
		return s.client.mediaEndpoint
	}

	return s.client.endpoint
}

func (s *MediaService) GetProfiles(ctx context.Context) ([]*Profile, error) {
	endpoint := s.getMediaEndpoint()

	req := GetProfiles{
		Xmlns: mediaNamespace,
	}

	var resp GetProfilesResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetProfiles failed: %w", err)
	}

	profiles := make([]*Profile, len(resp.Profiles))
	for i, p := range resp.Profiles {
		profile := &Profile{
			Token: p.Token,
			Name:  p.Name,
		}

		if p.VideoSourceConfiguration != nil {
			profile.VideoSourceConfiguration = &VideoSourceConfiguration{
				Token:       p.VideoSourceConfiguration.Token,
				Name:        p.VideoSourceConfiguration.Name,
				UseCount:    p.VideoSourceConfiguration.UseCount,
				SourceToken: p.VideoSourceConfiguration.SourceToken,
			}
			if p.VideoSourceConfiguration.Bounds != nil {
				profile.VideoSourceConfiguration.Bounds = &IntRectangle{
					X:      p.VideoSourceConfiguration.Bounds.X,
					Y:      p.VideoSourceConfiguration.Bounds.Y,
					Width:  p.VideoSourceConfiguration.Bounds.Width,
					Height: p.VideoSourceConfiguration.Bounds.Height,
				}
			}
		}

		if p.VideoEncoderConfiguration != nil {
			profile.VideoEncoderConfiguration = &VideoEncoderConfiguration{
				Token:    p.VideoEncoderConfiguration.Token,
				Name:     p.VideoEncoderConfiguration.Name,
				UseCount: p.VideoEncoderConfiguration.UseCount,
				Encoding: p.VideoEncoderConfiguration.Encoding,
				Quality:  p.VideoEncoderConfiguration.Quality,
			}
			if p.VideoEncoderConfiguration.Resolution != nil {
				profile.VideoEncoderConfiguration.Resolution = &VideoResolution{
					Width:  p.VideoEncoderConfiguration.Resolution.Width,
					Height: p.VideoEncoderConfiguration.Resolution.Height,
				}
			}
			if p.VideoEncoderConfiguration.RateControl != nil {
				profile.VideoEncoderConfiguration.RateControl = &VideoRateControl{
					FrameRateLimit:   p.VideoEncoderConfiguration.RateControl.FrameRateLimit,
					EncodingInterval: p.VideoEncoderConfiguration.RateControl.EncodingInterval,
					BitrateLimit:     p.VideoEncoderConfiguration.RateControl.BitrateLimit,
				}
			}
		}

		if p.PTZConfiguration != nil {
			profile.PTZConfiguration = &PTZConfiguration{
				Token:     p.PTZConfiguration.Token,
				Name:      p.PTZConfiguration.Name,
				UseCount:  p.PTZConfiguration.UseCount,
				NodeToken: p.PTZConfiguration.NodeToken,
			}
		}

		profiles[i] = profile
	}

	return profiles, nil
}

func (s *MediaService) CreateProfile(ctx context.Context, name, token string) (*Profile, error) {
	endpoint := s.getMediaEndpoint()

	req := CreateProfile{
		Xmlns: mediaNamespace,
		Name:  name,
	}
	if token != "" {
		req.Token = &token
	}

	var resp CreateProfileResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("CreateProfile failed: %w", err)
	}

	return &Profile{
		Token: resp.Profile.Token,
		Name:  resp.Profile.Name,
	}, nil
}

func (s *MediaService) DeleteProfile(ctx context.Context, profileToken string) error {
	endpoint := s.getMediaEndpoint()

	req := DeleteProfile{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("DeleteProfile failed: %w", err)
	}

	return nil
}

func (s *MediaService) GetMediaServiceCapabilities(ctx context.Context) (*MediaServiceCapabilities, error) {
	endpoint := s.getMediaEndpoint()

	req := GetServiceCapabilities{
		Xmlns: mediaNamespace,
	}

	var resp GetServiceCapabilitiesResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetMediaServiceCapabilities failed: %w", err)
	}

	caps := &MediaServiceCapabilities{
		SnapshotURI:      resp.Capabilities.SnapshotURI,
		Rotation:         resp.Capabilities.Rotation,
		VideoSourceMode:  resp.Capabilities.VideoSourceMode,
		OSD:              resp.Capabilities.OSD,
		TemporaryOSDText: resp.Capabilities.TemporaryOSDText,
		EXICompression:   resp.Capabilities.EXICompression,
	}

	if resp.Capabilities.ProfileCapabilities != nil {
		caps.MaximumNumberOfProfiles = resp.Capabilities.ProfileCapabilities.MaximumNumberOfProfiles
	}

	if resp.Capabilities.StreamingCapabilities != nil {
		caps.RTPMulticast = resp.Capabilities.StreamingCapabilities.RTPMulticast
		caps.RTPTCP = resp.Capabilities.StreamingCapabilities.RTPTCP
		caps.RTPRTSPTCP = resp.Capabilities.StreamingCapabilities.RTPRTSPTCP
	}

	return caps, nil
}

func (s *MediaService) GetProfile(ctx context.Context, profileToken string) (*Profile, error) {
	endpoint := s.getMediaEndpoint()

	req := GetProfile{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetProfileResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetProfile failed: %w", err)
	}

	return &Profile{
		Token: resp.Profile.Token,
		Name:  resp.Profile.Name,
	}, nil
}

func (s *MediaService) SetProfile(ctx context.Context, profile *Profile) error {
	endpoint := s.getMediaEndpoint()

	req := SetProfile{
		Xmlns:  mediaNamespace,
		Xmlnst: "http://www.onvif.org/ver10/schema",
	}
	req.Profile.Token = profile.Token
	req.Profile.Name = profile.Name

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetProfile failed: %w", err)
	}

	return nil
}

func (s *MediaService) AddPTZConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := s.getMediaEndpoint()

	req := AddPTZConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddPTZConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) RemovePTZConfiguration(ctx context.Context, profileToken string) error {
	endpoint := s.getMediaEndpoint()

	req := RemovePTZConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemovePTZConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) GetCompatiblePTZConfigurations(ctx context.Context, profileToken string) ([]*PTZConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetCompatiblePTZConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatiblePTZConfigurationsResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatiblePTZConfigurations failed: %w", err)
	}

	configs := make([]*PTZConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &PTZConfiguration{
			Token:     cfg.Token,
			Name:      cfg.Name,
			UseCount:  cfg.UseCount,
			NodeToken: cfg.NodeToken,
		}
	}

	return configs, nil
}

func (s *MediaService) GetVideoAnalyticsConfigurations(ctx context.Context) ([]*VideoAnalyticsConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetVideoAnalyticsConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetVideoAnalyticsConfigurationsResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoAnalyticsConfigurations failed: %w", err)
	}

	configs := make([]*VideoAnalyticsConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &VideoAnalyticsConfiguration{
			Token:    cfg.Token,
			Name:     cfg.Name,
			UseCount: cfg.UseCount,
		}
	}

	return configs, nil
}

func (s *MediaService) GetVideoAnalyticsConfiguration(
	ctx context.Context,
	configurationToken string,
) (*VideoAnalyticsConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetVideoAnalyticsConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetVideoAnalyticsConfigurationResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoAnalyticsConfiguration failed: %w", err)
	}

	return &VideoAnalyticsConfiguration{
		Token:    resp.Configuration.Token,
		Name:     resp.Configuration.Name,
		UseCount: resp.Configuration.UseCount,
	}, nil
}

func (s *MediaService) GetCompatibleVideoAnalyticsConfigurations(ctx context.Context, profileToken string) ([]*VideoAnalyticsConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetCompatibleVideoAnalyticsConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleVideoAnalyticsConfigurationsResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleVideoAnalyticsConfigurations failed: %w", err)
	}

	configs := make([]*VideoAnalyticsConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &VideoAnalyticsConfiguration{
			Token:    cfg.Token,
			Name:     cfg.Name,
			UseCount: cfg.UseCount,
		}
	}

	return configs, nil
}

func (s *MediaService) SetVideoAnalyticsConfiguration(ctx context.Context, config *VideoAnalyticsConfiguration, forcePersistence bool) error {
	endpoint := s.getMediaEndpoint()

	req := SetVideoAnalyticsConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           "http://www.onvif.org/ver10/schema",
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetVideoAnalyticsConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) GetVideoAnalyticsConfigurationOptions(
	ctx context.Context,
	configurationToken, profileToken string,
) (*VideoAnalyticsConfigurationOptions, error) {
	endpoint := s.getMediaEndpoint()

	req := GetVideoAnalyticsConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}
	if profileToken != "" {
		req.ProfileToken = profileToken
	}

	var resp GetVideoAnalyticsConfigurationOptionsResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoAnalyticsConfigurationOptions failed: %w", err)
	}

	return &VideoAnalyticsConfigurationOptions{}, nil
}

func (s *MediaService) AddVideoAnalyticsConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := s.getMediaEndpoint()

	req := AddVideoAnalyticsConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddVideoAnalyticsConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) RemoveVideoAnalyticsConfiguration(ctx context.Context, profileToken string) error {
	endpoint := s.getMediaEndpoint()

	req := RemoveVideoAnalyticsConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveVideoAnalyticsConfiguration failed: %w", err)
	}

	return nil
}
