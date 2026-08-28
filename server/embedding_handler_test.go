package server

// Tests for the embedding surface (#35): hosts mount the ONVIF services
// on their own mux alongside their own routes without calling Start,
// and embedding mode produces no stdout output — startup logs go to the
// injectable slog.Logger only.

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const deviceInfoProbe = `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><GetDeviceInformation xmlns="http://www.onvif.org/ver10/device/wsdl"/></s:Body></s:Envelope>`

// TestRegisterServicesOnCustomMux pins #35: RegisterServices mounts the
// ONVIF endpoints on a host-owned mux, host routes coexist, and Start is
// never involved.
func TestRegisterServicesOnCustomMux(t *testing.T) {
	srv, err := New(createTestConfig())
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	srv.RegisterServices(mux)
	mux.HandleFunc("/extra", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/onvif/device_service", "application/soap+xml", strings.NewReader(deviceInfoProbe))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("device service through custom mux: status %d body %s", resp.StatusCode, body)
	}

	if !strings.Contains(string(body), "Test") {
		t.Fatalf("GetDeviceInformation response missing manufacturer: %s", body)
	}

	extra, err := http.Get(ts.URL + "/extra")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = extra.Body.Close() }()

	if extra.StatusCode != http.StatusTeapot {
		t.Fatalf("host route /extra: status %d, want 418 (host routes must survive)", extra.StatusCode)
	}
}

// TestHandlerServesSnapshot pins the Handler() convenience path end to
// end: the snapshot endpoint answers at the default path.
func TestHandlerServesSnapshot(t *testing.T) {
	srv, err := New(createTestConfig())
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/onvif/snapshot")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("snapshot via Handler(): status %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "image/jpeg" {
		t.Fatalf("snapshot Content-Type = %q, want image/jpeg", ct)
	}
}

// recordingHandler captures slog records for assertions.
type recordingHandler struct {
	messages []string
}

func (h *recordingHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *recordingHandler) Handle(_ context.Context, r slog.Record) error {
	h.messages = append(h.messages, r.Message)

	return nil
}

func (h *recordingHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }

func (h *recordingHandler) WithGroup(_ string) slog.Handler { return h }

// captureStdout swaps os.Stdout for a pipe and returns the reader plus
// a restore func. The goroutine drain keeps the pipe from filling.
func captureStdout(t *testing.T) (*os.File, func()) {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	os.Stdout = w

	return r, func() {
		os.Stdout = original
		_ = w.Close()
	}
}

// TestStartupNoStdoutWithoutLogger pins #35: with no Logger configured
// the startup path writes nothing to stdout (embedded hosts keep a
// clean console; the CLI banner era is gone).
func TestStartupNoStdoutWithoutLogger(t *testing.T) {
	srv, err := New(createTestConfig())
	if err != nil {
		t.Fatal(err)
	}

	r, restore := captureStdout(t)
	srv.logStartup("127.0.0.1:8080")
	restore()

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}

	if len(data) != 0 {
		t.Fatalf("startup wrote %d bytes to stdout with no logger configured: %q", len(data), data)
	}
}

// TestStartupLogsViaLogger pins #35: with a Logger configured the same
// startup path emits structured records carrying the listen address.
func TestStartupLogsViaLogger(t *testing.T) {
	rec := &recordingHandler{}
	config := createTestConfig()
	config.Logger = slog.New(rec)

	srv, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	r, restore := captureStdout(t)
	srv.logStartup("127.0.0.1:8080")
	restore()

	data, _ := io.ReadAll(r)
	if len(data) != 0 {
		t.Fatalf("structured logger path still wrote %d bytes to stdout: %q", len(data), data)
	}

	if len(rec.messages) == 0 {
		t.Fatal("no slog records emitted through the configured logger")
	}
}
