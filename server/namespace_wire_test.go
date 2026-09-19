package server

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// WSDL-grounded namespace contract suite (follow-up to #90).
//
// Ground truth is the official ONVIF WSDL/XSD set (github.com/onvif/specs,
// every schema elementFormDefault="qualified", attributes unqualified):
//
//   - Elements locally declared inside a service WSDL resolve to that
//     WSDL's namespace — including the direct children of *Response
//     wrappers (Manufacturer is an inline xs:string in the device WSDL,
//     so it is tds, not tt).
//   - Children of ver10/schema types (tt:Capabilities, tt:Profile,
//     tt:PTZStatus, …) resolve to ver10/schema.
//   - WS-BaseNotification children (CurrentTime, NotificationMessage, …)
//     resolve to the wsn b-2 namespace; endpoint references to wsa.
//
// Each test marshals a served response and decodes it namespace-strictly:
// values only populate when every element resolves to the exact namespace
// the WSDL assigns. A wrong namespace (or no namespace at all, the shape
// unprefixed tags produce in explicit-prefix mode) leaves zero values and
// fails the assertion.

const (
	nsDeviceWSDL  = "http://www.onvif.org/ver10/device/wsdl"
	nsMediaWSDL   = "http://www.onvif.org/ver10/media/wsdl"
	nsEventsWSDL  = "http://www.onvif.org/ver10/events/wsdl"
	nsPTZWSDL     = "http://www.onvif.org/ver20/ptz/wsdl"
	nsImagingWSDL = "http://www.onvif.org/ver20/imaging/wsdl"
	nsSchema      = "http://www.onvif.org/ver10/schema"
	nsWSNT        = "http://docs.oasis-open.org/wsn/b-2"
	nsWSA         = "http://www.w3.org/2005/08/addressing"
)

func marshalOrFail(t *testing.T, v interface{}) []byte {
	t.Helper()

	data, err := xml.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	return data
}

func decodeOrFail(t *testing.T, data []byte, v interface{}) {
	t.Helper()

	if err := xml.Unmarshal(data, v); err != nil {
		t.Fatalf("decode: %v\nxml: %s", err, data)
	}
}

func TestDeviceInformationWireNamespaces(t *testing.T) {
	data := marshalOrFail(t, &GetDeviceInformationResponse{
		Manufacturer: "onvif-go", SerialNumber: "SN-1",
	})

	var resp struct {
		XMLName      xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetDeviceInformationResponse"`
		Manufacturer string   `xml:"http://www.onvif.org/ver10/device/wsdl Manufacturer"`
		SerialNumber string   `xml:"http://www.onvif.org/ver10/device/wsdl SerialNumber"`
	}
	decodeOrFail(t, data, &resp)

	if resp.Manufacturer != "onvif-go" || resp.SerialNumber != "SN-1" {
		t.Errorf("Manufacturer/SerialNumber not in %s:\nxml: %s", nsDeviceWSDL, data)
	}
}

func TestCapabilitiesWireNamespaces(t *testing.T) {
	data := marshalOrFail(t, &GetCapabilitiesResponse{
		Capabilities: &Capabilities{
			Device: &DeviceCapabilities{
				XAddr:   "http://cam/device_service",
				Network: &NetworkCapabilities{IPVersion6: true},
			},
			Media: &MediaCapabilities{
				XAddr:                 "http://cam/media_service",
				StreamingCapabilities: &StreamingCapabilities{RTPTCP: true},
			},
			Events:  &EventCapabilities{XAddr: "http://cam/events_service"},
			Imaging: &ImagingCapabilities{XAddr: "http://cam/imaging_service"},
			PTZ:     &PTZCapabilities{XAddr: "http://cam/ptz_service"},
		},
	})

	var resp struct {
		XMLName      xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetCapabilitiesResponse"`
		Capabilities struct {
			XMLName xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl Capabilities"`
			Device  struct {
				XMLName xml.Name `xml:"http://www.onvif.org/ver10/schema Device"`
				XAddr   string   `xml:"http://www.onvif.org/ver10/schema XAddr"`
				Network struct {
					XMLName    xml.Name `xml:"http://www.onvif.org/ver10/schema Network"`
					IPVersion6 bool     `xml:"IPVersion6,attr"`
				} `xml:"http://www.onvif.org/ver10/schema Network"`
			} `xml:"http://www.onvif.org/ver10/schema Device"`
			Media struct {
				XMLName               xml.Name `xml:"http://www.onvif.org/ver10/schema Media"`
				XAddr                 string   `xml:"http://www.onvif.org/ver10/schema XAddr"`
				StreamingCapabilities struct {
					XMLName xml.Name `xml:"http://www.onvif.org/ver10/schema StreamingCapabilities"`
					RTPTCP  bool     `xml:"RTP_TCP,attr"`
				} `xml:"http://www.onvif.org/ver10/schema StreamingCapabilities"`
			} `xml:"http://www.onvif.org/ver10/schema Media"`
			Events struct {
				XMLName xml.Name `xml:"http://www.onvif.org/ver10/schema Events"`
				XAddr   string   `xml:"http://www.onvif.org/ver10/schema XAddr"`
			} `xml:"http://www.onvif.org/ver10/schema Events"`
			Imaging struct {
				XMLName xml.Name `xml:"http://www.onvif.org/ver10/schema Imaging"`
				XAddr   string   `xml:"http://www.onvif.org/ver10/schema XAddr"`
			} `xml:"http://www.onvif.org/ver10/schema Imaging"`
			PTZ struct {
				XMLName xml.Name `xml:"http://www.onvif.org/ver10/schema PTZ"`
				XAddr   string   `xml:"http://www.onvif.org/ver10/schema XAddr"`
			} `xml:"http://www.onvif.org/ver10/schema PTZ"`
		} `xml:"http://www.onvif.org/ver10/device/wsdl Capabilities"`
	}
	decodeOrFail(t, data, &resp)

	caps := resp.Capabilities
	switch {
	case caps.XMLName.Local != "Capabilities":
		t.Errorf("Capabilities element = %q, want the %s namespace", caps.XMLName.Local, nsDeviceWSDL)
	case caps.Device.XAddr == "":
		t.Errorf("Device/XAddr not in %s:\nxml: %s", nsSchema, data)
	case !caps.Device.Network.IPVersion6:
		t.Errorf("Device/Network not in %s:\nxml: %s", nsSchema, data)
	case caps.Media.XAddr == "":
		t.Errorf("Media/XAddr not in %s:\nxml: %s", nsSchema, data)
	case !caps.Media.StreamingCapabilities.RTPTCP:
		t.Errorf("Media/StreamingCapabilities not in %s:\nxml: %s", nsSchema, data)
	case caps.Events.XAddr == "":
		t.Errorf("Events not in %s:\nxml: %s", nsSchema, data)
	case caps.Imaging.XAddr == "":
		t.Errorf("Imaging not in %s:\nxml: %s", nsSchema, data)
	case caps.PTZ.XAddr == "":
		t.Errorf("PTZ not in %s:\nxml: %s", nsSchema, data)
	}
}

func TestSystemDateAndTimeWireNamespaces(t *testing.T) {
	data := marshalOrFail(t, &soap.GetSystemDateAndTimeResponse{
		SystemDateAndTime: soap.SystemDateAndTime{
			DateTimeType: "NTP",
			TimeZone:     soap.TimeZone{TZ: "UTC"},
			UTCDateTime:  soap.DateTime{Time: soap.Time{Hour: 7}, Date: soap.Date{Year: 2026}},
		},
	})

	var resp struct {
		XMLName           xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetSystemDateAndTimeResponse"`
		SystemDateAndTime struct {
			XMLName      xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl SystemDateAndTime"`
			DateTimeType string   `xml:"http://www.onvif.org/ver10/schema DateTimeType"`
			TimeZone     struct {
				TZ string `xml:"http://www.onvif.org/ver10/schema TZ"`
			} `xml:"http://www.onvif.org/ver10/schema TimeZone"`
			UTCDateTime struct {
				Date struct {
					Year int `xml:"http://www.onvif.org/ver10/schema Year"`
				} `xml:"http://www.onvif.org/ver10/schema Date"`
				Time struct {
					Hour int `xml:"http://www.onvif.org/ver10/schema Hour"`
				} `xml:"http://www.onvif.org/ver10/schema Time"`
			} `xml:"http://www.onvif.org/ver10/schema UTCDateTime"`
		} `xml:"http://www.onvif.org/ver10/device/wsdl SystemDateAndTime"`
	}
	decodeOrFail(t, data, &resp)

	sdt := resp.SystemDateAndTime
	if sdt.XMLName.Local != "SystemDateAndTime" ||
		sdt.DateTimeType != "NTP" || sdt.TimeZone.TZ != "UTC" ||
		sdt.UTCDateTime.Date.Year != 2026 || sdt.UTCDateTime.Time.Hour != 7 {
		t.Errorf("SystemDateAndTime tree not in WSDL namespaces:\nxml: %s", data)
	}
}

func TestServicesWireNamespaces(t *testing.T) {
	data := marshalOrFail(t, &GetServicesResponse{
		Service: []Service{{
			Namespace: nsDeviceWSDL,
			XAddr:     "http://cam/device_service",
			Version:   Version{Major: 2, Minor: 5},
		}},
	})

	var resp struct {
		XMLName xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetServicesResponse"`
		Service []struct {
			Namespace string `xml:"http://www.onvif.org/ver10/device/wsdl Namespace"`
			XAddr     string `xml:"http://www.onvif.org/ver10/device/wsdl XAddr"`
			Version   struct {
				Major int `xml:"http://www.onvif.org/ver10/schema Major"`
				Minor int `xml:"http://www.onvif.org/ver10/schema Minor"`
			} `xml:"http://www.onvif.org/ver10/device/wsdl Version"`
		} `xml:"http://www.onvif.org/ver10/device/wsdl Service"`
	}
	decodeOrFail(t, data, &resp)

	if len(resp.Service) != 1 {
		t.Fatalf("Service count = %d, want 1 (element in %s)\nxml: %s", len(resp.Service), nsDeviceWSDL, data)
	}

	svc := resp.Service[0]
	if svc.Namespace == "" || svc.XAddr == "" || svc.Version.Major != 2 || svc.Version.Minor != 5 {
		t.Errorf("Service children not in WSDL namespaces:\nxml: %s", data)
	}
}

func TestSystemRebootWireNamespaces(t *testing.T) {
	data := marshalOrFail(t, &SystemRebootResponse{Message: "Device rebooting"})

	var resp struct {
		XMLName xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl SystemRebootResponse"`
		Message string   `xml:"http://www.onvif.org/ver10/device/wsdl Message"`
	}
	decodeOrFail(t, data, &resp)

	if resp.Message == "" {
		t.Errorf("Message not in %s:\nxml: %s", nsDeviceWSDL, data)
	}
}

func TestScopesWireNamespaces(t *testing.T) {
	data := marshalOrFail(t, &GetScopesResponse{
		Scopes: []ScopeDefinition{{ScopeDef: "Fixed", ScopeItem: "onvif://www.onvif.org/type/video_encoder"}},
	})

	var resp struct {
		XMLName xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetScopesResponse"`
		Scopes  []struct {
			ScopeDef  string `xml:"http://www.onvif.org/ver10/schema ScopeDef"`
			ScopeItem string `xml:"http://www.onvif.org/ver10/schema ScopeItem"`
		} `xml:"http://www.onvif.org/ver10/device/wsdl Scopes"`
	}
	decodeOrFail(t, data, &resp)

	if len(resp.Scopes) != 1 ||
		resp.Scopes[0].ScopeDef != "Fixed" ||
		resp.Scopes[0].ScopeItem != "onvif://www.onvif.org/type/video_encoder" {
		t.Errorf("scope not in the tt:Scope shape (ScopeDef/ScopeItem elements):\nxml: %s", data)
	}
}

func TestProfilesWireNamespaces(t *testing.T) {
	data := marshalOrFail(t, &GetProfilesResponse{
		Profiles: []MediaProfile{{
			Token: "profile_1", Fixed: true, Name: "Main",
			VideoSourceConfiguration: &VideoSourceConfiguration{
				Token: "src_1", Name: "Source", UseCount: 1, SourceToken: "src_1",
				Bounds: IntRectangle{Width: 1920, Height: 1080},
			},
			VideoEncoderConfiguration: &VideoEncoderConfiguration{
				Token: "enc_1", Name: "Encoder", UseCount: 1, Encoding: "H264",
				Resolution:  VideoResolution{Width: 1920, Height: 1080},
				Quality:     4,
				RateControl: &VideoRateControl{FrameRateLimit: 15, EncodingInterval: 1, BitrateLimit: 2500},
				H264:        &H264Configuration{GovLength: 30, H264Profile: "Main"},
			},
			PTZConfiguration:      &PTZConfiguration{Token: "ptz_1", Name: "PTZ", NodeToken: "node_1"},
			MetadataConfiguration: &MetadataConfiguration{Token: "meta_1", Name: "Metadata"},
		}},
	})

	var resp struct {
		XMLName  xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetProfilesResponse"`
		Profiles []struct {
			XMLName                  string `xml:"http://www.onvif.org/ver10/media/wsdl Profiles"`
			Name                     string `xml:"http://www.onvif.org/ver10/schema Name"`
			VideoSourceConfiguration struct {
				Name        string `xml:"http://www.onvif.org/ver10/schema Name"`
				SourceToken string `xml:"http://www.onvif.org/ver10/schema SourceToken"`
				Bounds      struct {
					Width int `xml:"width,attr"`
				} `xml:"http://www.onvif.org/ver10/schema Bounds"`
			} `xml:"http://www.onvif.org/ver10/schema VideoSourceConfiguration"`
			VideoEncoderConfiguration struct {
				Encoding   string `xml:"http://www.onvif.org/ver10/schema Encoding"`
				Resolution struct {
					Width int `xml:"http://www.onvif.org/ver10/schema Width"`
				} `xml:"http://www.onvif.org/ver10/schema Resolution"`
				RateControl struct {
					FrameRateLimit int `xml:"http://www.onvif.org/ver10/schema FrameRateLimit"`
				} `xml:"http://www.onvif.org/ver10/schema RateControl"`
				H264 struct {
					GovLength int `xml:"http://www.onvif.org/ver10/schema GovLength"`
				} `xml:"http://www.onvif.org/ver10/schema H264"`
			} `xml:"http://www.onvif.org/ver10/schema VideoEncoderConfiguration"`
			PTZConfiguration struct {
				NodeToken string `xml:"http://www.onvif.org/ver10/schema NodeToken"`
			} `xml:"http://www.onvif.org/ver10/schema PTZConfiguration"`
			MetadataConfiguration struct {
				Name string `xml:"http://www.onvif.org/ver10/schema Name"`
			} `xml:"http://www.onvif.org/ver10/schema MetadataConfiguration"`
		} `xml:"http://www.onvif.org/ver10/media/wsdl Profiles"`
	}
	decodeOrFail(t, data, &resp)

	if len(resp.Profiles) != 1 {
		t.Fatalf("Profiles count = %d, want 1 (element in %s)\nxml: %s", len(resp.Profiles), nsMediaWSDL, data)
	}

	p := resp.Profiles[0]
	switch {
	case p.Name != "Main":
		t.Errorf("Profile/Name not in %s:\nxml: %s", nsSchema, data)
	case p.VideoSourceConfiguration.SourceToken == "":
		t.Errorf("VideoSourceConfiguration children not in %s:\nxml: %s", nsSchema, data)
	case p.VideoEncoderConfiguration.Encoding != "H264":
		t.Errorf("VideoEncoderConfiguration/Encoding not in %s:\nxml: %s", nsSchema, data)
	case p.VideoEncoderConfiguration.Resolution.Width != 1920:
		t.Errorf("VideoEncoderConfiguration/Resolution not in %s:\nxml: %s", nsSchema, data)
	case p.VideoEncoderConfiguration.RateControl.FrameRateLimit != 15:
		t.Errorf("VideoEncoderConfiguration/RateControl not in %s:\nxml: %s", nsSchema, data)
	case p.VideoEncoderConfiguration.H264.GovLength != 30:
		t.Errorf("VideoEncoderConfiguration/H264 not in %s:\nxml: %s", nsSchema, data)
	case p.PTZConfiguration.NodeToken == "":
		t.Errorf("PTZConfiguration not in %s:\nxml: %s", nsSchema, data)
	case p.MetadataConfiguration.Name == "":
		t.Errorf("MetadataConfiguration not in %s:\nxml: %s", nsSchema, data)
	}
}

func TestVideoSourcesWireNamespaces(t *testing.T) {
	data := marshalOrFail(t, &GetVideoSourcesResponse{
		VideoSources: []VideoSource{{Framerate: 15, Resolution: VideoResolution{Width: 1280}}},
	})

	var resp struct {
		XMLName      xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetVideoSourcesResponse"`
		VideoSources []struct {
			Framerate  float64 `xml:"http://www.onvif.org/ver10/schema Framerate"`
			Resolution struct {
				Width int `xml:"http://www.onvif.org/ver10/schema Width"`
			} `xml:"http://www.onvif.org/ver10/schema Resolution"`
		} `xml:"http://www.onvif.org/ver10/media/wsdl VideoSources"`
	}
	decodeOrFail(t, data, &resp)

	if len(resp.VideoSources) != 1 || resp.VideoSources[0].Framerate != 15 || resp.VideoSources[0].Resolution.Width != 1280 {
		t.Errorf("VideoSources children not in %s:\nxml: %s", nsSchema, data)
	}
}

type nsUri struct {
	URI string `xml:"http://www.onvif.org/ver10/schema Uri"`
}

func TestSnapshotUriResolvesToSchemaNamespace(t *testing.T) {
	data := marshalOrFail(t, &GetSnapshotUriResponse{MediaUri: MediaUri{URI: "http://cam/snap.jpg"}})

	var env struct {
		XMLName  xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetSnapshotUriResponse"`
		MediaUri nsUri    `xml:"http://www.onvif.org/ver10/media/wsdl MediaUri"`
	}
	decodeOrFail(t, data, &env)

	if env.MediaUri.URI != "http://cam/snap.jpg" {
		t.Errorf("Uri = %q, want %q (not in %s)\nxml: %s", env.MediaUri.URI, "http://cam/snap.jpg", nsSchema, data)
	}
}

func TestPTZStatusWireNamespaces(t *testing.T) {
	data := marshalOrFail(t, &GetStatusResponse{
		PTZStatus: &PTZStatus{
			Position:   respPTZVector{PanTilt: &respVector2D{X: 0.5}},
			MoveStatus: PTZMoveStatus{PanTilt: "IDLE"},
		},
	})

	var env struct {
		XMLName   xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GetStatusResponse"`
		PTZStatus struct {
			XMLName  xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl PTZStatus"`
			Position struct {
				PanTilt struct {
					X float64 `xml:"x,attr"`
				} `xml:"http://www.onvif.org/ver10/schema PanTilt"`
			} `xml:"http://www.onvif.org/ver10/schema Position"`
			MoveStatus struct {
				PanTilt string `xml:"http://www.onvif.org/ver10/schema PanTilt"`
			} `xml:"http://www.onvif.org/ver10/schema MoveStatus"`
		} `xml:"http://www.onvif.org/ver20/ptz/wsdl PTZStatus"`
	}
	decodeOrFail(t, data, &env)

	if env.PTZStatus.XMLName.Local != "PTZStatus" ||
		env.PTZStatus.Position.PanTilt.X != 0.5 ||
		env.PTZStatus.MoveStatus.PanTilt != "IDLE" {
		t.Errorf("PTZStatus tree not in WSDL namespaces:\nxml: %s", data)
	}
}

func TestPresetsWireNamespaces(t *testing.T) {
	data := marshalOrFail(t, &GetPresetsResponse{
		Preset: []PTZPreset{{
			Token: "preset_1", Name: "Gate",
			PTZPosition: &respPTZVector{PanTilt: &respVector2D{X: 1}},
		}},
	})

	var resp struct {
		XMLName xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl GetPresetsResponse"`
		Preset  []struct {
			Name        string `xml:"http://www.onvif.org/ver10/schema Name"`
			PTZPosition struct {
				PanTilt struct {
					X float64 `xml:"x,attr"`
				} `xml:"http://www.onvif.org/ver10/schema PanTilt"`
			} `xml:"http://www.onvif.org/ver10/schema PTZPosition"`
		} `xml:"http://www.onvif.org/ver20/ptz/wsdl Preset"`
	}
	decodeOrFail(t, data, &resp)

	if len(resp.Preset) != 1 || resp.Preset[0].Name != "Gate" || resp.Preset[0].PTZPosition.PanTilt.X != 1 {
		t.Errorf("Preset tree not in WSDL namespaces:\nxml: %s", data)
	}
}

func TestImagingSettingsResolvesToSchemaNamespace(t *testing.T) {
	brightness := 60.0
	data := marshalOrFail(t, &GetImagingSettingsResponse{
		ImagingSettings: toRespImagingSettings(&ImagingSettings{Brightness: &brightness}),
	})

	var resp struct {
		XMLName         xml.Name `xml:"http://www.onvif.org/ver20/imaging/wsdl GetImagingSettingsResponse"`
		ImagingSettings struct {
			Brightness *float64 `xml:"http://www.onvif.org/ver10/schema Brightness"`
		} `xml:"http://www.onvif.org/ver20/imaging/wsdl ImagingSettings"`
	}
	decodeOrFail(t, data, &resp)

	if resp.ImagingSettings.Brightness == nil || *resp.ImagingSettings.Brightness != 60 {
		t.Errorf("ImagingSettings/Brightness not in %s:\nxml: %s", nsSchema, data)
	}
}

func TestImagingOptionsWireNamespaces(t *testing.T) {
	data := marshalOrFail(t, &GetOptionsResponse{
		ImagingOptions: toRespImagingOptions(&provider.ImagingOptions{
			Brightness: &provider.FloatRange{Min: 0, Max: 100},
			Exposure:   &provider.ExposureOptions{Mode: []string{"AUTO"}},
		}),
	})

	var resp struct {
		XMLName        xml.Name `xml:"http://www.onvif.org/ver20/imaging/wsdl GetOptionsResponse"`
		ImagingOptions struct {
			Brightness struct {
				Min float64 `xml:"http://www.onvif.org/ver10/schema Min"`
				Max float64 `xml:"http://www.onvif.org/ver10/schema Max"`
			} `xml:"http://www.onvif.org/ver10/schema Brightness"`
			Exposure struct {
				Mode []string `xml:"http://www.onvif.org/ver10/schema Mode"`
			} `xml:"http://www.onvif.org/ver10/schema Exposure"`
		} `xml:"http://www.onvif.org/ver20/imaging/wsdl ImagingOptions"`
	}
	decodeOrFail(t, data, &resp)

	if resp.ImagingOptions.Brightness.Min != 0 || resp.ImagingOptions.Brightness.Max != 100 ||
		len(resp.ImagingOptions.Exposure.Mode) != 1 {
		t.Errorf("ImagingOptions children not in %s:\nxml: %s", nsSchema, data)
	}
}

func TestEventServiceCapabilitiesWireNamespaces(t *testing.T) {
	resp := &GetEventServiceCapabilitiesResponse{}
	resp.Capabilities.WSPullPointSupport = true
	resp.Capabilities.MaxPullPoints = 10
	data := marshalOrFail(t, resp)

	var decoded struct {
		XMLName      xml.Name `xml:"http://www.onvif.org/ver10/events/wsdl GetServiceCapabilitiesResponse"`
		Capabilities struct {
			XMLName            xml.Name `xml:"http://www.onvif.org/ver10/events/wsdl Capabilities"`
			WSPullPointSupport bool     `xml:"WSPullPointSupport,attr"`
		} `xml:"http://www.onvif.org/ver10/events/wsdl Capabilities"`
	}
	decodeOrFail(t, data, &decoded)

	if decoded.Capabilities.XMLName.Local != "Capabilities" || !decoded.Capabilities.WSPullPointSupport {
		t.Errorf("Capabilities element not in %s:\nxml: %s", nsEventsWSDL, data)
	}
}

func TestCreatePullPointWireNamespaces(t *testing.T) {
	data := marshalOrFail(t, &CreatePullPointSubscriptionResponse{
		SubscriptionReference: endpointReference{Address: "http://cam/onvif/events_service/sub/abc"},
		CurrentTime:           "2026-09-19T00:00:00Z",
		TerminationTime:       "2026-09-19T01:00:00Z",
	})

	var resp struct {
		XMLName               xml.Name `xml:"http://www.onvif.org/ver10/events/wsdl CreatePullPointSubscriptionResponse"`
		SubscriptionReference struct {
			Address string `xml:"http://www.w3.org/2005/08/addressing Address"`
		} `xml:"http://www.onvif.org/ver10/events/wsdl SubscriptionReference"`
		CurrentTime     string `xml:"http://docs.oasis-open.org/wsn/b-2 CurrentTime"`
		TerminationTime string `xml:"http://docs.oasis-open.org/wsn/b-2 TerminationTime"`
	}
	decodeOrFail(t, data, &resp)

	if resp.SubscriptionReference.Address == "" || resp.CurrentTime == "" || resp.TerminationTime == "" {
		t.Errorf("subscription tree not in WSDL namespaces (tev/wsa/wsnt):\nxml: %s", data)
	}
}

func TestPullMessagesWireNamespaces(t *testing.T) {
	msg := notificationMessage{
		Topic: eventTopic{Value: "tns1:VideoSource/MotionAlarm"},
		ProducerReference: &endpointReference{
			Address: "http://cam/onvif/device_service",
		},
		Message: notificationPayload{Inner: eventMessage{
			PropertyOperation: "Changed",
			Source:            eventItemGroup{SimpleItems: []eventSimpleItem{{Name: "Source", Value: "CSI"}}},
		}},
	}
	data := marshalOrFail(t, &PullMessagesResponse{
		CurrentTime:          "2026-09-19T00:00:00Z",
		TerminationTime:      "2026-09-19T01:00:00Z",
		NotificationMessages: []notificationMessage{msg},
	})

	var resp struct {
		XMLName              xml.Name `xml:"http://www.onvif.org/ver10/events/wsdl PullMessagesResponse"`
		CurrentTime          string   `xml:"http://www.onvif.org/ver10/events/wsdl CurrentTime"`
		TerminationTime      string   `xml:"http://www.onvif.org/ver10/events/wsdl TerminationTime"`
		NotificationMessages []struct {
			Topic struct {
				Value string `xml:",chardata"`
			} `xml:"http://docs.oasis-open.org/wsn/b-2 Topic"`
			ProducerReference struct {
				Address string `xml:"http://www.w3.org/2005/08/addressing Address"`
			} `xml:"http://docs.oasis-open.org/wsn/b-2 ProducerReference"`
			Message struct {
				Inner struct {
					Source struct {
						SimpleItem []struct {
							Name  string `xml:"Name,attr"`
							Value string `xml:"Value,attr"`
						} `xml:"http://www.onvif.org/ver10/schema SimpleItem"`
					} `xml:"http://www.onvif.org/ver10/schema Source"`
				} `xml:"http://www.onvif.org/ver10/schema Message"`
			} `xml:"http://docs.oasis-open.org/wsn/b-2 Message"`
		} `xml:"http://docs.oasis-open.org/wsn/b-2 NotificationMessage"`
	}
	decodeOrFail(t, data, &resp)

	if resp.CurrentTime == "" || resp.TerminationTime == "" {
		t.Errorf("PullMessages times not in %s:\nxml: %s", nsEventsWSDL, data)
	}

	if len(resp.NotificationMessages) != 1 {
		t.Fatalf("NotificationMessage count = %d, want 1 (element in %s)\nxml: %s", len(resp.NotificationMessages), nsWSNT, data)
	}

	nm := resp.NotificationMessages[0]
	if nm.Topic.Value == "" || nm.ProducerReference.Address == "" ||
		len(nm.Message.Inner.Source.SimpleItem) != 1 || nm.Message.Inner.Source.SimpleItem[0].Value != "CSI" {
		t.Errorf("notification tree not in WSDL namespaces (wsnt/wsa/tt):\nxml: %s", data)
	}
}

func TestRenewWireNamespaces(t *testing.T) {
	data := marshalOrFail(t, &RenewResponse{
		CurrentTime:     "2026-09-19T00:00:00Z",
		TerminationTime: "2026-09-19T01:00:00Z",
	})

	var resp struct {
		XMLName         xml.Name `xml:"http://docs.oasis-open.org/wsn/b-2 RenewResponse"`
		CurrentTime     string   `xml:"http://docs.oasis-open.org/wsn/b-2 CurrentTime"`
		TerminationTime string   `xml:"http://docs.oasis-open.org/wsn/b-2 TerminationTime"`
	}
	decodeOrFail(t, data, &resp)

	if resp.CurrentTime == "" || resp.TerminationTime == "" {
		t.Errorf("RenewResponse times not in %s:\nxml: %s", nsWSNT, data)
	}
}

// TestExplicitPrefixConventionalNamespaces drives full HTTP responses in
// explicit-prefix mode: unprefixed tags (no namespace at all after the
// prefix rewriter) and wrong-namespace tags alike must not survive — every
// element reappears under its conventional prefix.
// postSOAPAllServices mounts every service (like RegisterServices) and
// posts one SOAP request.
func postSOAPAllServices(t *testing.T, s *Server, endpoint, innerXML string) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	s.RegisterServices(mux)

	req := httptest.NewRequest(http.MethodPost, s.config.BasePath+endpoint, strings.NewReader(
		`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>`+innerXML+
			`</s:Body></s:Envelope>`))
	req.RemoteAddr = "127.0.0.1:41000"

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	return w
}

func TestExplicitPrefixConventionalNamespaces(t *testing.T) {
	cfg := createTestConfig()
	cfg.ExplicitPrefixes = true
	cfg.SupportPTZ = true

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	rec := postSOAPAllServices(t, srv, "/device_service", `<GetServices xmlns="`+nsDeviceWSDL+`"/>`)
	if rec.Code != http.StatusOK {
		t.Fatalf("GetServices status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()

	for _, want := range []string{
		`tds:GetServicesResponse xmlns:tds="` + nsDeviceWSDL + `"`,
		"<tds:Service>",
		"<tds:XAddr",
		"<tds:Version>",
		`<tt:Major xmlns:tt="` + nsSchema + `">2</tt:Major>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("explicit-prefix GetServices body missing %q:\n%s", want, body)
		}
	}

	rec = postSOAPAllServices(t, srv, "/ptz_service",
		`<GetStatus xmlns="`+nsPTZWSDL+`"><ProfileToken>profile_token_1</ProfileToken></GetStatus>`)
	if rec.Code != http.StatusOK {
		t.Fatalf("GetStatus status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	body = rec.Body.String()

	for _, want := range []string{
		"<tptz:PTZStatus",
		"<tt:Position",
		"<tt:MoveStatus",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("explicit-prefix GetStatus body missing %q:\n%s", want, body)
		}
	}
}
