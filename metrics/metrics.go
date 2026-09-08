// Package metrics defines the library-neutral observability seam for the
// ONVIF stack (issue #66): the host bridges events to Prometheus (or any
// backend) without this module taking a metrics dependency.
//
// Hook methods must be cheap (they fire on every request) and safe for
// concurrent use; never block inside them.
package metrics

// Hooks receives ONVIF server observability events. Use NoopHooks (the
// zero configuration) or embed it to implement only what you need.
type Hooks interface {
	// SoapRequest counts a SOAP request that passed authentication and
	// was dispatched to its handler, with the action's local name
	// (e.g. "GetStreamUri").
	SoapRequest(action string)
	// SoapFault counts a dispatched request that completed with a SOAP
	// Fault response (handler error) — the fault-rate numerator.
	SoapFault(action string)
	// AuthFail counts a UsernameToken that failed verification (bad
	// credentials).
	AuthFail()
	// AuthLockout counts a request refused while its source was locked
	// out after repeated authentication failures.
	AuthLockout()
	// DiscoveryProbeAnswered counts a WS-Discovery Probe answered with
	// ProbeMatches (multicast and directed HTTP probes).
	DiscoveryProbeAnswered()
}

// NoopHooks is the no-op Hooks used when none are configured.
type NoopHooks struct{}

var _ Hooks = NoopHooks{}

func (NoopHooks) SoapRequest(string)      {}
func (NoopHooks) SoapFault(string)        {}
func (NoopHooks) AuthFail()               {}
func (NoopHooks) AuthLockout()            {}
func (NoopHooks) DiscoveryProbeAnswered() {}

// OrNoop normalizes a nil Hooks to NoopHooks so call sites never need a
// nil check.
func OrNoop(h Hooks) Hooks {
	if h == nil {
		return NoopHooks{}
	}
	return h
}
