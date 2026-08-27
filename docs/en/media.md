# Media and Streaming

The Media service facade (`client.Media()`) covers profiles, stream and
snapshot URIs, and encoder configuration. This document covers the parts with
field-tested semantics: profile selection, stream-setup parameterization,
and response parsing.

## Choosing the right profile

Blindly using `profiles[0]` silently records a substream on devices that list
the low-resolution profile first — a failure that looks like everything works
until you notice the resolution. Two helpers encode the field-tested
heuristics:

```go
profiles, _ := client.Media().GetProfiles(ctx)

mainToken := onvif.SelectMainProfile(profiles)
subToken  := onvif.SelectSubProfile(profiles, mainToken) // "" = no substream
```

**`SelectMainProfile`** — the highest-pixel-count profile (W×H). Exact ties
are settled by naming hints only (`main`/`primary`/`主流`/`主码流`/`channel1`
win; `sub`/`secondary`/`辅流`/`辅码流`/`extra` lose), because OEM firmware
assigns arbitrary names and naming can never be the primary signal. When
nothing carries resolution information, the first profile is returned
(list-order fallback).

**`SelectSubProfile`** — the largest remaining profile that is *strictly*
smaller than main. A second profile at the *same* resolution as main is not a
substream: on some hardware (the Amcrest IP4M pattern) two tokens at the same
resolution are two handles onto the same stream. `""` means no independent
substream exists.

## Stream URIs with an explicit transport

`GetStreamURI` requests RTP-Unicast + RTSP. Some devices decide what to
return based on the requested protocol — ESP32 firmwares return an RTSP URL
with a G.711 audio track only when asked for RTSP, and an HTTP video-only
URL otherwise. `GetStreamURIWithOptions` exposes the choice:

```go
uri, err := client.Media().GetStreamURIWithOptions(ctx, profileToken,
    onvif.StreamSetup{
        Stream:    onvif.StreamRTPUnicast,   // or StreamRTPMulticast
        Transport: &onvif.Transport{Protocol: onvif.ProtocolRTSP}, // HTTP / UDP / TCP
    })
```

A nil or empty transport defaults to RTSP. `GetStreamURI(ctx, token)` is
exactly `RTP-Unicast + RTSP` — unchanged behavior.

## Response parsing guarantees

ONVIF media responses vary more than the spec suggests: namespace prefixes
(`trt:`/`tt:`/default), SOAP 1.1 vs 1.2 envelopes, and the occasional missing
`MediaUri` wrapper. Parsing is layered:

1. typed structs match by local name — any prefix, either SOAP version;
2. if the typed path yields no URI, a local-name scan extracts the first
   `Uri` element from the raw response content;
3. if there is still no URI, you get an explicit `ErrEmptyMediaURI` error
   carrying a truncated body summary — never the historical silent
   empty-string-with-nil-error;
4. a SOAP Fault is detected regardless of the HTTP status it arrived with
   (200-with-Fault included) and returned as a structured `*FaultError`.

`GetSnapshotURI` shares the response shape and the same guarantees.

## Encoder and OSD configuration

Beyond streaming, the facade covers video/audio encoder configuration
(`Get/SetVideoEncoderConfiguration`, `SetVideoEncoderConfiguration` families),
OSD management, multicast configuration (`Start/StopMulticastStreaming`), and
synchronization points. See the
[Go reference](https://pkg.go.dev/github.com/mickeyzzc/onvif-go/v2) for the full
operation list.
