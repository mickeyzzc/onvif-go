// Package testutil provides the fake api.Caller used by the domain
// service tests: requests are recorded and answered from a pluggable
// handler through the exact xml.Unmarshal decode path of the real
// transport — no sockets, microsecond latency, deterministic.
package testutil

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"
	"sync"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
)

// FakeCaller implements api.Caller for table-driven service tests.
type FakeCaller struct {
	// Endpoint is returned by EndpointFor; empty simulates a service
	// the client has no endpoint for (ErrServiceNotSupported paths).
	Endpoint string

	mu       sync.Mutex
	requests []Request
	handler  func(action, reqXML string) (respXML string, err error)
}

// Request is one recorded SOAP call.
type Request struct {
	Action string // request element name, prefix included
	Body   string // marshaled request XML
	Target string // endpoint the call was sent to
}

// NewFakeCaller builds a caller answering via handler.
func NewFakeCaller(endpoint string, handler func(action, reqXML string) (string, error)) *FakeCaller {
	return &FakeCaller{Endpoint: endpoint, handler: handler}
}

// Call implements api.Caller.
func (f *FakeCaller) Call(_ context.Context, target, _ string, request, response interface{}) error {
	reqXML, err := xml.Marshal(request)
	if err != nil {
		return fmt.Errorf("fake caller: marshal request: %w", err)
	}

	f.mu.Lock()
	f.requests = append(f.requests, Request{Action: elementName(string(reqXML)), Body: string(reqXML), Target: target})
	handler := f.handler
	f.mu.Unlock()

	respXML, herr := handler(elementName(string(reqXML)), string(reqXML))
	if herr != nil {
		return herr
	}

	if respXML == "" || response == nil {
		return nil
	}

	if err := xml.Unmarshal([]byte(respXML), response); err != nil {
		return fmt.Errorf("fake caller: unmarshal response: %w", err)
	}

	return nil
}

// EndpointFor implements api.Caller.
func (f *FakeCaller) EndpointFor(api.Service) string {
	return f.Endpoint
}

// Requests returns a copy of the recorded requests.
func (f *FakeCaller) Requests() []Request {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]Request(nil), f.requests...)
}

// CountAction counts requests by element name.
func (f *FakeCaller) CountAction(action string) int {
	n := 0

	for _, r := range f.Requests() {
		if r.Action == action {
			n++
		}
	}

	return n
}

// elementName extracts the root element name of a marshaled request.
func elementName(reqXML string) string {
	start := strings.Index(reqXML, "<")
	if start < 0 {
		return ""
	}

	rest := reqXML[start+1:]
	if end := strings.Index(rest, ">"); end > 0 {
		return strings.SplitN(rest[:end], " ", 2)[0]
	}

	return ""
}
