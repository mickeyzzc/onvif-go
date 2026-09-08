# Security Policy

## Supported versions

Only the latest tagged release of the `v2` module and the current `main`
branch receive security fixes. Pre-`v2` versions (`mickeyzzc/onvif-go`
without `/v2`) are end-of-life; migrate first (see `MIGRATION.md`).

| Version | Supported |
|---------|-----------|
| `v2` latest tag | ✅ |
| `main` | ✅ (fixes land here first, PR-only) |
| `v1.x` | ❌ end-of-life |

## Reporting a vulnerability

**Please do not open a public issue for security problems.**

- Prefer a private [GitHub security advisory](https://github.com/mickeyzzc/onvif-go/security/advisories/new).
- Alternatively email the maintainer (see the GitHub profile); include
  `onvif-go security` in the subject.

Include reproduction details (packet capture, SOAP body, fuzz input) when
possible. You will get an acknowledgement within 7 days. Fixes are released
as patch versions out-of-band if urgent; otherwise they ship with the next
release (merge ≠ release — see `CONTRIBUTING.md`).

## Scope

Security-relevant surfaces maintained by this library:

- SOAP/XML parsing of **untrusted** device responses (client) and client
  requests (server): bounded reads (`MaxBodyBytes`, 1 MiB client cap),
  per-source auth-failure lockout, panic-free fuzz targets.
- WS-UsernameToken verification (digest + plaintext) and the per-action
  auth policy.
- WS-Discovery datagram parsing (multicast, untrusted by definition).
- RTP/RTSP/GB28181-adjacent transport is **out of scope** — this library
  speaks SOAP/WS-Discovery only.

Out of scope: consumers' TLS termination choices, credential storage,
HTTP server hardening beyond the library's own limits (those belong to the
embedding application).

## Safe harbor

Fuzzing and penetration testing against your own deployments, and
submitting crashers found by `go test -fuzz`, are welcome — please still
report anything that survives the in-repo fuzz targets privately first.
