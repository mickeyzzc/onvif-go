package soap

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// recordingHandler captures structured entries for assertions (issue #62).
type recordingHandler struct {
	entries []slog.Record
}

func (h *recordingHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *recordingHandler) Handle(_ context.Context, r slog.Record) error {
	h.entries = append(h.entries, r)
	return nil
}

func (h *recordingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *recordingHandler) WithGroup(string) slog.Handler      { return h }

func (h *recordingHandler) joined() string {
	msgs := make([]string, 0, len(h.entries))
	for _, e := range h.entries {
		msgs = append(msgs, e.Message)
	}
	return strings.Join(msgs, "\n")
}

// The default client stays silent (nil logger).
func TestSetLoggerSilentByDefault(t *testing.T) {
	c := NewClient(nil, "", "")
	if c.slogger != nil {
		t.Fatal("default client must not carry a logger")
	}
}

func TestSetLoggerRecordsRequestAndResponse(t *testing.T) {
	const reply = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<GetSystemDateAndTimeResponse xmlns="http://www.onvif.org/ver10/device/wsdl"/></s:Body></s:Envelope>`

	srv := httptest.NewServer(handler200(reply))
	defer srv.Close()

	rec := &recordingHandler{}
	c := NewClient(srv.Client(), "", "")
	c.SetLogger(slog.New(rec))

	type emptyRequest struct {
		XMLName struct{} `xml:"GetSystemDateAndTime"`
	}
	var response struct {
		LocalTime string `xml:"LocalDateTime>Time"`
	}
	action := "http://www.onvif.org/ver10/device/wsdl/GetSystemDateAndTime"
	if err := c.Call(context.Background(), srv.URL, action, emptyRequest{}, &response); err != nil {
		t.Fatalf("Call: %v", err)
	}

	got := rec.joined()
	for _, want := range []string{"soap request", "soap response"} {
		if !strings.Contains(got, want) {
			t.Errorf("log stream missing %q entry; got:\n%s", want, got)
		}
	}
}

func TestSetLoggerRecordsFaultOutcome(t *testing.T) {
	const fault = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><s:Fault>
<s:Code><s:Value>s:Sender</s:Value></s:Code>
<s:Reason><s:Text xml:lang="en">ter:NotAuthorized</s:Text></s:Reason>
</s:Fault></s:Body></s:Envelope>`

	srv := httptest.NewServer(handler200(fault))
	defer srv.Close()

	rec := &recordingHandler{}
	c := NewClient(srv.Client(), "admin", "wrong")
	c.SetLogger(slog.New(rec))

	type emptyRequest struct {
		XMLName struct{} `xml:"GetDeviceInformation"`
	}
	var response struct{}
	action := "http://www.onvif.org/ver10/device/wsdl/GetDeviceInformation"
	_ = c.Call(context.Background(), srv.URL, action, emptyRequest{}, &response)

	if got := rec.joined(); !strings.Contains(got, "fault") {
		t.Errorf("fault outcome must be logged; got:\n%s", got)
	}
}

func handler200(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}
}
