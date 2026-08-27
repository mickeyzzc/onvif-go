package soap

import (
	"encoding/xml"
	"testing"
	"time"
)

func TestNormalizeAction(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"GetProfiles", "GetProfiles"},
		{"tds:GetProfiles", "GetProfiles"},
		{"a:b:c:Op", "Op"},
		{"", ""},
	}

	for _, tt := range tests {
		if got := NormalizeAction(tt.in); got != tt.want {
			t.Errorf("NormalizeAction(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParseRequestPaths(t *testing.T) {
	// Raw bytes (the transport form) decode directly.
	raw := []byte(`<GetStreamUri xmlns="http://www.onvif.org/ver10/media/wsdl"><ProfileToken>tok-1</ProfileToken></GetStreamUri>`)

	var req GetStreamUriRequest
	if err := ParseRequest(raw, &req); err != nil {
		t.Fatalf("ParseRequest raw: %v", err)
	}

	if req.ProfileToken != "tok-1" {
		t.Errorf("token = %q", req.ProfileToken)
	}

	// Struct input takes the marshal round-trip (direct handler calls).
	structReq := GetSnapshotUriRequest{ProfileToken: "tok-2"}
	var out GetSnapshotUriRequest
	if err := ParseRequest(&structReq, &out); err != nil {
		t.Fatalf("ParseRequest struct: %v", err)
	}

	if out.ProfileToken != "tok-2" {
		t.Errorf("round-trip token = %q", out.ProfileToken)
	}
}

func TestRequestWrapperInnerXML(t *testing.T) {
	var wrapper RequestWrapper
	if err := xml.Unmarshal([]byte(`<Op xmlns="ns"><A>1</A></Op>`), &wrapper); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if wrapper.XMLName.Local != "Op" {
		t.Errorf("element = %v", wrapper.XMLName)
	}

	if string(wrapper.Content) != "<A>1</A>" {
		t.Errorf("inner = %q", wrapper.Content)
	}
}

func TestToDateTime(t *testing.T) {
	reference := time.Date(2026, 8, 28, 12, 34, 56, 789000000, time.UTC)
	dt := ToDateTime(reference)

	if dt.Date.Year != 2026 || dt.Date.Month != 8 || dt.Date.Day != 28 {
		t.Errorf("date = %+v", dt.Date)
	}

	if dt.Time.Hour != 12 || dt.Time.Minute != 34 || dt.Time.Second != 56 {
		t.Errorf("time = %+v", dt.Time)
	}
}
