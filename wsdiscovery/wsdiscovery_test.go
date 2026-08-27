package wsdiscovery

import (
	"strings"
	"testing"
)

func TestProbeRoundTrip(t *testing.T) {
	probe := BuildProbe("11111111-2222-3333-4444-555555555555")

	parsed := ParseProbe(probe)
	if parsed == nil {
		t.Fatal("built probe not recognized")
	}

	if want := "11111111-2222-3333-4444-555555555555"; parsed.MessageID != want {
		t.Errorf("MessageID = %q, want %q", parsed.MessageID, want)
	}

	if len(parsed.Types) == 0 || !strings.Contains(parsed.Types[0], "NetworkVideoTransmitter") {
		t.Errorf("Types = %v, want NetworkVideoTransmitter", parsed.Types)
	}
}

func TestParseProbeRejectsNonProbes(t *testing.T) {
	hello := BuildHello(Match{EndpointRef: "urn:uuid:abc", Types: "tds:Device"})
	if ParseProbe(hello) != nil {
		t.Error("Hello must not parse as a Probe")
	}

	bye := BuildBye(Match{EndpointRef: "urn:uuid:abc"})
	if ParseProbe(bye) != nil {
		t.Error("Bye must not parse as a Probe")
	}

	if ParseProbe([]byte("not xml at all")) != nil {
		t.Error("garbage must not parse as a Probe")
	}
}

func TestProbeMatchesRoundTrip(t *testing.T) {
	match := Match{
		EndpointRef:     "urn:uuid:99999999-8888-7777-6666-555555555555",
		Types:           "tds:Device dp0:NetworkVideoTransmitter",
		Scopes:          "onvif://www.onvif.org/name/Test onvif://www.onvif.org/location/Roof",
		XAddrs:          "http://198.51.100.23:80/onvif/device_service",
		MetadataVersion: 1,
	}

	answer := BuildProbeMatches("probe-msg-id", match)

	matches, err := ParseProbeMatches(answer)
	if err != nil {
		t.Fatalf("ParseProbeMatches: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("matches = %d, want 1", len(matches))
	}

	got := matches[0]
	if got.EndpointRef != match.EndpointRef || got.XAddrs != match.XAddrs ||
		got.Types != match.Types || got.Scopes != match.Scopes || got.MetadataVersion != 1 {
		t.Errorf("round-trip mismatch:\n got %+v\nwant %+v", got, match)
	}

	// The RelatesTo correlation header must reference the probe.
	if !strings.Contains(string(answer), "uuid:probe-msg-id") {
		t.Error("answer missing RelatesTo correlation")
	}
}

func TestAnnouncementParsing(t *testing.T) {
	hello := BuildHello(Match{
		EndpointRef:     "urn:uuid:hello-1",
		Types:           "tds:Device",
		XAddrs:          "http://198.51.100.24/onvif/device_service",
		MetadataVersion: 1,
	})

	got := ParseAnnouncement(hello)
	if got == nil {
		t.Fatal("Hello not parsed")
	}

	if got.EndpointRef != "urn:uuid:hello-1" {
		t.Errorf("EndpointRef = %q", got.EndpointRef)
	}

	// ProbeMatches also parse as announcements (client listener path).
	answer := BuildProbeMatches("m", Match{EndpointRef: "urn:uuid:pm-1"})
	if pm := ParseAnnouncement(answer); pm == nil || pm.EndpointRef != "urn:uuid:pm-1" {
		t.Errorf("ProbeMatches announcement not parsed: %+v", pm)
	}

	// Bye is not an announcement.
	if bye := ParseAnnouncement(BuildBye(Match{EndpointRef: "urn:uuid:x"})); bye != nil {
		t.Error("Bye must not parse as an announcement")
	}
}

func TestMatchesTypesFilter(t *testing.T) {
	deviceTypes := []string{"tds:Device", "dp0:NetworkVideoTransmitter"}

	tests := []struct {
		name  string
		probe *Probe
		want  bool
	}{
		{name: "empty filter matches", probe: &Probe{MessageID: "m"}, want: true},
		{name: "exact match", probe: &Probe{MessageID: "m", Types: []string{"NetworkVideoTransmitter"}}, want: true},
		{name: "prefixed match", probe: &Probe{MessageID: "m", Types: []string{"dp0:NetworkVideoTransmitter"}}, want: true},
		{name: "unrelated type", probe: &Probe{MessageID: "m", Types: []string{"dn:NetworkPrinter"}}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.probe.MatchesTypes(deviceTypes); got != tt.want {
				t.Errorf("MatchesTypes = %v, want %v", got, tt.want)
			}
		})
	}
}
