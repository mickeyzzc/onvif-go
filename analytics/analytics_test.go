package analytics_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/analytics"
	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
)

func TestGetServiceCapabilities(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/analytics", func(action, _ string) (string, error) {
		if action != "tan:GetServiceCapabilities" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetServiceCapabilitiesResponse><Capabilities RuleSupport="true" AnalyticsModuleSupport="true"/></GetServiceCapabilitiesResponse>`, nil
	})

	caps, err := analytics.New(caller).GetServiceCapabilities(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if !caps.RuleSupport || !caps.AnalyticsModuleSupport {
		t.Errorf("capabilities = %+v, want both true", caps)
	}
}

func TestGetSupportedAnalyticsModulesAndInstances(t *testing.T) {
	cases := []struct {
		name   string
		action string
		resp   string
		invoke func(s *analytics.Service) error
	}{
		{
			name:   "supported modules",
			action: "tan:GetSupportedAnalyticsModules",
			resp: `<GetSupportedAnalyticsModulesResponse><SupportedAnalyticsModules>
				<AnalyticsModuleContentSchemaLocation>http://example/schema.xsd</AnalyticsModuleContentSchemaLocation>
				<AnalyticsModuleDescription Name="tt:MyDetector">
					<Parameters>
						<Item Name="Sensitivity" Type="xs:float"/>
						<Item Name="Zone" Type="tt:Polygon"/>
					</Parameters>
				</AnalyticsModuleDescription>
				</SupportedAnalyticsModules></GetSupportedAnalyticsModulesResponse>`,
			invoke: func(s *analytics.Service) error {
				supported, err := s.GetSupportedAnalyticsModules(context.Background(), "video_analytics_1")
				if err != nil {
					return err
				}

				if len(supported.ContentSchemaLocations) != 1 {
					t.Errorf("schema locations = %v", supported.ContentSchemaLocations)
				}

				if len(supported.Descriptions) != 1 ||
					supported.Descriptions[0].Name != "tt:MyDetector" ||
					len(supported.Descriptions[0].Parameters) != 2 ||
					supported.Descriptions[0].Parameters[1].Name != "Zone" {
					t.Errorf("descriptions = %+v", supported.Descriptions)
				}

				return nil
			},
		},
		{
			name:   "supported rules",
			action: "tan:GetSupportedRules",
			resp: `<GetSupportedRulesResponse><SupportedRules>
				<RuleDescription Name="tt:LineDetector"><Parameters><Item Name="Direction" Type="xs:string"/></Parameters></RuleDescription>
				</SupportedRules></GetSupportedRulesResponse>`,
			invoke: func(s *analytics.Service) error {
				supported, err := s.GetSupportedRules(context.Background(), "video_analytics_1")
				if err != nil {
					return err
				}

				if len(supported.Descriptions) != 1 || supported.Descriptions[0].Name != "tt:LineDetector" {
					t.Errorf("descriptions = %+v", supported.Descriptions)
				}

				return nil
			},
		},
		{
			name:   "module instances",
			action: "tan:GetAnalyticsModules",
			resp: `<GetAnalyticsModulesResponse>
				<AnalyticsModule Name="Detector1" Type="tt:MyDetector">
					<Parameters>
						<SimpleItem Name="Sensitivity" Value="0.5"/>
						<SimpleItem Name="Zone" Value="polygon-data"/>
					</Parameters>
				</AnalyticsModule>
				</GetAnalyticsModulesResponse>`,
			invoke: func(s *analytics.Service) error {
				modules, err := s.GetAnalyticsModules(context.Background(), "video_analytics_1")
				if err != nil {
					return err
				}

				if len(modules) != 1 || modules[0].Name != "Detector1" || modules[0].Type != "tt:MyDetector" {
					t.Errorf("modules = %+v", modules)
				}

				if len(modules[0].Parameters) != 2 || modules[0].Parameters[0].Value != "0.5" {
					t.Errorf("parameters = %+v", modules[0].Parameters)
				}

				return nil
			},
		},
		{
			name:   "rule instances",
			action: "tan:GetRules",
			resp:   `<GetRulesResponse><Rule Name="Line1" Type="tt:LineDetector"><Parameters><SimpleItem Name="Direction" Value="Both"/></Parameters></Rule></GetRulesResponse>`,
			invoke: func(s *analytics.Service) error {
				rules, err := s.GetRules(context.Background(), "video_analytics_1")
				if err != nil {
					return err
				}

				if len(rules) != 1 || rules[0].Name != "Line1" || len(rules[0].Parameters) != 1 {
					t.Errorf("rules = %+v", rules)
				}

				return nil
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := strings.TrimPrefix(tc.action, "tan:")
			caller := testutil.NewFakeCaller("http://fake/analytics", func(action, _ string) (string, error) {
				if strings.TrimPrefix(action, "tan:") != want {
					return "", errors.New("unexpected action " + action)
				}

				return tc.resp, nil
			})

			if err := tc.invoke(analytics.New(caller)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAnalyticsWriteOps(t *testing.T) {
	steps := []struct {
		action string
		resp   string
		check  func(reqXML string)
	}{
		{
			action: "tan:CreateAnalyticsModules",
			resp:   `<CreateAnalyticsModulesResponse/>`,
			check: func(reqXML string) {
				if !strings.Contains(reqXML, `<tan:AnalyticsModule Name="Detector1" Type="tt:MyDetector">`) {
					t.Errorf("create payload missing module attrs:\n%s", reqXML)
				}

				if !strings.Contains(reqXML, `<tt:SimpleItem Name="Sensitivity" Value="0.5"/>`) &&
					!strings.Contains(reqXML, `<tt:SimpleItem Name="Sensitivity" Value="0.5"></tt:SimpleItem>`) {
					t.Errorf("create payload missing tt:SimpleItem:\n%s", reqXML)
				}
			},
		},
		{
			action: "tan:ModifyAnalyticsModules",
			resp:   `<ModifyAnalyticsModulesResponse/>`,
			check: func(reqXML string) {
				if !strings.Contains(reqXML, `Type="tt:MyDetector"`) {
					t.Errorf("modify payload missing module:\n%s", reqXML)
				}
			},
		},
		{
			action: "tan:DeleteAnalyticsModules",
			resp:   `<DeleteAnalyticsModulesResponse/>`,
			check: func(reqXML string) {
				if !strings.Contains(reqXML, "<tan:AnalyticsModuleName>Detector1</tan:AnalyticsModuleName>") {
					t.Errorf("delete payload missing name:\n%s", reqXML)
				}
			},
		},
	}

	step := 0
	caller := testutil.NewFakeCaller("http://fake/analytics", func(action, reqXML string) (string, error) {
		if step >= len(steps) || action != steps[step].action {
			return "", errors.New("unexpected action " + action)
		}

		resp := steps[step].resp
		steps[step].check(reqXML)
		step++

		return resp, nil
	})

	s := analytics.New(caller)
	ctx := context.Background()

	if err := s.CreateAnalyticsModules(ctx, "video_analytics_1", []*analytics.Config{{
		Name: "Detector1", Type: "tt:MyDetector",
		Parameters: []analytics.SimpleItem{{Name: "Sensitivity", Value: "0.5"}},
	}}); err != nil {
		t.Fatal(err)
	}

	if err := s.ModifyAnalyticsModules(ctx, "video_analytics_1", []*analytics.Config{{
		Name: "Detector1", Type: "tt:MyDetector",
		Parameters: []analytics.SimpleItem{{Name: "Sensitivity", Value: "0.7"}},
	}}); err != nil {
		t.Fatal(err)
	}

	if err := s.DeleteAnalyticsModules(ctx, "video_analytics_1", []string{"Detector1"}); err != nil {
		t.Fatal(err)
	}
}
