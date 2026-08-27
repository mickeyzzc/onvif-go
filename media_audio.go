package onvif

import (
	"context"
	"encoding/xml"
	"fmt"
)

// Request/response types hoisted from method bodies.

type AddAudioDecoderConfiguration struct {
	XMLName            xml.Name `xml:"trt:AddAudioDecoderConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ProfileToken       string   `xml:"trt:ProfileToken"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type AddAudioEncoderConfiguration struct {
	XMLName            xml.Name `xml:"trt:AddAudioEncoderConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ProfileToken       string   `xml:"trt:ProfileToken"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type AddAudioOutputConfiguration struct {
	XMLName            xml.Name `xml:"trt:AddAudioOutputConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ProfileToken       string   `xml:"trt:ProfileToken"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type AddAudioSourceConfiguration struct {
	XMLName            xml.Name `xml:"trt:AddAudioSourceConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ProfileToken       string   `xml:"trt:ProfileToken"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type AddMetadataConfiguration struct {
	XMLName            xml.Name `xml:"trt:AddMetadataConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ProfileToken       string   `xml:"trt:ProfileToken"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type GetAudioDecoderConfiguration struct {
	XMLName            xml.Name `xml:"trt:GetAudioDecoderConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type GetAudioDecoderConfigurationOptions struct {
	XMLName            xml.Name `xml:"trt:GetAudioDecoderConfigurationOptions"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
}

type GetAudioDecoderConfigurationOptionsResponse struct {
	XMLName xml.Name `xml:"GetAudioDecoderConfigurationOptionsResponse"`
	Options struct {
		AACDecOptions *struct {
			BitrateList    []int `xml:"BitrateList"`
			SampleRateList []int `xml:"SampleRateList"`
		} `xml:"AACDecOptions"`
		G711DecOptions *struct {
			BitrateList []int `xml:"BitrateList"`
		} `xml:"G711DecOptions"`
		G726DecOptions *struct {
			BitrateList []int `xml:"BitrateList"`
		} `xml:"G726DecOptions"`
	} `xml:"Options"`
}

type GetAudioDecoderConfigurationResponse struct {
	XMLName       xml.Name `xml:"GetAudioDecoderConfigurationResponse"`
	Configuration struct {
		Token    string `xml:"token,attr"`
		Name     string `xml:"Name"`
		UseCount int    `xml:"UseCount"`
	} `xml:"Configuration"`
}

type GetAudioDecoderConfigurations struct {
	XMLName xml.Name `xml:"trt:GetAudioDecoderConfigurations"`
	Xmlns   string   `xml:"xmlns:trt,attr"`
}

type GetAudioDecoderConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetAudioDecoderConfigurationsResponse"`
	Configurations []struct {
		Token    string `xml:"token,attr"`
		Name     string `xml:"Name"`
		UseCount int    `xml:"UseCount"`
	} `xml:"Configurations"`
}

type GetAudioEncoderConfiguration struct {
	XMLName            xml.Name `xml:"trt:GetAudioEncoderConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type GetAudioEncoderConfigurationOptions struct {
	XMLName            xml.Name `xml:"trt:GetAudioEncoderConfigurationOptions"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
	ProfileToken       string   `xml:"trt:ProfileToken,omitempty"`
}

type GetAudioEncoderConfigurationOptionsResponse struct {
	XMLName xml.Name `xml:"GetAudioEncoderConfigurationOptionsResponse"`
	Options struct {
		EncodingOptions []string `xml:"EncodingOptions"`
		BitrateList     []int    `xml:"BitrateList"`
		SampleRateList  []int    `xml:"SampleRateList"`
	} `xml:"Options"`
}

type GetAudioEncoderConfigurationResponse struct {
	XMLName       xml.Name `xml:"GetAudioEncoderConfigurationResponse"`
	Configuration struct {
		Token      string `xml:"token,attr"`
		Name       string `xml:"Name"`
		UseCount   int    `xml:"UseCount"`
		Encoding   string `xml:"Encoding"`
		Bitrate    int    `xml:"Bitrate"`
		SampleRate int    `xml:"SampleRate"`
		Multicast  *struct {
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
	} `xml:"Configuration"`
}

type GetAudioEncoderConfigurations struct {
	XMLName xml.Name `xml:"trt:GetAudioEncoderConfigurations"`
	Xmlns   string   `xml:"xmlns:trt,attr"`
}

type GetAudioEncoderConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetAudioEncoderConfigurationsResponse"`
	Configurations []struct {
		Token      string `xml:"token,attr"`
		Name       string `xml:"Name"`
		UseCount   int    `xml:"UseCount"`
		Encoding   string `xml:"Encoding"`
		Bitrate    int    `xml:"Bitrate"`
		SampleRate int    `xml:"SampleRate"`
		Multicast  *struct {
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

type GetAudioOutputConfiguration struct {
	XMLName            xml.Name `xml:"trt:GetAudioOutputConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type GetAudioOutputConfigurationOptions struct {
	XMLName            xml.Name `xml:"trt:GetAudioOutputConfigurationOptions"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
}

type GetAudioOutputConfigurationOptionsResponse struct {
	XMLName xml.Name `xml:"GetAudioOutputConfigurationOptionsResponse"`
	Options struct {
		OutputTokensAvailable []string `xml:"OutputTokensAvailable"`
	} `xml:"Options"`
}

type GetAudioOutputConfigurationResponse struct {
	XMLName       xml.Name `xml:"GetAudioOutputConfigurationResponse"`
	Configuration struct {
		Token       string `xml:"token,attr"`
		Name        string `xml:"Name"`
		UseCount    int    `xml:"UseCount"`
		OutputToken string `xml:"OutputToken"`
	} `xml:"Configuration"`
}

type GetAudioOutputConfigurations struct {
	XMLName xml.Name `xml:"trt:GetAudioOutputConfigurations"`
	Xmlns   string   `xml:"xmlns:trt,attr"`
}

type GetAudioOutputConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetAudioOutputConfigurationsResponse"`
	Configurations []struct {
		Token       string `xml:"token,attr"`
		Name        string `xml:"Name"`
		UseCount    int    `xml:"UseCount"`
		OutputToken string `xml:"OutputToken"`
	} `xml:"Configurations"`
}

type GetAudioOutputs struct {
	XMLName xml.Name `xml:"trt:GetAudioOutputs"`
	Xmlns   string   `xml:"xmlns:trt,attr"`
}

type GetAudioOutputsResponse struct {
	XMLName      xml.Name `xml:"GetAudioOutputsResponse"`
	AudioOutputs []struct {
		Token string `xml:"token,attr"`
	} `xml:"AudioOutputs"`
}

type GetAudioSourceConfiguration struct {
	XMLName            xml.Name `xml:"trt:GetAudioSourceConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type GetAudioSourceConfigurationOptions struct {
	XMLName            xml.Name `xml:"trt:GetAudioSourceConfigurationOptions"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
	ProfileToken       string   `xml:"trt:ProfileToken,omitempty"`
}

type GetAudioSourceConfigurationOptionsResponse struct {
	XMLName xml.Name `xml:"GetAudioSourceConfigurationOptionsResponse"`
	Options struct {
		InputTokensAvailable []string `xml:"InputTokensAvailable"`
	} `xml:"Options"`
}

type GetAudioSourceConfigurationResponse struct {
	XMLName       xml.Name `xml:"GetAudioSourceConfigurationResponse"`
	Configuration struct {
		Token       string `xml:"token,attr"`
		Name        string `xml:"Name"`
		UseCount    int    `xml:"UseCount"`
		SourceToken string `xml:"SourceToken"`
	} `xml:"Configuration"`
}

type GetAudioSourceConfigurations struct {
	XMLName xml.Name `xml:"trt:GetAudioSourceConfigurations"`
	Xmlns   string   `xml:"xmlns:trt,attr"`
}

type GetAudioSourceConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetAudioSourceConfigurationsResponse"`
	Configurations []struct {
		Token       string `xml:"token,attr"`
		Name        string `xml:"Name"`
		UseCount    int    `xml:"UseCount"`
		SourceToken string `xml:"SourceToken"`
	} `xml:"Configurations"`
}

type GetAudioSources struct {
	XMLName xml.Name `xml:"trt:GetAudioSources"`
	Xmlns   string   `xml:"xmlns:trt,attr"`
}

type GetAudioSourcesResponse struct {
	XMLName      xml.Name `xml:"GetAudioSourcesResponse"`
	AudioSources []struct {
		Token    string `xml:"token,attr"`
		Channels int    `xml:"Channels"`
	} `xml:"AudioSources"`
}

type GetCompatibleAudioDecoderConfigurations struct {
	XMLName      xml.Name `xml:"trt:GetCompatibleAudioDecoderConfigurations"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type GetCompatibleAudioDecoderConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetCompatibleAudioDecoderConfigurationsResponse"`
	Configurations []struct {
		Token    string `xml:"token,attr"`
		Name     string `xml:"Name"`
		UseCount int    `xml:"UseCount"`
	} `xml:"Configurations"`
}

type GetCompatibleAudioEncoderConfigurations struct {
	XMLName      xml.Name `xml:"trt:GetCompatibleAudioEncoderConfigurations"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type GetCompatibleAudioEncoderConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetCompatibleAudioEncoderConfigurationsResponse"`
	Configurations []struct {
		Token      string `xml:"token,attr"`
		Name       string `xml:"Name"`
		UseCount   int    `xml:"UseCount"`
		Encoding   string `xml:"Encoding"`
		Bitrate    int    `xml:"Bitrate"`
		SampleRate int    `xml:"SampleRate"`
	} `xml:"Configurations"`
}

type GetCompatibleAudioOutputConfigurations struct {
	XMLName      xml.Name `xml:"trt:GetCompatibleAudioOutputConfigurations"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type GetCompatibleAudioOutputConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetCompatibleAudioOutputConfigurationsResponse"`
	Configurations []struct {
		Token       string `xml:"token,attr"`
		Name        string `xml:"Name"`
		UseCount    int    `xml:"UseCount"`
		OutputToken string `xml:"OutputToken"`
	} `xml:"Configurations"`
}

type GetCompatibleAudioSourceConfigurations struct {
	XMLName      xml.Name `xml:"trt:GetCompatibleAudioSourceConfigurations"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type GetCompatibleAudioSourceConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetCompatibleAudioSourceConfigurationsResponse"`
	Configurations []struct {
		Token       string `xml:"token,attr"`
		Name        string `xml:"Name"`
		UseCount    int    `xml:"UseCount"`
		SourceToken string `xml:"SourceToken"`
	} `xml:"Configurations"`
}

type GetCompatibleMetadataConfigurations struct {
	XMLName      xml.Name `xml:"trt:GetCompatibleMetadataConfigurations"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type GetCompatibleMetadataConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetCompatibleMetadataConfigurationsResponse"`
	Configurations []struct {
		Token     string `xml:"token,attr"`
		Name      string `xml:"Name"`
		UseCount  int    `xml:"UseCount"`
		Analytics bool   `xml:"Analytics"`
	} `xml:"Configurations"`
}

type GetMetadataConfiguration struct {
	XMLName            xml.Name `xml:"trt:GetMetadataConfiguration"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken"`
}

type GetMetadataConfigurationOptions struct {
	XMLName            xml.Name `xml:"trt:GetMetadataConfigurationOptions"`
	Xmlns              string   `xml:"xmlns:trt,attr"`
	ConfigurationToken string   `xml:"trt:ConfigurationToken,omitempty"`
	ProfileToken       string   `xml:"trt:ProfileToken,omitempty"`
}

type GetMetadataConfigurationOptionsResponse struct {
	XMLName xml.Name `xml:"GetMetadataConfigurationOptionsResponse"`
	Options struct {
		PTZStatusFilterOptions *struct {
			Status   bool `xml:"Status"`
			Position bool `xml:"Position"`
		} `xml:"PTZStatusFilterOptions"`
		Extension struct{} `xml:"Extension"`
	} `xml:"Options"`
}

type GetMetadataConfigurationResponse struct {
	XMLName       xml.Name `xml:"GetMetadataConfigurationResponse"`
	Configuration struct {
		Token     string `xml:"token,attr"`
		Name      string `xml:"Name"`
		UseCount  int    `xml:"UseCount"`
		PTZStatus *struct {
			Status   bool `xml:"Status"`
			Position bool `xml:"Position"`
		} `xml:"PTZStatus"`
		Events    *struct{} `xml:"Events"`
		Analytics bool      `xml:"Analytics"`
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
	} `xml:"Configuration"`
}

type GetMetadataConfigurations struct {
	XMLName xml.Name `xml:"trt:GetMetadataConfigurations"`
	Xmlns   string   `xml:"xmlns:trt,attr"`
}

type GetMetadataConfigurationsResponse struct {
	XMLName        xml.Name `xml:"GetMetadataConfigurationsResponse"`
	Configurations []struct {
		Token     string `xml:"token,attr"`
		Name      string `xml:"Name"`
		UseCount  int    `xml:"UseCount"`
		Analytics bool   `xml:"Analytics"`
	} `xml:"Configurations"`
}

type RemoveAudioDecoderConfiguration struct {
	XMLName      xml.Name `xml:"trt:RemoveAudioDecoderConfiguration"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type RemoveAudioEncoderConfiguration struct {
	XMLName      xml.Name `xml:"trt:RemoveAudioEncoderConfiguration"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type RemoveAudioOutputConfiguration struct {
	XMLName      xml.Name `xml:"trt:RemoveAudioOutputConfiguration"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type RemoveAudioSourceConfiguration struct {
	XMLName      xml.Name `xml:"trt:RemoveAudioSourceConfiguration"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type RemoveMetadataConfiguration struct {
	XMLName      xml.Name `xml:"trt:RemoveMetadataConfiguration"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type SetAudioDecoderConfiguration struct {
	XMLName       xml.Name `xml:"trt:SetAudioDecoderConfiguration"`
	Xmlns         string   `xml:"xmlns:trt,attr"`
	Xmlnst        string   `xml:"xmlns:tt,attr"`
	Configuration struct {
		Token    string `xml:"token,attr"`
		Name     string `xml:"tt:Name"`
		UseCount int    `xml:"tt:UseCount"`
	} `xml:"trt:Configuration"`
	ForcePersistence bool `xml:"trt:ForcePersistence"`
}

type SetAudioEncoderConfiguration struct {
	XMLName       xml.Name `xml:"trt:SetAudioEncoderConfiguration"`
	Xmlns         string   `xml:"xmlns:trt,attr"`
	Xmlnst        string   `xml:"xmlns:tt,attr"`
	Configuration struct {
		Token      string `xml:"token,attr"`
		Name       string `xml:"tt:Name"`
		UseCount   int    `xml:"tt:UseCount"`
		Encoding   string `xml:"tt:Encoding"`
		Bitrate    int    `xml:"tt:Bitrate,omitempty"`
		SampleRate int    `xml:"tt:SampleRate,omitempty"`
		Multicast  *struct {
			Address *struct {
				Type        string `xml:"tt:Type"`
				IPv4Address string `xml:"tt:IPv4Address,omitempty"`
				IPv6Address string `xml:"tt:IPv6Address,omitempty"`
			} `xml:"tt:Address,omitempty"`
			Port      int  `xml:"tt:Port,omitempty"`
			TTL       int  `xml:"tt:TTL,omitempty"`
			AutoStart bool `xml:"tt:AutoStart,omitempty"`
		} `xml:"tt:Multicast,omitempty"`
		SessionTimeout string `xml:"tt:SessionTimeout,omitempty"`
	} `xml:"trt:Configuration"`
	ForcePersistence bool `xml:"trt:ForcePersistence"`
}

type SetAudioOutputConfiguration struct {
	XMLName       xml.Name `xml:"trt:SetAudioOutputConfiguration"`
	Xmlns         string   `xml:"xmlns:trt,attr"`
	Xmlnst        string   `xml:"xmlns:tt,attr"`
	Configuration struct {
		Token       string `xml:"token,attr"`
		Name        string `xml:"tt:Name"`
		UseCount    int    `xml:"tt:UseCount"`
		OutputToken string `xml:"tt:OutputToken"`
	} `xml:"trt:Configuration"`
	ForcePersistence bool `xml:"trt:ForcePersistence"`
}

type SetAudioSourceConfiguration struct {
	XMLName       xml.Name `xml:"trt:SetAudioSourceConfiguration"`
	Xmlns         string   `xml:"xmlns:trt,attr"`
	Xmlnst        string   `xml:"xmlns:tt,attr"`
	Configuration struct {
		Token       string `xml:"token,attr"`
		Name        string `xml:"tt:Name"`
		UseCount    int    `xml:"tt:UseCount"`
		SourceToken string `xml:"tt:SourceToken"`
	} `xml:"trt:Configuration"`
	ForcePersistence bool `xml:"trt:ForcePersistence"`
}

type SetMetadataConfiguration struct {
	XMLName       xml.Name `xml:"trt:SetMetadataConfiguration"`
	Xmlns         string   `xml:"xmlns:trt,attr"`
	Xmlnst        string   `xml:"xmlns:tt,attr"`
	Configuration struct {
		Token     string `xml:"token,attr"`
		Name      string `xml:"tt:Name"`
		UseCount  int    `xml:"tt:UseCount"`
		PTZStatus *struct {
			Status   bool `xml:"tt:Status"`
			Position bool `xml:"tt:Position"`
		} `xml:"tt:PTZStatus,omitempty"`
		Events    *struct{} `xml:"tt:Events,omitempty"`
		Analytics bool      `xml:"tt:Analytics,omitempty"`
		Multicast *struct {
			Address *struct {
				Type        string `xml:"tt:Type"`
				IPv4Address string `xml:"tt:IPv4Address,omitempty"`
				IPv6Address string `xml:"tt:IPv6Address,omitempty"`
			} `xml:"tt:Address,omitempty"`
			Port      int  `xml:"tt:Port,omitempty"`
			TTL       int  `xml:"tt:TTL,omitempty"`
			AutoStart bool `xml:"tt:AutoStart,omitempty"`
		} `xml:"tt:Multicast,omitempty"`
		SessionTimeout string `xml:"tt:SessionTimeout,omitempty"`
	} `xml:"trt:Configuration"`
	ForcePersistence bool `xml:"trt:ForcePersistence"`
}

func (s *MediaService) GetAudioSources(ctx context.Context) ([]*AudioSource, error) {
	endpoint := s.getMediaEndpoint()

	req := GetAudioSources{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioSourcesResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioSources failed: %w", err)
	}

	sources := make([]*AudioSource, len(resp.AudioSources))
	for i, s := range resp.AudioSources {
		sources[i] = &AudioSource{
			Token:    s.Token,
			Channels: s.Channels,
		}
	}

	return sources, nil
}

func (s *MediaService) GetAudioOutputs(ctx context.Context) ([]*AudioOutput, error) {
	endpoint := s.getMediaEndpoint()

	req := GetAudioOutputs{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioOutputsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioOutputs failed: %w", err)
	}

	outputs := make([]*AudioOutput, len(resp.AudioOutputs))
	for i, o := range resp.AudioOutputs {
		outputs[i] = &AudioOutput{
			Token: o.Token,
		}
	}

	return outputs, nil
}

func (s *MediaService) GetAudioEncoderConfiguration(
	ctx context.Context,
	configurationToken string,
) (*AudioEncoderConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetAudioEncoderConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetAudioEncoderConfigurationResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioEncoderConfiguration failed: %w", err)
	}

	config := &AudioEncoderConfiguration{
		Token:      resp.Configuration.Token,
		Name:       resp.Configuration.Name,
		UseCount:   resp.Configuration.UseCount,
		Encoding:   resp.Configuration.Encoding,
		Bitrate:    resp.Configuration.Bitrate,
		SampleRate: resp.Configuration.SampleRate,
	}

	if resp.Configuration.Multicast != nil {
		config.Multicast = &MulticastConfiguration{
			Port:      resp.Configuration.Multicast.Port,
			TTL:       resp.Configuration.Multicast.TTL,
			AutoStart: resp.Configuration.Multicast.AutoStart,
		}
		if resp.Configuration.Multicast.Address != nil {
			config.Multicast.Address = &IPAddress{
				Type:        resp.Configuration.Multicast.Address.Type,
				IPv4Address: resp.Configuration.Multicast.Address.IPv4Address,
				IPv6Address: resp.Configuration.Multicast.Address.IPv6Address,
			}
		}
	}

	return config, nil
}

func (s *MediaService) SetAudioEncoderConfiguration(
	ctx context.Context,
	config *AudioEncoderConfiguration,
	forcePersistence bool,
) error {
	endpoint := s.getMediaEndpoint()

	req := SetAudioEncoderConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           "http://www.onvif.org/ver10/schema",
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount
	req.Configuration.Encoding = config.Encoding
	if config.Bitrate > 0 {
		req.Configuration.Bitrate = config.Bitrate
	}
	if config.SampleRate > 0 {
		req.Configuration.SampleRate = config.SampleRate
	}

	if config.Multicast != nil {
		req.Configuration.Multicast = &struct {
			Address *struct {
				Type        string `xml:"tt:Type"`
				IPv4Address string `xml:"tt:IPv4Address,omitempty"`
				IPv6Address string `xml:"tt:IPv6Address,omitempty"`
			} `xml:"tt:Address,omitempty"`
			Port      int  `xml:"tt:Port,omitempty"`
			TTL       int  `xml:"tt:TTL,omitempty"`
			AutoStart bool `xml:"tt:AutoStart,omitempty"`
		}{
			Port:      config.Multicast.Port,
			TTL:       config.Multicast.TTL,
			AutoStart: config.Multicast.AutoStart,
		}
		if config.Multicast.Address != nil {
			req.Configuration.Multicast.Address = &struct {
				Type        string `xml:"tt:Type"`
				IPv4Address string `xml:"tt:IPv4Address,omitempty"`
				IPv6Address string `xml:"tt:IPv6Address,omitempty"`
			}{
				Type:        config.Multicast.Address.Type,
				IPv4Address: config.Multicast.Address.IPv4Address,
				IPv6Address: config.Multicast.Address.IPv6Address,
			}
		}
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetAudioEncoderConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) GetMetadataConfiguration(
	ctx context.Context,
	configurationToken string,
) (*MetadataConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetMetadataConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetMetadataConfigurationResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetMetadataConfiguration failed: %w", err)
	}

	config := &MetadataConfiguration{
		Token:     resp.Configuration.Token,
		Name:      resp.Configuration.Name,
		UseCount:  resp.Configuration.UseCount,
		Analytics: resp.Configuration.Analytics,
	}

	if resp.Configuration.PTZStatus != nil {
		config.PTZStatus = &PTZFilter{
			Status:   resp.Configuration.PTZStatus.Status,
			Position: resp.Configuration.PTZStatus.Position,
		}
	}

	if resp.Configuration.Events != nil {
		config.Events = &EventSubscription{}
	}

	if resp.Configuration.Multicast != nil {
		config.Multicast = &MulticastConfiguration{
			Port:      resp.Configuration.Multicast.Port,
			TTL:       resp.Configuration.Multicast.TTL,
			AutoStart: resp.Configuration.Multicast.AutoStart,
		}
		if resp.Configuration.Multicast.Address != nil {
			config.Multicast.Address = &IPAddress{
				Type:        resp.Configuration.Multicast.Address.Type,
				IPv4Address: resp.Configuration.Multicast.Address.IPv4Address,
				IPv6Address: resp.Configuration.Multicast.Address.IPv6Address,
			}
		}
	}

	return config, nil
}

func (s *MediaService) SetMetadataConfiguration(
	ctx context.Context,
	config *MetadataConfiguration,
	forcePersistence bool,
) error {
	endpoint := s.getMediaEndpoint()

	req := SetMetadataConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           "http://www.onvif.org/ver10/schema",
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount
	req.Configuration.Analytics = config.Analytics

	if config.PTZStatus != nil {
		req.Configuration.PTZStatus = &struct {
			Status   bool `xml:"tt:Status"`
			Position bool `xml:"tt:Position"`
		}{
			Status:   config.PTZStatus.Status,
			Position: config.PTZStatus.Position,
		}
	}

	if config.Events != nil {
		req.Configuration.Events = &struct{}{}
	}

	if config.Multicast != nil {
		req.Configuration.Multicast = &struct {
			Address *struct {
				Type        string `xml:"tt:Type"`
				IPv4Address string `xml:"tt:IPv4Address,omitempty"`
				IPv6Address string `xml:"tt:IPv6Address,omitempty"`
			} `xml:"tt:Address,omitempty"`
			Port      int  `xml:"tt:Port,omitempty"`
			TTL       int  `xml:"tt:TTL,omitempty"`
			AutoStart bool `xml:"tt:AutoStart,omitempty"`
		}{
			Port:      config.Multicast.Port,
			TTL:       config.Multicast.TTL,
			AutoStart: config.Multicast.AutoStart,
		}
		if config.Multicast.Address != nil {
			req.Configuration.Multicast.Address = &struct {
				Type        string `xml:"tt:Type"`
				IPv4Address string `xml:"tt:IPv4Address,omitempty"`
				IPv6Address string `xml:"tt:IPv6Address,omitempty"`
			}{
				Type:        config.Multicast.Address.Type,
				IPv4Address: config.Multicast.Address.IPv4Address,
				IPv6Address: config.Multicast.Address.IPv6Address,
			}
		}
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetMetadataConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) AddAudioEncoderConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := s.getMediaEndpoint()

	req := AddAudioEncoderConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddAudioEncoderConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) RemoveAudioEncoderConfiguration(ctx context.Context, profileToken string) error {
	endpoint := s.getMediaEndpoint()

	req := RemoveAudioEncoderConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveAudioEncoderConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) AddAudioSourceConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := s.getMediaEndpoint()

	req := AddAudioSourceConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddAudioSourceConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) RemoveAudioSourceConfiguration(ctx context.Context, profileToken string) error {
	endpoint := s.getMediaEndpoint()

	req := RemoveAudioSourceConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveAudioSourceConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) AddMetadataConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := s.getMediaEndpoint()

	req := AddMetadataConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddMetadataConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) RemoveMetadataConfiguration(ctx context.Context, profileToken string) error {
	endpoint := s.getMediaEndpoint()

	req := RemoveMetadataConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveMetadataConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) GetAudioEncoderConfigurationOptions(
	ctx context.Context,
	configurationToken, profileToken string,
) (*AudioEncoderConfigurationOptions, error) {
	endpoint := s.getMediaEndpoint()

	req := GetAudioEncoderConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}
	if profileToken != "" {
		req.ProfileToken = profileToken
	}

	var resp GetAudioEncoderConfigurationOptionsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioEncoderConfigurationOptions failed: %w", err)
	}

	return &AudioEncoderConfigurationOptions{
		EncodingOptions: resp.Options.EncodingOptions,
		BitrateList:     resp.Options.BitrateList,
		SampleRateList:  resp.Options.SampleRateList,
	}, nil
}

func (s *MediaService) GetMetadataConfigurationOptions(
	ctx context.Context,
	configurationToken, profileToken string,
) (*MetadataConfigurationOptions, error) {
	endpoint := s.getMediaEndpoint()

	req := GetMetadataConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}
	if profileToken != "" {
		req.ProfileToken = profileToken
	}

	var resp GetMetadataConfigurationOptionsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetMetadataConfigurationOptions failed: %w", err)
	}

	options := &MetadataConfigurationOptions{}
	if resp.Options.PTZStatusFilterOptions != nil {
		options.PTZStatusFilterOptions = &PTZFilter{
			Status:   resp.Options.PTZStatusFilterOptions.Status,
			Position: resp.Options.PTZStatusFilterOptions.Position,
		}
	}

	return options, nil
}

func (s *MediaService) GetAudioOutputConfiguration(ctx context.Context, configurationToken string) (*AudioOutputConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetAudioOutputConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetAudioOutputConfigurationResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioOutputConfiguration failed: %w", err)
	}

	return &AudioOutputConfiguration{
		Token:       resp.Configuration.Token,
		Name:        resp.Configuration.Name,
		UseCount:    resp.Configuration.UseCount,
		OutputToken: resp.Configuration.OutputToken,
	}, nil
}

func (s *MediaService) SetAudioOutputConfiguration(ctx context.Context, config *AudioOutputConfiguration, forcePersistence bool) error {
	endpoint := s.getMediaEndpoint()

	req := SetAudioOutputConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           "http://www.onvif.org/ver10/schema",
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount
	req.Configuration.OutputToken = config.OutputToken

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetAudioOutputConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) GetAudioOutputConfigurationOptions(
	ctx context.Context,
	configurationToken string,
) (*AudioOutputConfigurationOptions, error) {
	endpoint := s.getMediaEndpoint()

	req := GetAudioOutputConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}

	var resp GetAudioOutputConfigurationOptionsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioOutputConfigurationOptions failed: %w", err)
	}

	return &AudioOutputConfigurationOptions{
		OutputTokensAvailable: resp.Options.OutputTokensAvailable,
	}, nil
}

func (s *MediaService) GetAudioDecoderConfigurationOptions(
	ctx context.Context,
	configurationToken string,
) (*AudioDecoderConfigurationOptions, error) {
	endpoint := s.getMediaEndpoint()

	req := GetAudioDecoderConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}

	var resp GetAudioDecoderConfigurationOptionsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioDecoderConfigurationOptions failed: %w", err)
	}

	options := &AudioDecoderConfigurationOptions{}
	if resp.Options.AACDecOptions != nil {
		options.AACDecOptions = &AudioDecoderOptions{
			BitrateList:    resp.Options.AACDecOptions.BitrateList,
			SampleRateList: resp.Options.AACDecOptions.SampleRateList,
		}
	}
	if resp.Options.G711DecOptions != nil {
		options.G711DecOptions = &AudioDecoderOptions{
			BitrateList: resp.Options.G711DecOptions.BitrateList,
		}
	}
	if resp.Options.G726DecOptions != nil {
		options.G726DecOptions = &AudioDecoderOptions{
			BitrateList: resp.Options.G726DecOptions.BitrateList,
		}
	}

	return options, nil
}

func (s *MediaService) GetAudioSourceConfigurations(ctx context.Context) ([]*AudioSourceConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetAudioSourceConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioSourceConfigurationsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioSourceConfigurations failed: %w", err)
	}

	configs := make([]*AudioSourceConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioSourceConfiguration{
			Token:       cfg.Token,
			Name:        cfg.Name,
			UseCount:    cfg.UseCount,
			SourceToken: cfg.SourceToken,
		}
	}

	return configs, nil
}

func (s *MediaService) GetAudioEncoderConfigurations(ctx context.Context) ([]*AudioEncoderConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetAudioEncoderConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioEncoderConfigurationsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioEncoderConfigurations failed: %w", err)
	}

	configs := make([]*AudioEncoderConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		config := &AudioEncoderConfiguration{
			Token:      cfg.Token,
			Name:       cfg.Name,
			UseCount:   cfg.UseCount,
			Encoding:   cfg.Encoding,
			Bitrate:    cfg.Bitrate,
			SampleRate: cfg.SampleRate,
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

func (s *MediaService) GetAudioSourceConfiguration(ctx context.Context, configurationToken string) (*AudioSourceConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetAudioSourceConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetAudioSourceConfigurationResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioSourceConfiguration failed: %w", err)
	}

	return &AudioSourceConfiguration{
		Token:       resp.Configuration.Token,
		Name:        resp.Configuration.Name,
		UseCount:    resp.Configuration.UseCount,
		SourceToken: resp.Configuration.SourceToken,
	}, nil
}

func (s *MediaService) GetAudioSourceConfigurationOptions(
	ctx context.Context,
	configurationToken, profileToken string,
) (*AudioSourceConfigurationOptions, error) {
	endpoint := s.getMediaEndpoint()

	req := GetAudioSourceConfigurationOptions{
		Xmlns: mediaNamespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}
	if profileToken != "" {
		req.ProfileToken = profileToken
	}

	var resp GetAudioSourceConfigurationOptionsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioSourceConfigurationOptions failed: %w", err)
	}

	return &AudioSourceConfigurationOptions{
		InputTokensAvailable: resp.Options.InputTokensAvailable,
	}, nil
}

func (s *MediaService) SetAudioSourceConfiguration(ctx context.Context, config *AudioSourceConfiguration, forcePersistence bool) error {
	endpoint := s.getMediaEndpoint()

	req := SetAudioSourceConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           "http://www.onvif.org/ver10/schema",
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount
	req.Configuration.SourceToken = config.SourceToken

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetAudioSourceConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) GetCompatibleAudioEncoderConfigurations(
	ctx context.Context,
	profileToken string,
) ([]*AudioEncoderConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetCompatibleAudioEncoderConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleAudioEncoderConfigurationsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleAudioEncoderConfigurations failed: %w", err)
	}

	configs := make([]*AudioEncoderConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioEncoderConfiguration{
			Token:      cfg.Token,
			Name:       cfg.Name,
			UseCount:   cfg.UseCount,
			Encoding:   cfg.Encoding,
			Bitrate:    cfg.Bitrate,
			SampleRate: cfg.SampleRate,
		}
	}

	return configs, nil
}

func (s *MediaService) GetCompatibleAudioSourceConfigurations(ctx context.Context, profileToken string) ([]*AudioSourceConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetCompatibleAudioSourceConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleAudioSourceConfigurationsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleAudioSourceConfigurations failed: %w", err)
	}

	configs := make([]*AudioSourceConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioSourceConfiguration{
			Token:       cfg.Token,
			Name:        cfg.Name,
			UseCount:    cfg.UseCount,
			SourceToken: cfg.SourceToken,
		}
	}

	return configs, nil
}

func (s *MediaService) GetCompatibleMetadataConfigurations(ctx context.Context, profileToken string) ([]*MetadataConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetCompatibleMetadataConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleMetadataConfigurationsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleMetadataConfigurations failed: %w", err)
	}

	configs := make([]*MetadataConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &MetadataConfiguration{
			Token:     cfg.Token,
			Name:      cfg.Name,
			UseCount:  cfg.UseCount,
			Analytics: cfg.Analytics,
		}
	}

	return configs, nil
}

func (s *MediaService) GetCompatibleAudioOutputConfigurations(ctx context.Context, profileToken string) ([]*AudioOutputConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetCompatibleAudioOutputConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleAudioOutputConfigurationsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleAudioOutputConfigurations failed: %w", err)
	}

	configs := make([]*AudioOutputConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioOutputConfiguration{
			Token:       cfg.Token,
			Name:        cfg.Name,
			UseCount:    cfg.UseCount,
			OutputToken: cfg.OutputToken,
		}
	}

	return configs, nil
}

func (s *MediaService) GetCompatibleAudioDecoderConfigurations(ctx context.Context, profileToken string) ([]*AudioDecoderConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetCompatibleAudioDecoderConfigurations{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleAudioDecoderConfigurationsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatibleAudioDecoderConfigurations failed: %w", err)
	}

	configs := make([]*AudioDecoderConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioDecoderConfiguration{
			Token:    cfg.Token,
			Name:     cfg.Name,
			UseCount: cfg.UseCount,
		}
	}

	return configs, nil
}

func (s *MediaService) GetMetadataConfigurations(ctx context.Context) ([]*MetadataConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetMetadataConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetMetadataConfigurationsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetMetadataConfigurations failed: %w", err)
	}

	configs := make([]*MetadataConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &MetadataConfiguration{
			Token:     cfg.Token,
			Name:      cfg.Name,
			UseCount:  cfg.UseCount,
			Analytics: cfg.Analytics,
		}
	}

	return configs, nil
}

func (s *MediaService) GetAudioOutputConfigurations(ctx context.Context) ([]*AudioOutputConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetAudioOutputConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioOutputConfigurationsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioOutputConfigurations failed: %w", err)
	}

	configs := make([]*AudioOutputConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioOutputConfiguration{
			Token:       cfg.Token,
			Name:        cfg.Name,
			UseCount:    cfg.UseCount,
			OutputToken: cfg.OutputToken,
		}
	}

	return configs, nil
}

func (s *MediaService) GetAudioDecoderConfigurations(ctx context.Context) ([]*AudioDecoderConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetAudioDecoderConfigurations{
		Xmlns: mediaNamespace,
	}

	var resp GetAudioDecoderConfigurationsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioDecoderConfigurations failed: %w", err)
	}

	configs := make([]*AudioDecoderConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &AudioDecoderConfiguration{
			Token:    cfg.Token,
			Name:     cfg.Name,
			UseCount: cfg.UseCount,
		}
	}

	return configs, nil
}

func (s *MediaService) GetAudioDecoderConfiguration(
	ctx context.Context,
	configurationToken string,
) (*AudioDecoderConfiguration, error) {
	endpoint := s.getMediaEndpoint()

	req := GetAudioDecoderConfiguration{
		Xmlns:              mediaNamespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetAudioDecoderConfigurationResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAudioDecoderConfiguration failed: %w", err)
	}

	return &AudioDecoderConfiguration{
		Token:    resp.Configuration.Token,
		Name:     resp.Configuration.Name,
		UseCount: resp.Configuration.UseCount,
	}, nil
}

func (s *MediaService) SetAudioDecoderConfiguration(ctx context.Context, config *AudioDecoderConfiguration, forcePersistence bool) error {
	endpoint := s.getMediaEndpoint()

	req := SetAudioDecoderConfiguration{
		Xmlns:            mediaNamespace,
		Xmlnst:           "http://www.onvif.org/ver10/schema",
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetAudioDecoderConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) AddAudioOutputConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := s.getMediaEndpoint()

	req := AddAudioOutputConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddAudioOutputConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) RemoveAudioOutputConfiguration(ctx context.Context, profileToken string) error {
	endpoint := s.getMediaEndpoint()

	req := RemoveAudioOutputConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveAudioOutputConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) AddAudioDecoderConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := s.getMediaEndpoint()

	req := AddAudioDecoderConfiguration{
		Xmlns:              mediaNamespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddAudioDecoderConfiguration failed: %w", err)
	}

	return nil
}

func (s *MediaService) RemoveAudioDecoderConfiguration(ctx context.Context, profileToken string) error {
	endpoint := s.getMediaEndpoint()

	req := RemoveAudioDecoderConfiguration{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveAudioDecoderConfiguration failed: %w", err)
	}

	return nil
}
