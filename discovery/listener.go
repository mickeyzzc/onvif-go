package discovery

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// ErrListenerStopped is returned by Start after the listener was stopped —
// a stopped listener cannot be restarted; build a new one.
var ErrListenerStopped = errors.New("discovery listener already stopped")

// listenerReadTick is the read-deadline granularity: the loop re-arms its
// deadline every tick so cancellation (Stop / ctx) is noticed promptly even
// while blocked in ReadFromUDP. A blocked read would otherwise hold the
// close sequence hostage.
const listenerReadTick = 5 * time.Second

// listenerBufferSize absorbs power-on bursts when a camera rack boots at
// once; Hello messages are small but arrive in volleys.
const listenerBufferSize = 256 * 1024

// Listener is a passive WS-Discovery listener: a resident UDP :3702 socket
// that receives the Hello messages cameras broadcast the moment they power
// on — the "plug in and it appears instantly" experience — plus the
// ProbeMatches other devices send in answer to someone else's probe (a free
// second discovery source). Bye messages and anything unparseable are
// ignored.
//
// Coexistence: the listener binds the multicast group exactly like
// DiscoverWithOptions does; on Linux and macOS the kernel delivers a copy of
// each multicast datagram to every bound socket, so a passive listener and
// on-demand Discover() calls coexist in one process without stealing each
// other's traffic.
type Listener struct {
	ifaceName string
	handler   func(*Device)

	readTick time.Duration

	stopOnce sync.Once
	stopped  chan struct{}
	done     chan struct{}
}

// NewListener creates a passive discovery listener. ifaceName selects the
// multicast interface ("" = kernel default, recommended on single-NIC hosts;
// see DiscoverOptions.NetworkInterface for accepted forms). The handler runs
// on the listener goroutine: return quickly and run slow work on your own
// goroutine — a panicking handler is contained and does not stop the
// listener.
func NewListener(ifaceName string, handler func(*Device)) (*Listener, error) {
	if handler == nil {
		return nil, fmt.Errorf("discovery listener: handler is nil")
	}

	return &Listener{
		ifaceName: ifaceName,
		handler:   handler,
		readTick:  listenerReadTick,
		stopped:   make(chan struct{}),
		done:      make(chan struct{}),
	}, nil
}

// Start blocks listening for Hello/ProbeMatches until Stop is called or ctx
// is cancelled. Call it on its own goroutine. Returns ErrListenerStopped when
// called after Stop.
func (l *Listener) Start(ctx context.Context) error {
	select {
	case <-l.stopped:
		return ErrListenerStopped
	default:
	}

	var iface *net.Interface
	if l.ifaceName != "" {
		var err error
		iface, err = resolveNetworkInterface(l.ifaceName)
		if err != nil {
			return fmt.Errorf("discovery listener: %w", err)
		}
	}

	group, err := net.ResolveUDPAddr("udp", multicastAddr)
	if err != nil {
		return fmt.Errorf("discovery listener: failed to resolve multicast address: %w", err)
	}

	conn, err := net.ListenMulticastUDP("udp", iface, group)
	if err != nil {
		return fmt.Errorf("discovery listener: failed to listen on multicast address: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	return l.readLoop(ctx, conn)
}

// readLoop listens on conn until stopped/cancelled; it owns the Done signal
// so both Start and direct-loop consumers (tests) observe loop exit.
func (l *Listener) readLoop(ctx context.Context, conn *net.UDPConn) error {
	defer close(l.done)

	buffer := make([]byte, listenerBufferSize)

	for {
		if err := conn.SetReadDeadline(time.Now().Add(l.readTick)); err != nil {
			return fmt.Errorf("discovery listener: failed to set read deadline: %w", err)
		}

		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				// Deadline tick: check for cancellation, keep listening.
				select {
				case <-l.stopped:
					return nil
				case <-ctx.Done():
					return nil
				default:
					continue
				}
			}

			// A closed connection (Stop on another path) surfaces as a
			// non-timeout error; treat every error as terminal but check
			// whether it was an intentional stop for the return value.
			select {
			case <-l.stopped:
				return nil
			default:
				return fmt.Errorf("discovery listener: read failed: %w", err)
			}
		}

		if device := parseDiscoveryDatagram(buffer[:n]); device != nil {
			l.dispatch(device)
		}

		select {
		case <-l.stopped:
			return nil
		case <-ctx.Done():
			return nil
		default:
		}
	}
}

// dispatch hands a device to the handler with panic isolation.
func (l *Listener) dispatch(device *Device) {
	defer func() {
		_ = recover() // a panicking handler must not stop the listener
	}()

	l.handler(device)
}

// Stop stops the listener and unblocks Start. Idempotent; a stopped listener
// cannot be restarted (create a new one).
func (l *Listener) Stop() {
	l.stopOnce.Do(func() {
		close(l.stopped)
	})
}

// Done returns a channel closed when the read loop has fully exited.
func (l *Listener) Done() <-chan struct{} {
	return l.done
}

// discoveryEnvelope is the shape of both Hello and ProbeMatches messages.
type discoveryEnvelope struct {
	Body struct {
		Hello        *discoveryMatch `xml:"Hello"`
		ProbeMatches struct {
			ProbeMatch []discoveryMatch `xml:"ProbeMatch"`
		} `xml:"ProbeMatches"`
	} `xml:"Body"`
}

// discoveryMatch carries the announcement fields shared by Hello and
// ProbeMatch; there is deliberately no XMLName constraint — the envelope
// parser dispatches on the containing element.
type discoveryMatch struct {
	EndpointRef     string `xml:"EndpointReference>Address"`
	Types           string `xml:"Types"`
	Scopes          string `xml:"Scopes"`
	XAddrs          string `xml:"XAddrs"`
	MetadataVersion int    `xml:"MetadataVersion"`
}

// parseDiscoveryDatagram extracts a device from a Hello or ProbeMatches
// datagram. Returns nil for Bye messages, responses with no match, and
// anything unparseable — a listener must survive garbage on the wire.
func parseDiscoveryDatagram(data []byte) *Device {
	var envelope discoveryEnvelope
	if err := xml.Unmarshal(data, &envelope); err != nil {
		return nil
	}

	var match *discoveryMatch
	switch {
	case envelope.Body.Hello != nil && envelope.Body.Hello.EndpointRef != "":
		match = envelope.Body.Hello
	case len(envelope.Body.ProbeMatches.ProbeMatch) > 0:
		first := envelope.Body.ProbeMatches.ProbeMatch[0]
		match = &first
	default:
		// Bye, Probe acks, or empty body: not a device announcement.
		return nil
	}

	device := &Device{
		EndpointRef:     match.EndpointRef,
		XAddrs:          parseSpaceSeparated(match.XAddrs),
		Types:           parseSpaceSeparated(match.Types),
		Scopes:          parseSpaceSeparated(match.Scopes),
		MetadataVersion: match.MetadataVersion,
	}
	device.Name, device.Hardware, device.Location = scopeFields(device.Scopes)

	return device
}
