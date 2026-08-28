package server

import (
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// TestGetServicesIncludesEvents pins #46: with SupportEvents=true,
// GetServices must report the Events service (namespace + XAddr) so the
// enumeration agrees with GetCapabilities.
func TestGetServicesIncludesEvents(t *testing.T) {
	config := createTestConfig()
	config.SupportEvents = true

	srv, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	rc := &soap.RequestContext{RemoteIP: "192.0.2.9"}

	resp, err := srv.HandleGetServices(rc, []byte(`<GetServices/>`))
	if err != nil {
		t.Fatal(err)
	}

	services := resp.(*GetServicesResponse).Service

	find := func(ns string) *Service {
		for i := range services {
			if services[i].Namespace == ns {
				return &services[i]
			}
		}

		return nil
	}

	events := find("http://www.onvif.org/ver10/events/wsdl")
	if events == nil {
		t.Fatalf("Events service missing from GetServices: %+v", services)
	}

	if want := "http://192.0.2.9:8080/onvif/events_service"; events.XAddr != want {
		t.Errorf("Events XAddr = %q, want %q", events.XAddr, want)
	}

	if events.Version != (Version{Major: 2, Minor: 5}) {
		t.Errorf("Events version = %+v, want 2.5 like the other services", events.Version)
	}
}

// TestGetServicesOmitsEventsWhenDisabled: SupportEvents=false must not
// advertise the service.
func TestGetServicesOmitsEventsWhenDisabled(t *testing.T) {
	config := createTestConfig()
	config.SupportEvents = false

	srv, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := srv.HandleGetServices(nil, []byte(`<GetServices/>`))
	if err != nil {
		t.Fatal(err)
	}

	for _, svc := range resp.(*GetServicesResponse).Service {
		if svc.Namespace == "http://www.onvif.org/ver10/events/wsdl" {
			t.Errorf("Events advertised with SupportEvents=false: %+v", svc)
		}
	}
}
