# Device Discovery

onvif-go ships three complementary discovery modes plus a post-processing
layer. All live in the self-contained `discovery` package.

| Mode | Function | Reaches |
|---|---|---|
| Active probe | `Discover` / `DiscoverWithOptions` | Same subnet (UDP multicast) |
| Passive listener | `Listener` | Same subnet, zero latency |
| Directed probe | `ProbeEndpoint` / `ProbeSerial` | Any address, cross-subnet, pure HTTP |

## Active: multicast probe

```go
devices, err := discovery.Discover(ctx, 5*time.Second)
// or pin the multicast interface (name or IP) on multi-NIC hosts:
devices, err := discovery.DiscoverWithOptions(ctx, 5*time.Second,
    &discovery.DiscoverOptions{NetworkInterface: "eth0"})
```

Polling this once a minute is the usual pattern — which wastes the Hello a
camera broadcasts at power-on unless something is listening. That is what
the passive listener is for.

## Passive: the resident listener

```go
listener, _ := discovery.NewListener("", func(d *discovery.Device) {
    fmt.Println("camera online:", d.GetDeviceEndpoint())
})
go func() { _ = listener.Start(ctx) }()
defer listener.Stop()
```

- Hears **Hello** (power-on self-announcement) and **ProbeMatches** (other
  devices answering someone else's probe — a free second discovery source);
  Bye and garbage are ignored.
- **Coexists with active discovery**: it joins the WS-Discovery multicast
  group exactly like `DiscoverWithOptions` does (SO_REUSEADDR), and on
  Linux/macOS the kernel delivers a copy of each datagram to every bound
  socket — a listener plus on-demand probes in one process do not steal each
  other's traffic.
- `ifaceName` `""` = kernel default interface (recommended on single-NIC
  hosts).
- The handler runs on the listener goroutine: return quickly, offload slow
  work. Panics in the handler are contained.
- `Stop()` is idempotent; a stopped listener refuses restart. `Done()`
  closes when the loop has fully exited.

## Directed: pure-HTTP probing

Multicast does not route across subnets, and some devices never answer
WS-Discovery probes. The camera-rediscovery pattern (a camera changed IP and
must be found again by serial) needs pure HTTP:

```go
// Two strategies, in order: a WS-Discovery Probe POSTed to
// http://host:port/onvif/device_service, then an unauthenticated
// GetDeviceInformation. nil = not ONVIF / offline (the two are
// indistinguishable from outside; both mean "not found").
dev := discovery.ProbeEndpoint(ctx, "192.168.2.50", 80, 1200*time.Millisecond)

// Serial number via common-port scan (80, 8080, 8000 by default). The
// serial is extracted namespace-agnostically — it is the stable identity
// anchor for correlating one physical camera across protocols.
serial, ok := discovery.ProbeSerial(ctx, "192.168.2.50", nil)
```

Malformed responses from arbitrary hosts are contained (never a panic);
401/405-class answers count as "not ONVIF for probing purposes".

## Post-processing

Three helpers turn raw discovery results into usable device lists:

```go
usable := discovery.FilterONVIFDevices(devices) // drop ghost responders
info := discovery.ParseScopes(dev.Scopes)       // name / hardware / location
discovery.EnrichDevices(ctx, usable)            // parallel identity fetch
```

- **`FilterONVIFDevices`** — generic WS-Discovery responders (Synology DSM,
  Windows hosts, printers) answer any Probe regardless of its Types filter
  and otherwise become forever-pending ghosts in camera inventories. A
  device is kept when its Types contain `NetworkVideoTransmitter` (local
  part, prefix-agnostic) **or** any scope starts with `onvif://www.onvif.org/`
  — deliberately lenient so marginal implementations survive.
- **`ParseScopes`** — the ONVIF scope conventions (`onvif://www.onvif.org/
  name/X`, `.../hardware/Y`, `.../location/Z`), percent-decoded. The
  structured fields are also filled directly on every `Device` the package
  produces (`d.Name`, `d.Hardware`, `d.Location`).
- **`EnrichDevices`** — fetches unauthenticated `GetDeviceInformation` for
  every device in parallel (bounded concurrency via `WithEnrichConcurrency`,
  per-device timeout via `WithEnrichTimeout`), filling `d.Info`
  (manufacturer / model / firmware / serial). Best-effort: unreachable
  devices are skipped, never fatal; devices already carrying `Info` are
  untouched.

## A note on identity

`Device.EndpointRef` is the WS-Discovery endpoint address (a `urn:uuid:...`
form) — **not** the device serial number. Correlating a camera across
protocols (ONVIF vs GB28181) must use `Device.Info.SerialNumber`; comparing
`EndpointRef` against a serial silently never matches.
