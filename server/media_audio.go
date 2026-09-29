package server

import (
	"encoding/xml"

	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// Media audio configuration family (issue #115): this virtual camera has
// no audio hardware, so every enumeration answers a valid EMPTY set —
// clients get a real response to walk, not an ActionNotSupported fault
// they cannot distinguish from "device broken".

type getAudioSourcesResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetAudioSourcesResponse"`
}

type getAudioSourceConfigurationsResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetAudioSourceConfigurationsResponse"`
}

type getAudioEncoderConfigurationsResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetAudioEncoderConfigurationsResponse"`
}

type getAudioOutputsResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetAudioOutputsResponse"`
}

type getAudioDecoderConfigurationsResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetAudioDecoderConfigurationsResponse"`
}

// HandleGetAudioSources answers the (empty) audio source list.
func (s *Server) HandleGetAudioSources(_ *soap.RequestContext, _ []byte) (interface{}, error) {
	return getAudioSourcesResponse{}, nil
}

// HandleGetAudioSourceConfigurations answers the empty audio source
// configuration list.
func (s *Server) HandleGetAudioSourceConfigurations(_ *soap.RequestContext, _ []byte) (interface{}, error) {
	return getAudioSourceConfigurationsResponse{}, nil
}

// HandleGetAudioEncoderConfigurations answers the empty audio encoder
// configuration list.
func (s *Server) HandleGetAudioEncoderConfigurations(_ *soap.RequestContext, _ []byte) (interface{}, error) {
	return getAudioEncoderConfigurationsResponse{}, nil
}

// HandleGetAudioOutputs answers the (empty) audio output list.
func (s *Server) HandleGetAudioOutputs(_ *soap.RequestContext, _ []byte) (interface{}, error) {
	return getAudioOutputsResponse{}, nil
}

// HandleGetAudioDecoderConfigurations answers the empty audio decoder
// configuration list.
func (s *Server) HandleGetAudioDecoderConfigurations(_ *soap.RequestContext, _ []byte) (interface{}, error) {
	return getAudioDecoderConfigurationsResponse{}, nil
}
