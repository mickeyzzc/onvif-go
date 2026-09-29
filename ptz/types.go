// Package ptz hosts the PTZ-service (tptz) domain types.
package ptz

import (
	"time"

	"github.com/mickeyzzc/onvif-go/v2/types"
)

// PTZConfiguration represents PTZ configuration.
type PTZConfiguration struct {
	Token                                  string
	Name                                   string
	UseCount                               int
	NodeToken                              string
	DefaultAbsolutePantTiltPositionSpace   string
	DefaultAbsoluteZoomPositionSpace       string
	DefaultRelativePanTiltTranslationSpace string
	DefaultRelativeZoomTranslationSpace    string
	DefaultContinuousPanTiltVelocitySpace  string
	DefaultContinuousZoomVelocitySpace     string
	DefaultPTZSpeed                        *PTZSpeed
	DefaultPTZTimeout                      time.Duration
	PanTiltLimits                          *PanTiltLimits
	ZoomLimits                             *ZoomLimits
}

// PTZSpeed represents PTZ speed.
type PTZSpeed struct {
	PanTilt *Vector2D
	Zoom    *Vector1D
}

// Vector2D represents a 2D vector.
type Vector2D struct {
	X     float64
	Y     float64
	Space string
}

// Vector1D represents a 1D vector.
type Vector1D struct {
	X     float64
	Space string
}

// PanTiltLimits represents pan/tilt limits.
type PanTiltLimits struct {
	Range *Space2DDescription
}

// ZoomLimits represents zoom limits.
type ZoomLimits struct {
	Range *Space1DDescription
}

// Space2DDescription represents 2D space description.
type Space2DDescription struct {
	URI    string
	XRange *types.FloatRange
	YRange *types.FloatRange
}

// Space1DDescription represents 1D space description.
type Space1DDescription struct {
	URI    string
	XRange *types.FloatRange
}

// PTZFilter represents PTZ filter.
type PTZFilter struct {
	Status   bool
	Position bool
}

// PTZStatus represents PTZ status.
type PTZStatus struct {
	Position   *PTZVector
	MoveStatus *PTZMoveStatus
	Error      string
	UTCTime    time.Time
}

// PTZVector represents PTZ position.
type PTZVector struct {
	PanTilt *Vector2D
	Zoom    *Vector1D
}

// PTZMoveStatus represents PTZ movement status.
type PTZMoveStatus struct {
	PanTilt string // IDLE, MOVING, UNKNOWN
	Zoom    string // IDLE, MOVING, UNKNOWN
}

// PTZPreset represents a PTZ preset.
type PTZPreset struct {
	Token       string
	Name        string
	PTZPosition *PTZVector
}

// AuxiliaryData represents auxiliary command data.
type AuxiliaryData string

// PTZSpaces enumerates the coordinate spaces a device supports (part of
// PTZConfigurationOptions and PTZNode).
type PTZSpaces struct {
	AbsolutePanTiltPositionSpace    []Space2DDescription
	AbsoluteZoomPositionSpace       []Space1DDescription
	RelativePanTiltTranslationSpace []Space2DDescription
	RelativeZoomTranslationSpace    []Space1DDescription
	ContinuousPanTiltVelocitySpace  []Space2DDescription
	ContinuousZoomVelocitySpace     []Space1DDescription
}

// PTZConfigurationOptions describes the valid ranges of a PTZ
// configuration (speeds, spaces, timeouts) — the data a UI needs to
// render PTZ controls.
type PTZConfigurationOptions struct {
	Spaces     PTZSpaces
	PTZTimeout time.Duration
}

// PTZNode is a physical or virtual PTZ unit with the spaces it supports.
type PTZNode struct {
	Token                  string
	Name                   string
	FixedHomePosition      bool
	HomeSupported          bool
	SupportedPTZSpaces     *PTZSpaces
	MaximumNumberOfPresets int
}

// PTZServiceCapabilities reports the optional PTZ operations a device
// supports (attributes of GetServiceCapabilities/Capabilities).
type PTZServiceCapabilities struct {
	EFlip                       bool
	Reverse                     bool
	GetCompatibleConfigurations bool
}
