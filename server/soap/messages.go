package soap

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// RequestWrapper wraps incoming SOAP request structures.
type RequestWrapper struct {
	XMLName xml.Name
	Content []byte `xml:",innerxml"`
}

// ParseRequest parses a SOAP request into a specific structure. The body
// content handed to handlers is raw inner XML ([]byte) and is decoded
// directly; struct values take a marshal round-trip for compatibility
// with direct handler invocation.
func ParseRequest(bodyContent, target interface{}) error {
	if raw, ok := bodyContent.([]byte); ok {
		if err := xml.Unmarshal(raw, target); err != nil {
			return fmt.Errorf("failed to unmarshal request: %w", err)
		}

		return nil
	}

	// Marshal the body content back to XML
	bodyXML, err := xml.Marshal(bodyContent)
	if err != nil {
		return fmt.Errorf("failed to marshal body content: %w", err)
	}

	// Unmarshal into target structure
	if err := xml.Unmarshal(bodyXML, target); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	return nil
}

// NormalizeAction normalizes SOAP action names.
func NormalizeAction(action string) string {
	// Remove namespace prefixes
	if idx := strings.LastIndex(action, ":"); idx != -1 {
		action = action[idx+1:]
	}

	return action
}

// Common SOAP request/response structures for ONVIF

// GetSystemDateAndTimeRequest represents GetSystemDateAndTime request.
type GetSystemDateAndTimeRequest struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetSystemDateAndTime"`
}

// GetSystemDateAndTimeResponse represents GetSystemDateAndTime response.
// The SystemDateAndTime wrapper is locally declared in the device WSDL
// (tds); its typed children below are ver10/schema (tt).
type GetSystemDateAndTimeResponse struct {
	XMLName           xml.Name          `xml:"http://www.onvif.org/ver10/device/wsdl GetSystemDateAndTimeResponse"`
	SystemDateAndTime SystemDateAndTime `xml:"http://www.onvif.org/ver10/device/wsdl SystemDateAndTime"`
}

// SystemDateAndTime represents system date and time.
type SystemDateAndTime struct {
	DateTimeType    string   `xml:"http://www.onvif.org/ver10/schema DateTimeType"`
	DaylightSavings bool     `xml:"http://www.onvif.org/ver10/schema DaylightSavings"`
	TimeZone        TimeZone `xml:"http://www.onvif.org/ver10/schema TimeZone,omitempty"`
	UTCDateTime     DateTime `xml:"http://www.onvif.org/ver10/schema UTCDateTime,omitempty"`
	LocalDateTime   DateTime `xml:"http://www.onvif.org/ver10/schema LocalDateTime,omitempty"`
}

// TimeZone represents timezone information.
type TimeZone struct {
	TZ string `xml:"http://www.onvif.org/ver10/schema TZ"`
}

// DateTime represents date and time.
type DateTime struct {
	Time Time `xml:"http://www.onvif.org/ver10/schema Time"`
	Date Date `xml:"http://www.onvif.org/ver10/schema Date"`
}

// Time represents time components.
type Time struct {
	Hour   int `xml:"http://www.onvif.org/ver10/schema Hour"`
	Minute int `xml:"http://www.onvif.org/ver10/schema Minute"`
	Second int `xml:"http://www.onvif.org/ver10/schema Second"`
}

// Date represents date components.
type Date struct {
	Year  int `xml:"http://www.onvif.org/ver10/schema Year"`
	Month int `xml:"http://www.onvif.org/ver10/schema Month"`
	Day   int `xml:"http://www.onvif.org/ver10/schema Day"`
}

// ToDateTime converts time.Time to DateTime structure.
func ToDateTime(t time.Time) DateTime {
	return DateTime{
		Date: Date{
			Year:  t.Year(),
			Month: int(t.Month()),
			Day:   t.Day(),
		},
		Time: Time{
			Hour:   t.Hour(),
			Minute: t.Minute(),
			Second: t.Second(),
		},
	}
}

// GetCapabilitiesRequest represents GetCapabilities request.
type GetCapabilitiesRequest struct {
	XMLName  xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetCapabilities"`
	Category []string `xml:"Category,omitempty"`
}

// GetDeviceInformationRequest represents GetDeviceInformation request.
type GetDeviceInformationRequest struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetDeviceInformation"`
}

// GetServicesRequest represents GetServices request.
type GetServicesRequest struct {
	XMLName           xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetServices"`
	IncludeCapability bool     `xml:"IncludeCapability"`
}

// GetProfilesRequest represents GetProfiles request.
type GetProfilesRequest struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetProfiles"`
}

// GetStreamUriRequest represents GetStreamUri request.
type GetStreamUriRequest struct {
	XMLName      xml.Name    `xml:"http://www.onvif.org/ver10/media/wsdl GetStreamUri"`
	StreamSetup  StreamSetup `xml:"StreamSetup"`
	ProfileToken string      `xml:"ProfileToken"`
}

// StreamSetup represents stream setup parameters.
type StreamSetup struct {
	Stream    string    `xml:"Stream"`
	Transport Transport `xml:"Transport"`
}

// Transport represents transport parameters.
type Transport struct {
	Protocol string `xml:"Protocol"`
}

// GetSnapshotUriRequest represents GetSnapshotUri request.
type GetSnapshotUriRequest struct {
	XMLName      xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetSnapshotUri"`
	ProfileToken string   `xml:"ProfileToken"`
}
