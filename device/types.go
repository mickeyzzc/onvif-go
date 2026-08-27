// Package device hosts the device-service (tds) domain types.
package device

import (
	"time"

	"github.com/mickeyzzc/onvif-go/v2/types"
)

// DeviceInformation contains basic device information.
type DeviceInformation struct {
	Manufacturer    string
	Model           string
	FirmwareVersion string
	SerialNumber    string
	HardwareID      string
}

// Capabilities represents the device capabilities.
type Capabilities struct {
	Analytics *AnalyticsCapabilities
	Device    *DeviceCapabilities
	Events    *EventCapabilities
	Imaging   *ImagingCapabilities
	Media     *MediaCapabilities
	PTZ       *PTZCapabilities
	Extension *CapabilitiesExtension
}

// AnalyticsCapabilities represents analytics service capabilities.
type AnalyticsCapabilities struct {
	XAddr                  string
	RuleSupport            bool
	AnalyticsModuleSupport bool
}

// DeviceCapabilities represents device service capabilities.
type DeviceCapabilities struct {
	XAddr    string
	Network  *NetworkCapabilities
	System   *SystemCapabilities
	IO       *IOCapabilities
	Security *SecurityCapabilities
}

// EventCapabilities represents event service capabilities.
type EventCapabilities struct {
	XAddr                         string
	WSSubscriptionPolicySupport   bool
	WSPullPointSupport            bool
	WSPausableSubscriptionSupport bool
}

// ImagingCapabilities represents imaging service capabilities.
type ImagingCapabilities struct {
	XAddr string
}

// MediaCapabilities represents media service capabilities.
type MediaCapabilities struct {
	XAddr                 string
	StreamingCapabilities *StreamingCapabilities
}

// PTZCapabilities represents PTZ service capabilities.
type PTZCapabilities struct {
	XAddr string
}

// NetworkCapabilities represents network capabilities.
type NetworkCapabilities struct {
	IPFilter          bool
	ZeroConfiguration bool
	IPVersion6        bool
	DynDNS            bool
	Extension         *NetworkCapabilitiesExtension
}

// SystemCapabilities represents system capabilities.
type SystemCapabilities struct {
	DiscoveryResolve  bool
	DiscoveryBye      bool
	RemoteDiscovery   bool
	SystemBackup      bool
	SystemLogging     bool
	FirmwareUpgrade   bool
	SupportedVersions []string
	Extension         *SystemCapabilitiesExtension
}

// IOCapabilities represents I/O capabilities.
type IOCapabilities struct {
	InputConnectors int
	RelayOutputs    int
	Extension       *IOCapabilitiesExtension
}

// SecurityCapabilities represents security capabilities.
type SecurityCapabilities struct {
	TLS11                bool
	TLS12                bool
	OnboardKeyGeneration bool
	AccessPolicyConfig   bool
	X509Token            bool
	SAMLToken            bool
	KerberosToken        bool
	RELToken             bool
	Extension            *SecurityCapabilitiesExtension
}

// StreamingCapabilities represents streaming capabilities.
type StreamingCapabilities struct {
	RTPMulticast bool
	RTPTCP       bool
	RTPRTSPTCP   bool
	Extension    *StreamingCapabilitiesExtension
}

// HostnameInformation represents hostname configuration.
type HostnameInformation struct {
	FromDHCP bool
	Name     string
}

// DNSInformation represents DNS configuration.
type DNSInformation struct {
	FromDHCP     bool
	SearchDomain []string
	DNSFromDHCP  []types.IPAddress
	DNSManual    []types.IPAddress
}

// NTPInformation represents NTP configuration.
type NTPInformation struct {
	FromDHCP    bool
	NTPFromDHCP []NetworkHost
	NTPManual   []NetworkHost
}

// NetworkHost represents a network host.
type NetworkHost struct {
	Type        string // IPv4, IPv6, DNS
	IPv4Address string
	IPv6Address string
	DNSname     string
}

// NetworkInterface represents a network interface.
type NetworkInterface struct {
	Token   string
	Enabled bool
	Info    NetworkInterfaceInfo
	IPv4    *IPv4NetworkInterface
	IPv6    *IPv6NetworkInterface
}

// NetworkInterfaceInfo represents network interface info.
type NetworkInterfaceInfo struct {
	Name      string
	HwAddress string
	MTU       int
}

// IPv4NetworkInterface represents IPv4 configuration.
type IPv4NetworkInterface struct {
	Enabled bool
	Config  IPv4Configuration
}

// IPv6NetworkInterface represents IPv6 configuration.
type IPv6NetworkInterface struct {
	Enabled bool
	Config  IPv6Configuration
}

// IPv4Configuration represents IPv4 configuration.
type IPv4Configuration struct {
	Manual []types.PrefixedIPv4Address
	DHCP   bool
}

// IPv6Configuration represents IPv6 configuration.
type IPv6Configuration struct {
	Manual []types.PrefixedIPv6Address
	DHCP   bool
}

// types.PrefixedIPv4Address represents an IPv4 address with prefix. Netmask is a
// dotted-decimal convenience derived from PrefixLength (e.g. "255.255.255.0"

// Scope represents a device scope.
type Scope struct {
	ScopeDef  string
	ScopeItem string
}

// ServiceEntry represents one ONVIF service in the GetServices response.
type ServiceEntry struct {
	Namespace    string
	XAddr        string
	Capabilities interface{}
	Version      OnvifVersion
}

// OnvifVersion represents ONVIF version.
type OnvifVersion struct {
	Major int
	Minor int
}

// DeviceServiceCapabilities represents device service capabilities.
type DeviceServiceCapabilities struct {
	Network  *NetworkCapabilities
	Security *SecurityCapabilities
	System   *SystemCapabilities
	Misc     *MiscCapabilities
}

// MiscCapabilities represents miscellaneous capabilities.
type MiscCapabilities struct {
	AuxiliaryCommands []string
}

// DiscoveryMode represents discovery mode.
type DiscoveryMode string

// NetworkProtocol represents network protocol configuration.
type NetworkProtocol struct {
	Name    NetworkProtocolType
	Enabled bool
	Port    []int
}

// NetworkProtocolType represents protocol type.
type NetworkProtocolType string

// NetworkGateway represents default gateway.
type NetworkGateway struct {
	IPv4Address []string
	IPv6Address []string
}

// SystemDateTime represents system date and time.
type SystemDateTime struct {
	DateTimeType    SetDateTimeType
	DaylightSavings bool
	TimeZone        *TimeZone
	UTCDateTime     *DateTime
	LocalDateTime   *DateTime
}

// SetDateTimeType represents date/time set method.
type SetDateTimeType string

// TimeZone represents timezone.
type TimeZone struct {
	TZ string // POSIX format
}

// DateTime represents date and time.
type DateTime struct {
	Time Time
	Date Date
}

// Time represents time.
type Time struct {
	Hour   int
	Minute int
	Second int
}

// Date represents date.
type Date struct {
	Year  int
	Month int
	Day   int
}

// SystemLogType represents system log type.
type SystemLogType string

// SystemLog represents system log data.
type SystemLog struct {
	Binary *AttachmentData
	String string
}

// AttachmentData represents attachment/binary data.
type AttachmentData struct {
	ContentType string
	Include     *Include
}

// Include represents XOP include.
type Include struct {
	Href string
}

// BackupFile represents backup file.
type BackupFile struct {
	Name string
	Data AttachmentData
}

// FactoryDefaultType represents factory default type.
type FactoryDefaultType string

// SupportInformation represents support information.
type SupportInformation struct {
	Binary *AttachmentData
	String string
}

// SystemLogURIList represents system log URIs.
type SystemLogURIList struct {
	SystemLog []SystemLogURI
}

// SystemLogURI represents system log URI.
type SystemLogURI struct {
	Type SystemLogType
	URI  string
}

// NetworkZeroConfiguration represents zero-configuration.
type NetworkZeroConfiguration struct {
	InterfaceToken string
	Enabled        bool
	Addresses      []string
}

// DynamicDNSInformation represents dynamic DNS info.
type DynamicDNSInformation struct {
	Type DynamicDNSType
	Name string
	TTL  time.Duration
}

// DynamicDNSType represents dynamic DNS type.
type DynamicDNSType string

// Dot11Capabilities represents 802.11 capabilities.
type Dot11Capabilities struct {
	TKIP                  bool
	ScanAvailableNetworks bool
	MultipleConfiguration bool
	AdHocStationMode      bool
	WEP                   bool
}

// Dot11Status represents 802.11 status.
type Dot11Status struct {
	SSID              string
	BSSID             string
	PairCipher        Dot11Cipher
	GroupCipher       Dot11Cipher
	SignalStrength    Dot11SignalStrength
	ActiveConfigAlias string
}

// Dot11Cipher represents 802.11 cipher.
type Dot11Cipher string

// Dot11SignalStrength represents signal strength.
type Dot11SignalStrength string

// Dot1XConfiguration represents 802.1X configuration.
type Dot1XConfiguration struct {
	Dot1XConfigurationToken string
	Identity                string
	AnonymousID             string
	EAPMethod               int
	CACertificateID         []string
	EAPMethodConfiguration  *EAPMethodConfiguration
}

// EAPMethodConfiguration represents EAP method configuration.
type EAPMethodConfiguration struct {
	TLSConfiguration *TLSConfiguration
	Password         string
}

// TLSConfiguration represents TLS configuration.
type TLSConfiguration struct {
	CertificateID string
}

// Dot11AvailableNetworks represents available 802.11 networks.
type Dot11AvailableNetworks struct {
	SSID                  string
	BSSID                 string
	AuthAndMangementSuite []Dot11AuthAndMangementSuite
	PairCipher            []Dot11Cipher
	GroupCipher           []Dot11Cipher
	SignalStrength        Dot11SignalStrength
}

// Dot11AuthAndMangementSuite represents auth suite.
type Dot11AuthAndMangementSuite string

// StorageConfiguration represents storage configuration.
type StorageConfiguration struct {
	Token string
	Data  StorageConfigurationData
}

// StorageConfigurationData represents storage configuration data.
type StorageConfigurationData struct {
	Type                       string
	LocalPath                  string
	StorageURI                 string `xml:"StorageUri"`
	User                       *UserCredential
	CertPathValidationPolicyID string
}

// UserCredential represents user credentials.
type UserCredential struct {
	UserName string
	Password string
	Token    string
}

// LocationEntity represents geo location.
type LocationEntity struct {
	Entity    string  `xml:"Entity"`
	Token     string  `xml:"Token"`
	Fixed     bool    `xml:"Fixed"`
	Lon       float64 `xml:"Lon,attr"`
	Lat       float64 `xml:"Lat,attr"`
	Elevation float64 `xml:"Elevation,attr"`
}

// GeoLocation represents geographic location coordinates.
type GeoLocation struct {
	Lon       float64 `xml:"lon,attr,omitempty"`       // Longitude in degrees
	Lat       float64 `xml:"lat,attr,omitempty"`       // Latitude in degrees
	Elevation float64 `xml:"elevation,attr,omitempty"` // Elevation in meters
}

// Extension placeholder types for the capabilities tree.
type (
	CapabilitiesExtension          struct{}
	NetworkCapabilitiesExtension   struct{}
	SystemCapabilitiesExtension    struct{}
	IOCapabilitiesExtension        struct{}
	SecurityCapabilitiesExtension  struct{}
	StreamingCapabilitiesExtension struct{}
)

const (
	DiscoveryModeDiscoverable    DiscoveryMode = "Discoverable"
	DiscoveryModeNonDiscoverable DiscoveryMode = "NonDiscoverable"
)

const (
	NetworkProtocolHTTP  NetworkProtocolType = "HTTP"
	NetworkProtocolHTTPS NetworkProtocolType = "HTTPS"
	NetworkProtocolRTSP  NetworkProtocolType = "RTSP"
)

const (
	SetDateTimeManual SetDateTimeType = "Manual"
	SetDateTimeNTP    SetDateTimeType = "NTP"
)

const (
	SystemLogTypeSystem SystemLogType = "System"
	SystemLogTypeAccess SystemLogType = "Access"
)

const (
	FactoryDefaultHard FactoryDefaultType = "Hard"
	FactoryDefaultSoft FactoryDefaultType = "Soft"
)

const (
	DynamicDNSNoUpdate      DynamicDNSType = "NoUpdate"
	DynamicDNSClientUpdates DynamicDNSType = "ClientUpdates"
	DynamicDNSServerUpdates DynamicDNSType = "ServerUpdates"
)

const (
	Dot11CipherCCMP     Dot11Cipher = "CCMP"
	Dot11CipherTKIP     Dot11Cipher = "TKIP"
	Dot11CipherAny      Dot11Cipher = "Any"
	Dot11CipherExtended Dot11Cipher = "Extended"
)

const (
	Dot11SignalNone     Dot11SignalStrength = "None"
	Dot11SignalVeryBad  Dot11SignalStrength = "Very Bad"
	Dot11SignalBad      Dot11SignalStrength = "Bad"
	Dot11SignalGood     Dot11SignalStrength = "Good"
	Dot11SignalVeryGood Dot11SignalStrength = "Very Good"
	Dot11SignalExtended Dot11SignalStrength = "Extended"
)

const (
	Dot11AuthNone     Dot11AuthAndMangementSuite = "None"
	Dot11AuthDot1X    Dot11AuthAndMangementSuite = "Dot1X"
	Dot11AuthPSK      Dot11AuthAndMangementSuite = "PSK"
	Dot11AuthExtended Dot11AuthAndMangementSuite = "Extended"
)
