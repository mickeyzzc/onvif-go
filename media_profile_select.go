package onvif

import "strings"

// Naming hints used ONLY as a tie-breaker between profiles of equal
// resolution. OEM firmware assigns arbitrary profile names, so naming can
// never be the primary signal (issue #7); these hints settle exact ties and
// actively demote obviously-secondary profiles.
var (
	mainStreamHints = []string{"main", "primary", "主流", "主码流", "channel1"}
	subStreamHints  = []string{"sub", "secondary", "辅流", "辅码流", "extra"}
)

// SelectMainProfile returns the token of the main-stream profile: the one
// with the most pixels (W×H). Blindly using profiles[0] silently records a
// substream on devices that list the low-resolution profile first.
//
// Heuristics (field-tested against Hikvision, Dahua, Amcrest, HiSilicon OEM
// and ESP32 hardware):
//   - Resolution rules. Ties are settled by naming hints only: a candidate
//     whose token/name contains a main hint ("main", "primary", "主流",
//     "主码流", "channel1") wins; candidates carrying sub hints ("sub",
//     "secondary", "辅流", "辅码流", "extra") lose ties.
//   - When no profile carries resolution information at all, the first
//     profile is returned (list-order fallback).
//
// Returns "" for an empty list.
func SelectMainProfile(profiles []*Profile) string {
	if len(profiles) == 0 {
		return ""
	}

	best := -1
	bestPixels := int64(-1)
	anyKnown := false

	for i, p := range profiles {
		pixels := profilePixels(p)
		if pixels < 0 {
			continue
		}

		anyKnown = true
		switch {
		case pixels > bestPixels:
			best, bestPixels = i, pixels
		case pixels == bestPixels:
			if mainTieBreak(profiles[i], profiles[best]) < 0 {
				best = i
			}
		}
	}

	if !anyKnown {
		return profiles[0].Token
	}

	return profiles[best].Token
}

// SelectSubProfile returns the token of the sub-stream profile given the
// already-selected main token: among the remaining profiles it is the one
// with the most pixels that is still STRICTLY smaller than the main
// resolution. A second profile with the SAME resolution as main is not a
// substream — on some hardware (e.g. Amcrest IP4M) two tokens at the same
// resolution are two handles onto the same stream.
//
// Returns "" when there is no independent substream (or when mainToken does
// not resolve). If nothing carries resolution information, the first
// non-main profile is returned as a list-order fallback.
func SelectSubProfile(profiles []*Profile, mainToken string) string {
	if len(profiles) == 0 {
		return ""
	}

	mainPixels := int64(-1)
	anyKnown := false
	for _, p := range profiles {
		if pixels := profilePixels(p); pixels >= 0 {
			anyKnown = true
			if p != nil && p.Token == mainToken {
				mainPixels = pixels
			}
		}
	}

	best := ""
	bestPixels := int64(-1)

	for _, p := range profiles {
		if p == nil || p.Token == mainToken || p.Token == "" {
			continue
		}

		pixels := profilePixels(p)
		if pixels < 0 {
			continue
		}

		// Same resolution as main = same stream under another token.
		if mainPixels >= 0 && pixels >= mainPixels {
			continue
		}

		switch {
		case pixels > bestPixels:
			best, bestPixels = p.Token, pixels
		case pixels == bestPixels:
			if best != "" && subTieBreak(p, profileByToken(profiles, best)) < 0 {
				best = p.Token
			}
		}
	}

	if best != "" {
		return best
	}

	// List-order fallback when resolution information is absent.
	if !anyKnown {
		for _, p := range profiles {
			if p != nil && p.Token != "" && p.Token != mainToken {
				return p.Token
			}
		}
	}

	return ""
}

// profilePixels returns W×H of the profile's video encoder configuration, or
// -1 when the profile or its resolution is unknown.
func profilePixels(p *Profile) int64 {
	if p == nil || p.VideoEncoderConfiguration == nil || p.VideoEncoderConfiguration.Resolution == nil {
		return -1
	}

	w := int64(p.VideoEncoderConfiguration.Resolution.Width)
	h := int64(p.VideoEncoderConfiguration.Resolution.Height)
	if w <= 0 || h <= 0 {
		return -1
	}

	return w * h
}

// mainTieBreak compares two same-resolution profiles for main-stream
// preference. Returns <0 when a wins (main-ish naming beats neutral beats
// sub-ish).
func mainTieBreak(a, b *Profile) int {
	return hintScore(b) - hintScore(a)
}

// subTieBreak compares two same-resolution profiles for sub-stream
// preference. Returns <0 when a wins (sub-ish naming beats neutral).
func subTieBreak(a, b *Profile) int {
	return hintScore(a) - hintScore(b)
}

// hintScore rates a profile's naming hints: positive = main-ish, negative =
// sub-ish, zero = neutral.
func hintScore(p *Profile) int {
	if p == nil {
		return 0
	}

	name := strings.ToLower(p.Token + " " + p.Name)

	for _, hint := range mainStreamHints {
		if strings.Contains(name, strings.ToLower(hint)) {
			return 1
		}
	}

	for _, hint := range subStreamHints {
		if strings.Contains(name, strings.ToLower(hint)) {
			return -1
		}
	}

	return 0
}

// profileByToken finds the profile carrying the given token (nil when absent).
func profileByToken(profiles []*Profile, token string) *Profile {
	for _, p := range profiles {
		if p != nil && p.Token == token {
			return p
		}
	}

	return nil
}
