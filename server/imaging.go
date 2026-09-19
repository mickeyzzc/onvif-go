package server

import (
	"encoding/xml"
	"fmt"

	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// Imaging service SOAP message types

// GetImagingSettingsRequest represents GetImagingSettings request.
type GetImagingSettingsRequest struct {
	XMLName          xml.Name `xml:"http://www.onvif.org/ver20/imaging/wsdl GetImagingSettings"`
	VideoSourceToken string   `xml:"VideoSourceToken"`
}

// GetImagingSettingsResponse represents GetImagingSettings response.
type GetImagingSettingsResponse struct {
	XMLName         xml.Name             `xml:"http://www.onvif.org/ver20/imaging/wsdl GetImagingSettingsResponse"`
	ImagingSettings *respImagingSettings `xml:"http://www.onvif.org/ver20/imaging/wsdl ImagingSettings"`
}

// SetImagingSettingsRequest represents SetImagingSettings request.
type SetImagingSettingsRequest struct {
	XMLName          xml.Name         `xml:"http://www.onvif.org/ver20/imaging/wsdl SetImagingSettings"`
	VideoSourceToken string           `xml:"VideoSourceToken"`
	ImagingSettings  *ImagingSettings `xml:"ImagingSettings"`
	ForcePersistence bool             `xml:"ForcePersistence,omitempty"`
}

// SetImagingSettingsResponse represents SetImagingSettings response.
type SetImagingSettingsResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/imaging/wsdl SetImagingSettingsResponse"`
}

// GetOptionsRequest represents GetOptions request.
type GetOptionsRequest struct {
	XMLName          xml.Name `xml:"http://www.onvif.org/ver20/imaging/wsdl GetOptions"`
	VideoSourceToken string   `xml:"VideoSourceToken"`
}

// GetOptionsResponse represents GetOptions response.
type GetOptionsResponse struct {
	XMLName        xml.Name            `xml:"http://www.onvif.org/ver20/imaging/wsdl GetOptionsResponse"`
	ImagingOptions *respImagingOptions `xml:"http://www.onvif.org/ver20/imaging/wsdl ImagingOptions"`
}

// MoveRequest represents Move (focus) request.
type MoveRequest struct {
	XMLName          xml.Name   `xml:"http://www.onvif.org/ver20/imaging/wsdl Move"`
	VideoSourceToken string     `xml:"VideoSourceToken"`
	Focus            *FocusMove `xml:"Focus"`
}

// MoveResponse represents Move response.
type MoveResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/imaging/wsdl MoveResponse"`
}

// Response wire shapes for imaging (#90). The provider model types keep
// unprefixed tags — they double as lenient request decode targets — so the
// response path carries dedicated schema-namespaced types.

type respFloatRange struct {
	Min float64 `xml:"http://www.onvif.org/ver10/schema Min"`
	Max float64 `xml:"http://www.onvif.org/ver10/schema Max"`
}

func toRespFloatRange(r *FloatRange) *respFloatRange {
	if r == nil {
		return nil
	}
	return &respFloatRange{Min: r.Min, Max: r.Max}
}

type respImagingSettings struct {
	BacklightCompensation *struct {
		Mode  string   `xml:"http://www.onvif.org/ver10/schema Mode"`
		Level *float64 `xml:"http://www.onvif.org/ver10/schema Level,omitempty"`
	} `xml:"http://www.onvif.org/ver10/schema BacklightCompensation,omitempty"`
	Brightness      *float64 `xml:"http://www.onvif.org/ver10/schema Brightness,omitempty"`
	ColorSaturation *float64 `xml:"http://www.onvif.org/ver10/schema ColorSaturation,omitempty"`
	Contrast        *float64 `xml:"http://www.onvif.org/ver10/schema Contrast,omitempty"`
	Exposure        *struct {
		Mode            string   `xml:"http://www.onvif.org/ver10/schema Mode"`
		Priority        *string  `xml:"http://www.onvif.org/ver10/schema Priority,omitempty"`
		MinExposureTime *float64 `xml:"http://www.onvif.org/ver10/schema MinExposureTime,omitempty"`
		MaxExposureTime *float64 `xml:"http://www.onvif.org/ver10/schema MaxExposureTime,omitempty"`
		MinGain         *float64 `xml:"http://www.onvif.org/ver10/schema MinGain,omitempty"`
		MaxGain         *float64 `xml:"http://www.onvif.org/ver10/schema MaxGain,omitempty"`
		MinIris         *float64 `xml:"http://www.onvif.org/ver10/schema MinIris,omitempty"`
		MaxIris         *float64 `xml:"http://www.onvif.org/ver10/schema MaxIris,omitempty"`
		ExposureTime    *float64 `xml:"http://www.onvif.org/ver10/schema ExposureTime,omitempty"`
		Gain            *float64 `xml:"http://www.onvif.org/ver10/schema Gain,omitempty"`
		Iris            *float64 `xml:"http://www.onvif.org/ver10/schema Iris,omitempty"`
	} `xml:"http://www.onvif.org/ver10/schema Exposure,omitempty"`
	Focus *struct {
		AutoFocusMode string   `xml:"http://www.onvif.org/ver10/schema AutoFocusMode"`
		DefaultSpeed  *float64 `xml:"http://www.onvif.org/ver10/schema DefaultSpeed,omitempty"`
		NearLimit     *float64 `xml:"http://www.onvif.org/ver10/schema NearLimit,omitempty"`
		FarLimit      *float64 `xml:"http://www.onvif.org/ver10/schema FarLimit,omitempty"`
	} `xml:"http://www.onvif.org/ver10/schema Focus,omitempty"`
	IrCutFilter      *string  `xml:"http://www.onvif.org/ver10/schema IrCutFilter,omitempty"`
	Sharpness        *float64 `xml:"http://www.onvif.org/ver10/schema Sharpness,omitempty"`
	WideDynamicRange *struct {
		Mode  string   `xml:"http://www.onvif.org/ver10/schema Mode"`
		Level *float64 `xml:"http://www.onvif.org/ver10/schema Level,omitempty"`
	} `xml:"http://www.onvif.org/ver10/schema WideDynamicRange,omitempty"`
	WhiteBalance *struct {
		Mode   string   `xml:"http://www.onvif.org/ver10/schema Mode"`
		CrGain *float64 `xml:"http://www.onvif.org/ver10/schema CrGain,omitempty"`
		CbGain *float64 `xml:"http://www.onvif.org/ver10/schema CbGain,omitempty"`
	} `xml:"http://www.onvif.org/ver10/schema WhiteBalance,omitempty"`
}

func toRespImagingSettings(in *ImagingSettings) *respImagingSettings {
	if in == nil {
		return nil
	}
	out := &respImagingSettings{
		Brightness:      in.Brightness,
		ColorSaturation: in.ColorSaturation,
		Contrast:        in.Contrast,
		IrCutFilter:     in.IrCutFilter,
		Sharpness:       in.Sharpness,
	}
	if in.BacklightCompensation != nil {
		out.BacklightCompensation = &struct {
			Mode  string   `xml:"http://www.onvif.org/ver10/schema Mode"`
			Level *float64 `xml:"http://www.onvif.org/ver10/schema Level,omitempty"`
		}{Mode: in.BacklightCompensation.Mode, Level: in.BacklightCompensation.Level}
	}
	if in.Exposure != nil {
		out.Exposure = &struct {
			Mode            string   `xml:"http://www.onvif.org/ver10/schema Mode"`
			Priority        *string  `xml:"http://www.onvif.org/ver10/schema Priority,omitempty"`
			MinExposureTime *float64 `xml:"http://www.onvif.org/ver10/schema MinExposureTime,omitempty"`
			MaxExposureTime *float64 `xml:"http://www.onvif.org/ver10/schema MaxExposureTime,omitempty"`
			MinGain         *float64 `xml:"http://www.onvif.org/ver10/schema MinGain,omitempty"`
			MaxGain         *float64 `xml:"http://www.onvif.org/ver10/schema MaxGain,omitempty"`
			MinIris         *float64 `xml:"http://www.onvif.org/ver10/schema MinIris,omitempty"`
			MaxIris         *float64 `xml:"http://www.onvif.org/ver10/schema MaxIris,omitempty"`
			ExposureTime    *float64 `xml:"http://www.onvif.org/ver10/schema ExposureTime,omitempty"`
			Gain            *float64 `xml:"http://www.onvif.org/ver10/schema Gain,omitempty"`
			Iris            *float64 `xml:"http://www.onvif.org/ver10/schema Iris,omitempty"`
		}{
			Mode:            in.Exposure.Mode,
			Priority:        in.Exposure.Priority,
			MinExposureTime: in.Exposure.MinExposureTime,
			MaxExposureTime: in.Exposure.MaxExposureTime,
			MinGain:         in.Exposure.MinGain,
			MaxGain:         in.Exposure.MaxGain,
			MinIris:         in.Exposure.MinIris,
			MaxIris:         in.Exposure.MaxIris,
			ExposureTime:    in.Exposure.ExposureTime,
			Gain:            in.Exposure.Gain,
			Iris:            in.Exposure.Iris,
		}
	}
	if in.Focus != nil {
		out.Focus = &struct {
			AutoFocusMode string   `xml:"http://www.onvif.org/ver10/schema AutoFocusMode"`
			DefaultSpeed  *float64 `xml:"http://www.onvif.org/ver10/schema DefaultSpeed,omitempty"`
			NearLimit     *float64 `xml:"http://www.onvif.org/ver10/schema NearLimit,omitempty"`
			FarLimit      *float64 `xml:"http://www.onvif.org/ver10/schema FarLimit,omitempty"`
		}{
			AutoFocusMode: in.Focus.AutoFocusMode,
			DefaultSpeed:  in.Focus.DefaultSpeed,
			NearLimit:     in.Focus.NearLimit,
			FarLimit:      in.Focus.FarLimit,
		}
	}
	if in.WideDynamicRange != nil {
		out.WideDynamicRange = &struct {
			Mode  string   `xml:"http://www.onvif.org/ver10/schema Mode"`
			Level *float64 `xml:"http://www.onvif.org/ver10/schema Level,omitempty"`
		}{Mode: in.WideDynamicRange.Mode, Level: in.WideDynamicRange.Level}
	}
	if in.WhiteBalance != nil {
		out.WhiteBalance = &struct {
			Mode   string   `xml:"http://www.onvif.org/ver10/schema Mode"`
			CrGain *float64 `xml:"http://www.onvif.org/ver10/schema CrGain,omitempty"`
			CbGain *float64 `xml:"http://www.onvif.org/ver10/schema CbGain,omitempty"`
		}{Mode: in.WhiteBalance.Mode, CrGain: in.WhiteBalance.CrGain, CbGain: in.WhiteBalance.CbGain}
	}
	return out
}

type respImagingOptions struct {
	BacklightCompensation *struct {
		Mode  []string        `xml:"http://www.onvif.org/ver10/schema Mode"`
		Level *respFloatRange `xml:"http://www.onvif.org/ver10/schema Level,omitempty"`
	} `xml:"http://www.onvif.org/ver10/schema BacklightCompensation,omitempty"`
	Brightness      *respFloatRange `xml:"http://www.onvif.org/ver10/schema Brightness,omitempty"`
	ColorSaturation *respFloatRange `xml:"http://www.onvif.org/ver10/schema ColorSaturation,omitempty"`
	Contrast        *respFloatRange `xml:"http://www.onvif.org/ver10/schema Contrast,omitempty"`
	Exposure        *struct {
		Mode            []string        `xml:"http://www.onvif.org/ver10/schema Mode"`
		Priority        []string        `xml:"http://www.onvif.org/ver10/schema Priority,omitempty"`
		MinExposureTime *respFloatRange `xml:"http://www.onvif.org/ver10/schema MinExposureTime,omitempty"`
		MaxExposureTime *respFloatRange `xml:"http://www.onvif.org/ver10/schema MaxExposureTime,omitempty"`
		MinGain         *respFloatRange `xml:"http://www.onvif.org/ver10/schema MinGain,omitempty"`
		MaxGain         *respFloatRange `xml:"http://www.onvif.org/ver10/schema MaxGain,omitempty"`
		MinIris         *respFloatRange `xml:"http://www.onvif.org/ver10/schema MinIris,omitempty"`
		MaxIris         *respFloatRange `xml:"http://www.onvif.org/ver10/schema MaxIris,omitempty"`
		ExposureTime    *respFloatRange `xml:"http://www.onvif.org/ver10/schema ExposureTime,omitempty"`
		Gain            *respFloatRange `xml:"http://www.onvif.org/ver10/schema Gain,omitempty"`
		Iris            *respFloatRange `xml:"http://www.onvif.org/ver10/schema Iris,omitempty"`
	} `xml:"http://www.onvif.org/ver10/schema Exposure,omitempty"`
	Focus *struct {
		AutoFocusModes []string        `xml:"http://www.onvif.org/ver10/schema AutoFocusModes"`
		DefaultSpeed   *respFloatRange `xml:"http://www.onvif.org/ver10/schema DefaultSpeed,omitempty"`
		NearLimit      *respFloatRange `xml:"http://www.onvif.org/ver10/schema NearLimit,omitempty"`
		FarLimit       *respFloatRange `xml:"http://www.onvif.org/ver10/schema FarLimit,omitempty"`
	} `xml:"http://www.onvif.org/ver10/schema Focus,omitempty"`
	IrCutFilterModes []string        `xml:"http://www.onvif.org/ver10/schema IrCutFilterModes,omitempty"`
	Sharpness        *respFloatRange `xml:"http://www.onvif.org/ver10/schema Sharpness,omitempty"`
	WideDynamicRange *struct {
		Mode  []string        `xml:"http://www.onvif.org/ver10/schema Mode"`
		Level *respFloatRange `xml:"http://www.onvif.org/ver10/schema Level,omitempty"`
	} `xml:"http://www.onvif.org/ver10/schema WideDynamicRange,omitempty"`
	WhiteBalance *struct {
		Mode   []string        `xml:"http://www.onvif.org/ver10/schema Mode"`
		YrGain *respFloatRange `xml:"http://www.onvif.org/ver10/schema YrGain,omitempty"`
		YbGain *respFloatRange `xml:"http://www.onvif.org/ver10/schema YbGain,omitempty"`
	} `xml:"http://www.onvif.org/ver10/schema WhiteBalance,omitempty"`
}

func toRespImagingOptions(in *ImagingOptions) *respImagingOptions {
	if in == nil {
		return nil
	}
	out := &respImagingOptions{
		Brightness:       toRespFloatRange(in.Brightness),
		ColorSaturation:  toRespFloatRange(in.ColorSaturation),
		Contrast:         toRespFloatRange(in.Contrast),
		Sharpness:        toRespFloatRange(in.Sharpness),
		IrCutFilterModes: in.IrCutFilterModes,
	}
	if in.BacklightCompensation != nil {
		out.BacklightCompensation = &struct {
			Mode  []string        `xml:"http://www.onvif.org/ver10/schema Mode"`
			Level *respFloatRange `xml:"http://www.onvif.org/ver10/schema Level,omitempty"`
		}{Mode: in.BacklightCompensation.Mode, Level: toRespFloatRange(in.BacklightCompensation.Level)}
	}
	if in.Exposure != nil {
		out.Exposure = &struct {
			Mode            []string        `xml:"http://www.onvif.org/ver10/schema Mode"`
			Priority        []string        `xml:"http://www.onvif.org/ver10/schema Priority,omitempty"`
			MinExposureTime *respFloatRange `xml:"http://www.onvif.org/ver10/schema MinExposureTime,omitempty"`
			MaxExposureTime *respFloatRange `xml:"http://www.onvif.org/ver10/schema MaxExposureTime,omitempty"`
			MinGain         *respFloatRange `xml:"http://www.onvif.org/ver10/schema MinGain,omitempty"`
			MaxGain         *respFloatRange `xml:"http://www.onvif.org/ver10/schema MaxGain,omitempty"`
			MinIris         *respFloatRange `xml:"http://www.onvif.org/ver10/schema MinIris,omitempty"`
			MaxIris         *respFloatRange `xml:"http://www.onvif.org/ver10/schema MaxIris,omitempty"`
			ExposureTime    *respFloatRange `xml:"http://www.onvif.org/ver10/schema ExposureTime,omitempty"`
			Gain            *respFloatRange `xml:"http://www.onvif.org/ver10/schema Gain,omitempty"`
			Iris            *respFloatRange `xml:"http://www.onvif.org/ver10/schema Iris,omitempty"`
		}{
			Mode:            in.Exposure.Mode,
			Priority:        in.Exposure.Priority,
			MinExposureTime: toRespFloatRange(in.Exposure.MinExposureTime),
			MaxExposureTime: toRespFloatRange(in.Exposure.MaxExposureTime),
			MinGain:         toRespFloatRange(in.Exposure.MinGain),
			MaxGain:         toRespFloatRange(in.Exposure.MaxGain),
			MinIris:         toRespFloatRange(in.Exposure.MinIris),
			MaxIris:         toRespFloatRange(in.Exposure.MaxIris),
			ExposureTime:    toRespFloatRange(in.Exposure.ExposureTime),
			Gain:            toRespFloatRange(in.Exposure.Gain),
			Iris:            toRespFloatRange(in.Exposure.Iris),
		}
	}
	if in.Focus != nil {
		out.Focus = &struct {
			AutoFocusModes []string        `xml:"http://www.onvif.org/ver10/schema AutoFocusModes"`
			DefaultSpeed   *respFloatRange `xml:"http://www.onvif.org/ver10/schema DefaultSpeed,omitempty"`
			NearLimit      *respFloatRange `xml:"http://www.onvif.org/ver10/schema NearLimit,omitempty"`
			FarLimit       *respFloatRange `xml:"http://www.onvif.org/ver10/schema FarLimit,omitempty"`
		}{
			AutoFocusModes: in.Focus.AutoFocusModes,
			DefaultSpeed:   toRespFloatRange(in.Focus.DefaultSpeed),
			NearLimit:      toRespFloatRange(in.Focus.NearLimit),
			FarLimit:       toRespFloatRange(in.Focus.FarLimit),
		}
	}
	if in.WideDynamicRange != nil {
		out.WideDynamicRange = &struct {
			Mode  []string        `xml:"http://www.onvif.org/ver10/schema Mode"`
			Level *respFloatRange `xml:"http://www.onvif.org/ver10/schema Level,omitempty"`
		}{Mode: in.WideDynamicRange.Mode, Level: toRespFloatRange(in.WideDynamicRange.Level)}
	}
	if in.WhiteBalance != nil {
		out.WhiteBalance = &struct {
			Mode   []string        `xml:"http://www.onvif.org/ver10/schema Mode"`
			YrGain *respFloatRange `xml:"http://www.onvif.org/ver10/schema YrGain,omitempty"`
			YbGain *respFloatRange `xml:"http://www.onvif.org/ver10/schema YbGain,omitempty"`
		}{Mode: in.WhiteBalance.Mode, YrGain: toRespFloatRange(in.WhiteBalance.YrGain), YbGain: toRespFloatRange(in.WhiteBalance.YbGain)}
	}
	return out
}

// Imaging service handlers - stateless translations between SOAP and
// the imaging provider.

// HandleGetImagingSettings handles GetImagingSettings request.
func (s *Server) HandleGetImagingSettings(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req GetImagingSettingsRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	settings, err := s.imaging.ImagingSettings(req.VideoSourceToken)
	if err != nil {
		return nil, err
	}
	return &GetImagingSettingsResponse{
		ImagingSettings: toRespImagingSettings(settings),
	}, nil
}

// HandleSetImagingSettings handles SetImagingSettings request.
//
//nolint:gocyclo // SetImagingSettings has high complexity due to multiple validation and update paths
func (s *Server) HandleSetImagingSettings(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req SetImagingSettingsRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if req.ImagingSettings == nil {
		// Return success if no settings to update
		return &SetImagingSettingsResponse{}, nil
	}

	if err := s.imaging.SetImagingSettings(req.VideoSourceToken, req.ImagingSettings); err != nil {
		return nil, err
	}

	return &SetImagingSettingsResponse{}, nil
}

// HandleGetOptions handles GetOptions request.
func (s *Server) HandleGetOptions(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req GetOptionsRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	options, err := s.imaging.ImagingOptions(req.VideoSourceToken)
	if err != nil {
		return nil, err
	}

	return &GetOptionsResponse{
		ImagingOptions: toRespImagingOptions(options),
	}, nil
}

// HandleMove handles Move (focus) request.
func (s *Server) HandleMove(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req MoveRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if err := s.imaging.MoveFocus(req.VideoSourceToken, req.Focus); err != nil {
		return nil, err
	}

	return &MoveResponse{}, nil
}
