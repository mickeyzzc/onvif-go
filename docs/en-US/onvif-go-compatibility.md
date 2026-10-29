# Real-device compatibility matrix (per-model evidence)

The README's compatibility claims come from real deployments. This page
pins each claim to **verifiable model-level evidence**: firmware,
the operation surface exercised, the auth mode that actually worked,
quirks and their handling (linked to code / docs / fixtures / issues).
Long-tail searches ("DS-2CD2143G2-I onvif golang") land here directly.

**Policy**: only rows with first-hand evidence are listed — a diagnostics
report, a real-device field vector from an issue, or production logs.
"Should work" is not listed. The last column is the evidence date and
the library line at the time.

| Vendor | Model / firmware | Verified operation surface | Auth that worked | Quirk → handling | Last verified |
|---|---|---|---|---|---|
| Hikvision | DS-2CD2143G2-I @ 5.7.1 | All 11 diagnostics operations (device info/capabilities/time/profiles/encoders/stream URIs/snapshot/PTZ) | WSSE digest | none — the reference-behavior device; the diagnostics manual's sample artifact is named after it ([diagnostics manual](onvif-go-cli-diagnostics.md)) | 2026-09-20 · v2.2-era |
| Hikvision | DS-IPC-T12HV3-IA | MotionAlarm event fields | — | event field vector pinned as a fixture ([#126](https://github.com/mickeyzzc/onvif-go/issues/126)) | 2026-10 · v2.5-era |
| Hikvision | (model TBD) | digest auth | WSSE digest | "sender not authorized" was actually **clock skew** — a digest timestamp outside the window misreported as an auth failure; `WithAutoClockSkew` auto-corrects ([authentication](onvif-go-authentication.md)) | 2026-09 · v2.1-era |
| TP-Link | TL-IPC44AW-COLOR | Motion events (CellMotionDetector) | — | `IsMotion` dialect: this family uses CellMotionDetector instead of VideoSource/MotionAlarm (#126 field vector) | 2026-10 · v2.5-era |
| Amcrest | IP4M family | profiles/stream URIs/sub-stream selection | WSSE digest | same-resolution **dual tokens**: sub-stream selection by resolution heuristic ([media](onvif-go-media.md)) | 2026-10 · v2.3-era |
| iDRPENG (Zhongwei-era platform) | PSH_H_V1.0 @ V2.2.1144, ONVIF 2.4 | connect/device info/profiles/stream URIs | digest (after 401 challenge) | **Digest → 200 + empty body**: the old stack answers a valid response with an empty SOAP body; the auth ladder's empty-body tolerance is for this device ([#140](https://github.com/mickeyzzc/onvif-go/issues/140)/#141) | 2026-10-14 · v2.6.0 |
| Dahua | (model TBD) | capabilities/profiles | WSSE digest | strict envelope-namespace parsing — envelopes without default-namespace declarations were skipped (v2.1 wire fix) | 2026-09 · v2.1-era |
| Uniview / TP-Link / in-house firmware | — | wsnt:Subscribe push subscription + Notify receive | — | push-subscribe envelope verified on real devices (v2.4 batch) | 2026-10 · v2.4-era |
| HiSilicon OEM / ESP32 / fault-only odd firmware | — | connect ladder liveness | ladder levels | fault-only firmware: `DiagnoseAuth` finds the usable rung ([authentication](onvif-go-authentication.md)) | ongoing · production |

## Quirk → handling index

- **Clock skew misreported as auth failure** → `onvif.WithAutoClockSkew`
- **Digest 200 with empty body** → auth-ladder empty-body tolerance (#141, v2.6.0)
- **Same-resolution dual tokens** → `SelectSubProfile` resolution heuristic (media package)
- **CellMotionDetector dialect** → dual-topic event parsing (#126)
- **Strict-namespace envelopes** → per-prefix decoder binding (v2.1 fix)
- **Fault-only firmware** → `DiagnoseAuth` (auth diagnostics)

## Maintenance rules

1. After every real-device integration, **add the row in the same PR**:
   model/firmware/operations/auth rung/quirk, with the diagnostics-report
   file name or issue number;
2. in-row links must point inside this repository (docs pages / fixtures /
   issues) — external links rot;
3. "model TBD" rows **stay** (honest evidence gaps); remove the note once
   filled.
