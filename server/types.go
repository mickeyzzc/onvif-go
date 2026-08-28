package server

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/onvif"
	"github.com/mickeyzzc/onvif-go/v2/server/provider"
)

// The data model (profile configuration, PTZ/imaging state, imaging
// settings/options) lives in the provider package so state backends can
// implement the provider interfaces without importing the server. These
// aliases keep the historical server.* spellings source-compatible.
type (
	DeviceInfo                    = provider.DeviceInfo
	ProfileConfig                 = provider.ProfileConfig
	VideoSourceConfig             = provider.VideoSourceConfig
	AudioSourceConfig             = provider.AudioSourceConfig
	VideoEncoderConfig            = provider.VideoEncoderConfig
	AudioEncoderConfig            = provider.AudioEncoderConfig
	PTZConfig                     = provider.PTZConfig
	SnapshotConfig                = provider.SnapshotConfig
	Resolution                    = provider.Resolution
	Bounds                        = provider.Bounds
	Range                         = provider.Range
	PTZSpeed                      = provider.PTZSpeed
	Preset                        = provider.Preset
	PTZPosition                   = provider.PTZPosition
	StreamConfig                  = provider.StreamConfig
	PTZState                      = provider.PTZState
	ImagingState                  = provider.ImagingState
	BacklightCompensation         = provider.BacklightCompensation
	ExposureSettings              = provider.ExposureSettings
	FocusSettings                 = provider.FocusSettings
	WhiteBalanceSettings          = provider.WhiteBalanceSettings
	WDRSettings                   = provider.WDRSettings
	ImagingSettings               = provider.ImagingSettings
	BacklightCompensationSettings = provider.BacklightCompensationSettings
	ExposureSettings20            = provider.ExposureSettings20
	FocusConfiguration20          = provider.FocusConfiguration20
	WideDynamicRangeSettings      = provider.WideDynamicRangeSettings
	WhiteBalanceSettings20        = provider.WhiteBalanceSettings20
	Rectangle                     = provider.Rectangle
	ImagingOptions                = provider.ImagingOptions
	BacklightCompensationOptions  = provider.BacklightCompensationOptions
	ExposureOptions               = provider.ExposureOptions
	FocusOptions                  = provider.FocusOptions
	WideDynamicRangeOptions       = provider.WideDynamicRangeOptions
	WhiteBalanceOptions           = provider.WhiteBalanceOptions
	FocusMove                     = provider.FocusMove
	AbsoluteFocus                 = provider.AbsoluteFocus
	RelativeFocus                 = provider.RelativeFocus
	ContinuousFocus               = provider.ContinuousFocus
	PTZVector                     = provider.PTZVector
	Vector2D                      = provider.Vector2D
	Vector1D                      = provider.Vector1D
	FloatRange                    = provider.FloatRange
)

const (
	defaultPort       = 8080
	defaultTimeoutSec = 30
	defaultWidth      = 1920
	defaultHeight     = 1080
	defaultFramerate  = 30
	defaultRTSPPort   = 8554
	defaultQuality    = 80
	defaultBitrate    = 4096
	maxPan            = 180
	maxTilt           = 90
	defaultPTZSpeed   = 0.5
	mediumWidth       = 1280
	mediumHeight      = 720
	mediumQuality     = 75
	highQuality       = 85
	mediumBitrate     = 2048
	lowFramerate      = 25
	highBitrate       = 6144
	maxZoom           = 3
	lowPTZSpeed       = 0.3
	presetZoom        = 2
)

// Config represents the ONVIF server configuration.
type Config struct {
	// Server settings
	Host     string        // Bind address (e.g., "0.0.0.0")
	Port     int           // Server port (default: 8080)
	BasePath string        // Base path for services (default: "/onvif")
	Timeout  time.Duration // Request timeout

	// Device information
	DeviceInfo DeviceInfo

	// Authentication. With credentials configured, the default policy
	// authenticates write-style actions (Set*/Remove*/Create*/Go* plus
	// SystemReboot and AuthProtectedActions) and leaves reads open;
	// without credentials everything is open.
	Username string
	Password string

	// AuthProtectedActions lists extra action names (beyond the default
	// write-style prefixes) that require authentication.
	AuthProtectedActions []string

	// AdvertiseHost overrides the host published in XAddr responses and
	// stream/snapshot URIs. Empty → the requesting client's IP is echoed
	// back, so each peer receives addresses reachable from its own
	// network (real-camera behavior for multi-interface hosts).
	AdvertiseHost string

	// AdvertiseHostProvider is the dynamic source for the advertised
	// host, consulted on every advertised-URL construction (#45: DHCP
	// renewals change the device IP at runtime). A non-empty return
	// value takes precedence over AdvertiseHost; nil installs nothing.
	// The function may be called concurrently and must be safe for it.
	AdvertiseHostProvider func() string

	// ExplicitPrefixes emits response envelopes with explicit namespace
	// prefixes (s:/tds:/trt:/...) instead of default xmlns declarations.
	ExplicitPrefixes bool

	// Camera profiles (supports multi-lens cameras)
	Profiles []ProfileConfig

	// Scopes answered by the Device-service GetScopes action. Hosts
	// advertising WS-Discovery should set the same values on the
	// discovery Responder config so ProbeMatches and GetScopes agree.
	// Empty → DefaultScopes (#37).
	Scopes []string

	// SnapshotPath is the HTTP path of the snapshot endpoint. Empty →
	// BasePath + "/snapshot" (the historical form); a value like "/snap"
	// is used verbatim as an absolute path (#36).
	SnapshotPath string

	// SnapshotURIParameterless makes GetSnapshotUri advertise the bare
	// SnapshotPath without the ?profile=<token> query, and the endpoint
	// serves the default profile for parameterless requests (#36). The
	// default (false) keeps the historical query form.
	SnapshotURIParameterless bool

	// Logger receives the server's startup and shutdown messages
	// (structured, log/slog). nil → nothing is logged anywhere —
	// embedded hosts keep a clean stdout/stderr (#35).
	Logger *slog.Logger

	// Capabilities
	SupportPTZ     bool
	SupportImaging bool
	SupportEvents  bool
}

// Server represents the ONVIF server: stateless SOAP handlers over
// pluggable providers. The zero-value is not usable; call New.
type Server struct {
	config     *Config
	deviceInfo provider.DeviceInfoProvider
	stream     provider.StreamURIProvider
	snapshot   provider.SnapshotProvider
	imaging    provider.ImagingProvider
	ptz        provider.PTZProvider
	systemTime time.Time

	// advertiseFn is the effective dynamic host source (seeded from
	// Config.AdvertiseHostProvider, replaceable at runtime); guarded by
	// advertiseMu. nil → the static/requester resolution applies.
	advertiseMu sync.RWMutex
	advertiseFn func() string
}

// DefaultConfig returns a default server configuration with a multi-lens camera setup.
//
//nolint:funlen // DefaultConfig has many statements due to comprehensive default configuration
func DefaultConfig() *Config {
	return &Config{
		Host:     "0.0.0.0",
		Port:     defaultPort,
		BasePath: "/onvif",
		Timeout:  defaultTimeoutSec * time.Second,
		DeviceInfo: DeviceInfo{
			Manufacturer:    "onvif-go",
			Model:           "Virtual Multi-Lens Camera",
			FirmwareVersion: "1.0.0",
			SerialNumber:    "SN-12345678",
			HardwareID:      "HW-87654321",
		},
		Username:       "admin",
		Password:       "admin",
		SupportPTZ:     true,
		SupportImaging: true,
		SupportEvents:  false,
		Profiles: []ProfileConfig{
			{
				Token: "profile_0",
				Name:  "Main Camera - High Quality",
				VideoSource: VideoSourceConfig{
					Token:      "video_source_0",
					Name:       "Main Camera",
					Resolution: Resolution{Width: defaultWidth, Height: defaultHeight},
					Framerate:  defaultFramerate,
					Bounds:     Bounds{X: 0, Y: 0, Width: defaultWidth, Height: defaultHeight},
				},
				VideoEncoder: VideoEncoderConfig{
					Encoding:   "H264",
					Resolution: Resolution{Width: defaultWidth, Height: defaultHeight},
					Quality:    defaultQuality,
					Framerate:  defaultFramerate,
					Bitrate:    defaultBitrate,
					GovLength:  defaultFramerate,
				},
				PTZ: &PTZConfig{
					NodeToken: "ptz_node_0",
					PanRange:  Range{Min: -maxPan, Max: maxPan},
					TiltRange: Range{Min: -maxTilt, Max: maxTilt},
					ZoomRange: Range{Min: 0, Max: 1},
					DefaultSpeed: PTZSpeed{
						Pan: defaultPTZSpeed, Tilt: defaultPTZSpeed, Zoom: defaultPTZSpeed,
					},
					SupportsContinuous: true,
					SupportsAbsolute:   true,
					SupportsRelative:   true,
					Presets: []Preset{
						{Token: "preset_0", Name: "Home", Position: PTZPosition{Pan: 0, Tilt: 0, Zoom: 0}},
						{
							Token: "preset_1", Name: "Entrance",
							Position: PTZPosition{Pan: -45, Tilt: -10, Zoom: defaultPTZSpeed},
						},
					},
				},
				Snapshot: SnapshotConfig{
					Enabled:    true,
					Resolution: Resolution{Width: defaultWidth, Height: defaultHeight},
					Quality:    highQuality,
				},
			},
			{
				Token: "profile_1",
				Name:  "Wide Angle Camera",
				VideoSource: VideoSourceConfig{
					Token:      "video_source_1",
					Name:       "Wide Angle Camera",
					Resolution: Resolution{Width: mediumWidth, Height: mediumHeight},
					Framerate:  defaultFramerate,
					Bounds:     Bounds{X: 0, Y: 0, Width: mediumWidth, Height: mediumHeight},
				},
				VideoEncoder: VideoEncoderConfig{
					Encoding:   "H264",
					Resolution: Resolution{Width: mediumWidth, Height: mediumHeight},
					Quality:    mediumQuality,
					Framerate:  defaultFramerate,
					Bitrate:    mediumBitrate,
					GovLength:  defaultFramerate,
				},
				Snapshot: SnapshotConfig{
					Enabled:    true,
					Resolution: Resolution{Width: mediumWidth, Height: mediumHeight},
					Quality:    defaultQuality,
				},
			},
			{
				Token: "profile_2",
				Name:  "Telephoto Camera",
				VideoSource: VideoSourceConfig{
					Token:      "video_source_2",
					Name:       "Telephoto Camera",
					Resolution: Resolution{Width: defaultWidth, Height: defaultHeight},
					Framerate:  lowFramerate,
					Bounds:     Bounds{X: 0, Y: 0, Width: defaultWidth, Height: defaultHeight},
				},
				VideoEncoder: VideoEncoderConfig{
					Encoding:   "H264",
					Resolution: Resolution{Width: defaultWidth, Height: defaultHeight},
					Quality:    highQuality,
					Framerate:  lowFramerate,
					Bitrate:    highBitrate,
					GovLength:  lowFramerate,
				},
				PTZ: &PTZConfig{
					NodeToken: "ptz_node_2",
					PanRange:  Range{Min: -maxPan, Max: maxPan},
					TiltRange: Range{Min: -maxTilt, Max: maxTilt},
					ZoomRange: Range{Min: 0, Max: maxZoom},
					DefaultSpeed: PTZSpeed{
						Pan: lowPTZSpeed, Tilt: lowPTZSpeed, Zoom: lowPTZSpeed,
					},
					SupportsContinuous: true,
					SupportsAbsolute:   true,
					SupportsRelative:   true,
					Presets: []Preset{
						{Token: "preset_2_0", Name: "Home", Position: PTZPosition{Pan: 0, Tilt: 0, Zoom: 0}},
						{
							Token: "preset_2_1", Name: "Zoom In",
							Position: PTZPosition{Pan: 0, Tilt: 0, Zoom: presetZoom},
						},
					},
				},
				Snapshot: SnapshotConfig{
					Enabled:    true,
					Resolution: Resolution{Width: defaultWidth, Height: defaultHeight},
					Quality:    highQuality,
				},
			},
		},
	}
}

// ServiceEndpoints returns the service endpoint URLs.
func (c *Config) ServiceEndpoints(host string) map[string]string {
	if host == "" {
		host = c.Host
		if host == "0.0.0.0" || host == "" {
			host = "localhost"
		}
	}

	var baseURL string
	const httpPort = 80
	if c.Port == httpPort {
		baseURL = "http://" + host + c.BasePath
	} else {
		// Import fmt at the top to use Sprintf
		baseURL = fmt.Sprintf("http://%s:%d%s", host, c.Port, c.BasePath)
	}

	endpoints := map[string]string{
		"device":  baseURL + "/device_service",
		"media":   baseURL + "/media_service",
		"imaging": baseURL + "/imaging_service",
	}

	if c.SupportPTZ {
		endpoints["ptz"] = baseURL + "/ptz_service"
	}

	if c.SupportEvents {
		endpoints["events"] = baseURL + "/events_service"
	}

	return endpoints
}

// ProfileToONVIF converts a ProfileConfig to an ONVIF Profile (the
// client-side wire shape). Formerly a ProfileConfig method; methods
// cannot follow type aliases across packages.
func ProfileToONVIF(p *ProfileConfig) *onvif.Profile {
	profile := &onvif.Profile{
		Token: p.Token,
		Name:  p.Name,
		VideoSourceConfiguration: &onvif.VideoSourceConfiguration{
			Token:       p.VideoSource.Token,
			Name:        p.VideoSource.Name,
			SourceToken: p.VideoSource.Token,
			Bounds: &onvif.IntRectangle{
				X:      p.VideoSource.Bounds.X,
				Y:      p.VideoSource.Bounds.Y,
				Width:  p.VideoSource.Bounds.Width,
				Height: p.VideoSource.Bounds.Height,
			},
		},
		VideoEncoderConfiguration: &onvif.VideoEncoderConfiguration{
			Token:    p.Token + "_encoder",
			Name:     p.Name + " Encoder",
			Encoding: p.VideoEncoder.Encoding,
			Resolution: &onvif.VideoResolution{
				Width:  p.VideoEncoder.Resolution.Width,
				Height: p.VideoEncoder.Resolution.Height,
			},
			Quality: p.VideoEncoder.Quality,
			RateControl: &onvif.VideoRateControl{
				FrameRateLimit: p.VideoEncoder.Framerate,
				BitrateLimit:   p.VideoEncoder.Bitrate,
			},
		},
	}

	if p.PTZ != nil {
		profile.PTZConfiguration = &onvif.PTZConfiguration{
			Token:     p.PTZ.NodeToken,
			Name:      p.Name + " PTZ",
			NodeToken: p.PTZ.NodeToken,
		}
	}

	return profile
}
