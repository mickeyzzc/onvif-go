// Package analytics covers the ONVIF ver20 analytics service (tan): the
// rule and analytics-module configuration model behind Profile M. Wire
// contracts follow the official WSDL set — tan-locally-declared elements
// carry the service namespace, tt:Config-typed children the schema
// namespace.
package analytics

import "github.com/mickeyzzc/onvif-go/v2/internal/api"

// Namespace is the ver20 analytics service WSDL namespace (tan).
const Namespace = "http://www.onvif.org/ver20/analytics/wsdl"

// SchemaNamespace is the ver10 schema namespace (tt) — the namespace of
// Config/ItemList-typed payload children.
const SchemaNamespace = "http://www.onvif.org/ver10/schema"

// Service is the analytics service client.
type Service struct {
	c api.Caller
}

// New creates an analytics service bound to a caller.
func New(c api.Caller) *Service { return &Service{c: c} }

// SimpleItem is one tt:ItemList SimpleItem (name/value attribute pair).
type SimpleItem struct {
	Name  string `xml:"Name,attr"`
	Value string `xml:"Value,attr"`
}

// Config is one tt:Config — an analytics module or rule instance: a name,
// its type QName, and the parameter values.
type Config struct {
	Name       string
	Type       string
	Parameters []SimpleItem
}

// ItemDescription describes one configurable parameter of a Config.
type ItemDescription struct {
	Name string
	Type string
}

// ConfigDescription is one tt:ConfigDescription — the description of a
// supported module or rule type: its name and the parameters it accepts.
// Message descriptions are not modeled yet (metadata streaming follow-up).
type ConfigDescription struct {
	Name       string
	Parameters []ItemDescription
}

// Capabilities is the analytics service capability answer.
type Capabilities struct {
	RuleSupport            bool
	AnalyticsModuleSupport bool
}

// Supported is the GetSupported{Rules,AnalyticsModules} answer: schema
// locations plus one description per supported type.
type Supported struct {
	ContentSchemaLocations []string
	Descriptions           []ConfigDescription
}
