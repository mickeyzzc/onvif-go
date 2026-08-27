// Package server provides ONVIF server implementation for testing and simulation.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// New creates a new ONVIF server with the given configuration.
func New(config *Config) (*Server, error) {
	if config == nil {
		config = DefaultConfig()
	}

	server := &Server{
		config:       config,
		streams:      make(map[string]*StreamConfig),
		ptzState:     make(map[string]*PTZState),
		imagingState: make(map[string]*ImagingState),
		systemTime:   time.Now(),
	}

	// Initialize streams for each profile. StreamURI stays empty so
	// GetStreamUri derives it per request from the advertised host (the
	// requesting client's IP by default); UpdateStreamURI pins an explicit
	// override that then wins.
	for i := range config.Profiles {
		profile := &config.Profiles[i]
		streamPath := fmt.Sprintf("/stream%d", i)

		server.streams[profile.Token] = &StreamConfig{
			ProfileToken: profile.Token,
			RTSPPath:     streamPath,
		}

		// Initialize PTZ state if PTZ is supported
		if profile.PTZ != nil {
			server.ptzState[profile.Token] = &PTZState{
				Position:   PTZPosition{Pan: 0, Tilt: 0, Zoom: 0},
				Moving:     false,
				PanMoving:  false,
				TiltMoving: false,
				ZoomMoving: false,
				LastUpdate: time.Now(),
			}
		}

		// Initialize imaging state
		server.imagingState[profile.VideoSource.Token] = &ImagingState{
			Brightness:  50.0, //nolint:mnd // Default imaging value
			Contrast:    50.0, //nolint:mnd // Default imaging value
			Saturation:  50.0, //nolint:mnd // Default imaging value
			Sharpness:   50.0, //nolint:mnd // Default imaging value
			IrCutFilter: "AUTO",
			BacklightComp: BacklightCompensation{
				Mode:  "OFF",
				Level: 0,
			},
			Exposure: ExposureSettings{
				Mode:         "AUTO",
				Priority:     "FrameRate",
				MinExposure:  1,
				MaxExposure:  10000, //nolint:mnd // Exposure time in microseconds
				MinGain:      0,
				MaxGain:      100, //nolint:mnd // Gain value
				ExposureTime: 100, //nolint:mnd // Exposure time
				Gain:         50,  //nolint:mnd // Gain value
			},
			Focus: FocusSettings{
				AutoFocusMode: "AUTO",
				DefaultSpeed:  0.5, //nolint:mnd // Focus speed
				NearLimit:     0,
				FarLimit:      1,
				CurrentPos:    0.5, //nolint:mnd // Focus position
			},
			WhiteBalance: WhiteBalanceSettings{
				Mode:   "AUTO",
				CrGain: 128, //nolint:mnd // White balance gain
				CbGain: 128, //nolint:mnd // White balance gain
			},
			WideDynamicRange: WDRSettings{
				Mode:  "OFF",
				Level: 0,
			},
		}
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

	// Add snapshot endpoint
	mux.HandleFunc(s.config.BasePath+"/snapshot", s.handleSnapshot)

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
			stream := s.streams[profile.Token]
			fmt.Printf("   [%d] %s - %s (%dx%d @ %dfps)\n",
				i+1, profile.Name, stream.StreamURI,
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

// registerDeviceService registers the device service handler.
func (s *Server) registerDeviceService(mux *http.ServeMux) {
	handler := s.newSOAPHandler()

	// Register device service handlers
	handler.RegisterContextHandler("GetDeviceInformation", s.HandleGetDeviceInformation)
	handler.RegisterContextHandler("GetCapabilities", s.HandleGetCapabilities)
	handler.RegisterContextHandler("GetSystemDateAndTime", s.HandleGetSystemDateAndTime)
	handler.RegisterContextHandler("GetServices", s.HandleGetServices)
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

// handleSnapshot handles HTTP snapshot requests.
func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	// Get profile token from query parameter
	profileToken := r.URL.Query().Get("profile")
	if profileToken == "" {
		http.Error(w, "Missing profile parameter", http.StatusBadRequest)

		return
	}

	// Find the profile
	var profileCfg *ProfileConfig
	for i := range s.config.Profiles {
		if s.config.Profiles[i].Token == profileToken {
			profileCfg = &s.config.Profiles[i]

			break
		}
	}

	if profileCfg == nil {
		http.Error(w, "Profile not found", http.StatusNotFound)

		return
	}

	if !profileCfg.Snapshot.Enabled {
		http.Error(w, "Snapshot not supported", http.StatusNotImplemented)

		return
	}

	// In a real implementation, this would capture a frame from the video source
	// For now, return a placeholder response
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", "0")
	w.WriteHeader(http.StatusOK)

	// TODO: Generate or capture actual JPEG snapshot
}

// GetConfig returns the server configuration.
func (s *Server) GetConfig() *Config {
	return s.config
}

// GetStreamConfig returns the stream configuration for a profile.
func (s *Server) GetStreamConfig(profileToken string) (*StreamConfig, bool) {
	stream, ok := s.streams[profileToken]

	return stream, ok
}

// UpdateStreamURI updates the RTSP URI for a profile.
func (s *Server) UpdateStreamURI(profileToken, uri string) error {
	stream, ok := s.streams[profileToken]
	if !ok {
		return fmt.Errorf("%w: %s", ErrProfileNotFound, profileToken)
	}
	stream.StreamURI = uri

	return nil
}

// ListProfiles returns all configured profiles.
func (s *Server) ListProfiles() []ProfileConfig {
	return s.config.Profiles
}

// GetPTZState returns the current PTZ state for a profile.
func (s *Server) GetPTZState(profileToken string) (*PTZState, bool) {
	ptzMutex.RLock()
	defer ptzMutex.RUnlock()
	state, ok := s.ptzState[profileToken]

	return state, ok
}

// GetImagingState returns the current imaging state for a video source.
func (s *Server) GetImagingState(videoSourceToken string) (*ImagingState, bool) {
	imagingMutex.RLock()
	defer imagingMutex.RUnlock()
	state, ok := s.imagingState[videoSourceToken]

	return state, ok
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
		if stream, ok := s.streams[profile.Token]; ok {
			uri := stream.StreamURI
			if uri == "" {
				uri = fmt.Sprintf("rtsp://%s:8554%s", s.advertiseHost(nil), stream.RTSPPath)
			}

			fmt.Fprintf(&sb, "      RTSP: %s\n", uri)
		}
		if profile.PTZ != nil {
			sb.WriteString("      PTZ: Enabled\n")
			info += sb.String() //nolint:perfsprint // bounded loop over configured profiles
		}
	}
	info += "\nCapabilities:\n"
	info += fmt.Sprintf("  PTZ: %v\n", s.config.SupportPTZ)
	info += fmt.Sprintf("  Imaging: %v\n", s.config.SupportImaging)
	info += fmt.Sprintf("  Events: %v\n", s.config.SupportEvents)

	return info
}
