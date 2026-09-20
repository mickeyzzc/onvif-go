// Example: Profile M analytics — rules and analytics modules.
//
// onvif.Client.Analytics() is the ver20 analytics service client
// (v2.1.0): capabilities, the supported rule/module types, and the
// configured rule/module instances of one VideoAnalytics configuration.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/onvif"
)

func main() {
	endpoint := flag.String("endpoint", "http://192.0.2.100/onvif/device_service", "ONVIF device endpoint")
	username := flag.String("username", "admin", "ONVIF username")
	password := flag.String("password", "", "ONVIF password")
	// The analytics operations are scoped to one VideoAnalytics
	// configuration; its token is usually discoverable via the ver10
	// media service (GetCompatibleVideoAnalyticsConfigurations) —
	// cameras commonly name it VideoAnalytics_1 or similar.
	token := flag.String("token", "VideoAnalytics_1", "VideoAnalytics configuration token")
	flag.Parse()

	if *password == "" {
		log.Fatal("pass -password (see your camera's ONVIF credentials)")
	}

	client, err := onvif.NewClient(
		*endpoint,
		onvif.WithCredentials(*username, *password),
		onvif.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	_ = client.Initialize(ctx)

	// What the analytics service can do.
	caps, err := client.Analytics().GetServiceCapabilities(ctx)
	if err != nil {
		log.Fatalf("GetServiceCapabilities failed: %v", err)
	}
	fmt.Printf("Analytics capabilities: rule support=%v, module support=%v\n\n",
		caps.RuleSupport, caps.AnalyticsModuleSupport)

	// The rule types this camera can instantiate.
	ruleTypes, err := client.Analytics().GetSupportedRules(ctx, *token)
	if err != nil {
		log.Printf("GetSupportedRules failed: %v", err)
	} else {
		fmt.Printf("Supported rule types:\n")
		for _, d := range ruleTypes.Descriptions {
			fmt.Printf("  %s\n", d.Name)
		}
		fmt.Println()
	}

	// The analytics module types (detectors) this camera offers.
	moduleTypes, err := client.Analytics().GetSupportedAnalyticsModules(ctx, *token)
	if err != nil {
		log.Printf("GetSupportedAnalyticsModules failed: %v", err)
	} else {
		fmt.Printf("Supported analytics module types:\n")
		for _, d := range moduleTypes.Descriptions {
			fmt.Printf("  %s\n", d.Name)
		}
		fmt.Println()
	}

	// The configured rule instances and their current parameters.
	rules, err := client.Analytics().GetRules(ctx, *token)
	if err != nil {
		log.Printf("GetRules failed: %v", err)
	} else {
		fmt.Printf("Configured rules (%d):\n", len(rules))
		for _, r := range rules {
			fmt.Printf("  %q type=%s", r.Name, r.Type)
			for _, p := range r.Parameters {
				fmt.Printf(" %s=%s", p.Name, p.Value)
			}
			fmt.Println()
		}
	}

	// Same for analytics module instances (e.g. a motion detector with
	// its zones); Create/Modify/DeleteAnalyticsModules manage them.
	modules, err := client.Analytics().GetAnalyticsModules(ctx, *token)
	if err != nil {
		log.Printf("GetAnalyticsModules failed: %v", err)
	} else {
		fmt.Printf("\nConfigured analytics modules (%d):\n", len(modules))
		for _, m := range modules {
			fmt.Printf("  %q type=%s", m.Name, m.Type)
			for _, p := range m.Parameters {
				fmt.Printf(" %s=%s", p.Name, p.Value)
			}
			fmt.Println()
		}
	}
}
