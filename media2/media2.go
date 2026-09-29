// Package media2 covers the ONVIF Media2 service (tr2, ver20/media/wsdl):
// the codec-agnostic configuration model — Encoding is a free media
// subtype name (tt:VideoEncodingMimeNames: JPEG/H264/H265 + IANA types
// like AV1), which is what makes H.265/AV1 configuration possible.
package media2

import (
	"context"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
)

// Namespace is the Media2 service WSDL namespace (tr2).
const Namespace = "http://www.onvif.org/ver20/media/wsdl"

// SchemaNamespace is the ver10 schema namespace (tt) — the namespace of
// the configuration-typed payload children.
const SchemaNamespace = "http://www.onvif.org/ver10/schema"

// Service is the Media2 service client.
type Service struct {
	c api.Caller
}

// New creates a Media2 service bound to a caller.
func New(c api.Caller) *Service { return &Service{c: c} }

// Profile is one tr2:MediaProfile. The configurations carried inline are
// modeled for video encoding (fully) and by reference (token + name) for
// the other families.
type Profile struct {
	Token        string                     `xml:"token,attr"`
	Fixed        bool                       `xml:"fixed,attr"`
	Name         string                     `xml:"Name"`
	VideoEncoder *VideoEncoderConfiguration `xml:"Configurations>VideoEncoder"`
	VideoSource  *ConfigurationRef          `xml:"Configurations>VideoSource"`
	AudioSource  *ConfigurationRef          `xml:"Configurations>AudioSource"`
	AudioEncoder *ConfigurationRef          `xml:"Configurations>AudioEncoder"`
	Analytics    *ConfigurationRef          `xml:"Configurations>Analytics"`
	PTZ          *ConfigurationRef          `xml:"Configurations>PTZ"`
	Metadata     *ConfigurationRef          `xml:"Configurations>Metadata"`
}

// ConfigurationRef references one inline configuration by token and name.
type ConfigurationRef struct {
	Token    string `xml:"token,attr"`
	Name     string `xml:"Name"`
	UseCount int    `xml:"UseCount"`
}

// VideoEncoderConfiguration is tt:VideoEncoder2Configuration: a free
// Encoding media subtype (H264/H265/AV1/…) plus resolution, rate control,
// quality, and the GOP attributes.
type VideoEncoderConfiguration struct {
	Token               string       `xml:"token,attr"`
	Name                string       `xml:"Name"`
	UseCount            int          `xml:"UseCount"`
	Encoding            string       `xml:"Encoding"`
	Resolution          Resolution   `xml:"Resolution"`
	RateControl         *RateControl `xml:"RateControl"`
	Quality             float32      `xml:"Quality"`
	GovLength           *int         `xml:"GovLength,attr"`
	AnchorFrameDistance *int         `xml:"AnchorFrameDistance,attr"`
}

// Resolution is tt:VideoResolution2.
type Resolution struct {
	Width  int `xml:"Width"`
	Height int `xml:"Height"`
}

// RateControl is tt:VideoRateControl2.
type RateControl struct {
	FrameRateLimit   float32 `xml:"FrameRateLimit"`
	BitrateLimit     int     `xml:"BitrateLimit"`
	EncodingInterval *int    `xml:"EncodingInterval"`
}

// VideoEncoderOptions is one tt:VideoEncoder2ConfigurationOptions answer
// — the device returns one per supported encoding.
type VideoEncoderOptions struct {
	Encoding       string
	QualityRange   *FloatRange
	Resolutions    []Resolution
	BitrateRange   *IntRange
	GovLengthRange []int
	FrameRateRange []float32
}

// FloatRange is tt:FloatRange.
type FloatRange struct {
	Min float32 `xml:"Min"`
	Max float32 `xml:"Max"`
}

// IntRange is tt:IntRange.
type IntRange struct {
	Min int `xml:"Min"`
	Max int `xml:"Max"`
}

// Request wire shapes: tr2-locally-declared wrappers with tt-typed
// payload children (xmlns:tt declared on roots that carry them).

type getProfilesReq struct {
	XMLName xml.Name `xml:"tr2:GetProfiles"`
	Xmlns   string   `xml:"xmlns:tr2,attr"`
	Token   string   `xml:"tr2:Token,omitempty"`
	Type    []string `xml:"tr2:Type,omitempty"`
}

type getVideoEncoderConfigurationsReq struct {
	XMLName xml.Name `xml:"tr2:GetVideoEncoderConfigurations"`
	Xmlns   string   `xml:"xmlns:tr2,attr"`
}

type getConfigurationReq struct {
	XMLName            xml.Name `xml:"tr2:GetVideoEncoderConfigurationOptions"`
	Xmlns              string   `xml:"xmlns:tr2,attr"`
	ConfigurationToken string   `xml:"tr2:ConfigurationToken,omitempty"`
	ProfileToken       string   `xml:"tr2:ProfileToken,omitempty"`
}

type setVideoEncoderConfigurationReq struct {
	XMLName       xml.Name              `xml:"tr2:SetVideoEncoderConfiguration"`
	Xmlns         string                `xml:"xmlns:tr2,attr"`
	XmlnsTT       string                `xml:"xmlns:tt,attr"`
	Configuration videoEncoderConfigOut `xml:"tr2:Configuration"`
}

type getStreamUriReq struct {
	XMLName      xml.Name `xml:"tr2:GetStreamUri"`
	Xmlns        string   `xml:"xmlns:tr2,attr"`
	Protocol     string   `xml:"tr2:Protocol"`
	ProfileToken string   `xml:"tr2:ProfileToken"`
}

// videoEncoderConfigOut is the tt:VideoEncoder2Configuration
// serialization shape (ConfigurationEntity children and encoder fields
// are schema-typed → tt:).
type videoEncoderConfigOut struct {
	Token      string `xml:"token,attr"`
	Name       string `xml:"tt:Name"`
	UseCount   int    `xml:"tt:UseCount"`
	Encoding   string `xml:"tt:Encoding"`
	Resolution struct {
		Width  int `xml:"tt:Width"`
		Height int `xml:"tt:Height"`
	} `xml:"tt:Resolution"`
	RateControl *struct {
		FrameRateLimit   float32 `xml:"tt:FrameRateLimit"`
		BitrateLimit     int     `xml:"tt:BitrateLimit"`
		EncodingInterval *int    `xml:"tt:EncodingInterval"`
	} `xml:"tt:RateControl,omitempty"`
	Quality             float32 `xml:"tt:Quality"`
	GovLength           *int    `xml:"GovLength,attr,omitempty"`
	AnchorFrameDistance *int    `xml:"AnchorFrameDistance,attr,omitempty"`
}

// GetProfiles lists the Media2 profiles. Optional token selects one;
// optional types (VideoSource/AudioSource/VideoEncoder/…, see
// tr2:ConfigurationEnumeration) restrict the returned configurations.
func (s *Service) GetProfiles(ctx context.Context, token string, types []string) ([]*Profile, error) {
	type response struct {
		XMLName  xml.Name  `xml:"GetProfilesResponse"`
		Profiles []Profile `xml:"Profiles"`
	}

	var resp response
	req := &getProfilesReq{Xmlns: Namespace, Token: token, Type: types}
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceMedia2), "", req, &resp); err != nil {
		return nil, fmt.Errorf("media2 GetProfiles failed: %w", err)
	}

	profiles := make([]*Profile, 0, len(resp.Profiles))
	for i := range resp.Profiles {
		profiles = append(profiles, &resp.Profiles[i])
	}

	return profiles, nil
}

// GetVideoEncoderConfigurations lists the video encoder configurations.
func (s *Service) GetVideoEncoderConfigurations(ctx context.Context) ([]*VideoEncoderConfiguration, error) {
	type response struct {
		XMLName        xml.Name                    `xml:"GetVideoEncoderConfigurationsResponse"`
		Configurations []VideoEncoderConfiguration `xml:"Configurations"`
	}

	var resp response
	req := &getVideoEncoderConfigurationsReq{Xmlns: Namespace}
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceMedia2), "", req, &resp); err != nil {
		return nil, fmt.Errorf("media2 GetVideoEncoderConfigurations failed: %w", err)
	}

	configs := make([]*VideoEncoderConfiguration, 0, len(resp.Configurations))
	for i := range resp.Configurations {
		configs = append(configs, &resp.Configurations[i])
	}

	return configs, nil
}

// GetVideoEncoderConfigurationOptions returns the per-encoding options —
// one entry per supported codec (the H.265/AV1 answer Media2 was made
// for: Encoding is a free media subtype name).
func (s *Service) GetVideoEncoderConfigurationOptions(ctx context.Context, configurationToken, profileToken string) ([]*VideoEncoderOptions, error) {
	type response struct {
		XMLName xml.Name                `xml:"GetVideoEncoderConfigurationOptionsResponse"`
		Options []videoEncoderOptionsIn `xml:"Options"`
	}

	var resp response
	req := &getConfigurationReq{Xmlns: Namespace, ConfigurationToken: configurationToken, ProfileToken: profileToken}
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceMedia2), "", req, &resp); err != nil {
		return nil, fmt.Errorf("media2 GetVideoEncoderConfigurationOptions failed: %w", err)
	}

	options := make([]*VideoEncoderOptions, 0, len(resp.Options))
	for i := range resp.Options {
		options = append(options, resp.Options[i].toPublic())
	}

	return options, nil
}

type videoEncoderOptionsIn struct {
	Encoding             string      `xml:"Encoding"`
	QualityRange         *FloatRange `xml:"QualityRange"`
	ResolutionsAvailable []struct {
		Width  int `xml:"Width"`
		Height int `xml:"Height"`
	} `xml:"ResolutionsAvailable"`
	BitrateRange   *IntRange `xml:"BitrateRange"`
	GovLengthRange string    `xml:"GovLengthRange,attr"`
	FrameRateRange string    `xml:"FrameRateRange,attr"`
}

func (o *videoEncoderOptionsIn) toPublic() *VideoEncoderOptions {
	out := &VideoEncoderOptions{
		Encoding:       o.Encoding,
		QualityRange:   o.QualityRange,
		BitrateRange:   o.BitrateRange,
		GovLengthRange: parseIntList(o.GovLengthRange),
		FrameRateRange: parseFloatList(o.FrameRateRange),
	}

	for _, res := range o.ResolutionsAvailable {
		out.Resolutions = append(out.Resolutions, Resolution{Width: res.Width, Height: res.Height})
	}

	return out
}

// parseIntList decodes an xs:list attribute (space-separated integers).
func parseIntList(s string) []int {
	if strings.TrimSpace(s) == "" {
		return nil
	}

	parts := strings.Fields(s)
	values := make([]int, 0, len(parts))

	for _, p := range parts {
		if v, err := strconv.Atoi(p); err == nil {
			values = append(values, v)
		}
	}

	return values
}

// parseFloatList decodes an xs:list attribute (space-separated floats).
func parseFloatList(s string) []float32 {
	if strings.TrimSpace(s) == "" {
		return nil
	}

	parts := strings.Fields(s)
	values := make([]float32, 0, len(parts))

	for _, p := range parts {
		if v, err := strconv.ParseFloat(p, 32); err == nil {
			values = append(values, float32(v))
		}
	}

	return values
}

// SetVideoEncoderConfiguration writes a video encoder configuration; the
// Encoding passes through verbatim (H264/H265/AV1/…).
func (s *Service) SetVideoEncoderConfiguration(ctx context.Context, config *VideoEncoderConfiguration) error {
	type response struct {
		XMLName xml.Name `xml:"SetVideoEncoderConfigurationResponse"`
	}

	var resp response
	req := &setVideoEncoderConfigurationReq{
		Xmlns:         Namespace,
		XmlnsTT:       SchemaNamespace,
		Configuration: toVideoEncoderOut(config),
	}
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceMedia2), "", req, &resp); err != nil {
		return fmt.Errorf("media2 SetVideoEncoderConfiguration failed: %w", err)
	}

	return nil
}

func toVideoEncoderOut(c *VideoEncoderConfiguration) videoEncoderConfigOut {
	out := videoEncoderConfigOut{
		Token:               c.Token,
		Name:                c.Name,
		UseCount:            c.UseCount,
		Encoding:            c.Encoding,
		Quality:             c.Quality,
		GovLength:           c.GovLength,
		AnchorFrameDistance: c.AnchorFrameDistance,
	}
	out.Resolution.Width = c.Resolution.Width
	out.Resolution.Height = c.Resolution.Height

	if c.RateControl != nil {
		out.RateControl = &struct {
			FrameRateLimit   float32 `xml:"tt:FrameRateLimit"`
			BitrateLimit     int     `xml:"tt:BitrateLimit"`
			EncodingInterval *int    `xml:"tt:EncodingInterval"`
		}{
			FrameRateLimit:   c.RateControl.FrameRateLimit,
			BitrateLimit:     c.RateControl.BitrateLimit,
			EncodingInterval: c.RateControl.EncodingInterval,
		}
	}

	return out
}

// GetStreamUri returns the stream URI for a profile (protocol as in
// tr2:TransportProtocol, e.g. "RtspUnicast").
func (s *Service) GetStreamUri(ctx context.Context, protocol, profileToken string) (string, error) {
	type response struct {
		XMLName xml.Name `xml:"GetStreamUriResponse"`
		Uri     string   `xml:"Uri"`
	}

	var resp response
	req := &getStreamUriReq{Xmlns: Namespace, Protocol: protocol, ProfileToken: profileToken}
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceMedia2), "", req, &resp); err != nil {
		return "", fmt.Errorf("media2 GetStreamUri failed: %w", err)
	}

	return resp.Uri, nil
}

// EncoderInstance reports how many concurrent encoder applications one
// codec supports on a video source configuration.
type EncoderInstance struct {
	Encoding string
	Number   int
}

// EncoderInstanceInfo is the GetVideoEncoderInstances answer: per-codec
// limits plus the guaranteed total.
type EncoderInstanceInfo struct {
	Codecs []EncoderInstance
	Total  int
}

// ServiceCapabilities is the Media2 capabilities answer (the flags an
// NVR probes before using snapshot / OSD / streaming features).
type ServiceCapabilities struct {
	SnapshotUri bool
	Rotation    bool
	OSD         bool
	// MaximumNumberOfProfiles comes from ProfileCapabilities.
	MaximumNumberOfProfiles int
	// RTSPStreaming comes from StreamingCapabilities.
	RTSPStreaming bool
}

// AudioEncoderConfiguration is tt:AudioEncoder2Configuration — Media2's
// codec-agnostic audio config (Encoding is an IANA media subtype such as
// audio/PCMA).
type AudioEncoderConfiguration struct {
	Token      string
	Name       string
	UseCount   int
	Encoding   string
	Bitrate    int
	SampleRate int
}

// MetadataPTZStatus selects which PTZ fields a metadata stream carries.
type MetadataPTZStatus struct {
	Status   bool
	Position bool
}

// MetadataConfiguration is tt:MetadataConfiguration (Media2 flavor).
type MetadataConfiguration struct {
	Token     string
	Name      string
	UseCount  int
	PTZStatus *MetadataPTZStatus
	Analytics bool
}

// SetSynchronizationPoint requests an iframe on the profile's stream —
// the call an NVR makes when event-triggered recording starts.
func (s *Service) SetSynchronizationPoint(ctx context.Context, profileToken string) error {
	type SetSynchronizationPoint struct {
		XMLName      xml.Name `xml:"tr2:SetSynchronizationPoint"`
		Xmlns        string   `xml:"xmlns:tr2,attr"`
		ProfileToken string   `xml:"tr2:ProfileToken"`
	}

	req := SetSynchronizationPoint{Xmlns: Namespace, ProfileToken: profileToken}
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceMedia2), "", req, nil); err != nil {
		return fmt.Errorf("media2 SetSynchronizationPoint failed: %w", err)
	}

	return nil
}

// GetVideoEncoderInstances reports the guaranteed encoder instance counts
// for a video source configuration.
func (s *Service) GetVideoEncoderInstances(ctx context.Context, configurationToken string) (*EncoderInstanceInfo, error) {
	type GetVideoEncoderInstances struct {
		XMLName            xml.Name `xml:"tr2:GetVideoEncoderInstances"`
		Xmlns              string   `xml:"xmlns:tr2,attr"`
		ConfigurationToken string   `xml:"tr2:ConfigurationToken"`
	}

	type response struct {
		XMLName xml.Name `xml:"GetVideoEncoderInstancesResponse"`
		Info    struct {
			Codec []struct {
				Encoding string `xml:"Encoding"`
				Number   int    `xml:"Number"`
			} `xml:"Codec"`
			Total int `xml:"Total"`
		} `xml:"Info"`
	}

	req := GetVideoEncoderInstances{Xmlns: Namespace, ConfigurationToken: configurationToken}
	var resp response
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceMedia2), "", req, &resp); err != nil {
		return nil, fmt.Errorf("media2 GetVideoEncoderInstances failed: %w", err)
	}

	info := &EncoderInstanceInfo{Total: resp.Info.Total}
	for _, c := range resp.Info.Codec {
		info.Codecs = append(info.Codecs, EncoderInstance{Encoding: c.Encoding, Number: c.Number})
	}

	return info, nil
}

// GetSnapshotUri returns the JPEG snapshot URI for a profile — Media2's
// flavor answers a plain Uri element (no MediaUri wrapper).
func (s *Service) GetSnapshotUri(ctx context.Context, profileToken string) (string, error) {
	type GetSnapshotUri struct {
		XMLName      xml.Name `xml:"tr2:GetSnapshotUri"`
		Xmlns        string   `xml:"xmlns:tr2,attr"`
		ProfileToken string   `xml:"tr2:ProfileToken"`
	}

	type response struct {
		XMLName xml.Name `xml:"GetSnapshotUriResponse"`
		Uri     string   `xml:"Uri"`
	}

	req := GetSnapshotUri{Xmlns: Namespace, ProfileToken: profileToken}
	var resp response
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceMedia2), "", req, &resp); err != nil {
		return "", fmt.Errorf("media2 GetSnapshotUri failed: %w", err)
	}

	return resp.Uri, nil
}

// StartMulticastStreaming starts the multicast stream of a profile.
func (s *Service) StartMulticastStreaming(ctx context.Context, profileToken string) error {
	return s.multicastStreaming(ctx, "StartMulticastStreaming", profileToken)
}

// StopMulticastStreaming stops the multicast stream of a profile.
func (s *Service) StopMulticastStreaming(ctx context.Context, profileToken string) error {
	return s.multicastStreaming(ctx, "StopMulticastStreaming", profileToken)
}

func (s *Service) multicastStreaming(ctx context.Context, action, profileToken string) error {
	// Distinct concrete types per action: a shared struct with a
	// name-swapped XMLName field would lose the tr2: prefix on the wire.
	type startMulticast struct {
		XMLName      xml.Name `xml:"tr2:StartMulticastStreaming"`
		Xmlns        string   `xml:"xmlns:tr2,attr"`
		ProfileToken string   `xml:"tr2:ProfileToken"`
	}
	type stopMulticast struct {
		XMLName      xml.Name `xml:"tr2:StopMulticastStreaming"`
		Xmlns        string   `xml:"xmlns:tr2,attr"`
		ProfileToken string   `xml:"tr2:ProfileToken"`
	}

	var req any
	switch action {
	case "StartMulticastStreaming":
		req = startMulticast{Xmlns: Namespace, ProfileToken: profileToken}
	case "StopMulticastStreaming":
		req = stopMulticast{Xmlns: Namespace, ProfileToken: profileToken}
	default:
		return fmt.Errorf("media2 %s failed: unknown multicast action", action)
	}
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceMedia2), "", req, nil); err != nil {
		return fmt.Errorf("media2 %s failed: %w", action, err)
	}

	return nil
}

// GetServiceCapabilities reports the Media2 feature flags.
func (s *Service) GetServiceCapabilities(ctx context.Context) (*ServiceCapabilities, error) {
	type GetServiceCapabilities struct {
		XMLName xml.Name `xml:"tr2:GetServiceCapabilities"`
		Xmlns   string   `xml:"xmlns:tr2,attr"`
	}

	type response struct {
		XMLName xml.Name `xml:"GetServiceCapabilitiesResponse"`
		Caps    struct {
			SnapshotUri bool `xml:"SnapshotUri,attr"`
			Rotation    bool `xml:"Rotation,attr"`
			OSD         bool `xml:"OSD,attr"`
			Profile     struct {
				MaximumNumberOfProfiles int `xml:"MaximumNumberOfProfiles,attr"`
			} `xml:"ProfileCapabilities"`
			Streaming struct {
				RTSPStreaming bool `xml:"RTSPStreaming,attr"`
			} `xml:"StreamingCapabilities"`
		} `xml:"Capabilities"`
	}

	req := GetServiceCapabilities{Xmlns: Namespace}
	var resp response
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceMedia2), "", req, &resp); err != nil {
		return nil, fmt.Errorf("media2 GetServiceCapabilities failed: %w", err)
	}

	return &ServiceCapabilities{
		SnapshotUri:             resp.Caps.SnapshotUri,
		Rotation:                resp.Caps.Rotation,
		OSD:                     resp.Caps.OSD,
		MaximumNumberOfProfiles: resp.Caps.Profile.MaximumNumberOfProfiles,
		RTSPStreaming:           resp.Caps.Streaming.RTSPStreaming,
	}, nil
}

// GetAudioEncoderConfigurations lists audio encoder configurations; a
// non-empty configurationToken filters to one configuration.
func (s *Service) GetAudioEncoderConfigurations(ctx context.Context, configurationToken string) ([]*AudioEncoderConfiguration, error) {
	type getConfiguration struct {
		XMLName            xml.Name `xml:"tr2:GetAudioEncoderConfigurations"`
		Xmlns              string   `xml:"xmlns:tr2,attr"`
		ConfigurationToken string   `xml:"tr2:ConfigurationToken,omitempty"`
	}

	type response struct {
		XMLName        xml.Name `xml:"GetAudioEncoderConfigurationsResponse"`
		Configurations []struct {
			Token      string `xml:"token,attr"`
			Name       string `xml:"Name"`
			UseCount   int    `xml:"UseCount"`
			Encoding   string `xml:"Encoding"`
			Bitrate    int    `xml:"Bitrate"`
			SampleRate int    `xml:"SampleRate"`
		} `xml:"Configurations"`
	}

	req := getConfiguration{Xmlns: Namespace, ConfigurationToken: configurationToken}
	var resp response
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceMedia2), "", req, &resp); err != nil {
		return nil, fmt.Errorf("media2 GetAudioEncoderConfigurations failed: %w", err)
	}

	out := make([]*AudioEncoderConfiguration, 0, len(resp.Configurations))
	for i := range resp.Configurations {
		c := resp.Configurations[i]
		out = append(out, &AudioEncoderConfiguration{
			Token:      c.Token,
			Name:       c.Name,
			UseCount:   c.UseCount,
			Encoding:   c.Encoding,
			Bitrate:    c.Bitrate,
			SampleRate: c.SampleRate,
		})
	}

	return out, nil
}

// GetMetadataConfigurations lists metadata configurations; a non-empty
// configurationToken filters to one configuration.
func (s *Service) GetMetadataConfigurations(ctx context.Context, configurationToken string) ([]*MetadataConfiguration, error) {
	type getConfiguration struct {
		XMLName            xml.Name `xml:"tr2:GetMetadataConfigurations"`
		Xmlns              string   `xml:"xmlns:tr2,attr"`
		ConfigurationToken string   `xml:"tr2:ConfigurationToken,omitempty"`
	}

	type response struct {
		XMLName        xml.Name `xml:"GetMetadataConfigurationsResponse"`
		Configurations []struct {
			Token     string `xml:"token,attr"`
			Name      string `xml:"Name"`
			UseCount  int    `xml:"UseCount"`
			PTZStatus *struct {
				Status   bool `xml:"Status"`
				Position bool `xml:"Position"`
			} `xml:"PTZStatus"`
			Analytics bool `xml:"Analytics"`
		} `xml:"Configurations"`
	}

	req := getConfiguration{Xmlns: Namespace, ConfigurationToken: configurationToken}
	var resp response
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceMedia2), "", req, &resp); err != nil {
		return nil, fmt.Errorf("media2 GetMetadataConfigurations failed: %w", err)
	}

	out := make([]*MetadataConfiguration, 0, len(resp.Configurations))
	for _, c := range resp.Configurations {
		cfg := &MetadataConfiguration{
			Token:     c.Token,
			Name:      c.Name,
			UseCount:  c.UseCount,
			Analytics: c.Analytics,
		}
		if c.PTZStatus != nil {
			cfg.PTZStatus = &MetadataPTZStatus{Status: c.PTZStatus.Status, Position: c.PTZStatus.Position}
		}
		out = append(out, cfg)
	}

	return out, nil
}
