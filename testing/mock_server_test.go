package onviftesting

// Tests for the capture-replay mock server (v1): archive loading, the
// HTTP replay path, and SOAP operation extraction.

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// writeCaptureArchive builds an in-memory tar.gz capture archive with
// one JSON exchange per name.
func writeCaptureArchive(t *testing.T, entries map[string]string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "camera.tar.gz")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	gzw := gzip.NewWriter(f)
	defer func() { _ = gzw.Close() }()

	tw := tar.NewWriter(gzw)
	defer func() { _ = tw.Close() }()

	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		body := entries[name]
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: int64(len(body))}); err != nil {
			t.Fatal(err)
		}

		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}

	return path
}

func TestLoadCaptureFromArchive(t *testing.T) {
	path := writeCaptureArchive(t, map[string]string{
		"001.json":  `{"timestamp":"2026-01-02T03:04:05Z","operation":1,"operation_name":"GetDeviceInformation","endpoint":"http://cam/onvif/device_service","request_body":"<GetDeviceInformation/>","response_body":"<GetDeviceInformationResponse/>","status_code":200}`,
		"002.json":  `{"operation":2,"operation_name":"GetServices","response_body":"<GetServicesResponse/>","status_code":200}`,
		"notes.txt": "ignored",
	})

	capture, err := LoadCaptureFromArchive(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(capture.Exchanges) != 2 {
		t.Fatalf("exchanges = %d, want 2 (non-JSON skipped)", len(capture.Exchanges))
	}

	if capture.Exchanges[0].OperationName != "GetDeviceInformation" {
		t.Errorf("first exchange = %+v", capture.Exchanges[0])
	}
}

func TestLoadCaptureFromArchiveErrors(t *testing.T) {
	if _, err := LoadCaptureFromArchive(filepath.Join(t.TempDir(), "missing.tar.gz")); err == nil {
		t.Error("missing archive accepted")
	}

	notGzip := filepath.Join(t.TempDir(), "plain.tar.gz")
	if err := os.WriteFile(notGzip, []byte("not gzip"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadCaptureFromArchive(notGzip); err == nil {
		t.Error("non-gzip archive accepted")
	}
}

func TestMockSOAPServerReplay(t *testing.T) {
	path := writeCaptureArchive(t, map[string]string{
		"001.json": `{"operation_name":"GetDeviceInformation","request_body":"<GetDeviceInformation xmlns=\"http://www.onvif.org/ver10/device/wsdl\"/>","response_body":"<GetDeviceInformationResponse><Manufacturer>Acme</Manufacturer></GetDeviceInformationResponse>","status_code":200}`,
	})

	mock, err := NewMockSOAPServer(path)
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	if mock.URL() == "" {
		t.Fatal("URL() empty")
	}

	resp, err := http.Post(mock.URL(), "application/soap+xml",
		strings.NewReader(`<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope"><Body><GetDeviceInformation xmlns="http://www.onvif.org/ver10/device/wsdl"/></Body></Envelope>`))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d body %s", resp.StatusCode, body)
	}

	if !bytes.Contains(body, []byte("Acme")) {
		t.Fatalf("captured response not replayed: %s", body)
	}

	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/soap+xml") {
		t.Errorf("Content-Type = %q", ct)
	}
}

func TestMockSOAPServerUnknownOperation(t *testing.T) {
	path := writeCaptureArchive(t, map[string]string{
		"001.json": `{"operation_name":"GetServices","response_body":"<GetServicesResponse/>","status_code":200}`,
	})

	mock, err := NewMockSOAPServer(path)
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	resp, err := http.Post(mock.URL(), "application/soap+xml",
		strings.NewReader(`<Body><GetProfiles xmlns="http://www.onvif.org/ver10/media/wsdl"/></Body>`))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 for uncaptured operation", resp.StatusCode)
	}
}

func TestExtractOperationFromSOAP(t *testing.T) {
	cases := map[string]string{
		`<Body><GetProfiles xmlns="x"/></Body>`: "GetProfiles",
		`<Body><trt:GetVideoSources/></Body>`:   "GetVideoSources", // prefix stripped
		`<GetDeviceInformation/>`:               "",                // no Body element
		`garbage`:                               "",
	}

	for body, want := range cases {
		if got := extractOperationFromSOAP(body); got != want {
			t.Errorf("extractOperationFromSOAP(%q) = %q, want %q", body, got, want)
		}
	}
}
