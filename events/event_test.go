package events

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/internal/soap"
)

// fakeCaller is an in-memory api.Caller: it records requests and answers
// from a pluggable handler — no sockets, microsecond latency, and the
// exact decode paths of the real transport (xml.Unmarshal into the
// response pointer).
type fakeCaller struct {
	mu      sync.Mutex
	calls   []string // request element names, in order
	handler func(action string, reqXML string) (respXML string, err error)
}

func newFakeCaller(handler func(action, reqXML string) (string, error)) *fakeCaller {
	return &fakeCaller{handler: handler}
}

func (f *fakeCaller) Call(_ context.Context, _, _ string, request, response interface{}) error {
	reqXML, err := xml.Marshal(request)
	if err != nil {
		return fmt.Errorf("fake caller marshal: %w", err)
	}

	action := ""
	if start := strings.Index(string(reqXML), "<"); start >= 0 {
		rest := string(reqXML[start+1:])
		if end := strings.Index(rest, ">"); end > 0 {
			action = strings.SplitN(rest[:end], " ", 2)[0]
		}
	}

	f.mu.Lock()
	f.calls = append(f.calls, action)
	handler := f.handler
	f.mu.Unlock()

	respXML, herr := handler(action, string(reqXML))
	if herr != nil {
		return herr
	}

	if respXML != "" {
		if err := xml.Unmarshal([]byte(respXML), response); err != nil {
			return fmt.Errorf("fake caller unmarshal: %w", err)
		}
	}

	return nil
}

func (f *fakeCaller) EndpointFor(api.Service) string { return "http://fake/events" }

func (f *fakeCaller) recorded() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]string(nil), f.calls...)
}

func (f *fakeCaller) countAction(action string) int {
	n := 0

	for _, c := range f.recorded() {
		if c == action {
			n++
		}
	}

	return n
}

// waitFor is the bounded-wait guard every channel-based test must use:
// a deadlock or missed close fails the test in bounded time instead of
// hanging the suite.
func waitFor(t *testing.T, what string, ch <-chan struct{}) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatalf("%s: timed out waiting (possible deadlock)", what)
	}
}

func mustReceive[T any](t *testing.T, what string, ch <-chan T) T {
	t.Helper()

	select {
	case v := <-ch:
		return v
	case <-time.After(5 * time.Second):
		t.Fatalf("%s: timed out waiting (possible deadlock)", what)

		var zero T

		return zero
	}
}

// --- canned response snippets ---

// createSubResponse renders a subscription answer whose termination is
// offset from now (timezone-independent fixtures).
func createSubResponse(terminationOffset time.Duration) string {
	now := time.Now().UTC()

	return `<CreatePullPointSubscriptionResponse>
	<SubscriptionReference><Address>http://fake/sub/1</Address></SubscriptionReference>
	<CurrentTime>` + now.Format(time.RFC3339) + `</CurrentTime>
	<TerminationTime>` + now.Add(terminationOffset).Format(time.RFC3339) + `</TerminationTime>
</CreatePullPointSubscriptionResponse>`
}

var (
	pullMessagesOne = `<PullMessagesResponse>
	<NotificationMessage>
		<Topic xmlns:tns1="http://www.onvif.org/ver10/topics">tns1:VideoSource/MotionAlarm</Topic>
		<ProducerReference><Address>urn:uuid:producer-1</Address></ProducerReference>
		<Message UtcTime="2026-08-28T10:00:01Z" PropertyOperation="Changed">
			<Source><SimpleItem Name="Source" Value="VideoSource_1"/></Source>
			<Data><SimpleItem Name="State" Value="true"/></Data>
		</Message>
	</NotificationMessage>
</PullMessagesResponse>`

	pullMessagesEmpty = `<PullMessagesResponse>
</PullMessagesResponse>`

	unsubscribeResponse = `<UnsubscribeResponse/>`

	capabilitiesResponse = `<GetServiceCapabilitiesResponse>
	<Capabilities WSSubscriptionPolicySupport="true" MaxPullPoints="8" MaxNotificationProducers="16"
		PersistentStorage="true" EventBrokerProtocols="mqtt mqtts" MetadataOverMQTT="false"/>
</GetServiceCapabilitiesResponse>`

	propertiesResponse = `<GetEventPropertiesResponse>
	<TopicNamespaceLocation>http://www.onvif.org/onvif/ver10/topics/topicns.xml</TopicNamespaceLocation>
	<FixedTopicSet>true</FixedTopicSet>
	<TopicSet>
		<VideoSource><MotionAlarm><State/></MotionAlarm></VideoSource>
	</TopicSet>
	<TopicExpressionDialect>http://docs.oasis-open.org/wsn/t-1/TopicExpression/Concrete</TopicExpressionDialect>
	<MessageContentFilterDialect>http://www.onvif.org/ver10/tev/messageContentFilter/ItemFilter</MessageContentFilterDialect>
</GetEventPropertiesResponse>`
)

// renewResponse renders a fresh future termination each call.
func renewResponse() string {
	return `<RenewResponse>
	<TerminationTime>` + time.Now().UTC().Add(time.Hour).Format(time.RFC3339) + `</TerminationTime>
</RenewResponse>`
}

func TestGetEventServiceCapabilitiesParses(t *testing.T) {
	caller := newFakeCaller(func(action, _ string) (string, error) {
		if action != "tev:GetServiceCapabilities" {
			return "", fmt.Errorf("unexpected action %q", action)
		}

		return capabilitiesResponse, nil
	})

	caps, err := New(caller).GetEventServiceCapabilities(context.Background())
	if err != nil {
		t.Fatalf("GetEventServiceCapabilities: %v", err)
	}

	if !caps.WSSubscriptionPolicySupport || caps.MaxPullPoints != 8 || caps.MaxNotificationProducers != 16 {
		t.Errorf("capability attrs not parsed: %+v", caps)
	}

	if len(caps.EventBrokerProtocols) != 2 || caps.EventBrokerProtocols[0] != "mqtt" {
		t.Errorf("EventBrokerProtocols not split: %v", caps.EventBrokerProtocols)
	}
}

func TestCreatePullPointSubscriptionParsesTimes(t *testing.T) {
	caller := newFakeCaller(func(_, reqXML string) (string, error) {
		if !strings.Contains(reqXML, "InitialTerminationTime") {
			t.Errorf("InitialTerminationTime not sent: %s", reqXML)
		}

		return createSubResponse(time.Hour), nil
	})

	duration := 30 * time.Minute
	sub, err := New(caller).CreatePullPointSubscription(context.Background(), "tns1:VideoSource//.", &duration, "")
	if err != nil {
		t.Fatalf("CreatePullPointSubscription: %v", err)
	}

	if sub.SubscriptionReference != "http://fake/sub/1" {
		t.Errorf("SubscriptionReference = %q", sub.SubscriptionReference)
	}

	if age := time.Since(sub.TerminationTime); age < -61*time.Minute || age > -59*time.Minute {
		t.Errorf("TerminationTime = %v, want ~1h from now", sub.TerminationTime)
	}
}

func TestCreatePullPointSubscriptionValidation(t *testing.T) {
	svc := New(newFakeCaller(func(string, string) (string, error) { return "", nil }))

	bad := -time.Second
	if _, err := svc.CreatePullPointSubscription(context.Background(), "", &bad, ""); !errors.Is(err, ErrInvalidTerminationTime) {
		t.Errorf("negative duration error = %v, want ErrInvalidTerminationTime", err)
	}
}

func TestPullMessagesParsesAndValidates(t *testing.T) {
	caller := newFakeCaller(func(action, reqXML string) (string, error) {
		if action != "tev:PullMessages" {
			return "", fmt.Errorf("unexpected action %q", action)
		}

		if !strings.Contains(reqXML, "<tev:Timeout>PT10S</tev:Timeout>") {
			t.Errorf("timeout not encoded: %s", reqXML)
		}

		return pullMessagesOne, nil
	})

	svc := New(caller)

	messages, err := svc.PullMessages(context.Background(), "http://fake/sub/1", 10*time.Second, 5)
	if err != nil {
		t.Fatalf("PullMessages: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("messages = %d, want 1", len(messages))
	}

	msg := messages[0]
	if msg.Topic != "tns1:VideoSource/MotionAlarm" {
		t.Errorf("Topic = %q", msg.Topic)
	}

	if msg.Message.PropertyOperation != "Changed" {
		t.Errorf("PropertyOperation = %q", msg.Message.PropertyOperation)
	}

	if len(msg.Message.Data) != 1 || msg.Message.Data[0].Name != "State" || msg.Message.Data[0].Value != "true" {
		t.Errorf("Data items = %+v", msg.Message.Data)
	}

	// Validation errors.
	if _, err := svc.PullMessages(context.Background(), "", time.Second, 1); !errors.Is(err, ErrInvalidSubscriptionReference) {
		t.Errorf("empty ref error = %v", err)
	}

	if _, err := svc.PullMessages(context.Background(), "ref", 0, 1); !errors.Is(err, ErrInvalidTimeout) {
		t.Errorf("zero timeout error = %v", err)
	}

	if _, err := svc.PullMessages(context.Background(), "ref", time.Second, 0); !errors.Is(err, ErrInvalidMessageLimit) {
		t.Errorf("zero limit error = %v", err)
	}
}

func TestRenewSubscriptionValidationAndParse(t *testing.T) {
	caller := newFakeCaller(func(action, _ string) (string, error) {
		if action != "wsnt:Renew" {
			return "", fmt.Errorf("unexpected action %q", action)
		}

		return renewResponse(), nil
	})

	svc := New(caller)

	if _, _, err := svc.RenewSubscription(context.Background(), "", time.Minute); !errors.Is(err, ErrInvalidSubscriptionReference) {
		t.Errorf("empty ref error = %v", err)
	}

	if _, _, err := svc.RenewSubscription(context.Background(), "ref", 0); !errors.Is(err, ErrInvalidTerminationTime) {
		t.Errorf("zero duration error = %v", err)
	}

	_, termination, err := svc.RenewSubscription(context.Background(), "ref", time.Hour)
	if err != nil {
		t.Fatalf("RenewSubscription: %v", err)
	}

	if age := time.Since(termination); age < -61*time.Minute || age > -59*time.Minute {
		t.Errorf("termination = %v, want ~1h from now", termination)
	}
}

func TestUnsubscribeSendsReference(t *testing.T) {
	caller := newFakeCaller(func(action, _ string) (string, error) {
		if action != "wsnt:Unsubscribe" {
			return "", fmt.Errorf("unexpected action %q", action)
		}

		return unsubscribeResponse, nil
	})

	if err := New(caller).Unsubscribe(context.Background(), "http://fake/sub/9"); err != nil {
		t.Fatalf("Unsubscribe: %v", err)
	}

	if caller.countAction("wsnt:Unsubscribe") != 1 {
		t.Errorf("unsubscribe action count = %d, want 1", caller.countAction("wsnt:Unsubscribe"))
	}
}

func TestGetEventPropertiesParses(t *testing.T) {
	caller := newFakeCaller(func(action, _ string) (string, error) {
		if action != "tev:GetEventProperties" {
			return "", fmt.Errorf("unexpected action %q", action)
		}

		return propertiesResponse, nil
	})

	props, err := New(caller).GetEventProperties(context.Background())
	if err != nil {
		t.Fatalf("GetEventProperties: %v", err)
	}

	if !props.FixedTopicSet {
		t.Error("FixedTopicSet not parsed")
	}

	if len(props.TopicExpressionDialects) != 1 {
		t.Errorf("dialects = %v", props.TopicExpressionDialects)
	}
}

func TestClassifyEventsNotSupported(t *testing.T) {
	tests := []struct {
		code, subcode, reason string
		want                  bool
	}{
		{code: "soap:Sender", subcode: "tev:ActionNotSupported", reason: "", want: true},
		{code: "soap:Receiver", subcode: "", reason: "Action Not Implemented", want: true},
		{code: "soap:Sender", subcode: "ter:NotSupportedByDevice", reason: "", want: true},
		{code: "soap:Sender", subcode: "ter:Unauthorized", reason: "", want: false},
	}

	for _, tt := range tests {
		err := &soap.FaultError{Code: tt.code, Subcode: tt.subcode, Reason: tt.reason}

		if got := classifyEventsNotSupported(err); got != tt.want {
			t.Errorf("classify(%q,%q,%q) = %v, want %v", tt.code, tt.subcode, tt.reason, got, tt.want)
		}
	}

	if classifyEventsNotSupported(errors.New("plain")) {
		t.Error("non-fault error must not classify as not-supported")
	}
}

// --- managed EventStream lifecycle ---

// subscriptionScript drives the fake device for lifecycle tests: create
// answers once, then pulls deliver messages or fail per the script.
type subscriptionScript struct {
	mu           sync.Mutex
	pullCount    int
	pullResponse func(pull int) (string, error)
	renewErr     error
	unsubscribes int
}

func (s *subscriptionScript) handler(action, _ string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch action {
	case "tev:CreatePullPointSubscription":
		return createSubResponse(time.Hour), nil
	case "tev:PullMessages":
		s.pullCount++

		return s.pullResponse(s.pullCount)
	case "wsnt:Renew":
		if s.renewErr != nil {
			return "", s.renewErr
		}

		return renewResponse(), nil
	case "wsnt:Unsubscribe":
		s.unsubscribes++

		return unsubscribeResponse, nil
	default:
		return "", fmt.Errorf("unexpected action %q", action)
	}
}

// pullCount is accessed under s.mu via handler; expose a getter.
func (s *subscriptionScript) pulls() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.pullCount
}

// fastStreamOptions keeps lifecycle tests inside tens of milliseconds:
// the subscription is created with an 11:00 termination time far in the
// past only via renewFailure tests; for happy paths we rely on the
// createSubResponse termination being in the past relative to the test
// clock — so every loop pass renews first. That is the exercised path
// anyway (renewal-before-expiry).
func fastStreamOptions() *SubscribeEventsOptions {
	return &SubscribeEventsOptions{
		SubscriptionDuration: time.Hour,
		PullTimeout:          30 * time.Millisecond,
		MessageLimit:         5,
	}
}

func TestEventStreamDeliversAndStops(t *testing.T) {
	script := &subscriptionScript{
		pullResponse: func(pull int) (string, error) {
			if pull <= 2 {
				return pullMessagesOne, nil
			}

			return pullMessagesEmpty, nil
		},
	}
	caller := newFakeCaller(script.handler)
	svc := New(caller)

	delivered := make(chan NotificationMessage, 8)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := svc.SubscribeEvents(ctx, func(msg NotificationMessage) {
		delivered <- msg
	}, fastStreamOptions())
	if err != nil {
		t.Fatalf("SubscribeEvents: %v", err)
	}

	first := mustReceive(t, "first message", delivered)
	if first.Topic != "tns1:VideoSource/MotionAlarm" {
		t.Errorf("topic = %q", first.Topic)
	}

	mustReceive(t, "second message", delivered)

	// Explicit unsubscribe: loop stops, SOAP unsubscribe sent exactly once.
	if err := stream.Unsubscribe(context.Background()); err != nil {
		t.Fatalf("Unsubscribe: %v", err)
	}

	waitFor(t, "Done after unsubscribe", stream.Done())

	script.mu.Lock()
	unsubscribes := script.unsubscribes
	script.mu.Unlock()

	if unsubscribes != 1 {
		t.Errorf("SOAP unsubscribe count = %d, want exactly 1 (no cleanup duplicate)", unsubscribes)
	}
}

func TestEventStreamRenewalFailureTerminates(t *testing.T) {
	script := &subscriptionScript{
		pullResponse: func(int) (string, error) { return pullMessagesEmpty, nil },
		renewErr:     errors.New("device refuses renewal"),
	}
	// Expiry in the past → renewIn <= 0 on the first pass → renewal
	// fails → loop terminates with a best-effort unsubscribe.
	caller := newFakeCaller(func(action, req string) (string, error) {
		if action == "tev:CreatePullPointSubscription" {
			return createSubResponse(-time.Minute), nil
		}

		return script.handler(action, req)
	})

	stream, err := New(caller).SubscribeEvents(context.Background(), func(NotificationMessage) {}, fastStreamOptions())
	if err != nil {
		t.Fatalf("SubscribeEvents: %v", err)
	}

	waitFor(t, "Done after renewal failure", stream.Done())

	script.mu.Lock()
	unsubscribes := script.unsubscribes
	script.mu.Unlock()

	if unsubscribes != 1 {
		t.Errorf("best-effort unsubscribe count = %d, want 1", unsubscribes)
	}
}

func TestEventStreamPanicIsolation(t *testing.T) {
	script := &subscriptionScript{
		pullResponse: func(int) (string, error) { return pullMessagesOne, nil },
	}
	caller := newFakeCaller(script.handler)

	delivered := make(chan int, 8)
	n := 0

	stream, err := New(caller).SubscribeEvents(context.Background(), func(NotificationMessage) {
		n++
		if n == 1 {
			panic("handler bug must not kill the loop")
		}

		// Non-blocking send: the fake device answers every pull with a
		// message, but the test stops draining after message #2 — a
		// blocking send would wedge the handler forever and Done() could
		// never close (the loop cannot preempt a blocked handler).
		select {
		case delivered <- n:
		default:
		}
	}, fastStreamOptions())
	if err != nil {
		t.Fatalf("SubscribeEvents: %v", err)
	}
	defer func() {
		_ = stream.Unsubscribe(context.Background())
		waitFor(t, "Done", stream.Done())
	}()

	// The second message must still arrive despite the first panic.
	select {
	case got := <-delivered:
		if got != 2 {
			t.Errorf("got message #%d, want 2", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("loop died after handler panic (isolation broken)")
	}
}

func TestEventStreamContextCancellation(t *testing.T) {
	script := &subscriptionScript{
		pullResponse: func(int) (string, error) { return pullMessagesEmpty, nil },
	}
	caller := newFakeCaller(script.handler)

	ctx, cancel := context.WithCancel(context.Background())

	stream, err := New(caller).SubscribeEvents(ctx, func(NotificationMessage) {}, fastStreamOptions())
	if err != nil {
		t.Fatalf("SubscribeEvents: %v", err)
	}

	cancel()
	waitFor(t, "Done after context cancellation", stream.Done())

	script.mu.Lock()
	unsubscribes := script.unsubscribes
	script.mu.Unlock()

	if unsubscribes != 1 {
		t.Errorf("cleanup unsubscribe count = %d, want 1", unsubscribes)
	}
}

func TestSubscribeEventsNotSupported(t *testing.T) {
	caller := newFakeCaller(func(string, string) (string, error) {
		return "", &soap.FaultError{Code: "soap:Sender", Subcode: "tev:ActionNotSupported"}
	})

	_, err := New(caller).SubscribeEvents(context.Background(), func(NotificationMessage) {}, nil)
	if !errors.Is(err, ErrEventsNotSupported) {
		t.Errorf("error = %v, want ErrEventsNotSupported", err)
	}
}

func TestSubscribeEventsNilHandler(t *testing.T) {
	if _, err := New(newFakeCaller(nil)).SubscribeEvents(context.Background(), nil, nil); err == nil {
		t.Error("nil handler must be rejected")
	}
}

func TestEventStreamTransientPullErrorRecovers(t *testing.T) {
	script := &subscriptionScript{
		pullResponse: func(pull int) (string, error) {
			if pull == 1 {
				return "", errors.New("transient network blip")
			}

			return pullMessagesOne, nil
		},
	}
	caller := newFakeCaller(script.handler)

	// The handler must not block the loop goroutine: this script keeps
	// returning messages after recovery, and a raw `delivered <- m` send
	// wedges the loop once the buffer fills — cancellation cannot preempt
	// a channel send, so Unsubscribe would never observe a stopped loop
	// (the CI flake this case fixed). Non-blocking delivery models the
	// real consumer contract.
	delivered := make(chan NotificationMessage, 4)

	stream, err := New(caller).SubscribeEvents(context.Background(), func(m NotificationMessage) {
		select {
		case delivered <- m:
		default:
		}
	}, fastStreamOptions())
	if err != nil {
		t.Fatalf("SubscribeEvents: %v", err)
	}
	defer func() {
		_ = stream.Unsubscribe(context.Background())
		waitFor(t, "Done", stream.Done())
	}()

	// The first pull errors and backs off (1s start); the message after
	// recovery must still arrive.
	select {
	case <-delivered:
	case <-time.After(5 * time.Second):
		t.Fatal("loop did not recover after a transient pull error")
	}
}

func TestEventStreamUnsubscribeSOAPFailureStillStops(t *testing.T) {
	script := &subscriptionScript{
		pullResponse: func(int) (string, error) { return pullMessagesEmpty, nil },
	}
	caller := newFakeCaller(func(action, req string) (string, error) {
		if action == "wsnt:Unsubscribe" {
			return "", errors.New("device unreachable")
		}

		return script.handler(action, req)
	})

	stream, err := New(caller).SubscribeEvents(context.Background(), func(NotificationMessage) {}, fastStreamOptions())
	if err != nil {
		t.Fatalf("SubscribeEvents: %v", err)
	}

	if err := stream.Unsubscribe(context.Background()); err == nil {
		t.Error("SOAP unsubscribe failure must surface")
	}

	waitFor(t, "Done despite SOAP failure", stream.Done())
}

func TestEventStreamIdleThrottle(t *testing.T) {
	script := &subscriptionScript{
		pullResponse: func(int) (string, error) { return pullMessagesEmpty, nil },
	}
	caller := newFakeCaller(script.handler)

	stream, err := New(caller).SubscribeEvents(context.Background(), func(NotificationMessage) {}, fastStreamOptions())
	if err != nil {
		t.Fatalf("SubscribeEvents: %v", err)
	}
	defer func() {
		_ = stream.Unsubscribe(context.Background())
		waitFor(t, "Done", stream.Done())
	}()

	// Instantly-empty pulls must be throttled by eventIdleGap: roughly
	// bounded by elapsed/pulls, far below an unthrottled spin.
	time.Sleep(350 * time.Millisecond)

	pulls := script.pulls()
	if pulls == 0 {
		t.Fatal("no pulls happened")
	}

	// 350ms with 100ms gaps → at most ~4-5 pulls (allow slack, catch
	// only a full-speed spin which would be 10k+).
	if pulls > 20 {
		t.Errorf("idle loop spun: %d pulls in 350ms (throttle broken)", pulls)
	}
}
