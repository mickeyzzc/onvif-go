package discovery

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ScopeInfo is the structured form of the ONVIF scope conventions — the
// name/hardware/location hints devices bury in a space-separated scope list
// (onvif://www.onvif.org/name/XXX, .../hardware/YYY, .../location/ZZZ).
type ScopeInfo struct {
	Name     string
	Hardware string
	Location string
}

// onvifScopePrefix is the scope namespace defined by the ONVIF discovery
// specification; the path segments after it carry the semantics.
const onvifScopePrefix = "onvif://www.onvif.org/"

// ParseScopes extracts the well-known ONVIF scope entries. Values are
// percent-decoded when possible; malformed scopes are skipped. Later entries
// win over earlier ones for a repeated key (devices occasionally advertise
// twice, the refined value last).
func ParseScopes(scopes []string) ScopeInfo {
	var info ScopeInfo

	for _, scope := range scopes {
		if !strings.HasPrefix(scope, onvifScopePrefix) {
			continue
		}

		rest := scope[len(onvifScopePrefix):]
		key, value, found := strings.Cut(rest, "/")
		if !found {
			continue
		}

		if decoded, err := url.PathUnescape(value); err == nil {
			value = decoded
		}

		switch key {
		case "name":
			info.Name = value
		case "hardware":
			info.Hardware = value
		case "location":
			info.Location = value
		}
	}

	return info
}

// IsONVIFResponder reports whether a discovery response comes from an actual
// ONVIF device. Generic WS-Discovery responders answer any Probe regardless
// of its Types filter: Synology DSM (:5357/:5000), Windows hosts and printers
// routinely sneak into camera lists and become forever-pending ghosts in
// downstream inventories. The filter is deliberately lenient (OR of two
// signals) so marginal implementations are not culled:
//
//   - Types contains NetworkVideoTransmitter (compared by local part, so
//     dp0:/tns:-style prefixes do not matter), or
//   - any scope falls under onvif://www.onvif.org/.
func IsONVIFResponder(d *Device) bool {
	if d == nil {
		return false
	}

	for _, typ := range d.Types {
		if _, local, found := strings.Cut(typ, ":"); found {
			if local == "NetworkVideoTransmitter" {
				return true
			}
		} else if typ == "NetworkVideoTransmitter" {
			return true
		}
	}

	for _, scope := range d.Scopes {
		if strings.HasPrefix(scope, onvifScopePrefix) {
			return true
		}
	}

	return false
}

// FilterONVIFDevices returns the devices that are actual ONVIF responders
// (see IsONVIFResponder), preserving order. The returned slice reuses the
// input backing array; treat both as read-only.
func FilterONVIFDevices(devices []*Device) []*Device {
	filtered := devices[:0]

	for _, d := range devices {
		if IsONVIFResponder(d) {
			filtered = append(filtered, d)
		}
	}

	return filtered
}

// enrichConfig tunes EnrichDevices.
type enrichConfig struct {
	concurrency int
	timeout     time.Duration
}

// DefaultEnrichConcurrency spreads the unauthenticated GetDeviceInformation
// round so a large discovery result does not stampede one weak device or the
// network.
const DefaultEnrichConcurrency = 8

// EnrichOption configures EnrichDevices.
type EnrichOption func(*enrichConfig)

// WithEnrichConcurrency caps the number of parallel per-device requests.
func WithEnrichConcurrency(n int) EnrichOption {
	return func(c *enrichConfig) {
		if n > 0 {
			c.concurrency = n
		}
	}
}

// WithEnrichTimeout sets the per-device request timeout.
func WithEnrichTimeout(d time.Duration) EnrichOption {
	return func(c *enrichConfig) {
		if d > 0 {
			c.timeout = d
		}
	}
}

// EnrichDevices fetches identity information (manufacturer, model, firmware,
// serial number) for every device in parallel via unauthenticated
// GetDeviceInformation, filling Device.Info and the structured Name/Hardware
// scope fields. Best-effort by design: devices without XAddrs, unreachable
// devices, and garbage responses are silently skipped — enrichment never
// fails the discovery result it decorates. Devices that already carry Info
// are left untouched.
func EnrichDevices(ctx context.Context, devices []*Device, opts ...EnrichOption) {
	cfg := enrichConfig{
		concurrency: DefaultEnrichConcurrency,
		timeout:     defaultProbeTimeout,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	if len(devices) == 0 || cfg.concurrency <= 0 {
		return
	}

	client := probeHTTPClient(cfg.timeout)

	sem := make(chan struct{}, cfg.concurrency)
	var wg sync.WaitGroup

	for _, device := range devices {
		if err := ctx.Err(); err != nil {
			break
		}

		if device == nil || device.Info != nil || len(device.XAddrs) == 0 {
			continue
		}

		wg.Add(1)
		go func(device *Device) {
			defer wg.Done()

			// A stampede of goroutines must never panic the caller.
			defer func() {
				_ = recover()
			}()

			sem <- struct{}{}
			defer func() { <-sem }()

			info, err := fetchDeviceInfo(ctx, client, device.XAddrs[0])
			if err != nil || (info.Manufacturer == "" && info.SerialNumber == "" && info.Model == "") {
				return
			}

			device.Info = info
		}(device)
	}

	wg.Wait()
}
