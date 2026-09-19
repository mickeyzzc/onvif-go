package analytics_test

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/analytics"
	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	onvif "github.com/mickeyzzc/onvif-go/v2/onvif"
)

// WSDL facts (ver20 analytics + ver10 schema): the Create/Modify payload
// wraps tt:Config instances in tan-locally-declared AnalyticsModule
// elements; the Config children (Parameters/SimpleItem) are
// ver10/schema. Strict devices drop modules whose parameters resolve into
// the wrong namespace.
func TestCreateAnalyticsModulesResolvesToSchemaNamespace(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(`<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope"><Body><CreateAnalyticsModulesResponse/></Body></Envelope>`))
	}))
	defer srv.Close()

	c, err := onvif.NewClient(srv.URL, onvif.WithCredentials("u", "p"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.SetServiceEndpoint(api.ServiceAnalytics, srv.URL)

	err = c.Analytics().CreateAnalyticsModules(context.Background(), "video_analytics_1", []*analytics.Config{{
		Name: "Detector1", Type: "tt:MyDetector",
		Parameters: []analytics.SimpleItem{{Name: "Sensitivity", Value: "0.5"}},
	}})
	if err != nil {
		t.Fatalf("CreateAnalyticsModules: %v", err)
	}

	var env struct {
		XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
		Body    struct {
			Create struct {
				Token  string `xml:"http://www.onvif.org/ver20/analytics/wsdl ConfigurationToken"`
				Module []struct {
					Name       string `xml:"Name,attr"`
					Parameters struct {
						SimpleItems []struct {
							Name  string `xml:"Name,attr"`
							Value string `xml:"Value,attr"`
						} `xml:"http://www.onvif.org/ver10/schema SimpleItem"`
					} `xml:"http://www.onvif.org/ver10/schema Parameters"`
				} `xml:"http://www.onvif.org/ver20/analytics/wsdl AnalyticsModule"`
			} `xml:"http://www.onvif.org/ver20/analytics/wsdl CreateAnalyticsModules"`
		} `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
	}
	if err := xml.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode envelope: %v\nbody: %s", err, body)
	}

	create := env.Body.Create
	if create.Token != "video_analytics_1" || len(create.Module) != 1 {
		t.Fatalf("token/modules not in tan:\nbody: %s", body)
	}

	module := create.Module[0]
	if module.Name != "Detector1" ||
		len(module.Parameters.SimpleItems) != 1 ||
		module.Parameters.SimpleItems[0].Name != "Sensitivity" ||
		module.Parameters.SimpleItems[0].Value != "0.5" {
		t.Errorf("Config parameters not in ver10/schema:\nbody: %s", body)
	}
}
