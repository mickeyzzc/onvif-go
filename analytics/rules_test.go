package analytics_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/analytics"
	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
)

// Issue #112: the Rules write path — CreateRules / ModifyRules /
// DeleteRules + the GetRuleOptions / GetAnalyticsModuleOptions probes.

func TestRulesWriteOpsEncodeConfigurations(t *testing.T) {
	type capture struct {
		action string
		body   string
	}
	var calls []capture
	svc := analytics.New(testutil.NewFakeCaller("http://fake/analytics", func(action, reqXML string) (string, error) {
		calls = append(calls, capture{action, reqXML})
		switch action {
		case "tan:CreateRules", "tan:ModifyRules":
			return "<" + strings.TrimPrefix(action, "tan:") + "Response/>", nil
		case "tan:DeleteRules":
			return "<DeleteRulesResponse/>", nil
		default:
			return "", errors.New("unexpected action " + action)
		}
	}))

	rule := &analytics.Config{
		Name: "LineDetector_1",
		Type: "tt:LineDetector",
		Parameters: []analytics.SimpleItem{
			{Name: "Sensitivity", Value: "0.6"},
		},
	}

	if err := svc.CreateRules(context.Background(), "vac-1", []*analytics.Config{rule}); err != nil {
		t.Fatalf("CreateRules: %v", err)
	}
	if err := svc.ModifyRules(context.Background(), "vac-1", []*analytics.Config{rule}); err != nil {
		t.Fatalf("ModifyRules: %v", err)
	}
	if err := svc.DeleteRules(context.Background(), "vac-1", []string{"LineDetector_1"}); err != nil {
		t.Fatalf("DeleteRules: %v", err)
	}

	if len(calls) != 3 {
		t.Fatalf("calls = %d, want 3", len(calls))
	}

	create := calls[0]
	if create.action != "tan:CreateRules" {
		t.Errorf("create action = %q", create.action)
	}
	for _, want := range []string{
		"<tan:ConfigurationToken>vac-1</tan:ConfigurationToken>",
		`<tan:Rule Name="LineDetector_1" Type="tt:LineDetector">`,
		`<tt:SimpleItem Name="Sensitivity" Value="0.6">`,
	} {
		if !strings.Contains(create.body, want) {
			t.Errorf("CreateRules request missing %q:\n%s", want, create.body)
		}
	}

	modify := calls[1]
	if modify.action != "tan:ModifyRules" || !strings.Contains(modify.body, `<tan:Rule Name="LineDetector_1"`) {
		t.Errorf("ModifyRules wire wrong:\n%s", modify.body)
	}

	del := calls[2]
	if del.action != "tan:DeleteRules" || !strings.Contains(del.body, "<tan:RuleName>LineDetector_1</tan:RuleName>") {
		t.Errorf("DeleteRules wire wrong:\n%s", del.body)
	}
}

func TestGetRuleOptionsParsesOptions(t *testing.T) {
	svc := analytics.New(testutil.NewFakeCaller("http://fake/analytics", func(action, reqXML string) (string, error) {
		if action != "tan:GetRuleOptions" {
			return "", errors.New("unexpected action " + action)
		}
		if !strings.Contains(reqXML, "tt:LineDetector") {
			t.Errorf("request missing rule type filter:\n%s", reqXML)
		}

		return `<GetRuleOptionsResponse>
	<RuleOptions RuleType="tt:LineDetector" Name="Sensitivity" Type="tt:FloatItem">
		<tt:FloatItem><Min>0</Min><Max>1</Max></tt:FloatItem>
	</RuleOptions>
</GetRuleOptionsResponse>`, nil
	}))

	opts, err := svc.GetRuleOptions(context.Background(), "tt:LineDetector", "vac-1")
	if err != nil {
		t.Fatalf("GetRuleOptions: %v", err)
	}

	if len(opts) != 1 {
		t.Fatalf("options = %d, want 1", len(opts))
	}
	if opts[0].RuleType != "tt:LineDetector" || opts[0].Name != "Sensitivity" {
		t.Errorf("option = %+v", opts[0])
	}
	if !strings.Contains(opts[0].InnerXML, "FloatItem") {
		t.Errorf("InnerXML = %q, want the vendor option tree", opts[0].InnerXML)
	}
}

func TestGetAnalyticsModuleOptionsParsesOptions(t *testing.T) {
	svc := analytics.New(testutil.NewFakeCaller("http://fake/analytics", func(action, _ string) (string, error) {
		if action != "tan:GetAnalyticsModuleOptions" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetAnalyticsModuleOptionsResponse>
	<Options AnalyticsModule="tt:CellMotionDetector" Name="Sensitivity" Type="tt:IntItem">
		<tt:IntItem><Min>0</Min><Max>100</Max></tt:IntItem>
	</Options>
</GetAnalyticsModuleOptionsResponse>`, nil
	}))

	opts, err := svc.GetAnalyticsModuleOptions(context.Background(), "", "vac-1")
	if err != nil {
		t.Fatalf("GetAnalyticsModuleOptions: %v", err)
	}

	if len(opts) != 1 {
		t.Fatalf("options = %d, want 1", len(opts))
	}
	if opts[0].AnalyticsModule != "tt:CellMotionDetector" || opts[0].Name != "Sensitivity" {
		t.Errorf("option = %+v", opts[0])
	}
}
