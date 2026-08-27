// Package provider defines the pluggable state sources behind the ONVIF
// server: device identity, stream URIs, snapshots, imaging, and PTZ.
// The simulator in server/simulator is the reference implementation; a
// real camera supplies its own (hardware-backed) implementations and
// keeps the SOAP layer unchanged.
package provider

import "time"

// DeviceInfo contains device identification information.
type DeviceInfo struct {
	Manufacturer    string
	Model           string
	FirmwareVersion string
	SerialNumber    string
	HardwareID      string
}

// ProfileConfig represents a camera profile configuration.
type ProfileConfig struct {
	Token        string              // Profile token (unique identifier)
	Name         string              // Profile name
	VideoSource  VideoSourceConfig   // Video source configuration
	AudioSource  *AudioSourceConfig  // Audio source configuration (optional)
	VideoEncoder VideoEncoderConfig  // Video encoder configuration
	AudioEncoder *AudioEncoderConfig // Audio encoder configuration (optional)
	PTZ          *PTZConfig          // PTZ configuration (optional)
	Snapshot     SnapshotConfig      // Snapshot configuration
}

// VideoSourceConfig represents video source configuration.
type VideoSourceConfig struct {
	Token      string // Video source token
	Name       string // Video source name
	Resolution Resolution
	Framerate  int
	Bounds     Bounds
}

// AudioSourceConfig represents audio source configuration.
type AudioSourceConfig struct {
	Token      string // Audio source token
	Name       string // Audio source name
	SampleRate int    // Sample rate in Hz (e.g., 8000, 16000, 48000)
	Bitrate    int    // Bitrate in kbps
}

// VideoEncoderConfig represents video encoder configuration.
type VideoEncoderConfig struct {
	Encoding   string     // JPEG, H264, H265, MPEG4
	Resolution Resolution // Video resolution
	Quality    float64    // Quality (0-100)
	Framerate  int        // Frames per second
	Bitrate    int        // Bitrate in kbps
	GovLength  int        // GOP length
}

// AudioEncoderConfig represents audio encoder configuration.
type AudioEncoderConfig struct {
	Encoding   string // G711, G726, AAC
	Bitrate    int    // Bitrate in kbps
	SampleRate int    // Sample rate in Hz
}

// PTZConfig represents PTZ configuration.
type PTZConfig struct {
	NodeToken          string   // PTZ node token
	PanRange           Range    // Pan range in degrees
	TiltRange          Range    // Tilt range in degrees
	ZoomRange          Range    // Zoom range
	DefaultSpeed       PTZSpeed // Default speed
	SupportsContinuous bool     // Supports continuous move
	SupportsAbsolute   bool     // Supports absolute move
	SupportsRelative   bool     // Supports relative move
	Presets            []Preset // Predefined presets
}

// SnapshotConfig represents snapshot configuration.
type SnapshotConfig struct {
	Enabled    bool       // Whether snapshots are supported
	Resolution Resolution // Snapshot resolution
	Quality    float64    // JPEG quality (0-100)
}

// Resolution represents video resolution.
type Resolution struct {
	Width  int
	Height int
}

// Bounds represents video bounds.
type Bounds struct {
	X      int
	Y      int
	Width  int
	Height int
}

// Range represents a numeric range.
type Range struct {
	Min float64
	Max float64
}

// PTZSpeed represents PTZ movement speed.
type PTZSpeed struct {
	Pan  float64 // Pan speed (-1.0 to 1.0)
	Tilt float64 // Tilt speed (-1.0 to 1.0)
	Zoom float64 // Zoom speed (-1.0 to 1.0)
}

// Preset represents a PTZ preset position.
type Preset struct {
	Token    string      // Preset token
	Name     string      // Preset name
	Position PTZPosition // Position
}

// PTZPosition represents PTZ position.
type PTZPosition struct {
	Pan  float64 // Pan position
	Tilt float64 // Tilt position
	Zoom float64 // Zoom position
}

// StreamConfig represents an RTSP stream configuration.
type StreamConfig struct {
	ProfileToken string // Associated profile token
	RTSPPath     string // RTSP path (e.g., "/stream1")
	StreamURI    string // Full RTSP URI
}

// PTZState represents the current PTZ state.
type PTZState struct {
	Position   PTZPosition
	Moving     bool
	PanMoving  bool
	TiltMoving bool
	ZoomMoving bool
	LastUpdate time.Time
}

// ImagingState represents the current imaging settings state.
type ImagingState struct {
	Brightness       float64
	Contrast         float64
	Saturation       float64
	Sharpness        float64
	BacklightComp    BacklightCompensation
	Exposure         ExposureSettings
	Focus            FocusSettings
	WhiteBalance     WhiteBalanceSettings
	WideDynamicRange WDRSettings
	IrCutFilter      string // ON, OFF, AUTO
}

// BacklightCompensation represents backlight compensation settings.
type BacklightCompensation struct {
	Mode  string  // OFF, ON
	Level float64 // 0-100
}

// ExposureSettings represents exposure settings.
type ExposureSettings struct {
	Mode         string // AUTO, MANUAL
	Priority     string // LowNoise, FrameRate
	MinExposure  float64
	MaxExposure  float64
	MinGain      float64
	MaxGain      float64
	ExposureTime float64
	Gain         float64
}

// FocusSettings represents focus settings.
type FocusSettings struct {
	AutoFocusMode string // AUTO, MANUAL
	DefaultSpeed  float64
	NearLimit     float64
	FarLimit      float64
	CurrentPos    float64
}

// WhiteBalanceSettings represents white balance settings.
type WhiteBalanceSettings struct {
	Mode   string // AUTO, MANUAL
	CrGain float64
	CbGain float64
}

// WDRSettings represents wide dynamic range settings.
type WDRSettings struct {
	Mode  string  // OFF, ON
	Level float64 // 0-100
}

// ImagingSettings represents imaging settings.
type ImagingSettings struct {
	BacklightCompensation *BacklightCompensationSettings `xml:"BacklightCompensation,omitempty"`
	Brightness            *float64                       `xml:"Brightness,omitempty"`
	ColorSaturation       *float64                       `xml:"ColorSaturation,omitempty"`
	Contrast              *float64                       `xml:"Contrast,omitempty"`
	Exposure              *ExposureSettings20            `xml:"Exposure,omitempty"`
	Focus                 *FocusConfiguration20          `xml:"Focus,omitempty"`
	IrCutFilter           *string                        `xml:"IrCutFilter,omitempty"`
	Sharpness             *float64                       `xml:"Sharpness,omitempty"`
	WideDynamicRange      *WideDynamicRangeSettings      `xml:"WideDynamicRange,omitempty"`
	WhiteBalance          *WhiteBalanceSettings20        `xml:"WhiteBalance,omitempty"`
}

// BacklightCompensationSettings represents backlight compensation settings.
type BacklightCompensationSettings struct {
	Mode  string   `xml:"Mode"`
	Level *float64 `xml:"Level,omitempty"`
}

// ExposureSettings20 represents exposure settings for ONVIF 2.0.
type ExposureSettings20 struct {
	Mode            string     `xml:"Mode"`
	Priority        *string    `xml:"Priority,omitempty"`
	Window          *Rectangle `xml:"Window,omitempty"`
	MinExposureTime *float64   `xml:"MinExposureTime,omitempty"`
	MaxExposureTime *float64   `xml:"MaxExposureTime,omitempty"`
	MinGain         *float64   `xml:"MinGain,omitempty"`
	MaxGain         *float64   `xml:"MaxGain,omitempty"`
	MinIris         *float64   `xml:"MinIris,omitempty"`
	MaxIris         *float64   `xml:"MaxIris,omitempty"`
	ExposureTime    *float64   `xml:"ExposureTime,omitempty"`
	Gain            *float64   `xml:"Gain,omitempty"`
	Iris            *float64   `xml:"Iris,omitempty"`
}

// FocusConfiguration20 represents focus configuration for ONVIF 2.0.
type FocusConfiguration20 struct {
	AutoFocusMode string   `xml:"AutoFocusMode"`
	DefaultSpeed  *float64 `xml:"DefaultSpeed,omitempty"`
	NearLimit     *float64 `xml:"NearLimit,omitempty"`
	FarLimit      *float64 `xml:"FarLimit,omitempty"`
}

// WideDynamicRangeSettings represents WDR settings.
type WideDynamicRangeSettings struct {
	Mode  string   `xml:"Mode"`
	Level *float64 `xml:"Level,omitempty"`
}

// WhiteBalanceSettings20 represents white balance settings for ONVIF 2.0.
type WhiteBalanceSettings20 struct {
	Mode   string   `xml:"Mode"`
	CrGain *float64 `xml:"CrGain,omitempty"`
	CbGain *float64 `xml:"CbGain,omitempty"`
}

// Rectangle represents a rectangle.
type Rectangle struct {
	Bottom float64 `xml:"bottom,attr"`
	Top    float64 `xml:"top,attr"`
	Right  float64 `xml:"right,attr"`
	Left   float64 `xml:"left,attr"`
}

// ImagingOptions represents imaging options/capabilities.
type ImagingOptions struct {
	BacklightCompensation *BacklightCompensationOptions `xml:"BacklightCompensation,omitempty"`
	Brightness            *FloatRange                   `xml:"Brightness,omitempty"`
	ColorSaturation       *FloatRange                   `xml:"ColorSaturation,omitempty"`
	Contrast              *FloatRange                   `xml:"Contrast,omitempty"`
	Exposure              *ExposureOptions              `xml:"Exposure,omitempty"`
	Focus                 *FocusOptions                 `xml:"Focus,omitempty"`
	IrCutFilterModes      []string                      `xml:"IrCutFilterModes,omitempty"`
	Sharpness             *FloatRange                   `xml:"Sharpness,omitempty"`
	WideDynamicRange      *WideDynamicRangeOptions      `xml:"WideDynamicRange,omitempty"`
	WhiteBalance          *WhiteBalanceOptions          `xml:"WhiteBalance,omitempty"`
}

// BacklightCompensationOptions represents backlight compensation options.
type BacklightCompensationOptions struct {
	Mode  []string    `xml:"Mode"`
	Level *FloatRange `xml:"Level,omitempty"`
}

// ExposureOptions represents exposure options.
type ExposureOptions struct {
	Mode            []string    `xml:"Mode"`
	Priority        []string    `xml:"Priority,omitempty"`
	MinExposureTime *FloatRange `xml:"MinExposureTime,omitempty"`
	MaxExposureTime *FloatRange `xml:"MaxExposureTime,omitempty"`
	MinGain         *FloatRange `xml:"MinGain,omitempty"`
	MaxGain         *FloatRange `xml:"MaxGain,omitempty"`
	MinIris         *FloatRange `xml:"MinIris,omitempty"`
	MaxIris         *FloatRange `xml:"MaxIris,omitempty"`
	ExposureTime    *FloatRange `xml:"ExposureTime,omitempty"`
	Gain            *FloatRange `xml:"Gain,omitempty"`
	Iris            *FloatRange `xml:"Iris,omitempty"`
}

// FocusOptions represents focus options.
type FocusOptions struct {
	AutoFocusModes []string    `xml:"AutoFocusModes"`
	DefaultSpeed   *FloatRange `xml:"DefaultSpeed,omitempty"`
	NearLimit      *FloatRange `xml:"NearLimit,omitempty"`
	FarLimit       *FloatRange `xml:"FarLimit,omitempty"`
}

// WideDynamicRangeOptions represents WDR options.
type WideDynamicRangeOptions struct {
	Mode  []string    `xml:"Mode"`
	Level *FloatRange `xml:"Level,omitempty"`
}

// WhiteBalanceOptions represents white balance options.
type WhiteBalanceOptions struct {
	Mode   []string    `xml:"Mode"`
	YrGain *FloatRange `xml:"YrGain,omitempty"`
	YbGain *FloatRange `xml:"YbGain,omitempty"`
}

// FocusMove represents focus move parameters.
type FocusMove struct {
	Absolute   *AbsoluteFocus   `xml:"Absolute,omitempty"`
	Relative   *RelativeFocus   `xml:"Relative,omitempty"`
	Continuous *ContinuousFocus `xml:"Continuous,omitempty"`
}

// AbsoluteFocus represents absolute focus.
type AbsoluteFocus struct {
	Position float64  `xml:"Position"`
	Speed    *float64 `xml:"Speed,omitempty"`
}

// RelativeFocus represents relative focus.
type RelativeFocus struct {
	Distance float64  `xml:"Distance"`
	Speed    *float64 `xml:"Speed,omitempty"`
}

// ContinuousFocus represents continuous focus.
type ContinuousFocus struct {
	Speed float64 `xml:"Speed"`
}

// PTZVector represents PTZ position/velocity.
type PTZVector struct {
	PanTilt *Vector2D `xml:"PanTilt,omitempty"`
	Zoom    *Vector1D `xml:"Zoom,omitempty"`
}

// Vector2D represents a 2D vector.
type Vector2D struct {
	X     float64 `xml:"x,attr"`
	Y     float64 `xml:"y,attr"`
	Space string  `xml:"space,attr,omitempty"`
}

// Vector1D represents a 1D vector.
type Vector1D struct {
	X     float64 `xml:"x,attr"`
	Space string  `xml:"space,attr,omitempty"`
}

// FloatRange represents a float range.
type FloatRange struct {
	Min float64 `xml:"Min"`
	Max float64 `xml:"Max"`
}
