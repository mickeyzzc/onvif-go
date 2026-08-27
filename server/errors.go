package server

import "github.com/mickeyzzc/onvif-go/v2/server/provider"

// Domain error sentinels live in the provider package (where the state
// backends produce them); these aliases keep the historical server.*
// spellings source-compatible.

var (
	// ErrVideoSourceNotFound is returned when a video source is not found.
	ErrVideoSourceNotFound = provider.ErrVideoSourceNotFound

	// ErrProfileNotFound is returned when a profile is not found.
	ErrProfileNotFound = provider.ErrProfileNotFound

	// ErrSnapshotNotSupported is returned when snapshot is not supported for a profile.
	ErrSnapshotNotSupported = provider.ErrSnapshotNotSupported

	// ErrPTZNotSupported is returned when PTZ is not supported for a profile.
	ErrPTZNotSupported = provider.ErrPTZNotSupported

	// ErrPresetNotFound is returned when a preset is not found.
	ErrPresetNotFound = provider.ErrPresetNotFound
)
