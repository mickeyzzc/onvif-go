package onvif

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// eventsMock is a pull-point device with scriptable behavior.
type eventsMock struct {
	mu sync.Mutex
	// counters
	creates, pulls, renews, unsubscribes int
	// scripts
	createFault   string // when non-empty, CreatePullPointSubscription faults
	renewFault    string // when non-empty, Renew faults
	pullFailures  int    // first N pulls fail with HTTP 500
	messageEveryN int    // deliver a message on every Nth pull (0 = never)

	pullCount int
}

func (m *eventsMock) handler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	b := string(body)

	m.mu.Lock()
	defer m.mu.Unlock()

	switch {
	case strings.Contains(b, "CreatePullPointSubscription"):
		m.creates++
		if m.createFault != "" {
			w.Write([]byte(faultEnvelope(m.createFault)))

			return
		}

		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<tev:CreatePullPointSubscriptionResponse xmlns:tev="http://www.onvif.org/ver10/events/wsdl" xmlns:wsa="http://www.w3.org/2005/08/addressing">
<tev:SubscriptionReference><wsa:Address>http://` + r.Host + `/subscriptions/1</wsa:Address></tev:SubscriptionReference>
<wsnt:CurrentTime xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2">2026-01-01T00:00:00Z</wsnt:CurrentTime>
<wsnt:TerminationTime xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2">2026-01-01T01:00:00Z</wsnt:TerminationTime>
</tev:CreatePullPointSubscriptionResponse></s:Body></s:Envelope>`))

	case strings.Contains(b, "PullMessages"):
		m.pullCount++
		m.pulls++
		if m.pullFailures > 0 && m.pullCount <= m.pullFailures {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("boom"))

			return
		}

		if m.messageEveryN > 0 && m.pullCount%m.messageEveryN == 0 {
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<tev:PullMessagesResponse xmlns:tev="http://www.onvif.org/ver10/events/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema">
<tev:NotificationMessage>
<wsnt:Topic xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2">tns1:VideoSource/MotionAlarm</wsnt:Topic>
<tev:Message><tt:Source><tt:SimpleItem Name="Source" Value="Camera1"/></tt:Source>
<tt:Data><tt:SimpleItem Name="State" Value="active"/></tt:Data></tev:Message>
</tev:NotificationMessage>
</tev:PullMessagesResponse></s:Body></s:Envelope>`))

			return
		}

		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<tev:PullMessagesResponse xmlns:tev="http://www.onvif.org/ver10/events/wsdl">
</tev:PullMessagesResponse></s:Body></s:Envelope>`))

	case strings.Contains(b, "Renew"):
		m.renews++
		if m.renewFault != "" {
			w.Write([]byte(faultEnvelope(m.renewFault)))

			return
		}

		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<wsnt:RenewResponse xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2">
<wsnt:TerminationTime>2026-01-01T02:00:00Z</wsnt:TerminationTime>
</wsnt:RenewResponse></s:Body></s:Envelope>`))

	case strings.Contains(b, "Unsubscribe"):
		m.unsubscribes++
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<wsnt:UnsubscribeResponse xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2"/>
</s:Body></s:Envelope>`))

	default:
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "unexpected op: %.100s", b)
	}
}

func faultEnvelope(text string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<s:Fault><s:Code><s:Value>s:Sender</s:Value><s:Subcode><s:Value>ter:ActionNotSupported</s:Value></s:Subcode></s:Code>
<s:Reason><s:Text>` + text + `</s:Text></s:Reason></s:Fault>
</s:Body></s:Envelope>`
}

func newEventsClient(t *testing.T, mock *eventsMock) *Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(mock.handler))
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	return client
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("condition not met within timeout")
}

func fastEventOpts() *SubscribeEventsOptions {
	return &SubscribeEventsOptions{
		SubscriptionDuration: 400 * time.Millisecond,
		RenewMargin:          350 * time.Millisecond,
		PullTimeout:          50 * time.Millisecond,
		MessageLimit:         5,
	}
}

func TestSubscribeEventsFullLifecycle(t *testing.T) {
	mock := &eventsMock{messageEveryN: 2}
	client := newEventsClient(t, mock)

	var mu sync.Mutex
	var got []NotificationMessage

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := client.Events().SubscribeEvents(ctx, func(msg NotificationMessage) {
		mu.Lock()
		defer mu.Unlock()
		got = append(got, msg)
	}, fastEventOpts())
	if err != nil {
		t.Fatalf("SubscribeEvents() error = %v", err)
	}

	// Messages must flow (every 2nd pull carries one).
	waitFor(t, 5*time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()

		return len(got) >= 2
	})

	mu.Lock()
	first := got[0]
	mu.Unlock()

	if first.Topic != "tns1:VideoSource/MotionAlarm" {
		t.Errorf("Topic = %q, want motion alarm", first.Topic)
	}

	if !first.Message.UtcTime.IsZero() == false && first.Message.UtcTime.IsZero() {
		t.Error("UtcTime left zero")
	}

	if len(first.Message.Data) != 1 || first.Message.Data[0].Name != "State" {
		t.Errorf("Data parsed wrong: %+v", first.Message.Data)
	}

	// Renewal must fire before the short subscription expires.
	waitFor(t, 5*time.Second, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()

		return mock.renews >= 1
	})

	// Unsubscribe stops the loop and hits the device.
	if err := sub.Unsubscribe(context.Background()); err != nil {
		t.Fatalf("Unsubscribe() error = %v", err)
	}

	select {
	case <-sub.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("loop did not exit after Unsubscribe")
	}

	mock.mu.Lock()
	defer mock.mu.Unlock()

	if mock.unsubscribes != 1 {
		t.Errorf("device saw %d unsubscribes, want 1", mock.unsubscribes)
	}
}

func TestSubscribeEventsNotSupported(t *testing.T) {
	for _, phrasing := range []string{
		"Action Not Implemented",
		"action not supported by device",
		"NotImplemented",
	} {
		t.Run(phrasing, func(t *testing.T) {
			mock := &eventsMock{createFault: phrasing}
			client := newEventsClient(t, mock)

			_, err := client.Events().SubscribeEvents(context.Background(),
				func(NotificationMessage) {}, fastEventOpts())
			if err == nil {
				t.Fatal("SubscribeEvents() succeeded against an events-less device")
			}

			if !errors.Is(err, ErrEventsNotSupported) {
				t.Errorf("errors.Is(err, ErrEventsNotSupported) = false, err = %v", err)
			}
		})
	}
}

func TestSubscribeEventsTransientPullFailuresKeepLoopAlive(t *testing.T) {
	mock := &eventsMock{pullFailures: 2, messageEveryN: 1}
	client := newEventsClient(t, mock)

	var messages int
	var mu sync.Mutex

	sub, err := client.Events().SubscribeEvents(context.Background(), func(NotificationMessage) {
		mu.Lock()
		defer mu.Unlock()
		messages++
	}, fastEventOpts())
	if err != nil {
		t.Fatalf("SubscribeEvents() error = %v", err)
	}
	defer sub.stopLoop()

	// Two 500s then successes: the loop must survive and deliver.
	waitFor(t, 10*time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()

		return messages >= 2
	})
}

func TestSubscribeEventsRenewFailureTerminates(t *testing.T) {
	mock := &eventsMock{renewFault: " Sender not authorized ", messageEveryN: 1}
	client := newEventsClient(t, mock)

	sub, err := client.Events().SubscribeEvents(context.Background(),
		func(NotificationMessage) {}, fastEventOpts())
	if err != nil {
		t.Fatalf("SubscribeEvents() error = %v", err)
	}

	// Renewal failure must end the loop (and attempt cleanup).
	select {
	case <-sub.Done():
	case <-time.After(10 * time.Second):
		t.Fatal("loop did not exit after renewal failure")
	}

	mock.mu.Lock()
	defer mock.mu.Unlock()

	if mock.unsubscribes < 1 {
		t.Error("best-effort unsubscribe was not attempted after renewal failure")
	}
}

func TestSubscribeEventsHandlerPanicIsolated(t *testing.T) {
	mock := &eventsMock{messageEveryN: 1}
	client := newEventsClient(t, mock)

	var deliveries int
	var mu sync.Mutex

	sub, err := client.Events().SubscribeEvents(context.Background(), func(NotificationMessage) {
		mu.Lock()
		deliveries++
		boom := deliveries%2 == 1
		mu.Unlock()

		if boom {
			panic("handler bug")
		}
	}, fastEventOpts())
	if err != nil {
		t.Fatalf("SubscribeEvents() error = %v", err)
	}
	defer sub.stopLoop()

	waitFor(t, 5*time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()

		return deliveries >= 3
	})
}

func TestSubscribeEventsContextCancelEndsLoop(t *testing.T) {
	mock := &eventsMock{}
	client := newEventsClient(t, mock)

	ctx, cancel := context.WithCancel(context.Background())

	sub, err := client.Events().SubscribeEvents(ctx, func(NotificationMessage) {}, fastEventOpts())
	if err != nil {
		t.Fatalf("SubscribeEvents() error = %v", err)
	}

	cancel()

	select {
	case <-sub.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("loop did not exit after context cancellation")
	}
}

func TestSubscribeEventsRequiresHandler(t *testing.T) {
	client, err := NewClient("http://192.168.1.100")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if _, err := client.Events().SubscribeEvents(context.Background(), nil, nil); err == nil {
		t.Fatal("SubscribeEvents(nil handler) accepted, want error")
	}
}
