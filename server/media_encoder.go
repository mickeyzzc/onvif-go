package server

import (
	"encoding/xml"
	"fmt"

	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// The ver10 media encoder-configuration family + the ver10
// SetSynchronizationPoint — the read side NVRs probe before touching a
// stream and the GOP-puncture signal for loss recovery. Wire shapes
// mirror the onvif-rs twin (media.rs) and the GetProfiles encoder
// blocks this server already emits.

type getVideoEncoderConfigurationsResponse struct {
	XMLName        xml.Name                    `xml:"http://www.onvif.org/ver10/media/wsdl GetVideoEncoderConfigurationsResponse"`
	Configurations []VideoEncoderConfiguration `xml:"http://www.onvif.org/ver10/media/wsdl Configurations"`
}

// encoderConfigurations derives one configuration per advertised
// profile — the exact blocks HandleGetProfiles writes inline (token
// `<profile>_encoder`, RateControl from the encoder fields, H264 block
// when the encoding is H264).
func (s *Server) encoderConfigurations() []VideoEncoderConfiguration {
	out := make([]VideoEncoderConfiguration, 0, len(s.config.Profiles))
	for _, p := range s.config.Profiles {
		enc := VideoEncoderConfiguration{
			Token:    p.Token + "_encoder",
			Name:     p.Name + " Encoder",
			UseCount: 1,
			Encoding: p.VideoEncoder.Encoding,
			Resolution: VideoResolution{
				Width:  p.VideoEncoder.Resolution.Width,
				Height: p.VideoEncoder.Resolution.Height,
			},
			Quality: p.VideoEncoder.Quality,
			RateControl: &VideoRateControl{
				FrameRateLimit:   p.VideoEncoder.Framerate,
				EncodingInterval: 1,
				BitrateLimit:     p.VideoEncoder.Bitrate,
			},
			SessionTimeout: "PT60S",
		}
		if p.VideoEncoder.Encoding == "H264" {
			enc.H264 = &H264Configuration{
				GovLength:   p.VideoEncoder.GovLength,
				H264Profile: "Main",
			}
		}
		out = append(out, enc)
	}
	return out
}

// HandleGetVideoEncoderConfigurations lists the per-profile encoder
// configurations.
func (s *Server) HandleGetVideoEncoderConfigurations(_ *soap.RequestContext, _ []byte) (interface{}, error) {
	return getVideoEncoderConfigurationsResponse{Configurations: s.encoderConfigurations()}, nil
}

type videoEncoderOptionsIntRange struct {
	Min int `xml:"Min"`
	Max int `xml:"Max"`
}

type videoEncoderOptionsFloatRange struct {
	Min float64 `xml:"Min"`
	Max float64 `xml:"Max"`
}

type videoEncoderOptionsResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetVideoEncoderConfigurationOptionsResponse"`
	Options struct {
		QualityRange *videoEncoderOptionsFloatRange `xml:"http://www.onvif.org/ver10/schema QualityRange,omitempty"`
		H264         *struct {
			ResolutionsAvailable  []VideoResolution              `xml:"http://www.onvif.org/ver10/schema ResolutionsAvailable"`
			GovLengthRange        *videoEncoderOptionsIntRange   `xml:"http://www.onvif.org/ver10/schema GovLengthRange,omitempty"`
			FrameRateRange        *videoEncoderOptionsFloatRange `xml:"http://www.onvif.org/ver10/schema FrameRateRange,omitempty"`
			EncodingIntervalRange *videoEncoderOptionsIntRange   `xml:"http://www.onvif.org/ver10/schema EncodingIntervalRange,omitempty"`
			H264ProfilesSupported []string                       `xml:"http://www.onvif.org/ver10/schema H264ProfilesSupported"`
		} `xml:"http://www.onvif.org/ver10/schema H264,omitempty"`
	} `xml:"http://www.onvif.org/ver10/schema Options"`
}

// HandleGetVideoEncoderConfigurationOptions answers the H264 option
// ranges derived from the advertised profiles (deduplicated
// resolutions; min/max spans across profiles; encoding interval fixed
// at 1 — the only value this server encodes).
func (s *Server) HandleGetVideoEncoderConfigurationOptions(_ *soap.RequestContext, _ []byte) (interface{}, error) {
	resp := videoEncoderOptionsResponse{}
	resp.Options.QualityRange = &videoEncoderOptionsFloatRange{Min: 1, Max: 100}

	h264 := &struct {
		ResolutionsAvailable  []VideoResolution              `xml:"http://www.onvif.org/ver10/schema ResolutionsAvailable"`
		GovLengthRange        *videoEncoderOptionsIntRange   `xml:"http://www.onvif.org/ver10/schema GovLengthRange,omitempty"`
		FrameRateRange        *videoEncoderOptionsFloatRange `xml:"http://www.onvif.org/ver10/schema FrameRateRange,omitempty"`
		EncodingIntervalRange *videoEncoderOptionsIntRange   `xml:"http://www.onvif.org/ver10/schema EncodingIntervalRange,omitempty"`
		H264ProfilesSupported []string                       `xml:"http://www.onvif.org/ver10/schema H264ProfilesSupported"`
	}{}

	seenRes := map[[2]int]bool{}
	govMin, govMax := 0, 0
	fpsMin, fpsMax := 0.0, 0.0
	for _, p := range s.config.Profiles {
		key := [2]int{p.VideoEncoder.Resolution.Width, p.VideoEncoder.Resolution.Height}
		if !seenRes[key] {
			seenRes[key] = true
			h264.ResolutionsAvailable = append(h264.ResolutionsAvailable,
				VideoResolution{Width: key[0], Height: key[1]})
		}
		fps := float64(p.VideoEncoder.Framerate)
		if fpsMin == 0 || fps < fpsMin {
			fpsMin = fps
		}
		if fps > fpsMax {
			fpsMax = fps
		}
		gov := p.VideoEncoder.GovLength
		if govMin == 0 || gov < govMin {
			govMin = gov
		}
		if gov > govMax {
			govMax = gov
		}
	}
	if len(seenRes) == 0 {
		// No H264 profile advertised — the honest answer is an options
		// block without the H264 section.
		return resp, nil
	}
	h264.GovLengthRange = &videoEncoderOptionsIntRange{Min: govMin, Max: govMax}
	h264.FrameRateRange = &videoEncoderOptionsFloatRange{Min: fpsMin, Max: fpsMax}
	h264.EncodingIntervalRange = &videoEncoderOptionsIntRange{Min: 1, Max: 1}
	h264.H264ProfilesSupported = []string{"Main"}
	resp.Options.H264 = h264

	return resp, nil
}

type setSynchronizationPointRequest struct {
	XMLName      xml.Name `xml:"SetSynchronizationPoint"`
	ProfileToken string   `xml:"ProfileToken"`
}

// HandleSetSynchronizationPoint fires the configured keyframe hook
// (the host's force-IDR seam) and acks the ver10 empty response.
func (s *Server) HandleSetSynchronizationPoint(_ *soap.RequestContext, body []byte) (interface{}, error) {
	var req setSynchronizationPointRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	if s.keyframeHook != nil {
		s.keyframeHook()
	}
	return struct {
		XMLName xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl SetSynchronizationPointResponse"`
	}{}, nil
}

type media2GetVideoEncoderConfigurationsResponse struct {
	XMLName        xml.Name                 `xml:"http://www.onvif.org/ver20/media/wsdl GetVideoEncoderConfigurationsResponse"`
	Configurations []media2VideoEncoderWire `xml:"http://www.onvif.org/ver20/media/wsdl Configurations"`
}

// HandleMedia2GetVideoEncoderConfigurations answers the tr2 encoder
// list (one tr2:Configurations block per profile — the onvif-rs twin's
// element shape).
func (s *Server) HandleMedia2GetVideoEncoderConfigurations(_ *soap.RequestContext, _ []byte) (interface{}, error) {
	resp := media2GetVideoEncoderConfigurationsResponse{
		Configurations: make([]media2VideoEncoderWire, 0, len(s.config.Profiles)),
	}
	for _, p := range s.config.Profiles {
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
		resp.Configurations = append(resp.Configurations, enc)
	}
	return resp, nil
}
