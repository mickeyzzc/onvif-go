package server

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

const (
	defaultHost     = "0.0.0.0"
	defaultHostname = "localhost"
)

// Device service SOAP message types

// GetDeviceInformationResponse represents GetDeviceInformation response.
// Its children are inline xs:string elements in the device WSDL, so they
// carry the tds namespace (not ver10/schema).
type GetDeviceInformationResponse struct {
	XMLName         xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetDeviceInformationResponse"`
	Manufacturer    string   `xml:"http://www.onvif.org/ver10/device/wsdl Manufacturer"`
	Model           string   `xml:"http://www.onvif.org/ver10/device/wsdl Model"`
	FirmwareVersion string   `xml:"http://www.onvif.org/ver10/device/wsdl FirmwareVersion"`
	SerialNumber    string   `xml:"http://www.onvif.org/ver10/device/wsdl SerialNumber"`
	HardwareID      string   `xml:"http://www.onvif.org/ver10/device/wsdl HardwareId"`
}

// GetCapabilitiesResponse represents GetCapabilities response. The
// Capabilities wrapper element is locally declared in the device WSDL
// (tds); its children — elements of the schema type tt:Capabilities — are
// ver10/schema.
type GetCapabilitiesResponse struct {
	XMLName      xml.Name      `xml:"http://www.onvif.org/ver10/device/wsdl GetCapabilitiesResponse"`
	Capabilities *Capabilities `xml:"http://www.onvif.org/ver10/device/wsdl Capabilities"`
}

// Capabilities represents device capabilities.
type Capabilities struct {
	Analytics *AnalyticsCapabilities `xml:"http://www.onvif.org/ver10/schema Analytics,omitempty"`
	Device    *DeviceCapabilities    `xml:"http://www.onvif.org/ver10/schema Device"`
	Events    *EventCapabilities     `xml:"http://www.onvif.org/ver10/schema Events,omitempty"`
	Imaging   *ImagingCapabilities   `xml:"http://www.onvif.org/ver10/schema Imaging,omitempty"`
	Media     *MediaCapabilities     `xml:"http://www.onvif.org/ver10/schema Media"`
	PTZ       *PTZCapabilities       `xml:"http://www.onvif.org/ver10/schema PTZ,omitempty"`
}

// AnalyticsCapabilities represents analytics service capabilities.
type AnalyticsCapabilities struct {
	XAddr                  string `xml:"http://www.onvif.org/ver10/schema XAddr"`
	RuleSupport            bool   `xml:"RuleSupport,attr"`
	AnalyticsModuleSupport bool   `xml:"AnalyticsModuleSupport,attr"`
}

// DeviceCapabilities represents device service capabilities.
type DeviceCapabilities struct {
	XAddr    string                `xml:"http://www.onvif.org/ver10/schema XAddr"`
	Network  *NetworkCapabilities  `xml:"http://www.onvif.org/ver10/schema Network,omitempty"`
	System   *SystemCapabilities   `xml:"http://www.onvif.org/ver10/schema System,omitempty"`
	IO       *IOCapabilities       `xml:"http://www.onvif.org/ver10/schema IO,omitempty"`
	Security *SecurityCapabilities `xml:"http://www.onvif.org/ver10/schema Security,omitempty"`
}

// NetworkCapabilities represents network capabilities.
type NetworkCapabilities struct {
	IPFilter          bool `xml:"IPFilter,attr"`
	ZeroConfiguration bool `xml:"ZeroConfiguration,attr"`
	IPVersion6        bool `xml:"IPVersion6,attr"`
	DynDNS            bool `xml:"DynDNS,attr"`
}

// SystemCapabilities represents system capabilities.
type SystemCapabilities struct {
	DiscoveryResolve bool `xml:"DiscoveryResolve,attr"`
	DiscoveryBye     bool `xml:"DiscoveryBye,attr"`
	RemoteDiscovery  bool `xml:"RemoteDiscovery,attr"`
	SystemBackup     bool `xml:"SystemBackup,attr"`
	SystemLogging    bool `xml:"SystemLogging,attr"`
	FirmwareUpgrade  bool `xml:"FirmwareUpgrade,attr"`
}

// IOCapabilities represents I/O capabilities.
type IOCapabilities struct {
	InputConnectors int `xml:"InputConnectors,attr"`
	RelayOutputs    int `xml:"RelayOutputs,attr"`
}

// SecurityCapabilities represents security capabilities.
type SecurityCapabilities struct {
	TLS11                bool `xml:"TLS1.1,attr"`
	TLS12                bool `xml:"TLS1.2,attr"`
	OnboardKeyGeneration bool `xml:"OnboardKeyGeneration,attr"`
	AccessPolicyConfig   bool `xml:"AccessPolicyConfig,attr"`
	X509Token            bool `xml:"X.509Token,attr"`
	SAMLToken            bool `xml:"SAMLToken,attr"`
	KerberosToken        bool `xml:"KerberosToken,attr"`
	RELToken             bool `xml:"RELToken,attr"`
}

// EventCapabilities represents event service capabilities.
type EventCapabilities struct {
	XAddr                         string `xml:"http://www.onvif.org/ver10/schema XAddr"`
	WSSubscriptionPolicySupport   bool   `xml:"WSSubscriptionPolicySupport,attr"`
	WSPullPointSupport            bool   `xml:"WSPullPointSupport,attr"`
	WSPausableSubscriptionSupport bool   `xml:"WSPausableSubscriptionManagerInterfaceSupport,attr"`
}

// ImagingCapabilities represents imaging service capabilities.
type ImagingCapabilities struct {
	XAddr string `xml:"http://www.onvif.org/ver10/schema XAddr"`
}

// MediaCapabilities represents media service capabilities.
type MediaCapabilities struct {
	XAddr                 string                 `xml:"http://www.onvif.org/ver10/schema XAddr"`
	StreamingCapabilities *StreamingCapabilities `xml:"http://www.onvif.org/ver10/schema StreamingCapabilities"`
}

// StreamingCapabilities represents streaming capabilities.
type StreamingCapabilities struct {
	RTPMulticast bool `xml:"RTPMulticast,attr"`
	RTPTCP       bool `xml:"RTP_TCP,attr"`
	RTPRTSPTCP   bool `xml:"RTP_RTSP_TCP,attr"`
}

// PTZCapabilities represents PTZ service capabilities.
type PTZCapabilities struct {
	XAddr string `xml:"http://www.onvif.org/ver10/schema XAddr"`
}

// GetServicesResponse represents GetServices response. Service and its
// Namespace/XAddr/Version children are locally declared in the device WSDL
// (tds); the version numbers inside are children of the schema type
// tt:OnvifVersion, so Major/Minor are ver10/schema.
type GetServicesResponse struct {
	XMLName xml.Name  `xml:"http://www.onvif.org/ver10/device/wsdl GetServicesResponse"`
	Service []Service `xml:"http://www.onvif.org/ver10/device/wsdl Service"`
}

// Service represents a service.
type Service struct {
	Namespace string  `xml:"http://www.onvif.org/ver10/device/wsdl Namespace"`
	XAddr     string  `xml:"http://www.onvif.org/ver10/device/wsdl XAddr"`
	Version   Version `xml:"http://www.onvif.org/ver10/device/wsdl Version"`
}

// Version represents service version.
type Version struct {
	Major int `xml:"http://www.onvif.org/ver10/schema Major"`
	Minor int `xml:"http://www.onvif.org/ver10/schema Minor"`
}

// SystemRebootResponse represents SystemReboot response.
type SystemRebootResponse struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl SystemRebootResponse"`
	Message string   `xml:"http://www.onvif.org/ver10/device/wsdl Message"`
}

// Device service handlers

// HandleGetDeviceInformation handles GetDeviceInformation request.
func (s *Server) HandleGetDeviceInformation(rc *soap.RequestContext, body []byte) (interface{}, error) {
	info := s.deviceInfo.DeviceInfo()

	return &GetDeviceInformationResponse{
		Manufacturer:    info.Manufacturer,
		Model:           info.Model,
		FirmwareVersion: info.FirmwareVersion,
		SerialNumber:    info.SerialNumber,
		HardwareID:      info.HardwareID,
	}, nil
}

// HandleGetCapabilities handles GetCapabilities request.
func (s *Server) HandleGetCapabilities(rc *soap.RequestContext, body []byte) (interface{}, error) {
	host := s.advertiseHost(rc)

	baseURL := fmt.Sprintf("http://%s:%d%s", host, s.config.Port, s.config.BasePath)

	capabilities := &Capabilities{
		Device: &DeviceCapabilities{
			XAddr: baseURL + "/device_service",
			Network: &NetworkCapabilities{
				IPFilter:          false,
				ZeroConfiguration: false,
				IPVersion6:        false,
				DynDNS:            false,
			},
			System: &SystemCapabilities{
				DiscoveryResolve: true,
				DiscoveryBye:     true,
				RemoteDiscovery:  true,
				SystemBackup:     false,
				SystemLogging:    false,
				FirmwareUpgrade:  false,
			},
			IO: &IOCapabilities{
				InputConnectors: 0,
				RelayOutputs:    0,
			},
			Security: &SecurityCapabilities{
				TLS11:                false,
				TLS12:                false,
				OnboardKeyGeneration: false,
				AccessPolicyConfig:   false,
				X509Token:            false,
				SAMLToken:            false,
				KerberosToken:        false,
				RELToken:             false,
			},
		},
		Media: &MediaCapabilities{
			XAddr: baseURL + "/media_service",
			StreamingCapabilities: &StreamingCapabilities{
				RTPMulticast: false,
				RTPTCP:       true,
				RTPRTSPTCP:   true,
			},
		},
	}

	if s.config.SupportPTZ {
		capabilities.PTZ = &PTZCapabilities{
			XAddr: baseURL + "/ptz_service",
		}
	}

	if s.config.SupportImaging {
		capabilities.Imaging = &ImagingCapabilities{
			XAddr: baseURL + "/imaging_service",
		}
	}

	if s.config.SupportEvents {
		// WSPullPointSupport is true since #83: the events service really
		// serves CreatePullPointSubscription/PullMessages/Renew/Unsubscribe
		// on the advertised XAddr.
		capabilities.Events = &EventCapabilities{
			XAddr:                         baseURL + "/events_service",
			WSSubscriptionPolicySupport:   false,
			WSPullPointSupport:            true,
			WSPausableSubscriptionSupport: false,
		}
	}

	return &GetCapabilitiesResponse{
		Capabilities: capabilities,
	}, nil
}

// HandleGetSystemDateAndTime handles GetSystemDateAndTime request.
func (s *Server) HandleGetSystemDateAndTime(rc *soap.RequestContext, body []byte) (interface{}, error) {
	now := time.Now().UTC()

	return &soap.GetSystemDateAndTimeResponse{
		SystemDateAndTime: soap.SystemDateAndTime{
			DateTimeType:    "NTP",
			DaylightSavings: false,
			TimeZone: soap.TimeZone{
				TZ: "UTC",
			},
			UTCDateTime:   soap.ToDateTime(now),
			LocalDateTime: soap.ToDateTime(now.Local()),
		},
	}, nil
}

// HandleGetServices handles GetServices request.
func (s *Server) HandleGetServices(rc *soap.RequestContext, body []byte) (interface{}, error) {
	host := s.advertiseHost(rc)

	baseURL := fmt.Sprintf("http://%s:%d%s", host, s.config.Port, s.config.BasePath)

	services := []Service{
		{
			Namespace: "http://www.onvif.org/ver10/device/wsdl",
			XAddr:     baseURL + "/device_service",
			Version:   Version{Major: 2, Minor: 5}, //nolint:mnd // ONVIF version
		},
		{
			Namespace: "http://www.onvif.org/ver10/media/wsdl",
			XAddr:     baseURL + "/media_service",
			Version:   Version{Major: 2, Minor: 5}, //nolint:mnd // ONVIF version
		},
	}

	if s.config.SupportPTZ {
		services = append(services, Service{
			Namespace: "http://www.onvif.org/ver20/ptz/wsdl",
			XAddr:     baseURL + "/ptz_service",
			Version:   Version{Major: 2, Minor: 5}, //nolint:mnd // ONVIF version
		})
	}

	if s.config.SupportImaging {
		services = append(services, Service{
			Namespace: "http://www.onvif.org/ver20/imaging/wsdl",
			XAddr:     baseURL + "/imaging_service",
			Version:   Version{Major: 2, Minor: 5}, //nolint:mnd // ONVIF version
		})
	}

	// GetCapabilities advertises the Events XAddr under the same flag;
	// the two enumerations must agree (#46) — clients (including this
	// library's Initialize) enumerate services via GetServices.
	if s.config.SupportEvents {
		services = append(services, Service{
			Namespace: "http://www.onvif.org/ver10/events/wsdl",
			XAddr:     baseURL + "/events_service",
			Version:   Version{Major: 2, Minor: 5}, //nolint:mnd // ONVIF version
		})
	}

	return &GetServicesResponse{
		Service: services,
	}, nil
}

// HandleSystemReboot handles SystemReboot request.
func (s *Server) HandleSystemReboot(rc *soap.RequestContext, body []byte) (interface{}, error) {
	return &SystemRebootResponse{
		Message: "Device rebooting",
	}, nil
}

// DefaultScopes is the scope set GetScopes serves when Config.Scopes is
// empty — the conventional ONVIF device scope URIs. Hosts advertising
// WS-Discovery should mirror their Responder Scopes into Config.Scopes
// so ProbeMatches and GetScopes agree (#37).
//
// It is a function (not a package var) so callers always get a fresh
// slice — appending to a shared package-level var would mutate the
// served scope set process-wide.
func DefaultScopes() []string {
	return []string{
		"onvif://www.onvif.org/type/NetworkVideoTransmitter",
		"onvif://www.onvif.org/type/video_encoder",
	}
}

// GetScopesResponse represents the GetScopes response. Each Scopes entry
// is the schema type tt:Scope: a ScopeDef enum ("Fixed"|"Configurable")
// followed by the ScopeItem URI, both ver10/schema; the Scopes wrapper
// element itself is locally declared in the device WSDL (tds).
type GetScopesResponse struct {
	XMLName xml.Name          `xml:"http://www.onvif.org/ver10/device/wsdl GetScopesResponse"`
	Scopes  []ScopeDefinition `xml:"http://www.onvif.org/ver10/device/wsdl Scopes"`
}

// ScopeDefinition carries one advertised scope in the tt:Scope wire shape
// (ScopeDef + ScopeItem elements).
type ScopeDefinition struct {
	ScopeDef  string `xml:"http://www.onvif.org/ver10/schema ScopeDef"`
	ScopeItem string `xml:"http://www.onvif.org/ver10/schema ScopeItem"`
}

// HandleGetScopes handles the GetScopes request (#37): the configured
// scope list, or DefaultScopes when none is configured.
func (s *Server) HandleGetScopes(_ *soap.RequestContext, _ []byte) (interface{}, error) {
	scopes := s.config.Scopes
	if len(scopes) == 0 {
		scopes = DefaultScopes()
	}

	resp := &GetScopesResponse{Scopes: make([]ScopeDefinition, len(scopes))}
	for i, scope := range scopes {
		resp.Scopes[i] = ScopeDefinition{ScopeDef: "Fixed", ScopeItem: scope}
	}

	return resp, nil
}
