package soap

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// errServerBug stands in for any plain handler failure.
var errServerBug = errors.New("boom")

// TestSenderFaultErrorMapping: a handler error that is a *SenderFaultError
// must surface as a SOAP Sender fault with HTTP 400 — the channel for
// client mistakes (unknown subscription, invalid duration, …) — while a
// plain error keeps the historical Receiver fault.
func TestSenderFaultErrorMapping(t *testing.T) {
	handler := NewHandler("", "")
	handler.RegisterContextHandler("ClientMistake", func(_ *RequestContext, _ []byte) (interface{}, error) {
		return nil, &SenderFaultError{Reason: "Unknown subscription", Detail: "no such pull point"}
	})
	handler.RegisterContextHandler("ServerBug", func(_ *RequestContext, _ []byte) (interface{}, error) {
		return nil, errServerBug
	})

	post := func(action string) (int, string) {
		body := `<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope"><Body><` +
			action + `/></Body></Envelope>`

		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		if err != nil {
			t.Fatalf("build request: %v", err)
		}

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec.Code, rec.Body.String()
	}

	status, body := post("ClientMistake")
	if status != http.StatusBadRequest {
		t.Errorf("SenderFaultError status = %d, want 400\nbody: %s", status, body)
	}

	if !strings.Contains(body, "<Value>Sender</Value>") {
		t.Errorf("SenderFaultError must carry SOAP code Sender:\n%s", body)
	}

	if !strings.Contains(body, "Unknown subscription") {
		t.Errorf("SenderFaultError reason missing from fault:\n%s", body)
	}

	status, body = post("ServerBug")
	if status != http.StatusInternalServerError {
		t.Errorf("plain error status = %d, want 500 (unchanged)", status)
	}

	if !strings.Contains(body, "<Value>Receiver</Value>") {
		t.Errorf("plain error must keep the Receiver fault:\n%s", body)
	}
}
