package onvif

import (
	"context"
	"encoding/xml"
	"fmt"
)

// PTZ service namespace.
const ptzNamespace = "http://www.onvif.org/ver20/ptz/wsdl"

// ptzPanTiltXML is a shared type for PTZ pan/tilt XML serialization.
type ptzPanTiltXML struct {
	X     float64 `xml:"x,attr"`
	Y     float64 `xml:"y,attr"`
	Space string  `xml:"space,attr,omitempty"`
}

// ptzZoomXML is a shared type for PTZ zoom XML serialization.
type ptzZoomXML struct {
	X     float64 `xml:"x,attr"`
	Space string  `xml:"space,attr,omitempty"`
}

// ptzVectorXML is a shared type for PTZ position/velocity XML serialization.
type ptzVectorXML struct {
	PanTilt *ptzPanTiltXML `xml:"PanTilt,omitempty"`
	Zoom    *ptzZoomXML    `xml:"Zoom,omitempty"`
}

// ptzSpeedXML is a shared type for PTZ speed XML serialization.
type ptzSpeedXML struct {
	PanTilt *ptzPanTiltXML `xml:"PanTilt,omitempty"`
	Zoom    *ptzZoomXML    `xml:"Zoom,omitempty"`
}

// convertToPTZVectorXML converts PTZVector to XML struct.
func convertToPTZVectorXML(v *PTZVector) *ptzVectorXML {
	if v == nil {
		return nil
	}
	result := &ptzVectorXML{}
	if v.PanTilt != nil {
		result.PanTilt = &ptzPanTiltXML{X: v.PanTilt.X, Y: v.PanTilt.Y, Space: v.PanTilt.Space}
	}
	if v.Zoom != nil {
		result.Zoom = &ptzZoomXML{X: v.Zoom.X, Space: v.Zoom.Space}
	}

	return result
}

// convertToPTZSpeedXML converts PTZSpeed to XML struct.
func convertToPTZSpeedXML(s *PTZSpeed) *ptzSpeedXML {
	if s == nil {
		return nil
	}
	result := &ptzSpeedXML{}
	if s.PanTilt != nil {
		result.PanTilt = &ptzPanTiltXML{X: s.PanTilt.X, Y: s.PanTilt.Y, Space: s.PanTilt.Space}
	}
	if s.Zoom != nil {
		result.Zoom = &ptzZoomXML{X: s.Zoom.X, Space: s.Zoom.Space}
	}

	return result
}

// ContinuousMove starts continuous PTZ movement.
// getEndpoint returns the service endpoint, falling back to the device
// endpoint when Initialize has not resolved (or the device does not
// advertise) a service-specific address. The read is lock-guarded: the
// client may be initialized concurrently with calls (issue #12).
func (s *PTZService) getEndpoint() string {
	s.client.mu.RLock()
	defer s.client.mu.RUnlock()

	if s.client.ptzEndpoint != "" {
		return s.client.ptzEndpoint
	}

	return s.client.endpoint
}

func (s *PTZService) ContinuousMove(ctx context.Context, profileToken string, velocity *PTZSpeed, timeout *string) error {
	endpoint := s.getEndpoint()
	if endpoint == "" {
		return ErrServiceNotSupported
	}

	type ContinuousMove struct {
		XMLName      xml.Name     `xml:"tptz:ContinuousMove"`
		Xmlns        string       `xml:"xmlns:tptz,attr"`
		ProfileToken string       `xml:"tptz:ProfileToken"`
		Velocity     *ptzSpeedXML `xml:"tptz:Velocity"`
		Timeout      *string      `xml:"tptz:Timeout,omitempty"`
	}

	req := ContinuousMove{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
		Velocity:     convertToPTZSpeedXML(velocity),
		Timeout:      timeout,
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("ContinuousMove failed: %w", err)
	}

	return nil
}

// AbsoluteMove moves PTZ to an absolute position.
func (s *PTZService) AbsoluteMove(ctx context.Context, profileToken string, position *PTZVector, speed *PTZSpeed) error {
	endpoint := s.getEndpoint()
	if endpoint == "" {
		return ErrServiceNotSupported
	}

	type AbsoluteMove struct {
		XMLName      xml.Name      `xml:"tptz:AbsoluteMove"`
		Xmlns        string        `xml:"xmlns:tptz,attr"`
		ProfileToken string        `xml:"tptz:ProfileToken"`
		Position     *ptzVectorXML `xml:"tptz:Position"`
		Speed        *ptzSpeedXML  `xml:"tptz:Speed,omitempty"`
	}

	req := AbsoluteMove{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
		Position:     convertToPTZVectorXML(position),
		Speed:        convertToPTZSpeedXML(speed),
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("AbsoluteMove failed: %w", err)
	}

	return nil
}

// RelativeMove moves PTZ relative to current position.
func (s *PTZService) RelativeMove(ctx context.Context, profileToken string, translation *PTZVector, speed *PTZSpeed) error {
	endpoint := s.getEndpoint()
	if endpoint == "" {
		return ErrServiceNotSupported
	}

	type RelativeMove struct {
		XMLName      xml.Name      `xml:"tptz:RelativeMove"`
		Xmlns        string        `xml:"xmlns:tptz,attr"`
		ProfileToken string        `xml:"tptz:ProfileToken"`
		Translation  *ptzVectorXML `xml:"tptz:Translation"`
		Speed        *ptzSpeedXML  `xml:"tptz:Speed,omitempty"`
	}

	req := RelativeMove{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
		Translation:  convertToPTZVectorXML(translation),
		Speed:        convertToPTZSpeedXML(speed),
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RelativeMove failed: %w", err)
	}

	return nil
}

// Stop stops PTZ movement.
func (s *PTZService) Stop(ctx context.Context, profileToken string, panTilt, zoom bool) error {
	endpoint := s.getEndpoint()
	if endpoint == "" {
		return ErrServiceNotSupported
	}

	type Stop struct {
		XMLName      xml.Name `xml:"tptz:Stop"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
		PanTilt      *bool    `xml:"tptz:PanTilt,omitempty"`
		Zoom         *bool    `xml:"tptz:Zoom,omitempty"`
	}

	req := Stop{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
	}

	if panTilt {
		req.PanTilt = &panTilt
	}
	if zoom {
		req.Zoom = &zoom
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("Stop failed: %w", err)
	}

	return nil
}

// GetStatus retrieves PTZ status.
func (s *PTZService) GetStatus(ctx context.Context, profileToken string) (*PTZStatus, error) {
	endpoint := s.getEndpoint()
	if endpoint == "" {
		return nil, ErrServiceNotSupported
	}

	type GetStatus struct {
		XMLName      xml.Name `xml:"tptz:GetStatus"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
	}

	type GetStatusResponse struct {
		XMLName   xml.Name `xml:"GetStatusResponse"`
		PTZStatus struct {
			Position *struct {
				PanTilt *struct {
					X     float64 `xml:"x,attr"`
					Y     float64 `xml:"y,attr"`
					Space string  `xml:"space,attr,omitempty"`
				} `xml:"PanTilt"`
				Zoom *struct {
					X     float64 `xml:"x,attr"`
					Space string  `xml:"space,attr,omitempty"`
				} `xml:"Zoom"`
			} `xml:"Position"`
			MoveStatus *struct {
				PanTilt string `xml:"PanTilt"`
				Zoom    string `xml:"Zoom"`
			} `xml:"MoveStatus"`
			Error   string `xml:"Error"`
			UTCTime string `xml:"UtcTime"`
		} `xml:"PTZStatus"`
	}

	req := GetStatus{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
	}

	var resp GetStatusResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetStatus failed: %w", err)
	}

	status := &PTZStatus{
		Error: resp.PTZStatus.Error,
	}

	if resp.PTZStatus.Position != nil {
		status.Position = &PTZVector{}
		if resp.PTZStatus.Position.PanTilt != nil {
			status.Position.PanTilt = &Vector2D{
				X:     resp.PTZStatus.Position.PanTilt.X,
				Y:     resp.PTZStatus.Position.PanTilt.Y,
				Space: resp.PTZStatus.Position.PanTilt.Space,
			}
		}
		if resp.PTZStatus.Position.Zoom != nil {
			status.Position.Zoom = &Vector1D{
				X:     resp.PTZStatus.Position.Zoom.X,
				Space: resp.PTZStatus.Position.Zoom.Space,
			}
		}
	}

	if resp.PTZStatus.MoveStatus != nil {
		status.MoveStatus = &PTZMoveStatus{
			PanTilt: resp.PTZStatus.MoveStatus.PanTilt,
			Zoom:    resp.PTZStatus.MoveStatus.Zoom,
		}
	}

	return status, nil
}

// GetPresets retrieves PTZ presets.
func (s *PTZService) GetPresets(ctx context.Context, profileToken string) ([]*PTZPreset, error) {
	endpoint := s.getEndpoint()
	if endpoint == "" {
		return nil, ErrServiceNotSupported
	}

	type GetPresets struct {
		XMLName      xml.Name `xml:"tptz:GetPresets"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
	}

	type GetPresetsResponse struct {
		XMLName xml.Name `xml:"GetPresetsResponse"`
		Preset  []struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			PTZPosition *struct {
				PanTilt *struct {
					X     float64 `xml:"x,attr"`
					Y     float64 `xml:"y,attr"`
					Space string  `xml:"space,attr,omitempty"`
				} `xml:"PanTilt"`
				Zoom *struct {
					X     float64 `xml:"x,attr"`
					Space string  `xml:"space,attr,omitempty"`
				} `xml:"Zoom"`
			} `xml:"PTZPosition"`
		} `xml:"Preset"`
	}

	req := GetPresets{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
	}

	var resp GetPresetsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetPresets failed: %w", err)
	}

	presets := make([]*PTZPreset, len(resp.Preset))
	for i, p := range resp.Preset {
		preset := &PTZPreset{
			Token: p.Token,
			Name:  p.Name,
		}

		if p.PTZPosition != nil {
			preset.PTZPosition = &PTZVector{}
			if p.PTZPosition.PanTilt != nil {
				preset.PTZPosition.PanTilt = &Vector2D{
					X:     p.PTZPosition.PanTilt.X,
					Y:     p.PTZPosition.PanTilt.Y,
					Space: p.PTZPosition.PanTilt.Space,
				}
			}
			if p.PTZPosition.Zoom != nil {
				preset.PTZPosition.Zoom = &Vector1D{
					X:     p.PTZPosition.Zoom.X,
					Space: p.PTZPosition.Zoom.Space,
				}
			}
		}

		presets[i] = preset
	}

	return presets, nil
}

// GotoPreset moves PTZ to a preset position.
func (s *PTZService) GotoPreset(ctx context.Context, profileToken, presetToken string, speed *PTZSpeed) error {
	endpoint := s.getEndpoint()
	if endpoint == "" {
		return ErrServiceNotSupported
	}

	type GotoPreset struct {
		XMLName      xml.Name     `xml:"tptz:GotoPreset"`
		Xmlns        string       `xml:"xmlns:tptz,attr"`
		ProfileToken string       `xml:"tptz:ProfileToken"`
		PresetToken  string       `xml:"tptz:PresetToken"`
		Speed        *ptzSpeedXML `xml:"tptz:Speed,omitempty"`
	}

	req := GotoPreset{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
		PresetToken:  presetToken,
		Speed:        convertToPTZSpeedXML(speed),
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("GotoPreset failed: %w", err)
	}

	return nil
}

// SetPreset sets a preset position.
func (s *PTZService) SetPreset(ctx context.Context, profileToken, presetName, presetToken string) (string, error) {
	endpoint := s.getEndpoint()
	if endpoint == "" {
		return "", ErrServiceNotSupported
	}

	type SetPreset struct {
		XMLName      xml.Name `xml:"tptz:SetPreset"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
		PresetName   *string  `xml:"tptz:PresetName,omitempty"`
		PresetToken  *string  `xml:"tptz:PresetToken,omitempty"`
	}

	type SetPresetResponse struct {
		XMLName     xml.Name `xml:"SetPresetResponse"`
		PresetToken string   `xml:"PresetToken"`
	}

	req := SetPreset{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
	}

	if presetName != "" {
		req.PresetName = &presetName
	}
	if presetToken != "" {
		req.PresetToken = &presetToken
	}

	var resp SetPresetResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return "", fmt.Errorf("SetPreset failed: %w", err)
	}

	return resp.PresetToken, nil
}

// RemovePreset removes a preset.
func (s *PTZService) RemovePreset(ctx context.Context, profileToken, presetToken string) error {
	endpoint := s.getEndpoint()
	if endpoint == "" {
		return ErrServiceNotSupported
	}

	type RemovePreset struct {
		XMLName      xml.Name `xml:"tptz:RemovePreset"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
		PresetToken  string   `xml:"tptz:PresetToken"`
	}

	req := RemovePreset{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
		PresetToken:  presetToken,
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("RemovePreset failed: %w", err)
	}

	return nil
}

// GotoHomePosition moves PTZ to home position.
func (s *PTZService) GotoHomePosition(ctx context.Context, profileToken string, speed *PTZSpeed) error {
	endpoint := s.getEndpoint()
	if endpoint == "" {
		return ErrServiceNotSupported
	}

	type GotoHomePosition struct {
		XMLName      xml.Name     `xml:"tptz:GotoHomePosition"`
		Xmlns        string       `xml:"xmlns:tptz,attr"`
		ProfileToken string       `xml:"tptz:ProfileToken"`
		Speed        *ptzSpeedXML `xml:"tptz:Speed,omitempty"`
	}

	req := GotoHomePosition{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
		Speed:        convertToPTZSpeedXML(speed),
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("GotoHomePosition failed: %w", err)
	}

	return nil
}

// SetHomePosition sets the current position as home position.
func (s *PTZService) SetHomePosition(ctx context.Context, profileToken string) error {
	endpoint := s.getEndpoint()
	if endpoint == "" {
		return ErrServiceNotSupported
	}

	type SetHomePosition struct {
		XMLName      xml.Name `xml:"tptz:SetHomePosition"`
		Xmlns        string   `xml:"xmlns:tptz,attr"`
		ProfileToken string   `xml:"tptz:ProfileToken"`
	}

	req := SetHomePosition{
		Xmlns:        ptzNamespace,
		ProfileToken: profileToken,
	}

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetHomePosition failed: %w", err)
	}

	return nil
}

// GetConfiguration retrieves PTZ configuration.
func (s *PTZService) GetConfiguration(ctx context.Context, configurationToken string) (*PTZConfiguration, error) {
	endpoint := s.getEndpoint()
	if endpoint == "" {
		return nil, ErrServiceNotSupported
	}

	type GetConfiguration struct {
		XMLName               xml.Name `xml:"tptz:GetConfiguration"`
		Xmlns                 string   `xml:"xmlns:tptz,attr"`
		PTZConfigurationToken string   `xml:"tptz:PTZConfigurationToken"`
	}

	type GetConfigurationResponse struct {
		XMLName          xml.Name `xml:"GetConfigurationResponse"`
		PTZConfiguration struct {
			Token     string `xml:"token,attr"`
			Name      string `xml:"Name"`
			UseCount  int    `xml:"UseCount"`
			NodeToken string `xml:"NodeToken"`
		} `xml:"PTZConfiguration"`
	}

	req := GetConfiguration{
		Xmlns:                 ptzNamespace,
		PTZConfigurationToken: configurationToken,
	}

	var resp GetConfigurationResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetConfiguration failed: %w", err)
	}

	return &PTZConfiguration{
		Token:     resp.PTZConfiguration.Token,
		Name:      resp.PTZConfiguration.Name,
		UseCount:  resp.PTZConfiguration.UseCount,
		NodeToken: resp.PTZConfiguration.NodeToken,
	}, nil
}

// GetConfigurations retrieves all PTZ configurations.
func (s *PTZService) GetConfigurations(ctx context.Context) ([]*PTZConfiguration, error) {
	endpoint := s.getEndpoint()
	if endpoint == "" {
		return nil, ErrServiceNotSupported
	}

	type GetConfigurations struct {
		XMLName xml.Name `xml:"tptz:GetConfigurations"`
		Xmlns   string   `xml:"xmlns:tptz,attr"`
	}

	type GetConfigurationsResponse struct {
		XMLName          xml.Name `xml:"GetConfigurationsResponse"`
		PTZConfiguration []struct {
			Token     string `xml:"token,attr"`
			Name      string `xml:"Name"`
			UseCount  int    `xml:"UseCount"`
			NodeToken string `xml:"NodeToken"`
		} `xml:"PTZConfiguration"`
	}

	req := GetConfigurations{
		Xmlns: ptzNamespace,
	}

	var resp GetConfigurationsResponse

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetConfigurations failed: %w", err)
	}

	configs := make([]*PTZConfiguration, len(resp.PTZConfiguration))
	for i, cfg := range resp.PTZConfiguration {
		configs[i] = &PTZConfiguration{
			Token:     cfg.Token,
			Name:      cfg.Name,
			UseCount:  cfg.UseCount,
			NodeToken: cfg.NodeToken,
		}
	}

	return configs, nil
}
