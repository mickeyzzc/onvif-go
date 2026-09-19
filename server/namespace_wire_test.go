package server

import (
	"encoding/xml"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// Issue #90 (server side): response children typed in ver10/schema were
// marshaled unprefixed, so they inherited the response root's service
// namespace (or none at all in explicit-prefix mode) — strict ONVIF
// clients cannot read the fields. These tests marshal representative
// responses and decode namespace-strictly: values only populate when the
// elements resolve to the exact namespace.

const schemaNS = "http://www.onvif.org/ver10/schema"

type nsUri struct {
	URI string `xml:"http://www.onvif.org/ver10/schema Uri"`
}

type nsSnapshotEnvelope struct {
	XMLName  xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetSnapshotUriResponse"`
	MediaUri nsUri    `xml:"http://www.onvif.org/ver10/media/wsdl MediaUri"`
}

func TestSnapshotUriResolvesToSchemaNamespace(t *testing.T) {
	data, err := xml.Marshal(&GetSnapshotUriResponse{MediaUri: MediaUri{URI: "http://cam/snap.jpg"}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var env nsSnapshotEnvelope
	if err := xml.Unmarshal(data, &env); err != nil {
		t.Fatalf("decode: %v\nxml: %s", err, data)
	}
	if env.MediaUri.URI != "http://cam/snap.jpg" {
		t.Errorf("Uri = %q, want %q (not in ver10/schema namespace)\nxml: %s",
			env.MediaUri.URI, "http://cam/snap.jpg", data)
	}
}

func TestDeviceInformationResolvesToSchemaNamespace(t *testing.T) {
	data, err := xml.Marshal(&GetDeviceInformationResponse{
		Manufacturer: "onvif-go",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var resp struct {
		XMLName      xml.Name        `xml:"http://www.onvif.org/ver10/device/wsdl GetDeviceInformationResponse"`
		Manufacturer string `xml:"http://www.onvif.org/ver10/schema Manufacturer"`
	}
	if err := xml.Unmarshal(data, &resp); err != nil {
		t.Fatalf("decode: %v\nxml: %s", err, data)
	}
	if resp.Manufacturer != "onvif-go" {
		t.Errorf("Manufacturer = %q (not in ver10/schema namespace)\nxml: %s", resp.Manufacturer, data)
	}
}

func TestSystemDateAndTimeResolvesToSchemaNamespace(t *testing.T) {
	data, err := xml.Marshal(&soap.GetSystemDateAndTimeResponse{
		SystemDateAndTime: soap.SystemDateAndTime{DateTimeType: "NTP"},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var resp struct {
		XMLName           xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetSystemDateAndTimeResponse"`
		SystemDateAndTime struct {
			DateTimeType string `xml:"http://www.onvif.org/ver10/schema DateTimeType"`
		} `xml:"http://www.onvif.org/ver10/schema SystemDateAndTime"`
	}
	if err := xml.Unmarshal(data, &resp); err != nil {
		t.Fatalf("decode: %v\nxml: %s", err, data)
	}
	// The strict decode only populates DateTimeType when SystemDateAndTime
	// itself resolved into ver10/schema; NTP (a legal value) proves both.
	if resp.SystemDateAndTime.DateTimeType == "" {
		t.Errorf("SystemDateAndTime/DateTimeType not in ver10/schema namespace\nxml: %s", data)
	}
}

type nsStatusEnvelope struct {
	XMLName   xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GetStatusResponse"`
	PTZStatus struct {
		Position struct {
			PanTilt struct {
				X float64 `xml:"x,attr"`
			} `xml:"http://www.onvif.org/ver10/schema PanTilt"`
		} `xml:"http://www.onvif.org/ver10/schema Position"`
	} `xml:"http://www.onvif.org/ver10/schema PTZStatus"`
}

func TestPTZStatusResolvesToSchemaNamespace(t *testing.T) {
	x := 0.5
	data, err := xml.Marshal(&GetStatusResponse{
		PTZStatus: &PTZStatus{
			Position: respPTZVector{PanTilt: &respVector2D{X: x, Y: 0}},
		},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var env nsStatusEnvelope
	if err := xml.Unmarshal(data, &env); err != nil {
		t.Fatalf("decode: %v\nxml: %s", err, data)
	}
	if env.PTZStatus.Position.PanTilt.X != 0.5 {
		t.Errorf("PTZStatus/Position/PanTilt not in ver10/schema namespace\nxml: %s", data)
	}
}

func TestImagingSettingsResolvesToSchemaNamespace(t *testing.T) {
	brightness := 60.0
	data, err := xml.Marshal(&GetImagingSettingsResponse{
		ImagingSettings: toRespImagingSettings(&ImagingSettings{Brightness: &brightness}),
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var resp struct {
		XMLName          xml.Name `xml:"http://www.onvif.org/ver20/imaging/wsdl GetImagingSettingsResponse"`
		ImagingSettings  struct {
			Brightness *float64 `xml:"http://www.onvif.org/ver10/schema Brightness"`
		} `xml:"http://www.onvif.org/ver20/imaging/wsdl ImagingSettings"`
	}
	if err := xml.Unmarshal(data, &resp); err != nil {
		t.Fatalf("decode: %v\nxml: %s", err, data)
	}
	if resp.ImagingSettings.Brightness == nil || *resp.ImagingSettings.Brightness != 60 {
		t.Errorf("ImagingSettings/Brightness not in ver10/schema namespace\nxml: %s", data)
	}
}
