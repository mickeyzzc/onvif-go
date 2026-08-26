package onvif

import (
	"context"
	"encoding/xml"
	"fmt"
)

// Request/response types hoisted from method bodies.

type AddVideoEncoderConfiguration struct {
	XMLName            xml.Name `xml:"trt:AddVideoEncoderConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ProfileToken       string   `xml:"trt:ProfileToken"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type AddVideoSourceConfiguration struct {
	XMLName            xml.Name `xml:"trt:AddVideoSourceConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ProfileToken       string   `xml:"trt:ProfileToken"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type GetCompatibleVideoEncoderConfigurations struct {
	XMLName      xml.Name `xml:"trt:GetCompatibleVideoEncoderConfigurations"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type GetCompatibleVideoEncoderConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetCompatibleVideoEncoderConfigurationsResponse"`
	Configurations []struct {
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
	} `xml:"Configurations"`
}

type GetCompatibleVideoSourceConfigurations struct {
	XMLName      xml.Name `xml:"trt:GetCompatibleVideoSourceConfigurations"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type GetCompatibleVideoSourceConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetCompatibleVideoSourceConfigurationsResponse"`
	Configurations []struct {
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
	} `xml:"Configurations"`
}

type GetGuaranteedNumberOfVideoEncoderInstances struct {
	XMLName            xml.Name `xml:"trt:GetGuaranteedNumberOfVideoEncoderInstances"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type GetGuaranteedNumberOfVideoEncoderInstancesResponse struct {
	XMLName     xml.Name `xml:"GetGuaranteedNumberOfVideoEncoderInstancesResponse"`
	TotalNumber int      `xml:"TotalNumber"`
	JPEG        int      `xml:"JPEG"`
	H264        int      `xml:"H264"`
	MPEG4       int      `xml:"MPEG4"`
}

type GetVideoEncoderConfiguration struct {
	XMLName            xml.Name `xml:"trt:GetVideoEncoderConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type GetVideoEncoderConfigurationOptions struct {
	XMLName            xml.Name `xml:"trt:GetVideoEncoderConfigurationOptions"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
	ProfileToken       string   `xml:"trt:ProfileToken,omitempty"`
}

type GetVideoEncoderConfigurationOptionsResponse struct {
	XMLName xml.Name `xml:"GetVideoEncoderConfigurationOptionsResponse"`
	Options struct {
		QualityRange *struct {
			Min float64 `xml:"Min"`
			Max float64 `xml:"Max"`
		} `xml:"QualityRange"`
		JPEG *struct {
			ResolutionsAvailable []struct {
				Width  int `xml:"Width"`
				Height int `xml:"Height"`
			} `xml:"ResolutionsAvailable"`
			FrameRateRange *struct {
				Min float64 `xml:"Min"`
				Max float64 `xml:"Max"`
			} `xml:"FrameRateRange"`
			EncodingIntervalRange *struct {
				Min int `xml:"Min"`
				Max int `xml:"Max"`
			} `xml:"EncodingIntervalRange"`
		} `xml:"JPEG"`
		H264 *struct {
			ResolutionsAvailable []struct {
				Width  int `xml:"Width"`
				Height int `xml:"Height"`
			} `xml:"ResolutionsAvailable"`
			GovLengthRange *struct {
				Min int `xml:"Min"`
				Max int `xml:"Max"`
			} `xml:"GovLengthRange"`
			FrameRateRange *struct {
				Min float64 `xml:"Min"`
				Max float64 `xml:"Max"`
			} `xml:"FrameRateRange"`
			EncodingIntervalRange *struct {
				Min int `xml:"Min"`
				Max int `xml:"Max"`
			} `xml:"EncodingIntervalRange"`
			H264ProfilesSupported []string `xml:"H264ProfilesSupported"`
		} `xml:"H264"`
		Extension struct{} `xml:"Extension"`
	} `xml:"Options"`
}

type GetVideoEncoderConfigurationResponse struct {
	XMLName       xml.Name `xml:"GetVideoEncoderConfigurationResponse"`
	Configuration struct {
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
	} `xml:"Configuration"`
}

type GetVideoEncoderConfigurations struct {
	XMLName xml.Name `xml:"trt:GetVideoEncoderConfigurations"`
	Xmlns   string   `xml:"xmlns:trt,attr"`
}

type GetVideoEncoderConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetVideoEncoderConfigurationsResponse"`
	Configurations []struct {
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
		MPEG4 *struct {
			GovLength    int    `xml:"GovLength"`
			MPEG4Profile string `xml:"MPEG4Profile"`
		} `xml:"MPEG4"`
		H264 *struct {
			GovLength   int    `xml:"GovLength"`
			H264Profile string `xml:"H264Profile"`
		} `xml:"H264"`
		Multicast *struct {
			Address *struct {
				Type        string `xml:"Type"`
				IPv4Address string `xml:"IPv4Address"`
				IPv6Address string `xml:"IPv6Address"`
			} `xml:"Address"`
			Port      int  `xml:"Port"`
			TTL       int  `xml:"TTL"`
			AutoStart bool `xml:"AutoStart"`
		} `xml:"Multicast"`
		SessionTimeout string `xml:"SessionTimeout"`
	} `xml:"Configurations"`
}

type GetVideoSourceConfiguration struct {
	XMLName            xml.Name `xml:"trt:GetVideoSourceConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type GetVideoSourceConfigurationOptions struct {
	XMLName            xml.Name `xml:"trt:GetVideoSourceConfigurationOptions"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
	ProfileToken       string   `xml:"trt:ProfileToken,omitempty"`
}

type GetVideoSourceConfigurationOptionsResponse struct {
	XMLName xml.Name `xml:"GetVideoSourceConfigurationOptionsResponse"`
	Options struct {
		BoundsRange *struct {
			X      *IntRange `xml:"X"`
			Y      *IntRange `xml:"Y"`
			Width  *IntRange `xml:"Width"`
			Height *IntRange `xml:"Height"`
		} `xml:"BoundsRange"`
		VideoSourceTokensAvailable []string `xml:"VideoSourceTokensAvailable"`
	} `xml:"Options"`
}

type GetVideoSourceConfigurationResponse struct {
	XMLName       xml.Name `xml:"GetVideoSourceConfigurationResponse"`
	Configuration struct {
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
	} `xml:"Configuration"`
}

type GetVideoSourceConfigurations struct {
	XMLName xml.Name `xml:"trt:GetVideoSourceConfigurations"`
	Xmlns   string   `xml:"xmlns:trt,attr"`
}

type GetVideoSourceConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetVideoSourceConfigurationsResponse"`
	Configurations []struct {
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
	} `xml:"Configurations"`
}

type GetVideoSourceModes struct {
	XMLName          xml.Name `xml:"trt:GetVideoSourceModes"`
	Xmlns            string   `xml:"xmlns:trt,attr"`
	VideoSourceToken string   `xml:"trt:VideoSourceToken"`
}

type GetVideoSourceModesResponse struct {
	XMLName          xml.Name `xml:"GetVideoSourceModesResponse"`
	VideoSourceModes []struct {
		Token      string `xml:"token,attr"`
		Enabled    bool   `xml:"Enabled"`
		Resolution struct {
			Width  int `xml:"Width"`
			Height int `xml:"Height"`
		} `xml:"Resolution"`
	} `xml:"VideoSourceModes"`
}

type GetVideoSources struct {
	XMLName xml.Name `xml:"trt:GetVideoSources"`
	Xmlns   string   `xml:"xmlns:trt,attr"`
}

type GetVideoSourcesResponse struct {
	XMLName      xml.Name `xml:"GetVideoSourcesResponse"`
	VideoSources []struct {
		Token      string  `xml:"token,attr"`
		Framerate  float64 `xml:"Framerate"`
		Resolution struct {
			Width  int `xml:"Width"`
			Height int `xml:"Height"`
		} `xml:"Resolution"`
	} `xml:"VideoSources"`
}

type RemoveVideoEncoderConfiguration struct {
	XMLName      xml.Name `xml:"trt:RemoveVideoEncoderConfiguration"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type RemoveVideoSourceConfiguration struct {
	XMLName      xml.Name `xml:"trt:RemoveVideoSourceConfiguration"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type SetVideoEncoderConfiguration struct {
	XMLName       xml.Name `xml:"trt:SetVideoEncoderConfiguration"`
	Xmlns         string   `xml:"xmlns:trt,attr"`
	Xmlnst        string   `xml:"xmlns:tt,attr"`
	Configuration struct {
		Token      string `xml:"token,attr"`
		Name       string `xml:"tt:Name"`
		UseCount   int    `xml:"tt:UseCount"`
		Encoding   string `xml:"tt:Encoding"`
		Resolution *struct {
			Width  int `xml:"tt:Width"`
			Height int `xml:"tt:Height"`
		} `xml:"tt:Resolution,omitempty"`
		Quality     *float64 `xml:"tt:Quality,omitempty"`
		RateControl *struct {
			FrameRateLimit   int `xml:"tt:FrameRateLimit"`
			EncodingInterval int `xml:"tt:EncodingInterval"`
			BitrateLimit     int `xml:"tt:BitrateLimit"`
		} `xml:"tt:RateControl,omitempty"`
	} `xml:"trt:Configuration"`
	ForcePersistence bool `xml:"trt:ForcePersistence"`
}

type SetVideoSourceConfiguration struct {
	XMLName       xml.Name `xml:"trt:SetVideoSourceConfiguration"`
	Xmlns         string   `xml:"xmlns:trt,attr"`
	Xmlnst        string   `xml:"xmlns:tt,attr"`
	Configuration struct {
		Token       string `xml:"token,attr"`
		Name        string `xml:"tt:Name"`
		UseCount    int    `xml:"tt:UseCount"`
		SourceToken string `xml:"tt:SourceToken"`
		Bounds      *struct {
			X      int `xml:"x,attr"`
			Y      int `xml:"y,attr"`
			Width  int `xml:"width,attr"`
			Height int `xml:"height,attr"`
		} `xml:"tt:Bounds,omitempty"`
	} `xml:"trt:Configuration"`
	ForcePersistence bool `xml:"trt:ForcePersistence"`
}

type SetVideoSourceMode struct {
	XMLName          xml.Name `xml:"trt:SetVideoSourceMode"`
	Xmlns            string   `xml:"xmlns:trt,attr"`
	VideoSourceToken string   `xml:"trt:VideoSourceToken"`
	ModeToken        string   `xml:"trt:ModeToken"`
}

func (s *MediaService) GetVideoEncoderConfiguration(
	ctx context.Context,
	configurationToken string,
) (*VideoEncoderConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetVideoEncoderConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetVideoEncoderConfigurationResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoEncoderConfiguration failed: %w", err)
	}

	config := &VideoEncoderConfiguration{
		Token:    resp.Configuration.Token,
		Name:     resp.Configuration.Name,
		UseCount: resp.Configuration.UseCount,
		Encoding: resp.Configuration.Encoding,
		Quality:  resp.Configuration.Quality,
	}

	if resp.Configuration.Resolution != nil {
		config.Resolution = &VideoResolution{
			Width:  resp.Configuration.Resolution.Width,
			Height: resp.Configuration.Resolution.Height,
		}
	}

	if resp.Configuration.RateControl != nil {
		config.RateControl = &VideoRateControl{
			FrameRateLimit:   resp.Configuration.RateControl.FrameRateLimit,
			EncodingInterval: resp.Configuration.RateControl.EncodingInterval,
			BitrateLimit:     resp.Configuration.RateControl.BitrateLimit,
		}
	}

	return config, nil
}

func (s *MediaService) GetVideoSources(ctx context.Context) ([]*VideoSource, error) {
	endpoint := s.getMediaEndpoint()

	req := GetVideoSources{
		Xmlns: mediaNamespace,
	}

	var resp GetVideoSourcesResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoSources failed: %w", err)
	}

	sources := make([]*VideoSource, len(resp.VideoSources))
	for i, s := range resp.VideoSources {
		sources[i] = &VideoSource{
			Token:     s.Token,
			Framerate: s.Framerate,
			Resolution: &VideoResolution{
				Width:  s.Resolution.Width,
				Height: s.Resolution.Height,
			},
		}
	}

	return sources, nil
}

func (s *MediaService) SetVideoEncoderConfiguration(
	ctx context.Context,
	config *VideoEncoderConfiguration,
	forcePersistence bool,
) error {
	endpoint := s.getMediaEndpoint()

	req := SetVideoEncoderConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           "http://www.onvif.org/ver10/schema",
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount
	req.Configuration.Encoding = config.Encoding

	if config.Resolution != nil {
		req.Configuration.Resolution = &struct {
			Width  int `xml:"tt:Width"`
			Height int `xml:"tt:Height"`
		}{
			Width:  config.Resolution.Width,
			Height: config.Resolution.Height,
		}
	}

	if config.Quality > 0 {
		req.Configuration.Quality = &config.Quality
	}

	if config.RateControl != nil {
		req.Configuration.RateControl = &struct {
			FrameRateLimit   int `xml:"tt:FrameRateLimit"`
			EncodingInterval int `xml:"tt:EncodingInterval"`
			BitrateLimit     int `xml:"tt:BitrateLimit"`
		}{
			FrameRateLimit:   config.RateControl.FrameRateLimit,
			EncodingInterval: config.RateControl.EncodingInterval,
			BitrateLimit:     config.RateControl.BitrateLimit,
		}
	}

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetVideoEncoderConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) GetVideoEncoderConfigurationOptions(
	ctx context.Context, configurationToken string,
) (*VideoEncoderConfigurationOptions, error) {
	endpoint := s.getMediaEndpoint()

	req := GetVideoEncoderConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}

	var resp GetVideoEncoderConfigurationOptionsResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoEncoderConfigurationOptions failed: %w", err)
	}

	options := &VideoEncoderConfigurationOptions{}

	if resp.Options.QualityRange != nil {
		options.QualityRange = &FloatRange{
			Min: resp.Options.QualityRange.Min,
			Max: resp.Options.QualityRange.Max,
		}
	}

	if resp.Options.JPEG != nil {
		jpegOpts := &JPEGOptions{}
		if resp.Options.JPEG.FrameRateRange != nil {
			jpegOpts.FrameRateRange = &FloatRange{
				Min: resp.Options.JPEG.FrameRateRange.Min,
				Max: resp.Options.JPEG.FrameRateRange.Max,
			}
		}
		if resp.Options.JPEG.EncodingIntervalRange != nil {
			jpegOpts.EncodingIntervalRange = &IntRange{
				Min: resp.Options.JPEG.EncodingIntervalRange.Min,
				Max: resp.Options.JPEG.EncodingIntervalRange.Max,
			}
		}
		for _, res := range resp.Options.JPEG.ResolutionsAvailable {
			jpegOpts.ResolutionsAvailable = append(jpegOpts.ResolutionsAvailable, &VideoResolution{
				Width:  res.Width,
				Height: res.Height,
			})
		}
		options.JPEG = jpegOpts
	}

	if resp.Options.H264 != nil {
		h264Opts := &H264Options{}
		if resp.Options.H264.FrameRateRange != nil {
			h264Opts.FrameRateRange = &FloatRange{
				Min: resp.Options.H264.FrameRateRange.Min,
				Max: resp.Options.H264.FrameRateRange.Max,
			}
		}
		if resp.Options.H264.GovLengthRange != nil {
			h264Opts.GovLengthRange = &IntRange{
				Min: resp.Options.H264.GovLengthRange.Min,
				Max: resp.Options.H264.GovLengthRange.Max,
			}
		}
		if resp.Options.H264.EncodingIntervalRange != nil {
			h264Opts.EncodingIntervalRange = &IntRange{
				Min: resp.Options.H264.EncodingIntervalRange.Min,
				Max: resp.Options.H264.EncodingIntervalRange.Max,
			}
		}
		for _, res := range resp.Options.H264.ResolutionsAvailable {
			h264Opts.ResolutionsAvailable = append(h264Opts.ResolutionsAvailable, &VideoResolution{
				Width:  res.Width,
				Height: res.Height,
			})
		}
		h264Opts.H264ProfilesSupported = resp.Options.H264.H264ProfilesSupported
		options.H264 = h264Opts
	}

	return options, nil
}

func (s *MediaService) GetVideoSourceModes(ctx context.Context, videoSourceToken string) ([]*VideoSourceMode, error) {
	endpoint := s.getMediaEndpoint()

	req := GetVideoSourceModes{
		Xmlns:            mediaNamespace,
		VideoSourceToken: videoSourceToken,
	}

	var resp GetVideoSourceModesResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoSourceModes failed: %w", err)
	}

	modes := make([]*VideoSourceMode, len(resp.VideoSourceModes))
	for i, m := range resp.VideoSourceModes {
		modes[i] = &VideoSourceMode{
			Token:   m.Token,
			Enabled: m.Enabled,
			Resolution: &VideoResolution{
				Width:  m.Resolution.Width,
				Height: m.Resolution.Height,
			},
		}
	}

	return modes, nil
}

func (s *MediaService) SetVideoSourceMode(ctx context.Context, videoSourceToken, modeToken string) error {
	endpoint := s.getMediaEndpoint()

	req := SetVideoSourceMode{
		Xmlns:            mediaNamespace,
		VideoSourceToken: videoSourceToken,
		ModeToken:        modeToken,
	}

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetVideoSourceMode failed: %w", err)
	}

	return nil
}

func (s *MediaService) AddVideoEncoderConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := s.getMediaEndpoint()

	req := AddVideoEncoderConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddVideoEncoderConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) RemoveVideoEncoderConfiguration(ctx context.Context, profileToken string) error {
	endpoint := s.getMediaEndpoint()

	req := RemoveVideoEncoderConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveVideoEncoderConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) AddVideoSourceConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := s.getMediaEndpoint()

	req := AddVideoSourceConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddVideoSourceConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) RemoveVideoSourceConfiguration(ctx context.Context, profileToken string) error {
	endpoint := s.getMediaEndpoint()

	req := RemoveVideoSourceConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveVideoSourceConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) GetGuaranteedNumberOfVideoEncoderInstances(
	ctx context.Context,
	configurationToken string,
) (*GuaranteedNumberOfVideoEncoderInstances, error) {
	endpoint := s.getMediaEndpoint()

	req := GetGuaranteedNumberOfVideoEncoderInstances{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetGuaranteedNumberOfVideoEncoderInstancesResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetGuaranteedNumberOfVideoEncoderInstances failed: %w", err)
	}

	return &GuaranteedNumberOfVideoEncoderInstances{
		TotalNumber: resp.TotalNumber,
		JPEG:        resp.JPEG,
		H264:        resp.H264,
		MPEG4:       resp.MPEG4,
	}, nil
}

func (s *MediaService) GetVideoSourceConfigurations(ctx context.Context) ([]*VideoSourceConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetVideoSourceConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetVideoSourceConfigurationsResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoSourceConfigurations failed: %w", err)
	}

	configs := make([]*VideoSourceConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		config := &VideoSourceConfiguration{
			Token:       cfg.Token,
			Name:        cfg.Name,
			UseCount:    cfg.UseCount,
			SourceToken: cfg.SourceToken,
		}
		if cfg.Bounds != nil {
			config.Bounds = &IntRectangle{
				X:      cfg.Bounds.X,
				Y:      cfg.Bounds.Y,
				Width:  cfg.Bounds.Width,
				Height: cfg.Bounds.Height,
			}
		}
		configs[i] = config
	}

	return configs, nil
}

func (s *MediaService) GetVideoEncoderConfigurations(ctx context.Context) ([]*VideoEncoderConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetVideoEncoderConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetVideoEncoderConfigurationsResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoEncoderConfigurations failed: %w", err)
	}

	configs := make([]*VideoEncoderConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		config := &VideoEncoderConfiguration{
			Token:    cfg.Token,
			Name:     cfg.Name,
			UseCount: cfg.UseCount,
			Encoding: cfg.Encoding,
			Quality:  cfg.Quality,
		}

		if cfg.Resolution != nil {
			config.Resolution = &VideoResolution{
				Width:  cfg.Resolution.Width,
				Height: cfg.Resolution.Height,
			}
		}

		if cfg.RateControl != nil {
			config.RateControl = &VideoRateControl{
				FrameRateLimit:   cfg.RateControl.FrameRateLimit,
				EncodingInterval: cfg.RateControl.EncodingInterval,
				BitrateLimit:     cfg.RateControl.BitrateLimit,
			}
		}

		if cfg.MPEG4 != nil {
			config.MPEG4 = &MPEG4Configuration{
				GovLength:    cfg.MPEG4.GovLength,
				MPEG4Profile: cfg.MPEG4.MPEG4Profile,
			}
		}

		if cfg.H264 != nil {
			config.H264 = &H264Configuration{
				GovLength:   cfg.H264.GovLength,
				H264Profile: cfg.H264.H264Profile,
			}
		}

		if cfg.Multicast != nil {
			config.Multicast = &MulticastConfiguration{
				Port:      cfg.Multicast.Port,
				TTL:       cfg.Multicast.TTL,
				AutoStart: cfg.Multicast.AutoStart,
			}
			if cfg.Multicast.Address != nil {
				config.Multicast.Address = &IPAddress{
					Type:        cfg.Multicast.Address.Type,
					IPv4Address: cfg.Multicast.Address.IPv4Address,
					IPv6Address: cfg.Multicast.Address.IPv6Address,
				}
			}
		}

		configs[i] = config
	}

	return configs, nil
}

func (s *MediaService) GetVideoSourceConfiguration(
	ctx context.Context,
	configurationToken string,
) (*VideoSourceConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetVideoSourceConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetVideoSourceConfigurationResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoSourceConfiguration failed: %w", err)
	}

	config := &VideoSourceConfiguration{
		Token:       resp.Configuration.Token,
		Name:        resp.Configuration.Name,
		UseCount:    resp.Configuration.UseCount,
		SourceToken: resp.Configuration.SourceToken,
	}

	if resp.Configuration.Bounds != nil {
		config.Bounds = &IntRectangle{
			X:      resp.Configuration.Bounds.X,
			Y:      resp.Configuration.Bounds.Y,
			Width:  resp.Configuration.Bounds.Width,
			Height: resp.Configuration.Bounds.Height,
		}
	}

	return config, nil
}

func (s *MediaService) GetVideoSourceConfigurationOptions(
	ctx context.Context,
	configurationToken, profileToken string,
) (*VideoSourceConfigurationOptions, error) {
	endpoint := s.getMediaEndpoint()

	req := GetVideoSourceConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}
	if profileToken != "" {
		req.ProfileToken = profileToken
	}

	var resp GetVideoSourceConfigurationOptionsResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoSourceConfigurationOptions failed: %w", err)
	}

	options := &VideoSourceConfigurationOptions{}
	if resp.Options.BoundsRange != nil {
		options.BoundsRange = &BoundsRange{
			X:      resp.Options.BoundsRange.X,
			Y:      resp.Options.BoundsRange.Y,
			Width:  resp.Options.BoundsRange.Width,
			Height: resp.Options.BoundsRange.Height,
		}
	}
	options.VideoSourceTokensAvailable = resp.Options.VideoSourceTokensAvailable

	return options, nil
}

func (s *MediaService) SetVideoSourceConfiguration(
	ctx context.Context,
	config *VideoSourceConfiguration,
	forcePersistence bool,
) error {
	endpoint := s.getMediaEndpoint()

	req := SetVideoSourceConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           "http://www.onvif.org/ver10/schema",
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount
	req.Configuration.SourceToken = config.SourceToken

	if config.Bounds != nil {
		req.Configuration.Bounds = &struct {
			X      int `xml:"x,attr"`
			Y      int `xml:"y,attr"`
			Width  int `xml:"width,attr"`
			Height int `xml:"height,attr"`
		}{
			X:      config.Bounds.X,
			Y:      config.Bounds.Y,
			Width:  config.Bounds.Width,
			Height: config.Bounds.Height,
		}
	}

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetVideoSourceConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) GetCompatibleVideoEncoderConfigurations(
	ctx context.Context,
	profileToken string,
) ([]*VideoEncoderConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetCompatibleVideoEncoderConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleVideoEncoderConfigurationsResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleVideoEncoderConfigurations failed: %w", err)
	}

	configs := make([]*VideoEncoderConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		config := &VideoEncoderConfiguration{
			Token:    cfg.Token,
			Name:     cfg.Name,
			UseCount: cfg.UseCount,
			Encoding: cfg.Encoding,
			Quality:  cfg.Quality,
		}

		if cfg.Resolution != nil {
			config.Resolution = &VideoResolution{
				Width:  cfg.Resolution.Width,
				Height: cfg.Resolution.Height,
			}
		}

		if cfg.RateControl != nil {
			config.RateControl = &VideoRateControl{
				FrameRateLimit:   cfg.RateControl.FrameRateLimit,
				EncodingInterval: cfg.RateControl.EncodingInterval,
				BitrateLimit:     cfg.RateControl.BitrateLimit,
			}
		}

		configs[i] = config
	}

	return configs, nil
}

func (s *MediaService) GetCompatibleVideoSourceConfigurations(
	ctx context.Context,
	profileToken string,
) ([]*VideoSourceConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetCompatibleVideoSourceConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleVideoSourceConfigurationsResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleVideoSourceConfigurations failed: %w", err)
	}

	configs := make([]*VideoSourceConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		config := &VideoSourceConfiguration{
			Token:       cfg.Token,
			Name:        cfg.Name,
			UseCount:    cfg.UseCount,
			SourceToken: cfg.SourceToken,
		}
		if cfg.Bounds != nil {
			config.Bounds = &IntRectangle{
				X:      cfg.Bounds.X,
				Y:      cfg.Bounds.Y,
				Width:  cfg.Bounds.Width,
				Height: cfg.Bounds.Height,
			}
		}
		configs[i] = config
	}

	return configs, nil
}
