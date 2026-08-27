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
	XMLName         xml.Name         `xml:"http://www.onvif.org/ver20/imaging/wsdl GetImagingSettingsResponse"`
	ImagingSettings *ImagingSettings `xml:"ImagingSettings"`
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
	XMLName        xml.Name        `xml:"http://www.onvif.org/ver20/imaging/wsdl GetOptionsResponse"`
	ImagingOptions *ImagingOptions `xml:"ImagingOptions"`
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
		ImagingSettings: settings,
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
		ImagingOptions: options,
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
