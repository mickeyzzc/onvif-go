package imaging_test

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/imaging"
	"github.com/mickeyzzc/onvif-go/v2/onvif"
)

// Issue #90: ver10/schema elements inside timg:SetImagingSettings and
// timg:Move were serialized unprefixed, inheriting the SOAP envelope
// default namespace from Body — strict devices parse empty settings and
// acknowledge without acting. These tests capture the real envelope over
// HTTP and decode namespace-strictly.

type nsSetImagingSettingsEnvelope struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Body    struct {
		Set struct {
			Settings struct {
				Brightness *float64 `xml:"http://www.onvif.org/ver10/schema Brightness"`
				Contrast   *float64 `xml:"http://www.onvif.org/ver10/schema Contrast"`
			} `xml:"http://www.onvif.org/ver20/imaging/wsdl ImagingSettings"`
		} `xml:"http://www.onvif.org/ver20/imaging/wsdl SetImagingSettings"`
	} `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
}

type nsMoveEnvelope struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Body    struct {
		Move struct {
			Focus struct {
				Absolute struct {
					Position float64 `xml:"http://www.onvif.org/ver10/schema Position"`
				} `xml:"http://www.onvif.org/ver10/schema Absolute"`
			} `xml:"http://www.onvif.org/ver20/imaging/wsdl Focus"`
		} `xml:"http://www.onvif.org/ver20/imaging/wsdl Move"`
	} `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
}

func captureImagingClient(t *testing.T) (*onvif.Client, *[]string) {
	t.Helper()
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(`<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope"><Body><SetImagingSettingsResponse/></Body></Envelope>`))
	}))
	t.Cleanup(srv.Close)

	c, err := onvif.NewClient(srv.URL, onvif.WithCredentials("u", "p"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.SetServiceEndpoint(api.ServiceImaging, srv.URL)

	return c, &bodies
}

func TestSetImagingSettingsResolvesToSchemaNamespace(t *testing.T) {
	c, bodies := captureImagingClient(t)
	brightness, contrast := 60.0, 55.0

	err := c.Imaging().SetImagingSettings(context.Background(), "vs-1", &imaging.ImagingSettings{
		Brightness: &brightness,
		Contrast:   &contrast,
	}, false)
	if err != nil {
		t.Fatalf("SetImagingSettings: %v", err)
	}

	var env nsSetImagingSettingsEnvelope
	if err := xml.Unmarshal([]byte((*bodies)[0]), &env); err != nil {
		t.Fatalf("decode envelope: %v\nbody: %s", err, (*bodies)[0])
	}

	if env.Body.Set.Settings.Brightness == nil || *env.Body.Set.Settings.Brightness != 60 {
		t.Errorf("Brightness not in ver10/schema namespace: %+v\nbody: %s",
			env.Body.Set.Settings, (*bodies)[0])
	}
	if env.Body.Set.Settings.Contrast == nil || *env.Body.Set.Settings.Contrast != 55 {
		t.Errorf("Contrast not in ver10/schema namespace: %+v\nbody: %s",
			env.Body.Set.Settings, (*bodies)[0])
	}
}

func TestMoveFocusResolvesToSchemaNamespace(t *testing.T) {
	c, bodies := captureImagingClient(t)

	err := c.Imaging().Move(context.Background(), "vs-1", &imaging.FocusMove{
		Absolute: &imaging.AbsoluteFocus{Position: 0.7},
	})
	if err != nil {
		t.Fatalf("Move: %v", err)
	}

	var env nsMoveEnvelope
	if err := xml.Unmarshal([]byte((*bodies)[0]), &env); err != nil {
		t.Fatalf("decode envelope: %v\nbody: %s", err, (*bodies)[0])
	}

	if env.Body.Move.Focus.Absolute.Position != 0.7 {
		t.Errorf("Focus/Absolute/Position not in ver10/schema namespace: %+v\nbody: %s",
			env.Body.Move.Focus, (*bodies)[0])
	}
}
