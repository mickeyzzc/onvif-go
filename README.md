# onvif-go

[![Go Reference](https://pkg.go.dev/badge/github.com/mickeyzzc/onvif-go.svg)](https://pkg.go.dev/github.com/mickeyzzc/onvif-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/mickeyzzc/onvif-go)](https://goreportcard.com/report/github.com/mickeyzzc/onvif-go)
[![License](https://img.shields.io/github/license/mickeyzzc/onvif-go)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.26-00ADD8.svg)](https://go.dev/)

> Modern Go library for ONVIF IP camera integration — client + virtual camera server,
> zero third-party dependencies. Maintained by [@mickeyzzc](https://github.com/mickeyzzc).

A production-ready Go library for communicating with ONVIF-compliant IP cameras, NVRs and
surveillance devices. Verified against Hikvision, Axis, Dahua, Bosch, Amcrest, HiSilicon-OEM
hardware and minimal embedded implementations (ESP32). It powers
[MiBeeNvr](https://github.com/Mi-Bee-Studio/MiBeeNvr).

## Features

- **ONVIF client + virtual camera server** — simulate cameras for testing
- **Service-domain facade API** — `client.Media()`, `client.Device()`, `client.PTZ()`, …
- **WS-Discovery** — active multicast probe + directed probing helpers
- **WS-Security** — UsernameToken digest auth with clock-skew correction
- **Clock-skew-aware auth** — Hikvision time-divergence fixes built in
- **Capability XAddr repair** — stale advertised service IPs rewritten automatically
- **Zero runtime dependencies** — the library module requires nothing outside stdlib

## Install

```bash
go get github.com/mickeyzzc/onvif-go
```

Requires Go **1.26+**.

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/mickeyzzc/onvif-go"
)

func main() {
    client, err := onvif.NewClient("http://192.168.1.100/onvif/device_service",
        onvif.WithCredentials("admin", "pass"))
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
    for _, p := range profiles {
        fmt.Printf("profile %s: %dx%d %s\n",
            p.Token, p.VideoEncoderConfiguration.Resolution.Width,
            p.VideoEncoderConfiguration.Resolution.Height,
            p.VideoEncoderConfiguration.Encoding)
    }
}
```

Device discovery:

```go
import "github.com/mickeyzzc/onvif-go/discovery"

devices, err := discovery.Discover(ctx, 5*time.Second)
```

## Project layout

| Path | Purpose |
|---|---|
| `*.go` (root) | Client library — per-service facade files (`media*.go`, `device*.go`, `ptz.go`, …) |
| `discovery/` | WS-Discovery (multicast probe) |
| `internal/soap/` | SOAP transport + WS-Security digest (shared by client and server) |
| `server/` | Virtual ONVIF camera server (simulator for testing) |
| `testing/` | Test helpers: mock server, capture replay, golden files |
| `testdata/captures/` | Real-camera SOAP captures used as regression fixtures |
| `cmd/` | Helper CLIs: `discover`, `onvif-quick`, `onvif-diagnostics`, `onvif-server` |

## Development

```bash
make build    # go build ./...
make test     # go test -race ./...
make lint     # golangci-lint run   (install: make lint-install)
make fmt      # gofumpt + goimports via golangci-lint fmt
```

CI (`.github/workflows/ci.yml`) runs lint + race tests + build on every push to `main`.

## Attribution & license

This project is a maintained continuation of
[0x524a/onvif-go](https://github.com/0x524a/onvif-go) (itself derived from earlier ONVIF
Go efforts — see git history). The module path changed in v1.2.0; before that it was
published as `github.com/0x524a/onvif-go` (tags v1.0.0–v1.1.7 remain on this repo for
history).

Licensed under the [MIT License](LICENSE). Contributions from the original authors are
gratefully acknowledged and remain under the same license.
