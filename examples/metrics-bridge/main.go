// metrics-bridge demonstrates the observability seam (issue #66): one
// Hooks implementation bridges every ONVIF server event into host-side
// counters — swap the print loop for your Prometheus registry and the
// rest of the program stays unchanged.
//
// Run it, then fire requests at it:
//
//	go run ./examples/metrics-bridge
//	curl -s http://127.0.0.1:8080/onvif/device_service \
//	  -H 'Content-Type: application/soap+xml' \
//	  -d '<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope"><Body><GetDeviceInformation xmlns="http://www.onvif.org/ver10/device/wsdl"/></Body></Envelope>'
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/metrics"
	"github.com/mickeyzzc/onvif-go/v2/server"
	discoveryserver "github.com/mickeyzzc/onvif-go/v2/server/discovery"
)

// PrintBridge counts every hook event. A PrometheusBridge has the same
// shape: each method body becomes a counter increment (labelled by the
// action where one is provided) straight into the host's registry.
type PrintBridge struct {
	soapRequests atomic.Int64
	soapFaults   atomic.Int64
	authFails    atomic.Int64
	lockouts     atomic.Int64
	probes       atomic.Int64
}

var _ metrics.Hooks = (*PrintBridge)(nil)

func (b *PrintBridge) SoapRequest(action string) { b.soapRequests.Add(1) }
func (b *PrintBridge) SoapFault(action string)   { b.soapFaults.Add(1) }
func (b *PrintBridge) AuthFail()                 { b.authFails.Add(1) }
func (b *PrintBridge) AuthLockout()              { b.lockouts.Add(1) }
func (b *PrintBridge) DiscoveryProbeAnswered()   { b.probes.Add(1) }

func (b *PrintBridge) snapshot() string {
	return fmt.Sprintf("requests=%d faults=%d authFails=%d lockouts=%d probesAnswered=%d",
		b.soapRequests.Load(), b.soapFaults.Load(), b.authFails.Load(),
		b.lockouts.Load(), b.probes.Load())
}

func main() {
	bridge := &PrintBridge{}

	srv, err := server.New(server.DefaultConfig(), server.WithMetrics(bridge))
	if err != nil {
		log.Fatalf("server: %v", err)
	}

	mux := http.NewServeMux()
	srv.RegisterServices(mux)

	// The WS-Discovery responder takes the same bridge — probes answered
	// on either transport land in the same counters.
	responder := discoveryserver.NewResponder(discoveryserver.Config{
		XAddrs:  []string{"http://127.0.0.1:8080/onvif/device_service"},
		Metrics: bridge,
	})
	mux.Handle("/ws-discovery", responder)
	if err := responder.Start(context.Background()); err != nil {
		log.Printf("discovery responder: %v (multicast unavailable, HTTP probes still served)", err)
	}

	go func() {
		for range time.Tick(5 * time.Second) {
			fmt.Println("onvif metrics:", bridge.snapshot())
		}
	}()

	// The server runs until killed; Stop-on-error keeps the gocritic
	// exitAfterDefer contract (no defer racing a log.Fatal exit).
	log.Println("metrics-bridge listening on :8080 (SOAP at /onvif/device_service, probes at /ws-discovery)")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		responder.Stop()
		log.Fatal(err)
	}
}
