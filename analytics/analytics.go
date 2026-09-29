package analytics

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/metadata"
)

// Request wire shapes. tan: covers elements locally declared in the
// analytics WSDL; tt: covers children of the schema types (tt:Config's
// Parameters/SimpleItem tree), with xmlns:tt declared on the root — the
// post-#90 request convention.

type getServiceCapabilities struct {
	XMLName xml.Name `xml:"tan:GetServiceCapabilities"`
	Xmlns   string   `xml:"xmlns:tan,attr"`
}

type getSupported struct {
	XMLName            xml.Name `xml:"tan:GetSupportedAnalyticsModules"`
	Xmlns              string   `xml:"xmlns:tan,attr"`
	ConfigurationToken string   `xml:"tan:ConfigurationToken"`
}

type getSupportedRules struct {
	XMLName            xml.Name `xml:"tan:GetSupportedRules"`
	Xmlns              string   `xml:"xmlns:tan,attr"`
	ConfigurationToken string   `xml:"tan:ConfigurationToken"`
}

type getAnalyticsModules struct {
	XMLName            xml.Name `xml:"tan:GetAnalyticsModules"`
	Xmlns              string   `xml:"xmlns:tan,attr"`
	ConfigurationToken string   `xml:"tan:ConfigurationToken"`
}

type getRules struct {
	XMLName            xml.Name `xml:"tan:GetRules"`
	Xmlns              string   `xml:"xmlns:tan,attr"`
	ConfigurationToken string   `xml:"tan:ConfigurationToken"`
}

type createAnalyticsModules struct {
	XMLName            xml.Name    `xml:"tan:CreateAnalyticsModules"`
	Xmlns              string      `xml:"xmlns:tan,attr"`
	XmlnsTT            string      `xml:"xmlns:tt,attr"`
	ConfigurationToken string      `xml:"tan:ConfigurationToken"`
	AnalyticsModule    []configOut `xml:"tan:AnalyticsModule"`
}

type modifyAnalyticsModules struct {
	XMLName            xml.Name    `xml:"tan:ModifyAnalyticsModules"`
	Xmlns              string      `xml:"xmlns:tan,attr"`
	XmlnsTT            string      `xml:"xmlns:tt,attr"`
	ConfigurationToken string      `xml:"tan:ConfigurationToken"`
	AnalyticsModule    []configOut `xml:"tan:AnalyticsModule"`
}

type deleteAnalyticsModules struct {
	XMLName             xml.Name `xml:"tan:DeleteAnalyticsModules"`
	Xmlns               string   `xml:"xmlns:tan,attr"`
	ConfigurationToken  string   `xml:"tan:ConfigurationToken"`
	AnalyticsModuleName []string `xml:"tan:AnalyticsModuleName"`
}

// configOut is the tt:Config serialization shape: Parameters/SimpleItem
// are schema-typed children, so they carry tt:.
type configOut struct {
	Name       string       `xml:"Name,attr"`
	Type       string       `xml:"Type,attr"`
	Parameters parameterOut `xml:"tt:Parameters"`
}

type parameterOut struct {
	SimpleItems []SimpleItem `xml:"tt:SimpleItem"`
}

// GetServiceCapabilities returns the analytics service capabilities.
func (s *Service) GetServiceCapabilities(ctx context.Context) (*Capabilities, error) {
	type response struct {
		XMLName      xml.Name `xml:"GetServiceCapabilitiesResponse"`
		Capabilities struct {
			RuleSupport            bool `xml:"RuleSupport,attr"`
			AnalyticsModuleSupport bool `xml:"AnalyticsModuleSupport,attr"`
		} `xml:"Capabilities"`
	}

	var resp response
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceAnalytics), "",
		&getServiceCapabilities{Xmlns: Namespace}, &resp); err != nil {
		return nil, fmt.Errorf("GetServiceCapabilities failed: %w", err)
	}

	return &Capabilities{
		RuleSupport:            resp.Capabilities.RuleSupport,
		AnalyticsModuleSupport: resp.Capabilities.AnalyticsModuleSupport,
	}, nil
}

// GetSupportedAnalyticsModules returns the module types a configuration
// accepts, with their parameters.
func (s *Service) GetSupportedAnalyticsModules(ctx context.Context, configurationToken string) (*Supported, error) {
	type response struct {
		XMLName   xml.Name `xml:"GetSupportedAnalyticsModulesResponse"`
		Supported struct {
			ContentSchemaLocations []string `xml:"AnalyticsModuleContentSchemaLocation"`
			Descriptions           []struct {
				Name       string `xml:"Name,attr"`
				Parameters []struct {
					Name string `xml:"Name,attr"`
					Type string `xml:"Type,attr"`
				} `xml:"Parameters>Item"`
			} `xml:"AnalyticsModuleDescription"`
		} `xml:"SupportedAnalyticsModules"`
	}

	var resp response
	req := &getSupported{Xmlns: Namespace, ConfigurationToken: configurationToken}
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceAnalytics), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetSupportedAnalyticsModules failed: %w", err)
	}

	return toSupported(resp.Supported.ContentSchemaLocations, resp.Supported.Descriptions), nil
}

// GetSupportedRules returns the rule types a configuration accepts.
func (s *Service) GetSupportedRules(ctx context.Context, configurationToken string) (*Supported, error) {
	type response struct {
		XMLName   xml.Name `xml:"GetSupportedRulesResponse"`
		Supported struct {
			ContentSchemaLocations []string `xml:"RuleContentSchemaLocation"`
			Descriptions           []struct {
				Name       string `xml:"Name,attr"`
				Parameters []struct {
					Name string `xml:"Name,attr"`
					Type string `xml:"Type,attr"`
				} `xml:"Parameters>Item"`
			} `xml:"RuleDescription"`
		} `xml:"SupportedRules"`
	}

	var resp response
	req := &getSupportedRules{Xmlns: Namespace, ConfigurationToken: configurationToken}
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceAnalytics), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetSupportedRules failed: %w", err)
	}

	return toSupported(resp.Supported.ContentSchemaLocations, resp.Supported.Descriptions), nil
}

func toSupported(locations []string, descriptions []struct {
	Name       string `xml:"Name,attr"`
	Parameters []struct {
		Name string `xml:"Name,attr"`
		Type string `xml:"Type,attr"`
	} `xml:"Parameters>Item"`
},
) *Supported {
	supported := &Supported{ContentSchemaLocations: locations}
	for _, d := range descriptions {
		desc := ConfigDescription{Name: d.Name}
		for _, p := range d.Parameters {
			desc.Parameters = append(desc.Parameters, ItemDescription{Name: p.Name, Type: p.Type})
		}

		supported.Descriptions = append(supported.Descriptions, desc)
	}

	return supported
}

// GetAnalyticsModules returns the analytics module instances of a
// configuration.
func (s *Service) GetAnalyticsModules(ctx context.Context, configurationToken string) ([]*Config, error) {
	type response struct {
		XMLName         xml.Name `xml:"GetAnalyticsModulesResponse"`
		AnalyticsModule []struct {
			Name       string `xml:"Name,attr"`
			Type       string `xml:"Type,attr"`
			Parameters struct {
				SimpleItems []SimpleItem `xml:"SimpleItem"`
			} `xml:"Parameters"`
		} `xml:"AnalyticsModule"`
	}

	var resp response
	req := &getAnalyticsModules{Xmlns: Namespace, ConfigurationToken: configurationToken}
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceAnalytics), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAnalyticsModules failed: %w", err)
	}

	configs := make([]*Config, 0, len(resp.AnalyticsModule))
	for _, m := range resp.AnalyticsModule {
		configs = append(configs, &Config{
			Name: m.Name, Type: m.Type, Parameters: m.Parameters.SimpleItems,
		})
	}

	return configs, nil
}

// GetRules returns the rule instances of a configuration.
func (s *Service) GetRules(ctx context.Context, configurationToken string) ([]*Config, error) {
	type response struct {
		XMLName xml.Name `xml:"GetRulesResponse"`
		Rule    []struct {
			Name       string `xml:"Name,attr"`
			Type       string `xml:"Type,attr"`
			Parameters struct {
				SimpleItems []SimpleItem `xml:"SimpleItem"`
			} `xml:"Parameters"`
		} `xml:"Rule"`
	}

	var resp response
	req := &getRules{Xmlns: Namespace, ConfigurationToken: configurationToken}
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceAnalytics), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetRules failed: %w", err)
	}

	configs := make([]*Config, 0, len(resp.Rule))
	for _, r := range resp.Rule {
		configs = append(configs, &Config{
			Name: r.Name, Type: r.Type, Parameters: r.Parameters.SimpleItems,
		})
	}

	return configs, nil
}

// CreateAnalyticsModules adds analytics module instances to a
// configuration.
func (s *Service) CreateAnalyticsModules(ctx context.Context, configurationToken string, modules []*Config) error {
	req := &createAnalyticsModules{
		Xmlns:              Namespace,
		XmlnsTT:            SchemaNamespace,
		ConfigurationToken: configurationToken,
	}
	for _, m := range modules {
		req.AnalyticsModule = append(req.AnalyticsModule, toConfigOut(m))
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceAnalytics), "", req, nil); err != nil {
		return fmt.Errorf("CreateAnalyticsModules failed: %w", err)
	}

	return nil
}

// ModifyAnalyticsModules updates analytics module instances of a
// configuration.
func (s *Service) ModifyAnalyticsModules(ctx context.Context, configurationToken string, modules []*Config) error {
	req := &modifyAnalyticsModules{
		Xmlns:              Namespace,
		XmlnsTT:            SchemaNamespace,
		ConfigurationToken: configurationToken,
	}
	for _, m := range modules {
		req.AnalyticsModule = append(req.AnalyticsModule, toConfigOut(m))
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceAnalytics), "", req, nil); err != nil {
		return fmt.Errorf("ModifyAnalyticsModules failed: %w", err)
	}

	return nil
}

// DeleteAnalyticsModules removes analytics modules by name.
func (s *Service) DeleteAnalyticsModules(ctx context.Context, configurationToken string, names []string) error {
	req := &deleteAnalyticsModules{
		Xmlns:               Namespace,
		ConfigurationToken:  configurationToken,
		AnalyticsModuleName: names,
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceAnalytics), "", req, nil); err != nil {
		return fmt.Errorf("DeleteAnalyticsModules failed: %w", err)
	}

	return nil
}

func toConfigOut(c *Config) configOut {
	return configOut{Name: c.Name, Type: c.Type, Parameters: parameterOut{SimpleItems: c.Parameters}}
}

// getSupportedMetadata is the GetSupportedMetadata request (Type is an
// optional AnalyticsModule type QName).
type getSupportedMetadata struct {
	XMLName xml.Name `xml:"tan:GetSupportedMetadata"`
	Xmlns   string   `xml:"xmlns:tan,attr"`
	Type    string   `xml:"tan:Type,omitempty"`
}

// MetadataInfo is one tan:MetadataInfo answer: the sample frame
// describing the metadata coordinate system.
type MetadataInfo struct {
	SampleFrame *metadata.Frame
}

// GetSupportedMetadata returns the metadata descriptions (sample frames)
// of the analytics modules.
func (s *Service) GetSupportedMetadata(ctx context.Context, moduleType string) ([]*MetadataInfo, error) {
	type response struct {
		XMLName         xml.Name `xml:"GetSupportedMetadataResponse"`
		AnalyticsModule []struct {
			SampleFrame *metadata.Frame `xml:"SampleFrame"`
		} `xml:"AnalyticsModule"`
	}

	var resp response
	req := &getSupportedMetadata{Xmlns: Namespace, Type: moduleType}
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceAnalytics), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetSupportedMetadata failed: %w", err)
	}

	infos := make([]*MetadataInfo, 0, len(resp.AnalyticsModule))
	for _, m := range resp.AnalyticsModule {
		infos = append(infos, &MetadataInfo{SampleFrame: m.SampleFrame})
	}

	return infos, nil
}

// RuleOptions is one tan:ConfigOptions entry of GetRuleOptions — the
// vendor parameter tree is preserved verbatim in InnerXML for callers
// that need the typed ranges.
type RuleOptions struct {
	RuleType string
	Name     string
	Type     string
	InnerXML string
}

// ModuleOptions is one tan:ConfigOptions entry of GetAnalyticsModuleOptions.
type ModuleOptions struct {
	AnalyticsModule string
	Name            string
	Type            string
	InnerXML        string
}

type createRules struct {
	XMLName            xml.Name    `xml:"tan:CreateRules"`
	Xmlns              string      `xml:"xmlns:tan,attr"`
	XmlnsTT            string      `xml:"xmlns:tt,attr"`
	ConfigurationToken string      `xml:"tan:ConfigurationToken"`
	Rule               []configOut `xml:"tan:Rule"`
}

type modifyRules struct {
	XMLName            xml.Name    `xml:"tan:ModifyRules"`
	Xmlns              string      `xml:"xmlns:tan,attr"`
	XmlnsTT            string      `xml:"xmlns:tt,attr"`
	ConfigurationToken string      `xml:"tan:ConfigurationToken"`
	Rule               []configOut `xml:"tan:Rule"`
}

type deleteRules struct {
	XMLName            xml.Name `xml:"tan:DeleteRules"`
	Xmlns              string   `xml:"xmlns:tan,attr"`
	ConfigurationToken string   `xml:"tan:ConfigurationToken"`
	RuleName           []string `xml:"tan:RuleName"`
}

// CreateRules adds rules to a VideoAnalyticsConfiguration (e.g. Hikvision
// / Axis MotionDetector / FieldDetector provisioning).
func (s *Service) CreateRules(ctx context.Context, configurationToken string, rules []*Config) error {
	req := &createRules{
		Xmlns:              Namespace,
		XmlnsTT:            SchemaNamespace,
		ConfigurationToken: configurationToken,
	}
	for _, r := range rules {
		req.Rule = append(req.Rule, toConfigOut(r))
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceAnalytics), "", req, nil); err != nil {
		return fmt.Errorf("CreateRules failed: %w", err)
	}

	return nil
}

// ModifyRules updates existing rules of a VideoAnalyticsConfiguration.
func (s *Service) ModifyRules(ctx context.Context, configurationToken string, rules []*Config) error {
	req := &modifyRules{
		Xmlns:              Namespace,
		XmlnsTT:            SchemaNamespace,
		ConfigurationToken: configurationToken,
	}
	for _, r := range rules {
		req.Rule = append(req.Rule, toConfigOut(r))
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceAnalytics), "", req, nil); err != nil {
		return fmt.Errorf("ModifyRules failed: %w", err)
	}

	return nil
}

// DeleteRules removes named rules from a VideoAnalyticsConfiguration.
func (s *Service) DeleteRules(ctx context.Context, configurationToken string, ruleNames []string) error {
	req := &deleteRules{
		Xmlns:              Namespace,
		ConfigurationToken: configurationToken,
		RuleName:           ruleNames,
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceAnalytics), "", req, nil); err != nil {
		return fmt.Errorf("DeleteRules failed: %w", err)
	}

	return nil
}

// GetRuleOptions returns the configurable parameter ranges for a rule
// type (empty ruleType = all types; each entry then carries RuleType).
func (s *Service) GetRuleOptions(ctx context.Context, ruleType, configurationToken string) ([]*RuleOptions, error) {
	type getRuleOptions struct {
		XMLName            xml.Name `xml:"tan:GetRuleOptions"`
		Xmlns              string   `xml:"xmlns:tan,attr"`
		RuleType           string   `xml:"tan:RuleType,omitempty"`
		ConfigurationToken string   `xml:"tan:ConfigurationToken"`
	}

	type response struct {
		XMLName xml.Name `xml:"GetRuleOptionsResponse"`
		Options []struct {
			RuleType string `xml:"RuleType,attr"`
			Name     string `xml:"Name,attr"`
			Type     string `xml:"Type,attr"`
			Inner    string `xml:",innerxml"`
		} `xml:"RuleOptions"`
	}

	req := &getRuleOptions{
		Xmlns:              Namespace,
		RuleType:           ruleType,
		ConfigurationToken: configurationToken,
	}
	var resp response
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceAnalytics), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetRuleOptions failed: %w", err)
	}

	out := make([]*RuleOptions, 0, len(resp.Options))
	for _, o := range resp.Options {
		out = append(out, &RuleOptions{RuleType: o.RuleType, Name: o.Name, Type: o.Type, InnerXML: o.Inner})
	}

	return out, nil
}

// GetAnalyticsModuleOptions returns the configurable parameter ranges
// for an analytics module type (empty moduleType = all types).
func (s *Service) GetAnalyticsModuleOptions(ctx context.Context, moduleType, configurationToken string) ([]*ModuleOptions, error) {
	type getAnalyticsModuleOptions struct {
		XMLName            xml.Name `xml:"tan:GetAnalyticsModuleOptions"`
		Xmlns              string   `xml:"xmlns:tan,attr"`
		Type               string   `xml:"tan:Type,omitempty"`
		ConfigurationToken string   `xml:"tan:ConfigurationToken"`
	}

	type response struct {
		XMLName xml.Name `xml:"GetAnalyticsModuleOptionsResponse"`
		Options []struct {
			AnalyticsModule string `xml:"AnalyticsModule,attr"`
			Name            string `xml:"Name,attr"`
			Type            string `xml:"Type,attr"`
			Inner           string `xml:",innerxml"`
		} `xml:"Options"`
	}

	req := &getAnalyticsModuleOptions{
		Xmlns:              Namespace,
		Type:               moduleType,
		ConfigurationToken: configurationToken,
	}
	var resp response
	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceAnalytics), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetAnalyticsModuleOptions failed: %w", err)
	}

	out := make([]*ModuleOptions, 0, len(resp.Options))
	for _, o := range resp.Options {
		out = append(out, &ModuleOptions{AnalyticsModule: o.AnalyticsModule, Name: o.Name, Type: o.Type, InnerXML: o.Inner})
	}

	return out, nil
}
