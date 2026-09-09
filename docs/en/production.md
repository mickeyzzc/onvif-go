# Production deployment & hardening

Deploying the onvif-go server (or an embedding host) onto a real network.
The theme throughout: the library ships safe defaults, but a camera
backend is internet-adjacent infrastructure — every knob below exists
because a deployment needed it.

## TLS deployment

The SOAP server speaks plain HTTP (ONVIF's common case). Terminate TLS at
your ingress and keep the library behind it:

```nginx
# nginx: ONVIF over HTTPS, plain HTTP to the onvif-go server on :8080
location /onvif/ {
    proxy_pass http://127.0.0.1:8080/onvif/;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
}
```

Two things follow from termination:

1. **Advertise the right host.** Discovery `XAddrs` and `GetStreamUri`
   answers must carry the address the *client* can reach — set
   `Config.AdvertiseHost` (or `WithAdvertiseHostProvider` when the
   address changes at runtime, e.g. DHCP renewals). The responder never
   echoes the requester's address back (that bug registers the NVR as
   the camera — see the discovery guide).
2. **Client IP for auth lockout.** The brute-force backstop keys on the
   source IP. Behind a proxy, forward the real client IP (the snippet
   above does) and make sure your HTTP layer surfaces it as
   `RemoteAddr` or via the embedding host's `RequestContext`.

Multi-homed hosts (common on cameras with wired + wireless): set
`XAddrs` explicitly rather than relying on the derived per-peer address.

## Choosing an authentication policy

`Config.Username`/`Config.Password` enable WS-UsernameToken validation;
the per-action policy decides what actually requires it (details in the
[authentication guide](authentication.md)):

| Deployment | Policy |
|---|---|
| Closed monitoring LAN, NVR-only clients | Defaults: write-style actions authenticated (`Set*`, `Remove*`, `Create*`, `Go*`, `SystemReboot`), reads open — NVRs that only pull profiles/streams need no credentials. |
| Anything reachable beyond the switch | `AuthPolicy{All: true}` — every action authenticated. Pair with digest tokens (see below). |
| Lab bench / capture replay | `AllowAnonymous` documented mode — never on a production network. |

Prefer `PasswordDigest` clients: the password never crosses the wire and
the replay window (nonce + Created) holds. `AllowPasswordText` defaults
on for interop; turn it off (`AuthPolicy.AllowPasswordText = false`)
when you control both ends.

## Rate limiting & body bounds

- `HandlerOptions.MaxBodyBytes` (default 1 MiB): larger SOAP requests
  are refused with 413 before parsing. Keep the default; raise only for
  oversized vendor extensions you actually need.
- `AuthFailureLimit` (default 5) / `AuthLockout` (default 60s): the
  per-source brute-force backstop. Locked-out sources get 401 without
  any credential work. Alert on lockouts (below) — they are someone
  knocking.
- WS-Discovery: the responder answers Probes on the multicast group.
  It is unauthenticated by design; nothing sensitive is disclosed
  beyond scopes you configured. Keep `Scopes` free of site names if
  discovery reaches untrusted segments.

## Wiring observability

The `metrics` package is the seam (no Prometheus dependency in this
module):

```go
bridge := myprom.NewOnvifBridge()           // implements metrics.Hooks
srv, _ := server.New(cfg, server.WithMetrics(bridge))
responder := discoveryserver.NewResponder(discoveryserver.Config{
    // ...
    Metrics: bridge,
})
```

The five counters worth dashboards: `SoapRequest` by action (traffic
mix), `SoapFault` by action (fault-rate numerator — handler errors, not
auth rejections), `AuthFail` (credential guessing), `AuthLockout`
(active lockouts), `DiscoveryProbeAnswered` (scanner noise baseline).
A runnable bridge lives at `examples/metrics-bridge`.

## NVR integration checklist

Byte-stability is the contract — NVR-side parsers commonly match raw
SOAP local names:

- [ ] `GetStreamUriResponse → MediaUri/Uri` element order unchanged
      (golden-pinned; any change is a breaking release).
- [ ] ProbeMatches scopes advertise `onvif://www.onvif.org/name/…`,
      `hardware/…`, `location/…` — fill them in, NVRs display them.
- [ ] Stream URIs carry the advertised host, not loopback or 0.0.0.0.
- [ ] Digest auth interop: nonce + Created present, `PasswordDigest`
      type URI exact; the NVR's clock skew must stay inside your replay
      window.
- [ ] 401 (not 403) on missing credentials for protected actions —
      ONVIF clients treat 401 as the challenge signal.
- [ ] Snapshot endpoint (`/onvif/snapshot` by default) reachable if the
      NVR polls it; it is authenticated like any write action when
      `All` policy is set.

Run `cmd/onvif-diagnostics` against the deployed endpoint before
pointing the NVR at it — it exercises discovery, auth, profiles, and
stream URIs the way an NVR will.
