# The no-hardware ONVIF loop: virtual camera × CLI × NVR integration

The most-imported ONVIF library in the Go ecosystem,
[use-go/onvif](https://github.com/use-go/onvif), is no longer releasing
(last release 2022-07, v0.0.9). This tutorial is the "current path" that
fills the gap: every part of this repository — the virtual camera server,
the unified CLI, the diagnostics report, the client API — chained into
one **end-to-end loop that never needs a real camera**. Every step comes
with paste-ready commands and real output.

## 0. One binary to start

```bash
# Prebuilt bundle (linux/darwin/windows, static, zero dependencies)
gh release download v2.7.0 -R mickeyzzc/onvif-go -p 'onvif_linux_amd64.tar.gz'
tar xzf onvif_linux_amd64.tar.gz && sha256sum -c SHA256SUMS

# Or version-pinned install / run from a checkout
go install github.com/mickeyzzc/onvif-go/cmd/onvif@v2.7.1
go run ./cmd/onvif <subcommand>
```

## 1. `onvif server`: stand up a virtual multi-lens camera

```bash
onvif server -password demo -manufacturer TestCam -model VC-1 -serial SIM-0001 -profiles 2
```

A multi-lens IP camera is now running (service tree mounted at
`http://<host>:<port>/onvif/device_service`; `SIGINT` shuts down
gracefully):

- **Identity quadruple** `-manufacturer/-model/-firmware/-serial` —
  `GetDeviceInformation` reports exactly these; that is what a recorder's
  device list will show;
- **`-profiles 1..10`**: ten rotating templates (1080p main, 720p wide,
  4K UHD, fisheye, thermal, 60 fps license-plate, …), each with an H264
  encoder and snapshot configuration — main/sub-stream selection and
  multi-channel UI layout all have real data to chew on;
- **`-ptz` (default on)**: the PTZ service answers `GetStatus`,
  continuous/absolute moves and presets; **`-events`** enables the
  pull-point event service;
- **Password is mandatory** (flag or `ONVIF_SERVER_PASSWORD`) — the
  empty-credentials mode belongs to the embeddable package, not to a
  demo binary on a real network.

More flags, and how to embed the server in your own binary:
the [server manual](onvif-go-cli-server.md).

## 2. `onvif discover`: find it on the LAN

```bash
onvif discover              # active WS-Discovery Probe (all interfaces)
onvif discover -interface eth0 -timeout 10s
```

The virtual camera you just started shows up on the same segment with the
identity you set. The three discovery mechanisms — active probe, passive
Hello listening, cross-segment directed probing — plus **ghost responder
filtering** (garbling devices never reach the result list) are covered in
the [discover manual](onvif-go-cli-discover.md) and
[discovery guide](onvif-go-discovery.md).

## 3. `onvif quick`: the interactive checkup

```bash
onvif quick
```

Menu-driven: `3` connect → `5` stream URLs → `4` PTZ demo. **Key hint:
enter a bare `host:port` at "Camera IP:"** — the tool builds the
`http://` prefix and the `/onvif/device_service` suffix itself; a full
URL gets doubled:

```text
Choice: 3
Camera IP: 127.0.0.1:18081
Username [admin]: admin
Password: demo
✅ Connected!
📡 Stream: rtsp://127.0.0.1:8554/stream0
```

Once connected the menu also offers the PTZ demo (a lap of
continuous/absolute/preset against the virtual head). The full
interaction walkthrough: the [quick manual](onvif-go-cli-quick.md).

## 4. `onvif diagnostics`: the 11-operation JSON report

```bash
onvif diagnostics -endpoint http://127.0.0.1:18081/onvif/device_service \
  -username admin -password demo -output ./camera-logs
```

Eleven operation families in one pass (device information, capabilities,
time, profile list, per-profile video-encoder configurations and
stream/snapshot URIs, PTZ status and presets), written to a JSON report
named after the device identity and timestamp:

```text
✓ Diagnostic collection complete!
  Report saved to: ./camera-logs/TestCam_VC-1_1.0.0_20261029-064143.json
  Total errors: 2
```

The report is the chart: `device_info`/`profiles` for identity and
capability surface; error entries like `video_encoders[i].error` carry
**which operation, which profile, and the raw SOAP envelope** — look
there first when something misbehaves. The
[diagnostics manual](onvif-go-cli-diagnostics.md) walks every section and
the `-capture-xml` raw-traffic archive.

**House rule: file issues on top of this report.** Against real hardware
or the virtual camera alike, run `diagnostics` first and attach the JSON
(optionally the `-capture-xml` archive) — reproducible evidence saves a
round trip.

## 5. Close the loop: feed it to an NVR

The last step that proves the loop needs no hardware: hand the virtual
camera to a real recorder.

- Endpoint: `http://<host>:<port>/onvif/device_service`, credentials
  `admin` / your password;
- Any ONVIF-capable recorder (NVR/VMS): use "add by address" — discovery,
  live video, recording and PTZ all ride the standard protocol;
- `-profiles 10` gives one process ten channels — exactly what a UI's
  main/sub-stream switching and multi-channel layout tests want.

MiBeeNvr uses this as its standard onboarding flow; other vendors'
recorders work the same way — the protocol face is ONVIF Profile S +
Media2.

## 6. CLI → code: the same thing as an API

The CLI is a shell over the same client. Everything above, as Go:

```go
import "github.com/mickeyzzc/onvif-go/v2"

client, _ := onvif.NewClient("127.0.0.1:18081",
    onvif.WithCredentials("admin", "demo"))
profiles, _ := client.Media().GetProfiles(ctx)
main := onvif.SelectMainProfile(profiles) // main-stream profile token
uri, _ := client.Media().GetStreamURI(ctx, main)
```

The full client surface (authentication ladder, Media2/Profile M/imaging,
event subscriptions): the [quickstart](onvif-go-quickstart.md),
[authentication](onvif-go-authentication.md) and
[events](onvif-go-events.md) guides; putting the virtual camera into a
test pipeline: [testing](onvif-go-testing.md).
