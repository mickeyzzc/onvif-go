package onvif

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestGetStreamURIResponseVariants replays captured device responses through
// the full client path: every envelope variant (SOAP 1.2 with prefixes,
// SOAP 1.1 with default namespaces, a missing MediaUri wrapper) must yield a
// usable URI instead of the historical silent empty string (issue #3).
func TestGetStreamURIResponseVariants(t *testing.T) {
	tests := []struct {
		name     string
		fixture  string
		wantURI  string
		wantErr  bool
		errIs    error
		wantBody bool
	}{
		{
			name:    "SOAP 1.2 with trt/tt prefixes",
			fixture: "getstreamuri_normal_soap12.xml",
			wantURI: "rtsp://192.168.1.100:554/stream1",
		},
		{
			name:    "SOAP 1.1 envelope with default namespaces",
			fixture: "getstreamuri_soap11_prefixes.xml",
			wantURI: "rtsp://192.168.1.101:554/11/media/video",
		},
		{
			name:    "Uri without MediaUri wrapper (loose extraction)",
			fixture: "getstreamuri_no_mediauri_wrapper.xml",
			wantURI: "rtsp://192.168.1.102:8554/h264_pcm",
		},
		{
			name:     "empty Uri element",
			fixture:  "getstreamuri_empty_uri.xml",
			wantErr:  true,
			errIs:    ErrEmptyMediaURI,
			wantBody: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri, err := fetchStreamURIWithFixture(t, tt.fixture)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("GetStreamURI() = %v, nil, want error", uri)
				}

				if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("errors.Is(err, ErrEmptyMediaURI) = false, err = %v", err)
				}

				if tt.wantBody && len(err.Error()) < len(ErrEmptyMediaURI.Error())+20 {
					t.Errorf("error message lacks response body summary: %v", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("GetStreamURI() error = %v", err)
			}

			if uri.URI != tt.wantURI {
				t.Errorf("URI = %q, want %q", uri.URI, tt.wantURI)
			}
		})
	}
}

func TestGetSnapshotURIResponseVariants(t *testing.T) {
	// GetSnapshotUri shares the response shape; the same hardening applies.
	uri, err := fetchSnapshotURIWithFixture(t, "getsnapshoturi_normal.xml")
	if err != nil {
		t.Fatalf("GetSnapshotURI() error = %v", err)
	}

	if uri.URI != "http://192.168.1.100:80/onvif-http/snapshot.jpg" {
		t.Errorf("URI = %q, want snapshot URL", uri.URI)
	}
}

// fixtureServer serves a raw capture as the SOAP response.
func fixtureServer(t *testing.T, fixture string) *httptest.Server {
	t.Helper()

	body, err := os.ReadFile(filepath.Join("testdata", "captures", fixture))
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", fixture, err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)

	return server
}

func fixtureClient(t *testing.T, fixture string) *Client {
	t.Helper()

	server := fixtureServer(t, fixture)

	client, err := NewClient(server.URL, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	return client
}

func fetchStreamURIWithFixture(t *testing.T, fixture string) (*MediaURI, error) {
	t.Helper()

	client := fixtureClient(t, fixture)

	return client.Media().GetStreamURI(context.Background(), "profile-1")
}

func fetchSnapshotURIWithFixture(t *testing.T, fixture string) (*MediaURI, error) {
	t.Helper()

	client := fixtureClient(t, fixture)

	return client.Media().GetSnapshotURI(context.Background(), "profile-1")
}

func TestGetStreamURIFaultWithHTTP200(t *testing.T) {
	// 200-with-Fault must surface as an error (soap-layer fault detection),
	// never as a nil-error empty URI.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(notAuthorizedFaultBody))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.Media().GetStreamURI(context.Background(), "profile-1")
	if err == nil {
		t.Fatal("GetStreamURI() succeeded on fault response, want error")
	}

	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("errors.Is(err, ErrUnauthorized) = false, err = %v", err)
	}
}
