package metrics_test

// Bridge sketch: a host-side Hooks implementation counting into atomics
// (the shape a PrometheusBridge takes — counter increments or histogram
// observations straight into your registry instead of this snapshot).

import (
	"fmt"
	"sync/atomic"

	"github.com/mickeyzzc/onvif-go/v2/metrics"
	"github.com/mickeyzzc/onvif-go/v2/server"
)

// ExportedBridge counts every hook event; hosts publish the totals to
// their metrics backend of choice.
type ExportedBridge struct {
	soapRequests atomic.Int64
	soapFaults   atomic.Int64
	authFails    atomic.Int64
	lockouts     atomic.Int64
	probes       atomic.Int64
}

// Compile-time check that the bridge satisfies the seam.
var _ metrics.Hooks = (*ExportedBridge)(nil)

func (b *ExportedBridge) SoapRequest(action string) {
	b.soapRequests.Add(1)
}

func (b *ExportedBridge) SoapFault(action string) {
	b.soapFaults.Add(1)
}

func (b *ExportedBridge) AuthFail() {
	b.authFails.Add(1)
}

func (b *ExportedBridge) AuthLockout() {
	b.lockouts.Add(1)
}

func (b *ExportedBridge) DiscoveryProbeAnswered() {
	b.probes.Add(1)
}

func Example() {
	bridge := &ExportedBridge{}

	// Wire the same bridge into the SOAP server via Option …
	_, _ = server.New(&server.Config{}, server.WithMetrics(bridge))

	// … and every dispatched request now bumps bridge.soapRequests.
	fmt.Println("bridge wired")
	// Output: bridge wired
}
