// Package simulator is the reference provider implementation: all state
// lives in memory, initialized from the profile configuration. It is the
// default set of providers for server.New; real cameras bring their own.
package simulator

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"sync"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
)

// Simulator holds the in-memory device state behind every provider
// interface. It is safe for concurrent use.
type Simulator struct {
	info     provider.DeviceInfo
	profiles []provider.ProfileConfig

	mu        sync.RWMutex
	streams   map[string]provider.StreamInfo
	ptz       map[string]*provider.PTZState
	imaging   map[string]*provider.ImagingState
	jpegCache map[string][]byte
}

// New builds the simulator state from a profile set: stream paths
// (/stream0, /stream1, ...), PTZ state for PTZ-capable profiles, and
// default imaging settings per video source.
func New(profiles []provider.ProfileConfig, info provider.DeviceInfo) *Simulator {
	sim := &Simulator{
		info:      info,
		profiles:  profiles,
		streams:   make(map[string]provider.StreamInfo),
		ptz:       make(map[string]*provider.PTZState),
		imaging:   make(map[string]*provider.ImagingState),
		jpegCache: make(map[string][]byte),
	}

	for i := range profiles {
		profile := &profiles[i]
		sim.streams[profile.Token] = provider.StreamInfo{
			RTSPPath: fmt.Sprintf("/stream%d", i),
		}

		if profile.PTZ != nil {
			sim.ptz[profile.Token] = &provider.PTZState{
				Position:   provider.PTZPosition{Pan: 0, Tilt: 0, Zoom: 0},
				LastUpdate: time.Now(),
			}
		}

		sim.imaging[profile.VideoSource.Token] = defaultImagingState()
	}

	return sim
}

// DeviceInfo implements provider.DeviceInfoProvider.
func (s *Simulator) DeviceInfo() provider.DeviceInfo {
	return s.info
}

// Profile returns the configuration of a profile token.
func (s *Simulator) Profile(token string) (provider.ProfileConfig, bool) {
	for i := range s.profiles {
		if s.profiles[i].Token == token {
			return s.profiles[i], true
		}
	}

	return provider.ProfileConfig{}, false
}

// Stream implements provider.StreamURIProvider.
func (s *Simulator) Stream(profileToken string) (provider.StreamInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, ok := s.streams[profileToken]
	if !ok {
		return provider.StreamInfo{}, fmt.Errorf("%w: %s", provider.ErrProfileNotFound, profileToken)
	}

	return info, nil
}

// SetStreamURI implements provider.StreamURISetter.
func (s *Simulator) SetStreamURI(profileToken, uri string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, ok := s.streams[profileToken]
	if !ok {
		return fmt.Errorf("%w: %s", provider.ErrProfileNotFound, profileToken)
	}

	info.OverrideURI = uri
	s.streams[profileToken] = info

	return nil
}

// Snapshot implements provider.SnapshotProvider: a solid-color JPEG at
// the configured snapshot resolution, cached per profile.
func (s *Simulator) Snapshot(profileToken string) ([]byte, error) {
	profile, ok := s.Profile(profileToken)
	if !ok {
		return nil, fmt.Errorf("%w: %s", provider.ErrProfileNotFound, profileToken)
	}

	if !profile.Snapshot.Enabled {
		return nil, fmt.Errorf("%w: %s", provider.ErrSnapshotNotSupported, profileToken)
	}

	s.mu.RLock()
	cached, ok := s.jpegCache[profileToken]
	s.mu.RUnlock()

	if ok {
		return cached, nil
	}

	width := profile.Snapshot.Resolution.Width
	height := profile.Snapshot.Resolution.Height
	if width <= 0 || height <= 0 {
		width, height = profile.VideoEncoder.Resolution.Width, profile.VideoEncoder.Resolution.Height
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// A calm slate blue, distinguishable from a black or gray decoder
	// failure frame.
	for i := range img.Pix {
		switch i % 4 {
		case 0:
			img.Pix[i] = 0x6E
		case 1:
			img.Pix[i] = 0x8B
		case 2:
			img.Pix[i] = 0xD4
		default:
			img.Pix[i] = 0xFF
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return nil, fmt.Errorf("simulator snapshot encode: %w", err)
	}

	data := buf.Bytes()

	s.mu.Lock()
	s.jpegCache[profileToken] = data
	s.mu.Unlock()

	return data, nil
}
