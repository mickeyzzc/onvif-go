# Testing

The test suite runs entirely offline by default: every behavioral test works
against `httptest` mock devices or raw XML fixtures. Real-camera integration
tests exist but are gated behind environment variables and never run in CI.

## Running

```bash
make test        # go test -race ./...
make lint        # golangci-lint run (zero findings required)
make fmt         # gofumpt + goimports via golangci-lint fmt
make check       # lint + test
```

CI (`.github/workflows/ci.yml`) runs lint, a format check, race tests, and a
build on every push to `main`; all three jobs are required status checks.

A few discovery tests need the host to join the WS-Discovery multicast
group. On hosts without a usable multicast route (VPN-up laptops, some CI
runners) they skip automatically — the suite stays green.

## Test layers

| Layer | Where | What it covers |
|---|---|---|
| Mock-device tests | `*_test.go` next to the code | Full client behavior against `httptest` SOAP devices: auth ladder transitions, fault handling, response-shape variants, lifecycle of managed subscriptions |
| Raw fixtures | `testdata/captures/*.xml` | Hand-crafted real-shape envelopes replayed through the client (e.g. the GetStreamUri namespace variants behind issue #3) |
| Parser unit tests | e.g. `TestParseScopes`, `TestLooseExtractURI`, `TestSelectMainProfile` | Pure functions, no I/O |
| Concurrency matrix | `concurrency_test.go` | Mixed operations + config mutation on one shared client, meaningful under `-race` |
| Capture-replay helpers | `testing/` package | Mock server, capture registry, golden files — used by the larger suites |

## Real-camera integration tests

Selected test files (`device_real_camera_test.go`,
`media_real_camera_test.go`) run against actual hardware when (and only
when) you provide connection details:

```bash
export ONVIF_TEST_ENDPOINT="http://192.168.x.x/onvif/device_service"
export ONVIF_TEST_USERNAME="..."
export ONVIF_TEST_PASSWORD="..."
go test -v -run RealCamera ./...
```

Use placeholder credentials in any documentation or issues you file — never
real ones. When a device misbehaves, capture its raw SOAP response and add
it as a fixture under `testdata/captures/` so the regression lives on after
the camera is gone. `cmd/onvif-diagnostics` (see [cli.md](cli.md)) exists
exactly for producing those captures.

## Conventions for new tests

- Prefer `httptest` servers over hand-mocked transports; classify requests
  by body content, not by call order, so tests stay valid when the transport
  details change.
- Assert the *wire shape* for anything a device branches on (e.g. the
  StreamSetup fields of `GetStreamUri` requests).
- Time-sensitive loops (managed subscriptions, discovery listeners) should
  expose deterministic exit signals (`Done()` channels) rather than relying
  on sleeps.
