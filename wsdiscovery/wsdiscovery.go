// Package wsdiscovery is the shared WS-Discovery wire codec: Probe,
// ProbeMatches, Hello, and Bye messages for both sides of the protocol.
// The client (discovery package) uses it to build probes and parse
// announcements; the device side (server/discovery) uses it to parse
// probes and build answers. One codec, no private parsing drift.
package wsdiscovery

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrNoMatches is returned by ParseProbeMatches when a response carries
// no ProbeMatch element.
var ErrNoMatches = errors.New("no ProbeMatch in response")

// Protocol constants.
const (
	// Namespace is the WS-Discovery XML namespace.
	Namespace = "http://schemas.xmlsoap.org/ws/2005/04/discovery"

	// MulticastAddr is the WS-Discovery multicast group (UDP).
	MulticastAddr = "239.255.255.250:3702"
)

// Match is the announcement record shared by Hello and ProbeMatch
// messages. There is deliberately no XMLName constraint: the envelope
// parser dispatches on the containing element.
type Match struct {
	EndpointRef     string `xml:"EndpointReference>Address"`
	Types           string `xml:"Types"`
	Scopes          string `xml:"Scopes"`
	XAddrs          string `xml:"XAddrs"`
	MetadataVersion int    `xml:"MetadataVersion"`
}

// announcementEnvelope is the decode shape of Hello and ProbeMatches.
type announcementEnvelope struct {
	Body struct {
		Hello        *Match `xml:"Hello"`
		ProbeMatches struct {
			ProbeMatch []Match `xml:"ProbeMatch"`
		} `xml:"ProbeMatches"`
	} `xml:"Body"`
}

// ParseAnnouncement extracts the first device announcement (Hello or
// ProbeMatch) from a datagram. Returns nil for Bye messages, probe
// acknowledgements, and anything unparseable — listeners must survive
// garbage on the wire.
func ParseAnnouncement(data []byte) *Match {
	var envelope announcementEnvelope
	if err := xml.Unmarshal(data, &envelope); err != nil {
		return nil
	}

	switch {
	case envelope.Body.Hello != nil && envelope.Body.Hello.EndpointRef != "":
		return envelope.Body.Hello
	case len(envelope.Body.ProbeMatches.ProbeMatch) > 0:
		first := envelope.Body.ProbeMatches.ProbeMatch[0]

		return &first
	default:
		return nil
	}
}

// ParseProbeMatches extracts every match from a ProbeMatches response.
func ParseProbeMatches(data []byte) ([]Match, error) {
	var envelope announcementEnvelope
	if err := xml.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("failed to unmarshal probe response: %w", err)
	}

	if len(envelope.Body.ProbeMatches.ProbeMatch) == 0 {
		return nil, ErrNoMatches
	}

	return envelope.Body.ProbeMatches.ProbeMatch, nil
}

// probeEnvelope decodes the parts of a Probe a responder needs. The
// Probe field is a pointer so an absent body distinguishes "not a
// probe" from "probe with no filters".
type probeEnvelope struct {
	Header struct {
		Action    string `xml:"http://schemas.xmlsoap.org/ws/2004/08/addressing Action"`
		MessageID string `xml:"http://schemas.xmlsoap.org/ws/2004/08/addressing MessageID"`
	} `xml:"Header"`
	Body struct {
		Probe *struct {
			Types  string `xml:"Types"`
			Scopes string `xml:"Scopes"`
		} `xml:"Probe"`
	} `xml:"Body"`
}

// Probe is a decoded WS-Discovery Probe request.
type Probe struct {
	MessageID string
	Types     []string
}

// ParseProbe decodes a Probe datagram (nil when it is not a probe).
func ParseProbe(data []byte) *Probe {
	var envelope probeEnvelope
	if err := xml.Unmarshal(data, &envelope); err != nil {
		return nil
	}

	if envelope.Body.Probe == nil {
		return nil
	}

	return &Probe{
		MessageID: trimUUIDScheme(envelope.Header.MessageID),
		Types:     parseQNameList(envelope.Body.Probe.Types),
	}
}

// MatchesTypes reports whether a probe's Types filter selects a device
// advertising the given types: an empty filter matches everything;
// otherwise at least one probe qname must equal one of ours (namespace
// prefixes are ignored — devices disagree on them).
func (p *Probe) MatchesTypes(deviceTypes []string) bool {
	if len(p.Types) == 0 {
		return true
	}

	ours := make(map[string]bool, len(deviceTypes))
	for _, t := range deviceTypes {
		ours[localName(t)] = true
	}

	for _, t := range p.Types {
		if ours[localName(t)] {
			return true
		}
	}

	return false
}

// BuildProbe builds the client probe envelope ONVIF devices answer to
// (Types = NetworkVideoTransmitter).
func BuildProbe(messageID string) []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing">
<s:Header>
<a:Action s:mustUnderstand="1">` + Namespace + `/Probe</a:Action>
<a:MessageID>uuid:` + messageID + `</a:MessageID>
<a:ReplyTo><a:Address>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address></a:ReplyTo>
<a:To s:mustUnderstand="1">urn:schemas-xmlsoap-org:ws:2005:04:discovery</a:To>
</s:Header>
<s:Body>
<Probe xmlns="` + Namespace + `">
<d:Types xmlns:d="` + Namespace + `" xmlns:dp0="http://www.onvif.org/ver10/network/wsdl">dp0:NetworkVideoTransmitter</d:Types>
</Probe>
</s:Body>
</s:Envelope>`)
}

// BuildProbeMatches builds the device-side answer to a Probe; relatesTo
// must be the probe's MessageID.
func BuildProbeMatches(relatesTo string, match Match) []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing">
<s:Header>
<a:Action s:mustUnderstand="1">` + Namespace + `/ProbeMatches</a:Action>
<a:RelatesTo>uuid:` + relatesTo + `</a:RelatesTo>
</s:Header>
<s:Body>
<ProbeMatches xmlns="` + Namespace + `">
<ProbeMatch>
<a:EndpointReference><a:Address>` + match.EndpointRef + `</a:Address></a:EndpointReference>
<Types>` + match.Types + `</Types>
<Scopes>` + match.Scopes + `</Scopes>
<XAddrs>` + match.XAddrs + `</XAddrs>
<MetadataVersion>` + strconv.Itoa(match.MetadataVersion) + `</MetadataVersion>
</ProbeMatch>
</ProbeMatches>
</s:Body>
</s:Envelope>`)
}

// BuildHello builds the device-side power-on announcement.
func BuildHello(match Match) []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing">
<s:Header>
<a:Action s:mustUnderstand="1">` + Namespace + `/Hello</a:Action>
</s:Header>
<s:Body>
<Hello xmlns="` + Namespace + `">
<a:EndpointReference><a:Address>` + match.EndpointRef + `</a:Address></a:EndpointReference>
<Types>` + match.Types + `</Types>
<Scopes>` + match.Scopes + `</Scopes>
<XAddrs>` + match.XAddrs + `</XAddrs>
<MetadataVersion>` + strconv.Itoa(match.MetadataVersion) + `</MetadataVersion>
</Hello>
</s:Body>
</s:Envelope>`)
}

// BuildBye builds the device-side shutdown announcement.
func BuildBye(match Match) []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing">
<s:Header>
<a:Action s:mustUnderstand="1">` + Namespace + `/Bye</a:Action>
</s:Header>
<s:Body>
<Bye xmlns="` + Namespace + `">
<a:EndpointReference><a:Address>` + match.EndpointRef + `</a:Address></a:EndpointReference>
</Bye>
</s:Body>
</s:Envelope>`)
}

// parseQNameList splits a space-separated list of qnames.
func parseQNameList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	return strings.Fields(s)
}

// trimUUIDScheme strips the uuid:/urn:uuid: scheme prefix devices
// variously use in MessageID/RelatesTo values.
func trimUUIDScheme(id string) string {
	id = strings.TrimPrefix(id, "urn:uuid:")
	id = strings.TrimPrefix(id, "uuid:")

	return id
}

// localName strips a qname's namespace prefix.
func localName(qname string) string {
	if idx := strings.LastIndex(qname, ":"); idx != -1 {
		return qname[idx+1:]
	}

	return qname
}
