// Package discovery provides the device side of WS-Discovery: a
// resident multicast responder that answers Probe messages with
// ProbeMatches, announces itself with Hello on start and Bye on stop,
// and answers directed Probe-over-HTTP POSTs through an http.Handler —
// the same transport the client's ProbeEndpoint uses.
package discovery

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/wsdiscovery"
)

// readTick is the read-deadline granularity: the loop re-arms its
// deadline every tick so cancellation is noticed promptly while blocked
// in ReadFromUDP.
const readTick = 5 * time.Second

// listenerBufferSize absorbs probe bursts; probes are small but NVRs
// scan in volleys.
const listenerBufferSize = 256 * 1024

// DefaultMetadataVersion is the WS-Discovery metadata revision devices
// advertise; bump it whenever the advertised Types/Scopes/XAddrs change
// at runtime.
const DefaultMetadataVersion = 1

// Config configures a Responder.
type Config struct {
	// EndpointRef is the stable device identifier (an urn:uuid: URN).
	// Empty → a generated random UUID.
	EndpointRef string

	// Types advertised in ProbeMatches/Hello (ONVIF devices answer with
	// at least tds:Device and the NetworkVideoTransmitter type).
	// Empty → the conventional ONVIF set.
	Types []string

	// Scopes advertised (onvif://… URIs: name, location, hardware…).
	Scopes []string

	// XAddrs advertised verbatim. Empty → derived per request as
	// http://<requester IP>:<Port><DevicePath>, so every peer receives
	// an address reachable from its own network.
	XAddrs []string

	// Port and DevicePath build the derived XAddrs (when XAddrs is
	// empty). Defaults: 80 and /onvif/device_service.
	Port       int
	DevicePath string

	// MetadataVersion advertised in answers (DefaultMetadataVersion).
	MetadataVersion int

	// Interface selects the multicast interface ("" = kernel default;
	// accepts names or IPs, like discovery.DiscoverOptions).
	Interface string
}

// Responder is the device-side WS-Discovery responder. Start runs the
// multicast loop on its own goroutine; ServeHTTP answers directed
// HTTP probes and can be mounted next to the ONVIF service endpoints.
//
// Coexistence: the responder binds the multicast group exactly like
// the client's Discover and the passive discovery.Listener do; on
// Linux and macOS the kernel delivers a copy of each multicast
// datagram to every bound socket, so device side and client side
// coexist in one process without stealing each other's traffic.
// Answers are sent unicast to the probing socket, never to the group.
type Responder struct {
	config Config

	mu       sync.Mutex
	started  bool
	stopOnce sync.Once
	stopped  chan struct{}
	done     chan struct{}
}

// NewResponder creates a responder with normalized defaults.
func NewResponder(config Config) *Responder {
	if config.EndpointRef == "" {
		config.EndpointRef = "urn:uuid:" + randomUUID()
	}

	if len(config.Types) == 0 {
		config.Types = []string{"tds:Device", "dp0:NetworkVideoTransmitter"}
	}

	if config.Port == 0 {
		config.Port = 80
	}

	if config.DevicePath == "" {
		config.DevicePath = "/onvif/device_service"
	}

	if config.MetadataVersion == 0 {
		config.MetadataVersion = DefaultMetadataVersion
	}

	return &Responder{
		config:  config,
		stopped: make(chan struct{}),
		done:    make(chan struct{}),
	}
}

// Start begins listening on the WS-Discovery multicast group (on its
// own goroutine) and sends the Hello announcement. Call Stop to shut
// down (which sends Bye).
func (r *Responder) Start(ctx context.Context) error {
	r.mu.Lock()
	if r.started {
		r.mu.Unlock()

		return errors.New("discovery responder already started")
	}
	r.started = true
	r.mu.Unlock()

	go func() {
		defer close(r.done)
		_ = r.run(ctx)
	}()

	return nil
}

// run owns the multicast socket lifecycle; split from Start for testability.
func (r *Responder) run(ctx context.Context) error {
	var iface *net.Interface
	if r.config.Interface != "" {
		var err error
		iface, err = resolveInterface(r.config.Interface)
		if err != nil {
			return fmt.Errorf("discovery responder: %w", err)
		}
	}

	group, err := net.ResolveUDPAddr("udp", wsdiscovery.MulticastAddr)
	if err != nil {
		return fmt.Errorf("discovery responder: resolve multicast address: %w", err)
	}

	conn, err := net.ListenMulticastUDP("udp", iface, group)
	if err != nil {
		return fmt.Errorf("discovery responder: listen multicast: %w", err)
	}

	// Announce ourselves.
	if hello := wsdiscovery.BuildHello(r.matchFor("")); len(hello) > 0 {
		if _, err := conn.WriteToUDP(hello, group); err != nil {
			_ = conn.Close()

			return fmt.Errorf("discovery responder: send Hello: %w", err)
		}
	}

	defer func() {
		// Best-effort Bye on the way out.
		_, _ = conn.WriteToUDP(wsdiscovery.BuildBye(r.matchFor("")), group)
		_ = conn.Close()
	}()

	return r.readLoop(ctx, conn)
}

// readLoop answers probes until stopped or cancelled.
func (r *Responder) readLoop(ctx context.Context, conn *net.UDPConn) error {
	buffer := make([]byte, listenerBufferSize)

	for {
		if err := conn.SetReadDeadline(time.Now().Add(readTick)); err != nil {
			return fmt.Errorf("discovery responder: set read deadline: %w", err)
		}

		n, src, err := conn.ReadFromUDP(buffer)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				select {
				case <-r.stopped:
					return nil
				case <-ctx.Done():
					return nil
				default:
					continue
				}
			}

			select {
			case <-r.stopped:
				return nil
			default:
				return fmt.Errorf("discovery responder: read failed: %w", err)
			}
		}

		r.handleDatagram(buffer[:n], src, func(data []byte, to *net.UDPAddr) {
			_, _ = conn.WriteToUDP(data, to)
		})
	}
}

// handleDatagram answers one probe datagram from src; send delivers the
// reply (unicast). Non-probe traffic is ignored.
func (r *Responder) handleDatagram(data []byte, src *net.UDPAddr, send func([]byte, *net.UDPAddr)) {
	probe := wsdiscovery.ParseProbe(data)
	if probe == nil || probe.MessageID == "" {
		return
	}

	if !probe.MatchesTypes(r.config.Types) {
		return
	}

	answer := wsdiscovery.BuildProbeMatches(probe.MessageID, r.matchFor(hostOnly(src)))
	send(answer, src)
}

// ServeHTTP answers directed WS-Discovery Probes POSTed over HTTP —
// the transport ProbeEndpoint uses for cross-subnet discovery.
func (r *Responder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	body, err := io.ReadAll(io.LimitReader(req.Body, listenerBufferSize))
	if err != nil || len(body) == 0 {
		http.Error(w, "Empty request body", http.StatusBadRequest)

		return
	}

	probe := wsdiscovery.ParseProbe(body)
	if probe == nil || probe.MessageID == "" {
		http.Error(w, "Not a WS-Discovery Probe", http.StatusBadRequest)

		return
	}

	if !probe.MatchesTypes(r.config.Types) {
		w.WriteHeader(http.StatusNoContent)

		return
	}

	remote := req.RemoteAddr
	if host, _, splitErr := net.SplitHostPort(remote); splitErr == nil {
		remote = host
	}

	answer := wsdiscovery.BuildProbeMatches(probe.MessageID, r.matchFor(remote))

	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(answer)
}

// Stop shuts the responder down (sending Bye) and unblocks the loop.
// Idempotent.
func (r *Responder) Stop() {
	r.stopOnce.Do(func() {
		close(r.stopped)
	})
}

// Done returns a channel closed when the multicast loop has exited.
func (r *Responder) Done() <-chan struct{} {
	return r.done
}

// matchFor builds the advertised Match; remoteIP is the requesting
// peer when XAddrs are derived (empty remoteIP → loopback form).
func (r *Responder) matchFor(remoteIP string) wsdiscovery.Match {
	xaddrs := r.config.XAddrs
	if len(xaddrs) == 0 {
		if remoteIP == "" {
			remoteIP = "127.0.0.1"
		}

		hostPort := net.JoinHostPort(remoteIP, strconv.Itoa(r.config.Port))
		xaddrs = []string{"http://" + hostPort + r.config.DevicePath}
	}

	return wsdiscovery.Match{
		EndpointRef:     r.config.EndpointRef,
		Types:           strings.Join(r.config.Types, " "),
		Scopes:          strings.Join(r.config.Scopes, " "),
		XAddrs:          strings.Join(xaddrs, " "),
		MetadataVersion: r.config.MetadataVersion,
	}
}

// randomUUID builds a random v4 UUID string (crypto/rand).
func randomUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Fall back to a time-derived value; uniqueness is best-effort here.
		now := time.Now().UnixNano()

		return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			uint32(now), uint16(now>>32), uint16(now>>48), 0, now)
	}

	b[6] = (b[6] & 0x0f) | 0x40 //nolint:mnd // UUID v4 marker
	b[8] = (b[8] & 0x3f) | 0x80 //nolint:mnd // RFC 4122 variant

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// hostOnly strips the port from a UDP address.
func hostOnly(addr *net.UDPAddr) string {
	if addr == nil {
		return ""
	}

	return addr.IP.String()
}

// resolveInterface resolves an interface by name or IP.
func resolveInterface(spec string) (*net.Interface, error) {
	if iface, err := net.InterfaceByName(spec); err == nil {
		return iface, nil
	}

	if ip := net.ParseIP(spec); ip != nil {
		ifaces, err := net.Interfaces()
		if err != nil {
			return nil, fmt.Errorf("list interfaces: %w", err)
		}

		for _, iface := range ifaces {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range addrs {
				if netIP, ok := addr.(*net.IPNet); ok && netIP.IP.Equal(ip) {
					return &iface, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("interface not found: %q", spec)
}
