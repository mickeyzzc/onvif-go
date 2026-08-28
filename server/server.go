// Package server provides ONVIF server implementation for testing and simulation.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
	"github.com/mickeyzzc/onvif-go/v2/server/simulator"
	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// Option customizes a Server's state sources. Without options the
// in-memory simulator backs every provider; real devices inject their
// own implementations (no fork of the SOAP layer needed).
type Option func(*Server)

// WithDeviceInfoProvider replaces the device identity source.
func WithDeviceInfoProvider(p provider.DeviceInfoProvider) Option {
	return func(s *Server) {
		if p != nil {
			s.deviceInfo = p
		}
	}
}

// WithStreamURIProvider replaces the stream URI source.
func WithStreamURIProvider(p provider.StreamURIProvider) Option {
	return func(s *Server) {
		if p != nil {
			s.stream = p
		}
	}
}

// WithSnapshotProvider replaces the snapshot JPEG source.
func WithSnapshotProvider(p provider.SnapshotProvider) Option {
	return func(s *Server) {
		if p != nil {
			s.snapshot = p
		}
	}
}

// WithImagingProvider replaces the imaging settings backend.
func WithImagingProvider(p provider.ImagingProvider) Option {
	return func(s *Server) {
		if p != nil {
			s.imaging = p
		}
	}
}

// WithPTZProvider replaces the PTZ backend.
func WithPTZProvider(p provider.PTZProvider) Option {
	return func(s *Server) {
		if p != nil {
			s.ptz = p
		}
	}
}

// New creates a new ONVIF server with the given configuration. The
// default state backend is the in-memory simulator built from the
// profile configuration; options swap individual providers for
// hardware-backed ones.
func New(config *Config, opts ...Option) (*Server, error) {
	if config == nil {
		config = DefaultConfig()
	} else {
		// Normalize on a shallow copy so callers keep their own struct.
		cfg := *config
		config = &cfg
	}

	if config.SnapshotPath == "" {
		config.SnapshotPath = path.Join(config.BasePath, "snapshot")
	}

	sim := simulator.New(config.Profiles, config.DeviceInfo)

	server := &Server{
		config:     config,
		deviceInfo: sim,
		stream:     sim,
		snapshot:   sim,
		imaging:    sim,
		ptz:        sim,
		systemTime: time.Now(),
	}

	for _, opt := range opts {
		opt(server)
	}

	return server, nil
}

// Start starts the ONVIF server.
func (s *Server) Start(ctx context.Context) error {
	// Create HTTP server
	mux := http.NewServeMux()

	// Register service handlers
	s.registerDeviceService(mux)
	s.registerMediaService(mux)

	if s.config.SupportPTZ {
		s.registerPTZService(mux)
	}

	if s.config.SupportImaging {
		s.registerImagingService(mux)
	}

	// Add snapshot endpoint (SnapshotPath; defaults to BasePath/snapshot)
	mux.HandleFunc(s.config.SnapshotPath, s.handleSnapshot)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  s.config.Timeout,
		WriteTimeout: s.config.Timeout,
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		fmt.Printf("🎥 ONVIF Server starting on %s\n", addr)
		fmt.Printf("📡 Device Service: http://%s%s/device_service\n", addr, s.config.BasePath)
		fmt.Printf("🎬 Media Service: http://%s%s/media_service\n", addr, s.config.BasePath)
		if s.config.SupportPTZ {
			fmt.Printf("🎮 PTZ Service: http://%s%s/ptz_service\n", addr, s.config.BasePath)
		}
		if s.config.SupportImaging {
			fmt.Printf("📷 Imaging Service: http://%s%s/imaging_service\n", addr, s.config.BasePath)
		}
		fmt.Printf("\n🌐 Virtual Camera Profiles:\n")

		for i, profile := range s.config.Profiles {
			uri := ""
			if info, err := s.stream.Stream(profile.Token); err == nil {
				uri = s.deriveStreamURI(nil, info)
			}

			fmt.Printf("   [%d] %s - %s (%dx%d @ %dfps)\n",
				i+1, profile.Name, uri,
				profile.VideoEncoder.Resolution.Width,
				profile.VideoEncoder.Resolution.Height,
				profile.VideoEncoder.Framerate)
		}
		fmt.Printf("\n✅ Server is ready!\n\n")

		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		fmt.Println("\n🛑 Shutting down server...")
		const shutdownTimeout = 5 // Server shutdown timeout in seconds
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil { //nolint:contextcheck // shutdown context must outlive the cancelled parent
			return fmt.Errorf("server shutdown failed: %w", err)
		}

		return nil
	case err := <-errChan:
		return err
	}
}

// newSOAPHandler builds a SOAP handler with the server's credentials and
// response encoding settings. The default policy authenticates
// write-style actions plus SystemReboot (and any configured extras);
// read operations stay open for credential-less discovery clients.
func (s *Server) newSOAPHandler() *soap.Handler {
	policy := soap.DefaultAuthPolicy()
	policy.Actions = append([]string{"SystemReboot"}, s.config.AuthProtectedActions...)

	return soap.NewHandlerWithOptions(soap.HandlerOptions{
		Username:         s.config.Username,
		Password:         s.config.Password,
		Auth:             policy,
		ExplicitPrefixes: s.config.ExplicitPrefixes,
	})
}

// advertiseHost returns the host to publish in XAddr responses and
// stream/snapshot URIs: the configured override, else the requesting
// client's IP, else the bind address.
func (s *Server) advertiseHost(rc *soap.RequestContext) string {
	if s.config.AdvertiseHost != "" {
		return s.config.AdvertiseHost
	}

	if rc != nil && rc.RemoteIP != "" {
		return rc.RemoteIP
	}

	host := s.config.Host
	if host == defaultHost || host == "" {
		host = defaultHostname
	}

	return host
}

// deriveStreamURI renders a StreamInfo into the URI GetStreamUri
// returns: a pinned override verbatim, else
// rtsp://<advertised host>:<RTSPPort|8554><RTSPPath> (#34).
func (s *Server) deriveStreamURI(rc *soap.RequestContext, info provider.StreamInfo) string {
	if info.OverrideURI != "" {
		return info.OverrideURI
	}

	port := info.RTSPPort
	if port == 0 {
		port = defaultRTSPPort
	}

	return fmt.Sprintf("rtsp://%s:%d%s", s.advertiseHost(rc), port, info.RTSPPath)
}

// defaultSnapshotToken picks the profile served by parameterless
// snapshot requests: the first snapshot-enabled profile, else the first
// profile (whose own capture surfaces its not-supported error).
func (s *Server) defaultSnapshotToken() string {
	for i := range s.config.Profiles {
		if s.config.Profiles[i].Snapshot.Enabled {
			return s.config.Profiles[i].Token
		}
	}

	if len(s.config.Profiles) > 0 {
		return s.config.Profiles[0].Token
	}

	return ""
}

// registerDeviceService registers the device service handler.
func (s *Server) registerDeviceService(mux *http.ServeMux) {
	handler := s.newSOAPHandler()

	// Register device service handlers
	handler.RegisterContextHandler("GetDeviceInformation", s.HandleGetDeviceInformation)
	handler.RegisterContextHandler("GetCapabilities", s.HandleGetCapabilities)
	handler.RegisterContextHandler("GetSystemDateAndTime", s.HandleGetSystemDateAndTime)
	handler.RegisterContextHandler("GetServices", s.HandleGetServices)
	handler.RegisterContextHandler("GetScopes", s.HandleGetScopes)
	handler.RegisterContextHandler("SystemReboot", s.HandleSystemReboot)

	mux.Handle(s.config.BasePath+"/device_service", handler)
}

// registerMediaService registers the media service handler.
func (s *Server) registerMediaService(mux *http.ServeMux) {
	handler := s.newSOAPHandler()

	// Register media service handlers
	handler.RegisterContextHandler("GetProfiles", s.HandleGetProfiles)
	handler.RegisterContextHandler("GetStreamUri", s.HandleGetStreamUri)
	handler.RegisterContextHandler("GetSnapshotUri", s.HandleGetSnapshotUri)
	handler.RegisterContextHandler("GetVideoSources", s.HandleGetVideoSources)

	mux.Handle(s.config.BasePath+"/media_service", handler)
}

// registerPTZService registers the PTZ service handler.
func (s *Server) registerPTZService(mux *http.ServeMux) {
	handler := s.newSOAPHandler()

	// Register PTZ service handlers
	handler.RegisterContextHandler("ContinuousMove", s.HandleContinuousMove)
	handler.RegisterContextHandler("AbsoluteMove", s.HandleAbsoluteMove)
	handler.RegisterContextHandler("RelativeMove", s.HandleRelativeMove)
	handler.RegisterContextHandler("Stop", s.HandleStop)
	handler.RegisterContextHandler("GetStatus", s.HandleGetStatus)
	handler.RegisterContextHandler("GetPresets", s.HandleGetPresets)
	handler.RegisterContextHandler("GotoPreset", s.HandleGotoPreset)

	mux.Handle(s.config.BasePath+"/ptz_service", handler)
}

// registerImagingService registers the imaging service handler.
func (s *Server) registerImagingService(mux *http.ServeMux) {
	handler := s.newSOAPHandler()

	// Register imaging service handlers
	handler.RegisterContextHandler("GetImagingSettings", s.HandleGetImagingSettings)
	handler.RegisterContextHandler("SetImagingSettings", s.HandleSetImagingSettings)
	handler.RegisterContextHandler("GetOptions", s.HandleGetOptions)
	handler.RegisterContextHandler("Move", s.HandleMove)

	mux.Handle(s.config.BasePath+"/imaging_service", handler)
}

// handleSnapshot serves snapshot captures through the SnapshotProvider.
// The ?profile= parameter selects the profile; when absent the default
// (first snapshot-enabled) profile is served — real devices commonly
// expose parameterless snapshot endpoints (#36). The content type comes
// from the provider, defaulting to image/jpeg.
func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	profileToken := r.URL.Query().Get("profile")
	if profileToken == "" {
		profileToken = s.defaultSnapshotToken()
		if profileToken == "" {
			http.Error(w, "No camera profiles configured", http.StatusNotFound)

			return
		}
	}

	result, err := s.snapshot.Snapshot(profileToken)
	if err != nil {
		if errors.Is(err, ErrProfileNotFound) {
			http.Error(w, "Profile not found", http.StatusNotFound)

			return
		}

		if errors.Is(err, ErrSnapshotNotSupported) {
			http.Error(w, "Snapshot not supported", http.StatusNotImplemented)

			return
		}

		http.Error(w, "Snapshot capture failed", http.StatusInternalServerError)

		return
	}

	contentType := result.ContentType
	if contentType == "" {
		contentType = "image/jpeg"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(result.Data)))
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(result.Data)
}

// GetConfig returns the server configuration.
func (s *Server) GetConfig() *Config {
	return s.config
}

// GetStreamConfig returns the stream configuration for a profile.
func (s *Server) GetStreamConfig(profileToken string) (*StreamConfig, bool) {
	info, err := s.stream.Stream(profileToken)
	if err != nil {
		return nil, false
	}

	return &StreamConfig{
		ProfileToken: profileToken,
		RTSPPath:     info.RTSPPath,
		StreamURI:    info.OverrideURI,
	}, true
}

// UpdateStreamURI updates the RTSP URI for a profile. It requires the
// stream provider to implement provider.StreamURISetter (the simulator
// does).
func (s *Server) UpdateStreamURI(profileToken, uri string) error {
	setter, ok := s.stream.(provider.StreamURISetter)
	if !ok {
		return errors.New("stream provider does not support runtime URI updates")
	}

	return setter.SetStreamURI(profileToken, uri)
}

// ListProfiles returns all configured profiles.
func (s *Server) ListProfiles() []ProfileConfig {
	return s.config.Profiles
}

// GetPTZState returns the current PTZ state for a profile.
func (s *Server) GetPTZState(profileToken string) (*PTZState, bool) {
	state, err := s.ptz.Status(profileToken)
	if err != nil {
		return nil, false
	}

	return &state, true
}

// GetImagingState returns the current imaging state for a video source.
// It requires the imaging provider to expose raw state (the simulator
// does).
func (s *Server) GetImagingState(videoSourceToken string) (*ImagingState, bool) {
	reader, ok := s.imaging.(imagingStateReader)
	if !ok {
		return nil, false
	}

	state, ok := reader.ImagingStateOf(videoSourceToken)
	if !ok {
		return nil, false
	}

	return &state, true
}

// imagingStateReader is the optional raw-state view used by
// GetImagingState (implemented by the simulator).
type imagingStateReader interface {
	ImagingStateOf(videoSourceToken string) (ImagingState, bool)
}

// ServerInfo returns human-readable server information.
func (s *Server) ServerInfo() string {
	var info string
	info += "ONVIF Server Configuration\n"
	info += "==========================\n"
	info += fmt.Sprintf("Device: %s %s\n", s.config.DeviceInfo.Manufacturer, s.config.DeviceInfo.Model)
	info += fmt.Sprintf("Firmware: %s\n", s.config.DeviceInfo.FirmwareVersion)
	info += fmt.Sprintf("Serial: %s\n", s.config.DeviceInfo.SerialNumber)
	info += fmt.Sprintf("\nServer Address: %s:%d\n", s.config.Host, s.config.Port)
	info += fmt.Sprintf("Base Path: %s\n", s.config.BasePath)
	info += fmt.Sprintf("\nProfiles (%d):\n", len(s.config.Profiles))

	for i, profile := range s.config.Profiles {
		var sb strings.Builder
		fmt.Fprintf(&sb, "  [%d] %s (%s)\n", i+1, profile.Name, profile.Token)
		fmt.Fprintf(&sb, "      Video: %dx%d @ %dfps (%s)\n",
			profile.VideoEncoder.Resolution.Width,
			profile.VideoEncoder.Resolution.Height,
			profile.VideoEncoder.Framerate,
			profile.VideoEncoder.Encoding)
		if info, err := s.stream.Stream(profile.Token); err == nil {
			fmt.Fprintf(&sb, "      RTSP: %s\n", s.deriveStreamURI(nil, info))
		}
		if profile.PTZ != nil {
			sb.WriteString("      PTZ: Enabled\n")
			info += sb.String()
		}
	}
	info += "\nCapabilities:\n"
	info += fmt.Sprintf("  PTZ: %v\n", s.config.SupportPTZ)
	info += fmt.Sprintf("  Imaging: %v\n", s.config.SupportImaging)
	info += fmt.Sprintf("  Events: %v\n", s.config.SupportEvents)

	return info
}
