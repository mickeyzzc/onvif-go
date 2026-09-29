package server

import (
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// DeviceIO family (issue #115): relay outputs and digital inputs for NVR
// alarm-linkage integration. Following the onvif-go client's own
// deployment reality (and several major vendors), the relay ops ride the
// DEVICE service endpoint with the tds action shape; GetDigitalInputs
// uses the DeviceIO (tmd) shape. The namespace rules of #93 apply:
// elements locally declared in the deviceIO WSDL are tmd-namespaced,
// schema-typed children resolve to ver10/schema.

type relayOutputWire struct {
	Token        string              `xml:"token,attr"`
	Properties   relayPropertiesWire `xml:"http://www.onvif.org/ver10/deviceIO/wsdl Properties"`
	LogicalState string              `xml:"http://www.onvif.org/ver10/schema LogicalState"`
}

type relayPropertiesWire struct {
	Mode      string `xml:"http://www.onvif.org/ver10/schema Mode"`
	DelayTime string `xml:"http://www.onvif.org/ver10/schema DelayTime"`
	IdleState string `xml:"http://www.onvif.org/ver10/schema IdleState"`
}

type getRelayOutputsResponse struct {
	XMLName      xml.Name          `xml:"http://www.onvif.org/ver10/deviceIO/wsdl GetRelayOutputsResponse"`
	RelayOutputs []relayOutputWire `xml:"http://www.onvif.org/ver10/deviceIO/wsdl RelayOutputs"`
}

type setRelayOutputStateRequest struct {
	XMLName          xml.Name `xml:"SetRelayOutputState"`
	RelayOutputToken string   `xml:"RelayOutputToken"`
	LogicalState     string   `xml:"LogicalState"`
}

type digitalInputWire struct {
	Token     string `xml:"token,attr"`
	IdleState string `xml:"IdleState,attr"`
}

type getDigitalInputsResponse struct {
	XMLName       xml.Name           `xml:"http://www.onvif.org/ver10/deviceIO/wsdl GetDigitalInputsResponse"`
	DigitalInputs []digitalInputWire `xml:"http://www.onvif.org/ver10/deviceIO/wsdl DigitalInputs"`
}

type deviceIOCapabilitiesWire struct {
	XMLName      xml.Name `xml:"http://www.onvif.org/ver10/deviceIO/wsdl GetDeviceIOServiceCapabilitiesResponse"`
	Capabilities struct {
		XMLName       xml.Name `xml:"http://www.onvif.org/ver10/deviceIO/wsdl Capabilities"`
		AudioSources  int      `xml:"AudioSources,attr"`
		AudioOutputs  int      `xml:"AudioOutputs,attr"`
		DigitalInputs int      `xml:"DigitalInputs,attr"`
		RelayOutputs  int      `xml:"RelayOutputs,attr"`
	} `xml:"http://www.onvif.org/ver10/deviceIO/wsdl Capabilities"`
}

// HandleGetRelayOutputs lists the alarm relays with live logical states.
func (s *Server) HandleGetRelayOutputs(_ *soap.RequestContext, _ []byte) (interface{}, error) {
	relays := s.relays.RelayOutputs()
	resp := getRelayOutputsResponse{RelayOutputs: make([]relayOutputWire, 0, len(relays))}
	for _, r := range relays {
		resp.RelayOutputs = append(resp.RelayOutputs, relayOutputWire{
			Token: r.Token,
			Properties: relayPropertiesWire{
				Mode:      r.Mode,
				DelayTime: r.DelayTime,
				IdleState: r.IdleState,
			},
			LogicalState: r.LogicalState,
		})
	}

	return resp, nil
}

// HandleSetRelayOutputState flips a relay (NVR alarm linkage).
func (s *Server) HandleSetRelayOutputState(_ *soap.RequestContext, body []byte) (interface{}, error) {
	var req setRelayOutputStateRequest
	if err := unmarshalBody(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if err := s.relays.SetRelayState(req.RelayOutputToken, req.LogicalState); err != nil {
		if errors.Is(err, provider.ErrNotFound) {
			return nil, &soap.SenderFaultError{Reason: fmt.Sprintf("relay output %q not found", req.RelayOutputToken)}
		}

		return nil, &soap.SenderFaultError{Reason: err.Error()}
	}

	return struct {
		XMLName xml.Name `xml:"http://www.onvif.org/ver10/deviceIO/wsdl SetRelayOutputStateResponse"`
	}{}, nil
}

// HandleGetDigitalInputs lists the alarm input lines.
func (s *Server) HandleGetDigitalInputs(_ *soap.RequestContext, _ []byte) (interface{}, error) {
	inputs := s.relays.DigitalInputs()
	resp := getDigitalInputsResponse{DigitalInputs: make([]digitalInputWire, 0, len(inputs))}
	for _, i := range inputs {
		resp.DigitalInputs = append(resp.DigitalInputs, digitalInputWire{
			Token:     i.Token,
			IdleState: i.IdleState,
		})
	}

	return resp, nil
}

// HandleGetDeviceIOServiceCapabilities reports the I/O counts (all audio
// counts zero — no audio hardware).
func (s *Server) HandleGetDeviceIOServiceCapabilities(_ *soap.RequestContext, _ []byte) (interface{}, error) {
	var resp deviceIOCapabilitiesWire
	resp.Capabilities.RelayOutputs = len(s.relays.RelayOutputs())
	resp.Capabilities.DigitalInputs = len(s.relays.DigitalInputs())

	return resp, nil
}
