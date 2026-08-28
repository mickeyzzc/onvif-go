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

// multicastGroupHost is the host part of the WS-Discovery group
// address, used as the derivation peer for Hello/Bye announcements.
var multicastGroupHost = func() string {
	host, _, err := net.SplitHostPort(wsdiscovery.MulticastAddr)
	if err != nil {
		return wsdiscovery.MulticastAddr
	}

	return host
}()

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
	// http://<device's own address toward the peer>:<Port><DevicePath>
	// (the interface a reply to that peer leaves from). NOTE: the
	// device's own address, never the requester's — echoing the
	// requester would make NVR-style consumers register themselves as
	// the camera (#38). Multi-homed hosts with a stable service address
	// should set XAddrs explicitly.
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
	conn     *net.UDPConn
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
// down (which sends Bye). Configuration errors (unknown interface)
// surface here synchronously; late bind failures terminate the loop and
// close Done.
func (r *Responder) Start(ctx context.Context) error {
	r.mu.Lock()
	if r.started {
		r.mu.Unlock()

		return errors.New("discovery responder already started")
	}
	r.started = true
	r.mu.Unlock()

	// Validate the interface synchronously so misconfiguration is
	// reported by Start, not swallowed by the loop goroutine.
	if r.config.Interface != "" {
		if _, err := resolveInterface(r.config.Interface); err != nil {
			return fmt.Errorf("discovery responder: %w", err)
		}
	}

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

	// Expose the socket so Stop can unblock the read loop immediately
	// (closing the conn interrupts ReadFromUDP; without it, Stop would
	// only be noticed on the next read-deadline tick).
	r.mu.Lock()
	r.conn = conn
	r.mu.Unlock()

	// Cover the Stop-before-bind race: if Stop already ran, it found no
	// socket to close — close it here ourselves and bail out.
	select {
	case <-r.stopped:
		_ = conn.Close()

		return nil
	default:
	}

	// Announce ourselves. Deriving the XAddr host toward the multicast
	// group picks the interface the announcement leaves from.
	if hello := wsdiscovery.BuildHello(r.matchFor(ctx, multicastGroupHost)); len(hello) > 0 {
		if _, err := conn.WriteToUDP(hello, group); err != nil {
			_ = conn.Close()

			return fmt.Errorf("discovery responder: send Hello: %w", err)
		}
	}

	defer func() {
		// Best-effort Bye on the way out.
		_, _ = conn.WriteToUDP(wsdiscovery.BuildBye(r.matchFor(ctx, multicastGroupHost)), group)
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

		r.handleDatagram(ctx, buffer[:n], src, func(data []byte, to *net.UDPAddr) {
			_, _ = conn.WriteToUDP(data, to)
		})
	}
}

// handleDatagram answers one probe datagram from src; send delivers the
// reply (unicast). Non-probe traffic is ignored.
func (r *Responder) handleDatagram(ctx context.Context, data []byte, src *net.UDPAddr, send func([]byte, *net.UDPAddr)) {
	probe := wsdiscovery.ParseProbe(data)
	if probe == nil || probe.MessageID == "" {
		return
	}

	if !probe.MatchesTypes(r.config.Types) {
		return
	}

	answer := wsdiscovery.BuildProbeMatches(probe.MessageID, r.matchFor(ctx, hostOnly(src)))
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

	answer := wsdiscovery.BuildProbeMatches(probe.MessageID, r.matchFor(req.Context(), remote))

	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(answer)
}

// Stop shuts the responder down (sending Bye) and unblocks the loop
// immediately by closing the multicast socket. Idempotent.
func (r *Responder) Stop() {
	r.stopOnce.Do(func() {
		close(r.stopped)

		r.mu.Lock()
		conn := r.conn
		r.mu.Unlock()

		if conn != nil {
			_ = conn.Close()
		}
	})
}

// Done returns a channel closed when the multicast loop has exited.
func (r *Responder) Done() <-chan struct{} {
	return r.done
}

// matchFor builds the advertised Match. When XAddrs is unset, the XAddr
// host is the device's own source address toward the peer — the address
// a reply to that peer is sent from — never the peer's address. Echoing
// the requester would make NVR-style consumers register *themselves* as
// the camera (#38).
func (r *Responder) matchFor(ctx context.Context, peer string) wsdiscovery.Match {
	xaddrs := r.config.XAddrs
	if len(xaddrs) == 0 {
		hostPort := net.JoinHostPort(r.derivedHost(ctx, peer), strconv.Itoa(r.config.Port))
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

// derivedHost returns the device address to advertise toward peer: the
// local source address a reply to peer would use (falling back to the
// configured interface's address, then loopback). Consumers registering
// the device from ProbeMatches XAddrs always end up pointed at this
// device, not at themselves.
func (r *Responder) derivedHost(ctx context.Context, peer string) string {
	if peer != "" {
		if host := localAddrToward(ctx, peer); host != "" {
			return host
		}
	}

	if r.config.Interface != "" {
		if iface, err := resolveInterface(r.config.Interface); err == nil {
			if addrs, addrErr := iface.Addrs(); addrErr == nil {
				for _, addr := range addrs {
					if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
						return ipnet.IP.String()
					}
				}
			}
		}
	}

	return "127.0.0.1"
}

// localAddrToward dials a throwaway UDP socket toward peer (no packets
// are sent) and returns the local source address the kernel chose — the
// interface address traffic to that peer leaves from.
func localAddrToward(ctx context.Context, peer string) string {
	dialCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var dialer net.Dialer

	conn, err := dialer.DialContext(dialCtx, "udp", net.JoinHostPort(peer, "9"))
	if err != nil {
		return ""
	}
	defer func() { _ = conn.Close() }()

	host, _, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		return ""
	}

	return host
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
