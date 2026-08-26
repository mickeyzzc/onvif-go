package onvif

import (
	"context"
	"encoding/xml"
	"fmt"
)

// Request/response types hoisted from method bodies.

type GetSnapshotURI struct {
	XMLName      xml.Name `xml:"trt:GetSnapshotUri"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type GetSnapshotURIResponse struct {
	XMLName  xml.Name `xml:"GetSnapshotUriResponse"`
	MediaURI struct {
		URI                 string `xml:"Uri"`
		InvalidAfterConnect bool   `xml:"InvalidAfterConnect"`
		InvalidAfterReboot  bool   `xml:"InvalidAfterReboot"`
		Timeout             string `xml:"Timeout"`
	} `xml:"MediaUri"`
}

type GetStreamURI struct {
	XMLName     xml.Name `xml:"trt:GetStreamUri"`
	Xmlns       string   `xml:"xmlns:trt,attr"`
	Xmlnst      string   `xml:"xmlns:tt,attr"`
	StreamSetup struct {
		Stream    string `xml:"tt:Stream"`
		Transport struct {
			Protocol string `xml:"tt:Protocol"`
		} `xml:"tt:Transport"`
	} `xml:"trt:StreamSetup"`
	ProfileToken string `xml:"trt:ProfileToken"`
}

type GetStreamURIResponse struct {
	XMLName  xml.Name `xml:"GetStreamUriResponse"`
	MediaURI struct {
		URI                 string `xml:"Uri"`
		InvalidAfterConnect bool   `xml:"InvalidAfterConnect"`
		InvalidAfterReboot  bool   `xml:"InvalidAfterReboot"`
		Timeout             string `xml:"Timeout"`
	} `xml:"MediaUri"`
}

type SetSynchronizationPoint struct {
	XMLName      xml.Name `xml:"trt:SetSynchronizationPoint"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type StartMulticastStreaming struct {
	XMLName      xml.Name `xml:"trt:StartMulticastStreaming"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

type StopMulticastStreaming struct {
	XMLName      xml.Name `xml:"trt:StopMulticastStreaming"`
	Xmlns        string   `xml:"xmlns:trt,attr"`
	ProfileToken string   `xml:"trt:ProfileToken"`
}

func (s *MediaService) GetStreamURI(ctx context.Context, profileToken string) (*MediaURI, error) {
	endpoint := s.getMediaEndpoint()

	req := GetStreamURI{
		Xmlns:        mediaNamespace,
		Xmlnst:       "http://www.onvif.org/ver10/schema",
		ProfileToken: profileToken,
	}
	req.StreamSetup.Stream = "RTP-Unicast"
	req.StreamSetup.Transport.Protocol = "RTSP"

	var resp GetStreamURIResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetStreamURI failed: %w", err)
	}

	return &MediaURI{
		URI:                 resp.MediaURI.URI,
		InvalidAfterConnect: resp.MediaURI.InvalidAfterConnect,
		InvalidAfterReboot:  resp.MediaURI.InvalidAfterReboot,
	}, nil
}

func (s *MediaService) GetSnapshotURI(ctx context.Context, profileToken string) (*MediaURI, error) {
	endpoint := s.getMediaEndpoint()

	req := GetSnapshotURI{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	var resp GetSnapshotURIResponse

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetSnapshotURI failed: %w", err)
	}

	return &MediaURI{
		URI:                 resp.MediaURI.URI,
		InvalidAfterConnect: resp.MediaURI.InvalidAfterConnect,
		InvalidAfterReboot:  resp.MediaURI.InvalidAfterReboot,
	}, nil
}

func (s *MediaService) SetSynchronizationPoint(ctx context.Context, profileToken string) error {
	endpoint := s.getMediaEndpoint()

	req := SetSynchronizationPoint{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("SetSynchronizationPoint failed: %w", err)
	}

	return nil
}

func (s *MediaService) StartMulticastStreaming(ctx context.Context, profileToken string) error {
	endpoint := s.getMediaEndpoint()

	req := StartMulticastStreaming{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("StartMulticastStreaming failed: %w", err)
	}

	return nil
}

func (s *MediaService) StopMulticastStreaming(ctx context.Context, profileToken string) error {
	endpoint := s.getMediaEndpoint()

	req := StopMulticastStreaming{
		Xmlns:        mediaNamespace,
		ProfileToken: profileToken,
	}

	username, password := s.client.GetCredentials()
	soapClient := s.client.newSoapClient(username, password)

	if err := soapClient.Call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("StopMulticastStreaming failed: %w", err)
	}

	return nil
}
