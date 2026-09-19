package security_test

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/onvif"
	"github.com/mickeyzzc/onvif-go/v2/security"
)

// Issue #90: LoadCertificateWithPrivateKey serialized CertificateID /
// Certificate / PrivateKey unprefixed — they resolved into the SOAP
// envelope namespace instead of ver10/schema and strict devices ignored
// the load.

type nsLoadCertEnvelope struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Body    struct {
		Load struct {
			Item struct {
				CertificateID string `xml:"http://www.onvif.org/ver10/schema CertificateID"`
			} `xml:"http://www.onvif.org/ver10/device/wsdl CertificateWithPrivateKey"`
		} `xml:"http://www.onvif.org/ver10/device/wsdl LoadCertificateWithPrivateKey"`
	} `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
}

func TestLoadCertificateWithPrivateKeyResolvesToSchemaNamespace(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(`<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope"><Body><LoadCertificateWithPrivateKeyResponse/></Body></Envelope>`))
	}))
	defer srv.Close()

	c, err := onvif.NewClient(srv.URL, onvif.WithCredentials("u", "p"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.SetServiceEndpoint(api.ServiceDevice, srv.URL)

	svc := c.Security()
	err = svc.LoadCertificateWithPrivateKey(
		context.Background(),
		[]*security.Certificate{{CertificateID: "cert-1"}},
		nil,
		[]string{"cert-1"},
	)
	if err != nil {
		t.Fatalf("LoadCertificateWithPrivateKey: %v", err)
	}

	var env nsLoadCertEnvelope
	if err := xml.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode envelope: %v\nbody: %s", err, body)
	}

	if env.Body.Load.Item.CertificateID != "cert-1" {
		t.Errorf("CertificateID = %q, want %q (element not in ver10/schema namespace)\nbody: %s",
			env.Body.Load.Item.CertificateID, "cert-1", body)
	}
}
