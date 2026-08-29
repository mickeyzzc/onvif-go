# onvif-go

[![CI](https://github.com/mickeyzzc/onvif-go/actions/workflows/ci.yml/badge.svg)](https://github.com/mickeyzzc/onvif-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/mickeyzzc/onvif-go/v2.svg)](https://pkg.go.dev/github.com/mickeyzzc/onvif-go/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/mickeyzzc/onvif-go/v2)](https://goreportcard.com/report/github.com/mickeyzzc/onvif-go/v2)
[![License](https://img.shields.io/github/license/mickeyzzc/onvif-go)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.26-00ADD8.svg)](https://go.dev/)

> ONVIF client + virtual camera server for Go — service-facade API, active and
> passive discovery, an authentication ladder that survives weird firmware,
> PTZ / media / events / imaging. Zero third-party dependencies.
> Maintained by [@mickeyzzc](https://github.com/mickeyzzc).

> [**中文**](README.zh.md) — [English](README.md)

A production-hardened Go library for communicating with ONVIF-compliant IP
cameras, NVRs and surveillance devices. Its device compatibility comes from
real deployments: Hikvision, Axis, Dahua, Bosch, Amcrest, HiSilicon-OEM
hardware, minimal embedded implementations (ESP32), and the odd firmware that
answers everything with a SOAP Fault. It powers
[MiBeeNvr](https://github.com/Mi-Bee-Studio/MiBeeNvr), an NVR running
24/7 on ARM64.

## Features

**Client** — every ONVIF operation lives on the service object that owns it,
mirroring the ONVIF service model:

- `client.Device()` — info, capabilities (cached, single-flighted,
  degradable to a minimal set for weak devices), network/system/DNS/NTP
  configuration, storage, WiFi, certificates, users
- `client.Media()` — profiles with main/sub-stream selection helpers, stream
  URIs (`GetStreamURIWithOptions`: RTSP/HTTP/UDP × unicast/multicast,
  namespace-tolerant response parsing), snapshots, encoder/audio/OSD
  configuration
- `client.PTZ()` — moves, status, presets
- `client.Imaging()` — exposure, focus, imaging settings
- `client.Events()` — managed pull-point subscriptions (background polling,
  auto-renewal, `ErrEventsNotSupported` sentinel) plus the raw primitives
- `client.DeviceIO()`, `client.Security()` — relays, I/O, user management

**Authentication built for real firmware** — `WithAuthMode` selects
digest / password-text / HTTP Basic / none; `WithAuthFallback` adds an
automatic fallback ladder that remembers the first mode the device accepts.
`errors.Is(err, onvif.ErrUnauthorized)` reliably classifies auth failures
(HTTP 401/403, NotAuthorized faults, 200-with-fault). `WithAutoClockSkew`
measures and corrects device clock divergence (the Hikvision "sender not
authorized" trap), and `DiagnoseAuth` tells clock skew, wrong credentials and
auth-less devices apart.

**Discovery, three ways** — `discovery.Discover` (multicast probe),
`discovery.Listener` (passive: hears camera power-on Hello messages and other
probes' answers, coexisting with active discovery), and
`discovery.ProbeEndpoint` / `ProbeSerial` (pure-HTTP directed probing across
subnets for devices multicast can't reach). Post-processing:
`FilterONVIFDevices` (drops Synology/Windows/printer ghost responders),
`ParseScopes`, and parallel `EnrichDevices` (identity fetch with bounded
concurrency).

**Robustness** — fault detection on every call (a 200-with-Fault is never
"success"), structured `FaultError`/`HTTPStatusError`, empty-URI responses
become explicit errors with body summaries, capability XAddr repair (stale
advertised IPs after camera roaming), and a documented, tested concurrency
contract: one `Client`, many goroutines, no external locking.

**Virtual camera server** — `server/` simulates ONVIF cameras for testing
your recorder without hardware.

## Install

```bash
go get github.com/mickeyzzc/onvif-go/v2
```

Requires Go **1.26+**. The library module has zero third-party dependencies.

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/mickeyzzc/onvif-go/v2/onvif"
)

func main() {
    client, err := onvif.NewClient("http://192.168.1.100/onvif/device_service",
        onvif.WithCredentials("admin", "pass"),
        onvif.WithAutoClockSkew()) // survive clock-diverged cameras
    if err != nil {
        log.Fatal(err)
    }
    if err := client.Initialize(context.Background()); err != nil {
        log.Fatal(err)
    }

    info, err := client.Device().GetDeviceInformation(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s %s (fw %s, serial %s)\n",
        info.Manufacturer, info.Model, info.FirmwareVersion, info.SerialNumber)

    profiles, err := client.Media().GetProfiles(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    mainToken := onvif.SelectMainProfile(profiles) // highest resolution, not profiles[0]
    uri, err := client.Media().GetStreamURI(context.Background(), mainToken)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("main stream:", uri.URI)
}
```

Firmware rejects digest auth? Add a ladder:

```go
client, err := onvif.NewClient(endpoint,
    onvif.WithCredentials("admin", "pass"),
    onvif.WithAuthFallback(onvif.AuthPasswordText, onvif.AuthHTTPBasic, onvif.AuthNone),
)
```

Discovery — active, passive, or directed:

```go
devices, _ := discovery.Discover(ctx, 5*time.Second)          // multicast probe
usable := discovery.FilterONVIFDevices(devices)               // drop ghost responders
discovery.EnrichDevices(ctx, usable)                          // parallel identity fetch

listener, _ := discovery.NewListener("", func(d *discovery.Device) { // passive
    fmt.Println("camera came online:", d.GetDeviceEndpoint())
})
go listener.Start(ctx)
defer listener.Stop()

dev := discovery.ProbeEndpoint(ctx, "192.168.2.50", 80, 1200*time.Millisecond) // cross-subnet
serial, ok := discovery.ProbeSerial(ctx, "192.168.2.50", nil)                  // identity anchor
```

Events — managed, auto-renewing:

```go
sub, err := client.Events().SubscribeEvents(ctx, func(msg onvif.NotificationMessage) {
    fmt.Println(msg.Topic, msg.Message.Data)
}, nil)
if errors.Is(err, onvif.ErrEventsNotSupported) {
    // device has no events service; cache the negative result
}
defer sub.Unsubscribe(context.Background())
```

Auth triage when nothing works:

```go
diag, _ := client.DiagnoseAuth(ctx)
// diag.Status: "ok" | "clock-skew" (fix camera NTP) | "bad-credentials"
```

## Project layout

| Path | Purpose |
|---|---|
| `client.go` etc. (root) | `Client` (auth ladder, clock skew, download) + v1 compatibility aliases |
| `device/` `security/` `deviceio/` `media/` `ptz/` `imaging/` `events/` | domain service packages (v2 split, issue #20) |
| `types/` | shared data-model leaf |
| `discovery/` | WS-Discovery: active probe, passive listener, directed HTTP probing, post-processing |
| `internal/soap/` | SOAP transport + WS-Security (digest/text modes, fault detection) |
| `server/` | Virtual ONVIF camera server (simulator for testing) |
| `docs/{en,zh}/` | Topic documentation (architecture, auth, discovery, media, events, concurrency, testing, CLI) |
| `testdata/captures/` | Real-camera SOAP captures used as regression fixtures |
| `cmd/` | Helper CLIs: `discover`, `onvif-quick`, `onvif-diagnostics`, `onvif-server` |
| `examples/` | Runnable examples per feature area: discovery, device-info, imaging-settings, ptz-control, events (PullPoint subscription), simple-server, onvif-server, complete-demo |

## Documentation

Topic guides, English and Chinese:

| Topic | English | 中文 |
|---|---|---|
| Architecture | [architecture.md](docs/en/architecture.md) | [架构](docs/zh/architecture.md) |
| Authentication & security | [authentication.md](docs/en/authentication.md) | [鉴权与安全](docs/zh/authentication.md) |
| Device discovery | [discovery.md](docs/en/discovery.md) | [设备发现](docs/zh/discovery.md) |
| Media & streaming | [media.md](docs/en/media.md) | [媒体与码流](docs/zh/media.md) |
| Events | [events.md](docs/en/events.md) | [事件](docs/zh/events.md) |
| ONVIF server | [server.md](docs/en/server.md) | [服务端](docs/zh/server.md) |
| v1 → v2 migration | [MIGRATION.md](MIGRATION.md) | [迁移指南](MIGRATION.md) |
| Concurrency model | [concurrency.md](docs/en/concurrency.md) | [并发模型](docs/zh/concurrency.md) |
| Testing | [testing.md](docs/en/testing.md) | [测试](docs/zh/testing.md) |
| CLI tools | [cli.md](docs/en/cli.md) | [CLI 工具](docs/zh/cli.md) |

API reference: [pkg.go.dev/github.com/mickeyzzc/onvif-go/v2/onvif](https://pkg.go.dev/github.com/mickeyzzc/onvif-go/v2/onvif).

## Development

```bash
make build    # go build ./...
make test     # go test -race ./...
make lint     # golangci-lint run   (install: make lint-install)
make fmt      # gofumpt + goimports via golangci-lint fmt
```

CI ([ci.yml](.github/workflows/ci.yml)) runs lint + format check + race
tests + build on every push to `main`; the branch requires all three jobs
green.

## Attribution & license

This project is a maintained continuation of
[0x524a/onvif-go](https://github.com/0x524a/onvif-go) (itself derived from
earlier ONVIF Go efforts — see git history). The module path changed to
`github.com/mickeyzzc/onvif-go` in **v1.2.0**; before that it was published
as `github.com/0x524a/onvif-go` (tags v1.0.0–v1.1.7 remain on this repo for
history and are never deleted). Continuation-focused changes since v1.2.0
are tracked in the [changelog](CHANGELOG.md).

Licensed under the [MIT License](LICENSE). Contributions from the original
authors are gratefully acknowledged and remain under the same license.
