package server

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// PTZ service SOAP message types

// ContinuousMoveRequest represents ContinuousMove request.
type ContinuousMoveRequest struct {
	XMLName      xml.Name  `xml:"http://www.onvif.org/ver20/ptz/wsdl ContinuousMove"`
	ProfileToken string    `xml:"ProfileToken"`
	Velocity     PTZVector `xml:"Velocity"`
	Timeout      string    `xml:"Timeout,omitempty"`
}

// ContinuousMoveResponse represents ContinuousMove response.
type ContinuousMoveResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl ContinuousMoveResponse"`
}

// AbsoluteMoveRequest represents AbsoluteMove request.
type AbsoluteMoveRequest struct {
	XMLName      xml.Name  `xml:"http://www.onvif.org/ver20/ptz/wsdl AbsoluteMove"`
	ProfileToken string    `xml:"ProfileToken"`
	Position     PTZVector `xml:"Position"`
	Speed        PTZVector `xml:"Speed,omitempty"`
}

// AbsoluteMoveResponse represents AbsoluteMove response.
type AbsoluteMoveResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl AbsoluteMoveResponse"`
}

// RelativeMoveRequest represents RelativeMove request.
type RelativeMoveRequest struct {
	XMLName      xml.Name  `xml:"http://www.onvif.org/ver20/ptz/wsdl RelativeMove"`
	ProfileToken string    `xml:"ProfileToken"`
	Translation  PTZVector `xml:"Translation"`
	Speed        PTZVector `xml:"Speed,omitempty"`
}

// RelativeMoveResponse represents RelativeMove response.
type RelativeMoveResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl RelativeMoveResponse"`
}

// StopRequest represents Stop request.
type StopRequest struct {
	XMLName      xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl Stop"`
	ProfileToken string   `xml:"ProfileToken"`
	PanTilt      bool     `xml:"PanTilt,omitempty"`
	Zoom         bool     `xml:"Zoom,omitempty"`
}

// StopResponse represents Stop response.
type StopResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl StopResponse"`
}

// GetStatusRequest represents GetStatus request.
type GetStatusRequest struct {
	XMLName      xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GetStatus"`
	ProfileToken string   `xml:"ProfileToken"`
}

// GetStatusResponse represents GetStatus response.
type GetStatusResponse struct {
	XMLName   xml.Name   `xml:"http://www.onvif.org/ver20/ptz/wsdl GetStatusResponse"`
	PTZStatus *PTZStatus `xml:"PTZStatus"`
}

// PTZStatus represents PTZ status.
type PTZStatus struct {
	Position   PTZVector     `xml:"Position"`
	MoveStatus PTZMoveStatus `xml:"MoveStatus"`
	UTCTime    string        `xml:"UtcTime"`
}

// PTZMoveStatus represents PTZ movement status.
type PTZMoveStatus struct {
	PanTilt string `xml:"PanTilt,omitempty"`
	Zoom    string `xml:"Zoom,omitempty"`
}

// GetPresetsRequest represents GetPresets request.
type GetPresetsRequest struct {
	XMLName      xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GetPresets"`
	ProfileToken string   `xml:"ProfileToken"`
}

// GetPresetsResponse represents GetPresets response.
type GetPresetsResponse struct {
	XMLName xml.Name    `xml:"http://www.onvif.org/ver20/ptz/wsdl GetPresetsResponse"`
	Preset  []PTZPreset `xml:"Preset"`
}

// PTZPreset represents a PTZ preset.
type PTZPreset struct {
	Token       string     `xml:"token,attr"`
	Name        string     `xml:"Name"`
	PTZPosition *PTZVector `xml:"PTZPosition,omitempty"`
}

// GotoPresetRequest represents GotoPreset request.
type GotoPresetRequest struct {
	XMLName      xml.Name  `xml:"http://www.onvif.org/ver20/ptz/wsdl GotoPreset"`
	ProfileToken string    `xml:"ProfileToken"`
	PresetToken  string    `xml:"PresetToken"`
	Speed        PTZVector `xml:"Speed,omitempty"`
}

// GotoPresetResponse represents GotoPreset response.
type GotoPresetResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GotoPresetResponse"`
}

// SetPresetRequest represents SetPreset request.
type SetPresetRequest struct {
	XMLName      xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl SetPreset"`
	ProfileToken string   `xml:"ProfileToken"`
	PresetName   string   `xml:"PresetName,omitempty"`
	PresetToken  string   `xml:"PresetToken,omitempty"`
}

// SetPresetResponse represents SetPreset response.
type SetPresetResponse struct {
	XMLName     xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl SetPresetResponse"`
	PresetToken string   `xml:"PresetToken"`
}

// GetConfigurationsResponse represents GetConfigurations response.
type GetConfigurationsResponse struct {
	XMLName          xml.Name              `xml:"http://www.onvif.org/ver20/ptz/wsdl GetConfigurationsResponse"`
	PTZConfiguration []PTZConfigurationExt `xml:"PTZConfiguration"`
}

// PTZConfigurationExt represents PTZ configuration with extensions.
type PTZConfigurationExt struct {
	Token         string         `xml:"token,attr"`
	Name          string         `xml:"Name"`
	UseCount      int            `xml:"UseCount"`
	NodeToken     string         `xml:"NodeToken"`
	PanTiltLimits *PanTiltLimits `xml:"PanTiltLimits,omitempty"`
	ZoomLimits    *ZoomLimits    `xml:"ZoomLimits,omitempty"`
}

// PanTiltLimits represents pan/tilt limits.
type PanTiltLimits struct {
	Range Space2DDescription `xml:"Range"`
}

// ZoomLimits represents zoom limits.
type ZoomLimits struct {
	Range Space1DDescription `xml:"Range"`
}

// Space2DDescription represents 2D space description.
type Space2DDescription struct {
	URI    string     `xml:"URI"`
	XRange FloatRange `xml:"XRange"`
	YRange FloatRange `xml:"YRange"`
}

// Space1DDescription represents 1D space description.
type Space1DDescription struct {
	URI    string     `xml:"URI"`
	XRange FloatRange `xml:"XRange"`
}

// PTZ service handlers - stateless translations between SOAP and the
// PTZ provider.

// HandleContinuousMove handles ContinuousMove request.
func (s *Server) HandleContinuousMove(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req ContinuousMoveRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if err := s.ptz.ContinuousMove(req.ProfileToken, req.Velocity, req.Timeout); err != nil {
		return nil, err
	}

	return &ContinuousMoveResponse{}, nil
}

// HandleAbsoluteMove handles AbsoluteMove request.
func (s *Server) HandleAbsoluteMove(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req AbsoluteMoveRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if err := s.ptz.AbsoluteMove(req.ProfileToken, req.Position); err != nil {
		return nil, err
	}

	return &AbsoluteMoveResponse{}, nil
}

// HandleRelativeMove handles RelativeMove request.
func (s *Server) HandleRelativeMove(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req RelativeMoveRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if err := s.ptz.RelativeMove(req.ProfileToken, req.Translation); err != nil {
		return nil, err
	}

	return &RelativeMoveResponse{}, nil
}

// HandleStop handles Stop request.
func (s *Server) HandleStop(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req StopRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if err := s.ptz.Stop(req.ProfileToken, req.PanTilt, req.Zoom); err != nil {
		return nil, err
	}

	return &StopResponse{}, nil
}

// HandleGetStatus handles GetStatus request.
func (s *Server) HandleGetStatus(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req GetStatusRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	state, err := s.ptz.Status(req.ProfileToken)
	if err != nil {
		return nil, err
	}

	// Build status response
	status := &PTZStatus{
		Position: PTZVector{
			PanTilt: &Vector2D{
				X:     state.Position.Pan,
				Y:     state.Position.Tilt,
				Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/PositionGenericSpace",
			},
			Zoom: &Vector1D{
				X:     state.Position.Zoom,
				Space: "http://www.onvif.org/ver10/tptz/ZoomSpaces/PositionGenericSpace",
			},
		},
		MoveStatus: PTZMoveStatus{
			PanTilt: getMoveStatusString(state.PanMoving || state.TiltMoving),
			Zoom:    getMoveStatusString(state.ZoomMoving),
		},
		UTCTime: time.Now().UTC().Format(time.RFC3339),
	}

	return &GetStatusResponse{
		PTZStatus: status,
	}, nil
}

// HandleGetPresets handles GetPresets request.
func (s *Server) HandleGetPresets(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req GetPresetsRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Find the profile configuration
	var profileCfg *ProfileConfig
	for i := range s.config.Profiles {
		if s.config.Profiles[i].Token == req.ProfileToken {
			profileCfg = &s.config.Profiles[i]

			break
		}
	}

	if profileCfg == nil || profileCfg.PTZ == nil {
		return nil, fmt.Errorf("%w: %s", ErrPTZNotSupported, req.ProfileToken)
	}

	// Build presets response
	presets := make([]PTZPreset, len(profileCfg.PTZ.Presets))
	for i, preset := range profileCfg.PTZ.Presets {
		presets[i] = PTZPreset{
			Token: preset.Token,
			Name:  preset.Name,
			PTZPosition: &PTZVector{
				PanTilt: &Vector2D{
					X: preset.Position.Pan,
					Y: preset.Position.Tilt,
				},
				Zoom: &Vector1D{
					X: preset.Position.Zoom,
				},
			},
		}
	}

	return &GetPresetsResponse{
		Preset: presets,
	}, nil
}

// HandleGotoPreset handles GotoPreset request.
func (s *Server) HandleGotoPreset(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req GotoPresetRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if err := s.ptz.GotoPreset(req.ProfileToken, req.PresetToken); err != nil {
		return nil, err
	}

	return &GotoPresetResponse{}, nil
}

// Helper functions

func getMoveStatusString(moving bool) string {
	if moving {
		return "MOVING"
	}

	return "IDLE"
}
