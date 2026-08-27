package onvif

import (
	"context"
	"crypto/sha1" //nolint:gosec // SHA1 required for ONVIF digest auth
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// skewDevice is a test camera whose clock runs `offset` away from real time
// and which fully verifies WS-Security digests (digest value AND Created
// within its ±5min replay window, evaluated against the DEVICE clock).
type skewDevice struct {
	offset     time.Duration // deviceTime = now + offset
	password   string
	rttDelay   time.Duration // simulated processing delay before answering
	alwaysFail bool          // reject every authenticated call regardless
}

var (
	reUsername = regexp.MustCompile(`<Username>([^<]*)</Username>`)
	rePassword = regexp.MustCompile(`<Password[^>]*>([^<]*)</Password>`)
	reNonce    = regexp.MustCompile(`<Nonce[^>]*>([^<]*)</Nonce>`)
	reCreated  = regexp.MustCompile(`<Created[^>]*>([^<]*)</Created>`)
)

func (d *skewDevice) handler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	bodyStr := string(body)
	deviceNow := time.Now().UTC().Add(d.offset)

	if strings.Contains(bodyStr, "GetSystemDateAndTime") {
		if d.rttDelay > 0 {
			time.Sleep(d.rttDelay)
		}

		t := deviceNow
		_, _ = w.Write([]byte(xmlEnvelope(`
<tds:GetSystemDateAndTimeResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:SystemDateAndTime>
<tds:UTCDateTime><tds:Time><tds:Hour>` + itoa(t.Hour()) + `</tds:Hour><tds:Minute>` +
			itoa(t.Minute()) + `</tds:Minute><tds:Second>` + itoa(t.Second()) +
			`</tds:Second></tds:Time><tds:Date><tds:Year>` + itoa(t.Year()) +
			`</tds:Year><tds:Month>` + itoa(int(t.Month())) + `</tds:Month><tds:Day>` +
			itoa(t.Day()) + `</tds:Day></tds:Date></tds:UTCDateTime>
</tds:SystemDateAndTime>
</tds:GetSystemDateAndTimeResponse>`)))

		return
	}

	if !d.alwaysFail && verifyDigest(bodyStr, deviceNow, d.password) {
		if strings.Contains(bodyStr, "GetCapabilities") {
			_, _ = w.Write([]byte(xmlEnvelope(`
<tds:GetCapabilitiesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:Capabilities><tds:Device><tds:XAddr>http://device/onvif/device_service</tds:XAddr>
</tds:Device></tds:Capabilities>
</tds:GetCapabilitiesResponse>`)))

			return
		}

		_, _ = w.Write([]byte(xmlEnvelope(`
<tds:GetDeviceInformationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:Manufacturer>SkewCam</tds:Manufacturer>
</tds:GetDeviceInformationResponse>`)))

		return
	}

	_, _ = w.Write([]byte(notAuthorizedFaultBody))
}

func xmlEnvelope(body string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>` + body + `</s:Body></s:Envelope>`
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

// verifyDigest recomputes the WS-Security digest from the request body and
// checks it against the device password and clock window.
func verifyDigest(body string, deviceNow time.Time, password string) bool {
	user := reUsername.FindStringSubmatch(body)
	pwd := rePassword.FindStringSubmatch(body)
	nonce := reNonce.FindStringSubmatch(body)
	created := reCreated.FindStringSubmatch(body)
	if user == nil || pwd == nil || nonce == nil || created == nil {
		return false
	}

	createdTime, err := time.Parse(time.RFC3339, created[1])
	if err != nil {
		return false
	}

	// Replay window evaluated on the DEVICE clock (Hikvision behavior).
	if diff := createdTime.Sub(deviceNow); diff < -5*time.Minute || diff > 5*time.Minute {
		return false
	}

	nonceBytes, err := base64.StdEncoding.DecodeString(nonce[1])
	if err != nil {
		return false
	}

	hash := sha1.New() //nolint:gosec // ONVIF digest algorithm
	hash.Write(nonceBytes)
	hash.Write([]byte(created[1]))
	hash.Write([]byte(password))

	return base64.StdEncoding.EncodeToString(hash.Sum(nil)) == pwd[1]
}

func newSkewTestClient(t *testing.T, dev *skewDevice, opts ...ClientOption) *Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(dev.handler))
	t.Cleanup(server.Close)

	allOpts := append([]ClientOption{
		WithCredentials("admin", dev.password),
		WithHTTPClient(server.Client()),
	}, opts...)

	client, err := NewClient(server.URL, allOpts...)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	return client
}

// TestMeasureClockSkewRTTCompensation verifies that a device answering with a
// delay does not pollute the measurement: the skew must land near the true
// offset, not offset±RTT/2.
func TestMeasureClockSkewRTTCompensation(t *testing.T) {
	const trueOffset = 90 * time.Minute

	dev := &skewDevice{offset: trueOffset, password: "secret", rttDelay: 300 * time.Millisecond}
	client := newSkewTestClient(t, dev)

	skew, err := client.MeasureClockSkew(context.Background())
	if err != nil {
		t.Fatalf("MeasureClockSkew() error = %v", err)
	}

	if diff := skew - trueOffset; diff < -time.Second || diff > time.Second {
		t.Errorf("skew = %v, want ~%v (RTT compensation off by %v)", skew, trueOffset, diff)
	}
}

// TestMeasureClockSkewSafeFallbacks covers zero-time and malformed responses.
func TestMeasureClockSkewSafeFallbacks(t *testing.T) {
	bodies := map[string]string{
		"zero time": xmlEnvelope(`
<tds:GetSystemDateAndTimeResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:SystemDateAndTime><tds:UTCDateTime><tds:Time><tds:Hour>0</tds:Hour>
</tds:Time><tds:Date><tds:Year>0</tds:Year></tds:Date></tds:UTCDateTime>
</tds:SystemDateAndTime></tds:GetSystemDateAndTimeResponse>`),
		"malformed": `<not-a-response/>`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(body))
			}))
			t.Cleanup(server.Close)

			client, err := NewClient(server.URL, WithHTTPClient(server.Client()))
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			skew, err := client.MeasureClockSkew(context.Background())
			if err == nil {
				t.Fatalf("MeasureClockSkew() = %v, nil; want error", skew)
			}

			if skew != 0 {
				t.Errorf("skew = %v on failure, want 0", skew)
			}
		})
	}
}

// TestWithAutoClockSkewOnInitialize shows the end-to-end fix: a Hikvision-
// style clock-diverged camera rejects local-time digests; auto skew makes
// Initialize + calls succeed.
func TestWithAutoClockSkewOnInitialize(t *testing.T) {
	dev := &skewDevice{offset: 42 * time.Minute, password: "secret"}

	t.Run("without auto skew digest is rejected", func(t *testing.T) {
		client := newSkewTestClient(t, dev)
		if _, err := client.Device().GetDeviceInformation(context.Background()); err == nil {
			t.Fatal("GetDeviceInformation() succeeded against a 42-minute-diverged clock, want rejection")
		}
	})

	t.Run("with auto skew initialize and calls succeed", func(t *testing.T) {
		client := newSkewTestClient(t, dev, WithAutoClockSkew())

		if err := client.Initialize(context.Background()); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if _, err := client.Device().GetDeviceInformation(context.Background()); err != nil {
			t.Fatalf("GetDeviceInformation() error = %v", err)
		}
	})
}

func TestDiagnoseAuth(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		client := newSkewTestClient(t, &skewDevice{password: "secret"})

		diag, err := client.DiagnoseAuth(context.Background())
		if err != nil {
			t.Fatalf("DiagnoseAuth() error = %v", err)
		}

		if diag.Status != AuthStatusOK {
			t.Errorf("Status = %q, want %q (%s)", diag.Status, AuthStatusOK, diag.Detail)
		}
	})

	t.Run("clock skew confirmed by device-time retry", func(t *testing.T) {
		client := newSkewTestClient(t, &skewDevice{offset: 10 * time.Minute, password: "secret"})

		diag, err := client.DiagnoseAuth(context.Background())
		if err != nil {
			t.Fatalf("DiagnoseAuth() error = %v", err)
		}

		if diag.Status != AuthStatusClockSkew {
			t.Fatalf("Status = %q, want %q (%s)", diag.Status, AuthStatusClockSkew, diag.Detail)
		}

		if diag.ClockSkew < 9*time.Minute || diag.ClockSkew > 11*time.Minute {
			t.Errorf("ClockSkew = %v, want ~10m", diag.ClockSkew)
		}
	})

	t.Run("bad credentials despite skew", func(t *testing.T) {
		// Device clock diverged AND digest always rejected: the retry with
		// device time must still fail, pointing at credentials.
		client := newSkewTestClient(t, &skewDevice{offset: 10 * time.Minute, password: "right", alwaysFail: true})
		client.SetCredentials("admin", "wrong")

		diag, err := client.DiagnoseAuth(context.Background())
		if err != nil {
			t.Fatalf("DiagnoseAuth() error = %v", err)
		}

		if diag.Status != AuthStatusBadCredentials {
			t.Errorf("Status = %q, want %q (%s)", diag.Status, AuthStatusBadCredentials, diag.Detail)
		}
	})

	t.Run("no significant skew means credentials", func(t *testing.T) {
		client := newSkewTestClient(t, &skewDevice{password: "right", alwaysFail: true})
		client.SetCredentials("admin", "wrong")

		diag, err := client.DiagnoseAuth(context.Background())
		if err != nil {
			t.Fatalf("DiagnoseAuth() error = %v", err)
		}

		if diag.Status != AuthStatusBadCredentials {
			t.Errorf("Status = %q, want %q (%s)", diag.Status, AuthStatusBadCredentials, diag.Detail)
		}
	})
}
