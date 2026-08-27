# Architecture

onvif-go is a zero-dependency Go library for talking to ONVIF devices, plus a
virtual camera server for testing recorders without hardware. This document
describes how the pieces fit together.

## Layering

```mermaid
flowchart TD
    subgraph tools["cmd/ · examples/"]
        clis["helper CLIs & runnable examples"]
    end

    subgraph rootpkg["root package — Client"]
        facades["service facades<br/>Device · Media · PTZ · Imaging<br/>Events · DeviceIO · Security"]
        dispatcher["auth dispatcher Client.call<br/>ladder + sticky memory"]
        state["shared state, mutex-guarded<br/>credentials · clock skew ·<br/>capabilities cache · endpoints"]
    end

    subgraph internals["internal/"]
        soap["internal/soap<br/>envelope build/parse · WS-Security ·<br/>fault detection · auth modes"]
        digest["internal/httpdigest<br/>HTTP Digest round-tripper<br/>(snapshot downloads)"]
    end

    subgraph disc["discovery/ — self-contained, imports neither the root package nor internal/"]
        discbox["active probe · passive listener ·<br/>directed HTTP probing · post-processing"]
    end

    subgraph testbed["server/ · testing/ · testdata/captures/"]
        sim["virtual camera · mock server · real-camera fixtures"]
    end

    clis --> facades
    facades --> dispatcher
    dispatcher --> state
    dispatcher --> soap
    dispatcher --> digest
    sim --> soap
```

## File map

The root package is deliberately a single Go package — the facade API
(`client.Media().GetProfiles(ctx)`) only works when the service types and
their methods share one package — so the architecture is expressed in the
file layout instead:

| File | Domain |
|---|---|
| `client.go` | `Client`, functional options, the single call dispatcher, endpoint/XAddr repair |
| `auth.go` | auth strategy (modes, ladder stickiness), clock-skew measurement, `DiagnoseAuth` |
| `capabilities.go` | capabilities cache with single-flight and minimal-device fallback |
| `services.go` | the seven service facade types and their accessors |
| `errors.go` | package sentinels and `ONVIFError` |
| `types.go` | shared ONVIF data types |
| `device.go` | device identity, capabilities, scopes, users, misc `tds` operations |
| `device_mgmt.go` | system date/time, DNS/NTP, network interfaces & gateway |
| `device_security.go` | users/remote user, access policy, certificates |
| `device_storage.go`, `device_wifi.go` | storage and WiFi subsystems |
| `deviceio.go` | relay outputs, digital I/O |
| `media_profiles.go` | profile listing/parsing + main/sub stream selection |
| `media_stream.go` | stream & snapshot URIs, `StreamSetup` transport selection |
| `media_encoder.go`, `media_audio.go`, `media_osd.go` | encoder, audio, and OSD configuration |
| `ptz.go`, `imaging.go` | PTZ and imaging services |
| `event.go` | raw pull-point primitives + the managed subscription loop |
| `download.go` | snapshot/media file download (basic + digest) |

## The service-facade model

The `Client` is a connection and configuration holder; every ONVIF operation
lives on the service object that owns it, mirroring the ONVIF service model:

```go
client.Device().GetDeviceInformation(ctx)
client.Media().GetProfiles(ctx)
client.PTZ().ContinuousMove(ctx, token, speed, timeout)
```

Facades are stateless views: accessors construct them on demand, which keeps
a single `Client` safe to share across goroutines without facade-level
locking (see [concurrency.md](concurrency.md)).

## The single call path

All ~220 service operations route through one dispatcher,
[`Client.call`](../#client-configuration). It:

1. snapshots the auth configuration (primary mode, fallback ladder, sticky
   result) under the read lock,
2. builds a stateless SOAP exchange for the chosen mode,
3. retries auth-class failures through the ladder and remembers the first
   working mode,
4. wraps auth-class failures so `errors.Is(err, ErrUnauthorized)` always
   holds.

One audited path is also what makes the concurrency contract provable — see
[authentication.md](authentication.md) for the ladder semantics.

## SOAP transport (internal/soap)

A per-call, stateless client that:

- builds the envelope and the WS-Security header for the configured mode
  (digest, password-text, or none) or sets HTTP Basic auth,
- **detects SOAP Faults on every call regardless of HTTP status** — ONVIF
  devices frequently return faults with HTTP 200, and before fault detection
  existed such responses were mistaken for success on void operations,
- returns structured errors: `*FaultError` (code/subcode/reason, both SOAP
  1.1 and 1.2 layouts, any namespace prefix) and `*HTTPStatusError`.

## Capability XAddr repair

Cameras frequently advertise service endpoints (`GetCapabilities` XAddrs)
with loopback addresses or stale IPs after roaming to a new address. Since
the client reached the device through its device-service URL, that URL's
host is authoritative: mismatching XAddr hosts are rewritten (preserving the
service-specific port) during `Initialize`.

## Why discovery does not import the root package

The `discovery` package is deliberately self-contained: it ships its own
minimal SOAP probing (HTTP POST + regex/xml extraction) so that consumers
who only need discovery do not pull in the full client, and so the package
can never develop an import cycle with the root package.
