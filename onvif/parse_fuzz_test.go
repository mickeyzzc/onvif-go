package onvif_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	onvif "github.com/mickeyzzc/onvif-go/v2/onvif"
)

// Client-response parse fuzz (issue #61's client-side half, applied to the
// real per-operation decode paths): a hostile or simply broken device
// response must never panic or hang the client, whatever bytes arrive.
// Each iteration feeds the fuzzed body through a real HTTP server and
// drives the genuine Client.Call pipeline — envelope extraction, fault
// detection, per-op unmarshal, post-processing. Seeds are the wire shapes
// a conformant device sends (including the namespace-correct forms from
// the contract suites). The seed corpus runs on every normal `go test`;
// extended fuzzing is `go test -fuzz`.

// fuzzBodyServer answers every request with the given body and a 200.
func fuzzBodyServer(body []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write(body)
	}))
}

func fuzzClient(t *testing.T, url string) *onvif.Client {
	t.Helper()

	c, err := onvif.NewClient(url, onvif.WithCredentials("u", "p"), onvif.WithTimeout(2*time.Second))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	return c
}

const (
	envHead = `<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope"><Body>`
	envTail = `</Body></Envelope>`
)

func FuzzParseDeviceResponses(f *testing.F) {
	f.Add([]byte(envHead + `<GetDeviceInformationResponse xmlns="http://www.onvif.org/ver10/device/wsdl"><Manufacturer xmlns="http://www.onvif.org/ver10/device/wsdl">onvif-go</Manufacturer><SerialNumber xmlns="http://www.onvif.org/ver10/device/wsdl">SN-1</SerialNumber></GetDeviceInformationResponse>` + envTail))
	f.Add([]byte(envHead + `<GetSystemDateAndTimeResponse xmlns="http://www.onvif.org/ver10/device/wsdl"><SystemDateAndTime xmlns="http://www.onvif.org/ver10/device/wsdl"><DateTimeType xmlns="http://www.onvif.org/ver10/schema">NTP</DateTimeType><TimeZone xmlns="http://www.onvif.org/ver10/schema"><TZ xmlns="http://www.onvif.org/ver10/schema">UTC</TZ></TimeZone><UTCDateTime xmlns="http://www.onvif.org/ver10/schema"><Time xmlns="http://www.onvif.org/ver10/schema"><Hour>23</Hour><Minute>59</Second></Time><Date xmlns="http://www.onvif.org/ver10/schema"><Year>2026</Year><Month>13</Month><Day>0</Day></Date></UTCDateTime></SystemDateAndTime></GetSystemDateAndTimeResponse>` + envTail))
	f.Add([]byte(envHead + `<GetServicesResponse xmlns="http://www.onvif.org/ver10/device/wsdl"><Service xmlns="http://www.onvif.org/ver10/device/wsdl"><Namespace xmlns="http://www.onvif.org/ver10/device/wsdl">http://www.onvif.org/ver10/device/wsdl</Namespace><XAddr xmlns="http://www.onvif.org/ver10/device/wsdl">http://127.0.0.1/onvif/device_service</XAddr><Version xmlns="http://www.onvif.org/ver10/device/wsdl"><Major xmlns="http://www.onvif.org/ver10/schema">2</Major><Minor xmlns="http://www.onvif.org/ver10/schema">60</Minor></Version></Service></GetServicesResponse>` + envTail))
	f.Add([]byte(envHead + `<GetCapabilitiesResponse xmlns="http://www.onvif.org/ver10/device/wsdl"><Capabilities xmlns="http://www.onvif.org/ver10/device/wsdl"><Device xmlns="http://www.onvif.org/ver10/schema"><XAddr xmlns="http://www.onvif.org/ver10/schema">http://127.0.0.1/onvif/device_service</XAddr></Device></Capabilities></GetCapabilitiesResponse>` + envTail))
	f.Add([]byte(envHead + `<s:Fault xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Code><s:Value>s:Sender</s:Value></s:Code><s:Reason><s:Text>ter:InvalidArgs</s:Text></s:Reason></s:Fault>` + envTail))
	f.Add([]byte("<Envelope><Body>"))
	f.Add([]byte{0x00, 0xff, 0xfe})

	f.Fuzz(func(t *testing.T, body []byte) {
		srv := fuzzBodyServer(body)
		defer srv.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		client := fuzzClient(t, srv.URL)
		client.SetServiceEndpoint(api.ServiceDevice, srv.URL)

		_, _ = client.Device().GetDeviceInformation(ctx)
		_, _ = client.Device().FixedGetSystemDateAndTime(ctx)
		_, _ = client.Device().GetServices(ctx, false)
		_, _ = client.Device().GetCapabilities(ctx)
	})
}

func FuzzParseMediaProfiles(f *testing.F) {
	f.Add([]byte(envHead + `<GetProfilesResponse xmlns="http://www.onvif.org/ver10/media/wsdl"><Profiles xmlns="http://www.onvif.org/ver10/media/wsdl" token="p1" fixed="true"><Name xmlns="http://www.onvif.org/ver10/schema">Main</Name><VideoEncoderConfiguration xmlns="http://www.onvif.org/ver10/schema" token="e1"><Name xmlns="http://www.onvif.org/ver10/schema">Enc</Name><UseCount xmlns="http://www.onvif.org/ver10/schema">1</UseCount><Encoding xmlns="http://www.onvif.org/ver10/schema">H264</Encoding><Resolution xmlns="http://www.onvif.org/ver10/schema"><Width xmlns="http://www.onvif.org/ver10/schema">1920</Width><Height xmlns="http://www.onvif.org/ver10/schema">1080</Height></Resolution><RateControl xmlns="http://www.onvif.org/ver10/schema"><FrameRateLimit xmlns="http://www.onvif.org/ver10/schema">15</FrameRateLimit></RateControl></VideoEncoderConfiguration></Profiles></GetProfilesResponse>` + envTail))
	f.Add([]byte(envHead + `<GetStreamUriResponse xmlns="http://www.onvif.org/ver10/media/wsdl"><MediaUri xmlns="http://www.onvif.org/ver10/media/wsdl"><Uri xmlns="http://www.onvif.org/ver10/schema">rtsp://127.0.0.1:8554/live</Uri><InvalidAfterConnect xmlns="http://www.onvif.org/ver10/schema">false</InvalidAfterConnect></MediaUri></GetStreamUriResponse>` + envTail))
	f.Add([]byte("<Envelope><Body>"))
	f.Add([]byte{0x00, 0xff, 0xfe})

	f.Fuzz(func(t *testing.T, body []byte) {
		srv := fuzzBodyServer(body)
		defer srv.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		client := fuzzClient(t, srv.URL)
		client.SetServiceEndpoint(api.ServiceMedia, srv.URL)

		_, _ = client.Media().GetProfiles(ctx)
		_, _ = client.Media().GetStreamURI(ctx, "p1")
		_, _ = client.Media().GetSnapshotURI(ctx, "p1")
	})
}

func FuzzParseEventNotifications(f *testing.F) {
	f.Add([]byte(envHead + `<PullMessagesResponse xmlns="http://www.onvif.org/ver10/events/wsdl"><CurrentTime xmlns="http://www.onvif.org/ver10/events/wsdl">2026-09-19T00:00:00Z</CurrentTime><TerminationTime xmlns="http://www.onvif.org/ver10/events/wsdl">2026-09-19T01:00:00Z</TerminationTime><NotificationMessage xmlns="http://docs.oasis-open.org/wsn/b-2"><Topic xmlns="http://docs.oasis-open.org/wsn/b-2">tns1:VideoSource/MotionAlarm</Topic><Message xmlns="http://docs.oasis-open.org/wsn/b-2"><Message xmlns="http://www.onvif.org/ver10/schema" PropertyOperation="Changed" UtcTime="2026-09-19T00:00:00Z"><Source xmlns="http://www.onvif.org/ver10/schema"><SimpleItem xmlns="http://www.onvif.org/ver10/schema" Name="Source" Value="CSI"/></Source><Data xmlns="http://www.onvif.org/ver10/schema"><SimpleItem xmlns="http://www.onvif.org/ver10/schema" Name="State" Value="true"/></Data></Message></Message></NotificationMessage></PullMessagesResponse>` + envTail))
	f.Add([]byte(envHead + `<PullMessagesResponse xmlns="http://www.onvif.org/ver10/events/wsdl"><CurrentTime>not-a-time</CurrentTime><NotificationMessage><Topic><Message><Message UtcTime=""/></Message></Message></NotificationMessage></PullMessagesResponse>` + envTail))
	f.Add([]byte("<Envelope><Body>"))
	f.Add([]byte{0x00, 0xff, 0xfe})

	f.Fuzz(func(t *testing.T, body []byte) {
		srv := fuzzBodyServer(body)
		defer srv.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		client := fuzzClient(t, srv.URL)

		_, _ = client.Events().PullMessages(ctx, srv.URL, 100*time.Millisecond, 5)
	})
}
