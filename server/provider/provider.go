package provider

// Providers are the pluggable state sources behind the ONVIF server's
// SOAP layer. The reference implementation lives in server/simulator
// (in-memory state built from Config); a real camera implements these
// against its hardware and injects them via server.New options — no
// fork of the SOAP layer required.

// DeviceInfoProvider supplies the identity block returned by
// GetDeviceInformation.
type DeviceInfoProvider interface {
	DeviceInfo() DeviceInfo
}

// StreamInfo describes where a profile's RTSP stream lives.
type StreamInfo struct {
	// RTSPPath is the path suffix used to derive a URI from the
	// advertised host when OverrideURI is empty (rtsp://<host>:<port><path>).
	RTSPPath string

	// RTSPPort is the RTSP listener port used when deriving the URI.
	// 0 → the ONVIF-conventional 8554 default (#34: real devices expose
	// configurable RTSP ports).
	RTSPPort int

	// OverrideURI, when non-empty, is returned verbatim by GetStreamUri
	// (a pinned external address; the advertised host is not applied).
	OverrideURI string
}

// StreamURIProvider answers GetStreamUri lookups per profile token.
type StreamURIProvider interface {
	Stream(profileToken string) (StreamInfo, error)
}

// StreamURISetter is the optional write side for hosts that allow
// pinning stream URIs at runtime.
type StreamURISetter interface {
	SetStreamURI(profileToken, uri string) error
}

// SnapshotResult is a captured snapshot: raw bytes plus their content
// type. Devices with fallback chains may serve non-JPEG results (e.g. a
// cached H.264 IDR frame); an empty ContentType means image/jpeg.
type SnapshotResult struct {
	Data        []byte
	ContentType string
}

// SnapshotProvider supplies snapshot bytes for the HTTP snapshot
// endpoint and GetSnapshotUri consumers. Implementations return
// ErrSnapshotNotSupported when the profile cannot snapshot.
type SnapshotProvider interface {
	Snapshot(profileToken string) (SnapshotResult, error)
}

// ImagingProvider is the read/write backend behind the Imaging service
// (GetImagingSettings / SetImagingSettings / GetOptions / Move).
type ImagingProvider interface {
	ImagingSettings(videoSourceToken string) (*ImagingSettings, error)
	SetImagingSettings(videoSourceToken string, settings *ImagingSettings) error
	ImagingOptions(videoSourceToken string) (*ImagingOptions, error)
	MoveFocus(videoSourceToken string, focus *FocusMove) error
}

// PTZProvider drives PTZ state. The simulator keeps positions in memory
// with fake completion timers; real devices wrap their motor controls.
// Presets stay part of the profile configuration (PTZConfig.Presets).
type PTZProvider interface {
	ContinuousMove(profileToken string, velocity PTZVector, timeout string) error
	AbsoluteMove(profileToken string, position PTZVector) error
	RelativeMove(profileToken string, translation PTZVector) error
	Stop(profileToken string, panTilt, zoom bool) error
	Status(profileToken string) (PTZState, error)
	GotoPreset(profileToken, presetToken string) error
}
