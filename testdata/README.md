# Test Data for ONVIF Camera Testing

This directory contains **synthetic** camera data for testing the onvif-go
library. No real network data (live LAN addresses, real device inventories,
raw captures) is committed to this repository.

## Files

### discovered_cameras_synthetic.json
JSON file with fabricated camera entries:
- RFC 5737 documentation IPs only (`192.0.2.x`, `198.51.100.x`)
- Synthetic UUIDs (`a0000000-0000-4000-8000-*`)
- Generic vendor/model names

### test_cameras_config.go
Go package providing programmatic access to the same synthetic data (the go
tool ignores this directory — it is data, not compiled code):
- `TestCameras` slice
- `GetCameraByManufacturer()` / `GetCameraByProfile()` / `GetHTTPSCameras()`

### captures/
Hand-written XML fixtures used by unit tests (GetStreamUri/GetSnapshotUri
quirk matrix). These are minimal synthetic envelopes, not device captures.

## Policy

If you capture traffic from real cameras for debugging, keep it in `tmp/`
(gitignored) — never commit it. Regenerate synthetic fixtures from the
quirk being tested instead.

## See Also

- [Main Testing Documentation](../docs/testing/)
