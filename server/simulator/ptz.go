package simulator

import (
	"fmt"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
)

// PTZ movement completion delay — the simulator flips the moving flags
// back off after this duration, mimicking arrival at the target.
const ptzMoveDuration = 500 * time.Millisecond

// ContinuousMove implements provider.PTZProvider.
func (s *Simulator) ContinuousMove(profileToken string, velocity provider.PTZVector, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.ptz[profileToken]
	if !ok {
		return fmt.Errorf("%w: %s", provider.ErrPTZNotSupported, profileToken)
	}

	state.Moving = true
	if velocity.PanTilt != nil {
		state.PanMoving = velocity.PanTilt.X != 0 || velocity.PanTilt.Y != 0
		state.TiltMoving = state.PanMoving
	}

	if velocity.Zoom != nil {
		state.ZoomMoving = velocity.Zoom.X != 0
	}

	state.LastUpdate = time.Now()

	return nil
}

// AbsoluteMove implements provider.PTZProvider.
func (s *Simulator) AbsoluteMove(profileToken string, position provider.PTZVector) error {
	s.mu.Lock()

	state, ok := s.ptz[profileToken]
	if !ok {
		s.mu.Unlock()

		return fmt.Errorf("%w: %s", provider.ErrPTZNotSupported, profileToken)
	}

	if position.PanTilt != nil {
		state.Position.Pan = position.PanTilt.X
		state.Position.Tilt = position.PanTilt.Y
	}

	if position.Zoom != nil {
		state.Position.Zoom = position.Zoom.X
	}

	state.Moving = true
	state.PanMoving = position.PanTilt != nil
	state.TiltMoving = position.PanTilt != nil
	state.ZoomMoving = position.Zoom != nil
	state.LastUpdate = time.Now()
	s.mu.Unlock()

	s.scheduleStop(state, ptzMoveDuration)

	return nil
}

// RelativeMove implements provider.PTZProvider.
func (s *Simulator) RelativeMove(profileToken string, translation provider.PTZVector) error {
	s.mu.Lock()

	state, ok := s.ptz[profileToken]
	if !ok {
		s.mu.Unlock()

		return fmt.Errorf("%w: %s", provider.ErrPTZNotSupported, profileToken)
	}

	if translation.PanTilt != nil {
		state.Position.Pan += translation.PanTilt.X
		state.Position.Tilt += translation.PanTilt.Y
	}

	if translation.Zoom != nil {
		state.Position.Zoom += translation.Zoom.X
	}

	const maxPan = 180 // PTZ pan range
	const maxTilt = 90 // PTZ tilt range
	state.Position.Pan = clamp(state.Position.Pan, -maxPan, maxPan)
	state.Position.Tilt = clamp(state.Position.Tilt, -maxTilt, maxTilt)
	state.Position.Zoom = clamp(state.Position.Zoom, 0, 1)

	state.Moving = true
	state.LastUpdate = time.Now()
	s.mu.Unlock()

	s.scheduleStop(state, ptzMoveDuration)

	return nil
}

// Stop implements provider.PTZProvider.
func (s *Simulator) Stop(profileToken string, panTilt, zoom bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.ptz[profileToken]
	if !ok {
		return fmt.Errorf("%w: %s", provider.ErrPTZNotSupported, profileToken)
	}

	if panTilt {
		state.PanMoving = false
		state.TiltMoving = false
	}

	if zoom {
		state.ZoomMoving = false
	}

	if !panTilt && !zoom {
		state.PanMoving = false
		state.TiltMoving = false
		state.ZoomMoving = false
	}

	state.Moving = state.PanMoving || state.TiltMoving || state.ZoomMoving
	state.LastUpdate = time.Now()

	return nil
}

// Status implements provider.PTZProvider.
func (s *Simulator) Status(profileToken string) (provider.PTZState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.ptz[profileToken]
	if !ok {
		return provider.PTZState{}, fmt.Errorf("%w: %s", provider.ErrPTZNotSupported, profileToken)
	}

	return *state, nil
}

// GotoPreset implements provider.PTZProvider. The preset position is
// resolved from the profile configuration (kept by the server) — the
// simulator accepts the resolved position via MoveToPreset.
func (s *Simulator) GotoPreset(profileToken, presetToken string) error {
	profile, ok := s.Profile(profileToken)
	if !ok || profile.PTZ == nil {
		return fmt.Errorf("%w: %s", provider.ErrPTZNotSupported, profileToken)
	}

	var presetPos *provider.PTZPosition
	for i := range profile.PTZ.Presets {
		if profile.PTZ.Presets[i].Token == presetToken {
			presetPos = &profile.PTZ.Presets[i].Position

			break
		}
	}

	if presetPos == nil {
		return fmt.Errorf("%w: %s", provider.ErrPresetNotFound, presetToken)
	}

	return s.MoveToPreset(profileToken, *presetPos)
}

// MoveToPreset snaps the PTZ state to a resolved preset position and
// reports movement for a moment afterwards.
func (s *Simulator) MoveToPreset(profileToken string, pos provider.PTZPosition) error {
	s.mu.Lock()

	state, ok := s.ptz[profileToken]
	if !ok {
		s.mu.Unlock()

		return fmt.Errorf("%w: %s", provider.ErrPTZNotSupported, profileToken)
	}

	state.Position = pos
	state.Moving = true
	state.PanMoving = true
	state.TiltMoving = true
	state.ZoomMoving = true
	state.LastUpdate = time.Now()
	s.mu.Unlock()

	s.scheduleStop(state, time.Second)

	return nil
}

// scheduleStop flips the moving flags off after d, mirroring arrival.
func (s *Simulator) scheduleStop(state *provider.PTZState, d time.Duration) {
	time.AfterFunc(d, func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		state.Moving = false
		state.PanMoving = false
		state.TiltMoving = false
		state.ZoomMoving = false
	})
}

func clamp(value, minVal, maxVal float64) float64 {
	if value < minVal {
		return minVal
	}

	if value > maxVal {
		return maxVal
	}

	return value
}
