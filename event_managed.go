package onvif

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mickeyzzc/onvif-go/internal/soap"
)

// ErrEventsNotSupported is returned when the device does not implement the
// ONVIF events service (some cameras advertise it in GetCapabilities but
// fault on CreatePullPointSubscription with "Action Not Implemented").
// Without this sentinel, callers would retry an impossible subscription
// forever. Match with errors.Is(err, ErrEventsNotSupported) and cache the
// negative result.
var ErrEventsNotSupported = errors.New("device does not support the ONVIF events service")

// Default managed-subscription tuning.
const (
	defaultEventSubscriptionDuration = time.Hour
	defaultEventRenewMargin          = 5 * time.Minute
	defaultEventPullTimeout          = 30 * time.Second
	defaultEventMessageLimit         = 10
	eventErrorBackoffStart           = time.Second
	eventErrorBackoffMax             = 30 * time.Second
)

// SubscribeEventsOptions tunes the managed event loop. The zero value is
// valid and selects the defaults above.
type SubscribeEventsOptions struct {
	// Filter is a topic expression filter passed to CreatePullPointSubscription.
	Filter string
	// SubscriptionDuration is how long each subscription grant lasts before
	// renewal (default 1h).
	SubscriptionDuration time.Duration
	// RenewMargin is how early before expiry the subscription is renewed
	// (default 5m). Renewal failure terminates the loop.
	RenewMargin time.Duration
	// PullTimeout is the long-poll timeout of each PullMessages call
	// (default 30s).
	PullTimeout time.Duration
	// MessageLimit caps messages per pull (default 10).
	MessageLimit int
}

// EventStream is a managed pull-point subscription: a background
// goroutine long-polls PullMessages, delivers every notification to the
// handler, renews the subscription before it expires, and stops cleanly on
// Unsubscribe or context cancellation.
type EventStream struct {
	client  *Client
	handler func(NotificationMessage)
	opts    SubscribeEventsOptions

	ref      string
	cancel   context.CancelFunc
	done     chan struct{}
	doneOnce sync.Once

	mu           sync.Mutex
	expiresAt    time.Time
	explicitStop bool
}

// SubscribeEvents creates a pull-point subscription and manages its whole
// lifecycle. The handler runs on the polling goroutine: return quickly, run
// slow work on your own goroutine (a panicking handler is isolated and does
// not kill the loop). The subscription lives until Unsubscribe is called or
// ctx is cancelled; a renewal failure also terminates it (Done closes).
//
// Devices that do not implement the events service yield an error matching
// errors.Is(err, ErrEventsNotSupported).
func (s *EventService) SubscribeEvents(
	ctx context.Context,
	handler func(NotificationMessage),
	opts *SubscribeEventsOptions,
) (*EventStream, error) {
	if handler == nil {
		return nil, fmt.Errorf("SubscribeEvents: %w: handler is nil", ErrInvalidParameter)
	}

	effective := SubscribeEventsOptions{
		SubscriptionDuration: defaultEventSubscriptionDuration,
		RenewMargin:          defaultEventRenewMargin,
		PullTimeout:          defaultEventPullTimeout,
		MessageLimit:         defaultEventMessageLimit,
	}
	if opts != nil {
		if opts.Filter != "" {
			effective.Filter = opts.Filter
		}

		if opts.SubscriptionDuration > 0 {
			effective.SubscriptionDuration = opts.SubscriptionDuration
		}

		if opts.RenewMargin > 0 {
			effective.RenewMargin = opts.RenewMargin
		}

		if opts.PullTimeout > 0 {
			effective.PullTimeout = opts.PullTimeout
		}

		if opts.MessageLimit > 0 {
			effective.MessageLimit = opts.MessageLimit
		}
	}

	duration := effective.SubscriptionDuration
	sub, err := s.CreatePullPointSubscription(ctx, effective.Filter, &duration, "")
	if err != nil {
		if classifyEventsNotSupported(err) {
			return nil, fmt.Errorf("SubscribeEvents: %w: %w", ErrEventsNotSupported, err)
		}

		return nil, fmt.Errorf("SubscribeEvents: %w", err)
	}

	loopCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))

	es := &EventStream{
		client:  s.client,
		handler: handler,
		opts:    effective,
		ref:     sub.SubscriptionReference,
		cancel:  cancel,
		done:    make(chan struct{}),
	}

	es.mu.Lock()
	if !sub.TerminationTime.IsZero() {
		es.expiresAt = sub.TerminationTime
	} else {
		es.expiresAt = time.Now().Add(duration)
	}
	es.mu.Unlock()

	// Dying with the caller's context is part of the contract.
	go func() {
		<-ctx.Done()
		es.stopLoop()
	}()

	go es.run(loopCtx)

	return es, nil
}

// Done returns a channel that closes when the managed loop has fully exited
// (Unsubscribe, cancellation, or renewal failure). Useful for deterministic
// cleanup in tests and shutdown paths.
func (es *EventStream) Done() <-chan struct{} {
	return es.done
}

// Unsubscribe stops the managed loop and best-effort unsubscribes on the
// device: the loop is terminated even when the SOAP unsubscribe fails (the
// error is then returned, cleanup is not blocked by it).
func (es *EventStream) Unsubscribe(ctx context.Context) error {
	es.mu.Lock()
	es.explicitStop = true
	es.mu.Unlock()
	es.stopLoop()

	err := es.client.Events().Unsubscribe(ctx, es.ref)
	if err != nil {
		return fmt.Errorf("Unsubscribe: device unsubscribe failed (loop stopped anyway): %w", err)
	}

	return nil
}

// stopLoop cancels the loop exactly once.
func (es *EventStream) stopLoop() {
	es.cancel()
}

// run is the managed polling loop.
func (es *EventStream) run(ctx context.Context) {
	defer es.doneOnce.Do(func() { close(es.done) })

	backoff := eventErrorBackoffStart

	for {
		if ctx.Err() != nil {
			es.cleanupUnlessExplicit(ctx)

			return
		}

		// Renew before the subscription runs out.
		es.mu.Lock()
		renewIn := time.Until(es.expiresAt) - es.opts.RenewMargin
		es.mu.Unlock()

		if renewIn <= 0 && !es.renew(ctx) {
			// Renewal failed: the subscription is gone or the device is
			// refusing; retrying forever would just spam — terminate.
			es.cleanupUnlessExplicit(ctx)

			return
		}

		messages, err := es.client.Events().PullMessages(ctx, es.ref, es.opts.PullTimeout, es.opts.MessageLimit)
		if err != nil {
			if ctx.Err() != nil {
				es.cleanupUnlessExplicit(ctx)

				return
			}

			// Transient failure: back off, keep the loop alive.
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
			}

			backoff = min(2*backoff, eventErrorBackoffMax)

			continue
		}

		backoff = eventErrorBackoffStart

		es.deliver(messages)

		if len(messages) == 0 {
			// Devices that honor long-polling only get here once per
			// PullTimeout; devices that return instantly-empty would
			// otherwise spin the loop at full speed — throttle them.
			select {
			case <-time.After(eventIdleGap):
			case <-ctx.Done():
			}
		}
	}
}

// eventIdleGap throttles the loop against devices that ignore the long-poll
// timeout and answer immediately with no messages.
const eventIdleGap = 100 * time.Millisecond

// renew extends the subscription; reports success.
func (es *EventStream) renew(ctx context.Context) bool {
	_, termination, err := es.client.Events().RenewSubscription(ctx, es.ref, es.opts.SubscriptionDuration)
	if err != nil {
		return false
	}

	es.mu.Lock()
	if termination.IsZero() {
		es.expiresAt = time.Now().Add(es.opts.SubscriptionDuration)
	} else {
		es.expiresAt = termination
	}
	es.mu.Unlock()

	return true
}

// deliver hands messages to the handler with panic isolation and missing
// timestamp defaults.
func (es *EventStream) deliver(messages []NotificationMessage) {
	for i := range messages {
		msg := messages[i]

		if msg.Message.UtcTime.IsZero() {
			msg.Message.UtcTime = time.Now().UTC()
		}

		es.safeHandle(msg)
	}
}

// cleanupUnlessExplicit sends a best-effort SOAP unsubscribe when the loop
// ended on its own (renewal failure, caller-context death). When Unsubscribe()
// triggered the stop it already sent the SOAP request itself — a second one
// would be noise.
func (es *EventStream) cleanupUnlessExplicit(parent context.Context) {
	es.mu.Lock()
	explicit := es.explicitStop
	es.mu.Unlock()

	if explicit {
		return
	}

	// The loop context is already cancelled here; detach cancellation but
	// keep it bounded so cleanup cannot hang.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), es.opts.PullTimeout)
	defer cancel()

	_ = es.client.Events().Unsubscribe(ctx, es.ref)
}

// safeHandle invokes the handler with panic isolation: a panicking handler
// must not kill the polling loop.
func (es *EventStream) safeHandle(msg NotificationMessage) {
	defer func() {
		_ = recover()
	}()

	es.handler(msg)
}

// classifyEventsNotSupported recognizes the "events service not implemented"
// fault family across firmware phrasings.
func classifyEventsNotSupported(err error) bool {
	var faultErr *soap.FaultError
	if !errors.As(err, &faultErr) {
		return false
	}

	normalized := strings.ToLower(strings.NewReplacer(" ", "", "-", "", "_", "").
		Replace(faultErr.Code + "|" + faultErr.Subcode + "|" + faultErr.Reason))

	return strings.Contains(normalized, "notimplemented") ||
		strings.Contains(normalized, "actionnotsupported") ||
		strings.Contains(normalized, "notsupportedbydevice")
}
