# Architecture

onvif-go is a zero-dependency Go library for talking to ONVIF devices, plus a
virtual camera server for testing recorders without hardware. This document
describes how the pieces fit together.

## Layering

```
┌────────────────────────────────────────────────────────────┐
│  cmd/            helper CLIs (discover, diagnostics, …)    │
│  examples/       runnable examples                          │
├────────────────────────────────────────────────────────────┤
│  Client (root package)                                      │
│  ├─ service facades: Device() Media() PTZ() Imaging()      │
│  │   Events() DeviceIO() Security()                        │
│  ├─ auth dispatcher: Client.call (ladder + sticky memory)  │
│  └─ shared state: credentials, clock skew, caps cache,     │
│     service endpoints (all mutex-guarded, issue #12)       │
├──────────────────────┬─────────────────────────────────────┤
│  internal/soap/      │  discovery/                          │
│  transport: envelope │  active probe, passive listener,    │
│  build/parse, WS-    │  directed HTTP probing, post-        │
│  Security header,    │  processing (filter/scopes/enrich). │
│  fault detection,    │  Self-contained: no dependency on   │
│  auth modes          │  the root package.                   │
├──────────────────────┴─────────────────────────────────────┤
│  server/          virtual ONVIF camera (simulator)         │
│  testing/         mock server, capture replay, goldens     │
│  testdata/captures/  real-camera SOAP fixtures             │
└────────────────────────────────────────────────────────────┘
```

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
