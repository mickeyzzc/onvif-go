// Media profiles: listing, parsing, and main/sub stream selection.

package media

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/ptz"
	"github.com/mickeyzzc/onvif-go/v2/types"
)

const Namespace = "http://www.onvif.org/ver10/media/wsdl"

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
	XMLName xml.Name `xml:"GetServiceCapabilitiesResponse"`

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

func (s *Service) GetProfiles(ctx context.Context) ([]*Profile, error) {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := GetProfiles{
		Xmlns: Namespace,
	}

	var resp GetProfilesResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
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
				profile.VideoSourceConfiguration.Bounds = &types.IntRectangle{
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
			profile.PTZConfiguration = &ptz.PTZConfiguration{
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

func (s *Service) CreateProfile(ctx context.Context, name, token string) (*Profile, error) {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := CreateProfile{
		Xmlns: Namespace,
		Name:  name,
	}
	if token != "" {
		req.Token = &token
	}

	var resp CreateProfileResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("CreateProfile failed: %w", err)
	}

	return &Profile{
		Token: resp.Profile.Token,
		Name:  resp.Profile.Name,
	}, nil
}

func (s *Service) DeleteProfile(ctx context.Context, profileToken string) error {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := DeleteProfile{
		Xmlns:        Namespace,
		ProfileToken: profileToken,
	}

	if err := s.c.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("DeleteProfile failed: %w", err)
	}

	return nil
}

func (s *Service) GetMediaServiceCapabilities(ctx context.Context) (*MediaServiceCapabilities, error) {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := GetServiceCapabilities{
		Xmlns: Namespace,
	}

	var resp GetServiceCapabilitiesResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
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

func (s *Service) GetProfile(ctx context.Context, profileToken string) (*Profile, error) {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := GetProfile{
		Xmlns:        Namespace,
		ProfileToken: profileToken,
	}

	var resp GetProfileResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetProfile failed: %w", err)
	}

	return &Profile{
		Token: resp.Profile.Token,
		Name:  resp.Profile.Name,
	}, nil
}

func (s *Service) SetProfile(ctx context.Context, profile *Profile) error {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := SetProfile{
		Xmlns:  Namespace,
		Xmlnst: "http://www.onvif.org/ver10/schema",
	}
	req.Profile.Token = profile.Token
	req.Profile.Name = profile.Name

	if err := s.c.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetProfile failed: %w", err)
	}

	return nil
}

func (s *Service) AddPTZConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := AddPTZConfiguration{
		Xmlns:              Namespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	if err := s.c.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddPTZConfiguration failed: %w", err)
	}

	return nil
}

func (s *Service) RemovePTZConfiguration(ctx context.Context, profileToken string) error {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := RemovePTZConfiguration{
		Xmlns:        Namespace,
		ProfileToken: profileToken,
	}

	if err := s.c.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemovePTZConfiguration failed: %w", err)
	}

	return nil
}

func (s *Service) GetCompatiblePTZConfigurations(ctx context.Context, profileToken string) ([]*ptz.PTZConfiguration, error) {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := GetCompatiblePTZConfigurations{
		Xmlns:        Namespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatiblePTZConfigurationsResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetCompatiblePTZConfigurations failed: %w", err)
	}

	configs := make([]*ptz.PTZConfiguration, len(resp.Configurations))
	for i, cfg := range resp.Configurations {
		configs[i] = &ptz.PTZConfiguration{
			Token:     cfg.Token,
			Name:      cfg.Name,
			UseCount:  cfg.UseCount,
			NodeToken: cfg.NodeToken,
		}
	}

	return configs, nil
}

func (s *Service) GetVideoAnalyticsConfigurations(ctx context.Context) ([]*VideoAnalyticsConfiguration, error) {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := GetVideoAnalyticsConfigurations{
		Xmlns: Namespace,
	}

	var resp GetVideoAnalyticsConfigurationsResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
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

func (s *Service) GetVideoAnalyticsConfiguration(
	ctx context.Context,
	configurationToken string,
) (*VideoAnalyticsConfiguration, error) {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := GetVideoAnalyticsConfiguration{
		Xmlns:              Namespace,
		ConfigurationToken: configurationToken,
	}

	var resp GetVideoAnalyticsConfigurationResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoAnalyticsConfiguration failed: %w", err)
	}

	return &VideoAnalyticsConfiguration{
		Token:    resp.Configuration.Token,
		Name:     resp.Configuration.Name,
		UseCount: resp.Configuration.UseCount,
	}, nil
}

func (s *Service) GetCompatibleVideoAnalyticsConfigurations(ctx context.Context, profileToken string) ([]*VideoAnalyticsConfiguration, error) {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := GetCompatibleVideoAnalyticsConfigurations{
		Xmlns:        Namespace,
		ProfileToken: profileToken,
	}

	var resp GetCompatibleVideoAnalyticsConfigurationsResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
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

func (s *Service) SetVideoAnalyticsConfiguration(ctx context.Context, config *VideoAnalyticsConfiguration, forcePersistence bool) error {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := SetVideoAnalyticsConfiguration{
		Xmlns:            Namespace,
		Xmlnst:           "http://www.onvif.org/ver10/schema",
		ForcePersistence: forcePersistence,
	}

	req.Configuration.Token = config.Token
	req.Configuration.Name = config.Name
	req.Configuration.UseCount = config.UseCount

	if err := s.c.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetVideoAnalyticsConfiguration failed: %w", err)
	}

	return nil
}

func (s *Service) GetVideoAnalyticsConfigurationOptions(
	ctx context.Context,
	configurationToken, profileToken string,
) (*VideoAnalyticsConfigurationOptions, error) {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := GetVideoAnalyticsConfigurationOptions{
		Xmlns: Namespace,
	}
	if configurationToken != "" {
		req.ConfigurationToken = configurationToken
	}
	if profileToken != "" {
		req.ProfileToken = profileToken
	}

	var resp GetVideoAnalyticsConfigurationOptionsResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoAnalyticsConfigurationOptions failed: %w", err)
	}

	return &VideoAnalyticsConfigurationOptions{}, nil
}

func (s *Service) AddVideoAnalyticsConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := AddVideoAnalyticsConfiguration{
		Xmlns:              Namespace,
		ProfileToken:       profileToken,
		ConfigurationToken: configurationToken,
	}

	if err := s.c.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AddVideoAnalyticsConfiguration failed: %w", err)
	}

	return nil
}

func (s *Service) RemoveVideoAnalyticsConfiguration(ctx context.Context, profileToken string) error {
	endpoint := s.c.EndpointFor(api.ServiceMedia)

	req := RemoveVideoAnalyticsConfiguration{
		Xmlns:        Namespace,
		ProfileToken: profileToken,
	}

	if err := s.c.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemoveVideoAnalyticsConfiguration failed: %w", err)
	}

	return nil
}

// Naming hints used ONLY as a tie-breaker between profiles of equal
// resolution. OEM firmware assigns arbitrary profile names, so naming can
// never be the primary signal (issue #7); these hints settle exact ties and
// actively demote obviously-secondary profiles.
var (
	mainStreamHints = []string{"main", "primary", "主流", "主码流", "channel1"}
	subStreamHints  = []string{"sub", "secondary", "辅流", "辅码流", "extra"}
)

// SelectMainProfile returns the token of the main-stream profile: the one
// with the most pixels (W×H). Blindly using profiles[0] silently records a
// substream on devices that list the low-resolution profile first.
//
// Heuristics (field-tested against Hikvision, Dahua, Amcrest, HiSilicon OEM
// and ESP32 hardware):
//   - Resolution rules. Ties are settled by naming hints only: a candidate
//     whose token/name contains a main hint ("main", "primary", "主流",
//     "主码流", "channel1") wins; candidates carrying sub hints ("sub",
//     "secondary", "辅流", "辅码流", "extra") lose ties.
//   - When no profile carries resolution information at all, the first
//     profile is returned (list-order fallback).
//
// Returns "" for an empty list.
func SelectMainProfile(profiles []*Profile) string {
	if len(profiles) == 0 {
		return ""
	}

	best := -1
	bestPixels := int64(-1)
	anyKnown := false

	for i, p := range profiles {
		pixels := profilePixels(p)
		if pixels < 0 {
			continue
		}

		anyKnown = true
		switch {
		case pixels > bestPixels:
			best, bestPixels = i, pixels
		case pixels == bestPixels:
			if mainTieBreak(profiles[i], profiles[best]) < 0 {
				best = i
			}
		}
	}

	if !anyKnown {
		return profiles[0].Token
	}

	return profiles[best].Token
}

// SelectSubProfile returns the token of the sub-stream profile given the
// already-selected main token: among the remaining profiles it is the one
// with the most pixels that is still STRICTLY smaller than the main
// resolution. A second profile with the SAME resolution as main is not a
// substream — on some hardware (e.g. Amcrest IP4M) two tokens at the same
// resolution are two handles onto the same stream.
//
// Returns "" when there is no independent substream. If mainToken does not
// resolve to any profile, the same-resolution guard is disabled and the
// largest profile wins — the main stream itself is returned (pass a token
// from SelectMainProfile to avoid this). If nothing carries resolution
// information, the first non-main profile is returned as a list-order
// fallback.
func SelectSubProfile(profiles []*Profile, mainToken string) string {
	if len(profiles) == 0 {
		return ""
	}

	mainPixels := int64(-1)
	anyKnown := false
	for _, p := range profiles {
		if pixels := profilePixels(p); pixels >= 0 {
			anyKnown = true
			if p != nil && p.Token == mainToken {
				mainPixels = pixels
			}
		}
	}

	best := ""
	bestPixels := int64(-1)

	for _, p := range profiles {
		if p == nil || p.Token == mainToken || p.Token == "" {
			continue
		}

		pixels := profilePixels(p)
		if pixels < 0 {
			continue
		}

		// Same resolution as main = same stream under another token.
		if mainPixels >= 0 && pixels >= mainPixels {
			continue
		}

		switch {
		case pixels > bestPixels:
			best, bestPixels = p.Token, pixels
		case pixels == bestPixels:
			if best != "" && subTieBreak(p, profileByToken(profiles, best)) < 0 {
				best = p.Token
			}
		}
	}

	if best != "" {
		return best
	}

	// List-order fallback when resolution information is absent.
	if !anyKnown {
		for _, p := range profiles {
			if p != nil && p.Token != "" && p.Token != mainToken {
				return p.Token
			}
		}
	}

	return ""
}

// profilePixels returns W×H of the profile's video encoder configuration, or
// -1 when the profile or its resolution is unknown.
func profilePixels(p *Profile) int64 {
	if p == nil || p.VideoEncoderConfiguration == nil || p.VideoEncoderConfiguration.Resolution == nil {
		return -1
	}

	w := int64(p.VideoEncoderConfiguration.Resolution.Width)
	h := int64(p.VideoEncoderConfiguration.Resolution.Height)
	if w <= 0 || h <= 0 {
		return -1
	}

	return w * h
}

// mainTieBreak compares two same-resolution profiles for main-stream
// preference. Returns <0 when a wins (main-ish naming beats neutral beats
// sub-ish).
func mainTieBreak(a, b *Profile) int {
	return hintScore(b) - hintScore(a)
}

// subTieBreak compares two same-resolution profiles for sub-stream
// preference. Returns <0 when a wins (sub-ish naming beats neutral).
func subTieBreak(a, b *Profile) int {
	return hintScore(a) - hintScore(b)
}

// hintScore rates a profile's naming hints: positive = main-ish, negative =
// sub-ish, zero = neutral.
func hintScore(p *Profile) int {
	if p == nil {
		return 0
	}

	name := strings.ToLower(p.Token + " " + p.Name)

	for _, hint := range mainStreamHints {
		if strings.Contains(name, strings.ToLower(hint)) {
			return 1
		}
	}

	for _, hint := range subStreamHints {
		if strings.Contains(name, strings.ToLower(hint)) {
			return -1
		}
	}

	return 0
}

// profileByToken finds the profile carrying the given token (nil when absent).
func profileByToken(profiles []*Profile, token string) *Profile {
	for _, p := range profiles {
		if p != nil && p.Token == token {
			return p
		}
	}

	return nil
}
