package ptz_test

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/onvif"
	"github.com/mickeyzzc/onvif-go/v2/ptz"
)

// Issue #90: ver10/schema elements (PanTilt/Zoom inside Velocity/Position)
// were serialized unprefixed. The SOAP Body carries
// xmlns="…soap-envelope", so those elements resolved into the envelope
// namespace — strict devices (Dahua &co.) parse an empty Velocity and
// return 200 without moving. These tests capture the real envelope bytes
// over HTTP and decode them namespace-strictly: the vector fields only
// populate when the elements resolve to ver10/schema, no matter which
// prefix the wire happens to use.

const (
	envNS  = "http://www.w3.org/2003/05/soap-envelope"
	ptzNS  = "http://www.onvif.org/ver20/ptz/wsdl"
	schema = "http://www.onvif.org/ver10/schema"
)

// captureClient starts an httptest SOAP endpoint recording request bodies
// and returns a client wired to it.
func captureClient(t *testing.T) (*onvif.Client, *[]string) {
	t.Helper()
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(`<Envelope xmlns="` + envNS + `"><Body><ContinuousMoveResponse/></Body></Envelope>`))
	}))
	t.Cleanup(srv.Close)

	c, err := onvif.NewClient(srv.URL, onvif.WithCredentials("u", "p"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.SetServiceEndpoint(api.ServicePTZ, srv.URL)

	return c, &bodies
}

// Namespace-strict decoders: every element tag below demands the exact
// namespace, so values only populate when the request carries the
// ver10/schema bindings.

type nsPanTilt struct {
	X float64 `xml:"x,attr"`
}

type nsZoom struct {
	X float64 `xml:"x,attr"`
}

type nsVelocity struct {
	PanTilt nsPanTilt `xml:"http://www.onvif.org/ver10/schema PanTilt"`
	Zoom    nsZoom    `xml:"http://www.onvif.org/ver10/schema Zoom"`
}

type nsContinuousMoveEnvelope struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Body    struct {
		Move struct {
			ProfileToken string     `xml:"http://www.onvif.org/ver20/ptz/wsdl ProfileToken"`
			Velocity     nsVelocity `xml:"http://www.onvif.org/ver20/ptz/wsdl Velocity"`
		} `xml:"http://www.onvif.org/ver20/ptz/wsdl ContinuousMove"`
	} `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
}

type nsPosition struct {
	PanTilt nsPanTilt `xml:"http://www.onvif.org/ver10/schema PanTilt"`
	Zoom    nsZoom    `xml:"http://www.onvif.org/ver10/schema Zoom"`
}

type nsAbsoluteMoveEnvelope struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Body    struct {
		Move struct {
			Position nsPosition `xml:"http://www.onvif.org/ver20/ptz/wsdl Position"`
		} `xml:"http://www.onvif.org/ver20/ptz/wsdl AbsoluteMove"`
	} `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
}

func TestContinuousMoveVectorsResolveToSchemaNamespace(t *testing.T) {
	c, bodies := captureClient(t)
	timeout := "PT10S"

	err := c.PTZ().ContinuousMove(context.Background(), "Profile_1",
		&ptz.PTZSpeed{
			PanTilt: &ptz.Vector2D{X: 0.5, Y: 0},
			Zoom:    &ptz.Vector1D{X: 0.25},
		}, &timeout)
	if err != nil {
		t.Fatalf("ContinuousMove: %v", err)
	}

	body := (*bodies)[0]
	if !strings.Contains(body, "Velocity") {
		t.Fatalf("request body lacks Velocity: %s", body)
	}

	var env nsContinuousMoveEnvelope
	if err := xml.Unmarshal([]byte(body), &env); err != nil {
		t.Fatalf("decode envelope: %v\nbody: %s", err, body)
	}

	if env.Body.Move.Velocity.PanTilt.X != 0.5 {
		t.Errorf("PanTilt x = %v, want 0.5 (element not in ver10/schema namespace)\nbody: %s",
			env.Body.Move.Velocity.PanTilt.X, body)
	}
	if env.Body.Move.Velocity.Zoom.X != 0.25 {
		t.Errorf("Zoom x = %v, want 0.25 (element not in ver10/schema namespace)\nbody: %s",
			env.Body.Move.Velocity.Zoom.X, body)
	}
}

func TestAbsoluteMoveVectorsResolveToSchemaNamespace(t *testing.T) {
	c, bodies := captureClient(t)

	err := c.PTZ().AbsoluteMove(context.Background(), "Profile_1",
		&ptz.PTZVector{
			PanTilt: &ptz.Vector2D{X: 1, Y: -1},
			Zoom:    &ptz.Vector1D{X: 0.5},
		}, nil)
	if err != nil {
		t.Fatalf("AbsoluteMove: %v", err)
	}

	body := (*bodies)[0]
	var env nsAbsoluteMoveEnvelope
	if err := xml.Unmarshal([]byte(body), &env); err != nil {
		t.Fatalf("decode envelope: %v\nbody: %s", err, body)
	}

	if env.Body.Move.Position.PanTilt.X != 1 || env.Body.Move.Position.Zoom.X != 0.5 {
		t.Errorf("Position vectors = %+v, want pan 1 / zoom 0.5 (ver10/schema namespace)\nbody: %s",
			env.Body.Move.Position, body)
	}
}
