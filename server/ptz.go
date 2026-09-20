package server

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
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

// GetStatusResponse represents GetStatus response. The PTZStatus wrapper
// is locally declared in the PTZ WSDL (tptz); its typed children are
// ver10/schema.
type GetStatusResponse struct {
	XMLName   xml.Name   `xml:"http://www.onvif.org/ver20/ptz/wsdl GetStatusResponse"`
	PTZStatus *PTZStatus `xml:"http://www.onvif.org/ver20/ptz/wsdl PTZStatus"`
}

// PTZStatus represents PTZ status. Schema-typed children carry the
// ver10/schema namespace (#90).
type PTZStatus struct {
	Position   respPTZVector `xml:"http://www.onvif.org/ver10/schema Position"`
	MoveStatus PTZMoveStatus `xml:"http://www.onvif.org/ver10/schema MoveStatus"`
	UTCTime    string        `xml:"http://www.onvif.org/ver10/schema UtcTime"`
}

// respVector2D is the serialization shape of tt:Vector2D.
type respVector2D struct {
	X     float64 `xml:"x,attr"`
	Y     float64 `xml:"y,attr"`
	Space string  `xml:"space,attr,omitempty"`
}

// respVector1D is the serialization shape of tt:Vector1D.
type respVector1D struct {
	X     float64 `xml:"x,attr"`
	Space string  `xml:"space,attr,omitempty"`
}

// respPTZVector is the serialization shape of tt:PTZVector. The shared
// provider.PTZVector keeps unprefixed tags because it doubles as a
// lenient request decode target (#90).
type respPTZVector struct {
	PanTilt *respVector2D `xml:"http://www.onvif.org/ver10/schema PanTilt,omitempty"`
	Zoom    *respVector1D `xml:"http://www.onvif.org/ver10/schema Zoom,omitempty"`
}

// PTZMoveStatus represents PTZ movement status.
type PTZMoveStatus struct {
	PanTilt string `xml:"http://www.onvif.org/ver10/schema PanTilt,omitempty"`
	Zoom    string `xml:"http://www.onvif.org/ver10/schema Zoom,omitempty"`
}

// GetPresetsRequest represents GetPresets request.
type GetPresetsRequest struct {
	XMLName      xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GetPresets"`
	ProfileToken string   `xml:"ProfileToken"`
}

// GetPresetsResponse represents GetPresets response. The Preset wrapper
// element is locally declared in the PTZ WSDL (tptz); PTZPreset's own
// children are ver10/schema.
type GetPresetsResponse struct {
	XMLName xml.Name    `xml:"http://www.onvif.org/ver20/ptz/wsdl GetPresetsResponse"`
	Preset  []PTZPreset `xml:"http://www.onvif.org/ver20/ptz/wsdl Preset"`
}

// PTZPreset represents a PTZ preset.
type PTZPreset struct {
	Token       string         `xml:"token,attr"`
	Name        string         `xml:"http://www.onvif.org/ver10/schema Name"`
	PTZPosition *respPTZVector `xml:"http://www.onvif.org/ver10/schema PTZPosition,omitempty"`
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

// SetPresetResponse represents SetPreset response. The WSDL declares the
// response empty; the PresetToken echo is kept (tptz-namespaced) to
// match the Rust twin and common device practice — clients need the
// generated token when they did not supply one.
type SetPresetResponse struct {
	XMLName     xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl SetPresetResponse"`
	PresetToken string   `xml:"http://www.onvif.org/ver20/ptz/wsdl PresetToken,omitempty"`
}

// RemovePresetRequest represents RemovePreset request.
type RemovePresetRequest struct {
	XMLName      xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl RemovePreset"`
	ProfileToken string   `xml:"ProfileToken"`
	PresetName   string   `xml:"PresetName,omitempty"`
	PresetToken  string   `xml:"PresetToken,omitempty"`
}

// RemovePresetResponse represents RemovePreset response.
type RemovePresetResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl RemovePresetResponse"`
}

// GetConfigurationsRequest represents GetConfigurations request (no
// parameters).
type GetConfigurationsRequest struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GetConfigurations"`
}

// GetConfigurationsResponse represents GetConfigurations response. The
// PTZConfiguration wrapper elements are locally declared in the PTZ WSDL
// (tptz); the configuration's own children are ver10/schema.
type GetConfigurationsResponse struct {
	XMLName          xml.Name              `xml:"http://www.onvif.org/ver20/ptz/wsdl GetConfigurationsResponse"`
	PTZConfiguration []PTZConfigurationExt `xml:"http://www.onvif.org/ver20/ptz/wsdl PTZConfiguration"`
}

// PTZConfigurationExt represents PTZ configuration with extensions.
type PTZConfigurationExt struct {
	Token         string         `xml:"token,attr"`
	Name          string         `xml:"http://www.onvif.org/ver10/schema Name"`
	UseCount      int            `xml:"http://www.onvif.org/ver10/schema UseCount"`
	NodeToken     string         `xml:"http://www.onvif.org/ver10/schema NodeToken"`
	PanTiltLimits *PanTiltLimits `xml:"http://www.onvif.org/ver10/schema PanTiltLimits,omitempty"`
	ZoomLimits    *ZoomLimits    `xml:"http://www.onvif.org/ver10/schema ZoomLimits,omitempty"`
}

// PanTiltLimits represents pan/tilt limits.
type PanTiltLimits struct {
	Range Space2DDescription `xml:"http://www.onvif.org/ver10/schema Range"`
}

// ZoomLimits represents zoom limits.
type ZoomLimits struct {
	Range Space1DDescription `xml:"http://www.onvif.org/ver10/schema Range"`
}

// Space2DDescription represents 2D space description.
type Space2DDescription struct {
	URI    string         `xml:"http://www.onvif.org/ver10/schema URI"`
	XRange respFloatRange `xml:"http://www.onvif.org/ver10/schema XRange"`
	YRange respFloatRange `xml:"http://www.onvif.org/ver10/schema YRange"`
}

// Space1DDescription represents 1D space description.
type Space1DDescription struct {
	URI    string         `xml:"http://www.onvif.org/ver10/schema URI"`
	XRange respFloatRange `xml:"http://www.onvif.org/ver10/schema XRange"`
}

// respFloatRange is shared with the imaging service (imaging.go): the
// ns-correct serialization shape of tt:FloatRange.

// GetNodesRequest represents GetNodes request (no parameters).
type GetNodesRequest struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GetNodes"`
}

// GetNodesResponse represents GetNodes response. PTZNode is tptz-local;
// its children (per tt:PTZNode) are ver10/schema. The token attribute
// is inherited from tt:DeviceEntity.
type GetNodesResponse struct {
	XMLName xml.Name      `xml:"http://www.onvif.org/ver20/ptz/wsdl GetNodesResponse"`
	PTZNode []respPTZNode `xml:"http://www.onvif.org/ver20/ptz/wsdl PTZNode"`
}

// respPTZNode is the ns-correct serialization shape of tt:PTZNode.
type respPTZNode struct {
	Token                  string         `xml:"token,attr"`
	Name                   string         `xml:"http://www.onvif.org/ver10/schema Name"`
	SupportedPTZSpaces     *respPTZSpaces `xml:"http://www.onvif.org/ver10/schema SupportedPTZSpaces,omitempty"`
	MaximumNumberOfPresets int            `xml:"http://www.onvif.org/ver10/schema MaximumNumberOfPresets"`
	HomeSupported          bool           `xml:"http://www.onvif.org/ver10/schema HomeSupported"`
}

// respPTZSpaces is tt:SupportedPTZSpaces — the movement spaces the node
// serves, emitted per the configuration's capability flags.
type respPTZSpaces struct {
	AbsolutePanTiltPositionSpace   *respSpace2D `xml:"http://www.onvif.org/ver10/schema AbsolutePanTiltPositionSpace,omitempty"`
	AbsoluteZoomPositionSpace      *respSpace1D `xml:"http://www.onvif.org/ver10/schema AbsoluteZoomPositionSpace,omitempty"`
	ContinuousPanTiltVelocitySpace *respSpace2D `xml:"http://www.onvif.org/ver10/schema ContinuousPanTiltVelocitySpace,omitempty"`
	ContinuousZoomVelocitySpace    *respSpace1D `xml:"http://www.onvif.org/ver10/schema ContinuousZoomVelocitySpace,omitempty"`
}

// respSpace2D is one tt:Space2DDescription entry inside SupportedPTZSpaces.
type respSpace2D struct {
	URI    string         `xml:"http://www.onvif.org/ver10/schema URI"`
	XRange respFloatRange `xml:"http://www.onvif.org/ver10/schema XRange"`
	YRange respFloatRange `xml:"http://www.onvif.org/ver10/schema YRange"`
}

// respSpace1D is one tt:Space1DDescription entry inside SupportedPTZSpaces.
type respSpace1D struct {
	URI    string         `xml:"http://www.onvif.org/ver10/schema URI"`
	XRange respFloatRange `xml:"http://www.onvif.org/ver10/schema XRange"`
}

// Canonical PTZ space URIs (ver10/tptz).
const (
	spacePanTiltPosition = "http://www.onvif.org/ver10/tptz/PanTiltSpaces/PositionGenericSpace"
	spaceZoomPosition    = "http://www.onvif.org/ver10/tptz/ZoomSpaces/PositionGenericSpace"
	spacePanTiltVelocity = "http://www.onvif.org/ver10/tptz/PanTiltSpaces/VelocityGenericSpace"
	spaceZoomVelocity    = "http://www.onvif.org/ver10/tptz/ZoomSpaces/VelocityGenericSpace"
)

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
		Position: respPTZVector{
			PanTilt: &respVector2D{
				X:     state.Position.Pan,
				Y:     state.Position.Tilt,
				Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/PositionGenericSpace",
			},
			Zoom: &respVector1D{
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

// HandleGetPresets handles GetPresets request. When the PTZ provider
// implements PTZPresetReader (mutable storage — the simulator does), the
// live store is served; otherwise presets come from the static profile
// configuration.
func (s *Server) HandleGetPresets(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req GetPresetsRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if reader, ok := s.ptz.(provider.PTZPresetReader); ok {
		presets, err := reader.Presets(req.ProfileToken)
		if err != nil {
			return nil, err
		}

		out := make([]PTZPreset, len(presets))
		for i, preset := range presets {
			out[i] = presetToWire(preset)
		}

		return &GetPresetsResponse{Preset: out}, nil
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
		presets[i] = presetToWire(preset)
	}

	return &GetPresetsResponse{
		Preset: presets,
	}, nil
}

// presetToWire maps a provider.Preset onto the response shape.
func presetToWire(preset provider.Preset) PTZPreset {
	return PTZPreset{
		Token: preset.Token,
		Name:  preset.Name,
		PTZPosition: &respPTZVector{
			PanTilt: &respVector2D{
				X: preset.Position.Pan,
				Y: preset.Position.Tilt,
			},
			Zoom: &respVector1D{
				X: preset.Position.Zoom,
			},
		},
	}
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

// HandleSetPreset handles SetPreset request — registered only when the
// PTZ provider implements PTZPresetWriter.
func (s *Server) HandleSetPreset(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req SetPresetRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	writer, ok := s.ptz.(provider.PTZPresetWriter)
	if !ok {
		return nil, fmt.Errorf("%w: SetPreset", ErrPTZNotSupported)
	}

	token, err := writer.SetPreset(req.ProfileToken, req.PresetName, req.PresetToken)
	if err != nil {
		return nil, err
	}

	return &SetPresetResponse{PresetToken: token}, nil
}

// HandleRemovePreset handles RemovePreset request — registered only when
// the PTZ provider implements PTZPresetWriter.
func (s *Server) HandleRemovePreset(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req RemovePresetRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	writer, ok := s.ptz.(provider.PTZPresetWriter)
	if !ok {
		return nil, fmt.Errorf("%w: RemovePreset", ErrPTZNotSupported)
	}

	if err := writer.RemovePreset(req.ProfileToken, req.PresetToken); err != nil {
		return nil, err
	}

	return &RemovePresetResponse{}, nil
}

// HandleGetConfigurations handles GetConfigurations request — one
// PTZConfiguration per PTZ-capable profile, from the static profile
// configuration.
func (s *Server) HandleGetConfigurations(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req GetConfigurationsRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	configs := make([]PTZConfigurationExt, 0, len(s.config.Profiles))
	for i := range s.config.Profiles {
		profile := &s.config.Profiles[i]
		if profile.PTZ == nil {
			continue
		}

		configs = append(configs, PTZConfigurationExt{
			Token:     profile.Token,
			Name:      profile.Name,
			UseCount:  1,
			NodeToken: profile.PTZ.NodeToken,
			PanTiltLimits: &PanTiltLimits{Range: Space2DDescription{
				URI:    spacePanTiltPosition,
				XRange: respFloatRange{Min: profile.PTZ.PanRange.Min, Max: profile.PTZ.PanRange.Max},
				YRange: respFloatRange{Min: profile.PTZ.TiltRange.Min, Max: profile.PTZ.TiltRange.Max},
			}},
			ZoomLimits: &ZoomLimits{Range: Space1DDescription{
				URI:    spaceZoomPosition,
				XRange: respFloatRange{Min: profile.PTZ.ZoomRange.Min, Max: profile.PTZ.ZoomRange.Max},
			}},
		})
	}

	return &GetConfigurationsResponse{PTZConfiguration: configs}, nil
}

// HandleGetNodes handles GetNodes request — the distinct PTZ nodes
// behind the profile set, with the movement spaces the configuration
// flags advertise.
func (s *Server) HandleGetNodes(rc *soap.RequestContext, body []byte) (interface{}, error) {
	var req GetNodesRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	seen := make(map[string]bool)
	nodes := make([]respPTZNode, 0, len(s.config.Profiles))
	for i := range s.config.Profiles {
		profile := &s.config.Profiles[i]
		if profile.PTZ == nil || seen[profile.PTZ.NodeToken] {
			continue
		}

		seen[profile.PTZ.NodeToken] = true

		var spaces *respPTZSpaces
		if profile.PTZ.SupportsAbsolute || profile.PTZ.SupportsContinuous {
			spaces = &respPTZSpaces{}
			if profile.PTZ.SupportsAbsolute {
				spaces.AbsolutePanTiltPositionSpace = &respSpace2D{
					URI:    spacePanTiltPosition,
					XRange: respFloatRange{Min: profile.PTZ.PanRange.Min, Max: profile.PTZ.PanRange.Max},
					YRange: respFloatRange{Min: profile.PTZ.TiltRange.Min, Max: profile.PTZ.TiltRange.Max},
				}
				spaces.AbsoluteZoomPositionSpace = &respSpace1D{
					URI:    spaceZoomPosition,
					XRange: respFloatRange{Min: profile.PTZ.ZoomRange.Min, Max: profile.PTZ.ZoomRange.Max},
				}
			}
			if profile.PTZ.SupportsContinuous {
				spaces.ContinuousPanTiltVelocitySpace = &respSpace2D{
					URI:    spacePanTiltVelocity,
					XRange: respFloatRange{Min: -1, Max: 1},
					YRange: respFloatRange{Min: -1, Max: 1},
				}
				spaces.ContinuousZoomVelocitySpace = &respSpace1D{
					URI:    spaceZoomVelocity,
					XRange: respFloatRange{Min: -1, Max: 1},
				}
			}
		}

		nodes = append(nodes, respPTZNode{
			Token:                  profile.PTZ.NodeToken,
			Name:                   profile.PTZ.NodeToken,
			SupportedPTZSpaces:     spaces,
			MaximumNumberOfPresets: maxPresetsPerNode,
			HomeSupported:          false,
		})
	}

	return &GetNodesResponse{PTZNode: nodes}, nil
}

// maxPresetsPerNode is the simulator's advertised preset capacity.
const maxPresetsPerNode = 32

// Helper functions

func getMoveStatusString(moving bool) string {
	if moving {
		return "MOVING"
	}

	return "IDLE"
}
