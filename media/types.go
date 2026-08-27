// Package media hosts the media-service (trt) domain types.
package media

import (
	"time"

	"github.com/mickeyzzc/onvif-go/v2/imaging"
	"github.com/mickeyzzc/onvif-go/v2/ptz"
	"github.com/mickeyzzc/onvif-go/v2/types"
)

// Profile represents a media profile.
type Profile struct {
	Token                     string
	Name                      string
	VideoSourceConfiguration  *VideoSourceConfiguration
	AudioSourceConfiguration  *AudioSourceConfiguration
	VideoEncoderConfiguration *VideoEncoderConfiguration
	AudioEncoderConfiguration *AudioEncoderConfiguration

	PTZConfiguration      *ptz.PTZConfiguration
	MetadataConfiguration *MetadataConfiguration
	Extension             *ProfileExtension
}

// VideoSourceConfiguration represents video source configuration.
type VideoSourceConfiguration struct {
	Token       string
	Name        string
	UseCount    int
	SourceToken string
	Bounds      *types.IntRectangle
}

// AudioSourceConfiguration represents audio source configuration.
type AudioSourceConfiguration struct {
	Token       string
	Name        string
	UseCount    int
	SourceToken string
}

// VideoEncoderConfiguration represents video encoder configuration.
type VideoEncoderConfiguration struct {
	Token          string
	Name           string
	UseCount       int
	Encoding       string // JPEG, MPEG4, H264
	Resolution     *VideoResolution
	Quality        float64
	RateControl    *VideoRateControl
	MPEG4          *MPEG4Configuration
	H264           *H264Configuration
	Multicast      *MulticastConfiguration
	SessionTimeout time.Duration
}

// AudioEncoderConfiguration represents audio encoder configuration.
type AudioEncoderConfiguration struct {
	Token          string
	Name           string
	UseCount       int
	Encoding       string // G711, G726, AAC
	Bitrate        int
	SampleRate     int
	Multicast      *MulticastConfiguration
	SessionTimeout time.Duration
}

// MetadataConfiguration represents metadata configuration.
type MetadataConfiguration struct {
	Token    string
	Name     string
	UseCount int

	PTZStatus      *ptz.PTZFilter
	Events         *EventSubscription
	Analytics      bool
	Multicast      *MulticastConfiguration
	SessionTimeout time.Duration
}

// VideoResolution represents video resolution.
type VideoResolution struct {
	Width  int
	Height int
}

// VideoRateControl represents video rate control.
type VideoRateControl struct {
	FrameRateLimit   int
	EncodingInterval int
	BitrateLimit     int
}

// MPEG4Configuration represents MPEG4 configuration.
type MPEG4Configuration struct {
	GovLength    int
	MPEG4Profile string
}

// H264Configuration represents H264 configuration.
type H264Configuration struct {
	GovLength   int
	H264Profile string
}

// MulticastConfiguration represents multicast configuration.
type MulticastConfiguration struct {
	Address   *types.IPAddress
	Port      int
	TTL       int
	AutoStart bool
}

// EventSubscription represents event subscription.
type EventSubscription struct {
	Filter *FilterType
}

// FilterType represents filter type.
type FilterType struct {
	// Simplified for now
}

// ProfileExtension represents profile extension.
type ProfileExtension struct{}

// MediaServiceCapabilities represents media service capabilities.
type MediaServiceCapabilities struct {
	SnapshotURI             bool
	Rotation                bool
	VideoSourceMode         bool
	OSD                     bool
	TemporaryOSDText        bool
	EXICompression          bool
	MaximumNumberOfProfiles int
	RTPMulticast            bool
	RTPTCP                  bool
	RTPRTSPTCP              bool
}

// VideoEncoderConfigurationOptions represents available options for video encoder configuration.
type VideoEncoderConfigurationOptions struct {
	QualityRange *types.FloatRange
	JPEG         *JPEGOptions
	H264         *H264Options
}

// JPEGOptions represents JPEG encoder options.
type JPEGOptions struct {
	ResolutionsAvailable  []*VideoResolution
	FrameRateRange        *types.FloatRange
	EncodingIntervalRange *types.IntRange
}

// H264Options represents H264 encoder options.
type H264Options struct {
	ResolutionsAvailable  []*VideoResolution
	GovLengthRange        *types.IntRange
	FrameRateRange        *types.FloatRange
	EncodingIntervalRange *types.IntRange
	H264ProfilesSupported []string
}

// VideoSourceMode represents a video source mode.
type VideoSourceMode struct {
	Token      string
	Enabled    bool
	Resolution *VideoResolution
}

// OSDConfiguration represents OSD (On-Screen Display) configuration.
type OSDConfiguration struct {
	Token string
	// Additional fields can be added based on ONVIF spec
}

// AudioEncoderConfigurationOptions represents available options for audio encoder configuration.
type AudioEncoderConfigurationOptions struct {
	EncodingOptions []string
	BitrateList     []int
	SampleRateList  []int
}

// MetadataConfigurationOptions represents available options for metadata configuration.
type MetadataConfigurationOptions struct {
	PTZStatusFilterOptions *ptz.PTZFilter
}

// AudioOutputConfiguration represents audio output configuration.
type AudioOutputConfiguration struct {
	Token       string
	Name        string
	UseCount    int
	OutputToken string
}

// AudioOutputConfigurationOptions represents available options for audio output configuration.
type AudioOutputConfigurationOptions struct {
	OutputTokensAvailable []string
}

// AudioDecoderConfigurationOptions represents available options for audio decoder configuration.
type AudioDecoderConfigurationOptions struct {
	AACDecOptions  *AudioDecoderOptions
	G711DecOptions *AudioDecoderOptions
	G726DecOptions *AudioDecoderOptions
}

// AudioDecoderOptions represents audio decoder options.
type AudioDecoderOptions struct {
	BitrateList    []int
	SampleRateList []int
}

// GuaranteedNumberOfVideoEncoderInstances represents guaranteed number of video encoder instances.
type GuaranteedNumberOfVideoEncoderInstances struct {
	TotalNumber int
	JPEG        int
	H264        int
	MPEG4       int
}

// OSDConfigurationOptions represents available options for OSD configuration.
type OSDConfigurationOptions struct {
	MaximumNumberOfOSDs int
}

// VideoSourceConfigurationOptions represents available options for video source configuration.
type VideoSourceConfigurationOptions struct {
	BoundsRange                *BoundsRange
	VideoSourceTokensAvailable []string
}

// AudioSourceConfigurationOptions represents available options for audio source configuration.
type AudioSourceConfigurationOptions struct {
	InputTokensAvailable []string
}

// BoundsRange represents bounds range for video source configuration.
type BoundsRange struct {
	X      *types.IntRange
	Y      *types.IntRange
	Width  *types.IntRange
	Height *types.IntRange
}

// AudioDecoderConfiguration represents audio decoder configuration.
type AudioDecoderConfiguration struct {
	Token    string
	Name     string
	UseCount int
}

// VideoAnalyticsConfiguration represents video analytics configuration.
type VideoAnalyticsConfiguration struct {
	Token                        string
	Name                         string
	UseCount                     int
	AnalyticsEngineConfiguration *AnalyticsEngineConfiguration
	RuleEngineConfiguration      *RuleEngineConfiguration
}

// AnalyticsEngineConfiguration represents analytics engine configuration.
type AnalyticsEngineConfiguration struct {
	AnalyticsEngine *Config
	Parameters      *ItemList
}

// RuleEngineConfiguration represents rule engine configuration.
type RuleEngineConfiguration struct {
	Rule *Config
}

// Config represents a generic configuration.
type Config struct {
	Parameters *ItemList
}

// ItemList represents a list of configuration items.
type ItemList struct {
	SimpleItem  []types.SimpleItem
	ElementItem []types.ElementItem
}

// VideoAnalyticsConfigurationOptions represents available options for video analytics configuration.
type VideoAnalyticsConfigurationOptions struct {
	// Simplified for now - can be expanded based on ONVIF spec
}

// StreamSetup represents stream setup parameters.
type StreamSetup struct {
	Stream    string // RTP-Unicast, RTP-Multicast
	Transport *Transport
}

// Transport represents transport parameters.
type Transport struct {
	Protocol string // UDP, TCP, RTSP, HTTP
	Tunnel   *Tunnel
}

// Tunnel represents tunnel parameters.
type Tunnel struct{}

// MediaURI represents a media URI.
type MediaURI struct {
	URI                 string
	InvalidAfterConnect bool
	InvalidAfterReboot  bool
	Timeout             time.Duration
}

// VideoSource represents a video source.
type VideoSource struct {
	Token      string
	Framerate  float64
	Resolution *VideoResolution
	Imaging    *imaging.ImagingSettings
}

// AudioSource represents an audio source.
type AudioSource struct {
	Token    string
	Channels int
}

// AudioOutput represents an audio output.
type AudioOutput struct {
	Token string
}
