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

// WSDL fact: CreateUsers/SetUser carry `User` elements that are locally
// declared in the device WSDL (tds) but typed tt:User — so the
// Username/Password/UserLevel children resolve to ver10/schema, not tds.
// A strict (gSOAP) device drops users whose fields arrive in tds or in no
// namespace.

type nsUsersEnvelope struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Body    struct {
		Create struct {
			User []struct {
				Username  string `xml:"http://www.onvif.org/ver10/schema Username"`
				Password  string `xml:"http://www.onvif.org/ver10/schema Password"`
				UserLevel string `xml:"http://www.onvif.org/ver10/schema UserLevel"`
			} `xml:"http://www.onvif.org/ver10/device/wsdl User"`
		} `xml:"http://www.onvif.org/ver10/device/wsdl CreateUsers"`
	} `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
}

func TestCreateUsersResolvesToSchemaNamespace(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(`<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope"><Body><CreateUsersResponse/></Body></Envelope>`))
	}))
	defer srv.Close()

	c, err := onvif.NewClient(srv.URL, onvif.WithCredentials("u", "p"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.SetServiceEndpoint(api.ServiceDevice, srv.URL)

	err = c.Security().CreateUsers(context.Background(), []*security.User{{
		Username: "operator", Password: "s3cret", UserLevel: "User",
	}})
	if err != nil {
		t.Fatalf("CreateUsers: %v", err)
	}

	var env nsUsersEnvelope
	if err := xml.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode envelope: %v\nbody: %s", err, body)
	}

	if len(env.Body.Create.User) != 1 {
		t.Fatalf("User count = %d, want 1 (element in tds)\nbody: %s", len(env.Body.Create.User), body)
	}

	u := env.Body.Create.User[0]
	if u.Username != "operator" || u.Password != "s3cret" || u.UserLevel != "User" {
		t.Errorf("Username/Password/UserLevel not in ver10/schema:\nbody: %s", body)
	}
}

type nsSetUserEnvelope struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Body    struct {
		Set struct {
			User struct {
				Username  string `xml:"http://www.onvif.org/ver10/schema Username"`
				UserLevel string `xml:"http://www.onvif.org/ver10/schema UserLevel"`
			} `xml:"http://www.onvif.org/ver10/device/wsdl User"`
		} `xml:"http://www.onvif.org/ver10/device/wsdl SetUser"`
	} `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
}

func TestSetUserResolvesToSchemaNamespace(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(`<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope"><Body><SetUserResponse/></Body></Envelope>`))
	}))
	defer srv.Close()

	c, err := onvif.NewClient(srv.URL, onvif.WithCredentials("u", "p"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.SetServiceEndpoint(api.ServiceDevice, srv.URL)

	err = c.Security().SetUser(context.Background(), &security.User{
		Username: "operator", UserLevel: "Administrator",
	})
	if err != nil {
		t.Fatalf("SetUser: %v", err)
	}

	var env nsSetUserEnvelope
	if err := xml.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode envelope: %v\nbody: %s", err, body)
	}

	if env.Body.Set.User.Username != "operator" || env.Body.Set.User.UserLevel != "Administrator" {
		t.Errorf("Username/UserLevel not in ver10/schema:\nbody: %s", body)
	}
}
