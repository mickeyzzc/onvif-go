package server

import (
	"encoding/xml"
	"fmt"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// Media OSD family (issue #115): the minimal Text/String closed loop
// backed by the simulator's OSD store. Wire shapes follow the ver10
// media WSDL: response children are media-namespace, the OSD body's
// schema-typed children resolve to ver10/schema.

type getOSDsRequest struct {
	XMLName            xml.Name `xml:"GetOSDs"`
	ConfigurationToken string   `xml:"ConfigurationToken"`
}

type osdConfigurationWire struct {
	Token                         string       `xml:"token,attr"`
	VideoSourceConfigurationToken string       `xml:"http://www.onvif.org/ver10/schema VideoSourceConfigurationToken"`
	Type                          string       `xml:"http://www.onvif.org/ver10/schema Type"`
	TextString                    *osdTextWire `xml:"http://www.onvif.org/ver10/schema TextString,omitempty"`
}

type osdTextWire struct {
	Type    string `xml:"Type,attr"`
	Content string `xml:"http://www.onvif.org/ver10/schema Content"`
}

type getOSDsResponse struct {
	XMLName xml.Name               `xml:"http://www.onvif.org/ver10/media/wsdl GetOSDsResponse"`
	OSDs    []osdConfigurationWire `xml:"http://www.onvif.org/ver10/media/wsdl OSDs"`
}

type getOSDRequest struct {
	XMLName  xml.Name `xml:"GetOSD"`
	OSDToken string   `xml:"OSDToken"`
}

type getOSDResponse struct {
	XMLName xml.Name              `xml:"http://www.onvif.org/ver10/media/wsdl GetOSDResponse"`
	OSD     *osdConfigurationWire `xml:"http://www.onvif.org/ver10/media/wsdl OSD"`
}

type setOSDRequest struct {
	XMLName xml.Name             `xml:"SetOSD"`
	OSD     osdConfigurationWire `xml:"OSD"`
}

type createOSDRequest struct {
	XMLName                       xml.Name             `xml:"CreateOSD"`
	VideoSourceConfigurationToken string               `xml:"VideoSourceConfigurationToken"`
	OSD                           osdConfigurationWire `xml:"OSD"`
}

type createOSDResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl CreateOSDResponse"`
	// WSDL shape: OSDToken. The OSD echo with the assigned token keeps
	// clients that read the entity form working too.
	OSDToken string `xml:"http://www.onvif.org/ver10/media/wsdl OSDToken"`
	OSD      struct {
		Token string `xml:"token,attr"`
	} `xml:"http://www.onvif.org/ver10/media/wsdl OSD"`
}

type deleteOSDRequest struct {
	XMLName  xml.Name `xml:"DeleteOSD"`
	OSDToken string   `xml:"OSDToken"`
}

func osdToWire(o provider.OSD) osdConfigurationWire {
	return osdConfigurationWire{
		Token:                         o.Token,
		VideoSourceConfigurationToken: o.VideoSourceConfigurationToken,
		Type:                          o.Type,
		TextString:                    &osdTextWire{Type: "Plain", Content: o.Text},
	}
}

// HandleGetOSDs lists the simulator's OSDs, optionally filtered by the
// video source configuration token.
func (s *Server) HandleGetOSDs(_ *soap.RequestContext, body []byte) (interface{}, error) {
	var req getOSDsRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	stored := s.osd.OSDs(req.ConfigurationToken)
	resp := getOSDsResponse{OSDs: make([]osdConfigurationWire, 0, len(stored))}
	for _, o := range stored {
		resp.OSDs = append(resp.OSDs, osdToWire(o))
	}

	return resp, nil
}

// HandleGetOSD returns one OSD by token.
func (s *Server) HandleGetOSD(_ *soap.RequestContext, body []byte) (interface{}, error) {
	var req getOSDRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	for _, o := range s.osd.OSDs("") {
		if o.Token == req.OSDToken {
			return getOSDResponse{OSD: ptrOf(osdToWire(o))}, nil
		}
	}

	return nil, errOSDNotFound(req.OSDToken)
}

// HandleSetOSD updates an existing OSD.
func (s *Server) HandleSetOSD(_ *soap.RequestContext, body []byte) (interface{}, error) {
	var req setOSDRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	if req.OSD.Token == "" {
		return nil, errOSDNotFound("")
	}

	text := ""
	if req.OSD.TextString != nil {
		text = req.OSD.TextString.Content
	}
	_ = s.osd.UpsertOSD(provider.OSD{
		Token:                         req.OSD.Token,
		VideoSourceConfigurationToken: req.OSD.VideoSourceConfigurationToken,
		Type:                          req.OSD.Type,
		Text:                          text,
	})

	return struct {
		XMLName xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl SetOSDResponse"`
	}{}, nil
}

// HandleCreateOSD adds an OSD; the store assigns the token.
func (s *Server) HandleCreateOSD(_ *soap.RequestContext, body []byte) (interface{}, error) {
	var req createOSDRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	text := ""
	if req.OSD.TextString != nil {
		text = req.OSD.TextString.Content
	}
	created := s.osd.UpsertOSD(provider.OSD{
		Token:                         req.OSD.Token,
		VideoSourceConfigurationToken: req.VideoSourceConfigurationToken,
		Type:                          req.OSD.Type,
		Text:                          text,
	})

	var resp createOSDResponse
	resp.OSDToken = created.Token
	resp.OSD.Token = created.Token

	return resp, nil
}

// HandleDeleteOSD removes an OSD by token.
func (s *Server) HandleDeleteOSD(_ *soap.RequestContext, body []byte) (interface{}, error) {
	var req deleteOSDRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if !s.osd.DeleteOSD(req.OSDToken) {
		return nil, errOSDNotFound(req.OSDToken)
	}

	return struct {
		XMLName xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl DeleteOSDResponse"`
	}{}, nil
}

func errOSDNotFound(token string) error {
	return &soap.SenderFaultError{Reason: fmt.Sprintf("OSD %q not found", token)}
}

func ptrOf[T any](v T) *T { return &v }
