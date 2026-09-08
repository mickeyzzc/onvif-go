package wsdiscovery

import (
	"testing"
)

// Fuzz targets pin the "hostile datagram never panics" invariant for the
// untrusted UDP multicast parse surface (issue #61). The seed corpus runs
// on every normal `go test`; extended fuzzing is `go test -fuzz`.

func FuzzParseProbe(f *testing.F) {
	f.Add([]byte(`<?xml version="1.0" encoding="UTF-8"?><s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing"><s:Header><a:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</a:Action><a:MessageID>urn:uuid:1</a:MessageID></s:Header><s:Body><Probe xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery"><Types>tds:Device</Types><Scopes/></Probe></s:Body></s:Envelope>`))
	f.Add([]byte("<Probe>"))
	f.Add([]byte{0x00, 0xff, 0xfe})
	f.Fuzz(func(t *testing.T, data []byte) {
		_ = ParseProbe(data)
	})
}

func FuzzParseProbeMatches(f *testing.F) {
	f.Add([]byte(`<?xml version="1.0" encoding="UTF-8"?><s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><ProbeMatches xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery"><ProbeMatch><wsa:EndpointReference xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"><wsa:Address>urn:uuid:1</wsa:Address></wsa:EndpointReference><Types>tds:Device</Types><Scopes>onvif://www.onvif.org/Profile/Streaming</Scopes><XAddrs>http://192.168.63.118:8080/onvif/device_service</XAddrs><MetadataVersion>1</MetadataVersion></ProbeMatch></ProbeMatches></s:Body></s:Envelope>`))
	f.Add([]byte("<ProbeMatches/>"))
	f.Add([]byte{0x81, 0x40, 0x30, 0x00})
	f.Fuzz(func(t *testing.T, data []byte) {
		matches, err := ParseProbeMatches(data)
		if err == nil && len(matches) == 0 {
			// A nil error promises usable matches; an empty slice would
			// defang ErrNoMatches checks at every call site.
			t.Error("ParseProbeMatches: nil error but zero matches")
		}
	})
}

func FuzzParseAnnouncement(f *testing.F) {
	f.Add([]byte(`<?xml version="1.0" encoding="UTF-8"?><s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><Hello xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery"><wsa:EndpointReference xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"><wsa:Address>urn:uuid:2</wsa:Address></wsa:EndpointReference><Scopes/></Hello></s:Body></s:Envelope>`))
	f.Add([]byte("<Bye/>"))
	f.Add([]byte("\r\n\r\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		// No-panic only: an empty <ProbeMatch/> element legitimately decodes
		// to an all-empty Match, so field-level invariants would false-positive.
		_ = ParseAnnouncement(data)
	})
}
