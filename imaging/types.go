// Package imaging hosts the imaging-service (timg) domain types.
package imaging

import (
	"github.com/mickeyzzc/onvif-go/v2/types"
)

// ImagingSettings represents imaging settings.
type ImagingSettings struct {
	BacklightCompensation *BacklightCompensation
	Brightness            *float64
	ColorSaturation       *float64
	Contrast              *float64
	Exposure              *Exposure
	Focus                 *FocusConfiguration
	IrCutFilter           *string
	Sharpness             *float64
	WideDynamicRange      *WideDynamicRange
	WhiteBalance          *WhiteBalance
	Extension             *ImagingSettingsExtension
}

// BacklightCompensation represents backlight compensation.
type BacklightCompensation struct {
	Mode  string // OFF, ON
	Level float64
}

// Exposure represents exposure settings.
type Exposure struct {
	Mode            string // AUTO, MANUAL
	Priority        string // LowNoise, FrameRate
	MinExposureTime float64
	MaxExposureTime float64
	MinGain         float64
	MaxGain         float64
	MinIris         float64
	MaxIris         float64
	ExposureTime    float64
	Gain            float64
	Iris            float64
}

// FocusConfiguration represents focus configuration.
type FocusConfiguration struct {
	AutoFocusMode string // AUTO, MANUAL
	DefaultSpeed  float64
	NearLimit     float64
	FarLimit      float64
}

// WideDynamicRange represents WDR settings.
type WideDynamicRange struct {
	Mode  string // OFF, ON
	Level float64
}

// WhiteBalance represents white balance settings.
type WhiteBalance struct {
	Mode   string // AUTO, MANUAL
	CrGain float64
	CbGain float64
}

// ImagingSettingsExtension represents imaging settings extension.
type ImagingSettingsExtension struct{}

// ImagingOptions represents available imaging options.
type ImagingOptions struct {
	BacklightCompensation *BacklightCompensationOptions
	Brightness            *types.FloatRange
	ColorSaturation       *types.FloatRange
	Contrast              *types.FloatRange
	Exposure              *ExposureOptions
	Focus                 *FocusOptions
	IrCutFilterModes      []string
	Sharpness             *types.FloatRange
	WideDynamicRange      *WideDynamicRangeOptions
	WhiteBalance          *WhiteBalanceOptions
}

// BacklightCompensationOptions represents backlight compensation options.
type BacklightCompensationOptions struct {
	Mode  []string
	Level *types.FloatRange
}

// ExposureOptions represents exposure options.
type ExposureOptions struct {
	Mode            []string
	Priority        []string
	MinExposureTime *types.FloatRange
	MaxExposureTime *types.FloatRange
	MinGain         *types.FloatRange
	MaxGain         *types.FloatRange
	MinIris         *types.FloatRange
	MaxIris         *types.FloatRange
	ExposureTime    *types.FloatRange
	Gain            *types.FloatRange
	Iris            *types.FloatRange
}

// FocusOptions represents focus options.
type FocusOptions struct {
	AutoFocusModes []string
	DefaultSpeed   *types.FloatRange
	NearLimit      *types.FloatRange
	FarLimit       *types.FloatRange
}

// WideDynamicRangeOptions represents WDR options.
type WideDynamicRangeOptions struct {
	Mode  []string
	Level *types.FloatRange
}

// WhiteBalanceOptions represents white balance options.
type WhiteBalanceOptions struct {
	Mode   []string
	YrGain *types.FloatRange
	YbGain *types.FloatRange
}

// MoveOptions represents imaging move options.
type MoveOptions struct {
	Absolute   *AbsoluteFocusOptions
	Relative   *RelativeFocusOptions
	Continuous *ContinuousFocusOptions
}

// AbsoluteFocusOptions represents absolute focus options.
type AbsoluteFocusOptions struct {
	Position types.FloatRange
	Speed    types.FloatRange
}

// RelativeFocusOptions represents relative focus options.
type RelativeFocusOptions struct {
	Distance types.FloatRange
	Speed    types.FloatRange
}

// ContinuousFocusOptions represents continuous focus options.
type ContinuousFocusOptions struct {
	Speed types.FloatRange
}

// ImagingStatus represents imaging status.
type ImagingStatus struct {
	FocusStatus *FocusStatus
}

// FocusStatus represents focus status.
type FocusStatus struct {
	Position   float64
	MoveStatus string
	Error      string
}
