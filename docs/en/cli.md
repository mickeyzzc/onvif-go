# CLI Tools

The `cmd/` directory ships helper binaries. Build them for your platform:

```bash
make build          # go build ./... (fastest syntax/compile check)
make cross          # CGO_ENABLED=0 linux/arm64 binaries into build/
go build -o bin/ ./cmd/...
```

All tools are zero-dependency single binaries; there is no release pipeline
for them — build from source.

## discover

Multicast WS-Discovery probe with interface selection (useful on multi-NIC
hosts):

```bash
go run ./cmd/discover -timeout 5s
go run ./cmd/discover -interface eth0
```

## onvif-quick

One-shot device summary against one camera: device information, profiles,
stream URIs. Reads endpoint/credentials from flags. Handy as the "does this
library talk to my camera at all" probe.

```bash
go run ./cmd/onvif-quick
```

## onvif-diagnostics

Deep diagnostic collector for a specific camera — the tool to run when
filing an issue. Exercises all major operations, prints per-operation
results, and can capture raw SOAP exchanges for reuse as regression
fixtures:

```bash
go run ./cmd/onvif-diagnostics \
    -endpoint http://192.168.1.100/onvif/device_service \
    -username admin -password '***' \
    -verbose

# Capture raw XML request/response pairs (attach to issues after redacting
# credentials; they can become testdata/captures fixtures):
go run ./cmd/onvif-diagnostics -endpoint ... -username ... -password *** \
    -capture-xml -output diag.json
```

Flags: `-endpoint`, `-username`, `-password`, `-timeout` (seconds),
`-verbose`, `-capture-xml`, `-capture-all`, `-output`.

## onvif-server

The virtual ONVIF camera simulator, runnable standalone for testing a
recorder's discovery/onboarding flow without hardware:

```bash
go run ./cmd/onvif-server -port 8000 -manufacturer TestCam -model V1 \
    -serial SIM-0001
```

Feature toggles: `-info`, `-ptz`, `-imaging`, `-events`, `-version`.

## generate-tests

Developer tool: converts captured SOAP exchanges (from `onvif-diagnostics`)
into Go test scaffolding and maintains the capture registry. See
[testing.md](testing.md) for the fixture workflow.
