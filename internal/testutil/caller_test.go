package testutil

// Self-tests for the fake caller used across the service suites: the
// decode path, recording semantics, and the helpers tests rely on.

import (
	"context"
	"encoding/xml"
	"errors"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
)

type echoRequest struct {
	XMLName xml.Name `xml:"Echo"`
	Value   string   `xml:"Value"`
}

type echoResponse struct {
	XMLName xml.Name `xml:"EchoResponse"`
	Value   string   `xml:"Value"`
}

func TestFakeCallerRoundtrip(t *testing.T) {
	caller := NewFakeCaller("http://fake/echo", func(action, reqXML string) (string, error) {
		if action != "Echo" {
			return "", errors.New("unexpected action " + action)
		}

		if !contains(reqXML, "ping") {
			return "", errors.New("request value missing")
		}

		return `<EchoResponse><Value>pong</Value></EchoResponse>`, nil
	})

	var resp echoResponse
	if err := caller.Call(context.Background(), caller.EndpointFor(api.ServiceDevice), "", echoRequest{Value: "ping"}, &resp); err != nil {
		t.Fatal(err)
	}

	if resp.Value != "pong" {
		t.Fatalf("decode path broken: %+v", resp)
	}

	if got := caller.EndpointFor(api.ServiceDevice); got != "http://fake/echo" {
		t.Errorf("EndpointFor = %q", got)
	}
}

func TestFakeCallerErrorPropagation(t *testing.T) {
	wantErr := errors.New("boom")
	caller := NewFakeCaller("", func(_, _ string) (string, error) { return "", wantErr })

	if err := caller.Call(context.Background(), "", "", echoRequest{}, &echoResponse{}); !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want boom", err)
	}
}

func TestFakeCallerBadResponseXML(t *testing.T) {
	caller := NewFakeCaller("", func(_, _ string) (string, error) { return "<not-xml", nil })

	var resp echoResponse
	if err := caller.Call(context.Background(), "", "", echoRequest{}, &resp); err == nil {
		t.Fatal("malformed response accepted")
	}
}

func TestFakeCallerEmptyResponseSkipsDecode(t *testing.T) {
	caller := NewFakeCaller("", func(_, _ string) (string, error) { return "", nil })

	var resp echoResponse
	if err := caller.Call(context.Background(), "", "", echoRequest{}, &resp); err != nil {
		t.Fatalf("empty response must be a no-op: %v", err)
	}
}

func TestFakeCallerRecording(t *testing.T) {
	caller := NewFakeCaller("", func(_, _ string) (string, error) { return "", nil })

	_ = caller.Call(context.Background(), "target-1", "", echoRequest{Value: "a"}, nil)
	_ = caller.Call(context.Background(), "target-2", "", echoRequest{Value: "b"}, nil)

	requests := caller.Requests()
	if len(requests) != 2 {
		t.Fatalf("recorded %d requests", len(requests))
	}

	if requests[0].Action != "Echo" || requests[0].Target != "target-1" {
		t.Errorf("first request = %+v", requests[0])
	}

	if requests[1].Body == "" {
		t.Error("request body not recorded")
	}

	// The returned slice is a copy — mutating it must not affect the caller.
	requests[0].Action = "mutated"
	if caller.Requests()[0].Action == "mutated" {
		t.Error("Requests() leaks internal state")
	}

	if caller.CountAction("Echo") != 2 || caller.CountAction("Other") != 0 {
		t.Errorf("CountAction broken")
	}
}

func TestElementName(t *testing.T) {
	cases := map[string]string{
		"<Echo></Echo>":           "Echo",
		`<Echo xmlns="x"></Echo>`: "Echo",
		`<ns:Op a="b"/>`:          "ns:Op",
		"":                        "",
		"no-xml":                  "",
		"<unclosed":               "",
	}

	for xmlIn, want := range cases {
		if got := elementName(xmlIn); got != want {
			t.Errorf("elementName(%q) = %q, want %q", xmlIn, got, want)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}

	return false
}
