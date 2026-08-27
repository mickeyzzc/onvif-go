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
