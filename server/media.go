package server

import (
	"encoding/xml"
	"fmt"
	"net/url"

	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// Media service SOAP message types

// GetProfilesResponse represents GetProfiles response.
type GetProfilesResponse struct {
	XMLName  xml.Name       `xml:"http://www.onvif.org/ver10/media/wsdl GetProfilesResponse"`
	Profiles []MediaProfile `xml:"http://www.onvif.org/ver10/media/wsdl Profiles"`
}

// MediaProfile represents a media profile.
type MediaProfile struct {
	Token                       string                       `xml:"token,attr"`
	Fixed                       bool                         `xml:"fixed,attr"`
	Name                        string                       `xml:"http://www.onvif.org/ver10/schema Name"`
	VideoSourceConfiguration    *VideoSourceConfiguration    `xml:"http://www.onvif.org/ver10/schema VideoSourceConfiguration"`
	AudioSourceConfiguration    *AudioSourceConfiguration    `xml:"http://www.onvif.org/ver10/schema AudioSourceConfiguration,omitempty"`
	VideoEncoderConfiguration   *VideoEncoderConfiguration   `xml:"http://www.onvif.org/ver10/schema VideoEncoderConfiguration"`
	AudioEncoderConfiguration   *AudioEncoderConfiguration   `xml:"http://www.onvif.org/ver10/schema AudioEncoderConfiguration,omitempty"`
	VideoAnalyticsConfiguration *VideoAnalyticsConfiguration `xml:"http://www.onvif.org/ver10/schema VideoAnalyticsConfiguration,omitempty"`
	PTZConfiguration            *PTZConfiguration            `xml:"http://www.onvif.org/ver10/schema PTZConfiguration,omitempty"`
	MetadataConfiguration       *MetadataConfiguration       `xml:"http://www.onvif.org/ver10/schema MetadataConfiguration,omitempty"`
}

// VideoSourceConfiguration represents video source configuration.
type VideoSourceConfiguration struct {
	Token       string       `xml:"token,attr"`
	Name        string       `xml:"http://www.onvif.org/ver10/schema Name"`
	UseCount    int          `xml:"http://www.onvif.org/ver10/schema UseCount"`
	SourceToken string       `xml:"http://www.onvif.org/ver10/schema SourceToken"`
	Bounds      IntRectangle `xml:"http://www.onvif.org/ver10/schema Bounds"`
}

// AudioSourceConfiguration represents audio source configuration.
type AudioSourceConfiguration struct {
	Token       string `xml:"token,attr"`
	Name        string `xml:"http://www.onvif.org/ver10/schema Name"`
	UseCount    int    `xml:"http://www.onvif.org/ver10/schema UseCount"`
	SourceToken string `xml:"http://www.onvif.org/ver10/schema SourceToken"`
}

// VideoEncoderConfiguration represents video encoder configuration.
type VideoEncoderConfiguration struct {
	Token          string                  `xml:"token,attr"`
	Name           string                  `xml:"http://www.onvif.org/ver10/schema Name"`
	UseCount       int                     `xml:"http://www.onvif.org/ver10/schema UseCount"`
	Encoding       string                  `xml:"http://www.onvif.org/ver10/schema Encoding"`
	Resolution     VideoResolution         `xml:"http://www.onvif.org/ver10/schema Resolution"`
	Quality        float64                 `xml:"http://www.onvif.org/ver10/schema Quality"`
	RateControl    *VideoRateControl       `xml:"http://www.onvif.org/ver10/schema RateControl,omitempty"`
	H264           *H264Configuration      `xml:"http://www.onvif.org/ver10/schema H264,omitempty"`
	Multicast      *MulticastConfiguration `xml:"http://www.onvif.org/ver10/schema Multicast,omitempty"`
	SessionTimeout string                  `xml:"http://www.onvif.org/ver10/schema SessionTimeout"`
}

// AudioEncoderConfiguration represents audio encoder configuration.
type AudioEncoderConfiguration struct {
	Token          string                  `xml:"token,attr"`
	Name           string                  `xml:"http://www.onvif.org/ver10/schema Name"`
	UseCount       int                     `xml:"http://www.onvif.org/ver10/schema UseCount"`
	Encoding       string                  `xml:"http://www.onvif.org/ver10/schema Encoding"`
	Bitrate        int                     `xml:"http://www.onvif.org/ver10/schema Bitrate"`
	SampleRate     int                     `xml:"http://www.onvif.org/ver10/schema SampleRate"`
	Multicast      *MulticastConfiguration `xml:"http://www.onvif.org/ver10/schema Multicast,omitempty"`
	SessionTimeout string                  `xml:"http://www.onvif.org/ver10/schema SessionTimeout"`
}

// VideoAnalyticsConfiguration represents video analytics configuration.
type VideoAnalyticsConfiguration struct {
	Token    string `xml:"token,attr"`
	Name     string `xml:"http://www.onvif.org/ver10/schema Name"`
	UseCount int    `xml:"http://www.onvif.org/ver10/schema UseCount"`
}

// PTZConfiguration represents PTZ configuration.
type PTZConfiguration struct {
	Token     string `xml:"token,attr"`
	Name      string `xml:"http://www.onvif.org/ver10/schema Name"`
	UseCount  int    `xml:"http://www.onvif.org/ver10/schema UseCount"`
	NodeToken string `xml:"http://www.onvif.org/ver10/schema NodeToken"`
}

// MetadataConfiguration represents metadata configuration.
type MetadataConfiguration struct {
	Token          string `xml:"token,attr"`
	Name           string `xml:"http://www.onvif.org/ver10/schema Name"`
	UseCount       int    `xml:"http://www.onvif.org/ver10/schema UseCount"`
	SessionTimeout string `xml:"http://www.onvif.org/ver10/schema SessionTimeout"`
}

// IntRectangle represents a rectangle with integer coordinates.
type IntRectangle struct {
	X      int `xml:"x,attr"`
	Y      int `xml:"y,attr"`
	Width  int `xml:"width,attr"`
	Height int `xml:"height,attr"`
}

// VideoResolution represents video resolution.
type VideoResolution struct {
	Width  int `xml:"http://www.onvif.org/ver10/schema Width"`
	Height int `xml:"http://www.onvif.org/ver10/schema Height"`
}

// VideoRateControl represents video rate control.
type VideoRateControl struct {
	FrameRateLimit   int `xml:"http://www.onvif.org/ver10/schema FrameRateLimit"`
	EncodingInterval int `xml:"http://www.onvif.org/ver10/schema EncodingInterval"`
	BitrateLimit     int `xml:"http://www.onvif.org/ver10/schema BitrateLimit"`
}

// H264Configuration represents H264 configuration.
type H264Configuration struct {
	GovLength   int    `xml:"http://www.onvif.org/ver10/schema GovLength"`
	H264Profile string `xml:"http://www.onvif.org/ver10/schema H264Profile"`
}

// MulticastConfiguration represents multicast configuration.
type MulticastConfiguration struct {
	Address   IPAddress `xml:"http://www.onvif.org/ver10/schema Address"`
	Port      int       `xml:"http://www.onvif.org/ver10/schema Port"`
	TTL       int       `xml:"http://www.onvif.org/ver10/schema TTL"`
	AutoStart bool      `xml:"http://www.onvif.org/ver10/schema AutoStart"`
}

// IPAddress represents an IP address.
type IPAddress struct {
	Type        string `xml:"http://www.onvif.org/ver10/schema Type"`
	IPv4Address string `xml:"http://www.onvif.org/ver10/schema IPv4Address,omitempty"`
	IPv6Address string `xml:"http://www.onvif.org/ver10/schema IPv6Address,omitempty"`
}

// GetStreamUriResponse represents GetStreamUri response.
type GetStreamUriResponse struct {
	XMLName  xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetStreamUriResponse"`
	MediaUri MediaUri `xml:"http://www.onvif.org/ver10/media/wsdl MediaUri"`
}

// MediaUri represents a media URI.
// MediaUri children are ver10/schema elements (#90; real-device capture
// in testdata/captures/getsnapshoturi_normal.xml shows tt:Uri &co.).
type MediaUri struct {
	URI                 string `xml:"http://www.onvif.org/ver10/schema Uri"`
	InvalidAfterConnect bool   `xml:"http://www.onvif.org/ver10/schema InvalidAfterConnect"`
	InvalidAfterReboot  bool   `xml:"http://www.onvif.org/ver10/schema InvalidAfterReboot"`
	Timeout             string `xml:"http://www.onvif.org/ver10/schema Timeout"`
}

// GetSnapshotUriResponse represents GetSnapshotUri response.
type GetSnapshotUriResponse struct {
	XMLName  xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetSnapshotUriResponse"`
	MediaUri MediaUri `xml:"http://www.onvif.org/ver10/media/wsdl MediaUri"`
}

// GetVideoSourcesResponse represents GetVideoSources response.
type GetVideoSourcesResponse struct {
	XMLName      xml.Name      `xml:"http://www.onvif.org/ver10/media/wsdl GetVideoSourcesResponse"`
	VideoSources []VideoSource `xml:"http://www.onvif.org/ver10/media/wsdl VideoSources"`
}

// VideoSource represents a video source.
type VideoSource struct {
	Token      string          `xml:"token,attr"`
	Framerate  float64         `xml:"http://www.onvif.org/ver10/schema Framerate"`
	Resolution VideoResolution `xml:"http://www.onvif.org/ver10/schema Resolution"`
}

// Media service handlers

// HandleGetProfiles handles GetProfiles request.
func (s *Server) HandleGetProfiles(rc *soap.RequestContext, body []byte) (interface{}, error) {
	profiles := make([]MediaProfile, len(s.config.Profiles))

	for i, profileCfg := range s.config.Profiles {
		profile := MediaProfile{
			Token: profileCfg.Token,
			Fixed: true,
			Name:  profileCfg.Name,
			VideoSourceConfiguration: &VideoSourceConfiguration{
				Token:       profileCfg.VideoSource.Token,
				Name:        profileCfg.VideoSource.Name,
				UseCount:    1,
				SourceToken: profileCfg.VideoSource.Token,
				Bounds: IntRectangle{
					X:      profileCfg.VideoSource.Bounds.X,
					Y:      profileCfg.VideoSource.Bounds.Y,
					Width:  profileCfg.VideoSource.Bounds.Width,
					Height: profileCfg.VideoSource.Bounds.Height,
				},
			},
			VideoEncoderConfiguration: &VideoEncoderConfiguration{
				Token:    profileCfg.Token + "_encoder",
				Name:     profileCfg.Name + " Encoder",
				UseCount: 1,
				Encoding: profileCfg.VideoEncoder.Encoding,
				Resolution: VideoResolution{
					Width:  profileCfg.VideoEncoder.Resolution.Width,
					Height: profileCfg.VideoEncoder.Resolution.Height,
				},
				Quality: profileCfg.VideoEncoder.Quality,
				RateControl: &VideoRateControl{
					FrameRateLimit:   profileCfg.VideoEncoder.Framerate,
					EncodingInterval: 1,
					BitrateLimit:     profileCfg.VideoEncoder.Bitrate,
				},
				SessionTimeout: "PT60S",
			},
		}

		// Add H264 configuration if encoding is H264
		if profileCfg.VideoEncoder.Encoding == "H264" {
			profile.VideoEncoderConfiguration.H264 = &H264Configuration{
				GovLength:   profileCfg.VideoEncoder.GovLength,
				H264Profile: "Main",
			}
		}

		// Add audio configuration if present
		if profileCfg.AudioSource != nil {
			profile.AudioSourceConfiguration = &AudioSourceConfiguration{
				Token:       profileCfg.AudioSource.Token,
				Name:        profileCfg.AudioSource.Name,
				UseCount:    1,
				SourceToken: profileCfg.AudioSource.Token,
			}
		}

		if profileCfg.AudioEncoder != nil {
			profile.AudioEncoderConfiguration = &AudioEncoderConfiguration{
				Token:          profileCfg.Token + "_audio_encoder",
				Name:           profileCfg.Name + " Audio Encoder",
				UseCount:       1,
				Encoding:       profileCfg.AudioEncoder.Encoding,
				Bitrate:        profileCfg.AudioEncoder.Bitrate,
				SampleRate:     profileCfg.AudioEncoder.SampleRate,
				SessionTimeout: "PT60S",
			}
		}

		// Add PTZ configuration if present
		if profileCfg.PTZ != nil {
			profile.PTZConfiguration = &PTZConfiguration{
				Token:     profileCfg.PTZ.NodeToken,
				Name:      profileCfg.Name + " PTZ",
				UseCount:  1,
				NodeToken: profileCfg.PTZ.NodeToken,
			}
		}

		profiles[i] = profile
	}

	return &GetProfilesResponse{
		Profiles: profiles,
	}, nil
}

// HandleGetStreamUri handles GetStreamUri request.
func (s *Server) HandleGetStreamUri(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req struct {
		ProfileToken string `xml:"ProfileToken"`
	}

	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Resolve the stream through the provider; a pinned override wins,
	// otherwise the URI is derived from the advertised host and the
	// stream's RTSP port (#34).
	streamInfo, err := s.stream.Stream(req.ProfileToken)
	if err != nil {
		return nil, err
	}

	return &GetStreamUriResponse{
		MediaUri: MediaUri{
			URI:                 s.deriveStreamURI(rc, streamInfo),
			InvalidAfterConnect: false,
			InvalidAfterReboot:  true,
			Timeout:             "PT60S",
		},
	}, nil
}

// HandleGetSnapshotUri handles GetSnapshotUri request.
func (s *Server) HandleGetSnapshotUri(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req struct {
		ProfileToken string `xml:"ProfileToken"`
	}

	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Find the profile
	var profileCfg *ProfileConfig
	for i := range s.config.Profiles {
		if s.config.Profiles[i].Token == req.ProfileToken {
			profileCfg = &s.config.Profiles[i]

			break
		}
	}

	if profileCfg == nil {
		return nil, fmt.Errorf("%w: %s", ErrProfileNotFound, req.ProfileToken)
	}

	if !profileCfg.Snapshot.Enabled {
		return nil, fmt.Errorf("%w: %s", ErrSnapshotNotSupported, req.ProfileToken)
	}

	// Build the snapshot URI: SnapshotPath, with the ?profile= query
	// unless the parameterless form is configured (#36).
	host := s.advertiseHost(rc)
	uri := fmt.Sprintf("http://%s:%d%s", host, s.config.Port, s.config.SnapshotPath)
	if !s.config.SnapshotURIParameterless {
		uri += "?profile=" + url.QueryEscape(req.ProfileToken)
	}

	return &GetSnapshotUriResponse{
		MediaUri: MediaUri{
			URI:                 uri,
			InvalidAfterConnect: false,
			InvalidAfterReboot:  true,
			Timeout:             "PT5S",
		},
	}, nil
}

// HandleGetVideoSources handles GetVideoSources request.
func (s *Server) HandleGetVideoSources(rc *soap.RequestContext, body []byte) (interface{}, error) {
	sources := make([]VideoSource, 0)

	// Collect unique video sources from profiles
	seenSources := make(map[string]bool)

	for _, profileCfg := range s.config.Profiles {
		if !seenSources[profileCfg.VideoSource.Token] {
			sources = append(sources, VideoSource{
				Token:     profileCfg.VideoSource.Token,
				Framerate: float64(profileCfg.VideoSource.Framerate),
				Resolution: VideoResolution{
					Width:  profileCfg.VideoSource.Resolution.Width,
					Height: profileCfg.VideoSource.Resolution.Height,
				},
			})
			seenSources[profileCfg.VideoSource.Token] = true
		}
	}

	return &GetVideoSourcesResponse{
		VideoSources: sources,
	}, nil
}

// unmarshalBody is a helper to unmarshal SOAP body content.
func unmarshalBody(body, target interface{}) error {
	var bodyXML []byte
	var err error

	// If body is already []byte, use it directly
	if b, ok := body.([]byte); ok {
		bodyXML = b
	} else {
		bodyXML, err = xml.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal XML: %w", err)
		}
	}

	if err := xml.Unmarshal(bodyXML, target); err != nil {
		return fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return nil
}
