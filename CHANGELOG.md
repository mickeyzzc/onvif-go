# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added (server, embedding + discovery)
- **`Server.RegisterServices(mux)` / `Server.Handler()`** (#35): hosts
  mount the ONVIF services plus the snapshot endpoint on their own
  mux/router — their routes coexist, `Start` is never required. `Start`
  itself now serves through `Handler()`.
- **`Config.Logger *slog.Logger`** (#35): startup/shutdown messages go
  through structured logging; nil → completely silent (no stdout). The
  emoji banner is gone — CLIs pass their own logger.

### Changed (server/discovery)
- **WS-Discovery XAddrs default no longer echoes the requester** (#38):
  an unset `Config.XAddrs` now derives the device's own source address
  toward the peer (the interface a reply leaves from; throwaway UDP
  dial, no packets sent), falling back to the configured interface's
  IPv4 address, then loopback. Rationale: NVR-style consumers register
  the XAddrs host as the device endpoint — an echoed requester IP makes
  them register *themselves* as the camera (observed with MiBee NVR).
  Multi-homed hosts should still set `XAddrs` explicitly.

### Added (server, embedded-host provider surface)
- **`StreamInfo.RTSPPort`** (#34): GetStreamUri derives
  `rtsp://<host>:<RTSPPort|8554><path>` instead of hardcoding 8554;
  OverrideURI stays verbatim. The startup banner and ServerInfo use the
  same derivation.
- **Device-service `GetScopes`** (#37): returns `Config.Scopes`
  (empty → the conventional `DefaultScopes`), wire form
  `tt:ScopeDefinition Scopeitem="…"`. Hosts should mirror their
  discovery Responder Scopes so ProbeMatches and GetScopes agree.
- **Snapshot-chain adaptations** (#36): `SnapshotProvider` now returns
  `SnapshotResult{Data, ContentType}` (empty ContentType = image/jpeg)
  so devices with fallback captures (e.g. cached H.264 IDR frames) can
  express non-JPEG results; the HTTP snapshot endpoint accepts requests
  without `?profile=` (serves the first snapshot-enabled profile); the
  endpoint path and advertised URI follow `Config.SnapshotPath` (empty →
  the historical `BasePath/snapshot`) and `Config.SnapshotURIParameterless`
  drops the `?profile=` query for parameterless device semantics.

### Changed
- `New` normalizes a shallow copy of the caller's `*Config` (fills the
  SnapshotPath default); `GetConfig` no longer returns the caller's
  pointer.

### Fixed (server, embedded-host interop)
- **WS-Security digest auth now accepts the misspelled utility namespace**:
  many community clients emit `wsu:Created` under
  `oasis-200401-wss-wssecurity-utility-1.0.xsd` instead of the canonical
  `oasis-200401-wss-utility-1.0.xsd`; the strict namespace match silently
  decoded an empty timestamp and every digest failed with a misleading
  "invalid password" fault. Both variants now authenticate (#40).
- **New `soap.RawEnvelope` response channel**: a handler returning
  `RawEnvelope` has its bytes served as the complete response document —
  no wrapping, no re-serialization. Unblocks serving WS-Discovery
  ProbeMatches (with WS-Addressing RelatesTo) from inside the SOAP
  handler (#39).

## [v2.0.0-rc2] - 2026-08-27

### Changed (breaking vs rc1 — import path only)
- **Client package moved to `onvif/`**: the repository root no longer
  contains any Go source. Import
  `github.com/mickeyzzc/onvif-go/v2/onvif` instead of
  `github.com/mickeyzzc/onvif-go/v2`. The module path, package identifier
  (`onvif`), and every exported symbol are unchanged — consumers only
  adjust the import path. Rationale: the rc1 root directory still carried
  23 Go files (client + auth + errors + v1 aliases + integration tests);
  the subdirectory consolidates them while keeping the `onvif.` identifier
  stable. Tagged during the rc window; v2.0.0 final ships this layout.

### Added
- **Comprehensive TDD test backfill** (repo-wide): in-package suites for
  the domain services (`device`, `media`, `ptz`, `imaging`, `security`)
  through the new `internal/testutil.FakeCaller` (exact transport decode
  path, no sockets); full `events` coverage (service ops + the managed
  `EventStream` lifecycle: delivery, renewal, panic isolation,
  cancellation, renewal-failure termination, idle throttling, transient
  error recovery); `server/simulator` in-package suite (stream
  lifecycle, JPEG snapshot caching, PTZ ops + completion timers,
  imaging round-trip); `server/discovery` lifecycle edges;
  `server/soap` message helpers; and the `testing` capture/golden
  infrastructure. TDD conventions documented in
  `docs/{en,zh}/testing.md`.

### Fixed
- `server/soap.Handler`: registering handlers while serving was an
  unsynchronized map access (data race under `-race`) — the handler map
  is now RWMutex-guarded.
- `imaging.Move` never encoded its focus parameters (`FocusMove` was an
  empty placeholder): `FocusMove`/`AbsoluteFocus`/`RelativeFocus`/
  `ContinuousFocus` are real types now and reach the wire.
- `imaging.GetOptions` silently dropped every option group except
  Brightness/ColorSaturation/Contrast — full decode + mapping
  (BacklightCompensation, Exposure, Focus, WDR, WhiteBalance,
  IrCutFilterModes, Sharpness).
- Imaging endpoint guards were a dead no-op retry; they now return
  `types.ErrServiceNotSupported` like PTZ.
- `discovery.Listener.Stop` and `server/discovery.Responder.Stop` took
  up to one read-deadline tick (5s) to unblock the read loop — both now
  close the multicast socket for immediate shutdown (with the
  Stop-before-bind race covered), and `Responder.Start` validates its
  interface synchronously instead of swallowing the error in the loop
  goroutine.

## [v2.0.0-rc1] - 2026-08-27

The v2 release candidate: module path `/v2`, client domain-package
split (M1), rebuilt server SOAP transport (M2), pluggable provider
architecture (M3), and the shared WS-Discovery codec with a device-side
responder (M4). See `MIGRATION.md` for the complete v1→v2 mapping and
`docs/{en,zh}/v2-architecture.md` for the plan of record.

### Changed
- **v2 module path and package split** (#20/#21): the module is now
  `github.com/mickeyzzc/onvif-go/v2`. The client split into domain packages
  (`device`, `security`, `deviceio`, `media`, `ptz`, `imaging`, `events`) +
  a shared `types` leaf, unlocked by the new `internal/api.Caller`
  interface; the root package keeps `Client`, auth strategy, download, and
  v1-compatibility aliases so most v1 source compiles after the import-path
  swap. Facade accessors are unchanged; services are long-lived instances
  (the capabilities cache lives on `device.Service`); v1 per-service
  endpoint setters collapse into `Client.SetServiceEndpoint`.
- **Architecture-mapped file layout** (no API change — the root package
  stays one package, which is what the facade API requires): domain-sibling
  files merged into their leads (`device_additional`→`device`,
  `device_network_config`→`device_mgmt` (from `device_extended`),
  `device_certificates`→`device_security`, `media_profile_select`→
  `media_profiles`, `event_managed`→`event`), diagnostics/cache files
  renamed to `auth.go`/`capabilities.go`, and the HTTP Digest transport
  extracted to `internal/httpdigest` (27 → 22 source files; test files
  consolidated to match).
- **Documentation diagrams are now mermaid** instead of ASCII box art, with
  a root-package file map added to the architecture guide (en/zh).
- **`server/soap` transport rebuilt for real-device embedding** (#22,
  closes #16/#17/#18 — breaking):
  - Handlers receive raw inner-XML request bytes (`[]byte`); the v1
    transport handed them an `encoding/xml` value it could never populate,
    so request parameters (e.g. `ProfileToken`) arrived silently empty.
  - Authentication is per-action (#16): with credentials configured,
    write-style actions (`Set*`/`Remove*`/`Create*`/`Go*` + `SystemReboot`)
    require valid WS-Security; reads stay open (was: every action forced
    digest auth). Both PasswordDigest and PasswordText are accepted;
    `soap.AuthPolicy` configures prefixes/exact actions/strict mode.
  - `soap.RequestContext` (#17): `RegisterContextHandler` exposes the
    action name, client IP, and the request `context.Context`. Advertised
    URLs (GetCapabilities/GetServices XAddrs, stream/snapshot URIs) echo
    the requesting client's IP by default; `Config.AdvertiseHost` overrides.
  - Byte-predictable responses (#18): canonical WSDL casing
    (`GetStreamUriResponse`/`GetSnapshotUriResponse`, inner `MediaUri`/
    `Uri`; legacy incoming spellings still dispatch), golden-locked
    envelope layout, optional explicit prefixes
    (`Config.ExplicitPrefixes` → `s:`/`tds:`/`trt:`/...), and a
    `soap.RawXML` passthrough that embeds hand-built bodies verbatim.
  - The server package's `Server.Handle*` methods take
    `(*soap.RequestContext, []byte)`; simulator `StreamURI` values are
    derived per request instead of pre-baked at `New()` (explicit
    `UpdateStreamURI` pins still win).
- **Server provider architecture** (#23, closes #19 — the simulator is
  now a swappable backend): all state sources behind the SOAP layer are
  interfaces in the new `server/provider` package (`DeviceInfoProvider`,
  `StreamURIProvider` + optional `StreamURISetter`, `SnapshotProvider`,
  `ImagingProvider`, `PTZProvider`); the in-memory state moved to
  `server/simulator` as the default; handlers are stateless translators;
  the HTTP snapshot endpoint serves provider JPEGs (the simulator
  generates real, decodable JPEGs — the empty-body TODO is gone);
  `server.New(config, ...server.Option)` injects hardware-backed
  providers without forking. The data model and error sentinels live in
  `server/provider` with `server.*` aliases, so CLI/examples and the
  exported surface are unchanged (one exception:
  `ProfileConfig.ToONVIFProfile()` became `server.ProfileToONVIF()` —
  methods cannot follow type aliases across packages).
- **v2 finalization** (#25): full v1-path sweep across docs and READMEs,
  `MIGRATION.md` (complete v1→v2 mapping incl. the server transport and
  provider changes), architecture status updated to implemented, and
  this release-candidate section.
- **Shared WS-Discovery codec + device-side responder** (#24, closes
  #15): new `wsdiscovery` leaf package holds the wire codec for both
  sides of the protocol (`BuildProbe`/`ParseProbe`,
  `BuildProbeMatches`/`ParseProbeMatches` (+ `ErrNoMatches`),
  `BuildHello`/`BuildBye`, `ParseAnnouncement`) — the client
  (`discovery/`) now parses through it, eliminating its private
  decoders. New `server/discovery.Responder`: resident multicast :3702
  loop answering Probes with unicast ProbeMatches, Hello on start / Bye
  on stop, full `http.Handler` for directed Probe-over-HTTP POSTs;
  advertised Types/Scopes/XAddrs/MetadataVersion configurable, XAddrs
  defaulting to the per-peer `http://<requester IP>:<port><path>` echo;
  WS-Discovery Types filtering honored. `discovery.ProbeMatch` is now an
  alias of `wsdiscovery.Match` (field-compatible; the unused XMLName
  field is gone).
- **Fixed the WS-Security utility namespace typo** in the client's
  `UsernameToken.Created` tag (`oasis-200401-wsssecurity-utility-1.0.xsd`
  → the correct `oasis-200401-wss-utility-1.0.xsd`), so `Created` is now
  emitted in (and matched from) the namespace strict devices expect.

### Added
- **Bilingual documentation set**: topic guides under `docs/{en,zh}/`
  (architecture, authentication, discovery, media, events, concurrency,
  testing, CLI) and a full Chinese README (`README.zh.md`), mirroring the
  MiBeeNvr documentation standard.
- **Auth ladder** (#1): `WithAuthMode` (digest / password-text / HTTP Basic /
  none), `WithAuthFallback` with sticky first-working-mode memory, and the
  `errors.Is(err, onvif.ErrUnauthorized)` sentinel covering HTTP 401/403,
  NotAuthorized SOAP faults, and 200-with-fault responses. All service
  operations route through a single audited call dispatcher.
- **StreamSetup parameterization** (#2): `GetStreamURIWithOptions` selects
  stream type × transport protocol (RTSP/HTTP/UDP, unicast/multicast);
  `GetStreamURI` delegates with the historical defaults.
- **Robust media URI parsing** (#3): namespace/SOAP-version-tolerant response
  handling with a local-name fallback extractor; empty URIs are now explicit
  `ErrEmptyMediaURI` errors carrying a body summary (was: silent empty
  string + nil error). Regression fixtures under `testdata/captures/`.
- **Directed unicast probing** (#4): `discovery.ProbeEndpoint` (WS-Discovery
  Probe over HTTP with GetDeviceInformation fallback — works across subnets)
  and `discovery.ProbeSerial` (namespace-agnostic serial extraction, common
  port scan).
- **Passive discovery listener** (#5): `discovery.Listener` on the WS-Discovery
  multicast group for camera power-on Hello and third-party ProbeMatches,
  coexisting with active `Discover()` sockets; panic-isolated handlers,
  idempotent Stop, context-aware shutdown.
- **Managed pull-point subscriptions** (#6): `Events().SubscribeEvents` runs
  the whole lifecycle (long-poll dispatch, pre-expiry auto-renewal,
  termination on renewal failure, best-effort cleanup) with the
  `ErrEventsNotSupported` sentinel for devices that advertise but do not
  implement the events service.
- **Main/sub stream profile selection** (#7): `SelectMainProfile` /
  `SelectSubProfile` — resolution-first heuristics with bilingual naming
  tie-breaks and the Amcrest same-resolution-dual-token exclusion.
- **Clock-skew measurement & auth diagnosis** (#8): `WithAutoClockSkew`
  (RTT-compensated, failure-tolerant), `MeasureClockSkew`,
  `DiagnoseAuth` (clock-skew / bad-credentials / ok triage).
- **Discovery post-processing** (#9): `ParseScopes`, structured
  `Device.Name/Hardware/Location`, `FilterONVIFDevices` (ghost-responder
  cull), parallel `EnrichDevices`; `Device.EndpointRef` documented as
  NOT a serial number.
- **Network config setters** (#10): `SetNetworkInterfaces` (DHCP/manual IPv4,
  RebootNeeded round trip), `NetmaskFromPrefixLength` +
  `PrefixedIPv4Address.Netmask` convenience.
- **Capabilities cache** (#11): `GetCapabilitiesCached` (single-flighted,
  shared pointer semantics), `InvalidateCapabilitiesCache`, opt-in
  `WithMinimalCapsFallback` degradation for weak devices.

### Fixed
- **Fault detection everywhere** (#1, #3): SOAP Faults carried with HTTP 200
  were previously missed for void operations (reported as success) and
  surfaced as confusing unmarshal errors elsewhere; every call now returns a
  structured `*FaultError` / `*HTTPStatusError`.
- **Concurrency contract** (#12): endpoint getters raced `Initialize`;
  PTZ/Imaging calls before `Initialize` used an empty endpoint. All endpoint
  access is lock-guarded with device-endpoint fallback; the Client's
  concurrency safety is documented and pinned by a mixed-operation race
  matrix test.

## [1.2.0] - 2026-08-26

### Changed
- **Module path**: `github.com/0x524a/onvif-go` → `github.com/mickeyzzc/onvif-go`.
  The project became a fully-owned continuation of 0x524a/onvif-go; tags
  v1.0.0–v1.1.7 (0x524a lineage) remain on the repository for history.
- **Service-facade API**: `client.Device()`, `client.Media()`, `client.PTZ()`,
  `client.Imaging()`, `client.Events()`, `client.DeviceIO()`,
  `client.Security()` — the Client is a connection + configuration holder,
  operations live on the service that owns them (mirrors the ONVIF service
  model). Legacy delegators on Client were removed.
- **Media split**: media.go split into topic files; 130 method-local types
  hoisted.
- **Toolchain**: Go 1.26; zero third-party dependencies; gofumpt + goimports
  + golangci-lint v2 at zero findings; `main` is the protected primary branch
  with required lint/test/build checks.

## [1.1.3] - 2025-11-18

### Changed
- **Release Workflow**: Create releases as draft initially
  - Fixes "Cannot upload assets to an immutable release" error
  - Releases must be manually published after assets upload
  - Prevents race condition where release publishes before all assets finish uploading

## [1.1.2] - 2025-11-18

### Changed
- **Release Workflow**: Upgraded to `softprops/action-gh-release@v2`
  - Fixes asset upload race condition in v1
  - Better handling of concurrent file uploads
  - Added `fail_on_unmatched_files` and `make_latest` flags

## [1.1.1] - 2025-11-18

### Added
- **RTSPeek Library Integration**: RTSP stream inspection using `github.com/0x524A/rtspeek`
  - Replaced command-line `ffprobe` execution with library-based approach
  - Enhanced stream inspection with codec, resolution, and framerate detection
  - 5-second timeout for stream DESCRIBE operations
  - TCP fallback for basic connectivity checks
  - See `cmd/onvif-cli/main.go` for implementation

### Changed
- **Code Quality Improvements**: Fixed all linting errors
  - Removed unused `generateDemoASCII()` function
  - Fixed dynamic format strings (SA1006 errors)
  - Added proper error handling for Close() operations
  - Migrated to golangci-lint v2 configuration
  - CI/CD pipeline excludes utility tools and examples from linting
- **golangci-lint v2**: Updated configuration and GitHub Actions workflow
  - Created `.golangci.yml` with v2 schema
  - Updated CI to use golangci-lint-action@v8 with v2.2
  - Scoped linting to main packages only

## [1.1.0] - 2025-11-18

### Added
- **Simplified Endpoint API**: `NewClient()` now accepts multiple endpoint formats
  - Simple IP address: `"192.168.1.100"`
  - IP with port: `"192.168.1.100:8080"`
  - Full URL: `"http://192.168.1.100/onvif/device_service"` (backward compatible)
  - Automatically adds `http://` scheme and `/onvif/device_service` path when needed
  - See `docs/SIMPLIFIED_ENDPOINT.md` for details
- **Localhost URL Fix**: Automatic handling of cameras that report localhost addresses
  - Detects and fixes localhost/127.0.0.1/0.0.0.0/::1 in GetCapabilities response
  - Replaces with actual camera IP address
  - Preserves service-specific ports when specified
  - Handles common camera firmware bugs transparently
- Comprehensive test coverage for endpoint normalization (12 test cases)
- Comprehensive test coverage for localhost URL handling (10 test cases)
- New example: `examples/simplified-endpoint/` demonstrating all endpoint formats
- Documentation: `docs/PROJECT_STRUCTURE.md` explaining project organization
- Initial release of onvif-go library

### Changed
- **Project Structure**: Implemented ideal Go project layout
  - Moved `soap/` to `internal/soap/` (private implementation)
  - Moved `test/test-server.go` to `examples/test-server/` for clarity
  - Removed empty `test/` directory
  - Public API remains at root level for clean imports
  - Follows Standard Go Project Layout for libraries
  - Updated all imports throughout codebase
  - See `docs/PROJECT_STRUCTURE.md` and `docs/ARCHITECTURE.md` for details
- Updated `docs/ARCHITECTURE.md` to reflect new project structure
- Updated module path from `github.com/0x524A/onvif-go` to `github.com/0x524a/onvif-go` (lowercase)
- ONVIF Client with context support
- Device service implementation
  - GetDeviceInformation
  - GetCapabilities
  - GetSystemDateAndTime
  - SystemReboot
- Media service implementation
  - GetProfiles
  - GetStreamURI (RTSP/HTTP)
  - GetSnapshotURI
  - GetVideoEncoderConfiguration
- PTZ service implementation
  - ContinuousMove
  - AbsoluteMove
  - RelativeMove
  - Stop
  - GetStatus
  - GetPresets
  - GotoPreset
- Imaging service implementation
  - GetImagingSettings
  - SetImagingSettings
  - Move (focus control)
- WS-Discovery implementation
  - Automatic device discovery via multicast
- SOAP client with WS-Security
  - UsernameToken authentication
  - Password digest (SHA-1)
- Comprehensive type definitions
- Error handling with typed errors
- Connection pooling for performance
- Complete examples
  - Discovery
  - Device information
  - PTZ control
  - Imaging settings
- Comprehensive documentation
- README with usage guide

[Unreleased]: https://github.com/0x524a/onvif-go/compare/v1.1.3...HEAD
[1.1.3]: https://github.com/0x524a/onvif-go/compare/v1.1.2...v1.1.3
[1.1.2]: https://github.com/0x524a/onvif-go/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/0x524a/onvif-go/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/0x524a/onvif-go/compare/v1.0.3...v1.1.0
