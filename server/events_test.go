package server

import (
	"context"
	"encoding/xml"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	onvif "github.com/mickeyzzc/onvif-go/v2/onvif"
	"github.com/mickeyzzc/onvif-go/v2/server/soap"
	"github.com/mickeyzzc/onvif-go/v2/types"
)

// postEventsSOAP posts a SOAP 1.2 envelope whose body is bodyXML to path
// on mux and returns the status plus the response body. Unlike the
// device-level postSOAP helper it does not presume success — fault paths
// are part of the events contract.
func postEventsSOAP(t *testing.T, mux *http.ServeMux, path, bodyXML string) (int, string) {
	t.Helper()

	// createTestConfig sets credentials and the default policy protects
	// Create* — carry a PasswordText token like a real subscribing client.
	envelope := `<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">` +
		`<Header><Security xmlns="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">` +
		`<UsernameToken><Username>admin</Username>` +
		`<Password Type="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordText">` +
		`password</Password></UsernameToken></Security></Header>` +
		`<Body>` + bodyXML + `</Body></Envelope>`

	req, err := http.NewRequest(http.MethodPost, path, strings.NewReader(envelope))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec.Code, rec.Body.String()
}

// eventsTestServer builds a server with SupportEvents enabled and its mux.
func eventsTestServer(t *testing.T) (*Server, *http.ServeMux) {
	t.Helper()

	config := createTestConfig()
	config.SupportEvents = true

	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	mux := http.NewServeMux()
	srv.RegisterServices(mux)

	return srv, mux
}

// createPullPointResponse is the wire decode target for
// CreatePullPointSubscriptionResponse (local-name matching, like the
// library's own events client).
type createPullPointResponse struct {
	XMLName               xml.Name `xml:"CreatePullPointSubscriptionResponse"`
	SubscriptionReference struct {
		Address string `xml:"Address"`
	} `xml:"SubscriptionReference"`
	CurrentTime     string `xml:"CurrentTime"`
	TerminationTime string `xml:"TerminationTime"`
}

// decodeCreateResponse extracts CreatePullPointSubscriptionResponse from a
// full SOAP envelope.
func decodeCreateResponse(t *testing.T, body string) createPullPointResponse {
	t.Helper()

	var env struct {
		Body struct {
			Response createPullPointResponse `xml:"CreatePullPointSubscriptionResponse"`
		} `xml:"Body"`
	}

	if err := xml.Unmarshal([]byte(body), &env); err != nil {
		t.Fatalf("decode CreatePullPointSubscriptionResponse: %v\nbody:\n%s", err, body)
	}

	return env.Body.Response
}

// pullMessagesWireResponse decodes PullMessagesResponse including the
// canonical double-layer notification structure (wsnt:Message wrapping an
// inner ONVIF tt:Message — the shape pinned in issue #82/#83).
type pullMessagesWireResponse struct {
	XMLName              xml.Name `xml:"PullMessagesResponse"`
	CurrentTime          string   `xml:"CurrentTime"`
	TerminationTime      string   `xml:"TerminationTime"`
	NotificationMessages []struct {
		Topic struct {
			Value string `xml:",chardata"`
		} `xml:"Topic"`
		ProducerReference struct {
			Address string `xml:"Address"`
		} `xml:"ProducerReference"`
		Message struct {
			Inner struct {
				PropertyOperation string `xml:"PropertyOperation,attr"`
				UtcTime           string `xml:"UtcTime,attr"`
				Source            struct {
					SimpleItems []eventItemWire `xml:"SimpleItem"`
				} `xml:"Source"`
				Data struct {
					SimpleItems []eventItemWire `xml:"SimpleItem"`
				} `xml:"Data"`
			} `xml:"Message"`
		} `xml:"Message"`
	} `xml:"NotificationMessage"`
}

type eventItemWire struct {
	Name  string `xml:"Name,attr"`
	Value string `xml:"Value,attr"`
}

func decodePullResponse(t *testing.T, body string) pullMessagesWireResponse {
	t.Helper()

	var env struct {
		Body struct {
			Response pullMessagesWireResponse `xml:"PullMessagesResponse"`
		} `xml:"Body"`
	}

	if err := xml.Unmarshal([]byte(body), &env); err != nil {
		t.Fatalf("decode PullMessagesResponse: %v\nbody:\n%s", err, body)
	}

	return env.Body.Response
}

// TestEventsServiceRouteServed pins #83: with SupportEvents=true the
// advertised events_service XAddr must actually serve SOAP requests —
// previously the route was never registered and every POST got a bare 404.
func TestEventsServiceRouteServed(t *testing.T) {
	_, mux := eventsTestServer(t)

	status, body := postEventsSOAP(t, mux, "/onvif/events_service",
		`<GetServiceCapabilities xmlns="http://www.onvif.org/ver10/events/wsdl"/>`)

	if status != http.StatusOK {
		t.Fatalf("GetServiceCapabilities status = %d, want 200 (route must exist)\nbody: %s", status, body)
	}

	if !strings.Contains(body, "GetServiceCapabilitiesResponse") {
		t.Errorf("response missing GetServiceCapabilitiesResponse:\n%s", body)
	}
}

// TestEventsServiceRouteAbsentWhenDisabled: SupportEvents=false keeps the
// historical behavior — no route, no advertisement.
func TestEventsServiceRouteAbsentWhenDisabled(t *testing.T) {
	config := createTestConfig()
	config.SupportEvents = false

	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	mux := http.NewServeMux()
	srv.RegisterServices(mux)

	status, _ := postEventsSOAP(t, mux, "/onvif/events_service",
		`<GetServiceCapabilities xmlns="http://www.onvif.org/ver10/events/wsdl"/>`)

	if status != http.StatusNotFound {
		t.Errorf("events_service status with SupportEvents=false = %d, want 404", status)
	}
}

// TestGetEventServiceCapabilitiesGolden locks the wire bytes of the
// capabilities answer (no time fields → fully deterministic).
func TestGetEventServiceCapabilitiesGolden(t *testing.T) {
	_, mux := eventsTestServer(t)

	status, body := postEventsSOAP(t, mux, "/onvif/events_service",
		`<GetServiceCapabilities xmlns="http://www.onvif.org/ver10/events/wsdl"/>`)

	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}

	want := `<?xml version="1.0" encoding="UTF-8"?>
<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">
  <Body>
    <GetServiceCapabilitiesResponse xmlns="http://www.onvif.org/ver10/events/wsdl">
      <Capabilities xmlns="http://www.onvif.org/ver10/events/wsdl" WSPullPointSupport="true" MaxPullPoints="10"></Capabilities>
    </GetServiceCapabilitiesResponse>
  </Body>
</Envelope>`

	if body != want {
		t.Errorf("GetServiceCapabilities golden mismatch\n got:\n%s\nwant:\n%s", body, want)
	}
}

// TestGetEventPropertiesResponse: minimal honest properties — fixed empty
// topic set, no filter dialects (subscription filtering is not applied).
func TestGetEventPropertiesResponse(t *testing.T) {
	_, mux := eventsTestServer(t)

	status, body := postEventsSOAP(t, mux, "/onvif/events_service",
		`<GetEventProperties xmlns="http://www.onvif.org/ver10/events/wsdl"/>`)

	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200\nbody: %s", status, body)
	}

	if !strings.Contains(body, "GetEventPropertiesResponse") {
		t.Fatalf("response missing GetEventPropertiesResponse:\n%s", body)
	}

	if !strings.Contains(body, "<FixedTopicSet xmlns=\"http://docs.oasis-open.org/wsn/b-2\">true</FixedTopicSet>") {
		t.Errorf("FixedTopicSet not true:\n%s", body)
	}

	if strings.Contains(body, "TopicExpressionDialect>") {
		t.Errorf("no subscription filtering is applied, so no topic dialects may be advertised:\n%s", body)
	}
}

// TestCreatePullPointSubscriptionWire drives the full create dance over
// HTTP and checks the SubscriptionReference shape, the time formats, and
// the requested termination.
func TestCreatePullPointSubscriptionWire(t *testing.T) {
	_, mux := eventsTestServer(t)

	status, body := postEventsSOAP(t, mux, "/onvif/events_service",
		`<CreatePullPointSubscription xmlns="http://www.onvif.org/ver10/events/wsdl">`+
			`<InitialTerminationTime>PT5M</InitialTerminationTime>`+
			`</CreatePullPointSubscription>`)

	if status != http.StatusOK {
		t.Fatalf("CreatePullPointSubscription status = %d, want 200\nbody: %s", status, body)
	}

	resp := decodeCreateResponse(t, body)

	addr := resp.SubscriptionReference.Address
	if !strings.HasPrefix(addr, "http://") || !strings.Contains(addr, "/onvif/events_service/sub/") {
		t.Errorf("SubscriptionReference address = %q, want under <base>/onvif/events_service/sub/", addr)
	}

	id := addr[strings.LastIndex(addr, "/")+1:]
	if len(id) < 16 {
		t.Errorf("subscription id %q too short (want an opaque token)", id)
	}

	now := time.Now().UTC()
	term, err := time.Parse(time.RFC3339, resp.TerminationTime)
	if err != nil {
		t.Fatalf("TerminationTime %q not RFC3339: %v", resp.TerminationTime, err)
	}

	if remaining := term.Sub(now); remaining < 4*time.Minute || remaining > 6*time.Minute {
		t.Errorf("requested PT5M termination, remaining = %v", remaining)
	}

	if _, err := time.Parse(time.RFC3339, resp.CurrentTime); err != nil {
		t.Errorf("CurrentTime %q not RFC3339: %v", resp.CurrentTime, err)
	}
}

// TestCreatePullPointSubscriptionDefaults: absent InitialTerminationTime
// grants the 1h default; an unparseable value is a Sender fault.
func TestCreatePullPointSubscriptionDefaults(t *testing.T) {
	_, mux := eventsTestServer(t)

	status, body := postEventsSOAP(t, mux, "/onvif/events_service",
		`<CreatePullPointSubscription xmlns="http://www.onvif.org/ver10/events/wsdl"/>`)
	if status != http.StatusOK {
		t.Fatalf("default create status = %d, want 200\nbody: %s", status, body)
	}

	resp := decodeCreateResponse(t, body)
	term, err := time.Parse(time.RFC3339, resp.TerminationTime)
	if err != nil {
		t.Fatalf("TerminationTime not RFC3339: %v", err)
	}

	if remaining := time.Until(term); remaining < 55*time.Minute || remaining > time.Hour {
		t.Errorf("default termination remaining = %v, want ~1h", remaining)
	}

	status, body = postEventsSOAP(t, mux, "/onvif/events_service",
		`<CreatePullPointSubscription xmlns="http://www.onvif.org/ver10/events/wsdl">`+
			`<InitialTerminationTime>whenever</InitialTerminationTime>`+
			`</CreatePullPointSubscription>`)
	if status != http.StatusBadRequest {
		t.Errorf("invalid InitialTerminationTime status = %d, want 400\nbody: %s", status, body)
	}

	if !strings.Contains(body, "Sender") {
		t.Errorf("invalid InitialTerminationTime must be a Sender fault:\n%s", body)
	}
}

// subscribeTo pulls the SubscriptionReference address out of a fresh
// create call (termination PT10M unless overridden).
func subscribeTo(t *testing.T, mux *http.ServeMux, termination string) string {
	t.Helper()

	create := `<CreatePullPointSubscription xmlns="http://www.onvif.org/ver10/events/wsdl">` +
		`<InitialTerminationTime>` + termination + `</InitialTerminationTime>` +
		`</CreatePullPointSubscription>`

	status, body := postEventsSOAP(t, mux, "/onvif/events_service", create)
	if status != http.StatusOK {
		t.Fatalf("create (%s) status = %d, want 200\nbody: %s", termination, status, body)
	}

	return decodeCreateResponse(t, body).SubscriptionReference.Address
}

// TestPullMessagesDeliversPublishedEvents pins the canonical
// double-layer notification payload (#83 server side; shape pinned by the
// spec sample in #82): wsnt:NotificationMessage > wsnt:Message >
// tt:Message with Source/Data SimpleItems, delivered in publish order and
// bounded by MessageLimit.
func TestPullMessagesDeliversPublishedEvents(t *testing.T) {
	srv, mux := eventsTestServer(t)
	sub := subscribeTo(t, mux, "PT10M")

	srv.PublishEvent(Event{
		Topic: "tns1:VideoSource/MotionAlarm",
		Source: []SimpleItem{
			{Name: "Source", Value: "CSI"},
		},
		Data: []SimpleItem{
			{Name: "State", Value: "true"},
			{Name: "Score", Value: "87"},
		},
	})
	srv.PublishEvent(Event{Topic: "tns1:VideoSource/SignalLoss"})

	// MessageLimit=1 returns only the first event; the rest stay queued.
	status, body := postEventsSOAP(t, mux, sub,
		`<PullMessages xmlns="http://www.onvif.org/ver10/events/wsdl">`+
			`<Timeout>PT2S</Timeout><MessageLimit>1</MessageLimit>`+
			`</PullMessages>`)
	if status != http.StatusOK {
		t.Fatalf("PullMessages status = %d, want 200\nbody: %s", status, body)
	}

	resp := decodePullResponse(t, body)
	if len(resp.NotificationMessages) != 1 {
		t.Fatalf("got %d notification messages with MessageLimit=1, want 1", len(resp.NotificationMessages))
	}

	first := resp.NotificationMessages[0]
	if first.Topic.Value != "tns1:VideoSource/MotionAlarm" {
		t.Errorf("topic = %q, want MotionAlarm", first.Topic.Value)
	}

	if first.Message.Inner.PropertyOperation != "Changed" {
		t.Errorf("PropertyOperation = %q, want Changed (the default)", first.Message.Inner.PropertyOperation)
	}

	if len(first.Message.Inner.Source.SimpleItems) != 1 ||
		first.Message.Inner.Source.SimpleItems[0] != (eventItemWire{Name: "Source", Value: "CSI"}) {
		t.Errorf("Source SimpleItems = %+v, want [{Source CSI}] (inner tt:Message layer)",
			first.Message.Inner.Source.SimpleItems)
	}

	if len(first.Message.Inner.Data.SimpleItems) != 2 ||
		first.Message.Inner.Data.SimpleItems[0] != (eventItemWire{Name: "State", Value: "true"}) ||
		first.Message.Inner.Data.SimpleItems[1] != (eventItemWire{Name: "Score", Value: "87"}) {
		t.Errorf("Data SimpleItems = %+v, want [State=true Score=87] (inner tt:Message layer)",
			first.Message.Inner.Data.SimpleItems)
	}

	if _, err := time.Parse(time.RFC3339, first.Message.Inner.UtcTime); err != nil {
		t.Errorf("inner UtcTime %q not RFC3339: %v", first.Message.Inner.UtcTime, err)
	}

	if first.ProducerReference.Address == "" {
		t.Errorf("ProducerReference address missing")
	}

	// The second pull drains the still-queued SignalLoss event.
	status, body = postEventsSOAP(t, mux, sub,
		`<PullMessages xmlns="http://www.onvif.org/ver10/events/wsdl">`+
			`<Timeout>PT1S</Timeout><MessageLimit>10</MessageLimit>`+
			`</PullMessages>`)
	if status != http.StatusOK {
		t.Fatalf("second PullMessages status = %d, want 200\nbody: %s", status, body)
	}

	resp = decodePullResponse(t, body)
	if len(resp.NotificationMessages) != 1 || resp.NotificationMessages[0].Topic.Value != "tns1:VideoSource/SignalLoss" {
		t.Errorf("second pull = %+v, want the queued SignalLoss event", resp.NotificationMessages)
	}
}

// TestPullMessagesLongPollWaits: an empty queue holds the request for the
// asked Timeout (long-poll), then answers empty.
func TestPullMessagesLongPollWaits(t *testing.T) {
	_, mux := eventsTestServer(t)
	sub := subscribeTo(t, mux, "PT10M")

	start := time.Now()
	status, body := postEventsSOAP(t, mux, sub,
		`<PullMessages xmlns="http://www.onvif.org/ver10/events/wsdl">`+
			`<Timeout>PT1S</Timeout><MessageLimit>5</MessageLimit>`+
			`</PullMessages>`)
	elapsed := time.Since(start)

	if status != http.StatusOK {
		t.Fatalf("PullMessages status = %d, want 200\nbody: %s", status, body)
	}

	if elapsed < 900*time.Millisecond {
		t.Errorf("empty PullMessages returned after %v, want it held ~1s (long poll)", elapsed)
	}

	resp := decodePullResponse(t, body)
	if len(resp.NotificationMessages) != 0 {
		t.Errorf("got %d messages on an empty queue, want 0", len(resp.NotificationMessages))
	}
}

// TestPullMessagesUnknownSubscription: posting to a subscription address
// that never existed must be a Sender fault, not a bare 404 or a 500.
func TestPullMessagesUnknownSubscription(t *testing.T) {
	_, mux := eventsTestServer(t)

	status, body := postEventsSOAP(t, mux, "/onvif/events_service/sub/deadbeefdeadbeef",
		`<PullMessages xmlns="http://www.onvif.org/ver10/events/wsdl">`+
			`<Timeout>PT1S</Timeout><MessageLimit>5</MessageLimit>`+
			`</PullMessages>`)

	if status != http.StatusBadRequest {
		t.Errorf("unknown subscription status = %d, want 400\nbody: %s", status, body)
	}

	if !strings.Contains(body, "<Value>Sender</Value>") {
		t.Errorf("unknown subscription must fault as Sender:\n%s", body)
	}
}

// TestRenewAndUnsubscribeSemantics: Renew extends the termination time;
// after Unsubscribe the endpoint faults.
func TestRenewAndUnsubscribeSemantics(t *testing.T) {
	_, mux := eventsTestServer(t)
	sub := subscribeTo(t, mux, "PT2M")

	status, body := postEventsSOAP(t, mux, sub,
		`<Renew xmlns="http://docs.oasis-open.org/wsn/b-2">`+
			`<TerminationTime>PT10M</TerminationTime>`+
			`</Renew>`)
	if status != http.StatusOK {
		t.Fatalf("Renew status = %d, want 200\nbody: %s", status, body)
	}

	var env struct {
		Body struct {
			Response struct {
				CurrentTime     string `xml:"CurrentTime"`
				TerminationTime string `xml:"TerminationTime"`
			} `xml:"RenewResponse"`
		} `xml:"Body"`
	}

	if err := xml.Unmarshal([]byte(body), &env); err != nil {
		t.Fatalf("decode RenewResponse: %v\nbody:\n%s", err, body)
	}

	term, err := time.Parse(time.RFC3339, env.Body.Response.TerminationTime)
	if err != nil {
		t.Fatalf("Renewed TerminationTime not RFC3339: %v", err)
	}

	if remaining := time.Until(term); remaining < 9*time.Minute || remaining > 11*time.Minute {
		t.Errorf("after Renew PT10M remaining = %v, want ~10m", remaining)
	}

	status, body = postEventsSOAP(t, mux, sub,
		`<Unsubscribe xmlns="http://docs.oasis-open.org/wsn/b-2"/>`)
	if status != http.StatusOK {
		t.Fatalf("Unsubscribe status = %d, want 200\nbody: %s", status, body)
	}

	if !strings.Contains(body, "UnsubscribeResponse") {
		t.Errorf("response missing UnsubscribeResponse:\n%s", body)
	}

	status, _ = postEventsSOAP(t, mux, sub,
		`<PullMessages xmlns="http://www.onvif.org/ver10/events/wsdl">`+
			`<Timeout>PT1S</Timeout><MessageLimit>5</MessageLimit>`+
			`</PullMessages>`)
	if status != http.StatusBadRequest {
		t.Errorf("PullMessages after Unsubscribe status = %d, want 400 (unknown subscription)", status)
	}
}

// TestMaxPullPointsEnforced: the registry caps concurrent pull points;
// the create beyond the cap is a Sender fault.
func TestMaxPullPointsEnforced(t *testing.T) {
	_, mux := eventsTestServer(t)

	for i := range defaultMaxPullPoints {
		sub := subscribeTo(t, mux, "PT10M")
		if sub == "" {
			t.Fatalf("create %d returned empty address", i)
		}
	}

	status, body := postEventsSOAP(t, mux, "/onvif/events_service",
		`<CreatePullPointSubscription xmlns="http://www.onvif.org/ver10/events/wsdl"/>`)

	if status != http.StatusBadRequest {
		t.Errorf("create beyond cap status = %d, want 400\nbody: %s", status, body)
	}

	if !strings.Contains(body, "<Value>Sender</Value>") {
		t.Errorf("cap exceeded must fault as Sender:\n%s", body)
	}
}

// TestExpiredSubscriptionRejected: a subscription past its termination
// time is pruned lazily — later operations fault as unknown.
func TestExpiredSubscriptionRejected(t *testing.T) {
	_, mux := eventsTestServer(t)
	sub := subscribeTo(t, mux, "PT1S")

	time.Sleep(1200 * time.Millisecond)

	status, _ := postEventsSOAP(t, mux, sub,
		`<PullMessages xmlns="http://www.onvif.org/ver10/events/wsdl">`+
			`<Timeout>PT1S</Timeout><MessageLimit>5</MessageLimit>`+
			`</PullMessages>`)

	if status != http.StatusBadRequest {
		t.Errorf("expired subscription status = %d, want 400 (Sender fault)", status)
	}
}

// TestPublishEventFanout: every live subscription receives the event; a
// subscriber created after the publish does not.
func TestPublishEventFanout(t *testing.T) {
	srv, mux := eventsTestServer(t)

	first := subscribeTo(t, mux, "PT10M")
	second := subscribeTo(t, mux, "PT10M")

	srv.PublishEvent(Event{Topic: "tns1:Device/HardwareFailure"})

	for name, sub := range map[string]string{"first": first, "second": second} {
		status, body := postEventsSOAP(t, mux, sub,
			`<PullMessages xmlns="http://www.onvif.org/ver10/events/wsdl">`+
				`<Timeout>PT1S</Timeout><MessageLimit>5</MessageLimit>`+
				`</PullMessages>`)
		if status != http.StatusOK {
			t.Fatalf("%s pull status = %d\nbody: %s", name, status, body)
		}

		resp := decodePullResponse(t, body)
		if len(resp.NotificationMessages) != 1 ||
			resp.NotificationMessages[0].Topic.Value != "tns1:Device/HardwareFailure" {
			t.Errorf("%s got %+v, want the fanned-out event", name, resp.NotificationMessages)
		}
	}
}

// TestEventsClientServerInterop drives the library's own events client
// against the server end to end (#83): create → publish → pull → renew →
// unsubscribe. Client-side Message payload parsing is covered by #82
// (separate fix); here the Topic level plus the raw wire payload are
// pinned.
func TestEventsClientServerInterop(t *testing.T) {
	// The advertised port is config.Port (a server-side concern); bind
	// the test listener first so both agree.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()

	config := createTestConfig()
	config.SupportEvents = true
	config.Port = listener.Addr().(*net.TCPAddr).Port

	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	mux := http.NewServeMux()
	srv.RegisterServices(mux)

	ts := httptest.NewUnstartedServer(mux)
	ts.Listener = listener
	ts.Start()
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := onvif.NewClient(
		ts.URL+"/onvif/device_service",
		onvif.WithCredentials("admin", "password"),
		onvif.WithTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("onvif.NewClient() error = %v", err)
	}

	if err := client.Initialize(ctx); err != nil {
		t.Fatalf("client.Initialize() error = %v", err)
	}

	duration := 5 * time.Minute
	sub, err := client.Events().CreatePullPointSubscription(ctx, "", &duration, "")
	if err != nil {
		t.Fatalf("CreatePullPointSubscription() error = %v", err)
	}

	if !strings.Contains(sub.SubscriptionReference, "/onvif/events_service/sub/") {
		t.Errorf("client-visible SubscriptionReference = %q, want an events_service/sub/ address",
			sub.SubscriptionReference)
	}

	srv.PublishEvent(Event{
		Topic: "tns1:VideoSource/MotionAlarm",
		Source: []SimpleItem{
			{Name: "Source", Value: "CSI"},
		},
		Data: []SimpleItem{
			{Name: "State", Value: "true"},
			{Name: "Score", Value: "87"},
		},
	})

	messages, err := client.Events().PullMessages(ctx, sub.SubscriptionReference, 2*time.Second, 10)
	if err != nil {
		t.Fatalf("PullMessages() error = %v", err)
	}

	if len(messages) != 1 || messages[0].Topic != "tns1:VideoSource/MotionAlarm" {
		t.Fatalf("client pulled %+v, want the MotionAlarm notification", messages)
	}

	// The full double-layer payload must survive the client's parser
	// (issue #82: before that fix only the Topic made it through).
	pulled := messages[0].Message
	if pulled.PropertyOperation != "Changed" {
		t.Errorf("client PropertyOperation = %q, want Changed", pulled.PropertyOperation)
	}

	if len(pulled.Source) != 1 || pulled.Source[0] != (types.SimpleItem{Name: "Source", Value: "CSI"}) {
		t.Errorf("client Source = %+v, want [{Source CSI}]", pulled.Source)
	}

	if len(pulled.Data) != 2 ||
		pulled.Data[0] != (types.SimpleItem{Name: "State", Value: "true"}) ||
		pulled.Data[1] != (types.SimpleItem{Name: "Score", Value: "87"}) {
		t.Errorf("client Data = %+v, want [State=true Score=87]", pulled.Data)
	}

	if pulled.UtcTime.IsZero() {
		t.Error("client UtcTime not parsed from the inner tt:Message")
	}

	_, term, err := client.Events().RenewSubscription(ctx, sub.SubscriptionReference, 10*time.Minute)
	if err != nil {
		t.Fatalf("RenewSubscription() error = %v", err)
	}

	if time.Until(term) < 9*time.Minute {
		t.Errorf("renewed termination too early: %v", term)
	}

	if err := client.Events().Unsubscribe(ctx, sub.SubscriptionReference); err != nil {
		t.Fatalf("Unsubscribe() error = %v", err)
	}

	if _, err := client.Events().PullMessages(ctx, sub.SubscriptionReference, time.Second, 5); err == nil {
		t.Error("PullMessages after Unsubscribe must fail")
	}
}

// TestGetCapabilitiesEventsPullPointFlag: with the pull-point service
// implemented, the capability flags must say so — advertising
// WSPullPointSupport=false while serving pull points would repeat the
// #46/#83 fake-advertisement mistake in the other direction.
func TestGetCapabilitiesEventsPullPointFlag(t *testing.T) {
	config := createTestConfig()
	config.SupportEvents = true

	srv, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, err := srv.HandleGetCapabilities(nil, nil)
	if err != nil {
		t.Fatalf("HandleGetCapabilities() error = %v", err)
	}

	events := resp.(*GetCapabilitiesResponse).Capabilities.Events
	if events == nil {
		t.Fatal("Events capability missing with SupportEvents=true")
	}

	if !events.WSPullPointSupport {
		t.Error("WSPullPointSupport = false, want true (pull point is implemented)")
	}
}

// TestParseISO8601Duration pins the duration grammar accepted for
// InitialTerminationTime / Renew TerminationTime / PullMessages Timeout.
func TestParseISO8601Duration(t *testing.T) {
	cases := []struct {
		in      string
		want    time.Duration
		wantErr bool
	}{
		{in: "PT30S", want: 30 * time.Second},
		{in: "PT5M", want: 5 * time.Minute},
		{in: "PT1M30S", want: 90 * time.Second},
		{in: "PT1H", want: time.Hour},
		{in: "PT1H5M10S", want: time.Hour + 5*time.Minute + 10*time.Second},
		{in: "P1DT2H", want: 26 * time.Hour},
		{in: "PT0S", want: 0},
		{in: "", wantErr: true},
		{in: "5M", wantErr: true},
		{in: "PTxS", wantErr: true},
		{in: "PT", wantErr: true},
		{in: "P1D", wantErr: true}, // date-only carries no time component for these uses
	}

	for _, tc := range cases {
		got, err := parseISO8601Duration(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseISO8601Duration(%q) = %v, want error", tc.in, got)
			}

			continue
		}

		if err != nil {
			t.Errorf("parseISO8601Duration(%q) error = %v", tc.in, err)

			continue
		}

		if got != tc.want {
			t.Errorf("parseISO8601Duration(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// compile-time guard: the SOAP layer must expose the Sender-fault channel
// used by the events handlers for client mistakes.
var _ error = (*soap.SenderFaultError)(nil)
