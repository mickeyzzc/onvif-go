package device_test

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/device"
	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/onvif"
	"github.com/mickeyzzc/onvif-go/v2/types"
)

// WSDL fact: the SetDNS/SetNTP Manual wrappers are locally declared in the
// device WSDL (tds), but they are typed tt:IPAddress / tt:NetworkHost — so
// the Type/IPv4Address/IPv6Address/DNSname children resolve to
// ver10/schema. A strict (gSOAP) device drops manual entries whose fields
// arrive in tds or in no namespace.

type nsDNSManual struct {
	FromDHCP  bool `xml:"http://www.onvif.org/ver10/device/wsdl FromDHCP"`
	DNSManual []struct {
		Type        string `xml:"http://www.onvif.org/ver10/schema Type"`
		IPv4Address string `xml:"http://www.onvif.org/ver10/schema IPv4Address"`
	} `xml:"http://www.onvif.org/ver10/device/wsdl DNSManual"`
}

func TestSetDNSManualResolvesToSchemaNamespace(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(`<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope"><Body><SetDNSResponse/></Body></Envelope>`))
	}))
	defer srv.Close()

	c, err := onvif.NewClient(srv.URL, onvif.WithCredentials("u", "p"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.SetServiceEndpoint(api.ServiceDevice, srv.URL)

	err = c.Device().SetDNS(context.Background(), false, nil,
		[]types.IPAddress{{Type: "IPv4", IPv4Address: "10.1.1.53"}})
	if err != nil {
		t.Fatalf("SetDNS: %v", err)
	}

	var env struct {
		XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
		Body    struct {
			SetDNS nsDNSManual `xml:"http://www.onvif.org/ver10/device/wsdl SetDNS"`
		} `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
	}
	if err := xml.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode envelope: %v\nbody: %s", err, body)
	}

	if len(env.Body.SetDNS.DNSManual) != 1 {
		t.Fatalf("DNSManual count = %d, want 1 (element in tds)\nbody: %s", len(env.Body.SetDNS.DNSManual), body)
	}

	m := env.Body.SetDNS.DNSManual[0]
	if m.Type != "IPv4" || m.IPv4Address != "10.1.1.53" {
		t.Errorf("DNSManual Type/IPv4Address not in ver10/schema:\nbody: %s", body)
	}
}

type nsNTPManual struct {
	FromDHCP  bool `xml:"http://www.onvif.org/ver10/device/wsdl FromDHCP"`
	NTPManual []struct {
		Type        string `xml:"http://www.onvif.org/ver10/schema Type"`
		IPv4Address string `xml:"http://www.onvif.org/ver10/schema IPv4Address"`
	} `xml:"http://www.onvif.org/ver10/device/wsdl NTPManual"`
}

func TestSetNTPManualResolvesToSchemaNamespace(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(`<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope"><Body><SetNTPResponse/></Body></Envelope>`))
	}))
	defer srv.Close()

	c, err := onvif.NewClient(srv.URL, onvif.WithCredentials("u", "p"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.SetServiceEndpoint(api.ServiceDevice, srv.URL)

	err = c.Device().SetNTP(context.Background(), false,
		[]device.NetworkHost{{Type: "IPv4", IPv4Address: "10.0.0.9"}})
	if err != nil {
		t.Fatalf("SetNTP: %v", err)
	}

	var env struct {
		XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
		Body    struct {
			SetNTP nsNTPManual `xml:"http://www.onvif.org/ver10/device/wsdl SetNTP"`
		} `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
	}
	if err := xml.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode envelope: %v\nbody: %s", err, body)
	}

	if len(env.Body.SetNTP.NTPManual) != 1 {
		t.Fatalf("NTPManual count = %d, want 1 (element in tds)\nbody: %s", len(env.Body.SetNTP.NTPManual), body)
	}

	m := env.Body.SetNTP.NTPManual[0]
	if m.Type != "IPv4" || m.IPv4Address != "10.0.0.9" {
		t.Errorf("NTPManual Type/IPv4Address not in ver10/schema:\nbody: %s", body)
	}
}
