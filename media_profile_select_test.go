package onvif

import "testing"

// profile helper: token, name, W×H (-1 resolution = no encoder info).
func p(token, name string, w, h int) *Profile {
	profile := &Profile{Token: token, Name: name}
	if w > 0 {
		profile.VideoEncoderConfiguration = &VideoEncoderConfiguration{
			Resolution: &VideoResolution{Width: w, Height: h},
		}
	}

	return profile
}

func TestSelectMainProfile(t *testing.T) {
	tests := []struct {
		name     string
		profiles []*Profile
		want     string
	}{
		{
			name:     "empty list",
			profiles: nil,
			want:     "",
		},
		{
			name:     "single profile",
			profiles: []*Profile{p("main", "Main", 1920, 1080)},
			want:     "main",
		},
		{
			name: "sub listed first (resolution rules)",
			profiles: []*Profile{
				p("sub_1", "Sub", 640, 360),
				p("main_1", "Main", 1920, 1080),
			},
			want: "main_1",
		},
		{
			name: "main listed first",
			profiles: []*Profile{
				p("main_1", "Main", 1920, 1080),
				p("sub_1", "Sub", 640, 360),
			},
			want: "main_1",
		},
		{
			name: "tie broken by main naming hint",
			profiles: []*Profile{
				p("profile_1", "Quality1", 1920, 1080),
				p("mainStream", "Quality2", 1920, 1080),
			},
			want: "mainStream",
		},
		{
			name: "tie: sub-hint loses to neutral",
			profiles: []*Profile{
				p("extra_1", "Extra", 1920, 1080),
				p("profile_1", "Quality1", 1920, 1080),
			},
			want: "profile_1",
		},
		{
			name: "chinese hints: 主码流 wins tie",
			profiles: []*Profile{
				p("tok_a", "高清子码流", 1280, 720),
				p("tok_b", "主码流", 1280, 720),
			},
			want: "tok_b",
		},
		{
			name: "no resolution info falls back to list order",
			profiles: []*Profile{
				p("first", "Whatever", 0, 0),
				p("second", "Main", 0, 0),
			},
			want: "first",
		},
		{
			name:     "nil profile entries are safe",
			profiles: []*Profile{nil, p("main", "Main", 640, 480)},
			want:     "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SelectMainProfile(tt.profiles); got != tt.want {
				t.Errorf("SelectMainProfile() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSelectSubProfile(t *testing.T) {
	main1920 := p("main_1", "Main", 1920, 1080)

	tests := []struct {
		name      string
		profiles  []*Profile
		mainToken string
		want      string
	}{
		{
			name:      "picks largest strictly-smaller profile",
			profiles:  []*Profile{main1920, p("sub_hd", "HD", 1280, 720), p("sub_sd", "SD", 640, 360)},
			mainToken: "main_1",
			want:      "sub_hd",
		},
		{
			name:      "same resolution as main is not a substream (dual token)",
			profiles:  []*Profile{main1920, p("token_b", "MainB", 1920, 1080), p("sub", "Sub", 640, 360)},
			mainToken: "main_1",
			want:      "sub",
		},
		{
			name:      "only same-resolution profiles means no substream",
			profiles:  []*Profile{main1920, p("token_b", "MainB", 1920, 1080)},
			mainToken: "main_1",
			want:      "",
		},
		{
			name:      "single profile has no substream",
			profiles:  []*Profile{main1920},
			mainToken: "main_1",
			want:      "",
		},
		{
			name:      "empty list",
			profiles:  nil,
			mainToken: "main_1",
			want:      "",
		},
		{
			name:      "unknown main resolution falls back to list order",
			profiles:  []*Profile{p("first", "A", 0, 0), p("second", "B", 0, 0)},
			mainToken: "first",
			want:      "second",
		},
		{
			name:      "unknown resolutions besides main return nothing rankable",
			profiles:  []*Profile{p("main", "M", 1920, 1080), p("other", "O", 0, 0)},
			mainToken: "main",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SelectSubProfile(tt.profiles, tt.mainToken); got != tt.want {
				t.Errorf("SelectSubProfile() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSelectProfilesRoundTrip mirrors the documented usage: pick main, then
// sub, on a realistic 3-profile camera.
func TestSelectProfilesRoundTrip(t *testing.T) {
	profiles := []*Profile{
		p("profile_2", "substream", 640, 360),
		p("profile_1", "hq", 2560, 1440),
		p("profile_3", "mq", 1280, 720),
	}

	if got := SelectMainProfile(profiles); got != "profile_1" {
		t.Fatalf("SelectMainProfile() = %q, want profile_1", got)
	}

	if got := SelectSubProfile(profiles, "profile_1"); got != "profile_3" {
		t.Fatalf("SelectSubProfile() = %q, want profile_3", got)
	}
}
