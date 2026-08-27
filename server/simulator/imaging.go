package simulator

import (
	"fmt"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
)

// defaultImagingState builds the neutral imaging defaults the simulator
// boots with (mid-range values, auto modes).
//
//nolint:mnd // Imaging default values
func defaultImagingState() *provider.ImagingState {
	return &provider.ImagingState{
		Brightness:  50.0,
		Contrast:    50.0,
		Saturation:  50.0,
		Sharpness:   50.0,
		IrCutFilter: "AUTO",
		BacklightComp: provider.BacklightCompensation{
			Mode:  "OFF",
			Level: 0,
		},
		Exposure: provider.ExposureSettings{
			Mode:         "AUTO",
			Priority:     "FrameRate",
			MinExposure:  1,
			MaxExposure:  10000,
			MinGain:      0,
			MaxGain:      100,
			ExposureTime: 100,
			Gain:         50,
		},
		Focus: provider.FocusSettings{
			AutoFocusMode: "AUTO",
			DefaultSpeed:  0.5,
			NearLimit:     0,
			FarLimit:      1,
			CurrentPos:    0.5,
		},
		WhiteBalance: provider.WhiteBalanceSettings{
			Mode:   "AUTO",
			CrGain: 128,
			CbGain: 128,
		},
		WideDynamicRange: provider.WDRSettings{
			Mode:  "OFF",
			Level: 0,
		},
	}
}

// ImagingSettings implements provider.ImagingProvider.
func (s *Simulator) ImagingSettings(videoSourceToken string) (*provider.ImagingSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.imaging[videoSourceToken]
	if !ok {
		return nil, fmt.Errorf("%w: %s", provider.ErrVideoSourceNotFound, videoSourceToken)
	}

	return stateToSettings(state), nil
}

// SetImagingSettings implements provider.ImagingProvider.
//
//nolint:gocyclo // Mirrors the ONVIF field-by-field partial-update semantics
func (s *Simulator) SetImagingSettings(videoSourceToken string, settings *provider.ImagingSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.imaging[videoSourceToken]
	if !ok {
		return fmt.Errorf("%w: %s", provider.ErrVideoSourceNotFound, videoSourceToken)
	}

	if settings == nil {
		return nil
	}

	if settings.Brightness != nil {
		state.Brightness = *settings.Brightness
	}

	if settings.ColorSaturation != nil {
		state.Saturation = *settings.ColorSaturation
	}

	if settings.Contrast != nil {
		state.Contrast = *settings.Contrast
	}

	if settings.Sharpness != nil {
		state.Sharpness = *settings.Sharpness
	}

	if settings.IrCutFilter != nil {
		state.IrCutFilter = *settings.IrCutFilter
	}

	if settings.BacklightCompensation != nil {
		state.BacklightComp.Mode = settings.BacklightCompensation.Mode
		if settings.BacklightCompensation.Level != nil {
			state.BacklightComp.Level = *settings.BacklightCompensation.Level
		}
	}

	if settings.Exposure != nil {
		state.Exposure.Mode = settings.Exposure.Mode
		if settings.Exposure.Priority != nil {
			state.Exposure.Priority = *settings.Exposure.Priority
		}
		if settings.Exposure.ExposureTime != nil {
			state.Exposure.ExposureTime = *settings.Exposure.ExposureTime
		}
		if settings.Exposure.Gain != nil {
			state.Exposure.Gain = *settings.Exposure.Gain
		}
	}

	if settings.Focus != nil {
		state.Focus.AutoFocusMode = settings.Focus.AutoFocusMode
	}

	if settings.WideDynamicRange != nil {
		state.WideDynamicRange.Mode = settings.WideDynamicRange.Mode
		if settings.WideDynamicRange.Level != nil {
			state.WideDynamicRange.Level = *settings.WideDynamicRange.Level
		}
	}

	if settings.WhiteBalance != nil {
		state.WhiteBalance.Mode = settings.WhiteBalance.Mode
		if settings.WhiteBalance.CrGain != nil {
			state.WhiteBalance.CrGain = *settings.WhiteBalance.CrGain
		}
		if settings.WhiteBalance.CbGain != nil {
			state.WhiteBalance.CbGain = *settings.WhiteBalance.CbGain
		}
	}

	return nil
}

// ImagingOptions implements provider.ImagingProvider: the static
// option ranges the simulator's defaults imply.
//
//nolint:mnd // Imaging option ranges
func (s *Simulator) ImagingOptions(videoSourceToken string) (*provider.ImagingOptions, error) {
	if _, ok := s.imaging[videoSourceToken]; !ok {
		return nil, fmt.Errorf("%w: %s", provider.ErrVideoSourceNotFound, videoSourceToken)
	}

	const maxImagingValue = 100
	const maxExposureTime = 10000

	return &provider.ImagingOptions{
		Brightness:       &provider.FloatRange{Min: 0, Max: maxImagingValue},
		ColorSaturation:  &provider.FloatRange{Min: 0, Max: maxImagingValue},
		Contrast:         &provider.FloatRange{Min: 0, Max: maxImagingValue},
		Sharpness:        &provider.FloatRange{Min: 0, Max: maxImagingValue},
		IrCutFilterModes: []string{"ON", "OFF", "AUTO"},
		BacklightCompensation: &provider.BacklightCompensationOptions{
			Mode:  []string{"OFF", "ON"},
			Level: &provider.FloatRange{Min: 0, Max: maxImagingValue},
		},
		Exposure: &provider.ExposureOptions{
			Mode:            []string{"AUTO", "MANUAL"},
			Priority:        []string{"LowNoise", "FrameRate"},
			MinExposureTime: &provider.FloatRange{Min: 1, Max: maxExposureTime},
			MaxExposureTime: &provider.FloatRange{Min: 1, Max: maxExposureTime},
			MinGain:         &provider.FloatRange{Min: 0, Max: maxImagingValue},
			MaxGain:         &provider.FloatRange{Min: 0, Max: maxImagingValue},
			ExposureTime:    &provider.FloatRange{Min: 1, Max: maxExposureTime},
			Gain:            &provider.FloatRange{Min: 0, Max: maxImagingValue},
		},
		Focus: &provider.FocusOptions{
			AutoFocusModes: []string{"AUTO", "MANUAL"},
			DefaultSpeed:   &provider.FloatRange{Min: 0, Max: 1},
			NearLimit:      &provider.FloatRange{Min: 0, Max: 1},
			FarLimit:       &provider.FloatRange{Min: 0, Max: 1},
		},
		WideDynamicRange: &provider.WideDynamicRangeOptions{
			Mode:  []string{"OFF", "ON"},
			Level: &provider.FloatRange{Min: 0, Max: 100},
		},
		WhiteBalance: &provider.WhiteBalanceOptions{
			Mode:   []string{"AUTO", "MANUAL"},
			YrGain: &provider.FloatRange{Min: 0, Max: 255},
			YbGain: &provider.FloatRange{Min: 0, Max: 255},
		},
	}, nil
}

// MoveFocus implements provider.ImagingProvider (absolute/relative focus).
func (s *Simulator) MoveFocus(videoSourceToken string, focus *provider.FocusMove) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.imaging[videoSourceToken]
	if !ok {
		return fmt.Errorf("%w: %s", provider.ErrVideoSourceNotFound, videoSourceToken)
	}

	if focus == nil {
		return nil
	}

	if focus.Absolute != nil {
		state.Focus.CurrentPos = focus.Absolute.Position
	} else if focus.Relative != nil {
		state.Focus.CurrentPos += focus.Relative.Distance
		state.Focus.CurrentPos = clamp(state.Focus.CurrentPos, 0, 1)
	}

	return nil
}

// stateToSettings projects the simulator's flat state into the ONVIF
// ImagingSettings wire shape (pointer fields = present entries).
func stateToSettings(state *provider.ImagingState) *provider.ImagingSettings {
	return &provider.ImagingSettings{
		Brightness:      &state.Brightness,
		ColorSaturation: &state.Saturation,
		Contrast:        &state.Contrast,
		Sharpness:       &state.Sharpness,
		IrCutFilter:     &state.IrCutFilter,
		BacklightCompensation: &provider.BacklightCompensationSettings{
			Mode:  state.BacklightComp.Mode,
			Level: &state.BacklightComp.Level,
		},
		Exposure: &provider.ExposureSettings20{
			Mode:            state.Exposure.Mode,
			Priority:        &state.Exposure.Priority,
			MinExposureTime: &state.Exposure.MinExposure,
			MaxExposureTime: &state.Exposure.MaxExposure,
			MinGain:         &state.Exposure.MinGain,
			MaxGain:         &state.Exposure.MaxGain,
			ExposureTime:    &state.Exposure.ExposureTime,
			Gain:            &state.Exposure.Gain,
		},
		Focus: &provider.FocusConfiguration20{
			AutoFocusMode: state.Focus.AutoFocusMode,
			DefaultSpeed:  &state.Focus.DefaultSpeed,
			NearLimit:     &state.Focus.NearLimit,
			FarLimit:      &state.Focus.FarLimit,
		},
		WideDynamicRange: &provider.WideDynamicRangeSettings{
			Mode:  state.WideDynamicRange.Mode,
			Level: &state.WideDynamicRange.Level,
		},
		WhiteBalance: &provider.WhiteBalanceSettings20{
			Mode:   state.WhiteBalance.Mode,
			CrGain: &state.WhiteBalance.CrGain,
			CbGain: &state.WhiteBalance.CbGain,
		},
	}
}

// ImagingStateOf exposes the raw imaging state for hosts embedding the
// simulator (backing Server.GetImagingState).
func (s *Simulator) ImagingStateOf(videoSourceToken string) (provider.ImagingState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.imaging[videoSourceToken]
	if !ok {
		return provider.ImagingState{}, false
	}

	return *state, true
}
