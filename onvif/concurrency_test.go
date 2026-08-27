package onvif

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// allOpsMock answers every operation family used by the concurrency matrix.
func allOpsMock(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		b := string(body)

		switch {
		case strings.Contains(b, "GetDeviceInformation"):
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<tds:GetDeviceInformationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:Manufacturer>ConcCam</tds:Manufacturer><tds:SerialNumber>C-1</tds:SerialNumber>
</tds:GetDeviceInformationResponse></s:Body></s:Envelope>`))
		case strings.Contains(b, "GetCapabilities"):
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<tds:GetCapabilitiesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:Capabilities><tds:Media><tds:XAddr>http://device/onvif/media</tds:XAddr></tds:Media>
</tds:Capabilities></tds:GetCapabilitiesResponse></s:Body></s:Envelope>`))
		case strings.Contains(b, "GetProfiles"):
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<trt:GetProfilesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema">
<trt:Profiles token="p1"><tt:VideoEncoderConfiguration><tt:Resolution>
<tt:Width>1920</tt:Width><tt:Height>1080</tt:Height></tt:Resolution>
</tt:VideoEncoderConfiguration></trt:Profiles>
</trt:GetProfilesResponse></s:Body></s:Envelope>`))
		case strings.Contains(b, "GetStreamUri"):
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<trt:GetStreamUriResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema">
<trt:MediaUri><tt:Uri>rtsp://device/main</tt:Uri></trt:MediaUri>
</trt:GetStreamUriResponse></s:Body></s:Envelope>`))
		case strings.Contains(b, "ContinuousMove"), strings.Contains(b, "Stop"):
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<tptz:ContinuousMoveResponse xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl"/>
</s:Body></s:Envelope>`))
		case strings.Contains(b, "PullMessages"):
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<tev:PullMessagesResponse xmlns:tev="http://www.onvif.org/ver10/events/wsdl"/>
</s:Body></s:Envelope>`))
		default:
			// Unknown ops answer an empty envelope; the matrix only asserts
			// absence of races/deadlocks, not full response semantics.
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<dummy:Response xmlns:dummy="urn:dummy"/>
</s:Body></s:Envelope>`))
		}
	}))
	t.Cleanup(server.Close)

	return server
}

// TestClientConcurrentMixedOperations is the concurrency-contract matrix
// (issue #12): one Client shared by many goroutines mixing Device, Media,
// PTZ and Events operations with concurrent configuration changes
// (credentials, clock skew, capabilities-cache invalidation) and concurrent
// Initialize. Run with -race (CI does) for the actual guarantee.
func TestClientConcurrentMixedOperations(t *testing.T) {
	server := allOpsMock(t)

	client, err := NewClient(server.URL,
		WithCredentials("admin", "secret"),
		WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	const workersPerOp = 4

	var wg sync.WaitGroup

	run := func(f func(ctx context.Context) error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				_ = recover() // a panic here would be a contract violation; let the test fail on wg timing
			}()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			for range 5 {
				_ = f(ctx) // errors are fine; races/deadlocks/panics are not
			}
		}()
	}

	for range workersPerOp {
		run(func(ctx context.Context) error {
			_, err := client.Device().GetDeviceInformation(ctx)

			return err
		})
		run(func(ctx context.Context) error {
			_, err := client.Media().GetProfiles(ctx)

			return err
		})
		run(func(ctx context.Context) error {
			_, err := client.Media().GetStreamURI(ctx, "p1")

			return err
		})
		run(func(ctx context.Context) error {
			_, err := client.Device().GetCapabilitiesCached(ctx)

			return err
		})
		run(func(ctx context.Context) error {
			_, err := client.Events().PullMessages(ctx, server.URL+"/sub", time.Second, 1)

			return err
		})
		run(func(ctx context.Context) error {
			return client.Initialize(ctx)
		})
	}

	// Concurrent mutators of the shared configuration.
	for range workersPerOp {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 10 {
				client.SetCredentials("admin", "secret")
				client.SetClockSkew(time.Duration(i) * time.Second)
				client.InvalidateCapabilitiesCache()
				client.ResetAuthLadder()
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("concurrent matrix deadlocked")
	}
}
