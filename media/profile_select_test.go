package media

// Tests for the main/sub profile selection heuristics (the client-side
// half of sub-stream support) and the error truncation helper.

import (
	"strings"
	"testing"
)

func profileWithRes(token, name string, w, h int) *Profile {
	return &Profile{
		Token: token,
		Name:  name,
		VideoEncoderConfiguration: &VideoEncoderConfiguration{
			Resolution: &VideoResolution{Width: w, Height: h},
		},
	}
}

func TestSelectMainProfile(t *testing.T) {
	cases := []struct {
		name     string
		profiles []*Profile
		want     string
	}{
		{"empty", nil, ""},
		{"highest resolution wins", []*Profile{
			profileWithRes("sub_1", "Sub", 640, 480),
			profileWithRes("main_1", "Main", 2560, 1440),
		}, "main_1"},
		{"resolution beats list order", []*Profile{
			profileWithRes("first_low", "First", 800, 448),
			profileWithRes("second_high", "Second", 1920, 1080),
		}, "second_high"},
		{"tie broken by main hint", []*Profile{
			profileWithRes("profile_2", "Secondary", 1920, 1080),
			profileWithRes("profile_1_main", "Main", 1920, 1080),
		}, "profile_1_main"},
		{"no resolution info falls back to first", []*Profile{
			{Token: "opaque_a"},
			{Token: "opaque_b"},
		}, "opaque_a"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := SelectMainProfile(tc.profiles); got != tc.want {
				t.Errorf("SelectMainProfile = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSelectSubProfile(t *testing.T) {
	cases := []struct {
		name      string
		profiles  []*Profile
		mainToken string
		want      string
	}{
		{"empty", nil, "main_1", ""},
		{"largest strictly-smaller wins", []*Profile{
			profileWithRes("main_1", "Main", 2560, 1440),
			profileWithRes("sub_small", "Small", 320, 240),
			profileWithRes("sub_big", "Big", 704, 576),
		}, "main_1", "sub_big"},
		{"same resolution as main is not a substream", []*Profile{
			profileWithRes("main_1", "Main", 1920, 1080),
			profileWithRes("alias", "Alias", 1920, 1080),
		}, "main_1", ""},
		// An unresolvable main token disables the strictly-smaller guard —
		// every profile is a candidate and the largest wins.
		{"missing main token yields largest overall", []*Profile{
			profileWithRes("a", "A", 1920, 1080),
			profileWithRes("b", "B", 640, 360),
		}, "nonexistent", "a"},
		{"no resolution info falls back to first non-main", []*Profile{
			{Token: "opaque_main"},
			{Token: "opaque_other"},
		}, "opaque_main", "opaque_other"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := SelectSubProfile(tc.profiles, tc.mainToken); got != tc.want {
				t.Errorf("SelectSubProfile = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestProfilePixels(t *testing.T) {
	if got := profilePixels(nil); got != -1 {
		t.Errorf("nil profile = %d", got)
	}

	if got := profilePixels(&Profile{Token: "x"}); got != -1 {
		t.Errorf("no encoder config = %d", got)
	}

	if got := profilePixels(profileWithRes("x", "x", 0, 100)); got != -1 {
		t.Errorf("zero width = %d", got)
	}

	if got := profilePixels(profileWithRes("x", "x", 1920, 1080)); got != 1920*1080 {
		t.Errorf("valid = %d", got)
	}
}

func TestMainTieBreakPrefersMainHints(t *testing.T) {
	mainish := profileWithRes("token_main", "Main", 100, 100)
	subish := profileWithRes("token_sub", "Sub", 100, 100)

	if mainTieBreak(mainish, subish) >= 0 {
		t.Error("main-hint profile must win the main tie-break")
	}

	if mainTieBreak(subish, mainish) <= 0 {
		t.Error("tie-break must be antisymmetric")
	}

	if mainTieBreak(mainish, mainish) != 0 {
		t.Error("identical profiles must tie at zero")
	}
}

func TestTruncateForError(t *testing.T) {
	short := "err"
	if got := truncateForError(short, 100); got != short {
		t.Error("short message altered")
	}

	long := strings.Repeat("x", 10000)
	if got := truncateForError(long, 100); len(got) >= len(long) {
		t.Errorf("long message not truncated: %d bytes", len(got))
	}
}
