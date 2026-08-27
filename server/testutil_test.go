package server

import (
	"encoding/xml"
	"testing"
)

// testBody marshals a request value into the raw XML bytes handlers
// receive from the SOAP transport.
func testBody(t *testing.T, v interface{}) []byte {
	t.Helper()

	data, err := xml.Marshal(v)
	if err != nil {
		t.Fatalf("marshal test body: %v", err)
	}

	return data
}
