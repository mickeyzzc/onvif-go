package server

import (
	"encoding/xml"
	"fmt"

	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// Media2 minimal face (issue #115): GetProfiles / GetStreamUri /
// SetSynchronizationPoint under the ver20/media (tr2) namespace — the
// Profile-T entry path. Wire rules: GetProfilesResponse/Profiles and the
// ConfigurationSet children are tr2-local elements; the configuration
// bodies (Name, Encoding, Resolution, RateControl) resolve to
// ver10/schema (the media2.wsdl ground truth, mirrored by the nsMedia2*
// decoder test).

type media2GetProfilesRequest struct {
	XMLName xml.Name `xml:"GetProfiles"`
	Token   string   `xml:"Token"`
}

type media2ConfigurationSetWire struct {
	XMLSource    *media2VideoSourceWire  `xml:"http://www.onvif.org/ver20/media/wsdl VideoSource,omitempty"`
	VideoEncoder *media2VideoEncoderWire `xml:"http://www.onvif.org/ver20/media/wsdl VideoEncoder,omitempty"`
}

type media2VideoSourceWire struct {
	Token       string `xml:"token,attr"`
	Name        string `xml:"http://www.onvif.org/ver10/schema Name"`
	UseCount    int    `xml:"http://www.onvif.org/ver10/schema UseCount"`
	SourceToken string `xml:"http://www.onvif.org/ver10/schema SourceToken"`
}

type media2VideoEncoderWire struct {
	Token      string `xml:"token,attr"`
	Name       string `xml:"http://www.onvif.org/ver10/schema Name"`
	UseCount   int    `xml:"http://www.onvif.org/ver10/schema UseCount"`
	Encoding   string `xml:"http://www.onvif.org/ver10/schema Encoding"`
	Resolution struct {
		Width  int `xml:"http://www.onvif.org/ver10/schema Width"`
		Height int `xml:"http://www.onvif.org/ver10/schema Height"`
	} `xml:"http://www.onvif.org/ver10/schema Resolution"`
	RateControl *struct {
		FrameRateLimit int `xml:"http://www.onvif.org/ver10/schema FrameRateLimit"`
		BitrateLimit   int `xml:"http://www.onvif.org/ver10/schema BitrateLimit"`
	} `xml:"http://www.onvif.org/ver10/schema RateControl,omitempty"`
}

type media2ProfileWire struct {
	Token          string                      `xml:"token,attr"`
	Fixed          bool                        `xml:"fixed,attr"`
	Name           string                      `xml:"http://www.onvif.org/ver10/schema Name"`
	Configurations *media2ConfigurationSetWire `xml:"http://www.onvif.org/ver20/media/wsdl Configurations,omitempty"`
}

type media2GetProfilesResponse struct {
	XMLName  xml.Name            `xml:"http://www.onvif.org/ver20/media/wsdl GetProfilesResponse"`
	Profiles []media2ProfileWire `xml:"http://www.onvif.org/ver20/media/wsdl Profiles"`
}

type media2GetStreamUriRequest struct {
	XMLName      xml.Name `xml:"GetStreamUri"`
	Protocol     string   `xml:"Protocol"`
	ProfileToken string   `xml:"ProfileToken"`
}

type media2GetStreamUriResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/media/wsdl GetStreamUriResponse"`
	Uri     string   `xml:"http://www.onvif.org/ver20/media/wsdl Uri"`
}

type media2SetSynchronizationPointRequest struct {
	XMLName      xml.Name `xml:"SetSynchronizationPoint"`
	ProfileToken string   `xml:"ProfileToken"`
}

// HandleMedia2GetProfiles answers the tr2 profile list derived from the
// same simulator profiles the ver10 media face serves.
func (s *Server) HandleMedia2GetProfiles(_ *soap.RequestContext, body []byte) (interface{}, error) {
	var req media2GetProfilesRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	resp := media2GetProfilesResponse{Profiles: make([]media2ProfileWire, 0, len(s.config.Profiles))}
	for _, p := range s.config.Profiles {
		if req.Token != "" && p.Token != req.Token {
			continue
		}

		// Token/name derivation mirrors the ver10 media face
		// (`<profile>_encoder`, `<profile> Encoder`).
		enc := media2VideoEncoderWire{
			Token:    p.Token + "_encoder",
			Name:     p.Name + " Encoder",
			UseCount: 1,
			Encoding: p.VideoEncoder.Encoding,
		}
		enc.Resolution.Width = p.VideoEncoder.Resolution.Width
		enc.Resolution.Height = p.VideoEncoder.Resolution.Height

		rc := struct {
			FrameRateLimit int `xml:"http://www.onvif.org/ver10/schema FrameRateLimit"`
			BitrateLimit   int `xml:"http://www.onvif.org/ver10/schema BitrateLimit"`
		}{FrameRateLimit: p.VideoEncoder.Framerate, BitrateLimit: p.VideoEncoder.Bitrate}
		enc.RateControl = &rc

		resp.Profiles = append(resp.Profiles, media2ProfileWire{
			Token: p.Token,
			Fixed: true,
			Name:  p.Name,
			Configurations: &media2ConfigurationSetWire{
				XMLSource: &media2VideoSourceWire{
					Token:       p.VideoSource.Token,
					Name:        p.VideoSource.Name,
					UseCount:    1,
					SourceToken: p.VideoSource.Token,
				},
				VideoEncoder: &enc,
			},
		})
	}

	return resp, nil
}

// HandleMedia2GetStreamUri answers the tr2 stream URI flavor (plain Uri
// element, no MediaUri wrapper).
func (s *Server) HandleMedia2GetStreamUri(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req media2GetStreamUriRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	streamInfo, err := s.stream.Stream(req.ProfileToken)
	if err != nil {
		return nil, err
	}

	return media2GetStreamUriResponse{Uri: s.deriveStreamURI(rc, streamInfo)}, nil
}

// HandleMedia2SetSynchronizationPoint acknowledges the iframe request —
// the simulator's encoder has no real GOP to puncture.
func (s *Server) HandleMedia2SetSynchronizationPoint(_ *soap.RequestContext, body []byte) (interface{}, error) {
	var req media2SetSynchronizationPointRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	if s.keyframeHook != nil {
		s.keyframeHook()
	}

	return struct {
		XMLName xml.Name `xml:"http://www.onvif.org/ver20/media/wsdl SetSynchronizationPointResponse"`
	}{}, nil
}
