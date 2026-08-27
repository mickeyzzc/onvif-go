package onvif

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
)

// ErrEmptyMediaURI is returned when a GetStreamUri/GetSnapshotUri response
// parsed successfully but contained no usable Uri element. Previously such
// responses silently produced an empty URI with a nil error — the hardest
// failure mode to debug downstream (issue #3).
var ErrEmptyMediaURI = errors.New("device returned an empty media URI")

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
	// InnerXML captures the raw response content so the loose URI-extraction
	// fallback and diagnostic errors can work with what the device actually
	// sent, whatever its namespace prefixes (issue #3).
	InnerXML string `xml:",innerxml"`
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
	// InnerXML captures the raw response content so the loose URI-extraction
	// fallback and diagnostic errors can work with what the device actually
	// sent, whatever its namespace prefixes (issue #3).
	InnerXML string `xml:",innerxml"`
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

// maxMediaURIErrBody caps how much raw response XML is embedded in
// ErrEmptyMediaURI errors.
const maxMediaURIErrBody = 512

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

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetStreamURI failed: %w", err)
	}

	uri := resp.MediaURI.URI
	if uri == "" {
		uri = looseExtractURI(resp.InnerXML)
	}

	if uri == "" {
		return nil, fmt.Errorf("%w: GetStreamUri (profile %s) response carried no Uri element; body: %s",
			ErrEmptyMediaURI, profileToken, truncateForError(resp.InnerXML, maxMediaURIErrBody))
	}

	return &MediaURI{
		URI:                 uri,
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

	if err := s.client.call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetSnapshotURI failed: %w", err)
	}

	// GetSnapshotUri shares the GetStreamUri response shape and therefore the
	// same namespace-variant parsing pitfalls (issue #3 audit conclusion).
	uri := resp.MediaURI.URI
	if uri == "" {
		uri = looseExtractURI(resp.InnerXML)
	}

	if uri == "" {
		return nil, fmt.Errorf("%w: GetSnapshotUri (profile %s) response carried no Uri element; body: %s",
			ErrEmptyMediaURI, profileToken, truncateForError(resp.InnerXML, maxMediaURIErrBody))
	}

	return &MediaURI{
		URI:                 uri,
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

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
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

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
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

	if err := s.client.call(ctx, endpoint, "", req, nil); err != nil {
		return fmt.Errorf("StopMulticastStreaming failed: %w", err)
	}

	return nil
}

// looseExtractURI scans raw response XML for the first element whose local
// name is "Uri" (any namespace, any prefix) and returns its text. It is the
// last-resort fallback for devices whose response shape defeats the typed
// structs: unusual namespace bindings, SOAP 1.1 envelopes, or a missing
// MediaUri wrapper element. Empty string when nothing is found.
func looseExtractURI(raw string) string {
	dec := xml.NewDecoder(strings.NewReader(raw))
	for {
		tok, err := dec.Token()
		if err != nil {
			return ""
		}

		start, ok := tok.(xml.StartElement)
		if !ok || start.Name.Local != "Uri" {
			continue
		}

		var value string
		if err := dec.DecodeElement(&value, &start); err != nil {
			return ""
		}

		return strings.TrimSpace(value)
	}
}

// truncateForError shortens s to at most limit bytes for error messages.
func truncateForError(s string, limit int) string {
	if len(s) <= limit {
		return s
	}

	return s[:limit] + "...(truncated)"
}
